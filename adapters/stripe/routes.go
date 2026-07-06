package stripe

import (
	"encoding/json"
	"errors"
	"maps"
	"net/http"
	"strconv"
	"strings"

	"github.com/albert-einshutoin/mockport/internal/adapter"
	"github.com/albert-einshutoin/mockport/internal/adapter/httpx"
	"github.com/albert-einshutoin/mockport/internal/state"
)

type routes struct {
	basePath    string
	cfg         adapter.Config
	store       *state.Store
	idempotency *state.IdempotencyStore
	resolver    *adapter.ScenarioResolver
}

func (rt *routes) register(mux *http.ServeMux, prefix string) {
	rt.registerV1Routes(mux, prefix)
	rt.registerTestRoutes(mux, prefix)
}

func (rt *routes) registerV1Routes(mux *http.ServeMux, prefix string) {
	handleLimited(mux, "POST "+prefix+"/v1/checkout/sessions", rt.writeCheckoutSession)
	handleLimited(mux, "GET "+prefix+"/v1/checkout/sessions", func(w http.ResponseWriter, r *http.Request) {
		if _, ok := rt.resolveScenario(w, r); !ok {
			return
		}
		rt.writeList(w, "checkout_session")
	})
	handleLimited(mux, "GET "+prefix+"/v1/checkout/sessions/{id}", func(w http.ResponseWriter, r *http.Request) {
		if _, ok := rt.resolveScenario(w, r); !ok {
			return
		}
		rt.writeResource(w, "checkout_session", r.PathValue("id"), fallbackCheckoutSession)
	})
	handleLimited(mux, "POST "+prefix+"/v1/payment_intents", rt.writePaymentIntent)
	handleLimited(mux, "GET "+prefix+"/v1/payment_intents", func(w http.ResponseWriter, r *http.Request) {
		if _, ok := rt.resolveScenario(w, r); !ok {
			return
		}
		rt.writeList(w, "payment_intent")
	})
	handleLimited(mux, "GET "+prefix+"/v1/payment_intents/{id}", func(w http.ResponseWriter, r *http.Request) {
		if _, ok := rt.resolveScenario(w, r); !ok {
			return
		}
		rt.writeResource(w, "payment_intent", r.PathValue("id"), fallbackPaymentIntent)
	})
	rt.registerResource(mux, prefix, "customer", "/v1/customers", nil, map[string]any{"object": "customer"}, nil)
	rt.registerResource(mux, prefix, "product", "/v1/products", nil, map[string]any{"object": "product", "active": true}, []string{"name"})
	rt.registerResource(mux, prefix, "price", "/v1/prices", nil, map[string]any{"object": "price", "active": true}, []string{"product", "currency", "unit_amount"})
	rt.registerResource(mux, prefix, "subscription", "/v1/subscriptions", nil, map[string]any{"object": "subscription", "status": "active"}, []string{"customer"})
	rt.registerResource(mux, prefix, "invoice", "/v1/invoices", nil, map[string]any{"object": "invoice", "status": "draft"}, []string{"customer"})
	rt.registerResource(mux, prefix, "refund", "/v1/refunds", nil, map[string]any{"object": "refund", "status": "succeeded"}, []string{"payment_intent"})
}

func (rt *routes) registerTestRoutes(mux *http.ServeMux, prefix string) {
	handleLimited(mux, "POST "+prefix+"/test/webhook/send", rt.sendWebhook)
	handleLimited(mux, "POST "+prefix+"/test/reset", rt.handleReset)
}

func (rt *routes) registerResource(mux *http.ServeMux, prefix, resourceType, path string,
	fallback func(string) map[string]any, body map[string]any, required []string) {
	handleLimited(mux, "POST "+prefix+path, func(w http.ResponseWriter, r *http.Request) {
		if _, ok := rt.resolveScenario(w, r); !ok {
			return
		}
		rt.writeGenericResource(w, r, resourceType, body, required)
	})
	handleLimited(mux, "GET "+prefix+path, func(w http.ResponseWriter, r *http.Request) {
		if _, ok := rt.resolveScenario(w, r); !ok {
			return
		}
		rt.writeList(w, resourceType)
	})
	handleLimited(mux, "GET "+prefix+path+"/{id}", func(w http.ResponseWriter, r *http.Request) {
		if _, ok := rt.resolveScenario(w, r); !ok {
			return
		}
		rt.writeResource(w, resourceType, r.PathValue("id"), fallback)
	})
}

func handleLimited(mux *http.ServeMux, pattern string, h http.HandlerFunc) {
	mux.HandleFunc(pattern, withBodyLimit(h))
}

func withBodyLimit(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		httpx.LimitRequestBody(w, r)
		next(w, r)
	}
}

func (rt *routes) writeCheckoutSession(w http.ResponseWriter, r *http.Request) {
	scenario, ok := rt.resolveScenario(w, r)
	if !ok {
		return
	}
	switch scenario {
	case scenarioPaymentFailed:
		rt.writeStripeError(w, http.StatusPaymentRequired, "card_error", "card_declined", "Mockport simulated card decline")
	case scenarioAuthError:
		rt.writeStripeError(w, http.StatusUnauthorized, "invalid_request_error", "invalid_api_key", "Mockport simulated auth error")
	case scenarioRateLimited:
		rt.writeStripeError(w, http.StatusTooManyRequests, "rate_limit_error", "rate_limited", "Mockport simulated rate limit")
	case scenarioTimeout:
		rt.writeStripeError(w, http.StatusGatewayTimeout, "api_error", "mockport_timeout", "Mockport simulated timeout")
	default:
		fields := formFields(r)
		if !rt.validateFormFields(w, fields) {
			return
		}
		body := stripeDataFromStruct(checkoutSessionResponse{Object: "checkout.session", PaymentStatus: "paid"})
		if clientReferenceID := fields.Get("client_reference_id"); clientReferenceID != "" {
			body["client_reference_id"] = clientReferenceID
		}
		rt.createStatefulResource(w, r, "checkout_session", body)
	}
}

func (rt *routes) writePaymentIntent(w http.ResponseWriter, r *http.Request) {
	scenario, ok := rt.resolveScenario(w, r)
	if !ok {
		return
	}
	switch scenario {
	case scenarioPaymentFailed:
		rt.writeStripeError(w, http.StatusPaymentRequired, "card_error", "card_declined", "Mockport simulated card decline")
	case scenarioAuthError:
		rt.writeStripeError(w, http.StatusUnauthorized, "invalid_request_error", "invalid_api_key", "Mockport simulated auth error")
	case scenarioRateLimited:
		rt.writeStripeError(w, http.StatusTooManyRequests, "rate_limit_error", "rate_limited", "Mockport simulated rate limit")
	case scenarioTimeout:
		rt.writeStripeError(w, http.StatusGatewayTimeout, "api_error", "mockport_timeout", "Mockport simulated timeout")
	default:
		fields := formFields(r)
		if !rt.validateFormFields(w, fields) {
			return
		}
		if !fields.Empty() {
			if missing := missingFields(fields, "amount", "currency"); len(missing) > 0 {
				rt.writeValidationError(w, missing[0])
				return
			}
		}
		body := stripeDataFromStruct(paymentIntentResponse{Object: "payment_intent", Status: "succeeded"})
		if amount := fields.Get("amount"); amount != "" {
			if parsed, err := strconv.Atoi(amount); err == nil {
				body["amount"] = parsed
			}
		}
		if currency := fields.Get("currency"); currency != "" {
			body["currency"] = currency
		}
		rt.createStatefulResource(w, r, "payment_intent", body)
	}
}

func (rt *routes) writeGenericResource(w http.ResponseWriter, r *http.Request, resourceType string, body map[string]any, required []string) {
	// Clone the template map so registration-time closures do not share mutable state
	// across concurrent requests.
	body = maps.Clone(body)
	fields := formFields(r)
	if !rt.validateFormFields(w, fields) {
		return
	}
	if !fields.Empty() {
		if missing := missingFields(fields, required...); len(missing) > 0 {
			rt.writeValidationError(w, missing[0])
			return
		}
	}
	for _, field := range fields.Keys() {
		body[field] = parseFormValue(fields.Get(field))
	}
	rt.createStatefulResource(w, r, resourceType, body)
}

func (rt *routes) createStatefulResource(w http.ResponseWriter, r *http.Request, resourceType string, body map[string]any) {
	scope := "stripe:" + resourceType
	fingerprint := requestFingerprint(r)

	_, idempotentResponse, err := rt.idempotency.Do(scope, r.Header.Get("Idempotency-Key"), fingerprint, func() (state.IdempotentResponse, error) {
		resource, err := rt.store.Create("stripe", resourceType, body)
		if err != nil {
			return state.IdempotentResponse{}, err
		}
		response := resource.Data
		response["id"] = resource.ID
		return state.IdempotentResponse{Status: http.StatusOK, Body: response}, nil
	})
	if err != nil {
		var conflict *state.IdempotencyConflictError
		if errors.As(err, &conflict) {
			rt.writeStripeError(w, http.StatusConflict, "idempotency_error", "idempotency_key_in_use", "Keys for idempotent requests can only be reused with the same parameters")
			return
		}
		rt.writeStripeError(w, http.StatusInternalServerError, "api_error", "mockport_state_error", err.Error())
		return
	}
	rt.writeJSON(w, idempotentResponse.Status, idempotentResponse.Body)
}

func (rt *routes) writeResource(w http.ResponseWriter, resourceType, id string, fallback func(string) map[string]any) {
	if resource, ok := rt.store.Get("stripe", resourceType, id); ok {
		body := resource.Data
		body["id"] = resource.ID
		rt.writeJSON(w, http.StatusOK, body)
		return
	}
	if fallback != nil && looksLikeLegacyID(resourceType, id) {
		rt.writeJSON(w, http.StatusOK, fallback(id))
		return
	}
	rt.writeStripeError(w, http.StatusNotFound, "invalid_request_error", "resource_missing", "No such "+resourceType+": "+id)
}

func (rt *routes) writeList(w http.ResponseWriter, resourceType string) {
	var data []map[string]any
	for _, resource := range rt.store.List("stripe", resourceType) {
		body := resource.Data
		body["id"] = resource.ID
		data = append(data, body)
	}
	rt.writeJSON(w, http.StatusOK, listResponse{Object: "list", Data: data})
}

func fallbackCheckoutSession(id string) map[string]any {
	return map[string]any{
		"id":             id,
		"object":         "checkout.session",
		"payment_status": "paid",
	}
}

func fallbackPaymentIntent(id string) map[string]any {
	return map[string]any{
		"id":     id,
		"object": "payment_intent",
		"status": "succeeded",
	}
}

func formFields(r *http.Request) formValues {
	if err := r.ParseForm(); err != nil {
		return emptyValues{err: err}
	}
	return urlValues{values: r.Form}
}

func stripeDataFromStruct(value any) map[string]any {
	encoded, _ := json.Marshal(value)
	var decoded map[string]any
	_ = json.Unmarshal(encoded, &decoded)
	return decoded
}

type formValues interface {
	Get(string) string
	Keys() []string
	Empty() bool
	Err() error
}

type urlValues struct {
	values map[string][]string
}

func (v urlValues) Get(name string) string {
	values := v.values[name]
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func (v urlValues) Keys() []string {
	keys := make([]string, 0, len(v.values))
	for key := range v.values {
		keys = append(keys, key)
	}
	return keys
}

func (v urlValues) Empty() bool { return len(v.values) == 0 }
func (v urlValues) Err() error  { return nil }

type emptyValues struct{ err error }

func (emptyValues) Get(string) string { return "" }
func (emptyValues) Keys() []string    { return nil }
func (emptyValues) Empty() bool       { return true }
func (v emptyValues) Err() error      { return v.err }

func (rt *routes) validateFormFields(w http.ResponseWriter, fields formValues) bool {
	if fields.Err() == nil {
		return true
	}
	if httpx.IsRequestBodyTooLarge(fields.Err()) {
		rt.writeStripeError(w, http.StatusRequestEntityTooLarge, "invalid_request_error", "request_too_large", "Request body is too large")
		return false
	}
	rt.writeStripeError(w, http.StatusBadRequest, "invalid_request_error", "invalid_request", "Request body is invalid")
	return false
}

func requestFingerprint(r *http.Request) string {
	_ = r.ParseForm()
	return r.Method + " " + r.URL.Path + " " + r.Form.Encode()
}

func (rt *routes) writeStripeError(w http.ResponseWriter, status int, typ, code, message string) {
	rt.writeJSON(w, status, errorBody{Error: stripeError{Type: typ, Code: code, Message: message}})
}

func (rt *routes) writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Request-Id", "req_mockport")
	w.Header().Set("Stripe-Version", "2025-10-29.clover")
	httpx.WriteJSON(w, status, body)
}

func (rt *routes) writeValidationError(w http.ResponseWriter, field string) {
	rt.writeJSON(w, http.StatusBadRequest, errorBody{Error: stripeError{
		Type:    "invalid_request_error",
		Code:    "parameter_missing",
		Param:   field,
		Message: "Missing required param: " + field,
	}})
}

func missingFields(fields formValues, required ...string) []string {
	var missing []string
	for _, field := range required {
		if fields.Get(field) == "" {
			missing = append(missing, field)
		}
	}
	return missing
}

func parseFormValue(value string) any {
	if parsed, err := strconv.Atoi(value); err == nil {
		return parsed
	}
	return value
}

// resolveScenario はリクエストヘッダまたは設定からシナリオ名を解決する。
// 未知のシナリオが指定された場合は Stripe エラー形式で 400 を返し false を返す。
func (rt *routes) resolveScenario(w http.ResponseWriter, r *http.Request) (string, bool) {
	scenario, err := rt.resolver.Resolve(r)
	if err != nil {
		rt.writeStripeError(w, http.StatusBadRequest, "invalid_request_error", "unknown_mockport_scenario", err.Error())
		return "", false
	}
	return scenario, true
}

func looksLikeLegacyID(resourceType, id string) bool {
	switch resourceType {
	case "checkout_session":
		return strings.HasPrefix(id, "cs_")
	case "payment_intent":
		return strings.HasPrefix(id, "pi_")
	default:
		return false
	}
}
