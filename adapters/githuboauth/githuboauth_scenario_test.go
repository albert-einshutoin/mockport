package githuboauth_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	githuboauthadapter "github.com/albert-einshutoin/mockport/adapters/githuboauth"
	"github.com/albert-einshutoin/mockport/internal/adapter"
)

func newGitHubMuxForScenario(t *testing.T, cfg adapter.Config) *http.ServeMux {
	t.Helper()
	mux := http.NewServeMux()
	if err := githuboauthadapter.New().Register(mux, cfg); err != nil {
		t.Fatalf("Register: %v", err)
	}
	return mux
}

func TestGitHubOAuthHeaderOverridesConfig(t *testing.T) {
	mux := newGitHubMuxForScenario(t, adapter.Config{BasePath: "/github", Scenario: "oauth_success"})
	form := url.Values{"code": {"somecode"}, "client_id": {"mockport_github_client"}}.Encode()
	req := httptest.NewRequest(http.MethodPost, "/github/login/oauth/access_token", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Mockport-Scenario", "invalid_code")
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
	if errStr != "bad_verification_code" {
		t.Errorf("want error=bad_verification_code, got %q", errStr)
	}
}

func TestGitHubOAuthUnknownScenarioReturns400(t *testing.T) {
	mux := newGitHubMuxForScenario(t, adapter.Config{BasePath: "/github", Scenario: "oauth_success"})
	form := url.Values{"code": {"somecode"}, "client_id": {"mockport_github_client"}}.Encode()
	req := httptest.NewRequest(http.MethodPost, "/github/login/oauth/access_token", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Mockport-Scenario", "definitely_fake")
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

// TestGitHubAuthorizeUnknownScenarioReturns400 は GET /github/login/oauth/authorize
// （authorize エンドポイント）で未知の X-Mockport-Scenario が 400 になることを固定する（指摘3）。
// 修正前はこのエンドポイントが dispatcher で resolver を呼ばず 302 でリダイレクトしていた。
func TestGitHubAuthorizeUnknownScenarioReturns400(t *testing.T) {
	mux := newGitHubMuxForScenario(t, adapter.Config{BasePath: "/github", Scenario: "oauth_success"})
	req := httptest.NewRequest(http.MethodGet, "/github/login/oauth/authorize?client_id=testclient&redirect_uri=http://localhost/callback", nil)
	req.Header.Set("X-Mockport-Scenario", "totally_unknown")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d (body: %s)", rec.Code, rec.Body.String())
	}
	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if errStr, _ := body["error"].(string); errStr != "unknown_mockport_scenario" {
		t.Errorf("want error=unknown_mockport_scenario, got %q", errStr)
	}
}

// TestGitHubUserUnknownScenarioUsesErrorCode は user 情報エンドポイント（GitHub REST
// 形式 {"message": ..., "documentation_url": ...}）の未知シナリオ応答で、共通コードが
// メッセージ連結ではなく機械可読な error フィールドへ厳密一致で入ることを固定する。
func TestGitHubUserUnknownScenarioUsesErrorCode(t *testing.T) {
	mux := newGitHubMuxForScenario(t, adapter.Config{BasePath: "/github", Scenario: "oauth_success"})
	req := httptest.NewRequest(http.MethodGet, "/github/user", nil)
	req.Header.Set("X-Mockport-Scenario", "definitely_fake")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if errStr, _ := body["error"].(string); errStr != "unknown_mockport_scenario" {
		t.Errorf("want error=unknown_mockport_scenario, got %q", errStr)
	}
	// message にコードを連結していないこと（人間向け説明にとどめること）を確認する。
	if msg, _ := body["message"].(string); strings.HasPrefix(msg, "unknown_mockport_scenario") {
		t.Errorf("message must not carry the code, got %q", msg)
	}
}
