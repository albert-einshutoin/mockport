package line

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/albert-einshutoin/mockport/internal/adapter/httpx"
	"github.com/albert-einshutoin/mockport/internal/security"
)

func (r *routes) writeSetWebhookEndpoint(w http.ResponseWriter, req *http.Request) {
	payload, err := decodePayload(req)
	if err != nil {
		writeDecodeError(w, err)
		return
	}
	endpoint, _ := payload["endpoint"].(string)
	if !strings.HasPrefix(endpoint, "https://") || len(endpoint) > 500 {
		writeLINEError(w, http.StatusBadRequest, "Invalid webhook endpoint URL")
		return
	}
	r.setWebhookEndpoint(endpoint)
	writeEmptyJSON(w, http.StatusOK)
}

func (r *routes) writeGetWebhookEndpoint(w http.ResponseWriter) {
	endpoint := r.getWebhookEndpoint()
	if endpoint == "" {
		writeLINEError(w, http.StatusNotFound, "Webhook endpoint not found")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"endpoint": endpoint, "active": true})
}

func (r *routes) writeWebhookTest(w http.ResponseWriter) {
	if normalizeScenario(r.cfg.Scenario) == "auth_error" {
		writeLINEError(w, http.StatusUnauthorized, "Mockport simulated invalid channel access token")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"success": true, "timestamp": "2999-01-01T00:00:00Z", "statusCode": 200, "reason": "OK", "detail": "200"})
}

func (r *routes) sendWebhook(w http.ResponseWriter, req *http.Request) {
	if !security.IsLoopbackRemoteAddr(req.RemoteAddr) {
		writeLINEError(w, http.StatusForbidden, "webhook delivery can only be triggered from loopback")
		return
	}
	if r.cfg.WebhookTargetURL == "" {
		writeLINEError(w, http.StatusBadRequest, "webhook target URL is not configured")
		return
	}
	if !security.IsSafeWebhookTargetURL(r.cfg.WebhookTargetURL) {
		writeLINEError(w, http.StatusBadRequest, "webhook target URL must be a local Mockport target")
		return
	}
	payload, err := decodePayload(req)
	if err != nil {
		writeDecodeError(w, err)
		return
	}
	destination, _ := payload["destination"].(string)
	if destination == "" {
		destination = "U00000000000000000000000000000000"
	}
	events, _ := payload["events"].([]any)
	if len(events) == 0 {
		events = []any{map[string]any{
			"type":       "message",
			"replyToken": "mockport_line_reply_token",
			"message": map[string]any{
				"type": "text",
				"id":   "mockport_line_message",
				"text": "Mockport LINE webhook message",
			},
		}}
	}
	for i, event := range events {
		eventObject, ok := event.(map[string]any)
		if !ok {
			writeValidationDetails(w, []lineErrorDetail{{Message: "Must be a JSON object", Property: fmt.Sprintf("events[%d]", i)}})
			return
		}
		if eventObject["timestamp"] == nil {
			eventObject["timestamp"] = int64(4102444800000)
		}
		if eventObject["source"] == nil {
			eventObject["source"] = map[string]any{"type": "user", "userId": "Umockport"}
		}
		if eventObject["mode"] == nil {
			eventObject["mode"] = "active"
		}
		if eventObject["webhookEventId"] == nil {
			eventObject["webhookEventId"] = fmt.Sprintf("01MOCKPORTLINEEVENT%02d", i+1)
		}
		if eventObject["deliveryContext"] == nil {
			eventObject["deliveryContext"] = map[string]any{"isRedelivery": false}
		}
	}
	body, err := json.Marshal(map[string]any{"destination": destination, "events": events})
	if err != nil {
		writeLINEError(w, http.StatusInternalServerError, "failed to encode webhook payload")
		return
	}
	outbound, err := http.NewRequestWithContext(req.Context(), http.MethodPost, r.cfg.WebhookTargetURL, bytes.NewReader(body))
	if err != nil {
		writeLINEError(w, http.StatusBadRequest, "webhook target URL is invalid")
		return
	}
	secret := r.cfg.WebhookSigningSecret
	if secret == "" {
		secret = "mockport_line_secret"
	}
	outbound.Header.Set("Content-Type", "application/json")
	outbound.Header.Set("x-line-signature", signWebhookPayload(secret, body))
	resp, err := (&http.Client{Timeout: 5 * time.Second}).Do(outbound)
	if err != nil {
		writeLINEError(w, http.StatusBadGateway, "failed to send webhook")
		return
	}
	defer resp.Body.Close()
	httpx.WriteJSON(w, http.StatusAccepted, map[string]any{
		"sent":        true,
		"target_url":  r.cfg.WebhookTargetURL,
		"event_count": len(events),
		"status_code": resp.StatusCode,
	})
}

func signWebhookPayload(secret string, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(payload)
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}
