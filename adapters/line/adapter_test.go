package line

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"github.com/albert-einshutoin/mockport/internal/adapter"
)

func TestMessagingPushAndProfile(t *testing.T) {
	mux := newLineMux(t, adapter.Config{BasePath: "/line", Scenario: "line_success"})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/line/v2/bot/message/push", strings.NewReader(`{"to":"Umockport","messages":[{"type":"text","text":"hello"}]}`))
	req.Header.Set("Content-Type", "application/json")
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("push status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var push map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &push); err != nil {
		t.Fatalf("decode push: %v", err)
	}
	if _, ok := push["sentMessages"].([]any); !ok {
		t.Fatalf("push response missing sentMessages: %#v", push)
	}

	profileRec := httptest.NewRecorder()
	mux.ServeHTTP(profileRec, httptest.NewRequest(http.MethodGet, "/line/v2/bot/profile/Umockport", nil))
	if profileRec.Code != http.StatusOK {
		t.Fatalf("profile status = %d, want %d", profileRec.Code, http.StatusOK)
	}
	var profile map[string]any
	if err := json.Unmarshal(profileRec.Body.Bytes(), &profile); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	if profile["userId"] != "Umockport" || profile["displayName"] == "" {
		t.Fatalf("profile = %#v", profile)
	}
}

func TestMessagingAPICoreEndpoints(t *testing.T) {
	mux := newLineMux(t, adapter.Config{BasePath: "/line", Scenario: "line_success"})

	cases := []struct {
		method string
		path   string
		body   string
		status int
		want   string
	}{
		{http.MethodPost, "/line/v2/bot/message/validate/push", `{"messages":[{"type":"text","text":"hello"}]}`, http.StatusOK, `{}`},
		{http.MethodPost, "/line/v2/bot/message/multicast", `{"to":["Umockport"],"messages":[{"type":"text","text":"hello"}]}`, http.StatusOK, `{}`},
		{http.MethodPost, "/line/v2/bot/message/broadcast", `{"messages":[{"type":"text","text":"hello"}]}`, http.StatusOK, `{}`},
		{http.MethodPost, "/line/v2/bot/message/narrowcast", `{"messages":[{"type":"text","text":"hello"}]}`, http.StatusAccepted, `{}`},
		{http.MethodPost, "/line/v2/bot/chat/markAsRead", `{"markAsReadToken":"mockport_read"}`, http.StatusOK, `{}`},
		{http.MethodPost, "/line/v2/bot/chat/loading/start", `{"chatId":"Umockport","loadingSeconds":5}`, http.StatusAccepted, `{}`},
		{http.MethodGet, "/line/v2/bot/info", ``, http.StatusOK, `"basicId":"@mockport"`},
		{http.MethodGet, "/line/v2/bot/message/quota", ``, http.StatusOK, `"type":"limited"`},
		{http.MethodGet, "/line/v2/bot/message/delivery/reply?date=20260528", ``, http.StatusOK, `"status":"ready"`},
		{http.MethodGet, "/line/v2/bot/followers/ids?limit=2", ``, http.StatusOK, `"userIds"`},
	}

	for _, tc := range cases {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			mux.ServeHTTP(rec, req)
			if rec.Code != tc.status {
				t.Fatalf("status = %d, want %d: %s", rec.Code, tc.status, rec.Body.String())
			}
			if !strings.Contains(rec.Body.String(), tc.want) {
				t.Fatalf("body = %s, want to contain %s", rec.Body.String(), tc.want)
			}
			if rec.Header().Get("X-Line-Request-Id") == "" {
				t.Fatalf("missing X-Line-Request-Id")
			}
		})
	}
}

func TestMessageValidationReturnsLINEStyleDetails(t *testing.T) {
	mux := newLineMux(t, adapter.Config{BasePath: "/line", Scenario: "line_success"})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/line/v2/bot/message/validate/push", strings.NewReader(`{"messages":[{"type":"text","text":""},{"type":"unknown"}]}`))
	req.Header.Set("Content-Type", "application/json")
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"message":"The request body has 2 error(s)"`) {
		t.Fatalf("body missing error count: %s", rec.Body.String())
	}
	for _, want := range []string{`"property":"messages[0].text"`, `"property":"messages[1].type"`} {
		if !strings.Contains(rec.Body.String(), want) {
			t.Fatalf("body missing %s: %s", want, rec.Body.String())
		}
	}
}

func TestMessageValidationRejectsTrailingJSONToken(t *testing.T) {
	mux := newLineMux(t, adapter.Config{BasePath: "/line", Scenario: "line_success"})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/line/v2/bot/message/validate/push", strings.NewReader(`{"messages":[{"type":"text","text":"hello"}]}{}`))
	req.Header.Set("Content-Type", "application/json")
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "Request body must be JSON") {
		t.Fatalf("trailing JSON token = status %d body %s", rec.Code, rec.Body.String())
	}
}

func TestWebhookEndpointSettingsAndContent(t *testing.T) {
	mux := newLineMux(t, adapter.Config{BasePath: "/line", Scenario: "line_success"})

	setRec := httptest.NewRecorder()
	setReq := httptest.NewRequest(http.MethodPut, "/line/v2/bot/channel/webhook/endpoint", strings.NewReader(`{"endpoint":"https://example.com/webhook"}`))
	setReq.Header.Set("Content-Type", "application/json")
	mux.ServeHTTP(setRec, setReq)
	if setRec.Code != http.StatusOK {
		t.Fatalf("set webhook status = %d, want %d: %s", setRec.Code, http.StatusOK, setRec.Body.String())
	}

	getRec := httptest.NewRecorder()
	mux.ServeHTTP(getRec, httptest.NewRequest(http.MethodGet, "/line/v2/bot/channel/webhook/endpoint", nil))
	if getRec.Code != http.StatusOK || !strings.Contains(getRec.Body.String(), `"endpoint":"https://example.com/webhook"`) {
		t.Fatalf("get webhook = status %d body %s", getRec.Code, getRec.Body.String())
	}

	contentRec := httptest.NewRecorder()
	mux.ServeHTTP(contentRec, httptest.NewRequest(http.MethodGet, "/line/v2/bot/message/mock-message/content", nil))
	if contentRec.Code != http.StatusOK || contentRec.Header().Get("Content-Type") == "" || contentRec.Body.Len() == 0 {
		t.Fatalf("content response = status %d content-type %q len %d", contentRec.Code, contentRec.Header().Get("Content-Type"), contentRec.Body.Len())
	}

	transcodingRec := httptest.NewRecorder()
	mux.ServeHTTP(transcodingRec, httptest.NewRequest(http.MethodGet, "/line/v2/bot/message/mock-message/content/transcoding", nil))
	if transcodingRec.Code != http.StatusOK || !strings.Contains(transcodingRec.Body.String(), `"status":"succeeded"`) {
		t.Fatalf("transcoding = status %d body %s", transcodingRec.Code, transcodingRec.Body.String())
	}
}

func TestWebhookDeliverySignsLINERequest(t *testing.T) {
	const secret = "mockport_line_secret"
	received := make(chan []byte, 1)
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read target body: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		mac := hmac.New(sha256.New, []byte(secret))
		_, _ = mac.Write(body)
		wantSignature := base64.StdEncoding.EncodeToString(mac.Sum(nil))
		if r.Header.Get("x-line-signature") != wantSignature {
			t.Errorf("signature = %q, want %q", r.Header.Get("x-line-signature"), wantSignature)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("content-type = %q, want application/json", r.Header.Get("Content-Type"))
		}
		received <- body
		w.WriteHeader(http.StatusOK)
	}))
	defer target.Close()

	mux := newLineMux(t, adapter.Config{
		BasePath:             "/line",
		Scenario:             "line_success",
		WebhookTargetURL:     target.URL,
		WebhookSigningSecret: secret,
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/line/test/webhook/send", strings.NewReader(`{"events":[{"type":"message","message":{"type":"text","id":"m1","text":"hello"}}]}`))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "127.0.0.1:12345"
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusAccepted, rec.Body.String())
	}
	select {
	case body := <-received:
		if !strings.Contains(string(body), `"destination":"U00000000000000000000000000000000"`) || !strings.Contains(string(body), `"webhookEventId"`) {
			t.Fatalf("webhook body missing LINE fields: %s", string(body))
		}
	default:
		t.Fatal("webhook target did not receive request")
	}
}

func TestWebhookDeliveryRejectsNonLocalTriggerAndUnsafeTarget(t *testing.T) {
	mux := newLineMux(t, adapter.Config{
		BasePath:             "/line",
		Scenario:             "line_success",
		WebhookTargetURL:     "http://127.0.0.1:3000/webhooks/line",
		WebhookSigningSecret: "mockport_line_secret",
	})
	remoteRec := httptest.NewRecorder()
	remoteReq := httptest.NewRequest(http.MethodPost, "/line/test/webhook/send", strings.NewReader(`{}`))
	remoteReq.Header.Set("Content-Type", "application/json")
	remoteReq.RemoteAddr = "192.168.1.10:12345"
	mux.ServeHTTP(remoteRec, remoteReq)
	if remoteRec.Code != http.StatusForbidden || !strings.Contains(remoteRec.Body.String(), "loopback") {
		t.Fatalf("remote trigger = status %d body %s", remoteRec.Code, remoteRec.Body.String())
	}

	unsafeMux := newLineMux(t, adapter.Config{
		BasePath:             "/line",
		Scenario:             "line_success",
		WebhookTargetURL:     "http://169.254.169.254/latest/meta-data",
		WebhookSigningSecret: "mockport_line_secret",
	})
	unsafeRec := httptest.NewRecorder()
	unsafeReq := httptest.NewRequest(http.MethodPost, "/line/test/webhook/send", strings.NewReader(`{}`))
	unsafeReq.Header.Set("Content-Type", "application/json")
	unsafeReq.RemoteAddr = "127.0.0.1:12345"
	unsafeMux.ServeHTTP(unsafeRec, unsafeReq)
	if unsafeRec.Code != http.StatusBadRequest || !strings.Contains(unsafeRec.Body.String(), "local Mockport target") {
		t.Fatalf("unsafe target = status %d body %s", unsafeRec.Code, unsafeRec.Body.String())
	}
}

func TestChannelAccessTokenEndpoints(t *testing.T) {
	mux := newLineMux(t, adapter.Config{BasePath: "/line", Scenario: "line_success"})

	form := url.Values{"grant_type": {"client_credentials"}, "client_id": {"mockport_line_channel"}, "client_secret": {"mockport_line_secret"}}
	tokenRec := httptest.NewRecorder()
	tokenReq := httptest.NewRequest(http.MethodPost, "/line/oauth2/v3/token", strings.NewReader(form.Encode()))
	tokenReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	mux.ServeHTTP(tokenRec, tokenReq)
	if tokenRec.Code != http.StatusOK {
		t.Fatalf("stateless token status = %d, want %d: %s", tokenRec.Code, http.StatusOK, tokenRec.Body.String())
	}
	var token map[string]any
	if err := json.Unmarshal(tokenRec.Body.Bytes(), &token); err != nil {
		t.Fatalf("decode token: %v", err)
	}
	accessToken, _ := token["access_token"].(string)
	if accessToken == "" || token["expires_in"].(float64) != 900 {
		t.Fatalf("token = %#v", token)
	}

	verifyRec := httptest.NewRecorder()
	mux.ServeHTTP(verifyRec, httptest.NewRequest(http.MethodGet, "/line/oauth2/v2.1/verify?access_token="+url.QueryEscape(accessToken), nil))
	if verifyRec.Code != http.StatusOK || !strings.Contains(verifyRec.Body.String(), `"client_id":"mockport_line_channel"`) {
		t.Fatalf("verify = status %d body %s", verifyRec.Code, verifyRec.Body.String())
	}

	revokeRec := httptest.NewRecorder()
	revokeReq := httptest.NewRequest(http.MethodPost, "/line/v2/oauth/revoke", strings.NewReader(url.Values{"access_token": {accessToken}}.Encode()))
	revokeReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	mux.ServeHTTP(revokeRec, revokeReq)
	if revokeRec.Code != http.StatusOK || strings.TrimSpace(revokeRec.Body.String()) != "" {
		t.Fatalf("revoke = status %d body %q", revokeRec.Code, revokeRec.Body.String())
	}
}

func TestRichMenuAndChatResources(t *testing.T) {
	mux := newLineMux(t, adapter.Config{BasePath: "/line", Scenario: "line_success"})

	createRec := httptest.NewRecorder()
	createReq := httptest.NewRequest(http.MethodPost, "/line/v2/bot/richmenu", strings.NewReader(`{"size":{"width":2500,"height":1686},"selected":false,"name":"Nice rich menu","chatBarText":"Tap","areas":[{"bounds":{"x":0,"y":0,"width":2500,"height":1686},"action":{"type":"postback","data":"a=b"}}]}`))
	createReq.Header.Set("Content-Type", "application/json")
	mux.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusOK {
		t.Fatalf("create rich menu status = %d, want %d: %s", createRec.Code, http.StatusOK, createRec.Body.String())
	}
	var created map[string]any
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode rich menu: %v", err)
	}
	richMenuID, _ := created["richMenuId"].(string)
	if richMenuID == "" {
		t.Fatalf("richMenuId missing: %#v", created)
	}

	uploadRec := httptest.NewRecorder()
	uploadReq := httptest.NewRequest(http.MethodPost, "/line/v2/bot/richmenu/"+richMenuID+"/content", strings.NewReader("image"))
	uploadReq.Header.Set("Content-Type", "image/png")
	mux.ServeHTTP(uploadRec, uploadReq)
	if uploadRec.Code != http.StatusOK {
		t.Fatalf("upload rich menu status = %d, want %d: %s", uploadRec.Code, http.StatusOK, uploadRec.Body.String())
	}

	defaultRec := httptest.NewRecorder()
	mux.ServeHTTP(defaultRec, httptest.NewRequest(http.MethodPost, "/line/v2/bot/user/all/richmenu/"+richMenuID, nil))
	if defaultRec.Code != http.StatusOK {
		t.Fatalf("set default rich menu status = %d, want %d: %s", defaultRec.Code, http.StatusOK, defaultRec.Body.String())
	}

	aliasRec := httptest.NewRecorder()
	aliasReq := httptest.NewRequest(http.MethodPost, "/line/v2/bot/richmenu/alias", strings.NewReader(`{"richMenuAliasId":"richmenu-alias-a","richMenuId":"`+richMenuID+`"}`))
	aliasReq.Header.Set("Content-Type", "application/json")
	mux.ServeHTTP(aliasRec, aliasReq)
	if aliasRec.Code != http.StatusOK {
		t.Fatalf("create rich menu alias status = %d, want %d: %s", aliasRec.Code, http.StatusOK, aliasRec.Body.String())
	}

	groupRec := httptest.NewRecorder()
	mux.ServeHTTP(groupRec, httptest.NewRequest(http.MethodGet, "/line/v2/bot/group/Cmockport/summary", nil))
	if groupRec.Code != http.StatusOK || !strings.Contains(groupRec.Body.String(), `"groupId":"Cmockport"`) {
		t.Fatalf("group summary = status %d body %s", groupRec.Code, groupRec.Body.String())
	}
}

func TestLoginTokenAndProfile(t *testing.T) {
	mux := newLineMux(t, adapter.Config{BasePath: "/line", Scenario: "line_success"})
	clientID := "custom_line_channel"

	authorize := httptest.NewRecorder()
	mux.ServeHTTP(authorize, httptest.NewRequest(http.MethodGet, "/line/oauth2/v2.1/authorize?client_id="+clientID+"&redirect_uri=http://localhost/callback&state=abc&scope=profile%20openid", nil))
	if authorize.Code != http.StatusFound {
		t.Fatalf("authorize status = %d, want %d", authorize.Code, http.StatusFound)
	}
	location, err := url.Parse(authorize.Header().Get("Location"))
	if err != nil {
		t.Fatalf("parse redirect: %v", err)
	}
	code := location.Query().Get("code")
	if code == "" || location.Query().Get("state") != "abc" {
		t.Fatalf("redirect location = %s", authorize.Header().Get("Location"))
	}

	form := url.Values{"grant_type": {"authorization_code"}, "code": {code}, "client_id": {clientID}, "redirect_uri": {"http://localhost/callback"}}
	tokenRec := httptest.NewRecorder()
	tokenReq := httptest.NewRequest(http.MethodPost, "/line/oauth2/v2.1/token", strings.NewReader(form.Encode()))
	tokenReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	mux.ServeHTTP(tokenRec, tokenReq)
	if tokenRec.Code != http.StatusOK {
		t.Fatalf("token status = %d, want %d: %s", tokenRec.Code, http.StatusOK, tokenRec.Body.String())
	}
	var token map[string]any
	if err := json.Unmarshal(tokenRec.Body.Bytes(), &token); err != nil {
		t.Fatalf("decode token: %v", err)
	}
	accessToken, _ := token["access_token"].(string)
	if accessToken == "" || token["token_type"] != "Bearer" {
		t.Fatalf("token = %#v", token)
	}

	profileRec := httptest.NewRecorder()
	profileReq := httptest.NewRequest(http.MethodGet, "/line/v2/profile", nil)
	profileReq.Header.Set("Authorization", "Bearer "+accessToken)
	mux.ServeHTTP(profileRec, profileReq)
	if profileRec.Code != http.StatusOK {
		t.Fatalf("login profile status = %d, want %d", profileRec.Code, http.StatusOK)
	}

	verifyRec := httptest.NewRecorder()
	mux.ServeHTTP(verifyRec, httptest.NewRequest(http.MethodGet, "/line/oauth2/v2.1/verify?access_token="+url.QueryEscape(accessToken), nil))
	if verifyRec.Code != http.StatusOK || !strings.Contains(verifyRec.Body.String(), `"client_id":"`+clientID+`"`) {
		t.Fatalf("verify token = status %d body %s", verifyRec.Code, verifyRec.Body.String())
	}
}

func TestLoginTokenRequiresAuthorizationCode(t *testing.T) {
	mux := newLineMux(t, adapter.Config{BasePath: "/line", Scenario: "line_success"})

	form := url.Values{"grant_type": {"authorization_code"}, "redirect_uri": {"http://localhost/callback"}}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/line/oauth2/v2.1/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "invalid_grant") {
		t.Fatalf("token without code = status %d body %s", rec.Code, rec.Body.String())
	}
}

func TestLoginAuthorizeRequiresClientID(t *testing.T) {
	mux := newLineMux(t, adapter.Config{BasePath: "/line", Scenario: "line_success"})

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/line/oauth2/v2.1/authorize?redirect_uri=http://localhost/callback&state=abc", nil))
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "invalid_request") {
		t.Fatalf("authorize without client_id = status %d body %s", rec.Code, rec.Body.String())
	}
}

func TestLoginTokenRejectsClientIDMismatch(t *testing.T) {
	mux := newLineMux(t, adapter.Config{BasePath: "/line", Scenario: "line_success"})

	authorize := httptest.NewRecorder()
	mux.ServeHTTP(authorize, httptest.NewRequest(http.MethodGet, "/line/oauth2/v2.1/authorize?client_id=mockport_line_channel&redirect_uri=http://localhost/callback&state=abc", nil))
	code := redirectCodeFromLocation(t, authorize.Header().Get("Location"))

	mismatch := url.Values{"grant_type": {"authorization_code"}, "code": {code}, "client_id": {"other_channel"}, "redirect_uri": {"http://localhost/callback"}}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/line/oauth2/v2.1/token", strings.NewReader(mismatch.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized || !strings.Contains(rec.Body.String(), "invalid_client") {
		t.Fatalf("mismatch token = status %d body %s", rec.Code, rec.Body.String())
	}

	valid := url.Values{"grant_type": {"authorization_code"}, "code": {code}, "client_id": {"mockport_line_channel"}, "redirect_uri": {"http://localhost/callback"}}
	validRec := httptest.NewRecorder()
	validReq := httptest.NewRequest(http.MethodPost, "/line/oauth2/v2.1/token", strings.NewReader(valid.Encode()))
	validReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	mux.ServeHTTP(validRec, validReq)
	if validRec.Code != http.StatusOK {
		t.Fatalf("valid retry status = %d, want %d: %s", validRec.Code, http.StatusOK, validRec.Body.String())
	}
}

func TestLoginTokenRequiresClientID(t *testing.T) {
	mux := newLineMux(t, adapter.Config{BasePath: "/line", Scenario: "line_success"})

	authorize := httptest.NewRecorder()
	mux.ServeHTTP(authorize, httptest.NewRequest(http.MethodGet, "/line/oauth2/v2.1/authorize?client_id=mockport_line_channel&redirect_uri=http://localhost/callback&state=abc", nil))
	code := redirectCodeFromLocation(t, authorize.Header().Get("Location"))

	missing := url.Values{"grant_type": {"authorization_code"}, "code": {code}, "redirect_uri": {"http://localhost/callback"}}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/line/oauth2/v2.1/token", strings.NewReader(missing.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized || !strings.Contains(rec.Body.String(), "invalid_client") {
		t.Fatalf("missing client_id token = status %d body %s", rec.Code, rec.Body.String())
	}

	valid := url.Values{"grant_type": {"authorization_code"}, "code": {code}, "client_id": {"mockport_line_channel"}, "redirect_uri": {"http://localhost/callback"}}
	validRec := httptest.NewRecorder()
	validReq := httptest.NewRequest(http.MethodPost, "/line/oauth2/v2.1/token", strings.NewReader(valid.Encode()))
	validReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	mux.ServeHTTP(validRec, validReq)
	if validRec.Code != http.StatusOK {
		t.Fatalf("valid retry status = %d, want %d: %s", validRec.Code, http.StatusOK, validRec.Body.String())
	}
}

func TestLoginTokenConsumesAuthorizationCode(t *testing.T) {
	mux := newLineMux(t, adapter.Config{BasePath: "/line", Scenario: "line_success"})

	authorize := httptest.NewRecorder()
	mux.ServeHTTP(authorize, httptest.NewRequest(http.MethodGet, "/line/oauth2/v2.1/authorize?client_id=mockport_line_channel&redirect_uri=http://localhost/callback&state=abc", nil))
	code := redirectCodeFromLocation(t, authorize.Header().Get("Location"))

	form := url.Values{"grant_type": {"authorization_code"}, "code": {code}, "client_id": {"mockport_line_channel"}, "redirect_uri": {"http://localhost/callback"}}
	first := httptest.NewRecorder()
	firstReq := httptest.NewRequest(http.MethodPost, "/line/oauth2/v2.1/token", strings.NewReader(form.Encode()))
	firstReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	mux.ServeHTTP(first, firstReq)
	if first.Code != http.StatusOK {
		t.Fatalf("first token status = %d, want %d: %s", first.Code, http.StatusOK, first.Body.String())
	}

	second := httptest.NewRecorder()
	secondReq := httptest.NewRequest(http.MethodPost, "/line/oauth2/v2.1/token", strings.NewReader(form.Encode()))
	secondReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	mux.ServeHTTP(second, secondReq)
	if second.Code != http.StatusBadRequest || !strings.Contains(second.Body.String(), "invalid_grant") {
		t.Fatalf("second token = status %d body %s", second.Code, second.Body.String())
	}
}

func TestLoginTokenConsumesAuthorizationCodeConcurrently(t *testing.T) {
	mux := newLineMux(t, adapter.Config{BasePath: "/line", Scenario: "line_success"})

	authorize := httptest.NewRecorder()
	mux.ServeHTTP(authorize, httptest.NewRequest(http.MethodGet, "/line/oauth2/v2.1/authorize?client_id=mockport_line_channel&redirect_uri=http://localhost/callback&state=abc", nil))
	code := redirectCodeFromLocation(t, authorize.Header().Get("Location"))

	form := url.Values{"grant_type": {"authorization_code"}, "code": {code}, "client_id": {"mockport_line_channel"}, "redirect_uri": {"http://localhost/callback"}}
	statuses := exchangeLineTokenConcurrently(mux, form.Encode(), 50)
	okCount := 0
	for _, status := range statuses {
		if status == http.StatusOK {
			okCount++
		}
	}
	if okCount != 1 {
		t.Fatalf("successful token exchanges = %d, want 1; statuses=%v", okCount, statuses)
	}
}

func exchangeLineTokenConcurrently(mux http.Handler, body string, attempts int) []int {
	var wg sync.WaitGroup
	start := make(chan struct{})
	statuses := make([]int, attempts)
	for i := range attempts {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/line/oauth2/v2.1/token", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			mux.ServeHTTP(rec, req)
			statuses[i] = rec.Code
		}(i)
	}
	close(start)
	wg.Wait()
	return statuses
}

func TestLoginAuthorizeRejectsExternalRedirectURI(t *testing.T) {
	mux := newLineMux(t, adapter.Config{BasePath: "/line", Scenario: "line_success"})

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/line/oauth2/v2.1/authorize?client_id=mockport_line_channel&redirect_uri=https://example.com/callback&state=abc", nil))
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "redirect_uri") {
		t.Fatalf("authorize external redirect = status %d body %s", rec.Code, rec.Body.String())
	}
}

func TestLinePayRequestAndConfirm(t *testing.T) {
	mux := newLineMux(t, adapter.Config{BasePath: "/line", Scenario: "line_success"})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/line/v3/payments/request", strings.NewReader(`{"amount":1200,"currency":"JPY","orderId":"order-1","packages":[{"id":"pkg-1","amount":1200,"products":[{"name":"ticket","quantity":1,"price":1200}]}]}`))
	req.Header.Set("Content-Type", "application/json")
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("pay request status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode pay request: %v", err)
	}
	if body["returnCode"] != "0000" {
		t.Fatalf("pay request = %#v", body)
	}
	info, _ := body["info"].(map[string]any)
	transactionID, _ := info["transactionId"].(string)
	if transactionID == "" {
		t.Fatalf("transaction id missing: %#v", body)
	}

	confirm := httptest.NewRecorder()
	confirmReq := httptest.NewRequest(http.MethodPost, "/line/v3/payments/"+transactionID+"/confirm", strings.NewReader(`{"amount":1200,"currency":"JPY"}`))
	confirmReq.Header.Set("Content-Type", "application/json")
	mux.ServeHTTP(confirm, confirmReq)
	if confirm.Code != http.StatusOK {
		t.Fatalf("confirm status = %d, want %d: %s", confirm.Code, http.StatusOK, confirm.Body.String())
	}
}

func TestMiniAppServiceMessageAndMiniDappPayment(t *testing.T) {
	mux := newLineMux(t, adapter.Config{BasePath: "/line", Scenario: "line_success"})

	tokenRec := httptest.NewRecorder()
	tokenReq := httptest.NewRequest(http.MethodPost, "/line/message/v3/notifier/token", strings.NewReader(`{"liffAccessToken":"mockport_liff_token"}`))
	tokenReq.Header.Set("Content-Type", "application/json")
	mux.ServeHTTP(tokenRec, tokenReq)
	if tokenRec.Code != http.StatusOK {
		t.Fatalf("notifier token status = %d, want %d: %s", tokenRec.Code, http.StatusOK, tokenRec.Body.String())
	}
	var notifier map[string]any
	if err := json.Unmarshal(tokenRec.Body.Bytes(), &notifier); err != nil {
		t.Fatalf("decode notifier token: %v", err)
	}
	notificationToken, _ := notifier["notificationToken"].(string)
	if notificationToken == "" || notifier["remainingCount"].(float64) == 0 {
		t.Fatalf("notifier = %#v", notifier)
	}

	sendRec := httptest.NewRecorder()
	sendReq := httptest.NewRequest(http.MethodPost, "/line/message/v3/notifier/send?target=service", strings.NewReader(`{"templateName":"reservation_complete_en","notificationToken":"`+notificationToken+`","params":{"username":"Brown"}}`))
	sendReq.Header.Set("Content-Type", "application/json")
	mux.ServeHTTP(sendRec, sendReq)
	if sendRec.Code != http.StatusOK {
		t.Fatalf("service message status = %d, want %d: %s", sendRec.Code, http.StatusOK, sendRec.Body.String())
	}

	paymentRec := httptest.NewRecorder()
	paymentReq := httptest.NewRequest(http.MethodPost, "/line/mini-dapp/v1/payments", strings.NewReader(`{"itemId":"item-1","amount":"10","currency":"KAIA"}`))
	paymentReq.Header.Set("Content-Type", "application/json")
	mux.ServeHTTP(paymentRec, paymentReq)
	if paymentRec.Code != http.StatusOK {
		t.Fatalf("mini dapp payment status = %d, want %d: %s", paymentRec.Code, http.StatusOK, paymentRec.Body.String())
	}
}

func TestLineErrorScenarios(t *testing.T) {
	auth := performLineRequest(t, adapter.Config{BasePath: "/line", Scenario: "auth_error"}, http.MethodPost, "/line/v2/bot/message/push", `{"to":"Umockport","messages":[{"type":"text","text":"hello"}]}`)
	if auth.Code != http.StatusUnauthorized {
		t.Fatalf("auth status = %d, want %d", auth.Code, http.StatusUnauthorized)
	}
	rate := performLineRequest(t, adapter.Config{BasePath: "/line", Scenario: "rate_limited"}, http.MethodPost, "/line/v2/bot/message/push", `{"to":"Umockport","messages":[{"type":"text","text":"hello"}]}`)
	if rate.Code != http.StatusTooManyRequests {
		t.Fatalf("rate status = %d, want %d", rate.Code, http.StatusTooManyRequests)
	}
	pay := performLineRequest(t, adapter.Config{BasePath: "/line", Scenario: "pay_failed"}, http.MethodPost, "/line/v3/payments/request", `{"amount":1200,"currency":"JPY","orderId":"order-1"}`)
	if pay.Code != http.StatusOK || !strings.Contains(pay.Body.String(), `"returnCode":"1169"`) {
		t.Fatalf("pay failure = status %d body %s", pay.Code, pay.Body.String())
	}
}

func performLineRequest(t *testing.T, cfg adapter.Config, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	mux := newLineMux(t, cfg)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	mux.ServeHTTP(rec, req)
	return rec
}

// TestConcurrentSingletonStateNoRace hammers the endpoints that mutate the
// adapter's singleton webhook endpoint and default rich menu state from many
// goroutines at once. It is meant to be run with -race: without locking around
// those fields, the detector flags the concurrent read/write.
func TestConcurrentSingletonStateNoRace(t *testing.T) {
	mux := newLineMux(t, adapter.Config{BasePath: "/line", Scenario: "line_success"})

	// Seed a rich menu with an image so it can be set as the default.
	createRec := httptest.NewRecorder()
	createReq := httptest.NewRequest(http.MethodPost, "/line/v2/bot/richmenu", strings.NewReader(`{"name":"race"}`))
	createReq.Header.Set("Content-Type", "application/json")
	mux.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusOK {
		t.Fatalf("create rich menu status = %d: %s", createRec.Code, createRec.Body.String())
	}
	var created map[string]any
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode rich menu: %v", err)
	}
	richMenuID, _ := created["richMenuId"].(string)
	if richMenuID == "" {
		t.Fatalf("missing richMenuId: %#v", created)
	}
	uploadRec := httptest.NewRecorder()
	mux.ServeHTTP(uploadRec, httptest.NewRequest(http.MethodPost, "/line/v2/bot/richmenu/"+richMenuID+"/content", nil))
	if uploadRec.Code != http.StatusOK {
		t.Fatalf("upload image status = %d: %s", uploadRec.Code, uploadRec.Body.String())
	}

	requests := []struct {
		method string
		path   string
		body   string
	}{
		{http.MethodPut, "/line/v2/bot/channel/webhook/endpoint", `{"endpoint":"https://example.test/webhook"}`},
		{http.MethodGet, "/line/v2/bot/channel/webhook/endpoint", ""},
		{http.MethodPost, "/line/v2/bot/user/all/richmenu/" + richMenuID, ""},
		{http.MethodGet, "/line/v2/bot/user/all/richmenu", ""},
		{http.MethodDelete, "/line/v2/bot/user/all/richmenu", ""},
		{http.MethodDelete, "/line/v2/bot/richmenu/" + richMenuID, ""},
	}

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		for _, rc := range requests {
			wg.Add(1)
			go func(method, path, body string) {
				defer wg.Done()
				var reader io.Reader
				if body != "" {
					reader = strings.NewReader(body)
				}
				req := httptest.NewRequest(method, path, reader)
				if body != "" {
					req.Header.Set("Content-Type", "application/json")
				}
				mux.ServeHTTP(httptest.NewRecorder(), req)
			}(rc.method, rc.path, rc.body)
		}
	}
	wg.Wait()
}

// TestConcurrentDefaultRichMenuStaysConsistent reproduces the check-then-set
// race between adopting a default rich menu and deleting that same menu. Each
// round races a set-default and a delete on one id, then verifies the invariant
// while state is quiescent: the default rich menu must never reference a menu
// that no longer exists in the store. Without serializing the set's existence
// check with the assignment (and re-checking on read), a set whose check ran
// before the delete but whose assignment lands after it leaves the default
// dangling, which this test catches.
func TestConcurrentDefaultRichMenuStaysConsistent(t *testing.T) {
	mux := newLineMux(t, adapter.Config{BasePath: "/line", Scenario: "line_success"})

	const settersPerRound = 24
	for round := 0; round < 300; round++ {
		id := createRichMenuWithImage(t, mux)

		// Race many set-default calls against a delete on the same menu id. The
		// set handler first checks the menu exists and then adopts it as the
		// default; with that check-then-set unsynchronized, a setter whose check
		// passed before the delete but whose assignment runs after it leaves the
		// default dangling. Many concurrent setters widen the chance of straddling
		// the delete.
		start := make(chan struct{})
		var wg sync.WaitGroup
		wg.Add(settersPerRound + 1)
		for s := 0; s < settersPerRound; s++ {
			go func() {
				defer wg.Done()
				<-start
				mux.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/line/v2/bot/user/all/richmenu/"+id, nil))
			}()
		}
		go func() {
			defer wg.Done()
			<-start
			mux.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodDelete, "/line/v2/bot/richmenu/"+id, nil))
		}()
		close(start) // release all goroutines as simultaneously as possible
		wg.Wait()

		// Quiescent invariant check: nothing mutates state now, so the default
		// reported here is authoritative and must point at a menu that exists.
		getRec := httptest.NewRecorder()
		mux.ServeHTTP(getRec, httptest.NewRequest(http.MethodGet, "/line/v2/bot/user/all/richmenu", nil))
		if getRec.Code == http.StatusNotFound {
			continue // no default set — consistent
		}
		if getRec.Code != http.StatusOK {
			t.Fatalf("round %d: get default status = %d: %s", round, getRec.Code, getRec.Body.String())
		}
		var got map[string]any
		if err := json.Unmarshal(getRec.Body.Bytes(), &got); err != nil {
			t.Fatalf("round %d: decode default: %v", round, err)
		}
		defaultID, _ := got["richMenuId"].(string)
		lookupRec := httptest.NewRecorder()
		mux.ServeHTTP(lookupRec, httptest.NewRequest(http.MethodGet, "/line/v2/bot/richmenu/"+defaultID, nil))
		if lookupRec.Code != http.StatusOK {
			t.Fatalf("round %d: default rich menu %q does not exist in store (lookup status %d): the default points at a deleted menu", round, defaultID, lookupRec.Code)
		}
	}
}

func createRichMenuWithImage(t *testing.T, mux *http.ServeMux) string {
	t.Helper()
	createRec := httptest.NewRecorder()
	createReq := httptest.NewRequest(http.MethodPost, "/line/v2/bot/richmenu", strings.NewReader(`{"name":"race"}`))
	createReq.Header.Set("Content-Type", "application/json")
	mux.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusOK {
		t.Fatalf("create rich menu status = %d: %s", createRec.Code, createRec.Body.String())
	}
	var created map[string]any
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode rich menu: %v", err)
	}
	id, _ := created["richMenuId"].(string)
	if id == "" {
		t.Fatalf("missing richMenuId: %#v", created)
	}
	uploadRec := httptest.NewRecorder()
	mux.ServeHTTP(uploadRec, httptest.NewRequest(http.MethodPost, "/line/v2/bot/richmenu/"+id+"/content", nil))
	if uploadRec.Code != http.StatusOK {
		t.Fatalf("upload image status = %d: %s", uploadRec.Code, uploadRec.Body.String())
	}
	return id
}

func newLineMux(t *testing.T, cfg adapter.Config) *http.ServeMux {
	t.Helper()
	mux := http.NewServeMux()
	if err := New().Register(mux, cfg); err != nil {
		t.Fatalf("register line adapter: %v", err)
	}
	return mux
}

func redirectCodeFromLocation(t *testing.T, location string) string {
	t.Helper()
	parsed, err := url.Parse(location)
	if err != nil {
		t.Fatalf("parse redirect location: %v", err)
	}
	code := parsed.Query().Get("code")
	if code == "" {
		t.Fatalf("missing code in location: %q", location)
	}
	return code
}
