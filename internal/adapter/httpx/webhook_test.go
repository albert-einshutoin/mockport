package httpx

import (
	"context"
	"errors"
	"net"
	"net/http"
	"testing"
)

func TestIsWebhookSendTimeout(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "deadline", err: context.DeadlineExceeded, want: true},
		{name: "wrapped deadline", err: errors.Join(errors.New("send failed"), context.DeadlineExceeded), want: true},
		{name: "network timeout", err: &net.DNSError{IsTimeout: true}, want: true},
		{name: "other", err: errors.New("connection refused"), want: false},
		{name: "nil", err: nil, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsWebhookSendTimeout(tt.err); got != tt.want {
				t.Fatalf("IsWebhookSendTimeout(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

func TestIsWebhookTargetSuccess(t *testing.T) {
	for _, tt := range []struct {
		status int
		want   bool
	}{
		{status: http.StatusOK, want: true},
		{status: http.StatusAccepted, want: true},
		{status: http.StatusNoContent, want: true},
		{status: http.StatusMultipleChoices, want: false},
		{status: http.StatusBadGateway, want: false},
	} {
		if got := IsWebhookTargetSuccess(tt.status); got != tt.want {
			t.Fatalf("IsWebhookTargetSuccess(%d) = %v, want %v", tt.status, got, tt.want)
		}
	}
}
