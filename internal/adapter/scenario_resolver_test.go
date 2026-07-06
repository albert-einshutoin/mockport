package adapter_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/albert-einshutoin/mockport/internal/adapter"
)

func TestScenarioResolver(t *testing.T) {
	knownScenarios := []adapter.Scenario{
		{Name: "payment_success", Supported: true},
		{Name: "payment_failed", Supported: true},
		{Name: "auth_error", Supported: true},
	}
	meta := adapter.Metadata{
		Name:            "stripe",
		Maturity:        adapter.MaturityWorkflowCompatible,
		ProviderVersion: "2025-10-29.clover",
		Scenarios:       knownScenarios,
	}

	tests := []struct {
		name           string
		configScenario string
		defaultName    string
		headerValue    string
		want           string
		wantErr        bool
	}{
		{
			name:           "ヘッダなし・config値あり→config値を返す",
			configScenario: "payment_failed",
			defaultName:    "payment_success",
			headerValue:    "",
			want:           "payment_failed",
		},
		{
			name:           "ヘッダなし・config値が未知→ErrUnknownScenario",
			configScenario: "payment_faled", // タイプミス等の未知値
			defaultName:    "payment_success",
			headerValue:    "",
			wantErr:        true,
		},
		{
			name:           "ヘッダなし・config空→default値を返す",
			configScenario: "",
			defaultName:    "payment_success",
			headerValue:    "",
			want:           "payment_success",
		},
		{
			name:           "ヘッダあり既知→ヘッダ値を返す",
			configScenario: "payment_success",
			defaultName:    "payment_success",
			headerValue:    "payment_failed",
			want:           "payment_failed",
		},
		{
			name:           "ヘッダあり未知→ErrUnknownScenario",
			configScenario: "payment_success",
			defaultName:    "payment_success",
			headerValue:    "nonexistent_scenario",
			wantErr:        true,
		},
		{
			name:           "ヘッダ前後空白トリム→正常解決",
			configScenario: "payment_success",
			defaultName:    "payment_success",
			headerValue:    "  payment_failed  ",
			want:           "payment_failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := adapter.Config{Scenario: tt.configScenario}
			resolver := adapter.NewScenarioResolver(cfg, tt.defaultName, meta)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.headerValue != "" {
				req.Header.Set(adapter.ScenarioHeader, tt.headerValue)
			}

			got, err := resolver.Resolve(req)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("want error, got nil (scenario=%q)", got)
				}
				if !errors.Is(err, adapter.ErrUnknownScenario) {
					t.Fatalf("want ErrUnknownScenario, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

// TestResolveStoresResolvedScenarioInContext は Resolve 成功時に検証済みシナリオが
// context へ記録され、ResolvedScenarioFromContext で取得できることを固定する。
func TestResolveStoresResolvedScenarioInContext(t *testing.T) {
	meta := adapter.Metadata{
		Name:      "stripe",
		Scenarios: []adapter.Scenario{{Name: "payment_success", Supported: true}, {Name: "payment_failed", Supported: true}},
	}
	resolver := adapter.NewScenarioResolver(adapter.Config{Scenario: "payment_success"}, "payment_success", meta)

	req := adapter.WithScenarioCapture(httptest.NewRequest(http.MethodGet, "/", nil))
	req.Header.Set(adapter.ScenarioHeader, "payment_failed")

	got, err := resolver.Resolve(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "payment_failed" {
		t.Fatalf("resolve = %q, want payment_failed", got)
	}
	stored, ok := adapter.ResolvedScenarioFromContext(req.Context())
	if !ok {
		t.Fatal("ResolvedScenarioFromContext ok = false, want true")
	}
	if stored != "payment_failed" {
		t.Fatalf("stored scenario = %q, want payment_failed", stored)
	}
}

// TestResolveDoesNotStoreUnknownScenario は未知シナリオで Resolve が失敗した場合、
// context に何も書き込まれないこと（不正値が混入しないこと）を固定する。
func TestResolveDoesNotStoreUnknownScenario(t *testing.T) {
	meta := adapter.Metadata{
		Name:      "stripe",
		Scenarios: []adapter.Scenario{{Name: "payment_success", Supported: true}},
	}
	resolver := adapter.NewScenarioResolver(adapter.Config{Scenario: "payment_success"}, "payment_success", meta)

	req := adapter.WithScenarioCapture(httptest.NewRequest(http.MethodGet, "/", nil))
	req.Header.Set(adapter.ScenarioHeader, "totally_unknown")

	if _, err := resolver.Resolve(req); !errors.Is(err, adapter.ErrUnknownScenario) {
		t.Fatalf("err = %v, want ErrUnknownScenario", err)
	}
	if stored, ok := adapter.ResolvedScenarioFromContext(req.Context()); ok {
		t.Fatalf("stored scenario = %q, must not be set on unknown scenario", stored)
	}
}
