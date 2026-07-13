package httpx

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"
)

// DefaultWebhookSenderTimeout bounds local webhook delivery so a slow or
// unreachable application cannot stall CI and development automation.
const DefaultWebhookSenderTimeout = 5 * time.Second

// NewWebhookSenderClient creates the dedicated client used by webhook helpers.
// Keeping it separate from http.DefaultClient prevents unrelated transports or
// timeout policy from leaking into this local-only integration boundary.
func NewWebhookSenderClient(timeout time.Duration) *http.Client {
	return &http.Client{Timeout: timeout}
}

// IsWebhookSendTimeout reports whether delivery failed because its deadline
// elapsed, including wrapped context and network timeout errors.
func IsWebhookSendTimeout(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}

// IsWebhookTargetSuccess reports whether the target accepted the delivery.
func IsWebhookTargetSuccess(statusCode int) bool {
	return statusCode >= http.StatusOK && statusCode < http.StatusMultipleChoices
}
