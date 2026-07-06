package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/albert-einshutoin/mockport/adapters/stripe"
	"github.com/albert-einshutoin/mockport/internal/adapter"
	"github.com/albert-einshutoin/mockport/internal/compat"
	"github.com/albert-einshutoin/mockport/internal/config"
	"github.com/albert-einshutoin/mockport/internal/report"
)

func TestReportEndpointReturnsRequestsAndSafety(t *testing.T) {
	cfg := config.Config{
		Mode:   "ai-safe",
		Server: config.ServerConfig{Host: "127.0.0.1", Port: 43101},
		Adapters: map[string]config.AdapterConfig{
			"stripe": {Enabled: true, BasePath: "/stripe", Scenario: "payment_success", FakeSecret: "sk_live_123"},
		},
	}
	if err := config.Validate(&cfg); err != nil {
		t.Fatalf("validate config: %v", err)
	}
	reg := adapter.NewRegistry()
	if err := reg.Register(stripe.New()); err != nil {
		t.Fatalf("register stripe: %v", err)
	}
	recorder := report.NewRecorder()
	handler, err := NewConfiguredHandler(cfg, reg, recorder)
	if err != nil {
		t.Fatalf("new handler: %v", err)
	}

	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/stripe/v1/checkout/sessions", nil))
	reportRec := httptest.NewRecorder()
	handler.ServeHTTP(reportRec, httptest.NewRequest(http.MethodGet, "/_mockport/report", nil))

	if reportRec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", reportRec.Code, http.StatusOK)
	}
	var snapshot report.Snapshot
	if err := json.Unmarshal(reportRec.Body.Bytes(), &snapshot); err != nil {
		t.Fatalf("decode report: %v", err)
	}
	if snapshot.Mode != "ai-safe" {
		t.Fatalf("mode = %q", snapshot.Mode)
	}
	if len(snapshot.Adapters) != 1 || snapshot.Adapters[0].Name != "stripe" {
		t.Fatalf("adapters = %#v", snapshot.Adapters)
	}
	if len(snapshot.Requests) != 1 || snapshot.Requests[0].Path != "/stripe/v1/checkout/sessions" {
		t.Fatalf("requests = %#v", snapshot.Requests)
	}
	if len(snapshot.SafetyWarnings) != 1 {
		t.Fatalf("safety warnings = %#v", snapshot.SafetyWarnings)
	}
	if snapshot.Safety.Safe {
		t.Fatal("safety safe = true, want false")
	}
	if snapshot.Safety.Mode != "ai-safe" {
		t.Fatalf("safety mode = %q, want ai-safe", snapshot.Safety.Mode)
	}
	if snapshot.Safety.RealLookingSecrets != 1 {
		t.Fatalf("real-looking secrets = %d, want 1", snapshot.Safety.RealLookingSecrets)
	}
	if snapshot.Safety.PublicEnvSafe {
		t.Fatal("public env safe = true, want false")
	}
	if len(snapshot.ScenarioCoverage) != 1 || snapshot.ScenarioCoverage[0].Adapter != "stripe" {
		t.Fatalf("scenario coverage = %#v", snapshot.ScenarioCoverage)
	}
	if len(snapshot.BehaviorMatrix) == 0 {
		t.Fatal("behavior matrix is empty")
	}
	if snapshot.Adapters[0].Maturity != "workflow-compatible" {
		t.Fatalf("maturity = %q, want workflow-compatible", snapshot.Adapters[0].Maturity)
	}
	if len(snapshot.Compatibility) != 1 {
		t.Fatalf("compatibility = %#v, want one entry", snapshot.Compatibility)
	}
	if snapshot.Compatibility[0].Adapter != "stripe" || snapshot.Compatibility[0].Score == 0 {
		t.Fatalf("compatibility entry = %#v", snapshot.Compatibility[0])
	}
	if snapshot.Compatibility[0].ProviderVersion == "" {
		t.Fatalf("provider version is empty: %#v", snapshot.Compatibility[0])
	}
	if snapshot.Compatibility[0].SDKCoverage != 100 || snapshot.Compatibility[0].StateCoverage != 100 || snapshot.Compatibility[0].ErrorCoverage != 100 {
		t.Fatalf("compatibility coverage = %#v", snapshot.Compatibility[0])
	}
	if !snapshot.Compatibility[0].PromotionEligible {
		t.Fatalf("stripe should be promotion-eligible: %#v", snapshot.Compatibility[0])
	}
	if len(snapshot.StateCoverage) != 1 || snapshot.StateCoverage[0].Adapter != "stripe" || !snapshot.StateCoverage[0].Idempotency {
		t.Fatalf("state coverage = %#v", snapshot.StateCoverage)
	}
}

func TestReportEndpointRecordsUnsupportedEndpoint(t *testing.T) {
	cfg := config.Config{
		Mode:   "ai-safe",
		Server: config.ServerConfig{Host: "127.0.0.1", Port: 43101},
		Adapters: map[string]config.AdapterConfig{
			"stripe": {Enabled: true, BasePath: "/stripe", Scenario: "payment_success", FakeSecret: "mockport_stripe_secret"},
		},
	}
	if err := config.Validate(&cfg); err != nil {
		t.Fatalf("validate config: %v", err)
	}
	reg := adapter.NewRegistry()
	if err := reg.Register(stripe.New()); err != nil {
		t.Fatalf("register stripe: %v", err)
	}
	handler, err := NewConfiguredHandler(cfg, reg, report.NewRecorder())
	if err != nil {
		t.Fatalf("new handler: %v", err)
	}

	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/stripe/v1/not-supported", nil))
	reportRec := httptest.NewRecorder()
	handler.ServeHTTP(reportRec, httptest.NewRequest(http.MethodGet, "/_mockport/report", nil))

	var snapshot report.Snapshot
	if err := json.Unmarshal(reportRec.Body.Bytes(), &snapshot); err != nil {
		t.Fatalf("decode report: %v", err)
	}
	if len(snapshot.UnsupportedEndpoints) != 1 {
		t.Fatalf("unsupported endpoint count = %d, want 1", len(snapshot.UnsupportedEndpoints))
	}
	if snapshot.UnsupportedEndpoints[0].Reason != "unsupported_endpoint" {
		t.Fatalf("reason = %q", snapshot.UnsupportedEndpoints[0].Reason)
	}
	if snapshot.Requests[0].Adapter != "stripe" || snapshot.Requests[0].Scenario != "payment_success" {
		t.Fatalf("request metadata = %#v", snapshot.Requests[0])
	}
}

// newStripeHandler は X-Mockport-Scenario の記録を検証するための stripe 単体ハンドラを作る。
func newStripeHandler(t *testing.T, recorder *report.Recorder) http.Handler {
	t.Helper()
	cfg := config.Config{
		Mode:   "ai-safe",
		Server: config.ServerConfig{Host: "127.0.0.1", Port: 43101},
		Adapters: map[string]config.AdapterConfig{
			"stripe": {Enabled: true, BasePath: "/stripe", Scenario: "payment_success", FakeSecret: "mockport_stripe_secret"},
		},
	}
	if err := config.Validate(&cfg); err != nil {
		t.Fatalf("validate config: %v", err)
	}
	reg := adapter.NewRegistry()
	if err := reg.Register(stripe.New()); err != nil {
		t.Fatalf("register stripe: %v", err)
	}
	handler, err := NewConfiguredHandler(cfg, reg, recorder)
	if err != nil {
		t.Fatalf("new handler: %v", err)
	}
	return handler
}

// TestReportRejectsUnknownScenarioHeaderInRecord は未知シナリオ→400 のリクエストで、
// 不正なヘッダ値（totally_unknown）がレポートに混入せず、設定済みの有効なシナリオ
// （payment_success）のまま記録されることを固定する（B1）。
func TestReportRejectsUnknownScenarioHeaderInRecord(t *testing.T) {
	recorder := report.NewRecorder()
	handler := newStripeHandler(t, recorder)

	req := httptest.NewRequest(http.MethodPost, "/stripe/v1/checkout/sessions", nil)
	req.Header.Set(adapter.ScenarioHeader, "totally_unknown")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d (unknown scenario should be rejected)", rec.Code, http.StatusBadRequest)
	}

	snapshot := recorder.Snapshot()
	if len(snapshot.Requests) != 1 {
		t.Fatalf("requests = %#v, want one", snapshot.Requests)
	}
	got := snapshot.Requests[0]
	if got.Scenario == "totally_unknown" {
		t.Fatalf("scenario = %q, must not record the unvalidated header value", got.Scenario)
	}
	if got.Scenario != "payment_success" {
		t.Fatalf("scenario = %q, want payment_success (configured default)", got.Scenario)
	}
}

// TestReportRecordsValidScenarioHeaderOverride は有効なヘッダ override が
// レポートへ反映されることを固定する（B1）。
func TestReportRecordsValidScenarioHeaderOverride(t *testing.T) {
	recorder := report.NewRecorder()
	handler := newStripeHandler(t, recorder)

	req := httptest.NewRequest(http.MethodPost, "/stripe/v1/checkout/sessions", nil)
	req.Header.Set(adapter.ScenarioHeader, "payment_failed")
	handler.ServeHTTP(httptest.NewRecorder(), req)

	snapshot := recorder.Snapshot()
	if len(snapshot.Requests) != 1 {
		t.Fatalf("requests = %#v, want one", snapshot.Requests)
	}
	if got := snapshot.Requests[0]; got.Scenario != "payment_failed" {
		t.Fatalf("scenario = %q, want payment_failed (valid header override)", got.Scenario)
	}
}

func TestReportEndpointRecordsUnsupportedMethodAs405(t *testing.T) {
	cfg := config.Config{
		Mode:   "ai-safe",
		Server: config.ServerConfig{Host: "127.0.0.1", Port: 43101},
		Adapters: map[string]config.AdapterConfig{
			"stripe": {Enabled: true, BasePath: "/stripe", Scenario: "payment_success", FakeSecret: "mockport_stripe_secret"},
		},
	}
	if err := config.Validate(&cfg); err != nil {
		t.Fatalf("validate config: %v", err)
	}
	reg := adapter.NewRegistry()
	if err := reg.Register(stripe.New()); err != nil {
		t.Fatalf("register stripe: %v", err)
	}
	handler, err := NewConfiguredHandler(cfg, reg, report.NewRecorder())
	if err != nil {
		t.Fatalf("new handler: %v", err)
	}

	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodDelete, "/stripe/v1/customers", nil))
	reportRec := httptest.NewRecorder()
	handler.ServeHTTP(reportRec, httptest.NewRequest(http.MethodGet, "/_mockport/report", nil))

	var snapshot report.Snapshot
	if err := json.Unmarshal(reportRec.Body.Bytes(), &snapshot); err != nil {
		t.Fatalf("decode report: %v", err)
	}
	if len(snapshot.UnsupportedEndpoints) != 1 {
		t.Fatalf("unsupported endpoint count = %d, want 1", len(snapshot.UnsupportedEndpoints))
	}
	if snapshot.UnsupportedEndpoints[0].Status != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", snapshot.UnsupportedEndpoints[0].Status, http.StatusMethodNotAllowed)
	}
	if snapshot.UnsupportedEndpoints[0].Reason != "unsupported_endpoint" {
		t.Fatalf("reason = %q", snapshot.UnsupportedEndpoints[0].Reason)
	}
}

func TestStateCoverageFromAdapterMetadata(t *testing.T) {
	got, ok := stateCoverage(adapter.Metadata{
		Name:              "stripe",
		StatefulResources: []string{"checkout_session", "payment_intent"},
		Idempotency:       true,
		Reset:             true,
	})
	if !ok {
		t.Fatal("state coverage ok = false, want true")
	}
	if got.Adapter != "stripe" || !got.Idempotency || !got.Reset {
		t.Fatalf("state coverage = %#v", got)
	}
	if len(got.StatefulResources) != 2 {
		t.Fatalf("resources = %#v", got.StatefulResources)
	}
}

func TestStateCoverageSkipsStatelessAdapterMetadata(t *testing.T) {
	if got, ok := stateCoverage(adapter.Metadata{Name: "openai"}); ok {
		t.Fatalf("state coverage = %#v, want skipped", got)
	}
}

// TestCompatibilityStatusReportsPromotionEligibility pins that the PromotionEligible
// value emitted in the report is computed from internal/compat.CanPromote as the single
// source of truth. In particular, a declaration missing a hierarchy level (e.g.
// LevelWorkflow) must be emitted as not promotion-eligible even when coverage is full.
func TestCompatibilityStatusReportsPromotionEligibility(t *testing.T) {
	eligible := compat.Manifest{
		Adapter:         "demo",
		Maturity:        adapter.MaturityWorkflowCompatible,
		ProviderVersion: "1",
		Levels:          []compat.Level{compat.LevelWire, compat.LevelWorkflow, compat.LevelState, compat.LevelError},
		Endpoints:       []compat.Endpoint{{ID: "one", Supported: true}},
		Scenarios: []compat.Scenario{
			{Name: "ok", BuiltIn: true, Supported: true},
			{Name: "failed", BuiltIn: true, Supported: true},
		},
		StateEvidence: &compat.StateEvidence{StatefulResources: []string{"resource"}, Reset: true},
	}
	if status := compatibilityStatus(eligible); !status.PromotionEligible {
		t.Fatalf("workflow-compatible manifest with full evidence should be promotion-eligible: %#v", status)
	}

	// Declares provider-compatible but lacks LevelWorkflow. Coverage can be full, yet it
	// must not be eligible because it fails CanPromote's hierarchy condition.
	ineligible := eligible
	ineligible.Maturity = adapter.MaturityProviderCompatible
	ineligible.Levels = []compat.Level{compat.LevelWire, compat.LevelSDK, compat.LevelState, compat.LevelError, compat.LevelContract}
	ineligible.SDKVersions = []compat.SDKVersion{{Name: "lib", Version: "1.0.0"}}
	ineligible.ContractEvidence = &compat.ContractEvidence{
		Fixtures:     []string{"compat/fixtures/demo/success.json"},
		SDKContracts: []string{"contract/sdk/demo"},
		KnownGaps:    []string{"docs/compatibility-reports/latest.json#demo"},
	}
	if status := compatibilityStatus(ineligible); status.PromotionEligible {
		t.Fatalf("provider-compatible without workflow level must not be promotion-eligible: %#v", status)
	}
	status := compatibilityStatus(ineligible)
	if status.ContractEvidence == nil || len(status.ContractEvidence.Fixtures) != 1 {
		t.Fatalf("contract evidence was not reported: %#v", status)
	}
}
