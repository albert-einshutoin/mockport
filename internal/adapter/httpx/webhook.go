package httpx

import (
	"context"
	"errors"
	"net"
	"net/http"
	"testing"
	"time"
)

// DefaultWebhookSenderTimeout bounds outbound webhook helper delivery so local
// automation does not hang when a configured target is down or slow.
const DefaultWebhookSenderTimeout = 5 * time.Second

var webhookSenderTimeout = DefaultWebhookSenderTimeout

// WebhookSenderClient returns an HTTP client for adapter webhook send helpers.
func WebhookSenderClient() *http.Client {
	return &http.Client{Timeout: webhookSenderTimeout}
}

// WebhookSendFailure classifies outbound webhook delivery failures.
type WebhookSendFailure struct {
	Kind string
}

// IsWebhookTargetSuccess reports whether the target returned a 2xx status.
func IsWebhookTargetSuccess(statusCode int) bool {
	return statusCode >= http.StatusOK && statusCode < http.StatusMultipleChoices
}

// ClassifyWebhookSendError maps outbound delivery errors to timeout vs other failures.
func ClassifyWebhookSendError(err error) (WebhookSendFailure, bool) {
	if err == nil {
		return WebhookSendFailure{}, false
	}
	if isWebhookSendTimeout(err) {
		return WebhookSendFailure{Kind: "timeout"}, true
	}
	return WebhookSendFailure{Kind: "send"}, true
}

func isWebhookSendTimeout(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}

// SetWebhookSenderTimeoutForTest overrides the outbound webhook timeout for t's duration.
func SetWebhookSenderTimeoutForTest(t testing.TB, timeout time.Duration) {
	t.Helper()
	previous := webhookSenderTimeout
	webhookSenderTimeout = timeout
	t.Cleanup(func() {
		webhookSenderTimeout = previous
	})
}
