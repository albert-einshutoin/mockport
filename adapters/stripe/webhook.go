package stripe

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/albert-einshutoin/mockport/internal/adapter/httpx"
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
	eventID := "evt_mockport"
	object := map[string]any{"id": "cs_test_mockport", "object": "checkout.session"}
	if r.ContentLength != 0 {
		var input struct {
			SessionID string `json:"session_id"`
			EventID   string `json:"event_id"`
			EventType string `json:"event_type"`
		}
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&input); err != nil {
			rt.writeStripeError(w, http.StatusBadRequest, "invalid_request_error", "invalid_webhook_event", "webhook event must be a JSON object with session_id, event_id, and event_type")
			return
		}
		if err := decoder.Decode(new(any)); err != io.EOF || input.SessionID == "" || !validWebhookEventID(input.EventID) ||
			(input.EventType != "checkout.session.completed" && input.EventType != "checkout.session.async_payment_failed") {
			rt.writeStripeError(w, http.StatusBadRequest, "invalid_request_error", "invalid_webhook_event", "webhook event requires an existing session, an event ID, and a supported event type")
			return
		}
		resource, ok := rt.store.Get("stripe", "checkout_session", input.SessionID)
		if !ok {
			rt.writeStripeError(w, http.StatusNotFound, "invalid_request_error", "checkout_session_not_found", "webhook session was not created in this Mockport process")
			return
		}
		object = resource.Data
		object["id"] = resource.ID
		if input.EventType == "checkout.session.async_payment_failed" {
			object["payment_status"] = "unpaid"
		} else {
			object["payment_status"] = "paid"
		}
		eventID, eventType = input.EventID, input.EventType
	}
	payload, err := json.Marshal(map[string]any{
		"id":   eventID,
		"type": eventType,
		"data": map[string]any{
			"object": object,
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

	resp, err := rt.webhookClient.Do(req)
	if err != nil {
		if httpx.IsWebhookSendTimeout(err) {
			rt.writeStripeError(w, http.StatusGatewayTimeout, "api_error", "webhook_send_timeout", "webhook target did not respond before timeout")
			return
		}
		rt.writeStripeError(w, http.StatusBadGateway, "api_error", "webhook_send_failed", "failed to send webhook")
		return
	}
	defer resp.Body.Close()
	if !httpx.IsWebhookTargetSuccess(resp.StatusCode) {
		rt.writeStripeError(w, http.StatusBadGateway, "api_error", "webhook_target_non_2xx", fmt.Sprintf("webhook target returned status %d", resp.StatusCode))
		return
	}

	rt.writeJSON(w, http.StatusAccepted, map[string]any{
		"sent":        true,
		"target_url":  rt.cfg.WebhookTargetURL,
		"event_type":  eventType,
		"event_id":    eventID,
		"status_code": resp.StatusCode,
	})
}

func validWebhookEventID(id string) bool {
	if len(id) < 5 || len(id) > 128 || !strings.HasPrefix(id, "evt_") {
		return false
	}
	for _, ch := range id[4:] {
		if (ch < 'a' || ch > 'z') && (ch < 'A' || ch > 'Z') && (ch < '0' || ch > '9') && ch != '_' && ch != '-' {
			return false
		}
	}
	return true
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
