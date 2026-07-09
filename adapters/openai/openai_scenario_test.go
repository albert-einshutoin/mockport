package openai_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	openai "github.com/albert-einshutoin/mockport/adapters/openai"
	"github.com/albert-einshutoin/mockport/internal/adapter"
)

func newOpenAIMuxScenario(t *testing.T, cfg adapter.Config) *http.ServeMux {
	t.Helper()
	mux := http.NewServeMux()
	if err := openai.New().Register(mux, cfg); err != nil {
		t.Fatalf("Register: %v", err)
	}
	return mux
}

func TestOpenAIHeaderOverridesConfig(t *testing.T) {
	mux := newOpenAIMuxScenario(t, adapter.Config{BasePath: "/openai", Scenario: "chat_success"})
	body := strings.NewReader(`{"model":"gpt-4o","messages":[{"role":"user","content":"hi"}]}`)
	req := httptest.NewRequest(http.MethodPost, "/openai/v1/chat/completions", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Mockport-Scenario", "auth_error")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("want 401, got %d", rec.Code)
	}
}

func TestOpenAIResetExemptFromScenarioValidation(t *testing.T) {
	// /test/reset は状態リセット専用の管理エンドポイントなので、未知シナリオヘッダを
	// 付けても 400 で弾かず通常どおりリセットが成功する（ループバックからの実行）。
	mux := newOpenAIMuxScenario(t, adapter.Config{BasePath: "/openai", Scenario: "chat_success"})
	req := httptest.NewRequest(http.MethodPost, "/openai/test/reset", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	req.Header.Set("X-Mockport-Scenario", "no_such_scenario")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200 (reset exempt from scenario validation), got %d, body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["reset"] != true {
		t.Errorf("want reset=true, got %#v", resp)
	}
}

func TestOpenAIUnknownScenarioReturns400(t *testing.T) {
	mux := newOpenAIMuxScenario(t, adapter.Config{BasePath: "/openai", Scenario: "chat_success"})
	body := strings.NewReader(`{"model":"gpt-4o","messages":[{"role":"user","content":"hi"}]}`)
	req := httptest.NewRequest(http.MethodPost, "/openai/v1/chat/completions", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Mockport-Scenario", "no_such_scenario")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", rec.Code)
	}
	var resp map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	errObj, _ := resp["error"].(map[string]any)
	if errObj == nil {
		t.Fatal("want error object")
	}
	code, _ := errObj["code"].(string)
	if code != "unknown_mockport_scenario" {
		t.Errorf("want code=unknown_mockport_scenario, got %q", code)
	}
}

// TestOpenAIModelsUnknownScenarioReturns400 は GET /openai/v1/models（read エンドポイント）で
// 未知の X-Mockport-Scenario が 400 になることを固定する（指摘3）。
// 修正前はこのエンドポイントが dispatcher で resolver を呼ばず 200 で成功していた。
func TestOpenAIModelsUnknownScenarioReturns400(t *testing.T) {
	mux := newOpenAIMuxScenario(t, adapter.Config{BasePath: "/openai", Scenario: "chat_success"})
	req := httptest.NewRequest(http.MethodGet, "/openai/v1/models", nil)
	req.Header.Set("X-Mockport-Scenario", "totally_unknown")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", rec.Code)
	}
	var resp map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	errObj, _ := resp["error"].(map[string]any)
	if errObj == nil {
		t.Fatal("want error object")
	}
	code, _ := errObj["code"].(string)
	if code != "unknown_mockport_scenario" {
		t.Errorf("want code=unknown_mockport_scenario, got %q", code)
	}
}
