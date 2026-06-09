package cli

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthcheckCommandPassesForHealthyEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			t.Fatalf("path = %q, want /health", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	cmd := NewRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"healthcheck", "--url", server.URL + "/health"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute healthcheck: %v", err)
	}
	if got := out.String(); !strings.Contains(got, "mockport healthcheck: ok") {
		t.Fatalf("output = %q, want success message", got)
	}
}

func TestHealthcheckCommandRejectsNon200Status(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "busy", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	cmd := NewRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"healthcheck", "--url", server.URL + "/health"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("execute healthcheck returned nil, want error")
	}
	if !strings.Contains(err.Error(), "healthcheck status: 503") {
		t.Fatalf("error = %q, want healthcheck status: 503", err)
	}
}

func TestHealthcheckCommandRejectsUnexpectedBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "degraded"})
	}))
	defer server.Close()

	cmd := NewRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"healthcheck", "--url", server.URL + "/health"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("execute healthcheck returned nil, want error")
	}
	if !strings.Contains(err.Error(), "healthcheck response") {
		t.Fatalf("error = %q, want healthcheck response", err)
	}
}
