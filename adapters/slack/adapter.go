package slack

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/albert-einshutoin/mockport/internal/adapter"
	"github.com/albert-einshutoin/mockport/internal/adapter/httpx"
	"github.com/albert-einshutoin/mockport/internal/security"
	"github.com/albert-einshutoin/mockport/internal/state"
)

type Adapter struct{}

const slackSignatureTolerance = 5 * time.Minute

func New() Adapter { return Adapter{} }

func (a Adapter) Name() string { return "slack" }

func (a Adapter) Register(mux *http.ServeMux, cfg adapter.Config) error {
	basePath := cfg.BasePath
	if basePath == "" {
		basePath = "/slack"
	}
	r := &routes{
		basePath: strings.TrimRight(basePath, "/"),
		cfg:      cfg,
		store:    state.NewStore(),
		resolver: adapter.NewScenarioResolver(cfg, "message_success", a.Metadata()),
	}
	mux.HandleFunc(r.basePath+"/", r.handle)
	return nil
}

func (a Adapter) FakeEnv(cfg adapter.Config) map[string]string {
	basePath := cfg.BasePath
	if basePath == "" {
		basePath = "/slack"
	}
	return map[string]string{
		"SLACK_API_URL":   adapter.LocalBaseURL(basePath + "/api"),
		"SLACK_BOT_TOKEN": "mockport_slack_token",
	}
}

func (a Adapter) Metadata() adapter.Metadata {
	return adapter.Metadata{
		Name:            "slack",
		Maturity:        adapter.MaturityWorkflowCompatible,
		ProviderVersion: "2025-02-01",
		ClientEvidence:  []string{"slack-client-contract"},
		Levels:          []adapter.Level{adapter.LevelWire, adapter.LevelClient, adapter.LevelWorkflow, adapter.LevelState, adapter.LevelError},
		Capabilities:    []string{"auth_test", "chat_post_message", "chat_update", "chat_delete", "conversations_list", "conversations_history", "events_url_verification", "events_message_callback"},
		StatefulResources: []string{
			"channel",
			"user",
			"bot",
			"message",
		},
		Reset: true,
		Scenarios: []adapter.Scenario{
			{Name: "message_success", Supported: true},
			{Name: "auth_error", Supported: true},
			{Name: "rate_limited", Supported: true},
			{Name: "delivery_failed", Supported: true},
			{Name: "channel_not_found", Supported: true},
			{Name: "not_in_channel", Supported: true},
		},
		Endpoints: []adapter.Endpoint{
			{Method: http.MethodPost, Path: "/slack/api/auth.test", SupportedScenarios: []string{"message_success", "auth_error"}, Notes: "Slack-like auth test"},
			{Method: http.MethodPost, Path: "/slack/api/chat.postMessage", SupportedScenarios: []string{"message_success", "auth_error", "rate_limited", "delivery_failed"}, Notes: "Slack-like message post"},
			{Method: http.MethodPost, Path: "/slack/api/chat.update", SupportedScenarios: []string{"message_success", "auth_error", "channel_not_found"}, Notes: "Slack-like message update"},
			{Method: http.MethodPost, Path: "/slack/api/chat.delete", SupportedScenarios: []string{"message_success", "auth_error", "channel_not_found"}, Notes: "Slack-like message delete"},
			{Method: http.MethodPost, Path: "/slack/api/conversations.list", SupportedScenarios: []string{"message_success", "auth_error"}, Notes: "Slack-like conversation listing"},
			{Method: http.MethodGet, Path: "/slack/api/conversations.history", SupportedScenarios: []string{"message_success", "auth_error", "channel_not_found"}, Notes: "Deterministic Slack-like channel history"},
			{Method: http.MethodPost, Path: "/slack/api/conversations.history", SupportedScenarios: []string{"message_success", "auth_error", "channel_not_found"}, Notes: "Deterministic Slack-like channel history"},
			{Method: http.MethodPost, Path: "/slack/events", SupportedScenarios: []string{"message_success", "auth_error"}, Notes: "Slack-like Events API URL verification and message callback subset"},
			{Method: http.MethodPost, Path: "/slack/test/reset", SupportedScenarios: []string{"message_success", "auth_error", "rate_limited", "delivery_failed", "channel_not_found", "not_in_channel"}, Notes: "Clears state for test isolation"},
		},
	}
}

type routes struct {
	basePath string
	cfg      adapter.Config
	store    *state.Store
	resolver *adapter.ScenarioResolver
}

func (r *routes) handle(w http.ResponseWriter, req *http.Request) {
	httpx.LimitRequestBody(w, req)
	path := strings.TrimPrefix(req.URL.Path, r.basePath)
	switch {
	case req.Method == http.MethodPost && path == "/api/auth.test":
		r.writeAuthTest(w, req)
	case req.Method == http.MethodPost && path == "/api/chat.postMessage":
		r.writePostMessage(w, req)
	case req.Method == http.MethodPost && path == "/api/chat.update":
		r.writeUpdateMessage(w, req)
	case req.Method == http.MethodPost && path == "/api/chat.delete":
		r.writeDeleteMessage(w, req)
	case req.Method == http.MethodPost && path == "/api/conversations.list":
		r.writeConversationsList(w, req)
	case (req.Method == http.MethodGet || req.Method == http.MethodPost) && path == "/api/conversations.history":
		r.writeHistory(w, req)
	case req.Method == http.MethodPost && path == "/events":
		r.writeEvent(w, req)
	case req.Method == http.MethodPost && path == "/test/reset":
		r.handleReset(w, req)
	default:
		http.NotFound(w, req)
	}
}

func (r *routes) handleReset(w http.ResponseWriter, req *http.Request) {
	if !security.IsLoopbackRemoteAddr(req.RemoteAddr) {
		writeSlackError(w, http.StatusForbidden, "local_request_required")
		return
	}
	resourceTypes := r.store.ResetAll("slack")
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"reset":          true,
		"adapter":        "slack",
		"resource_types": resourceTypes,
	})
}

func (r *routes) writeAuthTest(w http.ResponseWriter, req *http.Request) {
	scenario, err := r.resolver.Resolve(req)
	if err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, slackErrorResponse{OK: false, Error: "unknown_mockport_scenario"})
		return
	}
	if scenario == "auth_error" {
		writeSlackError(w, http.StatusOK, "invalid_auth")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, authTestResponse{
		OK: true, URL: "https://mockport.slack.test/", Team: "Mockport", TeamID: "T_MOCKPORT",
		User: "mockport-bot", UserID: "U_MOCKPORT", BotID: "B_MOCKPORT",
	})
}

func (r *routes) writePostMessage(w http.ResponseWriter, req *http.Request) {
	if !r.validateWriteScenario(w, req) {
		return
	}
	// ここに到達するのは message_success (default) のみ
	if !parseSlackForm(w, req) {
		return
	}
	channel := req.Form.Get("channel")
	if channel == "" {
		channel = "C_MOCKPORT"
	}
	if !knownChannel(channel) {
		writeSlackError(w, http.StatusOK, "channel_not_found")
		return
	}
	text := req.Form.Get("text")
	if text == "" {
		text = "Mockport message"
	}
	message, err := r.store.Create("slack", "message", map[string]any{
		"channel": channel,
		"deleted": false,
		"text":    text,
		"team":    "T_MOCKPORT",
		"user":    "U_MOCKPORT",
	})
	if err != nil {
		writeSlackError(w, http.StatusInternalServerError, "mockport_state_error")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, postMessageResponse{OK: true, Channel: channel, TS: message.ID, Message: messageBody(message.ID, message.Data)})
}

func (r *routes) writeUpdateMessage(w http.ResponseWriter, req *http.Request) {
	if !r.validateWriteScenario(w, req) {
		return
	}
	if !parseSlackForm(w, req) {
		return
	}
	channel := defaultChannel(req.Form.Get("channel"))
	if !knownChannel(channel) {
		writeSlackError(w, http.StatusOK, "channel_not_found")
		return
	}
	ts := req.Form.Get("ts")
	resource, ok := r.store.Get("slack", "message", ts)
	if !ok || resource.Data["channel"] != channel || resource.Data["deleted"] == true {
		writeSlackError(w, http.StatusOK, "message_not_found")
		return
	}
	text := req.Form.Get("text")
	if text == "" {
		text = "Mockport message"
	}
	resource.Data["text"] = text
	r.store.Update("slack", "message", ts, resource.Data)
	httpx.WriteJSON(w, http.StatusOK, postMessageResponse{OK: true, Channel: channel, TS: ts, Message: messageBody(ts, resource.Data)})
}

func (r *routes) writeDeleteMessage(w http.ResponseWriter, req *http.Request) {
	if !r.validateWriteScenario(w, req) {
		return
	}
	if !parseSlackForm(w, req) {
		return
	}
	channel := defaultChannel(req.Form.Get("channel"))
	if !knownChannel(channel) {
		writeSlackError(w, http.StatusOK, "channel_not_found")
		return
	}
	ts := req.Form.Get("ts")
	resource, ok := r.store.Get("slack", "message", ts)
	if !ok || resource.Data["channel"] != channel {
		writeSlackError(w, http.StatusOK, "message_not_found")
		return
	}
	resource.Data["deleted"] = true
	r.store.Update("slack", "message", ts, resource.Data)
	httpx.WriteJSON(w, http.StatusOK, struct {
		OK      bool   `json:"ok"`
		Channel string `json:"channel"`
		TS      string `json:"ts"`
	}{OK: true, Channel: channel, TS: ts})
}

func (r *routes) writeConversationsList(w http.ResponseWriter, req *http.Request) {
	scenario, err := r.resolver.Resolve(req)
	if err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, slackErrorResponse{OK: false, Error: "unknown_mockport_scenario"})
		return
	}
	if scenario == "auth_error" {
		writeSlackError(w, http.StatusOK, "invalid_auth")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, conversationsListResponse{
		OK:               true,
		Channels:         []channelData{{ID: "C_MOCKPORT", Name: "mockport", IsChannel: true, IsMember: true}},
		ResponseMetadata: responseMetadata{NextCursor: ""},
	})
}

func (r *routes) writeHistory(w http.ResponseWriter, req *http.Request) {
	scenario, err := r.resolver.Resolve(req)
	if err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, slackErrorResponse{OK: false, Error: "unknown_mockport_scenario"})
		return
	}
	if scenario == "auth_error" {
		writeSlackError(w, http.StatusOK, "invalid_auth")
		return
	}
	if !parseSlackForm(w, req) {
		return
	}
	channel := req.URL.Query().Get("channel")
	if channel == "" {
		channel = req.Form.Get("channel")
	}
	channel = defaultChannel(channel)
	if !knownChannel(channel) {
		writeSlackError(w, http.StatusOK, "channel_not_found")
		return
	}
	var messages []messageData
	for _, resource := range r.store.List("slack", "message") {
		if channel != "" && resource.Data["channel"] != channel {
			continue
		}
		if resource.Data["deleted"] == true {
			continue
		}
		messages = append(messages, messageBody(resource.ID, resource.Data))
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "messages": messages})
}

func (r *routes) writeEvent(w http.ResponseWriter, req *http.Request) {
	if _, err := r.resolver.Resolve(req); err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, slackErrorResponse{OK: false, Error: "unknown_mockport_scenario"})
		return
	}
	raw, err := io.ReadAll(req.Body)
	if err != nil {
		if httpx.IsRequestBodyTooLarge(err) {
			writeSlackError(w, http.StatusRequestEntityTooLarge, "request_too_large")
			return
		}
		writeSlackError(w, http.StatusBadRequest, "invalid_payload")
		return
	}
	if !r.validSignature(req, raw) {
		writeSlackError(w, http.StatusUnauthorized, "invalid_signature")
		return
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		writeSlackError(w, http.StatusBadRequest, "invalid_payload")
		return
	}
	switch payload["type"] {
	case "url_verification":
		httpx.WriteJSON(w, http.StatusOK, map[string]any{"challenge": payload["challenge"]})
	case "event_callback":
		if event, ok := payload["event"].(map[string]any); ok && event["type"] == "message" {
			channel, _ := event["channel"].(string)
			channel = defaultChannel(channel)
			text, _ := event["text"].(string)
			if text == "" {
				text = "Mockport event message"
			}
			user, _ := event["user"].(string)
			if user == "" {
				user = "U_MOCKPORT"
			}
			if _, err := r.store.Create("slack", "message", map[string]any{
				"channel": channel,
				"deleted": false,
				"text":    text,
				"team":    "T_MOCKPORT",
				"user":    user,
			}); err != nil {
				writeSlackError(w, http.StatusInternalServerError, "mockport_state_error")
				return
			}
		}
		httpx.WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
	default:
		writeSlackError(w, http.StatusOK, "unsupported_event")
	}
}

func (r *routes) validSignature(req *http.Request, raw []byte) bool {
	secret := r.cfg.FakeSecret
	if secret == "" {
		secret = "mockport_slack_signing_secret"
	}
	signature := req.Header.Get("X-Slack-Signature")
	timestamp := req.Header.Get("X-Slack-Request-Timestamp")
	if signature == "" || timestamp == "" {
		return false
	}
	requestTime, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return false
	}
	now := time.Now().Unix()
	toleranceSeconds := int64(slackSignatureTolerance / time.Second)
	if requestTime < now-toleranceSeconds || requestTime > now+toleranceSeconds {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte("v0:" + timestamp + ":" + string(raw)))
	want := "v0=" + hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(signature), []byte(want))
}

func (r *routes) validateWriteScenario(w http.ResponseWriter, req *http.Request) bool {
	scenario, err := r.resolver.Resolve(req)
	if err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, slackErrorResponse{OK: false, Error: "unknown_mockport_scenario"})
		return false
	}
	switch scenario {
	case "auth_error":
		writeSlackError(w, http.StatusOK, "invalid_auth")
	case "rate_limited":
		w.Header().Set("Retry-After", "1")
		writeSlackError(w, http.StatusTooManyRequests, "ratelimited")
	case "delivery_failed":
		writeSlackError(w, http.StatusOK, "message_delivery_failed")
	case "channel_not_found":
		writeSlackError(w, http.StatusOK, "channel_not_found")
	case "not_in_channel":
		writeSlackError(w, http.StatusOK, "not_in_channel")
	default:
		return true
	}
	return false
}

func defaultChannel(channel string) string {
	if channel == "" {
		return "C_MOCKPORT"
	}
	return channel
}

func knownChannel(channel string) bool {
	return channel == "C_MOCKPORT" || channel == "C_TEST"
}

func messageBody(ts string, data map[string]any) messageData {
	return messageData{Type: "message", Team: data["team"], Channel: data["channel"], TS: ts, User: data["user"], Text: data["text"]}
}

func writeSlackError(w http.ResponseWriter, status int, code string) {
	httpx.WriteJSON(w, status, slackErrorResponse{OK: false, Error: code})
}

func parseSlackForm(w http.ResponseWriter, req *http.Request) bool {
	if err := req.ParseForm(); err != nil {
		if httpx.IsRequestBodyTooLarge(err) {
			writeSlackError(w, http.StatusRequestEntityTooLarge, "request_too_large")
			return false
		}
		writeSlackError(w, http.StatusBadRequest, "invalid_payload")
		return false
	}
	return true
}
