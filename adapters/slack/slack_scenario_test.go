package slack_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	slackadapter "github.com/albert-einshutoin/mockport/adapters/slack"
	"github.com/albert-einshutoin/mockport/internal/adapter"
)

func newSlackScenarioMux(t *testing.T, cfg adapter.Config) *http.ServeMux {
	t.Helper()
	mux := http.NewServeMux()
	if err := slackadapter.New().Register(mux, cfg); err != nil {
		t.Fatalf("Register: %v", err)
	}
	return mux
}

func TestSlackHeaderOverridesConfig(t *testing.T) {
	// config=message_success でサーバーを起動し、ヘッダで auth_error を指定
	mux := newSlackScenarioMux(t, adapter.Config{BasePath: "/slack", Scenario: "message_success"})
	form := url.Values{"channel": {"C_MOCKPORT"}, "text": {"hello"}}.Encode()
	req := httptest.NewRequest(http.MethodPost, "/slack/api/chat.postMessage", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Mockport-Scenario", "auth_error")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("want 200 (Slack always returns 200 for auth errors), got %d", rec.Code)
	}
	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["ok"] != false {
		t.Errorf("want ok=false for auth_error, got %v", body["ok"])
	}
	errStr, _ := body["error"].(string)
	if errStr != "invalid_auth" {
		t.Errorf("want error=invalid_auth, got %q", errStr)
	}
}

func TestSlackUnknownScenarioReturns400(t *testing.T) {
	mux := newSlackScenarioMux(t, adapter.Config{BasePath: "/slack", Scenario: "message_success"})
	form := url.Values{"channel": {"C_MOCKPORT"}, "text": {"hello"}}.Encode()
	req := httptest.NewRequest(http.MethodPost, "/slack/api/chat.postMessage", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Mockport-Scenario", "no_such_scenario")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", rec.Code)
	}
	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	errStr, _ := body["error"].(string)
	if errStr != "unknown_mockport_scenario" {
		t.Errorf("want error=unknown_mockport_scenario, got %q", errStr)
	}
}

func TestSlackEventsUnknownScenarioReturns400(t *testing.T) {
	// /events も dispatch 前に未知シナリオを 400 で拒否する（署名検証より前で早期リターン）
	mux := newSlackScenarioMux(t, adapter.Config{BasePath: "/slack", Scenario: "message_success"})
	req := httptest.NewRequest(http.MethodPost, "/slack/events", strings.NewReader(`{"type":"url_verification","challenge":"abc"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Mockport-Scenario", "no_such_scenario")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", rec.Code)
	}
	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	errStr, _ := body["error"].(string)
	if errStr != "unknown_mockport_scenario" {
		t.Errorf("want error=unknown_mockport_scenario, got %q", errStr)
	}
}
