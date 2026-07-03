package slack

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/albert-einshutoin/mockport/internal/adapter"
	"github.com/albert-einshutoin/mockport/internal/adapter/adaptertest"
)

func TestAuthTest(t *testing.T) {
	rec := performRequest(t, adapter.Config{BasePath: "/slack", Scenario: "message_success"}, http.MethodPost, "/slack/api/auth.test")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode auth.test: %v", err)
	}
	if body["ok"] != true {
		t.Fatalf("ok = %v, want true", body["ok"])
	}
	if body["team_id"] != "T_MOCKPORT" || body["bot_id"] != "B_MOCKPORT" {
		t.Fatalf("auth body = %#v", body)
	}
}

func TestPostMessageScenarios(t *testing.T) {
	tests := []struct {
		name     string
		scenario string
		status   int
		ok       bool
	}{
		{"success", "message_success", http.StatusOK, true},
		{"auth", "auth_error", http.StatusOK, false},
		{"rate", "rate_limited", http.StatusTooManyRequests, false},
		{"delivery", "delivery_failed", http.StatusOK, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := performRequest(t, adapter.Config{BasePath: "/slack", Scenario: tt.scenario}, http.MethodPost, "/slack/api/chat.postMessage")
			if rec.Code != tt.status {
				t.Fatalf("status = %d, want %d", rec.Code, tt.status)
			}
		})
	}
}

func TestPostMessagePersistsConversationHistory(t *testing.T) {
	mux := newSlackMux(t, adapter.Config{BasePath: "/slack", Scenario: "message_success"})

	post := serveSlackRequest(mux, http.MethodPost, "/slack/api/chat.postMessage", "channel=C_TEST&text=hello")
	if post.Code != http.StatusOK {
		t.Fatalf("post status = %d, want %d, body=%s", post.Code, http.StatusOK, post.Body.String())
	}
	var posted map[string]any
	if err := json.Unmarshal(post.Body.Bytes(), &posted); err != nil {
		t.Fatalf("decode post: %v", err)
	}
	if posted["channel"] != "C_TEST" || posted["ts"] != "slack_message_000001" {
		t.Fatalf("posted = %#v", posted)
	}
	message, ok := posted["message"].(map[string]any)
	if !ok || message["text"] != "hello" || message["team"] != "T_MOCKPORT" {
		t.Fatalf("posted message = %#v", posted["message"])
	}

	history := serveSlackRequest(mux, http.MethodGet, "/slack/api/conversations.history?channel=C_TEST", "")
	if history.Code != http.StatusOK {
		t.Fatalf("history status = %d, want %d, body=%s", history.Code, http.StatusOK, history.Body.String())
	}
	var body struct {
		OK       bool             `json:"ok"`
		Messages []map[string]any `json:"messages"`
	}
	if err := json.Unmarshal(history.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode history: %v", err)
	}
	if !body.OK || len(body.Messages) != 1 || body.Messages[0]["text"] != "hello" {
		t.Fatalf("history = %#v", body)
	}
}

func TestMessageLifecycleHistoryVisibility(t *testing.T) {
	mux := newSlackMux(t, adapter.Config{BasePath: "/slack", Scenario: "message_success"})
	const channel = "C_TEST"

	post := serveSlackRequest(mux, http.MethodPost, "/slack/api/chat.postMessage", "channel="+channel+"&text=hello")
	if post.Code != http.StatusOK {
		t.Fatalf("post status = %d, want %d, body=%s", post.Code, http.StatusOK, post.Body.String())
	}
	var posted map[string]any
	if err := json.Unmarshal(post.Body.Bytes(), &posted); err != nil {
		t.Fatalf("decode post: %v", err)
	}
	ts, _ := posted["ts"].(string)
	if ts != "slack_message_000001" {
		t.Fatalf("post ts = %q, want slack_message_000001", ts)
	}

	historyAfterPost := serveSlackRequest(mux, http.MethodGet, "/slack/api/conversations.history?channel="+channel, "")
	if historyAfterPost.Code != http.StatusOK {
		t.Fatalf("history after post status = %d, want %d, body=%s", historyAfterPost.Code, http.StatusOK, historyAfterPost.Body.String())
	}
	var afterPost struct {
		OK       bool             `json:"ok"`
		Messages []map[string]any `json:"messages"`
	}
	if err := json.Unmarshal(historyAfterPost.Body.Bytes(), &afterPost); err != nil {
		t.Fatalf("decode history after post: %v", err)
	}
	if !afterPost.OK || len(afterPost.Messages) != 1 {
		t.Fatalf("history after post = %#v", afterPost)
	}
	if afterPost.Messages[0]["ts"] != "slack_message_000001" || afterPost.Messages[0]["text"] != "hello" {
		t.Fatalf("history after post message = %#v", afterPost.Messages[0])
	}

	update := serveSlackRequest(mux, http.MethodPost, "/slack/api/chat.update", "channel="+channel+"&ts="+ts+"&text=edited")
	if update.Code != http.StatusOK {
		t.Fatalf("update status = %d, want %d, body=%s", update.Code, http.StatusOK, update.Body.String())
	}

	historyAfterUpdate := serveSlackRequest(mux, http.MethodPost, "/slack/api/conversations.history", "channel="+channel)
	if historyAfterUpdate.Code != http.StatusOK {
		t.Fatalf("history after update status = %d, want %d, body=%s", historyAfterUpdate.Code, http.StatusOK, historyAfterUpdate.Body.String())
	}
	var afterUpdate struct {
		OK       bool             `json:"ok"`
		Messages []map[string]any `json:"messages"`
	}
	if err := json.Unmarshal(historyAfterUpdate.Body.Bytes(), &afterUpdate); err != nil {
		t.Fatalf("decode history after update: %v", err)
	}
	if !afterUpdate.OK || len(afterUpdate.Messages) != 1 {
		t.Fatalf("history after update = %#v", afterUpdate)
	}
	if afterUpdate.Messages[0]["ts"] != ts || afterUpdate.Messages[0]["text"] != "edited" {
		t.Fatalf("history after update message = %#v", afterUpdate.Messages[0])
	}

	del := serveSlackRequest(mux, http.MethodPost, "/slack/api/chat.delete", "channel="+channel+"&ts="+ts)
	if del.Code != http.StatusOK {
		t.Fatalf("delete status = %d, want %d, body=%s", del.Code, http.StatusOK, del.Body.String())
	}
	var deleted struct {
		OK bool `json:"ok"`
	}
	if err := json.Unmarshal(del.Body.Bytes(), &deleted); err != nil {
		t.Fatalf("decode delete: %v", err)
	}
	if !deleted.OK {
		t.Fatalf("delete ok = false, body=%s", del.Body.String())
	}

	historyAfterDelete := serveSlackRequest(mux, http.MethodPost, "/slack/api/conversations.history", "channel="+channel)
	if historyAfterDelete.Code != http.StatusOK {
		t.Fatalf("history after delete status = %d, want %d, body=%s", historyAfterDelete.Code, http.StatusOK, historyAfterDelete.Body.String())
	}
	var afterDelete struct {
		OK       bool             `json:"ok"`
		Messages []map[string]any `json:"messages"`
	}
	if err := json.Unmarshal(historyAfterDelete.Body.Bytes(), &afterDelete); err != nil {
		t.Fatalf("decode history after delete: %v", err)
	}
	if !afterDelete.OK {
		t.Fatalf("history after delete ok = false, body=%s", historyAfterDelete.Body.String())
	}
	if len(afterDelete.Messages) != 0 {
		t.Fatalf("history after delete should be empty, got %#v", afterDelete.Messages)
	}
}

func TestConversationsListUpdateAndDelete(t *testing.T) {
	mux := newSlackMux(t, adapter.Config{BasePath: "/slack", Scenario: "message_success"})

	list := serveSlackRequest(mux, http.MethodPost, "/slack/api/conversations.list", "")
	if list.Code != http.StatusOK || !strings.Contains(list.Body.String(), `"id":"C_MOCKPORT"`) {
		t.Fatalf("list status/body = %d %s", list.Code, list.Body.String())
	}

	post := serveSlackRequest(mux, http.MethodPost, "/slack/api/chat.postMessage", "channel=C_MOCKPORT&text=hello")
	var posted map[string]any
	if err := json.Unmarshal(post.Body.Bytes(), &posted); err != nil {
		t.Fatalf("decode post: %v", err)
	}
	ts, _ := posted["ts"].(string)

	update := serveSlackRequest(mux, http.MethodPost, "/slack/api/chat.update", "channel=C_MOCKPORT&ts="+ts+"&text=edited")
	if update.Code != http.StatusOK || !strings.Contains(update.Body.String(), `"text":"edited"`) {
		t.Fatalf("update status/body = %d %s", update.Code, update.Body.String())
	}

	del := serveSlackRequest(mux, http.MethodPost, "/slack/api/chat.delete", "channel=C_MOCKPORT&ts="+ts)
	if del.Code != http.StatusOK || !strings.Contains(del.Body.String(), `"ok":true`) {
		t.Fatalf("delete status/body = %d %s", del.Code, del.Body.String())
	}

	history := serveSlackRequest(mux, http.MethodPost, "/slack/api/conversations.history", "channel=C_MOCKPORT")
	if history.Code != http.StatusOK || strings.Contains(history.Body.String(), `"text":"edited"`) {
		t.Fatalf("history after delete status/body = %d %s", history.Code, history.Body.String())
	}
}

func TestSlackErrorsAndHeaders(t *testing.T) {
	auth := performRequest(t, adapter.Config{BasePath: "/slack", Scenario: "auth_error"}, http.MethodPost, "/slack/api/auth.test")
	if auth.Code != http.StatusOK {
		t.Fatalf("auth error status = %d, body=%s", auth.Code, auth.Body.String())
	}
	assertSlackError(t, auth, "invalid_auth")

	rate := performRequest(t, adapter.Config{BasePath: "/slack", Scenario: "rate_limited"}, http.MethodPost, "/slack/api/chat.postMessage")
	if rate.Code != http.StatusTooManyRequests || rate.Header().Get("Retry-After") != "1" {
		t.Fatalf("rate status/header/body = %d %q %s", rate.Code, rate.Header().Get("Retry-After"), rate.Body.String())
	}
	assertSlackError(t, rate, "ratelimited")

	mux := newSlackMux(t, adapter.Config{BasePath: "/slack", Scenario: "message_success"})
	unknownChannel := serveSlackRequest(mux, http.MethodPost, "/slack/api/chat.postMessage", "channel=C_UNKNOWN&text=hello")
	if unknownChannel.Code != http.StatusOK {
		t.Fatalf("unknown channel status = %d, body=%s", unknownChannel.Code, unknownChannel.Body.String())
	}
	assertSlackError(t, unknownChannel, "channel_not_found")

	unsupported := serveSlackRequest(mux, http.MethodPost, "/slack/api/views.open", "")
	if unsupported.Code != http.StatusNotFound {
		t.Fatalf("unsupported status = %d, body=%s", unsupported.Code, unsupported.Body.String())
	}
}

func TestEventsURLVerificationAndCallback(t *testing.T) {
	mux := newSlackMux(t, adapter.Config{BasePath: "/slack", Scenario: "message_success", FakeSecret: "mockport_slack_signing_secret"})

	challenge := `{"type":"url_verification","challenge":"challenge-123"}`
	verify := serveSlackSignedRequest(mux, http.MethodPost, "/slack/events", challenge, "mockport_slack_signing_secret")
	if verify.Code != http.StatusOK || !strings.Contains(verify.Body.String(), `"challenge":"challenge-123"`) {
		t.Fatalf("verify status/body = %d %s", verify.Code, verify.Body.String())
	}

	callback := `{"type":"event_callback","event":{"type":"message","channel":"C_MOCKPORT","text":"event hello","user":"U_EVENT"}}`
	event := serveSlackSignedRequest(mux, http.MethodPost, "/slack/events", callback, "mockport_slack_signing_secret")
	if event.Code != http.StatusOK || !strings.Contains(event.Body.String(), `"ok":true`) {
		t.Fatalf("event status/body = %d %s", event.Code, event.Body.String())
	}

	history := serveSlackRequest(mux, http.MethodPost, "/slack/api/conversations.history", "channel=C_MOCKPORT")
	if !strings.Contains(history.Body.String(), `"text":"event hello"`) {
		t.Fatalf("history after event = %s", history.Body.String())
	}

	bad := serveSlackSignedRequest(mux, http.MethodPost, "/slack/events", callback, "wrong_secret")
	if bad.Code != http.StatusUnauthorized {
		t.Fatalf("bad signature status = %d, body=%s", bad.Code, bad.Body.String())
	}
	assertSlackError(t, bad, "invalid_signature")

	old := serveSlackSignedRequestWithTimestamp(mux, http.MethodPost, "/slack/events", callback, "mockport_slack_signing_secret", strconv.FormatInt(time.Now().Add(-10*time.Minute).Unix(), 10))
	if old.Code != http.StatusUnauthorized {
		t.Fatalf("old signature status = %d, body=%s", old.Code, old.Body.String())
	}
	assertSlackError(t, old, "invalid_signature")
}

func TestMetadata(t *testing.T) {
	meta := New().Metadata()
	if meta.Name != "slack" || meta.Maturity != "workflow-compatible" {
		t.Fatalf("metadata = %#v", meta)
	}
	if meta.ProviderVersion != "2025-02-01" || len(meta.Levels) < 5 || len(meta.Endpoints) < 7 {
		t.Fatalf("compat metadata = %#v", meta)
	}
	if !meta.Reset || len(meta.StatefulResources) != 4 {
		t.Fatalf("state metadata = %#v", meta)
	}
	found := false
	for _, endpoint := range meta.Endpoints {
		if endpoint.Path == "/slack/test/reset" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("metadata missing /slack/test/reset")
	}
}

func TestResetClearsStateAndRequiresLoopback(t *testing.T) {
	mux := newSlackMux(t, adapter.Config{BasePath: "/slack", Scenario: "message_success"})

	post := serveSlackRequest(mux, http.MethodPost, "/slack/api/chat.postMessage", "channel=C_MOCKPORT&text=hello")
	if post.Code != http.StatusOK {
		t.Fatalf("post status = %d, body=%s", post.Code, post.Body.String())
	}
	historyBefore := serveSlackRequest(mux, http.MethodGet, "/slack/api/conversations.history?channel=C_MOCKPORT", "")
	if historyBefore.Code != http.StatusOK {
		t.Fatalf("history before reset status = %d, body=%s", historyBefore.Code, historyBefore.Body.String())
	}
	var beforeBody struct {
		Messages []map[string]any `json:"messages"`
	}
	if err := json.Unmarshal(historyBefore.Body.Bytes(), &beforeBody); err != nil {
		t.Fatalf("decode history before reset: %v", err)
	}
	if len(beforeBody.Messages) != 1 {
		t.Fatalf("messages before reset = %#v", beforeBody.Messages)
	}

	reset := serveSlackResetRequest(mux, "127.0.0.1:12345")
	if reset.Code != http.StatusOK {
		t.Fatalf("reset status = %d, body=%s", reset.Code, reset.Body.String())
	}
	var resetBody map[string]any
	if err := json.Unmarshal(reset.Body.Bytes(), &resetBody); err != nil {
		t.Fatalf("decode reset response: %v", err)
	}
	if resetBody["reset"] != true || resetBody["adapter"] != "slack" {
		t.Fatalf("reset body = %#v", resetBody)
	}

	historyAfter := serveSlackRequest(mux, http.MethodGet, "/slack/api/conversations.history?channel=C_MOCKPORT", "")
	if historyAfter.Code != http.StatusOK {
		t.Fatalf("history after reset status = %d, body=%s", historyAfter.Code, historyAfter.Body.String())
	}
	var afterBody struct {
		Messages []map[string]any `json:"messages"`
	}
	if err := json.Unmarshal(historyAfter.Body.Bytes(), &afterBody); err != nil {
		t.Fatalf("decode history after reset: %v", err)
	}
	if len(afterBody.Messages) != 0 {
		t.Fatalf("messages after reset should be empty: %#v", afterBody.Messages)
	}

	remoteReset := serveSlackResetRequestWithRemote(mux, "192.168.0.2:12345")
	if remoteReset.Code != http.StatusForbidden {
		t.Fatalf("remote reset status = %d, body=%s", remoteReset.Code, remoteReset.Body.String())
	}
	if !strings.Contains(remoteReset.Body.String(), "local_request_required") {
		t.Fatalf("remote reset body = %s", remoteReset.Body.String())
	}
}

func performRequest(t *testing.T, cfg adapter.Config, method, path string) *httptest.ResponseRecorder {
	t.Helper()
	mux := newSlackMux(t, cfg)
	return serveSlackRequest(mux, method, path, "")
}

func newSlackMux(t *testing.T, cfg adapter.Config) *http.ServeMux {
	t.Helper()
	return adaptertest.NewMux(t, New(), cfg)
}

func serveSlackRequest(mux http.Handler, method, path, body string) *httptest.ResponseRecorder {
	header := http.Header{}
	if body != "" {
		header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	return adaptertest.Serve(mux, method, path, strings.NewReader(body), header)
}

func serveSlackSignedRequest(mux http.Handler, method, path, body, secret string) *httptest.ResponseRecorder {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	return serveSlackSignedRequestWithTimestamp(mux, method, path, body, secret, timestamp)
}

func serveSlackSignedRequestWithTimestamp(mux http.Handler, method, path, body, secret, timestamp string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Slack-Request-Timestamp", timestamp)
	req.Header.Set("X-Slack-Signature", slackSignature(secret, timestamp, body))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

func slackSignature(secret, timestamp, body string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte("v0:" + timestamp + ":" + body))
	return "v0=" + hex.EncodeToString(mac.Sum(nil))
}

func serveSlackResetRequest(mux http.Handler, remoteAddr string) *httptest.ResponseRecorder {
	return adaptertest.ServeWithRemote(mux, http.MethodPost, "/slack/test/reset", nil, nil, remoteAddr)
}

func serveSlackResetRequestWithRemote(mux http.Handler, remoteAddr string) *httptest.ResponseRecorder {
	return serveSlackResetRequest(mux, remoteAddr)
}

func assertSlackError(t *testing.T, rec *httptest.ResponseRecorder, want string) {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if body["ok"] != false || body["error"] != want {
		t.Fatalf("error body = %#v", body)
	}
}
