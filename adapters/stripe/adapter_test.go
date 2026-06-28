package stripe

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/albert-einshutoin/mockport/internal/adapter"
	"github.com/albert-einshutoin/mockport/internal/adapter/adaptertest"
)

func TestCheckoutSessionSuccess(t *testing.T) {
	rec := performStripeRequest(t, adapter.Config{BasePath: "/stripe", Scenario: "payment_success"}, http.MethodPost, "/stripe/v1/checkout/sessions")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["object"] != "checkout.session" || body["payment_status"] != "paid" {
		t.Fatalf("unexpected body: %#v", body)
	}
}

func TestCheckoutSessionPaymentFailed(t *testing.T) {
	rec := performStripeRequest(t, adapter.Config{BasePath: "/stripe", Scenario: "payment_failed"}, http.MethodPost, "/stripe/v1/checkout/sessions")
	if rec.Code != http.StatusPaymentRequired {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusPaymentRequired)
	}
	assertStripeErrorCode(t, rec, "card_declined")
}

func TestPaymentIntentScenarios(t *testing.T) {
	tests := []struct {
		name     string
		scenario string
		wantCode int
		wantErr  string
	}{
		{"success", "payment_success", http.StatusOK, ""},
		{"failed", "payment_failed", http.StatusPaymentRequired, "card_declined"},
		{"auth", "auth_error", http.StatusUnauthorized, "invalid_api_key"},
		{"rate limited", "rate_limited", http.StatusTooManyRequests, "rate_limited"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := performStripeRequest(t, adapter.Config{BasePath: "/stripe", Scenario: tt.scenario}, http.MethodPost, "/stripe/v1/payment_intents")
			if rec.Code != tt.wantCode {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantCode)
			}
			if tt.wantErr != "" {
				assertStripeErrorCode(t, rec, tt.wantErr)
			}
		})
	}
}

func TestGetPaymentIntent(t *testing.T) {
	rec := performStripeRequest(t, adapter.Config{BasePath: "/stripe", Scenario: "payment_success"}, http.MethodGet, "/stripe/v1/payment_intents/pi_mockport")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestCheckoutSessionCreateRetrieveListAndIdempotency(t *testing.T) {
	mux := newStripeMux(t, adapter.Config{BasePath: "/stripe", Scenario: "payment_success"})

	first := serveStripeRequest(mux, http.MethodPost, "/stripe/v1/checkout/sessions", "client_reference_id=cart_1", map[string]string{"Idempotency-Key": "cart-1"})
	if first.Code != http.StatusOK {
		t.Fatalf("first status = %d, want %d, body=%s", first.Code, http.StatusOK, first.Body.String())
	}
	var created map[string]any
	if err := json.Unmarshal(first.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode first response: %v", err)
	}
	if created["id"] != "stripe_checkout_session_000001" {
		t.Fatalf("created id = %#v", created["id"])
	}

	replay := serveStripeRequest(mux, http.MethodPost, "/stripe/v1/checkout/sessions", "client_reference_id=cart_1", map[string]string{"Idempotency-Key": "cart-1"})
	if replay.Body.String() != first.Body.String() {
		t.Fatalf("replay body = %s, want %s", replay.Body.String(), first.Body.String())
	}

	conflict := serveStripeRequest(mux, http.MethodPost, "/stripe/v1/checkout/sessions", "client_reference_id=cart_2", map[string]string{"Idempotency-Key": "cart-1"})
	if conflict.Code != http.StatusConflict {
		t.Fatalf("conflict status = %d, want %d, body=%s", conflict.Code, http.StatusConflict, conflict.Body.String())
	}
	assertStripeErrorCode(t, conflict, "idempotency_key_in_use")

	retrieved := serveStripeRequest(mux, http.MethodGet, "/stripe/v1/checkout/sessions/stripe_checkout_session_000001", "", nil)
	if retrieved.Code != http.StatusOK {
		t.Fatalf("retrieve status = %d, want %d", retrieved.Code, http.StatusOK)
	}
	var got map[string]any
	if err := json.Unmarshal(retrieved.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode retrieve: %v", err)
	}
	if got["id"] != created["id"] || got["client_reference_id"] != "cart_1" {
		t.Fatalf("retrieved = %#v", got)
	}

	list := serveStripeRequest(mux, http.MethodGet, "/stripe/v1/checkout/sessions", "", nil)
	if list.Code != http.StatusOK {
		t.Fatalf("list status = %d, want %d", list.Code, http.StatusOK)
	}
	var listed struct {
		Object string           `json:"object"`
		Data   []map[string]any `json:"data"`
	}
	if err := json.Unmarshal(list.Body.Bytes(), &listed); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if listed.Object != "list" || len(listed.Data) != 1 || listed.Data[0]["id"] != created["id"] {
		t.Fatalf("list = %#v", listed)
	}
}

func TestCheckoutSessionIdempotencyIsAtomicUnderConcurrentRequests(t *testing.T) {
	mux := newStripeMux(t, adapter.Config{BasePath: "/stripe", Scenario: "payment_success"})
	const requests = 64

	bodies := adaptertest.ConcurrentResults(requests, func() map[string]any {
		rec := serveStripeRequest(mux, http.MethodPost, "/stripe/v1/checkout/sessions", "client_reference_id=cart_race", map[string]string{"Idempotency-Key": "cart-race"})
		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
			return nil
		}
		var body map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Errorf("decode response: %v", err)
			return nil
		}
		return body
	})
	if t.Failed() {
		return
	}

	ids := map[string]int{}
	for _, body := range bodies {
		id, ok := body["id"].(string)
		if !ok || id == "" {
			t.Fatalf("response without id: %#v", body)
		}
		ids[id]++
	}
	if len(ids) != 1 {
		t.Fatalf("created ids = %#v, want one replayed id", ids)
	}

	list := serveStripeRequest(mux, http.MethodGet, "/stripe/v1/checkout/sessions", "", nil)
	if list.Code != http.StatusOK {
		t.Fatalf("list status = %d, want %d", list.Code, http.StatusOK)
	}
	var listed struct {
		Data []map[string]any `json:"data"`
	}
	if err := json.Unmarshal(list.Body.Bytes(), &listed); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(listed.Data) != 1 {
		t.Fatalf("listed sessions = %#v, want one side effect", listed.Data)
	}
}

func TestPaymentIntentCreateRetrieveAndList(t *testing.T) {
	mux := newStripeMux(t, adapter.Config{BasePath: "/stripe", Scenario: "payment_success"})

	create := serveStripeRequest(mux, http.MethodPost, "/stripe/v1/payment_intents", "amount=1200&currency=usd", nil)
	if create.Code != http.StatusOK {
		t.Fatalf("create status = %d, want %d, body=%s", create.Code, http.StatusOK, create.Body.String())
	}
	var created map[string]any
	if err := json.Unmarshal(create.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode create: %v", err)
	}
	if created["id"] != "stripe_payment_intent_000001" || created["amount"].(float64) != 1200 {
		t.Fatalf("created = %#v", created)
	}

	retrieved := serveStripeRequest(mux, http.MethodGet, "/stripe/v1/payment_intents/stripe_payment_intent_000001", "", nil)
	if retrieved.Code != http.StatusOK {
		t.Fatalf("retrieve status = %d, want %d", retrieved.Code, http.StatusOK)
	}
	list := serveStripeRequest(mux, http.MethodGet, "/stripe/v1/payment_intents", "", nil)
	if list.Code != http.StatusOK || !strings.Contains(list.Body.String(), "stripe_payment_intent_000001") {
		t.Fatalf("list status/body = %d %s", list.Code, list.Body.String())
	}
}

func TestMajorStripeResourcesCreateRetrieveAndList(t *testing.T) {
	mux := newStripeMux(t, adapter.Config{BasePath: "/stripe", Scenario: "payment_success"})

	customer := createStripeResource(t, mux, "/stripe/v1/customers", "email=customer@example.test")
	product := createStripeResource(t, mux, "/stripe/v1/products", "name=Mockport+Product")
	price := createStripeResource(t, mux, "/stripe/v1/prices", "product="+product["id"].(string)+"&currency=usd&unit_amount=1200")
	subscription := createStripeResource(t, mux, "/stripe/v1/subscriptions", "customer="+customer["id"].(string)+"&items[0][price]="+price["id"].(string))
	invoice := createStripeResource(t, mux, "/stripe/v1/invoices", "customer="+customer["id"].(string))
	paymentIntent := createStripeResource(t, mux, "/stripe/v1/payment_intents", "amount=1200&currency=usd")
	refund := createStripeResource(t, mux, "/stripe/v1/refunds", "payment_intent="+paymentIntent["id"].(string))

	for _, entry := range []struct {
		path string
		body map[string]any
	}{
		{"/stripe/v1/customers", customer},
		{"/stripe/v1/products", product},
		{"/stripe/v1/prices", price},
		{"/stripe/v1/subscriptions", subscription},
		{"/stripe/v1/invoices", invoice},
		{"/stripe/v1/refunds", refund},
	} {
		retrieved := serveStripeRequest(mux, http.MethodGet, entry.path+"/"+entry.body["id"].(string), "", nil)
		if retrieved.Code != http.StatusOK {
			t.Fatalf("retrieve %s status = %d body=%s", entry.path, retrieved.Code, retrieved.Body.String())
		}
		listed := serveStripeRequest(mux, http.MethodGet, entry.path, "", nil)
		if listed.Code != http.StatusOK || !strings.Contains(listed.Body.String(), entry.body["id"].(string)) {
			t.Fatalf("list %s status/body = %d %s", entry.path, listed.Code, listed.Body.String())
		}
	}
}

func TestStripeValidationAndErrorHeaders(t *testing.T) {
	mux := newStripeMux(t, adapter.Config{BasePath: "/stripe", Scenario: "payment_success"})

	missing := serveStripeRequest(mux, http.MethodPost, "/stripe/v1/payment_intents", "currency=usd", nil)
	if missing.Code != http.StatusBadRequest {
		t.Fatalf("missing status = %d, want %d, body=%s", missing.Code, http.StatusBadRequest, missing.Body.String())
	}
	assertStripeErrorCode(t, missing, "parameter_missing")
	if missing.Header().Get("Request-Id") == "" || missing.Header().Get("Stripe-Version") == "" {
		t.Fatalf("missing stripe-like headers: %#v", missing.Header())
	}

	malformed := serveStripeRequest(mux, http.MethodGet, "/stripe/v1/payment_intents/not_a_pi", "", nil)
	if malformed.Code != http.StatusNotFound {
		t.Fatalf("malformed status = %d, want %d, body=%s", malformed.Code, http.StatusNotFound, malformed.Body.String())
	}
	assertStripeErrorCode(t, malformed, "resource_missing")
}

func TestStripeRejectsOversizedFormBody(t *testing.T) {
	mux := newStripeMux(t, adapter.Config{BasePath: "/stripe", Scenario: "payment_success"})
	body := "amount=1200&currency=usd&metadata=" + strings.Repeat("x", 1<<20)

	rec := serveStripeRequest(mux, http.MethodPost, "/stripe/v1/payment_intents", body, nil)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusRequestEntityTooLarge, rec.Body.String())
	}
	assertStripeErrorCode(t, rec, "request_too_large")
}

func TestTimeoutScenario(t *testing.T) {
	rec := performStripeRequest(t, adapter.Config{BasePath: "/stripe", Scenario: "timeout"}, http.MethodPost, "/stripe/v1/checkout/sessions")
	if rec.Code != http.StatusGatewayTimeout {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusGatewayTimeout)
	}
	assertStripeErrorCode(t, rec, "mockport_timeout")
}

func TestWebhookSender(t *testing.T) {
	received := make(chan *http.Request, 1)
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Stripe-Signature") == "" {
			t.Error("missing Stripe-Signature header")
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode webhook body: %v", err)
		}
		if body["type"] != "checkout.session.completed" {
			t.Errorf("event type = %v, want checkout.session.completed", body["type"])
		}
		received <- r
		w.WriteHeader(http.StatusNoContent)
	}))
	defer target.Close()

	rec := performStripeRequest(t, adapter.Config{
		BasePath:             "/stripe",
		Scenario:             "payment_success",
		WebhookTargetURL:     target.URL,
		WebhookSigningSecret: "whsec_mockport",
	}, http.MethodPost, "/stripe/test/webhook/send")
	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusAccepted, rec.Body.String())
	}
	select {
	case <-received:
	default:
		t.Fatal("webhook target did not receive request")
	}
}

func TestWebhookSenderRejectsNonLocalTriggerAndUnsafeTarget(t *testing.T) {
	mux := newStripeMux(t, adapter.Config{
		BasePath:             "/stripe",
		Scenario:             "payment_success",
		WebhookTargetURL:     "http://127.0.0.1:3000/webhooks/stripe",
		WebhookSigningSecret: "whsec_mockport",
	})
	remote := serveStripeRequestWithRemote(mux, http.MethodPost, "/stripe/test/webhook/send", "", nil, "192.168.1.10:12345")
	if remote.Code != http.StatusForbidden {
		t.Fatalf("remote trigger status = %d, want %d, body=%s", remote.Code, http.StatusForbidden, remote.Body.String())
	}
	assertStripeErrorCode(t, remote, "local_request_required")

	unsafe := performStripeRequest(t, adapter.Config{
		BasePath:             "/stripe",
		Scenario:             "payment_success",
		WebhookTargetURL:     "http://169.254.169.254/latest/meta-data",
		WebhookSigningSecret: "whsec_mockport",
	}, http.MethodPost, "/stripe/test/webhook/send")
	if unsafe.Code != http.StatusBadRequest {
		t.Fatalf("unsafe target status = %d, want %d, body=%s", unsafe.Code, http.StatusBadRequest, unsafe.Body.String())
	}
	assertStripeErrorCode(t, unsafe, "unsafe_webhook_target")
}

func TestMetadata(t *testing.T) {
	meta := New().Metadata()
	if meta.Name != "stripe" {
		t.Fatalf("name = %q, want stripe", meta.Name)
	}
	if meta.Maturity != "workflow-compatible" {
		t.Fatalf("maturity = %q, want workflow-compatible", meta.Maturity)
	}
	if meta.ProviderVersion != "2025-10-29.clover" || len(meta.SDKVersions) != 1 {
		t.Fatalf("compat metadata = %#v", meta)
	}
	if len(meta.Capabilities) < 9 {
		t.Fatal("expected capabilities")
	}
	if len(meta.Scenarios) != 5 {
		t.Fatalf("scenario count = %d, want 5", len(meta.Scenarios))
	}
	if len(meta.Endpoints) != 26 {
		t.Fatalf("endpoint count = %d, want 26", len(meta.Endpoints))
	}
	if !meta.Idempotency || !meta.Reset || len(meta.StatefulResources) != 8 {
		t.Fatalf("state metadata = %#v", meta)
	}
}

func TestResetClearsStateAndRequiresLoopback(t *testing.T) {
	mux := newStripeMux(t, adapter.Config{
		BasePath:             "/stripe",
		Scenario:             "payment_success",
		WebhookSigningSecret: "",
		WebhookTargetURL:     "http://127.0.0.1:3000/webhook",
	})

	created := createStripeResource(t, mux, "/stripe/v1/customers", `{"email":"user@example.test"}`)
	customerID, _ := created["id"].(string)

	list := serveStripeRequest(mux, http.MethodGet, "/stripe/v1/customers", "", nil)
	if list.Code != http.StatusOK {
		t.Fatalf("customers list before reset status = %d, body=%s", list.Code, list.Body.String())
	}
	var before struct {
		Data []map[string]any `json:"data"`
	}
	if err := json.Unmarshal(list.Body.Bytes(), &before); err != nil {
		t.Fatalf("decode customers list before reset: %v", err)
	}
	if len(before.Data) != 1 {
		t.Fatalf("customers before reset = %#v", before.Data)
	}

	reset := serveStripeRequestWithRemote(mux, http.MethodPost, "/stripe/test/reset", "", nil, "127.0.0.1:12345")
	if reset.Code != http.StatusOK {
		t.Fatalf("reset status = %d, body=%s", reset.Code, reset.Body.String())
	}
	var resetBody map[string]any
	if err := json.Unmarshal(reset.Body.Bytes(), &resetBody); err != nil {
		t.Fatalf("decode reset response: %v", err)
	}
	if resetBody["reset"] != true || resetBody["adapter"] != "stripe" {
		t.Fatalf("reset body = %#v", resetBody)
	}

	lookupAfter := serveStripeRequest(mux, http.MethodGet, "/stripe/v1/customers/"+customerID, "", nil)
	if lookupAfter.Code != http.StatusNotFound {
		t.Fatalf("customer lookup after reset status = %d, body=%s", lookupAfter.Code, lookupAfter.Body.String())
	}

	listAfter := serveStripeRequest(mux, http.MethodGet, "/stripe/v1/customers", "", nil)
	if listAfter.Code != http.StatusOK {
		t.Fatalf("customers list after reset status = %d, body=%s", listAfter.Code, listAfter.Body.String())
	}
	var after struct {
		Data []map[string]any `json:"data"`
	}
	if err := json.Unmarshal(listAfter.Body.Bytes(), &after); err != nil {
		t.Fatalf("decode customers list after reset: %v", err)
	}
	if len(after.Data) != 0 {
		t.Fatalf("customers after reset should be empty: %#v", after.Data)
	}

	remoteReset := serveStripeRequestWithRemote(mux, http.MethodPost, "/stripe/test/reset", "", nil, "192.168.0.2:12345")
	if remoteReset.Code != http.StatusForbidden {
		t.Fatalf("remote reset status = %d, body=%s", remoteReset.Code, remoteReset.Body.String())
	}
	assertStripeErrorCode(t, remoteReset, "local_request_required")
}

func TestUnsupportedMethodReturns405(t *testing.T) {
	mux := newStripeMux(t, adapter.Config{BasePath: "/stripe", Scenario: "payment_success"})

	for _, path := range []string{"/stripe/v1/customers", "/v1/customers"} {
		rec := serveStripeRequest(mux, http.MethodDelete, path, "", nil)
		if rec.Code != http.StatusMethodNotAllowed {
			t.Fatalf("DELETE %s status = %d, want %d", path, rec.Code, http.StatusMethodNotAllowed)
		}
		if allow := rec.Header().Get("Allow"); !strings.Contains(allow, http.MethodGet) || !strings.Contains(allow, http.MethodPost) {
			t.Fatalf("DELETE %s Allow = %q, want GET and POST", path, allow)
		}
	}
}

func TestRootAliasPaths(t *testing.T) {
	mux := newStripeMux(t, adapter.Config{BasePath: "/stripe", Scenario: "payment_success"})

	create := serveStripeRequest(mux, http.MethodPost, "/v1/customers", "email=alias@example.test", nil)
	if create.Code != http.StatusOK {
		t.Fatalf("create via /v1 alias status = %d, body=%s", create.Code, create.Body.String())
	}
	var customer map[string]any
	if err := json.Unmarshal(create.Body.Bytes(), &customer); err != nil {
		t.Fatalf("decode create: %v", err)
	}
	id, _ := customer["id"].(string)
	if id == "" {
		t.Fatalf("created customer without id: %#v", customer)
	}

	retrieved := serveStripeRequest(mux, http.MethodGet, "/v1/customers/"+id, "", nil)
	if retrieved.Code != http.StatusOK {
		t.Fatalf("retrieve via /v1 alias status = %d, body=%s", retrieved.Code, retrieved.Body.String())
	}
}

func TestGenericResourceDoesNotLeakFormFieldsBetweenRequests(t *testing.T) {
	mux := newStripeMux(t, adapter.Config{BasePath: "/stripe", Scenario: "payment_success"})

	first := serveStripeRequest(mux, http.MethodPost, "/stripe/v1/products", "name=First+Product&description=leak-me", nil)
	if first.Code != http.StatusOK {
		t.Fatalf("first create status = %d, body=%s", first.Code, first.Body.String())
	}
	second := serveStripeRequest(mux, http.MethodPost, "/stripe/v1/products", "name=Second+Product", nil)
	if second.Code != http.StatusOK {
		t.Fatalf("second create status = %d, body=%s", second.Code, second.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(second.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode second response: %v", err)
	}
	if _, ok := body["description"]; ok {
		t.Fatalf("second product leaked description from first request: %#v", body)
	}
	if body["name"] != "Second Product" {
		t.Fatalf("name = %#v, want Second Product", body["name"])
	}
}

func performStripeRequest(t *testing.T, cfg adapter.Config, method, path string) *httptest.ResponseRecorder {
	t.Helper()
	mux := newStripeMux(t, cfg)
	return serveStripeRequest(mux, method, path, "", nil)
}

func newStripeMux(t *testing.T, cfg adapter.Config) *http.ServeMux {
	t.Helper()
	return adaptertest.NewMux(t, New(), cfg)
}

func createStripeResource(t *testing.T, mux http.Handler, path, body string) map[string]any {
	t.Helper()
	rec := serveStripeRequest(mux, http.MethodPost, path, body, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("create %s status = %d, body=%s", path, rec.Code, rec.Body.String())
	}
	var decoded map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &decoded); err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
	if decoded["id"] == "" {
		t.Fatalf("created %s without id: %#v", path, decoded)
	}
	return decoded
}

func serveStripeRequest(mux http.Handler, method, path, body string, headers map[string]string) *httptest.ResponseRecorder {
	return serveStripeRequestWithRemote(mux, method, path, body, headers, "127.0.0.1:12345")
}

func serveStripeRequestWithRemote(mux http.Handler, method, path, body string, headers map[string]string, remoteAddr string) *httptest.ResponseRecorder {
	header := http.Header{}
	if body != "" {
		header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	for name, value := range headers {
		header.Set(name, value)
	}
	return adaptertest.ServeWithRemote(mux, method, path, strings.NewReader(body), header, remoteAddr)
}

func assertStripeErrorCode(t *testing.T, rec *httptest.ResponseRecorder, want string) {
	t.Helper()
	adaptertest.AssertJSONField(t, rec, "error.code", want)
}
