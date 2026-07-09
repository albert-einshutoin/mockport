package zohooauth_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	zohoadapter "github.com/albert-einshutoin/mockport/adapters/zohooauth"
	"github.com/albert-einshutoin/mockport/internal/adapter"
)

func newZohoMuxForScenario(t *testing.T, cfg adapter.Config) *http.ServeMux {
	t.Helper()
	mux := http.NewServeMux()
	if err := zohoadapter.New().Register(mux, cfg); err != nil {
		t.Fatalf("Register: %v", err)
	}
	return mux
}

func TestZohoOAuthHeaderOverridesConfig(t *testing.T) {
	mux := newZohoMuxForScenario(t, adapter.Config{BasePath: "/zoho", Scenario: "oauth_success"})
	form := url.Values{
		"grant_type":   {"authorization_code"},
		"code":         {"somecode"},
		"client_id":    {"mockport_zoho_client"},
		"redirect_uri": {"http://localhost/callback"},
	}.Encode()
	req := httptest.NewRequest(http.MethodPost, "/zoho/oauth/v2/token", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Mockport-Scenario", "invalid_code")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	// Zoho は 200 でエラーを返す
	if rec.Code != http.StatusOK {
		t.Errorf("want 200, got %d", rec.Code)
	}
	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	errStr, _ := body["error"].(string)
	if errStr != "invalid_code" {
		t.Errorf("want error=invalid_code, got %q", errStr)
	}
}

func TestZohoOAuthUnknownScenarioReturns400(t *testing.T) {
	mux := newZohoMuxForScenario(t, adapter.Config{BasePath: "/zoho", Scenario: "oauth_success"})
	form := url.Values{
		"grant_type": {"authorization_code"},
		"code":       {"somecode"},
	}.Encode()
	req := httptest.NewRequest(http.MethodPost, "/zoho/oauth/v2/token", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Mockport-Scenario", "totally_fake")
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

// TestZohoAuthorizeUnknownScenarioReturns400 は GET /zoho/oauth/v2/auth（authorize エンドポイント）で
// 未知の X-Mockport-Scenario が 400 になることを固定する（指摘3）。
// 修正前はこのエンドポイントが dispatcher で resolver を呼ばず 302 でリダイレクトしていた。
func TestZohoAuthorizeUnknownScenarioReturns400(t *testing.T) {
	mux := newZohoMuxForScenario(t, adapter.Config{BasePath: "/zoho", Scenario: "oauth_success"})
	req := httptest.NewRequest(http.MethodGet, "/zoho/oauth/v2/auth?client_id=testclient&redirect_uri=http://localhost/callback", nil)
	req.Header.Set("X-Mockport-Scenario", "totally_fake")
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

// TestZohoUserInfoUnknownScenarioReturns400 は userInfo 経路の未知シナリオが
// 400 + 機械可読な error=unknown_mockport_scenario を返すことを固定する（S3）。
// userInfo は token 経路（Zoho の業務エラーは 200）と異なり、Mockport レベルの
// 設定エラーとして真の 400 を返すのが妥当であり、その挙動をここで固定する。
func TestZohoUserInfoUnknownScenarioReturns400(t *testing.T) {
	mux := newZohoMuxForScenario(t, adapter.Config{BasePath: "/zoho", Scenario: "oauth_success"})
	req := httptest.NewRequest(http.MethodGet, "/zoho/oauth/user/info", nil)
	req.Header.Set("Authorization", "Zoho-oauthtoken sometoken")
	req.Header.Set("X-Mockport-Scenario", "totally_fake")
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
}
