package stripe_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/albert-einshutoin/mockport/adapters/stripe"
	"github.com/albert-einshutoin/mockport/internal/adapter"
)

func newStripeMux(t *testing.T, cfg adapter.Config) *http.ServeMux {
	t.Helper()
	mux := http.NewServeMux()
	if err := stripe.New().Register(mux, cfg); err != nil {
		t.Fatalf("Register: %v", err)
	}
	return mux
}

func TestStripeHeaderOverridesConfig(t *testing.T) {
	// config=payment_success でサーバーを起動し、ヘッダで payment_failed を指定
	mux := newStripeMux(t, adapter.Config{BasePath: "/stripe", Scenario: "payment_success"})
	req := httptest.NewRequest(http.MethodPost, "/stripe/v1/checkout/sessions", nil)
	req.Header.Set("X-Mockport-Scenario", "payment_failed")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusPaymentRequired {
		t.Errorf("want 402, got %d", rec.Code)
	}
	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	errObj, _ := body["error"].(map[string]any)
	if errObj == nil {
		t.Fatal("want error object in response body")
	}
}

func TestStripeUnknownScenarioReturns400(t *testing.T) {
	mux := newStripeMux(t, adapter.Config{BasePath: "/stripe", Scenario: "payment_success"})
	req := httptest.NewRequest(http.MethodPost, "/stripe/v1/checkout/sessions", nil)
	req.Header.Set("X-Mockport-Scenario", "totally_unknown")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", rec.Code)
	}
	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	errObj, _ := body["error"].(map[string]any)
	if errObj == nil {
		t.Fatal("want error object")
	}
	code, _ := errObj["code"].(string)
	if code != "unknown_mockport_scenario" {
		t.Errorf("want code=unknown_mockport_scenario, got %q", code)
	}
}

// TestStripeGenericResourceUnknownScenarioReturns400 は generic resource/list/lookup ハンドラ
// （POST /stripe/v1/customers など）でも未知の X-Mockport-Scenario が 400 になることを固定する。
// 修正前はこれらのエンドポイントが resolver を呼ばず 200 で成功していた（指摘2）。
func TestStripeGenericResourceUnknownScenarioReturns400(t *testing.T) {
	endpoints := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/stripe/v1/customers"},
		{http.MethodGet, "/stripe/v1/customers"},
		{http.MethodGet, "/stripe/v1/customers/cus_mockport"},
	}
	for _, ep := range endpoints {
		ep := ep
		t.Run(ep.method+" "+ep.path, func(t *testing.T) {
			mux := newStripeMux(t, adapter.Config{BasePath: "/stripe", Scenario: "payment_success"})
			req := httptest.NewRequest(ep.method, ep.path, nil)
			req.Header.Set("X-Mockport-Scenario", "totally_unknown")
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)
			if rec.Code != http.StatusBadRequest {
				t.Errorf("want 400, got %d", rec.Code)
			}
			var body map[string]any
			if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
				t.Fatalf("decode: %v", err)
			}
			errObj, _ := body["error"].(map[string]any)
			if errObj == nil {
				t.Fatal("want error object")
			}
			code, _ := errObj["code"].(string)
			if code != "unknown_mockport_scenario" {
				t.Errorf("want code=unknown_mockport_scenario, got %q", code)
			}
		})
	}
}

// TestStripeCoreGetUnknownScenarioReturns400 は checkout session / payment intent の
// GET(取得/list/lookup)系ハンドラでも未知の X-Mockport-Scenario が 400 になることを固定する。
// 修正前はこれらのGET系が registerV1Routes に直書きされ resolver を呼ばず 200 で成功していた（Codex High）。
func TestStripeCoreGetUnknownScenarioReturns400(t *testing.T) {
	endpoints := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/stripe/v1/checkout/sessions"},
		{http.MethodGet, "/stripe/v1/checkout/sessions/cs_test_mockport"},
		{http.MethodGet, "/stripe/v1/payment_intents"},
		{http.MethodGet, "/stripe/v1/payment_intents/pi_mockport"},
	}
	for _, ep := range endpoints {
		ep := ep
		t.Run(ep.method+" "+ep.path, func(t *testing.T) {
			mux := newStripeMux(t, adapter.Config{BasePath: "/stripe", Scenario: "payment_success"})
			req := httptest.NewRequest(ep.method, ep.path, nil)
			req.Header.Set("X-Mockport-Scenario", "totally_unknown")
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)
			if rec.Code != http.StatusBadRequest {
				t.Errorf("want 400, got %d", rec.Code)
			}
			var body map[string]any
			if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
				t.Fatalf("decode: %v", err)
			}
			errObj, _ := body["error"].(map[string]any)
			if errObj == nil {
				t.Fatal("want error object")
			}
			code, _ := errObj["code"].(string)
			if code != "unknown_mockport_scenario" {
				t.Errorf("want code=unknown_mockport_scenario, got %q", code)
			}
		})
	}
}
