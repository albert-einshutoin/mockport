package stripe

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/albert-einshutoin/mockport/internal/security"
)

func (rt *routes) sendWebhook(w http.ResponseWriter, r *http.Request) {
	if !security.IsLoopbackRemoteAddr(r.RemoteAddr) {
		rt.writeStripeError(w, http.StatusForbidden, "invalid_request_error", "local_request_required", "webhook delivery can only be triggered from loopback")
		return
	}
	if rt.cfg.WebhookTargetURL == "" {
		rt.writeStripeError(w, http.StatusBadRequest, "invalid_request_error", "missing_webhook_target", "webhook target URL is not configured")
		return
	}
	if !security.IsSafeWebhookTargetURL(rt.cfg.WebhookTargetURL) {
		rt.writeStripeError(w, http.StatusBadRequest, "invalid_request_error", "unsafe_webhook_target", "webhook target URL must be a local Mockport target")
		return
	}
	secret := rt.cfg.WebhookSigningSecret
	if secret == "" {
		secret = "whsec_mockport"
	}
	scenario, err := rt.resolver.Resolve(r)
	if err != nil {
		rt.writeStripeError(w, http.StatusBadRequest, "invalid_request_error", "unknown_mockport_scenario", err.Error())
		return
	}
	eventType := "checkout.session.completed"
	if scenario == scenarioPaymentFailed {
		eventType = "payment_intent.payment_failed"
	}
	payload, err := json.Marshal(map[string]any{
		"id":   "evt_mockport",
		"type": eventType,
		"data": map[string]any{
			"object": map[string]any{
				"id":     "cs_test_mockport",
				"object": "checkout.session",
			},
		},
	})
	if err != nil {
		rt.writeStripeError(w, http.StatusInternalServerError, "api_error", "webhook_encode_failed", "failed to encode webhook payload")
		return
	}

	req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, rt.cfg.WebhookTargetURL, bytes.NewReader(payload))
	if err != nil {
		rt.writeStripeError(w, http.StatusBadRequest, "invalid_request_error", "invalid_webhook_target", "webhook target URL is invalid")
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Stripe-Signature", signPayload(secret, nowUnix(), payload))

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		rt.writeStripeError(w, http.StatusBadGateway, "api_error", "webhook_send_failed", "failed to send webhook")
		return
	}
	defer resp.Body.Close()

	rt.writeJSON(w, http.StatusAccepted, map[string]any{
		"sent":        true,
		"target_url":  rt.cfg.WebhookTargetURL,
		"event_type":  eventType,
		"status_code": resp.StatusCode,
	})
}

func (rt *routes) handleReset(w http.ResponseWriter, r *http.Request) {
	if !security.IsLoopbackRemoteAddr(r.RemoteAddr) {
		rt.writeStripeError(w, http.StatusForbidden, "invalid_request_error", "local_request_required", "state reset can only be triggered from loopback")
		return
	}
	resourceTypes := rt.store.ResetAll("stripe")
	rt.idempotency.ResetAll()
	rt.writeJSON(w, http.StatusOK, map[string]any{
		"reset":          true,
		"adapter":        "stripe",
		"resource_types": resourceTypes,
	})
}
