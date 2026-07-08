package zohooauth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/albert-einshutoin/mockport/internal/adapter"
	"github.com/albert-einshutoin/mockport/internal/adapter/adaptertest"
)

func TestAuthorizeRedirectsWithCodeAndEchoesState(t *testing.T) {
	mux := newZohoMux(t, adapter.Config{BasePath: "/zoho", Scenario: "oauth_success"})
	rec := serve(mux, http.MethodGet, "/zoho/oauth/v2/auth?response_type=code&client_id=mockport_zoho_client&scope=email&redirect_uri=http://localhost/callback&state=s1&access_type=online", "", nil)
	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusFound, rec.Body.String())
	}
	loc := parseLocation(t, rec)
	if loc.Query().Get("code") == "" {
		t.Fatalf("missing code in redirect: %q", loc.String())
	}
	if got := loc.Query().Get("state"); got != "s1" {
		t.Fatalf("state echo = %q, want %q", got, "s1")
	}
}

func TestAuthorizeRequiresClientID(t *testing.T) {
	mux := newZohoMux(t, adapter.Config{BasePath: "/zoho"})
	rec := serve(mux, http.MethodGet, "/zoho/oauth/v2/auth?redirect_uri=http://localhost/callback&state=s1", "", nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestAuthorizeRejectsExternalRedirectURI(t *testing.T) {
	mux := newZohoMux(t, adapter.Config{BasePath: "/zoho"})
	rec := serve(mux, http.MethodGet, "/zoho/oauth/v2/auth?client_id=mockport_zoho_client&redirect_uri=https://evil.example.com/callback&state=s1", "", nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestTokenSuccessReturnsAccessToken(t *testing.T) {
	mux := newZohoMux(t, adapter.Config{BasePath: "/zoho"})
	code := authorizeCode(t, mux, "")
	rec := serve(mux, http.MethodPost, "/zoho/oauth/v2/token", tokenForm(code), formHeaders())
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	body := decode(t, rec)
	if body["access_token"] == nil || body["access_token"] == "" {
		t.Fatalf("missing access_token: %#v", body)
	}
	if _, hasError := body["error"]; hasError {
		t.Fatalf("unexpected error field: %#v", body)
	}
}

func TestTokenRejectsBadGrantType(t *testing.T) {
	mux := newZohoMux(t, adapter.Config{BasePath: "/zoho"})
	code := authorizeCode(t, mux, "")
	form := "grant_type=client_credentials&client_id=mockport_zoho_client&client_secret=mockport_zoho_secret&redirect_uri=http://localhost/callback&code=" + code
	rec := serve(mux, http.MethodPost, "/zoho/oauth/v2/token", form, formHeaders())
	assertTokenError(t, rec)
}

func TestTokenRejectsUnknownCode(t *testing.T) {
	mux := newZohoMux(t, adapter.Config{BasePath: "/zoho"})
	rec := serve(mux, http.MethodPost, "/zoho/oauth/v2/token", tokenForm("zoho_oauth_oauth_code_999999"), formHeaders())
	assertTokenError(t, rec)
}

func TestTokenScenarioForcesError(t *testing.T) {
	mux := newZohoMux(t, adapter.Config{BasePath: "/zoho", Scenario: "invalid_code"})
	rec := serve(mux, http.MethodPost, "/zoho/oauth/v2/token", tokenForm("anything"), formHeaders())
	assertTokenError(t, rec)
}

func TestTokenConsumesAuthorizationCode(t *testing.T) {
	mux := newZohoMux(t, adapter.Config{BasePath: "/zoho"})
	code := authorizeCode(t, mux, "")
	form := tokenForm(code)

	first := serve(mux, http.MethodPost, "/zoho/oauth/v2/token", form, formHeaders())
	if first.Code != http.StatusOK || decode(t, first)["access_token"] == nil {
		t.Fatalf("first exchange failed: %d %s", first.Code, first.Body.String())
	}
	second := serve(mux, http.MethodPost, "/zoho/oauth/v2/token", form, formHeaders())
	assertTokenError(t, second)
}

func TestTokenRejectsRedirectURIMismatch(t *testing.T) {
	mux := newZohoMux(t, adapter.Config{BasePath: "/zoho"})
	code := authorizeCode(t, mux, "")
	form := "grant_type=authorization_code&client_id=mockport_zoho_client&client_secret=mockport_zoho_secret&redirect_uri=http://localhost/other&code=" + code
	rec := serve(mux, http.MethodPost, "/zoho/oauth/v2/token", form, formHeaders())
	assertTokenError(t, rec)
}

func TestTokenRejectsClientIDMismatch(t *testing.T) {
	mux := newZohoMux(t, adapter.Config{BasePath: "/zoho"})
	code := authorizeCode(t, mux, "")
	form := "grant_type=authorization_code&client_id=other_client&client_secret=mockport_zoho_secret&redirect_uri=http://localhost/callback&code=" + code
	rec := serve(mux, http.MethodPost, "/zoho/oauth/v2/token", form, formHeaders())
	assertTokenError(t, rec)
}

func TestTokenExchangedExactlyOnceUnderConcurrency(t *testing.T) {
	mux := newZohoMux(t, adapter.Config{BasePath: "/zoho"})
	code := authorizeCode(t, mux, "")
	form := tokenForm(code)

	const attempts = 50
	results := adaptertest.ConcurrentResults(attempts, func() map[string]any {
		rec := serve(mux, http.MethodPost, "/zoho/oauth/v2/token", form, formHeaders())
		var body map[string]any
		_ = json.Unmarshal(rec.Body.Bytes(), &body)
		return body
	})

	success := 0
	for _, body := range results {
		if token, _ := body["access_token"].(string); token != "" {
			success++
		}
	}
	if success != 1 {
		t.Fatalf("successful token exchanges = %d, want 1", success)
	}
}

func TestUserInfoUsesEnvDefaults(t *testing.T) {
	t.Setenv("ZOHO_USER_EMAIL", "env@example.test")
	t.Setenv("ZOHO_USER_NAME", "Env User")
	mux := newZohoMux(t, adapter.Config{BasePath: "/zoho"})
	token := issueToken(t, mux, "")
	rec := serve(mux, http.MethodGet, "/zoho/oauth/user/info", "", map[string]string{"Authorization": authScheme + token})
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", rec.Code, rec.Body.String())
	}
	body := decode(t, rec)
	if body["Email"] != "env@example.test" || body["Display_Name"] != "Env User" {
		t.Fatalf("env defaults not applied: %#v", body)
	}
}

func TestUserInfoSuccessShape(t *testing.T) {
	mux := newZohoMux(t, adapter.Config{BasePath: "/zoho"})
	token := issueToken(t, mux, "")
	rec := serve(mux, http.MethodGet, "/zoho/oauth/user/info", "", map[string]string{"Authorization": authScheme + token})
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	body := decode(t, rec)
	if body["Email"] != defaultUserEmail {
		t.Fatalf("Email = %v, want %q", body["Email"], defaultUserEmail)
	}
	if body["Display_Name"] != defaultUserName {
		t.Fatalf("Display_Name = %v, want %q", body["Display_Name"], defaultUserName)
	}
}

func TestUserInfoOverrideEmailViaAuthorizeQuery(t *testing.T) {
	mux := newZohoMux(t, adapter.Config{BasePath: "/zoho"})
	token := issueToken(t, mux, "mock_email=alice@example.test&mock_name=Alice")
	rec := serve(mux, http.MethodGet, "/zoho/oauth/user/info", "", map[string]string{"Authorization": authScheme + token})
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", rec.Code, rec.Body.String())
	}
	body := decode(t, rec)
	if body["Email"] != "alice@example.test" || body["Display_Name"] != "Alice" {
		t.Fatalf("override not applied: %#v", body)
	}
}

func TestUserInfoRejectsNonZohoScheme(t *testing.T) {
	mux := newZohoMux(t, adapter.Config{BasePath: "/zoho"})
	token := issueToken(t, mux, "")

	// A valid token with the wrong scheme (Bearer) must fail.
	bearer := serve(mux, http.MethodGet, "/zoho/oauth/user/info", "", map[string]string{"Authorization": "Bearer " + token})
	if bearer.Code != http.StatusUnauthorized {
		t.Fatalf("bearer status = %d, want %d", bearer.Code, http.StatusUnauthorized)
	}

	missing := serve(mux, http.MethodGet, "/zoho/oauth/user/info", "", nil)
	if missing.Code != http.StatusUnauthorized {
		t.Fatalf("missing status = %d, want %d", missing.Code, http.StatusUnauthorized)
	}

	unknown := serve(mux, http.MethodGet, "/zoho/oauth/user/info", "", map[string]string{"Authorization": authScheme + "unknown"})
	if unknown.Code != http.StatusUnauthorized {
		t.Fatalf("unknown token status = %d, want %d", unknown.Code, http.StatusUnauthorized)
	}
}

func TestUserInfoScenarioForcesUnauthorized(t *testing.T) {
	mux := newZohoMux(t, adapter.Config{BasePath: "/zoho", Scenario: "invalid_token"})
	token := issueToken(t, mux, "")
	rec := serve(mux, http.MethodGet, "/zoho/oauth/user/info", "", map[string]string{"Authorization": authScheme + token})
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestMetadata(t *testing.T) {
	meta := New().Metadata()
	if meta.Name != adapterName || meta.Maturity != adapter.MaturityWorkflowCompatible {
		t.Fatalf("metadata = %#v", meta)
	}
	if !meta.Reset || len(meta.StatefulResources) != 2 || len(meta.Endpoints) != 4 {
		t.Fatalf("metadata surface = %#v", meta)
	}
}

func TestResetClearsStateAndRequiresLoopback(t *testing.T) {
	mux := newZohoMux(t, adapter.Config{BasePath: "/zoho"})
	token := issueToken(t, mux, "")

	before := serve(mux, http.MethodGet, "/zoho/oauth/user/info", "", map[string]string{"Authorization": authScheme + token})
	if before.Code != http.StatusOK {
		t.Fatalf("user before reset = %d", before.Code)
	}

	reset := serveZohoReset(mux, "127.0.0.1:12345")
	if reset.Code != http.StatusOK {
		t.Fatalf("reset status = %d, body=%s", reset.Code, reset.Body.String())
	}
	body := decode(t, reset)
	if body["reset"] != true || body["adapter"] != "zoho-oauth" {
		t.Fatalf("reset body = %#v", body)
	}
	assertResetResponseDoesNotExposeSecrets(t, reset, token)

	after := serve(mux, http.MethodGet, "/zoho/oauth/user/info", "", map[string]string{"Authorization": authScheme + token})
	if after.Code != http.StatusUnauthorized {
		t.Fatalf("user after reset = %d, want 401", after.Code)
	}

	remote := serveZohoReset(mux, "192.168.0.2:12345")
	if remote.Code != http.StatusForbidden {
		t.Fatalf("remote reset = %d, want 403", remote.Code)
	}
	if !strings.Contains(remote.Body.String(), "loopback") {
		t.Fatalf("remote reset body = %s", remote.Body.String())
	}
	assertResetResponseDoesNotExposeSecrets(t, remote, token)
}

func assertResetResponseDoesNotExposeSecrets(t *testing.T, rec *httptest.ResponseRecorder, knownSecrets ...string) {
	t.Helper()
	body := rec.Body.String()

	var parsed map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &parsed); err == nil {
		for _, key := range []string{"access_token", "client_secret", "Email", "Display_Name", "email", "name"} {
			if _, ok := parsed[key]; ok {
				t.Fatalf("reset response exposes forbidden key %q: %s", key, body)
			}
		}
	}

	forbidden := append([]string{
		"mockport_zoho_secret",
		defaultUserEmail,
		defaultUserName,
	}, knownSecrets...)
	for _, secret := range forbidden {
		if secret == "" {
			continue
		}
		if strings.Contains(body, secret) {
			t.Fatalf("reset response leaks sensitive value: %s", body)
		}
	}
}

func serveZohoReset(mux http.Handler, remoteAddr string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/zoho/test/reset", nil)
	req.RemoteAddr = remoteAddr
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

// --- helpers ---

func newZohoMux(t *testing.T, cfg adapter.Config) *http.ServeMux {
	t.Helper()
	return adaptertest.NewMux(t, New(), cfg)
}

func serve(mux http.Handler, method, path, body string, headers map[string]string) *httptest.ResponseRecorder {
	header := http.Header{}
	for name, value := range headers {
		header.Set(name, value)
	}
	return adaptertest.Serve(mux, method, path, strings.NewReader(body), header)
}

func formHeaders() map[string]string {
	return map[string]string{"Content-Type": "application/x-www-form-urlencoded"}
}

func tokenForm(code string) string {
	return "grant_type=authorization_code&client_id=mockport_zoho_client&client_secret=mockport_zoho_secret&redirect_uri=http://localhost/callback&code=" + code
}

func authorizeCode(t *testing.T, mux http.Handler, extraQuery string) string {
	t.Helper()
	path := "/zoho/oauth/v2/auth?client_id=mockport_zoho_client&redirect_uri=http://localhost/callback&state=s1"
	if extraQuery != "" {
		path += "&" + extraQuery
	}
	rec := serve(mux, http.MethodGet, path, "", nil)
	if rec.Code != http.StatusFound {
		t.Fatalf("authorize status = %d, body=%s", rec.Code, rec.Body.String())
	}
	code := parseLocation(t, rec).Query().Get("code")
	if code == "" {
		t.Fatal("missing code")
	}
	return code
}

func issueToken(t *testing.T, mux http.Handler, extraQuery string) string {
	t.Helper()
	code := authorizeCode(t, mux, extraQuery)
	rec := serve(mux, http.MethodPost, "/zoho/oauth/v2/token", tokenForm(code), formHeaders())
	if rec.Code != http.StatusOK {
		t.Fatalf("token status = %d, body=%s", rec.Code, rec.Body.String())
	}
	token, _ := decode(t, rec)["access_token"].(string)
	if token == "" {
		t.Fatal("missing access_token")
	}
	return token
}

func parseLocation(t *testing.T, rec *httptest.ResponseRecorder) *url.URL {
	t.Helper()
	parsed, err := url.Parse(rec.Header().Get("Location"))
	if err != nil {
		t.Fatalf("parse Location: %v", err)
	}
	return parsed
}

func decode(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v (%s)", err, rec.Body.String())
	}
	return body
}

func assertTokenError(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()
	// Zoho returns HTTP 200 even for token-exchange failures; the client
	// inspects the error field, not the status code.
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (Zoho returns 200 on token error)", rec.Code, http.StatusOK)
	}
	body := decode(t, rec)
	if errVal, ok := body["error"].(string); !ok || errVal == "" {
		t.Fatalf("expected error field, got %#v (status %d)", body, rec.Code)
	}
	if body["access_token"] != nil {
		t.Fatalf("unexpected access_token in error body: %#v", body)
	}
}
