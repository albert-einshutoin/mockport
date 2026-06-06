package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/albert-einshutoin/mockport/adapters/githuboauth"
	"github.com/albert-einshutoin/mockport/adapters/line"
	"github.com/albert-einshutoin/mockport/adapters/openai"
	"github.com/albert-einshutoin/mockport/adapters/slack"
	"github.com/albert-einshutoin/mockport/adapters/stripe"
	"github.com/albert-einshutoin/mockport/internal/adapter"
	"github.com/albert-einshutoin/mockport/internal/config"
	"github.com/albert-einshutoin/mockport/internal/report"
)

func TestConfiguredHandlerServesMultipleAdapters(t *testing.T) {
	cfg := config.Config{
		Mode:   "ai-safe",
		Server: config.ServerConfig{Host: "127.0.0.1", Port: 43101},
		Adapters: map[string]config.AdapterConfig{
			"stripe":       {Enabled: true, BasePath: "/stripe", Scenario: "payment_success", FakeSecret: "mockport_stripe_secret"},
			"openai":       {Enabled: true, BasePath: "/openai", Scenario: "chat_success", FakeSecret: "mockport_openai_key"},
			"github-oauth": {Enabled: true, BasePath: "/github", Scenario: "oauth_success", FakeSecret: "mockport_github_secret"},
			"slack":        {Enabled: true, BasePath: "/slack", Scenario: "message_success", FakeSecret: "mockport_slack_token"},
			"line":         {Enabled: true, BasePath: "/line", Scenario: "line_success", FakeSecret: "mockport_line_channel_token"},
		},
	}
	if err := config.Validate(&cfg); err != nil {
		t.Fatalf("validate config: %v", err)
	}
	reg := adapter.NewRegistry()
	registerTestAdapters(t, reg)
	handler, err := NewConfiguredHandler(cfg, reg, report.NewRecorder())
	if err != nil {
		t.Fatalf("new handler: %v", err)
	}

	for _, tc := range []struct {
		method string
		path   string
		status int
	}{
		{http.MethodPost, "/stripe/v1/checkout/sessions", http.StatusOK},
		{http.MethodGet, "/openai/v1/models", http.StatusOK},
		{http.MethodGet, "/github/login/oauth/authorize?client_id=mockport_github_client", http.StatusFound},
		{http.MethodPost, "/slack/api/auth.test", http.StatusOK},
		{http.MethodPost, "/line/v2/bot/message/push", http.StatusOK},
	} {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(tc.method, tc.path, nil))
		if rec.Code != tc.status {
			t.Fatalf("%s %s status = %d, want %d", tc.method, tc.path, rec.Code, tc.status)
		}
	}

	reportRec := httptest.NewRecorder()
	handler.ServeHTTP(reportRec, httptest.NewRequest(http.MethodGet, "/_mockport/report", nil))
	var snapshot report.Snapshot
	if err := json.Unmarshal(reportRec.Body.Bytes(), &snapshot); err != nil {
		t.Fatalf("decode report: %v", err)
	}
	if len(snapshot.Adapters) != 5 {
		t.Fatalf("adapter count = %d, want 5: %#v", len(snapshot.Adapters), snapshot.Adapters)
	}
	if len(snapshot.ScenarioCoverage) != 5 {
		t.Fatalf("coverage count = %d, want 5", len(snapshot.ScenarioCoverage))
	}
}

func TestConfiguredHandlerReportOrderIsDeterministic(t *testing.T) {
	cfg := config.Config{
		Mode:   "ai-safe",
		Server: config.ServerConfig{Host: "127.0.0.1", Port: 43101},
		Adapters: map[string]config.AdapterConfig{
			"stripe":       {Enabled: true, BasePath: "/stripe", Scenario: "payment_success", FakeSecret: "mockport_stripe_secret"},
			"openai":       {Enabled: true, BasePath: "/openai", Scenario: "chat_success", FakeSecret: "mockport_openai_key"},
			"github-oauth": {Enabled: true, BasePath: "/github", Scenario: "oauth_success", FakeSecret: "mockport_github_secret"},
			"slack":        {Enabled: true, BasePath: "/slack", Scenario: "message_success", FakeSecret: "mockport_slack_token"},
			"line":         {Enabled: true, BasePath: "/line", Scenario: "line_success", FakeSecret: "mockport_line_channel_token"},
		},
	}
	if err := config.Validate(&cfg); err != nil {
		t.Fatalf("validate config: %v", err)
	}
	want := []string{"github-oauth", "line", "openai", "slack", "stripe"}
	for i := 0; i < 25; i++ {
		reg := adapter.NewRegistry()
		registerTestAdapters(t, reg)
		handler, err := NewConfiguredHandler(cfg, reg, report.NewRecorder())
		if err != nil {
			t.Fatalf("new handler: %v", err)
		}

		reportRec := httptest.NewRecorder()
		handler.ServeHTTP(reportRec, httptest.NewRequest(http.MethodGet, "/_mockport/report", nil))
		var snapshot report.Snapshot
		if err := json.Unmarshal(reportRec.Body.Bytes(), &snapshot); err != nil {
			t.Fatalf("decode report: %v", err)
		}
		got := make([]string, 0, len(snapshot.Adapters))
		for _, status := range snapshot.Adapters {
			got = append(got, status.Name)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("adapter order = %#v, want %#v", got, want)
		}
	}
}

func registerTestAdapters(t *testing.T, reg *adapter.Registry) {
	t.Helper()
	for _, adapterImpl := range []adapter.Adapter{stripe.New(), openai.New(), githuboauth.New(), slack.New(), line.New()} {
		if err := reg.Register(adapterImpl); err != nil {
			t.Fatalf("register %s: %v", adapterImpl.Name(), err)
		}
	}
}
