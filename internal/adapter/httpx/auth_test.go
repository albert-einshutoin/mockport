package httpx

import (
	"net/http/httptest"
	"testing"
)

func TestRequireBearerAuth(t *testing.T) {
	for _, tc := range []struct {
		name   string
		values []string
		want   string
		ok     bool
	}{
		{"correct", []string{"Bearer mockport_key"}, "mockport_key", true},
		{"wrong", []string{"Bearer mockport_wrong"}, "mockport_key", false},
		{"missing", nil, "mockport_key", false},
		{"empty configured key", []string{"Bearer "}, "", false},
		{"malformed", []string{"Basic mockport_key"}, "mockport_key", false},
		{"duplicate", []string{"Bearer mockport_key", "Bearer mockport_key"}, "mockport_key", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			for _, value := range tc.values {
				req.Header.Add("Authorization", value)
			}
			if got := RequireBearerAuth(req, tc.want); got != tc.ok {
				t.Fatalf("RequireBearerAuth = %v, want %v", got, tc.ok)
			}
		})
	}
}
