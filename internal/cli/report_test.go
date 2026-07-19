package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestReportCommandPrintsTextSummary(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/_mockport/report" {
			t.Fatalf("path = %q, want /_mockport/report", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"mode":"ai-safe","adapters":[{"name":"stripe","base_path":"/stripe","enabled":true}],"requests":[{"method":"POST","path":"/stripe/v1/checkout/sessions","status":200}],"safety_warnings":[]}`))
	}))
	defer server.Close()

	cmd, out := newTestCommand(t, "report", "--url", server.URL+"/_mockport/report")
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute report: %v", err)
	}

	got := out.String()
	for _, want := range []string{
		"Mockport Report",
		"Mode: ai-safe",
		"stripe enabled at /stripe",
		"POST /stripe/v1/checkout/sessions -> 200",
		"Safety warnings: 0",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("report output missing %q:\n%s", want, got)
		}
	}
}

func TestReportCommandPrintsJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"mode":"ai-safe","safety":{"mode":"ai-safe","safe":false,"real_looking_secrets":1,"external_urls":1,"public_env_safe":false},"adapters":[{"name":"stripe","base_path":"/stripe","enabled":true}],"requests":[{"method":"POST","path":"/stripe/v1/checkout/sessions","status":200}],"safety_warnings":[{"field":"stripe.fake_secret","category":"real_looking_secret","message":"real-looking secret detected"}],"unexpected_secret":"sk_live_secret_should_not_print","unexpected_live_url":"https://api.stripe.com"}`))
	}))
	defer server.Close()

	cmd, out := newTestCommand(t, "report", "--url", server.URL, "--json")
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute report: %v", err)
	}
	got := out.Bytes()
	if !json.Valid(got) {
		t.Fatalf("report output is not valid JSON:\n%s", got)
	}
	var payload map[string]any
	if err := json.Unmarshal(got, &payload); err != nil {
		t.Fatalf("decode JSON report: %v", err)
	}
	safety, ok := payload["safety"].(map[string]any)
	if !ok || safety["public_env_safe"] != false || safety["real_looking_secrets"] != float64(1) || safety["external_urls"] != float64(1) {
		t.Fatalf("safety summary = %#v", payload["safety"])
	}
	if _, ok := payload["adapters"].([]any); !ok {
		t.Fatalf("adapters = %#v", payload["adapters"])
	}
	if _, ok := payload["requests"].([]any); !ok {
		t.Fatalf("requests = %#v", payload["requests"])
	}
	for _, forbidden := range []string{"\x1b[", "sk_live_secret_should_not_print", "https://api.stripe.com"} {
		if strings.Contains(string(got), forbidden) {
			t.Fatalf("JSON report contains forbidden value %q:\n%s", forbidden, got)
		}
	}
}

func TestReportCommandFormatJSONRemainsSupported(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"mode":"ai-safe","safety":{"public_env_safe":true},"adapters":[],"requests":[]}`))
	}))
	defer server.Close()

	cmd, out := newTestCommand(t, "report", "--url", server.URL, "--format", "json")
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute report: %v", err)
	}
	if !json.Valid(out.Bytes()) {
		t.Fatalf("--format json output is not valid JSON:\n%s", out.String())
	}
}

func TestReportCommandRejectsConflictingJSONFormat(t *testing.T) {
	cmd, _ := newTestCommand(t, "report", "--json", "--format", "text")

	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), `--json cannot be combined with --format "text"`) {
		t.Fatalf("error = %v, want conflicting format error", err)
	}
}

func TestReportHelpDocumentsJSONFlag(t *testing.T) {
	cmd, out := newTestCommand(t, "report", "--help")
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute report help: %v", err)
	}
	if !strings.Contains(out.String(), "--json") || !strings.Contains(out.String(), "shorthand for --format json") {
		t.Fatalf("report help missing JSON flag:\n%s", out.String())
	}
}
