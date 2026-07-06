package line

import (
	"net/http"
	"strings"

	"github.com/albert-einshutoin/mockport/internal/adapter/httpx"
	"github.com/albert-einshutoin/mockport/internal/security"
)

func (r *routes) handle(w http.ResponseWriter, req *http.Request) {
	httpx.LimitRequestBody(w, req)
	w.Header().Set("X-Line-Request-Id", "line-request-mockport")
	path := strings.TrimPrefix(req.URL.Path, r.basePath)
	// Validate X-Mockport-Scenario once at the dispatch layer so every LINE
	// endpoint rejects unknown scenario names with a 400 instead of silently
	// falling through to a success path. The error body shape depends on the
	// path group (Messaging API vs OAuth vs LINE Pay), so pick the matching
	// resolver helper before routing to the concrete handler.
	if !r.validateScenarioForPath(w, req, path) {
		return
	}
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
		r.writeMessagingProfile(w, req, strings.TrimPrefix(path, "/v2/bot/profile/"))
	case req.Method == http.MethodPut && path == "/v2/bot/channel/webhook/endpoint":
		r.writeSetWebhookEndpoint(w, req)
	case req.Method == http.MethodGet && path == "/v2/bot/channel/webhook/endpoint":
		r.writeGetWebhookEndpoint(w)
	case req.Method == http.MethodPost && path == "/v2/bot/channel/webhook/test":
		r.writeWebhookTest(w, req)
	case req.Method == http.MethodPost && path == "/test/webhook/send":
		r.sendWebhook(w, req)
	case req.Method == http.MethodPost && path == "/test/reset":
		r.handleReset(w, req)
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
		r.writeLIFFProfile(w, req)
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
		r.writePayConfirm(w, req, id)
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

// validateScenarioForPath resolves the request scenario using the resolver
// helper that matches the path group's error format, returning false (after
// writing a 400) when the X-Mockport-Scenario header names an unknown scenario.
// Grouping mirrors the error shapes the concrete handlers already emit:
//   - LINE Pay (/v3/payments/*): resolveScenarioPay -> {"returnCode","returnMessage"}
//   - OAuth/Login (/oauth2/*, /v2/oauth/*): resolveScenarioOAuth -> {"error",...}
//   - everything else (Messaging API, LIFF, MINI App, Mini Dapp): resolveScenario -> {"message"}
//
// The /test/reset management endpoint only clears state and is exempt.
func (r *routes) validateScenarioForPath(w http.ResponseWriter, req *http.Request, path string) bool {
	switch {
	case path == "/test/reset":
		return true
	case strings.HasPrefix(path, "/v3/payments/"):
		_, ok := r.resolveScenarioPay(w, req)
		return ok
	case strings.HasPrefix(path, "/oauth2/") || strings.HasPrefix(path, "/v2/oauth/"):
		_, ok := r.resolveScenarioOAuth(w, req)
		return ok
	default:
		_, ok := r.resolveScenario(w, req)
		return ok
	}
}

func (r *routes) handleReset(w http.ResponseWriter, req *http.Request) {
	if !security.IsLoopbackRemoteAddr(req.RemoteAddr) {
		writeLINEError(w, http.StatusForbidden, "line reset can only be triggered from loopback")
		return
	}
	r.resetSingletonState()
	resourceTypes := r.store.ResetAll("line")
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"reset":          true,
		"adapter":        "line",
		"resource_types": resourceTypes,
	})
}
