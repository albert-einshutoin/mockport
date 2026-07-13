package httpx

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestIsWebhookTargetSuccess(t *testing.T) {
	tests := []struct {
		code int
		want bool
	}{
		{http.StatusOK, true},
		{http.StatusAccepted, true},
		{http.StatusNoContent, true},
		{http.StatusBadRequest, false},
		{http.StatusInternalServerError, false},
	}
	for _, tt := range tests {
		if got := IsWebhookTargetSuccess(tt.code); got != tt.want {
			t.Fatalf("IsWebhookTargetSuccess(%d) = %v, want %v", tt.code, got, tt.want)
		}
	}
}

func TestClassifyWebhookSendError(t *testing.T) {
	timeoutErr := &net.DNSError{IsTimeout: true, Err: "i/o timeout", Name: "example.test"}
	failure, ok := ClassifyWebhookSendError(timeoutErr)
	if !ok || failure.Kind != "timeout" {
		t.Fatalf("timeout classification = %#v, %v, want timeout", failure, ok)
	}

	deadlineErr := context.DeadlineExceeded
	failure, ok = ClassifyWebhookSendError(deadlineErr)
	if !ok || failure.Kind != "timeout" {
		t.Fatalf("deadline classification = %#v, %v, want timeout", failure, ok)
	}

	otherErr := errors.New("connection refused")
	failure, ok = ClassifyWebhookSendError(otherErr)
	if !ok || failure.Kind != "send" {
		t.Fatalf("generic classification = %#v, %v, want send", failure, ok)
	}
}

func TestWebhookSenderClientTimesOut(t *testing.T) {
	SetWebhookSenderTimeoutForTest(t, 25*time.Millisecond)
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer target.Close()

	_, err := WebhookSenderClient().Get(target.URL)
	failure, ok := ClassifyWebhookSendError(err)
	if !ok || failure.Kind != "timeout" {
		t.Fatalf("client error = %v, classification = %#v, %v, want timeout", err, failure, ok)
	}
}
