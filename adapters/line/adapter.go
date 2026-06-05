package line

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/albert-einshutoin/mockport/internal/adapter"
	"github.com/albert-einshutoin/mockport/internal/adapter/httpx"
	"github.com/albert-einshutoin/mockport/internal/security"
	"github.com/albert-einshutoin/mockport/internal/state"
)

type Adapter struct{}

func New() Adapter { return Adapter{} }

func (a Adapter) Name() string { return "line" }

func (a Adapter) Register(mux *http.ServeMux, cfg adapter.Config) error {
	basePath := cfg.BasePath
	if basePath == "" {
		basePath = "/line"
	}
	r := &routes{basePath: strings.TrimRight(basePath, "/"), cfg: cfg, store: state.NewStore()}
	mux.HandleFunc(r.basePath+"/", r.handle)
	return nil
}

func (a Adapter) FakeEnv(cfg adapter.Config) map[string]string {
	basePath := cfg.BasePath
	if basePath == "" {
		basePath = "/line"
	}
	return map[string]string{
		"LINE_API_BASE_URL":        "http://localhost:43101" + basePath,
		"LINE_CHANNEL_ID":          "mockport_line_channel",
		"LINE_CHANNEL_SECRET":      "mockport_line_secret",
		"LINE_CHANNEL_TOKEN":       "mockport_line_channel_token",
		"LINE_LIFF_ID":             "mockport-line-liff",
		"LINE_PAY_CHANNEL_ID":      "mockport_line_pay_channel",
		"LINE_PAY_CHANNEL_SECRET":  "mockport_line_pay_secret",
		"LINE_MINI_DAPP_CLIENT_ID": "mockport_line_mini_dapp_client",
	}
}

func (a Adapter) Metadata() adapter.Metadata {
	scenarios := []adapter.Scenario{
		{Name: "line_success", Supported: true},
		{Name: "auth_error", Supported: true},
		{Name: "rate_limited", Supported: true},
		{Name: "invalid_request", Supported: true},
		{Name: "pay_failed", Supported: true},
	}
	return adapter.Metadata{
		Name:            "line",
		Maturity:        adapter.MaturityWorkflowCompatible,
		ProviderVersion: "Messaging API v2 / Login v2.1 / Pay v3 / MINI App service messages / Mini Dapp SDK",
		Levels:          []adapter.Level{adapter.LevelWire, adapter.LevelWorkflow, adapter.LevelState, adapter.LevelError},
		Capabilities: []string{
			"channel_access_token",
			"messaging_bot_info",
			"messaging_content",
			"messaging_group_room",
			"messaging_mark_as_read",
			"messaging_multicast_broadcast_narrowcast",
			"messaging_push",
			"messaging_quota_delivery",
			"messaging_reply",
			"messaging_profile",
			"messaging_rich_menu",
			"messaging_webhook_delivery",
			"messaging_webhook_settings",
			"login_oauth",
			"login_profile",
			"liff_profile",
			"mini_app_service_messages",
			"line_pay_request_confirm",
			"mini_dapp_wallet_payment",
		},
		Scenarios: scenarios,
		StatefulResources: []string{
			"message",
			"oauth_code",
			"oauth_token",
			"rich_menu",
			"rich_menu_alias",
			"user_rich_menu",
			"notification_token",
			"line_pay_payment",
			"mini_dapp_payment",
		},
		Reset: true,
		Endpoints: []adapter.Endpoint{
			{Method: http.MethodPost, Path: "/line/v2/bot/message/push", SupportedScenarios: []string{"line_success", "auth_error", "rate_limited", "invalid_request"}, Notes: "LINE Messaging API-like push message"},
			{Method: http.MethodPost, Path: "/line/v2/bot/message/reply", SupportedScenarios: []string{"line_success", "auth_error", "rate_limited", "invalid_request"}, Notes: "LINE Messaging API-like reply message"},
			{Method: http.MethodPost, Path: "/line/v2/bot/message/multicast", SupportedScenarios: []string{"line_success", "auth_error", "rate_limited", "invalid_request"}, Notes: "LINE Messaging API-like multicast message"},
			{Method: http.MethodPost, Path: "/line/v2/bot/message/broadcast", SupportedScenarios: []string{"line_success", "auth_error", "rate_limited", "invalid_request"}, Notes: "LINE Messaging API-like broadcast message"},
			{Method: http.MethodPost, Path: "/line/v2/bot/message/narrowcast", SupportedScenarios: []string{"line_success", "auth_error", "rate_limited", "invalid_request"}, Notes: "LINE Messaging API-like narrowcast acceptance"},
			{Method: http.MethodGet, Path: "/line/v2/bot/message/progress/narrowcast", SupportedScenarios: []string{"line_success"}, Notes: "LINE Messaging API-like narrowcast progress"},
			{Method: http.MethodPost, Path: "/line/v2/bot/message/validate/{type}", SupportedScenarios: []string{"line_success", "invalid_request"}, Notes: "LINE Messaging API-like message object validation"},
			{Method: http.MethodPost, Path: "/line/v2/bot/chat/markAsRead", SupportedScenarios: []string{"line_success"}, Notes: "LINE Messaging API-like mark-as-read acknowledgement"},
			{Method: http.MethodPost, Path: "/line/v2/bot/chat/loading/start", SupportedScenarios: []string{"line_success"}, Notes: "LINE Messaging API-like loading animation acknowledgement"},
			{Method: http.MethodGet, Path: "/line/v2/bot/message/{messageId}/content", SupportedScenarios: []string{"line_success"}, Notes: "LINE Messaging API-like binary content download"},
			{Method: http.MethodGet, Path: "/line/v2/bot/message/{messageId}/content/transcoding", SupportedScenarios: []string{"line_success"}, Notes: "LINE Messaging API-like content transcoding status"},
			{Method: http.MethodGet, Path: "/line/v2/bot/message/{messageId}/content/preview", SupportedScenarios: []string{"line_success"}, Notes: "LINE Messaging API-like content preview download"},
			{Method: http.MethodGet, Path: "/line/v2/bot/profile/{userId}", SupportedScenarios: []string{"line_success", "auth_error"}, Notes: "LINE Messaging API-like user profile"},
			{Method: http.MethodPut, Path: "/line/v2/bot/channel/webhook/endpoint", SupportedScenarios: []string{"line_success", "invalid_request"}, Notes: "LINE Messaging API-like webhook endpoint update"},
			{Method: http.MethodGet, Path: "/line/v2/bot/channel/webhook/endpoint", SupportedScenarios: []string{"line_success", "invalid_request"}, Notes: "LINE Messaging API-like webhook endpoint lookup"},
			{Method: http.MethodPost, Path: "/line/v2/bot/channel/webhook/test", SupportedScenarios: []string{"line_success", "auth_error"}, Notes: "LINE Messaging API-like webhook test"},
			{Method: http.MethodPost, Path: "/line/test/webhook/send", SupportedScenarios: []string{"line_success", "invalid_request"}, Notes: "Sends a fake LINE signed webhook to the configured target"},
			{Method: http.MethodGet, Path: "/line/v2/bot/info", SupportedScenarios: []string{"line_success"}, Notes: "LINE Messaging API-like bot information"},
			{Method: http.MethodGet, Path: "/line/v2/bot/message/quota", SupportedScenarios: []string{"line_success"}, Notes: "LINE Messaging API-like quota lookup"},
			{Method: http.MethodGet, Path: "/line/v2/bot/message/quota/consumption", SupportedScenarios: []string{"line_success"}, Notes: "LINE Messaging API-like quota consumption lookup"},
			{Method: http.MethodGet, Path: "/line/v2/bot/message/delivery/{type}", SupportedScenarios: []string{"line_success"}, Notes: "LINE Messaging API-like delivery statistics"},
			{Method: http.MethodGet, Path: "/line/v2/bot/message/aggregation/info", SupportedScenarios: []string{"line_success"}, Notes: "LINE Messaging API-like aggregation unit summary"},
			{Method: http.MethodGet, Path: "/line/v2/bot/message/aggregation/list", SupportedScenarios: []string{"line_success"}, Notes: "LINE Messaging API-like aggregation unit list"},
			{Method: http.MethodGet, Path: "/line/v2/bot/followers/ids", SupportedScenarios: []string{"line_success"}, Notes: "LINE Messaging API-like follower IDs"},
			{Method: http.MethodGet, Path: "/line/v2/bot/group/{groupId}/summary", SupportedScenarios: []string{"line_success"}, Notes: "LINE Messaging API-like group summary"},
			{Method: http.MethodGet, Path: "/line/v2/bot/group/{groupId}/members/ids", SupportedScenarios: []string{"line_success"}, Notes: "LINE Messaging API-like group member IDs"},
			{Method: http.MethodGet, Path: "/line/v2/bot/group/{groupId}/member/{userId}", SupportedScenarios: []string{"line_success"}, Notes: "LINE Messaging API-like group member profile"},
			{Method: http.MethodPost, Path: "/line/v2/bot/group/{groupId}/leave", SupportedScenarios: []string{"line_success"}, Notes: "LINE Messaging API-like group leave acknowledgement"},
			{Method: http.MethodGet, Path: "/line/v2/bot/room/{roomId}/members/ids", SupportedScenarios: []string{"line_success"}, Notes: "LINE Messaging API-like room member IDs"},
			{Method: http.MethodGet, Path: "/line/v2/bot/room/{roomId}/member/{userId}", SupportedScenarios: []string{"line_success"}, Notes: "LINE Messaging API-like room member profile"},
			{Method: http.MethodPost, Path: "/line/v2/bot/room/{roomId}/leave", SupportedScenarios: []string{"line_success"}, Notes: "LINE Messaging API-like room leave acknowledgement"},
			{Method: http.MethodPost, Path: "/line/v2/bot/richmenu", SupportedScenarios: []string{"line_success", "invalid_request"}, Notes: "LINE Messaging API-like rich menu creation"},
			{Method: http.MethodPost, Path: "/line/v2/bot/richmenu/validate", SupportedScenarios: []string{"line_success", "invalid_request"}, Notes: "LINE Messaging API-like rich menu validation"},
			{Method: http.MethodGet, Path: "/line/v2/bot/richmenu/list", SupportedScenarios: []string{"line_success"}, Notes: "LINE Messaging API-like rich menu list"},
			{Method: http.MethodGet, Path: "/line/v2/bot/richmenu/{richMenuId}", SupportedScenarios: []string{"line_success", "invalid_request"}, Notes: "LINE Messaging API-like rich menu lookup"},
			{Method: http.MethodDelete, Path: "/line/v2/bot/richmenu/{richMenuId}", SupportedScenarios: []string{"line_success", "invalid_request"}, Notes: "LINE Messaging API-like rich menu deletion"},
			{Method: http.MethodPost, Path: "/line/v2/bot/richmenu/{richMenuId}/content", SupportedScenarios: []string{"line_success", "invalid_request"}, Notes: "LINE Messaging API-like rich menu image upload"},
			{Method: http.MethodGet, Path: "/line/v2/bot/richmenu/{richMenuId}/content", SupportedScenarios: []string{"line_success", "invalid_request"}, Notes: "LINE Messaging API-like rich menu image download"},
			{Method: http.MethodPost, Path: "/line/v2/bot/richmenu/alias", SupportedScenarios: []string{"line_success", "invalid_request"}, Notes: "LINE Messaging API-like rich menu alias creation"},
			{Method: http.MethodGet, Path: "/line/v2/bot/richmenu/alias/list", SupportedScenarios: []string{"line_success"}, Notes: "LINE Messaging API-like rich menu alias list"},
			{Method: http.MethodGet, Path: "/line/v2/bot/richmenu/alias/{aliasId}", SupportedScenarios: []string{"line_success", "invalid_request"}, Notes: "LINE Messaging API-like rich menu alias lookup"},
			{Method: http.MethodPost, Path: "/line/v2/bot/richmenu/alias/{aliasId}", SupportedScenarios: []string{"line_success", "invalid_request"}, Notes: "LINE Messaging API-like rich menu alias update"},
			{Method: http.MethodDelete, Path: "/line/v2/bot/richmenu/alias/{aliasId}", SupportedScenarios: []string{"line_success", "invalid_request"}, Notes: "LINE Messaging API-like rich menu alias deletion"},
			{Method: http.MethodGet, Path: "/line/v2/bot/richmenu/progress/batch", SupportedScenarios: []string{"line_success"}, Notes: "LINE Messaging API-like rich menu batch progress"},
			{Method: http.MethodPost, Path: "/line/v2/bot/richmenu/bulk/link", SupportedScenarios: []string{"line_success"}, Notes: "LINE Messaging API-like bulk rich menu link acknowledgement"},
			{Method: http.MethodPost, Path: "/line/v2/bot/richmenu/bulk/unlink", SupportedScenarios: []string{"line_success"}, Notes: "LINE Messaging API-like bulk rich menu unlink acknowledgement"},
			{Method: http.MethodPost, Path: "/line/v2/bot/richmenu/batch", SupportedScenarios: []string{"line_success"}, Notes: "LINE Messaging API-like rich menu batch acknowledgement"},
			{Method: http.MethodPost, Path: "/line/v2/bot/richmenu/validate/batch", SupportedScenarios: []string{"line_success"}, Notes: "LINE Messaging API-like rich menu batch validation"},
			{Method: http.MethodPost, Path: "/line/v2/bot/user/all/richmenu/{richMenuId}", SupportedScenarios: []string{"line_success", "invalid_request"}, Notes: "LINE Messaging API-like default rich menu link"},
			{Method: http.MethodGet, Path: "/line/v2/bot/user/all/richmenu", SupportedScenarios: []string{"line_success", "invalid_request"}, Notes: "LINE Messaging API-like default rich menu lookup"},
			{Method: http.MethodDelete, Path: "/line/v2/bot/user/all/richmenu", SupportedScenarios: []string{"line_success"}, Notes: "LINE Messaging API-like default rich menu unlink"},
			{Method: http.MethodPost, Path: "/line/v2/bot/user/{userId}/richmenu/{richMenuId}", SupportedScenarios: []string{"line_success", "invalid_request"}, Notes: "LINE Messaging API-like user rich menu link"},
			{Method: http.MethodGet, Path: "/line/v2/bot/user/{userId}/richmenu", SupportedScenarios: []string{"line_success", "invalid_request"}, Notes: "LINE Messaging API-like user rich menu lookup"},
			{Method: http.MethodDelete, Path: "/line/v2/bot/user/{userId}/richmenu", SupportedScenarios: []string{"line_success"}, Notes: "LINE Messaging API-like user rich menu unlink"},
			{Method: http.MethodPost, Path: "/line/v2/bot/user/{userId}/linkToken", SupportedScenarios: []string{"line_success"}, Notes: "LINE Messaging API-like account link token issue"},
			{Method: http.MethodGet, Path: "/line/oauth2/v2.1/authorize", SupportedScenarios: []string{"line_success", "invalid_request"}, Notes: "LINE Login-like authorization redirect"},
			{Method: http.MethodPost, Path: "/line/oauth2/v2.1/token", SupportedScenarios: []string{"line_success", "auth_error", "invalid_request"}, Notes: "LINE Login-like token exchange"},
			{Method: http.MethodGet, Path: "/line/oauth2/v2.1/verify", SupportedScenarios: []string{"line_success", "invalid_request"}, Notes: "LINE channel access token v2.1-like verification"},
			{Method: http.MethodGet, Path: "/line/oauth2/v2.1/tokens/kid", SupportedScenarios: []string{"line_success"}, Notes: "LINE channel access token v2.1-like key ID list"},
			{Method: http.MethodPost, Path: "/line/oauth2/v2.1/revoke", SupportedScenarios: []string{"line_success"}, Notes: "LINE channel access token v2.1-like revocation"},
			{Method: http.MethodPost, Path: "/line/oauth2/v3/token", SupportedScenarios: []string{"line_success", "auth_error", "invalid_request"}, Notes: "LINE stateless channel access token-like issue"},
			{Method: http.MethodPost, Path: "/line/v2/oauth/accessToken", SupportedScenarios: []string{"line_success", "auth_error", "invalid_request"}, Notes: "LINE short-lived channel access token-like issue"},
			{Method: http.MethodPost, Path: "/line/v2/oauth/verify", SupportedScenarios: []string{"line_success", "invalid_request"}, Notes: "LINE short-lived channel access token-like verification"},
			{Method: http.MethodPost, Path: "/line/v2/oauth/revoke", SupportedScenarios: []string{"line_success"}, Notes: "LINE short-lived channel access token-like revocation"},
			{Method: http.MethodGet, Path: "/line/v2/profile", SupportedScenarios: []string{"line_success", "auth_error"}, Notes: "LINE Login-like profile endpoint"},
			{Method: http.MethodGet, Path: "/line/liff/v2/profile", SupportedScenarios: []string{"line_success", "auth_error"}, Notes: "LIFF profile helper for local app tests"},
			{Method: http.MethodGet, Path: "/line/liff/v2/context", SupportedScenarios: []string{"line_success"}, Notes: "LIFF context helper for local app tests"},
			{Method: http.MethodPost, Path: "/line/message/v3/notifier/token", SupportedScenarios: []string{"line_success", "auth_error", "invalid_request"}, Notes: "LINE MINI App Service Message API-like notification token issue"},
			{Method: http.MethodPost, Path: "/line/message/v3/notifier/send", SupportedScenarios: []string{"line_success", "auth_error", "invalid_request"}, Notes: "LINE MINI App Service Message API-like send"},
			{Method: http.MethodPost, Path: "/line/v3/payments/request", SupportedScenarios: []string{"line_success", "auth_error", "pay_failed", "invalid_request"}, Notes: "LINE Pay v3-like payment request"},
			{Method: http.MethodPost, Path: "/line/v3/payments/{transactionId}/confirm", SupportedScenarios: []string{"line_success", "auth_error", "pay_failed"}, Notes: "LINE Pay v3-like payment confirm"},
			{Method: http.MethodGet, Path: "/line/v3/payments/requests/{transactionId}/check", SupportedScenarios: []string{"line_success", "pay_failed"}, Notes: "LINE Pay v3-like payment request status"},
			{Method: http.MethodPost, Path: "/line/mini-dapp/v1/wallet/sessions", SupportedScenarios: []string{"line_success", "auth_error", "invalid_request"}, Notes: "Mini Dapp wallet session helper"},
			{Method: http.MethodPost, Path: "/line/mini-dapp/v1/payments", SupportedScenarios: []string{"line_success", "auth_error", "pay_failed", "invalid_request"}, Notes: "Mini Dapp SDK-like payment helper"},
			{Method: http.MethodGet, Path: "/line/mini-dapp/v1/payments/{id}", SupportedScenarios: []string{"line_success", "pay_failed"}, Notes: "Mini Dapp SDK-like payment lookup"},
		},
	}
}

type routes struct {
	basePath string
	cfg      adapter.Config
	store    *state.Store

	// mu guards the singleton mutable state below. net/http dispatches
	// concurrent requests to the same routes instance, so these fields must
	// not be touched without holding the lock.
	mu                sync.RWMutex
	webhookEndpoint   string
	defaultRichMenuID string
}

func (r *routes) setWebhookEndpoint(endpoint string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.webhookEndpoint = endpoint
}

func (r *routes) getWebhookEndpoint() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.webhookEndpoint
}

func (r *routes) setDefaultRichMenu(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.defaultRichMenuID = id
}

// deleteRichMenu removes a rich menu and clears the default pointer in the same
// critical section, so the default can never be left referencing a menu that no
// longer exists in the store.
func (r *routes) deleteRichMenu(id string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.store.Delete("line", "rich_menu", id) {
		return false
	}
	if r.defaultRichMenuID == id {
		r.defaultRichMenuID = ""
	}
	return true
}

// setDefaultRichMenuChecked validates that the rich menu still exists with an
// uploaded image and adopts it as the default within a single lock, closing the
// check-then-set race against a concurrent delete. On success it returns
// http.StatusOK with an empty message.
func (r *routes) setDefaultRichMenuChecked(id string) (status int, message string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	resource, ok := r.store.Get("line", "rich_menu", id)
	if !ok {
		return http.StatusNotFound, "Not found"
	}
	if resource.Data["hasImage"] != true {
		return http.StatusBadRequest, "must upload richmenu image before applying it to user"
	}
	r.defaultRichMenuID = id
	return http.StatusOK, ""
}

// currentDefaultRichMenu returns the default rich menu id only while it still
// exists in the store, so a concurrently deleted menu is never reported.
func (r *routes) currentDefaultRichMenu() (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.defaultRichMenuID == "" {
		return "", false
	}
	if _, ok := r.store.Get("line", "rich_menu", r.defaultRichMenuID); !ok {
		return "", false
	}
	return r.defaultRichMenuID, true
}

func (r *routes) handle(w http.ResponseWriter, req *http.Request) {
	httpx.LimitRequestBody(w, req)
	w.Header().Set("X-Line-Request-Id", "line-request-mockport")
	path := strings.TrimPrefix(req.URL.Path, r.basePath)
	switch {
	case req.Method == http.MethodPost && strings.HasPrefix(path, "/v2/bot/message/validate/"):
		r.writeValidateMessage(w, req)
	case req.Method == http.MethodPost && (path == "/v2/bot/message/push" || path == "/v2/bot/message/reply"):
		r.writeMessage(w, req, true, http.StatusOK)
	case req.Method == http.MethodPost && (path == "/v2/bot/message/multicast" || path == "/v2/bot/message/broadcast"):
		r.writeMessage(w, req, false, http.StatusOK)
	case req.Method == http.MethodPost && path == "/v2/bot/message/narrowcast":
		r.writeMessage(w, req, false, http.StatusAccepted)
	case req.Method == http.MethodGet && path == "/v2/bot/message/progress/narrowcast":
		r.writeNarrowcastProgress(w)
	case req.Method == http.MethodPost && path == "/v2/bot/chat/markAsRead":
		writeEmptyJSON(w, http.StatusOK)
	case req.Method == http.MethodPost && path == "/v2/bot/chat/loading/start":
		writeEmptyJSON(w, http.StatusAccepted)
	case req.Method == http.MethodGet && strings.HasPrefix(path, "/v2/bot/message/") && strings.Contains(path, "/content"):
		r.writeContentEndpoint(w, path)
	case req.Method == http.MethodGet && strings.HasPrefix(path, "/v2/bot/profile/"):
		r.writeMessagingProfile(w, strings.TrimPrefix(path, "/v2/bot/profile/"))
	case req.Method == http.MethodPut && path == "/v2/bot/channel/webhook/endpoint":
		r.writeSetWebhookEndpoint(w, req)
	case req.Method == http.MethodGet && path == "/v2/bot/channel/webhook/endpoint":
		r.writeGetWebhookEndpoint(w)
	case req.Method == http.MethodPost && path == "/v2/bot/channel/webhook/test":
		r.writeWebhookTest(w)
	case req.Method == http.MethodPost && path == "/test/webhook/send":
		r.sendWebhook(w, req)
	case req.Method == http.MethodGet && path == "/oauth2/v2.1/authorize":
		r.writeAuthorize(w, req)
	case req.Method == http.MethodPost && path == "/oauth2/v2.1/token":
		r.writeToken(w, req)
	case req.Method == http.MethodGet && path == "/oauth2/v2.1/verify":
		r.writeVerifyToken(w, req)
	case req.Method == http.MethodGet && path == "/oauth2/v2.1/tokens/kid":
		httpx.WriteJSON(w, http.StatusOK, map[string]any{"kids": []string{"mockport_line_key_id"}})
	case req.Method == http.MethodPost && path == "/oauth2/v2.1/revoke":
		writeNoBody(w, http.StatusOK)
	case req.Method == http.MethodPost && path == "/oauth2/v3/token":
		r.writeStatelessToken(w, req)
	case req.Method == http.MethodPost && path == "/v2/oauth/accessToken":
		r.writeShortLivedToken(w, req)
	case req.Method == http.MethodPost && path == "/v2/oauth/verify":
		r.writeVerifyOAuthToken(w, req)
	case req.Method == http.MethodPost && path == "/v2/oauth/revoke":
		writeNoBody(w, http.StatusOK)
	case req.Method == http.MethodGet && path == "/v2/profile":
		r.writeLoginProfile(w, req)
	case req.Method == http.MethodGet && path == "/v2/bot/info":
		r.writeBotInfo(w)
	case req.Method == http.MethodGet && path == "/v2/bot/message/quota":
		httpx.WriteJSON(w, http.StatusOK, map[string]any{"type": "limited", "value": 10000})
	case req.Method == http.MethodGet && path == "/v2/bot/message/quota/consumption":
		httpx.WriteJSON(w, http.StatusOK, map[string]any{"totalUsage": len(r.store.List("line", "message"))})
	case req.Method == http.MethodGet && strings.HasPrefix(path, "/v2/bot/message/delivery/"):
		httpx.WriteJSON(w, http.StatusOK, map[string]any{"status": "ready", "success": len(r.store.List("line", "message"))})
	case req.Method == http.MethodGet && path == "/v2/bot/message/aggregation/info":
		httpx.WriteJSON(w, http.StatusOK, map[string]any{"numOfCustomAggregationUnits": 1})
	case req.Method == http.MethodGet && path == "/v2/bot/message/aggregation/list":
		httpx.WriteJSON(w, http.StatusOK, map[string]any{"customAggregationUnits": []string{"mockport_unit"}})
	case req.Method == http.MethodGet && path == "/v2/bot/followers/ids":
		httpx.WriteJSON(w, http.StatusOK, map[string]any{"userIds": []string{"Umockport", "Ulocaluser"}})
	case req.Method == http.MethodGet && strings.HasPrefix(path, "/v2/bot/group/"):
		r.writeGroupEndpoint(w, path)
	case req.Method == http.MethodPost && strings.HasPrefix(path, "/v2/bot/group/") && strings.HasSuffix(path, "/leave"):
		writeEmptyJSON(w, http.StatusOK)
	case req.Method == http.MethodGet && strings.HasPrefix(path, "/v2/bot/room/"):
		r.writeRoomEndpoint(w, path)
	case req.Method == http.MethodPost && strings.HasPrefix(path, "/v2/bot/room/") && strings.HasSuffix(path, "/leave"):
		writeEmptyJSON(w, http.StatusOK)
	case req.Method == http.MethodGet && path == "/liff/v2/profile":
		r.writeLIFFProfile(w)
	case req.Method == http.MethodGet && path == "/liff/v2/context":
		r.writeLIFFContext(w)
	case req.Method == http.MethodPost && path == "/v2/bot/richmenu":
		r.writeCreateRichMenu(w, req)
	case req.Method == http.MethodPost && path == "/v2/bot/richmenu/validate":
		r.writeValidateRichMenu(w, req)
	case req.Method == http.MethodGet && path == "/v2/bot/richmenu/list":
		r.writeRichMenuList(w)
	case req.Method == http.MethodPost && strings.HasPrefix(path, "/v2/bot/richmenu/") && strings.HasSuffix(path, "/content"):
		r.writeUploadRichMenuImage(w, path)
	case req.Method == http.MethodGet && strings.HasPrefix(path, "/v2/bot/richmenu/") && strings.HasSuffix(path, "/content"):
		r.writeDownloadRichMenuImage(w, path)
	case req.Method == http.MethodGet && path == "/v2/bot/richmenu/alias/list":
		r.writeRichMenuAliasList(w)
	case req.Method == http.MethodGet && strings.HasPrefix(path, "/v2/bot/richmenu/alias/"):
		r.writeGetRichMenuAlias(w, strings.TrimPrefix(path, "/v2/bot/richmenu/alias/"))
	case req.Method == http.MethodPost && path == "/v2/bot/richmenu/alias":
		r.writeCreateRichMenuAlias(w, req)
	case req.Method == http.MethodPost && strings.HasPrefix(path, "/v2/bot/richmenu/alias/"):
		r.writeUpdateRichMenuAlias(w, req, strings.TrimPrefix(path, "/v2/bot/richmenu/alias/"))
	case req.Method == http.MethodDelete && strings.HasPrefix(path, "/v2/bot/richmenu/alias/"):
		r.writeDeleteRichMenuAlias(w, strings.TrimPrefix(path, "/v2/bot/richmenu/alias/"))
	case req.Method == http.MethodGet && path == "/v2/bot/richmenu/progress/batch":
		httpx.WriteJSON(w, http.StatusOK, map[string]any{"phase": "succeeded", "acceptedTime": "2999-01-01T00:00:00.000Z", "completedTime": "2999-01-01T00:00:01.000Z"})
	case req.Method == http.MethodGet && strings.HasPrefix(path, "/v2/bot/richmenu/"):
		r.writeGetRichMenu(w, strings.TrimPrefix(path, "/v2/bot/richmenu/"))
	case req.Method == http.MethodDelete && strings.HasPrefix(path, "/v2/bot/richmenu/"):
		r.writeDeleteRichMenu(w, strings.TrimPrefix(path, "/v2/bot/richmenu/"))
	case req.Method == http.MethodPost && strings.HasPrefix(path, "/v2/bot/user/all/richmenu/"):
		r.writeSetDefaultRichMenu(w, strings.TrimPrefix(path, "/v2/bot/user/all/richmenu/"))
	case req.Method == http.MethodGet && path == "/v2/bot/user/all/richmenu":
		r.writeGetDefaultRichMenu(w)
	case req.Method == http.MethodDelete && path == "/v2/bot/user/all/richmenu":
		r.setDefaultRichMenu("")
		writeEmptyJSON(w, http.StatusOK)
	case req.Method == http.MethodPost && strings.HasPrefix(path, "/v2/bot/user/") && strings.Contains(path, "/richmenu/"):
		r.writeLinkRichMenuToUser(w, path)
	case req.Method == http.MethodGet && strings.HasPrefix(path, "/v2/bot/user/") && strings.HasSuffix(path, "/richmenu"):
		r.writeGetUserRichMenu(w, path)
	case req.Method == http.MethodDelete && strings.HasPrefix(path, "/v2/bot/user/") && strings.HasSuffix(path, "/richmenu"):
		r.writeUnlinkRichMenuFromUser(w, path)
	case req.Method == http.MethodPost && path == "/v2/bot/richmenu/bulk/link":
		writeEmptyJSON(w, http.StatusAccepted)
	case req.Method == http.MethodPost && path == "/v2/bot/richmenu/bulk/unlink":
		writeEmptyJSON(w, http.StatusAccepted)
	case req.Method == http.MethodPost && path == "/v2/bot/richmenu/batch":
		writeEmptyJSON(w, http.StatusOK)
	case req.Method == http.MethodPost && path == "/v2/bot/richmenu/validate/batch":
		writeEmptyJSON(w, http.StatusOK)
	case req.Method == http.MethodPost && strings.HasPrefix(path, "/v2/bot/user/") && strings.HasSuffix(path, "/linkToken"):
		httpx.WriteJSON(w, http.StatusOK, map[string]any{"linkToken": "mockport_line_link_token"})
	case req.Method == http.MethodPost && path == "/message/v3/notifier/token":
		r.writeNotificationToken(w, req)
	case req.Method == http.MethodPost && path == "/message/v3/notifier/send":
		r.writeServiceMessage(w, req)
	case req.Method == http.MethodPost && path == "/v3/payments/request":
		r.writePayRequest(w, req)
	case req.Method == http.MethodPost && strings.HasPrefix(path, "/v3/payments/") && strings.HasSuffix(path, "/confirm"):
		id := strings.TrimSuffix(strings.TrimPrefix(path, "/v3/payments/"), "/confirm")
		r.writePayConfirm(w, id)
	case req.Method == http.MethodGet && strings.HasPrefix(path, "/v3/payments/requests/") && strings.HasSuffix(path, "/check"):
		id := strings.TrimSuffix(strings.TrimPrefix(path, "/v3/payments/requests/"), "/check")
		r.writePayCheck(w, id)
	case req.Method == http.MethodPost && path == "/mini-dapp/v1/wallet/sessions":
		r.writeMiniDappWalletSession(w, req)
	case req.Method == http.MethodPost && path == "/mini-dapp/v1/payments":
		r.writeMiniDappPayment(w, req)
	case req.Method == http.MethodGet && strings.HasPrefix(path, "/mini-dapp/v1/payments/"):
		r.writeMiniDappPaymentLookup(w, strings.TrimPrefix(path, "/mini-dapp/v1/payments/"))
	default:
		http.NotFound(w, req)
	}
}

func (r *routes) writeValidateMessage(w http.ResponseWriter, req *http.Request) {
	payload, err := decodePayload(req)
	if err != nil {
		writeDecodeError(w, err)
		return
	}
	messages, _ := payload["messages"].([]any)
	if len(messages) == 0 || len(messages) > 5 {
		writeValidationDetails(w, []lineErrorDetail{{Message: "Size must be between 1 and 5", Property: "messages"}})
		return
	}
	details := make([]lineErrorDetail, 0)
	allowedMessageTypes := map[string]bool{
		"text": true, "image": true, "video": true, "audio": true, "location": true,
		"sticker": true, "template": true, "imagemap": true, "flex": true, "coupon": true,
	}
	for i, message := range messages {
		messageObject, ok := message.(map[string]any)
		if !ok {
			details = append(details, lineErrorDetail{Message: "Must be a JSON object", Property: fmt.Sprintf("messages[%d]", i)})
			continue
		}
		messageType, _ := messageObject["type"].(string)
		if !allowedMessageTypes[messageType] {
			details = append(details, lineErrorDetail{
				Message:  "Must be one of the following values: [text, image, video, audio, location, sticker, template, imagemap, flex, coupon]",
				Property: fmt.Sprintf("messages[%d].type", i),
			})
			continue
		}
		if messageType == "text" {
			text, _ := messageObject["text"].(string)
			if strings.TrimSpace(text) == "" {
				details = append(details, lineErrorDetail{Message: "May not be empty", Property: fmt.Sprintf("messages[%d].text", i)})
			}
		}
	}
	if len(details) > 0 {
		writeValidationDetails(w, details)
		return
	}
	writeEmptyJSON(w, http.StatusOK)
}

func (r *routes) writeMessage(w http.ResponseWriter, req *http.Request, includeSentMessages bool, status int) {
	switch normalizeScenario(r.cfg.Scenario) {
	case "auth_error":
		writeLINEError(w, http.StatusUnauthorized, "Mockport simulated invalid channel access token")
	case "rate_limited":
		writeLINEError(w, http.StatusTooManyRequests, "Mockport simulated rate limit")
	case "invalid_request":
		writeLINEError(w, http.StatusBadRequest, "Mockport simulated invalid message request")
	default:
		payload, err := decodePayload(req)
		if err != nil {
			writeDecodeError(w, err)
			return
		}
		if len(payload) > 0 && payload["messages"] == nil {
			writeLINEError(w, http.StatusBadRequest, "messages is required")
			return
		}
		resource, err := r.store.Create("line", "message", map[string]any{
			"to":       firstNonEmpty(payload["to"], payload["replyToken"], "Umockport"),
			"messages": firstNonEmpty(payload["messages"], []any{map[string]any{"type": "text", "text": "Mockport LINE message"}}),
			"status":   "sent",
		})
		if err != nil {
			writeLINEError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if !includeSentMessages {
			writeEmptyJSON(w, status)
			return
		}
		httpx.WriteJSON(w, status, map[string]any{"sentMessages": []map[string]any{{"id": resource.ID, "quoteToken": "mockport_line_quote_token"}}})
	}
}

func (r *routes) writeNarrowcastProgress(w http.ResponseWriter) {
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"phase":         "succeeded",
		"successCount":  len(r.store.List("line", "message")),
		"failureCount":  0,
		"targetCount":   len(r.store.List("line", "message")),
		"acceptedTime":  "2999-01-01T00:00:00.000Z",
		"completedTime": "2999-01-01T00:00:01.000Z",
	})
}

func (r *routes) writeMessagingProfile(w http.ResponseWriter, userID string) {
	if normalizeScenario(r.cfg.Scenario) == "auth_error" {
		writeLINEError(w, http.StatusUnauthorized, "Mockport simulated invalid channel access token")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, lineProfile(userID))
}

func (r *routes) writeSetWebhookEndpoint(w http.ResponseWriter, req *http.Request) {
	payload, err := decodePayload(req)
	if err != nil {
		writeDecodeError(w, err)
		return
	}
	endpoint, _ := payload["endpoint"].(string)
	if !strings.HasPrefix(endpoint, "https://") || len(endpoint) > 500 {
		writeLINEError(w, http.StatusBadRequest, "Invalid webhook endpoint URL")
		return
	}
	r.setWebhookEndpoint(endpoint)
	writeEmptyJSON(w, http.StatusOK)
}

func (r *routes) writeGetWebhookEndpoint(w http.ResponseWriter) {
	endpoint := r.getWebhookEndpoint()
	if endpoint == "" {
		writeLINEError(w, http.StatusNotFound, "Webhook endpoint not found")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"endpoint": endpoint, "active": true})
}

func (r *routes) writeWebhookTest(w http.ResponseWriter) {
	if normalizeScenario(r.cfg.Scenario) == "auth_error" {
		writeLINEError(w, http.StatusUnauthorized, "Mockport simulated invalid channel access token")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"success": true, "timestamp": "2999-01-01T00:00:00Z", "statusCode": 200, "reason": "OK", "detail": "200"})
}

func (r *routes) sendWebhook(w http.ResponseWriter, req *http.Request) {
	if !security.IsLoopbackRemoteAddr(req.RemoteAddr) {
		writeLINEError(w, http.StatusForbidden, "webhook delivery can only be triggered from loopback")
		return
	}
	if r.cfg.WebhookTargetURL == "" {
		writeLINEError(w, http.StatusBadRequest, "webhook target URL is not configured")
		return
	}
	if !security.IsSafeWebhookTargetURL(r.cfg.WebhookTargetURL) {
		writeLINEError(w, http.StatusBadRequest, "webhook target URL must be a local Mockport target")
		return
	}
	payload, err := decodePayload(req)
	if err != nil {
		writeDecodeError(w, err)
		return
	}
	destination, _ := payload["destination"].(string)
	if destination == "" {
		destination = "U00000000000000000000000000000000"
	}
	events, _ := payload["events"].([]any)
	if len(events) == 0 {
		events = []any{map[string]any{
			"type":       "message",
			"replyToken": "mockport_line_reply_token",
			"message": map[string]any{
				"type": "text",
				"id":   "mockport_line_message",
				"text": "Mockport LINE webhook message",
			},
		}}
	}
	for i, event := range events {
		eventObject, ok := event.(map[string]any)
		if !ok {
			writeValidationDetails(w, []lineErrorDetail{{Message: "Must be a JSON object", Property: fmt.Sprintf("events[%d]", i)}})
			return
		}
		if eventObject["timestamp"] == nil {
			eventObject["timestamp"] = int64(4102444800000)
		}
		if eventObject["source"] == nil {
			eventObject["source"] = map[string]any{"type": "user", "userId": "Umockport"}
		}
		if eventObject["mode"] == nil {
			eventObject["mode"] = "active"
		}
		if eventObject["webhookEventId"] == nil {
			eventObject["webhookEventId"] = fmt.Sprintf("01MOCKPORTLINEEVENT%02d", i+1)
		}
		if eventObject["deliveryContext"] == nil {
			eventObject["deliveryContext"] = map[string]any{"isRedelivery": false}
		}
	}
	body, err := json.Marshal(map[string]any{"destination": destination, "events": events})
	if err != nil {
		writeLINEError(w, http.StatusInternalServerError, "failed to encode webhook payload")
		return
	}
	outbound, err := http.NewRequestWithContext(req.Context(), http.MethodPost, r.cfg.WebhookTargetURL, bytes.NewReader(body))
	if err != nil {
		writeLINEError(w, http.StatusBadRequest, "webhook target URL is invalid")
		return
	}
	secret := r.cfg.WebhookSigningSecret
	if secret == "" {
		secret = "mockport_line_secret"
	}
	outbound.Header.Set("Content-Type", "application/json")
	outbound.Header.Set("x-line-signature", signWebhookPayload(secret, body))
	resp, err := (&http.Client{Timeout: 5 * time.Second}).Do(outbound)
	if err != nil {
		writeLINEError(w, http.StatusBadGateway, "failed to send webhook")
		return
	}
	defer resp.Body.Close()
	httpx.WriteJSON(w, http.StatusAccepted, map[string]any{
		"sent":        true,
		"target_url":  r.cfg.WebhookTargetURL,
		"event_count": len(events),
		"status_code": resp.StatusCode,
	})
}

func (r *routes) writeContentEndpoint(w http.ResponseWriter, path string) {
	switch {
	case strings.HasSuffix(path, "/content/transcoding"):
		httpx.WriteJSON(w, http.StatusOK, map[string]any{"status": "succeeded"})
	case strings.HasSuffix(path, "/content/preview"):
		w.Header().Set("Content-Type", "image/png")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("mockport line preview image"))
	case strings.HasSuffix(path, "/content"):
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("mockport line message content"))
	default:
		writeLINEError(w, http.StatusNotFound, "not found")
	}
}

func (r *routes) writeAuthorize(w http.ResponseWriter, req *http.Request) {
	redirectURI := req.URL.Query().Get("redirect_uri")
	if redirectURI == "" {
		redirectURI = "http://localhost/callback"
	}
	if !security.IsSafeOAuthRedirectURL(redirectURI) {
		writeOAuthError(w, http.StatusBadRequest, "invalid_request", "redirect_uri must be a loopback URL in Mockport")
		return
	}
	if normalizeScenario(r.cfg.Scenario) == "invalid_request" {
		redirectWithQuery(w, req, redirectURI, map[string]string{
			"error":             "invalid_request",
			"error_description": "Mockport simulated invalid LINE Login request",
			"state":             req.URL.Query().Get("state"),
		})
		return
	}
	scope := req.URL.Query().Get("scope")
	if scope == "" {
		scope = "profile openid"
	}
	resource, err := r.store.Create("line", "oauth_code", map[string]any{
		"client_id":    req.URL.Query().Get("client_id"),
		"redirect_uri": redirectURI,
		"scope":        scope,
		"user_id":      "Umockport",
		"expires_at":   "2999-01-01T00:00:00Z",
	})
	if err != nil {
		writeLINEError(w, http.StatusInternalServerError, err.Error())
		return
	}
	redirectWithQuery(w, req, redirectURI, map[string]string{"code": resource.ID, "state": req.URL.Query().Get("state")})
}

func (r *routes) writeToken(w http.ResponseWriter, req *http.Request) {
	switch normalizeScenario(r.cfg.Scenario) {
	case "auth_error":
		writeOAuthError(w, http.StatusUnauthorized, "invalid_client", "Mockport simulated invalid LINE Login client")
	case "invalid_request":
		writeOAuthError(w, http.StatusBadRequest, "invalid_request", "Mockport simulated invalid token request")
	default:
		if err := req.ParseForm(); err != nil {
			writeOAuthError(w, http.StatusBadRequest, "invalid_request", "Request body must be form encoded")
			return
		}
		code := req.Form.Get("code")
		if code == "" {
			writeOAuthError(w, http.StatusBadRequest, "invalid_grant", "Authorization code is invalid or expired")
			return
		}
		codeResource, ok := r.store.Get("line", "oauth_code", code)
		if !ok {
			writeOAuthError(w, http.StatusBadRequest, "invalid_grant", "Authorization code is invalid or expired")
			return
		}
		if want, _ := codeResource.Data["redirect_uri"].(string); want != "" && req.Form.Get("redirect_uri") != "" && req.Form.Get("redirect_uri") != want {
			writeOAuthError(w, http.StatusBadRequest, "invalid_request", "redirect_uri does not match")
			return
		}
		if !clientIDMatches(codeResource, req.Form.Get("client_id")) {
			writeOAuthError(w, http.StatusUnauthorized, "invalid_client", "client_id does not match authorization request")
			return
		}
		codeResource, ok = r.store.Take("line", "oauth_code", code)
		if !ok {
			writeOAuthError(w, http.StatusBadRequest, "invalid_grant", "Authorization code is invalid or expired")
			return
		}
		token, err := r.store.Create("line", "oauth_token", map[string]any{
			"scope":      codeResource.Data["scope"],
			"user_id":    codeResource.Data["user_id"],
			"expires_at": "2999-01-01T00:00:00Z",
		})
		if err != nil {
			writeOAuthError(w, http.StatusInternalServerError, "server_error", err.Error())
			return
		}
		body := map[string]any{
			"access_token":  token.ID,
			"expires_in":    2592000,
			"id_token":      "mockport.line.id-token",
			"refresh_token": "mockport_line_refresh_token",
			"scope":         codeResource.Data["scope"],
			"token_type":    "Bearer",
		}
		if req.Form.Get("grant_type") == "client_credentials" {
			body["key_id"] = "mockport_line_key_id"
		}
		httpx.WriteJSON(w, http.StatusOK, body)
	}
}

func (r *routes) writeStatelessToken(w http.ResponseWriter, req *http.Request) {
	if normalizeScenario(r.cfg.Scenario) == "auth_error" {
		writeOAuthError(w, http.StatusBadRequest, "invalid_request", "Invalid 'client_credentials'.")
		return
	}
	if err := req.ParseForm(); err != nil {
		writeOAuthError(w, http.StatusBadRequest, "invalid_request", "Request body must be form encoded")
		return
	}
	token, err := r.store.Create("line", "oauth_token", map[string]any{
		"scope":      "profile chat_message.write",
		"user_id":    "Umockport",
		"client_id":  firstNonEmpty(req.Form.Get("client_id"), "mockport_line_channel"),
		"expires_in": 900,
	})
	if err != nil {
		writeOAuthError(w, http.StatusInternalServerError, "server_error", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"token_type": "Bearer", "access_token": token.ID, "expires_in": 900})
}

func (r *routes) writeShortLivedToken(w http.ResponseWriter, req *http.Request) {
	if normalizeScenario(r.cfg.Scenario) == "auth_error" {
		writeOAuthError(w, http.StatusBadRequest, "invalid_client", "invalid client_secret")
		return
	}
	if err := req.ParseForm(); err != nil {
		writeOAuthError(w, http.StatusBadRequest, "invalid_request", "Request body must be form encoded")
		return
	}
	token, err := r.store.Create("line", "oauth_token", map[string]any{
		"scope":      "profile chat_message.write",
		"user_id":    "Umockport",
		"client_id":  firstNonEmpty(req.Form.Get("client_id"), "mockport_line_channel"),
		"expires_in": 2592000,
	})
	if err != nil {
		writeOAuthError(w, http.StatusInternalServerError, "server_error", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"access_token": token.ID, "expires_in": 2592000, "token_type": "Bearer"})
}

func (r *routes) writeVerifyToken(w http.ResponseWriter, req *http.Request) {
	token := req.URL.Query().Get("access_token")
	if token == "" {
		writeOAuthError(w, http.StatusBadRequest, "invalid_request", "The access token expired")
		return
	}
	resource, ok := r.store.Get("line", "oauth_token", token)
	if !ok {
		writeOAuthError(w, http.StatusBadRequest, "invalid_request", "The access token expired")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"client_id":  firstNonEmpty(resource.Data["client_id"], "mockport_line_channel"),
		"expires_in": firstNonEmpty(resource.Data["expires_in"], 2592000),
		"scope":      firstNonEmpty(resource.Data["scope"], "profile chat_message.write"),
	})
}

func (r *routes) writeVerifyOAuthToken(w http.ResponseWriter, req *http.Request) {
	if err := req.ParseForm(); err != nil {
		writeOAuthError(w, http.StatusBadRequest, "invalid_request", "Request body must be form encoded")
		return
	}
	token := req.Form.Get("access_token")
	if token == "" {
		writeOAuthError(w, http.StatusBadRequest, "invalid_request", "access_token invalid")
		return
	}
	resource, ok := r.store.Get("line", "oauth_token", token)
	if !ok {
		writeOAuthError(w, http.StatusBadRequest, "invalid_request", "access_token invalid")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"client_id":  firstNonEmpty(resource.Data["client_id"], "mockport_line_channel"),
		"expires_in": firstNonEmpty(resource.Data["expires_in"], 2592000),
		"scope":      firstNonEmpty(resource.Data["scope"], "profile chat_message.write"),
	})
}

func (r *routes) writeLoginProfile(w http.ResponseWriter, req *http.Request) {
	if normalizeScenario(r.cfg.Scenario) == "auth_error" {
		writeLINEError(w, http.StatusUnauthorized, "Mockport simulated invalid access token")
		return
	}
	resource, ok := r.tokenResource(req)
	if !ok {
		writeLINEError(w, http.StatusUnauthorized, "Access token is invalid")
		return
	}
	userID, _ := resource.Data["user_id"].(string)
	httpx.WriteJSON(w, http.StatusOK, lineProfile(userID))
}

func (r *routes) writeLIFFProfile(w http.ResponseWriter) {
	if normalizeScenario(r.cfg.Scenario) == "auth_error" {
		writeLINEError(w, http.StatusUnauthorized, "Mockport simulated LIFF access token error")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, lineProfile("Umockport"))
}

func (r *routes) writeLIFFContext(w http.ResponseWriter) {
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"type":        "utou",
		"userId":      "Umockport",
		"viewType":    "full",
		"endpointUrl": "http://localhost:3000",
		"liffId":      "mockport-line-liff",
	})
}

func (r *routes) writeBotInfo(w http.ResponseWriter) {
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"userId":         "U00000000000000000000000000000000",
		"basicId":        "@mockport",
		"displayName":    "Mockport LINE Official Account",
		"pictureUrl":     "https://example.test/mockport-line-bot.png",
		"chatMode":       "bot",
		"markAsReadMode": "auto",
	})
}

func (r *routes) writeGroupEndpoint(w http.ResponseWriter, path string) {
	rest := strings.TrimPrefix(path, "/v2/bot/group/")
	parts := strings.Split(rest, "/")
	if len(parts) < 2 {
		writeLINEError(w, http.StatusBadRequest, "The value for the 'groupId' parameter is invalid")
		return
	}
	groupID := parts[0]
	switch {
	case len(parts) == 2 && parts[1] == "summary":
		httpx.WriteJSON(w, http.StatusOK, map[string]any{"groupId": groupID, "groupName": "Mockport Group", "pictureUrl": "https://example.test/mockport-line-group.png"})
	case len(parts) == 3 && parts[1] == "members" && parts[2] == "count":
		httpx.WriteJSON(w, http.StatusOK, map[string]any{"count": 3})
	case len(parts) == 3 && parts[1] == "members" && parts[2] == "ids":
		httpx.WriteJSON(w, http.StatusOK, map[string]any{"memberIds": []string{"Umockport", "Ulocaluser"}})
	case len(parts) == 3 && parts[1] == "member":
		httpx.WriteJSON(w, http.StatusOK, lineProfile(parts[2]))
	default:
		writeLINEError(w, http.StatusNotFound, "Not found")
	}
}

func (r *routes) writeRoomEndpoint(w http.ResponseWriter, path string) {
	rest := strings.TrimPrefix(path, "/v2/bot/room/")
	parts := strings.Split(rest, "/")
	if len(parts) < 2 {
		writeLINEError(w, http.StatusBadRequest, "The value for the 'roomId' parameter is invalid")
		return
	}
	switch {
	case len(parts) == 3 && parts[1] == "members" && parts[2] == "count":
		httpx.WriteJSON(w, http.StatusOK, map[string]any{"count": 3})
	case len(parts) == 3 && parts[1] == "members" && parts[2] == "ids":
		httpx.WriteJSON(w, http.StatusOK, map[string]any{"memberIds": []string{"Umockport", "Ulocaluser"}})
	case len(parts) == 3 && parts[1] == "member":
		httpx.WriteJSON(w, http.StatusOK, lineProfile(parts[2]))
	default:
		writeLINEError(w, http.StatusNotFound, "Not found")
	}
}

func (r *routes) writeValidateRichMenu(w http.ResponseWriter, req *http.Request) {
	payload, err := decodePayload(req)
	if err != nil {
		writeDecodeError(w, err)
		return
	}
	if _, ok := payload["name"].(string); !ok {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"message": "The request body has 1 error(s)",
			"details": []map[string]string{{
				"message":  "must be specified",
				"property": "name",
			}},
		})
		return
	}
	writeEmptyJSON(w, http.StatusOK)
}

func (r *routes) writeCreateRichMenu(w http.ResponseWriter, req *http.Request) {
	payload, err := decodePayload(req)
	if err != nil {
		writeDecodeError(w, err)
		return
	}
	if _, ok := payload["name"].(string); !ok {
		writeLINEError(w, http.StatusBadRequest, "The request body has 1 error(s)")
		return
	}
	resource, err := r.store.Create("line", "rich_menu", payload)
	if err != nil {
		writeLINEError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"richMenuId": resource.ID})
}

func (r *routes) writeRichMenuList(w http.ResponseWriter) {
	menus := []map[string]any{}
	for _, resource := range r.store.List("line", "rich_menu") {
		menus = append(menus, richMenuResponse(resource))
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"richmenus": menus})
}

func (r *routes) writeGetRichMenu(w http.ResponseWriter, id string) {
	resource, ok := r.store.Get("line", "rich_menu", id)
	if !ok {
		writeLINEError(w, http.StatusNotFound, "Not found")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, richMenuResponse(resource))
}

func (r *routes) writeDeleteRichMenu(w http.ResponseWriter, id string) {
	if !r.deleteRichMenu(id) {
		writeLINEError(w, http.StatusNotFound, "Not found")
		return
	}
	writeEmptyJSON(w, http.StatusOK)
}

func (r *routes) writeUploadRichMenuImage(w http.ResponseWriter, path string) {
	id := strings.TrimSuffix(strings.TrimPrefix(path, "/v2/bot/richmenu/"), "/content")
	if _, ok := r.store.Get("line", "rich_menu", id); !ok {
		writeLINEError(w, http.StatusNotFound, "Not found")
		return
	}
	if _, err := r.store.Update("line", "rich_menu", id, map[string]any{"hasImage": true}); err != nil {
		writeLINEError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeEmptyJSON(w, http.StatusOK)
}

func (r *routes) writeDownloadRichMenuImage(w http.ResponseWriter, path string) {
	id := strings.TrimSuffix(strings.TrimPrefix(path, "/v2/bot/richmenu/"), "/content")
	resource, ok := r.store.Get("line", "rich_menu", id)
	if !ok || resource.Data["hasImage"] != true {
		writeLINEError(w, http.StatusNotFound, "Not found")
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("mockport rich menu image"))
}

func (r *routes) writeSetDefaultRichMenu(w http.ResponseWriter, id string) {
	if status, message := r.setDefaultRichMenuChecked(id); status != http.StatusOK {
		writeLINEError(w, status, message)
		return
	}
	writeEmptyJSON(w, http.StatusOK)
}

func (r *routes) writeGetDefaultRichMenu(w http.ResponseWriter) {
	id, ok := r.currentDefaultRichMenu()
	if !ok {
		writeLINEError(w, http.StatusNotFound, "no default richmenu")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"richMenuId": id})
}

func (r *routes) writeLinkRichMenuToUser(w http.ResponseWriter, path string) {
	parts := strings.Split(strings.TrimPrefix(path, "/v2/bot/user/"), "/richmenu/")
	if len(parts) != 2 {
		writeLINEError(w, http.StatusBadRequest, "The value for the 'userId' parameter is invalid")
		return
	}
	if _, ok := r.store.Get("line", "rich_menu", parts[1]); !ok {
		writeLINEError(w, http.StatusNotFound, "Not found")
		return
	}
	_, err := r.store.Create("line", "user_rich_menu", map[string]any{"userId": parts[0], "richMenuId": parts[1]})
	if err != nil {
		writeLINEError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeEmptyJSON(w, http.StatusOK)
}

func (r *routes) writeGetUserRichMenu(w http.ResponseWriter, path string) {
	userID := strings.TrimSuffix(strings.TrimPrefix(path, "/v2/bot/user/"), "/richmenu")
	for _, resource := range r.store.List("line", "user_rich_menu") {
		if resource.Data["userId"] == userID {
			httpx.WriteJSON(w, http.StatusOK, map[string]any{"richMenuId": resource.Data["richMenuId"]})
			return
		}
	}
	writeLINEError(w, http.StatusNotFound, "the user has no richmenu")
}

func (r *routes) writeUnlinkRichMenuFromUser(w http.ResponseWriter, path string) {
	userID := strings.TrimSuffix(strings.TrimPrefix(path, "/v2/bot/user/"), "/richmenu")
	for _, resource := range r.store.List("line", "user_rich_menu") {
		if resource.Data["userId"] == userID {
			r.store.Delete("line", "user_rich_menu", resource.ID)
		}
	}
	writeEmptyJSON(w, http.StatusOK)
}

func (r *routes) writeCreateRichMenuAlias(w http.ResponseWriter, req *http.Request) {
	payload, err := decodePayload(req)
	if err != nil {
		writeDecodeError(w, err)
		return
	}
	aliasID, _ := payload["richMenuAliasId"].(string)
	richMenuID, _ := payload["richMenuId"].(string)
	if aliasID == "" || richMenuID == "" {
		writeLINEError(w, http.StatusBadRequest, "The request body has 1 error(s)")
		return
	}
	if _, ok := r.store.Get("line", "rich_menu", richMenuID); !ok {
		writeLINEError(w, http.StatusBadRequest, "richmenu not found")
		return
	}
	if _, ok := r.findAlias(aliasID); ok {
		writeLINEError(w, http.StatusBadRequest, "conflict richmenu alias id")
		return
	}
	if _, err := r.store.Create("line", "rich_menu_alias", map[string]any{"richMenuAliasId": aliasID, "richMenuId": richMenuID}); err != nil {
		writeLINEError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeEmptyJSON(w, http.StatusOK)
}

func (r *routes) writeUpdateRichMenuAlias(w http.ResponseWriter, req *http.Request, aliasID string) {
	resource, ok := r.findAlias(aliasID)
	if !ok {
		writeLINEError(w, http.StatusNotFound, "richmenu alias not found")
		return
	}
	payload, err := decodePayload(req)
	if err != nil {
		writeDecodeError(w, err)
		return
	}
	richMenuID, _ := payload["richMenuId"].(string)
	if _, ok := r.store.Get("line", "rich_menu", richMenuID); !ok {
		writeLINEError(w, http.StatusBadRequest, "richmenu not found")
		return
	}
	if _, err := r.store.Update("line", "rich_menu_alias", resource.ID, map[string]any{"richMenuId": richMenuID}); err != nil {
		writeLINEError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeEmptyJSON(w, http.StatusOK)
}

func (r *routes) writeDeleteRichMenuAlias(w http.ResponseWriter, aliasID string) {
	resource, ok := r.findAlias(aliasID)
	if !ok {
		writeLINEError(w, http.StatusNotFound, "richmenu alias not found")
		return
	}
	r.store.Delete("line", "rich_menu_alias", resource.ID)
	writeEmptyJSON(w, http.StatusOK)
}

func (r *routes) writeGetRichMenuAlias(w http.ResponseWriter, aliasID string) {
	resource, ok := r.findAlias(aliasID)
	if !ok {
		writeLINEError(w, http.StatusNotFound, "richmenu alias not found")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"richMenuAliasId": resource.Data["richMenuAliasId"], "richMenuId": resource.Data["richMenuId"]})
}

func (r *routes) writeRichMenuAliasList(w http.ResponseWriter) {
	aliases := []map[string]any{}
	for _, resource := range r.store.List("line", "rich_menu_alias") {
		aliases = append(aliases, map[string]any{"richMenuAliasId": resource.Data["richMenuAliasId"], "richMenuId": resource.Data["richMenuId"]})
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"aliases": aliases})
}

func (r *routes) findAlias(aliasID string) (state.Resource, bool) {
	for _, resource := range r.store.List("line", "rich_menu_alias") {
		if resource.Data["richMenuAliasId"] == aliasID {
			return resource, true
		}
	}
	return state.Resource{}, false
}

func (r *routes) writeNotificationToken(w http.ResponseWriter, req *http.Request) {
	switch normalizeScenario(r.cfg.Scenario) {
	case "auth_error":
		writeLINEError(w, http.StatusUnauthorized, "Mockport simulated invalid MINI App channel token")
	case "invalid_request":
		writeLINEError(w, http.StatusBadRequest, "[liffAccessToken] must not be blank")
	default:
		payload, err := decodePayload(req)
		if err != nil {
			writeDecodeError(w, err)
			return
		}
		if len(payload) > 0 && strings.TrimSpace(fmt.Sprint(payload["liffAccessToken"])) == "" {
			writeLINEError(w, http.StatusBadRequest, "[liffAccessToken] must not be blank")
			return
		}
		resource, err := r.store.Create("line", "notification_token", map[string]any{
			"remaining_count": 5,
			"session_id":      "line_service_session_mockport",
		})
		if err != nil {
			writeLINEError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeNotification(w, resource.ID, 5)
	}
}

func (r *routes) writeServiceMessage(w http.ResponseWriter, req *http.Request) {
	switch normalizeScenario(r.cfg.Scenario) {
	case "auth_error":
		writeLINEError(w, http.StatusUnauthorized, "Mockport simulated invalid MINI App channel token")
	case "invalid_request":
		writeLINEError(w, http.StatusBadRequest, "templateName is required")
	default:
		payload, err := decodePayload(req)
		if err != nil {
			writeDecodeError(w, err)
			return
		}
		token, _ := payload["notificationToken"].(string)
		if token != "" {
			if _, ok := r.store.Get("line", "notification_token", token); !ok {
				writeLINEError(w, http.StatusBadRequest, "notificationToken is invalid")
				return
			}
		}
		resource, err := r.store.Create("line", "notification_token", map[string]any{
			"remaining_count": 4,
			"session_id":      "line_service_session_mockport",
		})
		if err != nil {
			writeLINEError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeNotification(w, resource.ID, 4)
	}
}

func (r *routes) writePayRequest(w http.ResponseWriter, req *http.Request) {
	switch normalizeScenario(r.cfg.Scenario) {
	case "auth_error":
		writePayError(w, "1104", "Mockport simulated LINE Pay authorization error")
	case "pay_failed":
		writePayError(w, "1169", "LINE Pay requires payment method selection and password authentication.")
	case "invalid_request":
		writePayError(w, "2101", "Mockport simulated invalid LINE Pay request")
	default:
		payload, err := decodePayload(req)
		if err != nil {
			writePayError(w, "2101", "Request body must be JSON")
			return
		}
		orderID := fmt.Sprint(firstNonEmpty(payload["orderId"], "order_mockport"))
		amount := firstNonEmpty(payload["amount"], 1000)
		currency := fmt.Sprint(firstNonEmpty(payload["currency"], "JPY"))
		resource, err := r.store.Create("line", "line_pay_payment", map[string]any{
			"orderId":  orderID,
			"amount":   amount,
			"currency": currency,
			"status":   "reserved",
		})
		if err != nil {
			writePayError(w, "9000", err.Error())
			return
		}
		httpx.WriteJSON(w, http.StatusOK, map[string]any{
			"returnCode":    "0000",
			"returnMessage": "Success.",
			"info": map[string]any{
				"transactionId":      resource.ID,
				"paymentAccessToken": "mockport_line_pay_access_token",
				"paymentUrl": map[string]string{
					"web": "http://localhost:43101" + r.basePath + "/line-pay/authorize/" + resource.ID,
					"app": "line://pay/mockport/" + resource.ID,
				},
			},
		})
	}
}

func (r *routes) writePayConfirm(w http.ResponseWriter, id string) {
	switch normalizeScenario(r.cfg.Scenario) {
	case "auth_error":
		writePayError(w, "1104", "Mockport simulated LINE Pay authorization error")
	case "pay_failed":
		writePayError(w, "1169", "LINE Pay requires payment method selection and password authentication.")
	default:
		resource, ok := r.store.Get("line", "line_pay_payment", id)
		if !ok {
			writePayError(w, "1150", "Transaction not found")
			return
		}
		updated, err := r.store.Update("line", "line_pay_payment", id, map[string]any{"status": "confirmed"})
		if err != nil {
			writePayError(w, "9000", err.Error())
			return
		}
		httpx.WriteJSON(w, http.StatusOK, payInfoResponse("0000", "Success.", updated.ID, resource.Data["orderId"], updated.Data["amount"], updated.Data["currency"], updated.Data["status"]))
	}
}

func (r *routes) writePayCheck(w http.ResponseWriter, id string) {
	resource, ok := r.store.Get("line", "line_pay_payment", id)
	if !ok {
		writePayError(w, "1150", "Transaction not found")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, payInfoResponse("0000", "Success.", resource.ID, resource.Data["orderId"], resource.Data["amount"], resource.Data["currency"], resource.Data["status"]))
}

func (r *routes) writeMiniDappWalletSession(w http.ResponseWriter, req *http.Request) {
	if normalizeScenario(r.cfg.Scenario) == "auth_error" {
		writeLINEError(w, http.StatusUnauthorized, "Mockport simulated Mini Dapp client authorization error")
		return
	}
	if normalizeScenario(r.cfg.Scenario) == "invalid_request" {
		writeLINEError(w, http.StatusBadRequest, "chainId is required")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"sessionId":     "line_mini_dapp_wallet_session_mockport",
		"chainId":       1001,
		"walletAddress": "0x0000000000000000000000000000000000431010",
		"status":        "connected",
	})
}

func (r *routes) writeMiniDappPayment(w http.ResponseWriter, req *http.Request) {
	switch normalizeScenario(r.cfg.Scenario) {
	case "auth_error":
		writeLINEError(w, http.StatusUnauthorized, "Mockport simulated Mini Dapp client authorization error")
	case "pay_failed":
		writeLINEError(w, http.StatusPaymentRequired, "Mockport simulated Mini Dapp payment failure")
	case "invalid_request":
		writeLINEError(w, http.StatusBadRequest, "itemId is required")
	default:
		payload, err := decodePayload(req)
		if err != nil {
			writeDecodeError(w, err)
			return
		}
		resource, err := r.store.Create("line", "mini_dapp_payment", map[string]any{
			"itemId":   firstNonEmpty(payload["itemId"], "item_mockport"),
			"amount":   firstNonEmpty(payload["amount"], "10"),
			"currency": firstNonEmpty(payload["currency"], "KAIA"),
			"status":   "approved",
		})
		if err != nil {
			writeLINEError(w, http.StatusInternalServerError, err.Error())
			return
		}
		body := resource.Data
		body["id"] = resource.ID
		body["checkoutUrl"] = "http://localhost:43101" + r.basePath + "/mini-dapp/checkout/" + resource.ID
		httpx.WriteJSON(w, http.StatusOK, body)
	}
}

func (r *routes) writeMiniDappPaymentLookup(w http.ResponseWriter, id string) {
	resource, ok := r.store.Get("line", "mini_dapp_payment", id)
	if !ok {
		writeLINEError(w, http.StatusNotFound, "Mini Dapp payment not found")
		return
	}
	body := resource.Data
	body["id"] = resource.ID
	httpx.WriteJSON(w, http.StatusOK, body)
}

func (r *routes) tokenResource(req *http.Request) (state.Resource, bool) {
	token := bearerToken(req)
	if token == "" {
		return state.Resource{}, false
	}
	return r.store.Get("line", "oauth_token", token)
}

func decodePayload(req *http.Request) (map[string]any, error) {
	if req.Body == nil {
		return map[string]any{}, nil
	}
	defer req.Body.Close()
	var payload map[string]any
	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&payload); err != nil {
		if err == io.EOF {
			return map[string]any{}, nil
		}
		return nil, err
	}
	var trailing json.RawMessage
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return nil, fmt.Errorf("unexpected trailing JSON value")
		}
		return nil, err
	}
	if payload == nil {
		payload = map[string]any{}
	}
	return payload, nil
}

func writeDecodeError(w http.ResponseWriter, err error) {
	if httpx.IsRequestBodyTooLarge(err) {
		writeLINEError(w, http.StatusRequestEntityTooLarge, "Request body is too large")
		return
	}
	writeLINEError(w, http.StatusBadRequest, "Request body must be JSON")
}

type lineErrorDetail struct {
	Message  string `json:"message"`
	Property string `json:"property"`
}

func writeValidationDetails(w http.ResponseWriter, details []lineErrorDetail) {
	httpx.WriteJSON(w, http.StatusBadRequest, struct {
		Message string            `json:"message"`
		Details []lineErrorDetail `json:"details"`
	}{
		Message: fmt.Sprintf("The request body has %d error(s)", len(details)),
		Details: details,
	})
}

type lineProfileResponse struct {
	UserID        string `json:"userId"`
	DisplayName   string `json:"displayName"`
	PictureURL    string `json:"pictureUrl"`
	StatusMessage string `json:"statusMessage"`
}

type linePayResponse struct {
	ReturnCode    string      `json:"returnCode"`
	ReturnMessage string      `json:"returnMessage"`
	Info          linePayInfo `json:"info"`
}

type linePayInfo struct {
	TransactionID string `json:"transactionId"`
	OrderID       any    `json:"orderId"`
	Amount        any    `json:"amount"`
	Currency      any    `json:"currency"`
	Status        any    `json:"status"`
}

func lineProfile(userID string) lineProfileResponse {
	if userID == "" {
		userID = "Umockport"
	}
	return lineProfileResponse{
		UserID:        userID,
		DisplayName:   "Mockport LINE User",
		PictureURL:    "https://example.test/mockport-line-user.png",
		StatusMessage: "Mockport local LINE profile",
	}
}

func writeLINEError(w http.ResponseWriter, status int, message string) {
	httpx.WriteJSON(w, status, map[string]any{"message": message})
}

func signWebhookPayload(secret string, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(payload)
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func writeOAuthError(w http.ResponseWriter, status int, code, description string) {
	httpx.WriteJSON(w, status, map[string]any{"error": code, "error_description": description})
}

func writePayError(w http.ResponseWriter, code, message string) {
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"returnCode": code, "returnMessage": message})
}

func writeNotification(w http.ResponseWriter, token string, remaining int) {
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"notificationToken": token,
		"expiresIn":         31536000,
		"remainingCount":    remaining,
		"sessionId":         "line_service_session_mockport",
	})
}

func payInfoResponse(code, message, transactionID string, orderID, amount, currency, status any) linePayResponse {
	return linePayResponse{
		ReturnCode:    code,
		ReturnMessage: message,
		Info: linePayInfo{
			TransactionID: transactionID,
			OrderID:       orderID,
			Amount:        amount,
			Currency:      currency,
			Status:        status,
		},
	}
}

type richMenuData map[string]any

func richMenuResponse(resource state.Resource) richMenuData {
	body := make(richMenuData, len(resource.Data)+1)
	for key, value := range resource.Data {
		body[key] = value
	}
	body["richMenuId"] = resource.ID
	delete(body, "hasImage")
	return body
}

func writeEmptyJSON(w http.ResponseWriter, status int) {
	httpx.WriteJSON(w, status, map[string]any{})
}

func writeNoBody(w http.ResponseWriter, status int) {
	w.WriteHeader(status)
}

func redirectWithQuery(w http.ResponseWriter, req *http.Request, redirectURI string, values map[string]string) {
	target, err := url.Parse(redirectURI)
	if err != nil {
		http.Error(w, "invalid redirect_uri", http.StatusBadRequest)
		return
	}
	query := target.Query()
	for key, value := range values {
		if value != "" {
			query.Set(key, value)
		}
	}
	target.RawQuery = query.Encode()
	http.Redirect(w, req, target.String(), http.StatusFound)
}

func bearerToken(r *http.Request) string {
	value := r.Header.Get("Authorization")
	if !strings.HasPrefix(value, "Bearer ") {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(value, "Bearer "))
}

func normalizeScenario(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "line_success"
	}
	return value
}

func firstNonEmpty(values ...any) any {
	for _, value := range values {
		switch typed := value.(type) {
		case nil:
			continue
		case string:
			if strings.TrimSpace(typed) == "" {
				continue
			}
		}
		return value
	}
	return nil
}

func clientIDMatches(resource state.Resource, got string) bool {
	want, _ := resource.Data["client_id"].(string)
	return want == "" || got == want
}
