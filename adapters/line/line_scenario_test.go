package line_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	lineadapter "github.com/albert-einshutoin/mockport/adapters/line"
	"github.com/albert-einshutoin/mockport/internal/adapter"
)

func newLineMuxForScenario(t *testing.T, cfg adapter.Config) *http.ServeMux {
	t.Helper()
	mux := http.NewServeMux()
	if err := lineadapter.New().Register(mux, cfg); err != nil {
		t.Fatalf("Register: %v", err)
	}
	return mux
}

func TestLINEHeaderOverridesConfig(t *testing.T) {
	// config=line_success でサーバーを起動し、ヘッダで auth_error を指定
	mux := newLineMuxForScenario(t, adapter.Config{BasePath: "/line", Scenario: "line_success"})
	body := strings.NewReader(`{"to":"Umockport","messages":[{"type":"text","text":"hello"}]}`)
	req := httptest.NewRequest(http.MethodPost, "/line/v2/bot/message/push", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Mockport-Scenario", "auth_error")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("want 401, got %d", rec.Code)
	}
}

func TestLINEUnknownScenarioReturns400(t *testing.T) {
	// Messaging API のエラー本文は {"message": ...} 形式で専用コードフィールドが無いため、
	// 共通コードは message 先頭の固定プレフィックスとして判別する（厳密にプレフィックス一致）。
	mux := newLineMuxForScenario(t, adapter.Config{BasePath: "/line", Scenario: "line_success"})
	body := strings.NewReader(`{"to":"Umockport","messages":[{"type":"text","text":"hello"}]}`)
	req := httptest.NewRequest(http.MethodPost, "/line/v2/bot/message/push", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Mockport-Scenario", "not_a_real_scenario")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", rec.Code)
	}
	var resp map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	msg, _ := resp["message"].(string)
	if !strings.HasPrefix(msg, "unknown_mockport_scenario:") {
		t.Errorf("want message prefixed with unknown_mockport_scenario:, got %q", msg)
	}
}

// TestLINEOAuthUnknownScenarioUsesErrorCode は OAuth エンドポイント（{"error": ...}
// 形式）で共通コードが機械可読な error フィールドへ入ることを厳密一致で固定する。
func TestLINEOAuthUnknownScenarioUsesErrorCode(t *testing.T) {
	mux := newLineMuxForScenario(t, adapter.Config{BasePath: "/line", Scenario: "line_success"})
	req := httptest.NewRequest(http.MethodPost, "/line/oauth2/v2.1/token", strings.NewReader("grant_type=authorization_code&code=x"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Mockport-Scenario", "not_a_real_scenario")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
	var resp map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if code, _ := resp["error"].(string); code != "unknown_mockport_scenario" {
		t.Errorf("want error=unknown_mockport_scenario, got %q", code)
	}
}

// TestLINEBotInfoUnknownScenarioReturns400 は GET /line/v2/bot/info（read エンドポイント）で
// 未知の X-Mockport-Scenario が 400 になることを固定する（指摘3）。
// 修正前はこのエンドポイントが resolver を呼ばず 200 で成功していた。
func TestLINEBotInfoUnknownScenarioReturns400(t *testing.T) {
	mux := newLineMuxForScenario(t, adapter.Config{BasePath: "/line", Scenario: "line_success"})
	req := httptest.NewRequest(http.MethodGet, "/line/v2/bot/info", nil)
	req.Header.Set("X-Mockport-Scenario", "totally_unknown")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d (body: %s)", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	msg, _ := resp["message"].(string)
	if !strings.HasPrefix(msg, "unknown_mockport_scenario:") {
		t.Errorf("want message prefixed with unknown_mockport_scenario:, got %q", msg)
	}
}

// TestLINEPayUnknownScenarioReturnsCommonCode は LINE Pay エンドポイント
// （{"returnCode": ..., "returnMessage": ...} 形式）で returnCode の数値契約を壊さず、
// 共通コードが returnMessage 先頭プレフィックスで判別できることを固定する。
//
// LINE Pay は実 API 仕様に合わせて業務エラーを HTTP 200 + returnCode で表すため、
// 未知シナリオも意図的に 400 ではなく HTTP 200 を返す（provider-shaped な例外）。
// その 200 契約を明示的に固定する。
func TestLINEPayUnknownScenarioReturnsCommonCode(t *testing.T) {
	mux := newLineMuxForScenario(t, adapter.Config{BasePath: "/line", Scenario: "line_success"})
	req := httptest.NewRequest(http.MethodPost, "/line/v3/payments/request", strings.NewReader(`{"amount":1000}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Mockport-Scenario", "not_a_real_scenario")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	// LINE Pay の provider-shaped 例外: 未知シナリオでも HTTP 200 を返す。
	if rec.Code != http.StatusOK {
		t.Errorf("want 200 (LINE Pay returns business errors as HTTP 200), got %d", rec.Code)
	}
	var resp map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if code, _ := resp["returnCode"].(string); code != "2101" {
		t.Errorf("want returnCode=2101 (numeric contract preserved), got %q", code)
	}
	msg, _ := resp["returnMessage"].(string)
	if !strings.HasPrefix(msg, "unknown_mockport_scenario:") {
		t.Errorf("want returnMessage prefixed with unknown_mockport_scenario:, got %q", msg)
	}
}

// TestLINEMessagingFollowersUnknownScenarioReturns400 は dispatch 層一元検証により
// これまで resolver を通していなかった Messaging API の read エンドポイント
// GET /v2/bot/followers/ids も未知シナリオで 400（{"message": ...} 形式）を返すことを固定する。
func TestLINEMessagingFollowersUnknownScenarioReturns400(t *testing.T) {
	mux := newLineMuxForScenario(t, adapter.Config{BasePath: "/line", Scenario: "line_success"})
	req := httptest.NewRequest(http.MethodGet, "/line/v2/bot/followers/ids", nil)
	req.Header.Set("X-Mockport-Scenario", "not_a_real_scenario")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d (body: %s)", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	msg, _ := resp["message"].(string)
	if !strings.HasPrefix(msg, "unknown_mockport_scenario:") {
		t.Errorf("want message prefixed with unknown_mockport_scenario:, got %q", msg)
	}
}

// TestLINEOAuthVerifyUnknownScenarioReturns400 は dispatch 層一元検証により
// GET /oauth2/v2.1/verify（OAuth/Login 系 read エンドポイント）も未知シナリオで
// 400 + 機械可読な error=unknown_mockport_scenario を返すことを固定する。
func TestLINEOAuthVerifyUnknownScenarioReturns400(t *testing.T) {
	mux := newLineMuxForScenario(t, adapter.Config{BasePath: "/line", Scenario: "line_success"})
	req := httptest.NewRequest(http.MethodGet, "/line/oauth2/v2.1/verify?access_token=mockport", nil)
	req.Header.Set("X-Mockport-Scenario", "not_a_real_scenario")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d (body: %s)", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if code, _ := resp["error"].(string); code != "unknown_mockport_scenario" {
		t.Errorf("want error=unknown_mockport_scenario, got %q", code)
	}
}

// TestLINEPayCheckUnknownScenarioReturnsCommonCode は dispatch 層一元検証により
// GET /v3/payments/requests/{id}/check（LINE Pay 系 read エンドポイント）も未知シナリオで
// 共通コードを returnMessage プレフィックスで返すことを固定する。LINE Pay は業務エラーを
// HTTP 200 + returnCode で表す契約のため、ステータスは 200 のまま returnCode の数値契約を保つ。
func TestLINEPayCheckUnknownScenarioReturnsCommonCode(t *testing.T) {
	mux := newLineMuxForScenario(t, adapter.Config{BasePath: "/line", Scenario: "line_success"})
	req := httptest.NewRequest(http.MethodGet, "/line/v3/payments/requests/tx_mockport/check", nil)
	req.Header.Set("X-Mockport-Scenario", "not_a_real_scenario")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	// LINE Pay の provider-shaped 例外: 未知シナリオでも HTTP 200 を返す。
	if rec.Code != http.StatusOK {
		t.Errorf("want 200 (LINE Pay returns business errors as HTTP 200), got %d", rec.Code)
	}
	var resp map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if code, _ := resp["returnCode"].(string); code != "2101" {
		t.Errorf("want returnCode=2101 (numeric contract preserved), got %q", code)
	}
	msg, _ := resp["returnMessage"].(string)
	if !strings.HasPrefix(msg, "unknown_mockport_scenario:") {
		t.Errorf("want returnMessage prefixed with unknown_mockport_scenario:, got %q", msg)
	}
}
