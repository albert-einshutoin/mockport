package githuboauth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"github.com/albert-einshutoin/mockport/internal/adapter"
)

func TestAuthorizeRedirect(t *testing.T) {
	rec := performRequest(t, adapter.Config{BasePath: "/github", Scenario: "oauth_success"}, http.MethodGet, "/github/login/oauth/authorize?client_id=mockport_github_client&redirect_uri=http://localhost/callback&state=s1")
	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusFound)
	}
	if rec.Header().Get("Location") == "" {
		t.Fatal("missing Location header")
	}
}

func TestAuthorizeRequiresClientID(t *testing.T) {
	rec := performRequest(t, adapter.Config{BasePath: "/github", Scenario: "oauth_success"}, http.MethodGet, "/github/login/oauth/authorize?redirect_uri=http://localhost/callback&state=s1")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
	assertGitHubOAuthError(t, rec, "invalid_request")
}

func TestAuthorizeRejectsExternalRedirectURI(t *testing.T) {
	rec := performRequest(t, adapter.Config{BasePath: "/github", Scenario: "oauth_success"}, http.MethodGet, "/github/login/oauth/authorize?client_id=mockport_github_client&redirect_uri=https://example.com/callback&state=s1")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
	assertGitHubOAuthError(t, rec, "redirect_uri_mismatch")
}

func TestAccessTokenScenarios(t *testing.T) {
	tests := []struct {
		name     string
		scenario string
		status   int
	}{
		{"invalid", "invalid_code", http.StatusBadRequest},
		{"expired", "expired_token", http.StatusUnauthorized},
		{"scope", "scope_missing", http.StatusForbidden},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := performRequest(t, adapter.Config{BasePath: "/github", Scenario: tt.scenario}, http.MethodPost, "/github/login/oauth/access_token")
			if rec.Code != tt.status {
				t.Fatalf("status = %d, want %d", rec.Code, tt.status)
			}
		})
	}
}

func TestAccessTokenRequiresAuthorizationCode(t *testing.T) {
	mux := newGitHubMux(t, adapter.Config{BasePath: "/github", Scenario: "oauth_success"})
	rec := serveGitHubRequest(mux, http.MethodPost, "/github/login/oauth/access_token", "", map[string]string{"Accept": "application/json"})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
	assertGitHubOAuthError(t, rec, "bad_verification_code")
}

func TestUser(t *testing.T) {
	mux := newGitHubMux(t, adapter.Config{BasePath: "/github", Scenario: "oauth_success"})
	token := issueGitHubToken(t, mux, "read:user")
	rec := serveGitHubRequest(mux, http.MethodGet, "/github/user", "", map[string]string{"Authorization": "Bearer " + token})
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var body map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode user: %v", err)
	}
	if body["login"] != "mockport-user" {
		t.Fatalf("login = %v", body["login"])
	}
}

func TestUserRequiresValidBearerToken(t *testing.T) {
	mux := newGitHubMux(t, adapter.Config{BasePath: "/github", Scenario: "oauth_success"})

	missing := serveGitHubRequest(mux, http.MethodGet, "/github/user", "", nil)
	if missing.Code != http.StatusUnauthorized {
		t.Fatalf("missing token status = %d, body=%s", missing.Code, missing.Body.String())
	}
	assertGitHubMessage(t, missing, "Bad credentials")

	unknown := serveGitHubRequest(mux, http.MethodGet, "/github/user", "", map[string]string{"Authorization": "Bearer unknown"})
	if unknown.Code != http.StatusUnauthorized {
		t.Fatalf("unknown token status = %d, body=%s", unknown.Code, unknown.Body.String())
	}
	assertGitHubMessage(t, unknown, "Bad credentials")
}

func TestOAuthCodeTokenAndUserAreStateful(t *testing.T) {
	mux := newGitHubMux(t, adapter.Config{BasePath: "/github", Scenario: "oauth_success"})

	auth := serveGitHubRequest(mux, http.MethodGet, "/github/login/oauth/authorize?client_id=mockport_github_client&redirect_uri=http://localhost/callback&state=s1&scope=read:user,user:email", "", nil)
	if auth.Code != http.StatusFound {
		t.Fatalf("authorize status = %d, want %d", auth.Code, http.StatusFound)
	}
	location := auth.Header().Get("Location")
	parsed, err := url.Parse(location)
	if err != nil {
		t.Fatalf("parse redirect location: %v", err)
	}
	code := parsed.Query().Get("code")
	if code != "github_oauth_oauth_code_000001" {
		t.Fatalf("code = %q", code)
	}

	token := serveGitHubRequest(mux, http.MethodPost, "/github/login/oauth/access_token", "code="+code+"&redirect_uri=http://localhost/callback&client_id=mockport_github_client", map[string]string{"Accept": "application/json"})
	if token.Code != http.StatusOK {
		t.Fatalf("token status = %d, want %d, body=%s", token.Code, http.StatusOK, token.Body.String())
	}
	var tokenBody map[string]any
	if err := json.Unmarshal(token.Body.Bytes(), &tokenBody); err != nil {
		t.Fatalf("decode token: %v", err)
	}
	if tokenBody["access_token"] != "github_oauth_oauth_token_000001" || tokenBody["scope"] != "read:user,user:email" {
		t.Fatalf("token body = %#v", tokenBody)
	}

	user := serveGitHubRequest(mux, http.MethodGet, "/github/user", "", map[string]string{"Authorization": "Bearer github_oauth_oauth_token_000001"})
	if user.Code != http.StatusOK {
		t.Fatalf("user status = %d, want %d, body=%s", user.Code, http.StatusOK, user.Body.String())
	}
	var userBody map[string]any
	if err := json.Unmarshal(user.Body.Bytes(), &userBody); err != nil {
		t.Fatalf("decode user: %v", err)
	}
	if userBody["login"] != "mockport-user" || userBody["scope"] != "read:user,user:email" {
		t.Fatalf("user body = %#v", userBody)
	}
}

func TestEmailsAndOrgsRequireScopes(t *testing.T) {
	mux := newGitHubMux(t, adapter.Config{BasePath: "/github", Scenario: "oauth_success"})
	token := issueGitHubToken(t, mux, "read:user user:email read:org")

	emails := serveGitHubRequest(mux, http.MethodGet, "/github/user/emails", "", map[string]string{"Authorization": "Bearer " + token})
	if emails.Code != http.StatusOK || !strings.Contains(emails.Body.String(), `"email":"mockport@example.test"`) {
		t.Fatalf("emails status/body = %d %s", emails.Code, emails.Body.String())
	}

	orgs := serveGitHubRequest(mux, http.MethodGet, "/github/user/orgs", "", map[string]string{"Authorization": "Bearer " + token})
	if orgs.Code != http.StatusOK || !strings.Contains(orgs.Body.String(), `"login":"mockport-org"`) {
		t.Fatalf("orgs status/body = %d %s", orgs.Code, orgs.Body.String())
	}

	readOnly := issueGitHubToken(t, mux, "read:user")
	for _, path := range []string{"/github/user/emails", "/github/user/orgs"} {
		rec := serveGitHubRequest(mux, http.MethodGet, path, "", map[string]string{"Authorization": "Bearer " + readOnly})
		if rec.Code != http.StatusForbidden {
			t.Fatalf("%s status = %d, body=%s", path, rec.Code, rec.Body.String())
		}
		assertGitHubMessage(t, rec, "Resource not accessible by token")
	}
}

func TestTokenRejectsRedirectURIMismatch(t *testing.T) {
	mux := newGitHubMux(t, adapter.Config{BasePath: "/github", Scenario: "oauth_success"})
	auth := serveGitHubRequest(mux, http.MethodGet, "/github/login/oauth/authorize?client_id=mockport_github_client&redirect_uri=http://localhost/callback&state=s1", "", nil)
	code := redirectCode(t, auth)

	rec := serveGitHubRequest(mux, http.MethodPost, "/github/login/oauth/access_token", "code="+code+"&redirect_uri=http://other/callback&client_id=mockport_github_client", map[string]string{"Accept": "application/json"})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
	assertGitHubOAuthError(t, rec, "redirect_uri_mismatch")
}

func TestTokenRejectsClientIDMismatch(t *testing.T) {
	mux := newGitHubMux(t, adapter.Config{BasePath: "/github", Scenario: "oauth_success"})
	auth := serveGitHubRequest(mux, http.MethodGet, "/github/login/oauth/authorize?client_id=mockport_github_client&redirect_uri=http://localhost/callback&state=s1", "", nil)
	code := redirectCode(t, auth)

	rec := serveGitHubRequest(mux, http.MethodPost, "/github/login/oauth/access_token", "code="+code+"&redirect_uri=http://localhost/callback&client_id=other_client", map[string]string{"Accept": "application/json"})
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
	assertGitHubOAuthError(t, rec, "incorrect_client_credentials")

	valid := serveGitHubRequest(mux, http.MethodPost, "/github/login/oauth/access_token", "code="+code+"&redirect_uri=http://localhost/callback&client_id=mockport_github_client", map[string]string{"Accept": "application/json"})
	if valid.Code != http.StatusOK {
		t.Fatalf("valid retry status = %d, want %d, body=%s", valid.Code, http.StatusOK, valid.Body.String())
	}
}

func TestTokenRequiresClientID(t *testing.T) {
	mux := newGitHubMux(t, adapter.Config{BasePath: "/github", Scenario: "oauth_success"})
	auth := serveGitHubRequest(mux, http.MethodGet, "/github/login/oauth/authorize?client_id=mockport_github_client&redirect_uri=http://localhost/callback&state=s1", "", nil)
	code := redirectCode(t, auth)

	rec := serveGitHubRequest(mux, http.MethodPost, "/github/login/oauth/access_token", "code="+code+"&redirect_uri=http://localhost/callback", map[string]string{"Accept": "application/json"})
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
	assertGitHubOAuthError(t, rec, "incorrect_client_credentials")

	valid := serveGitHubRequest(mux, http.MethodPost, "/github/login/oauth/access_token", "code="+code+"&redirect_uri=http://localhost/callback&client_id=mockport_github_client", map[string]string{"Accept": "application/json"})
	if valid.Code != http.StatusOK {
		t.Fatalf("valid retry status = %d, want %d, body=%s", valid.Code, http.StatusOK, valid.Body.String())
	}
}

func TestTokenConsumesAuthorizationCode(t *testing.T) {
	mux := newGitHubMux(t, adapter.Config{BasePath: "/github", Scenario: "oauth_success"})
	auth := serveGitHubRequest(mux, http.MethodGet, "/github/login/oauth/authorize?client_id=mockport_github_client&redirect_uri=http://localhost/callback&state=s1", "", nil)
	code := redirectCode(t, auth)
	body := "code=" + code + "&redirect_uri=http://localhost/callback&client_id=mockport_github_client"

	first := serveGitHubRequest(mux, http.MethodPost, "/github/login/oauth/access_token", body, map[string]string{"Accept": "application/json"})
	if first.Code != http.StatusOK {
		t.Fatalf("first token status = %d, want %d, body=%s", first.Code, http.StatusOK, first.Body.String())
	}
	second := serveGitHubRequest(mux, http.MethodPost, "/github/login/oauth/access_token", body, map[string]string{"Accept": "application/json"})
	if second.Code != http.StatusBadRequest {
		t.Fatalf("second token status = %d, want %d, body=%s", second.Code, http.StatusBadRequest, second.Body.String())
	}
	assertGitHubOAuthError(t, second, "bad_verification_code")
}

func TestTokenConsumesAuthorizationCodeConcurrently(t *testing.T) {
	mux := newGitHubMux(t, adapter.Config{BasePath: "/github", Scenario: "oauth_success"})
	auth := serveGitHubRequest(mux, http.MethodGet, "/github/login/oauth/authorize?client_id=mockport_github_client&redirect_uri=http://localhost/callback&state=s1", "", nil)
	code := redirectCode(t, auth)
	body := "code=" + code + "&redirect_uri=http://localhost/callback&client_id=mockport_github_client"

	statuses := exchangeGitHubTokenConcurrently(mux, body, 50)
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

func TestAuthorizeReportsUnsupportedScope(t *testing.T) {
	mux := newGitHubMux(t, adapter.Config{BasePath: "/github", Scenario: "oauth_success"})
	rec := serveGitHubRequest(mux, http.MethodGet, "/github/login/oauth/authorize?client_id=mockport_github_client&redirect_uri=http://localhost/callback&state=s1&scope=repo", "", nil)
	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusFound)
	}
	location := rec.Header().Get("Location")
	if !strings.Contains(location, "error=unsupported_scope") || !strings.Contains(location, "state=s1") {
		t.Fatalf("location = %q", location)
	}
}

func TestOAuthRejectsUnknownCode(t *testing.T) {
	mux := newGitHubMux(t, adapter.Config{BasePath: "/github", Scenario: "oauth_success"})
	rec := serveGitHubRequest(mux, http.MethodPost, "/github/login/oauth/access_token", "code=unknown", nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestMetadata(t *testing.T) {
	meta := New().Metadata()
	if meta.Name != "github-oauth" || meta.Maturity != "workflow-compatible" {
		t.Fatalf("metadata = %#v", meta)
	}
	if meta.ProviderVersion != "2022-11-28" || len(meta.Levels) < 5 || len(meta.Endpoints) < 5 {
		t.Fatalf("compat metadata = %#v", meta)
	}
	if !meta.Reset || len(meta.StatefulResources) != 3 {
		t.Fatalf("state metadata = %#v", meta)
	}
}

func performRequest(t *testing.T, cfg adapter.Config, method, path string) *httptest.ResponseRecorder {
	t.Helper()
	mux := newGitHubMux(t, cfg)
	return serveGitHubRequest(mux, method, path, "", nil)
}

func newGitHubMux(t *testing.T, cfg adapter.Config) *http.ServeMux {
	t.Helper()
	mux := http.NewServeMux()
	if err := New().Register(mux, cfg); err != nil {
		t.Fatalf("register adapter: %v", err)
	}
	return mux
}

func serveGitHubRequest(mux http.Handler, method, path, body string, headers map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	for name, value := range headers {
		req.Header.Set(name, value)
	}
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

func exchangeGitHubTokenConcurrently(mux http.Handler, body string, attempts int) []int {
	var wg sync.WaitGroup
	start := make(chan struct{})
	statuses := make([]int, attempts)
	for i := range attempts {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			rec := serveGitHubRequest(mux, http.MethodPost, "/github/login/oauth/access_token", body, map[string]string{"Accept": "application/json"})
			statuses[i] = rec.Code
		}(i)
	}
	close(start)
	wg.Wait()
	return statuses
}

func issueGitHubToken(t *testing.T, mux http.Handler, scope string) string {
	t.Helper()
	auth := serveGitHubRequest(mux, http.MethodGet, "/github/login/oauth/authorize?client_id=mockport_github_client&redirect_uri=http://localhost/callback&state=s1&scope="+url.QueryEscape(scope), "", nil)
	code := redirectCode(t, auth)
	token := serveGitHubRequest(mux, http.MethodPost, "/github/login/oauth/access_token", "code="+code+"&redirect_uri=http://localhost/callback&client_id=mockport_github_client", map[string]string{"Accept": "application/json"})
	if token.Code != http.StatusOK {
		t.Fatalf("issue token status = %d, body=%s", token.Code, token.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(token.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode token: %v", err)
	}
	accessToken, _ := body["access_token"].(string)
	if accessToken == "" {
		t.Fatalf("missing access token: %#v", body)
	}
	return accessToken
}

func redirectCode(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()
	location := rec.Header().Get("Location")
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

func assertGitHubOAuthError(t *testing.T, rec *httptest.ResponseRecorder, want string) {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if body["error"] != want || body["error_description"] == "" || body["error_uri"] == "" {
		t.Fatalf("error body = %#v", body)
	}
}

func assertGitHubMessage(t *testing.T, rec *httptest.ResponseRecorder, want string) {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode message: %v", err)
	}
	if body["message"] != want || body["documentation_url"] == "" || body["status"] == "" {
		t.Fatalf("message body = %#v", body)
	}
}
