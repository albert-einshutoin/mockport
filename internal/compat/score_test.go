package compat

import (
	"net/http"
	"testing"

	"github.com/albert-einshutoin/mockport/internal/adapter"
)

func TestCalculateScoreUsesEndpointScenarioSDKStateAndErrorCoverage(t *testing.T) {
	score := CalculateScore(Manifest{
		Adapter:         "stripe",
		ProviderVersion: "2026-05-26",
		Levels:          []Level{LevelWire, LevelSDK, LevelState, LevelError},
		SDKVersions:     []SDKVersion{{Name: "stripe-go", Version: "v83.0.0"}},
		Endpoints: []Endpoint{
			{ID: "supported", Supported: true},
			{ID: "unsupported", Supported: false},
		},
		Scenarios: []Scenario{
			{Name: "built_in_supported", BuiltIn: true, Supported: true},
			{Name: "built_in_unsupported", BuiltIn: true, Supported: false},
			{Name: "payment_failed", BuiltIn: true, Supported: true},
			{Name: "local_custom", BuiltIn: false, Supported: true},
		},
		StateEvidence: &StateEvidence{StatefulResources: []string{"checkout_session"}, Idempotency: true},
	})

	if score.EndpointCoverage != 50 {
		t.Fatalf("endpoint coverage = %d, want 50", score.EndpointCoverage)
	}
	if score.ScenarioCoverage != 66 {
		t.Fatalf("scenario coverage = %d, want 66", score.ScenarioCoverage)
	}
	if score.SDKCoverage != 100 || score.StateCoverage != 100 || score.ErrorCoverage != 100 {
		t.Fatalf("score = %#v, want sdk/state/error coverage", score)
	}
	if score.Total != 83 {
		t.Fatalf("total = %d, want 83", score.Total)
	}
	if score.Level != string(LevelState) {
		t.Fatalf("level = %q, want state", score.Level)
	}
}

func TestCanPromoteRequiresContractEvidenceForProviderCompatible(t *testing.T) {
	manifest := Manifest{
		Adapter:         "stripe",
		ProviderVersion: "2026-05-26",
		Levels:          []Level{LevelWire, LevelSDK, LevelWorkflow, LevelState, LevelError},
		SDKVersions:     []SDKVersion{{Name: "stripe-go", Version: "v83.0.0"}},
		Endpoints:       []Endpoint{{ID: "one", Supported: true}},
		Scenarios: []Scenario{
			{Name: "payment_success", BuiltIn: true, Supported: true},
			{Name: "payment_failed", BuiltIn: true, Supported: true},
		},
		StateEvidence: &StateEvidence{StatefulResources: []string{"payment_intent"}, Idempotency: true},
	}
	score := CalculateScore(manifest)

	if CanPromote(manifest, score, "provider-compatible") {
		t.Fatal("provider-compatible promotion passed without contract level")
	}
	manifest.Levels = append(manifest.Levels, LevelContract)
	score = CalculateScore(manifest)
	if !CanPromote(manifest, score, "provider-compatible") {
		t.Fatalf("provider-compatible promotion failed with contract evidence: %#v", score)
	}
}

func TestStateCoverageRequiresConcreteEvidence(t *testing.T) {
	base := Manifest{
		Adapter:         "stripe",
		ProviderVersion: "2026-05-26",
		Levels:          []Level{LevelWire, LevelState},
		Endpoints:       []Endpoint{{ID: "one", Supported: true}},
		Scenarios:       []Scenario{{Name: "payment_success", BuiltIn: true, Supported: true}},
	}

	if score := CalculateScore(base); score.StateCoverage != 0 {
		t.Fatalf("state coverage = %d with declared level but no evidence, want 0", score.StateCoverage)
	}

	withEvidence := base
	withEvidence.StateEvidence = &StateEvidence{StatefulResources: []string{"checkout_session"}}
	if score := CalculateScore(withEvidence); score.StateCoverage != 100 {
		t.Fatalf("state coverage = %d with stateful resource evidence, want 100", score.StateCoverage)
	}

	withResetOnly := base
	withResetOnly.StateEvidence = &StateEvidence{Reset: true}
	if score := CalculateScore(withResetOnly); score.StateCoverage != 100 {
		t.Fatalf("state coverage = %d with reset evidence, want 100", score.StateCoverage)
	}

	withScenarioLevel := base
	withScenarioLevel.Scenarios = []Scenario{{Name: "payment_success", BuiltIn: true, Supported: true, Levels: []Level{LevelState}}}
	if score := CalculateScore(withScenarioLevel); score.StateCoverage != 100 {
		t.Fatalf("state coverage = %d with scenario state level, want 100", score.StateCoverage)
	}
}

func TestErrorCoverageRequiresConcreteEvidence(t *testing.T) {
	base := Manifest{
		Adapter:         "stripe",
		ProviderVersion: "2026-05-26",
		Levels:          []Level{LevelWire, LevelError},
		Endpoints:       []Endpoint{{ID: "one", Supported: true}},
		Scenarios:       []Scenario{{Name: "payment_success", BuiltIn: true, Supported: true}},
	}

	if score := CalculateScore(base); score.ErrorCoverage != 0 {
		t.Fatalf("error coverage = %d with declared level but no error scenario, want 0", score.ErrorCoverage)
	}

	withErrorScenario := base
	withErrorScenario.Scenarios = append(append([]Scenario(nil), base.Scenarios...), Scenario{Name: "payment_failed", BuiltIn: true, Supported: true})
	if score := CalculateScore(withErrorScenario); score.ErrorCoverage != 100 {
		t.Fatalf("error coverage = %d with error scenario, want 100", score.ErrorCoverage)
	}

	withCategory := base
	withCategory.Scenarios = append(append([]Scenario(nil), base.Scenarios...), Scenario{Name: "edge_case", BuiltIn: true, Supported: true, Category: "error"})
	if score := CalculateScore(withCategory); score.ErrorCoverage != 100 {
		t.Fatalf("error coverage = %d with error-category scenario, want 100", score.ErrorCoverage)
	}

	unsupportedError := base
	unsupportedError.Scenarios = append(append([]Scenario(nil), base.Scenarios...), Scenario{Name: "payment_failed", BuiltIn: true, Supported: false})
	if score := CalculateScore(unsupportedError); score.ErrorCoverage != 0 {
		t.Fatalf("error coverage = %d with unsupported error scenario, want 0", score.ErrorCoverage)
	}
}

func TestCanPromoteWorkflowRequiresStateAndErrorEvidence(t *testing.T) {
	manifest := Manifest{
		Adapter:         "stripe",
		ProviderVersion: "2026-05-26",
		Levels:          []Level{LevelWire, LevelSDK, LevelWorkflow, LevelState, LevelError},
		SDKVersions:     []SDKVersion{{Name: "stripe-go", Version: "v83.0.0"}},
		Endpoints:       []Endpoint{{ID: "one", Supported: true}},
		Scenarios:       []Scenario{{Name: "payment_success", BuiltIn: true, Supported: true}},
	}

	score := CalculateScore(manifest)
	if CanPromote(manifest, score, "workflow-compatible") {
		t.Fatal("workflow-compatible promotion passed without state/error evidence")
	}

	manifest.StateEvidence = &StateEvidence{StatefulResources: []string{"payment_intent"}, Idempotency: true}
	manifest.Scenarios = append(manifest.Scenarios, Scenario{Name: "payment_failed", BuiltIn: true, Supported: true})
	score = CalculateScore(manifest)
	if !CanPromote(manifest, score, "workflow-compatible") {
		t.Fatalf("workflow-compatible promotion failed with concrete evidence: %#v", score)
	}
}

func TestCalculateScoreUsesClientEvidence(t *testing.T) {
	score := CalculateScore(Manifest{
		Adapter:         "slack",
		ProviderVersion: "2025-02-01",
		Levels:          []Level{LevelWire, LevelClient, LevelWorkflow, LevelState, LevelError},
		ClientEvidence:  []string{"slack-client-contract"},
		Endpoints:       []Endpoint{{ID: "chat_post_message", Supported: true}},
		Scenarios: []Scenario{
			{Name: "message_success", BuiltIn: true, Supported: true},
			{Name: "rate_limited", BuiltIn: true, Supported: true},
		},
		StateEvidence: &StateEvidence{StatefulResources: []string{"conversation"}, Reset: true},
	})

	if score.SDKCoverage != 100 {
		t.Fatalf("sdk/client coverage = %d, want 100", score.SDKCoverage)
	}
	if score.Total != 100 {
		t.Fatalf("total = %d, want 100", score.Total)
	}
	if score.Level != string(LevelWorkflow) {
		t.Fatalf("level = %q, want workflow", score.Level)
	}
}

func TestCalculateScoreReportsWorkflowAsHigherThanState(t *testing.T) {
	score := CalculateScore(Manifest{
		Adapter:         "stripe",
		ProviderVersion: "2025-10-29.clover",
		Levels:          []Level{LevelWire, LevelSDK, LevelWorkflow, LevelState, LevelError},
		SDKVersions:     []SDKVersion{{Name: "stripe", Version: "22.1.1"}},
		Endpoints:       []Endpoint{{ID: "checkout_sessions_create", Supported: true}},
		Scenarios:       []Scenario{{Name: "payment_success", BuiltIn: true, Supported: true}},
	})

	if score.Level != string(LevelWorkflow) {
		t.Fatalf("level = %q, want workflow", score.Level)
	}
}

// TestCalculateScoreFromMetadataRequiresConcreteStateAndErrorEvidence は runtime-report
// パス（FromMetadata 経由）でも、宣言しただけの state/error level が coverage を
// 過大評価しないことを固定する（#21）。FromMetadata が endpoint/scenario へ adapter 全体の
// levels をバックフィルしていた頃は、具体証跡ゼロでも 100 になっていた。
func TestCalculateScoreFromMetadataRequiresConcreteStateAndErrorEvidence(t *testing.T) {
	// state/error を宣言しているが、具体証跡（StatefulResources/Idempotency/Reset/
	// error scenario）は一切持たないメタデータ。
	declaredOnly := adapter.Metadata{
		Name:            "openai",
		Maturity:        adapter.MaturityExperimental,
		ProviderVersion: "2025-10-29.clover",
		Levels:          []adapter.Level{adapter.LevelWire, adapter.LevelState, adapter.LevelError},
		Scenarios:       []adapter.Scenario{{Name: "chat_success", Supported: true}},
		Endpoints:       []adapter.Endpoint{{Method: http.MethodPost, Path: "/openai/v1/chat/completions", SupportedScenarios: []string{"chat_success"}}},
	}

	score := CalculateScore(FromMetadata(declaredOnly))
	if score.StateCoverage != 0 {
		t.Fatalf("state coverage = %d from declared level without evidence, want 0", score.StateCoverage)
	}
	if score.ErrorCoverage != 0 {
		t.Fatalf("error coverage = %d from declared level without evidence, want 0", score.ErrorCoverage)
	}

	// 具体証跡（fake-state surface と error scenario）を加えると 100 になる。
	withEvidence := declaredOnly
	withEvidence.StatefulResources = []string{"thread"}
	withEvidence.Idempotency = true
	withEvidence.Scenarios = append(append([]adapter.Scenario(nil), declaredOnly.Scenarios...),
		adapter.Scenario{Name: "chat_failed", Supported: true, Category: adapter.ScenarioCategoryError})

	score = CalculateScore(FromMetadata(withEvidence))
	if score.StateCoverage != 100 {
		t.Fatalf("state coverage = %d with concrete state evidence, want 100", score.StateCoverage)
	}
	if score.ErrorCoverage != 100 {
		t.Fatalf("error coverage = %d with error scenario, want 100", score.ErrorCoverage)
	}
}

// TestNonBuiltInScenariosDoNotInflateCoverage は、supported でも BuiltIn=false の
// local/custom scenario が error 風の名前や state level を持っていても evidence として
// 数えないことを固定する。scenarioCoverage が built-in のみを対象とするのと揃え、
// ユーザー定義 scenario だけで state/error coverage や workflow 昇格を自己申告で
// 水増しできないようにする。
func TestNonBuiltInScenariosDoNotInflateCoverage(t *testing.T) {
	manifest := Manifest{
		Adapter:         "stripe",
		ProviderVersion: "2026-05-26",
		Levels:          []Level{LevelWire, LevelWorkflow, LevelState, LevelError},
		Endpoints:       []Endpoint{{ID: "one", Supported: true}},
		Scenarios: []Scenario{
			{Name: "payment_success", BuiltIn: true, Supported: true},
			{Name: "local_timeout_repro", BuiltIn: false, Supported: true},
			{Name: "local_state_repro", BuiltIn: false, Supported: true, Levels: []Level{LevelState}},
		},
	}

	score := CalculateScore(manifest)
	if score.ErrorCoverage != 0 {
		t.Fatalf("error coverage = %d from non-built-in error-like scenario, want 0", score.ErrorCoverage)
	}
	if score.StateCoverage != 0 {
		t.Fatalf("state coverage = %d from non-built-in scenario state level, want 0", score.StateCoverage)
	}
	if CanPromote(manifest, score, "workflow-compatible") {
		t.Fatal("workflow-compatible promotion passed using only non-built-in scenarios")
	}

	// 同じ証跡を built-in scenario として宣言すれば evidence として数える。
	builtIn := manifest
	builtIn.Scenarios = []Scenario{
		{Name: "payment_success", BuiltIn: true, Supported: true},
		{Name: "payment_failed", BuiltIn: true, Supported: true},
		{Name: "state_repro", BuiltIn: true, Supported: true, Levels: []Level{LevelState}},
	}
	score = CalculateScore(builtIn)
	if score.ErrorCoverage != 100 {
		t.Fatalf("error coverage = %d from built-in error scenario, want 100", score.ErrorCoverage)
	}
	if score.StateCoverage != 100 {
		t.Fatalf("state coverage = %d from built-in scenario state level, want 100", score.StateCoverage)
	}
}

// TestCanPromoteProviderCompatibleRequiresConcreteEvidence は、最上位 maturity の
// provider-compatible が total 閾値だけで昇格できないことを固定する。contract level と
// total>=80 を満たしても、error/state/SDK のいずれかの具体証跡が欠けていれば昇格しない。
func TestCanPromoteProviderCompatibleRequiresConcreteEvidence(t *testing.T) {
	// SDK=100 / State=100 / Endpoint=100 / Scenario=100 だが Error=0。total はちょうど 80。
	manifest := Manifest{
		Adapter:         "stripe",
		ProviderVersion: "2026-05-26",
		Levels:          []Level{LevelWire, LevelSDK, LevelState, LevelContract},
		SDKVersions:     []SDKVersion{{Name: "stripe-go", Version: "v83.0.0"}},
		Endpoints:       []Endpoint{{ID: "one", Supported: true}},
		Scenarios:       []Scenario{{Name: "payment_success", BuiltIn: true, Supported: true}},
		StateEvidence:   &StateEvidence{StatefulResources: []string{"payment_intent"}, Idempotency: true},
	}

	score := CalculateScore(manifest)
	if score.Total < 80 {
		t.Fatalf("precondition: total = %d, want >= 80", score.Total)
	}
	if score.ErrorCoverage != 0 {
		t.Fatalf("precondition: error coverage = %d, want 0", score.ErrorCoverage)
	}
	if CanPromote(manifest, score, "provider-compatible") {
		t.Fatal("provider-compatible promotion passed without error evidence")
	}

	// workflow/error level と built-in error scenario を加えると、全証跡が揃い昇格できる。
	full := manifest
	full.Levels = append(append([]Level(nil), manifest.Levels...), LevelWorkflow, LevelError)
	full.Scenarios = append(append([]Scenario(nil), manifest.Scenarios...), Scenario{Name: "payment_failed", BuiltIn: true, Supported: true})
	score = CalculateScore(full)
	if !CanPromote(full, score, "provider-compatible") {
		t.Fatalf("provider-compatible promotion failed with full concrete evidence: %#v", score)
	}

	// workflow level が欠けていれば、他の証跡が揃っていても provider-compatible にならない
	// （最上位 maturity は workflow-compatible 条件を包含する）。
	noWorkflow := full
	noWorkflow.Levels = []Level{LevelWire, LevelSDK, LevelState, LevelError, LevelContract}
	score = CalculateScore(noWorkflow)
	if CanPromote(noWorkflow, score, "provider-compatible") {
		t.Fatal("provider-compatible promotion passed without workflow level")
	}
}

// TestSDKCoverageRejectsEmptyEvidence は、空文字の SDK version / client evidence では
// SDKCoverage が 100 にならないことを固定する。空証跡で sdk-compatible や
// provider-compatible の SDK 条件を満たせてしまう穴を防ぐ。
func TestSDKCoverageRejectsEmptyEvidence(t *testing.T) {
	base := Manifest{
		Adapter:         "x",
		ProviderVersion: "1",
		Levels:          []Level{LevelWire, LevelSDK},
		Endpoints:       []Endpoint{{ID: "one", Supported: true}},
		Scenarios:       []Scenario{{Name: "ok", BuiltIn: true, Supported: true}},
	}

	emptySDK := base
	emptySDK.SDKVersions = []SDKVersion{{Name: "", Version: ""}}
	if score := CalculateScore(emptySDK); score.SDKCoverage != 0 {
		t.Fatalf("sdk coverage = %d with empty sdk version, want 0", score.SDKCoverage)
	}

	emptyClient := base
	emptyClient.Levels = []Level{LevelWire, LevelClient}
	emptyClient.ClientEvidence = []string{"", "   "}
	if score := CalculateScore(emptyClient); score.SDKCoverage != 0 {
		t.Fatalf("sdk coverage = %d with empty client evidence, want 0", score.SDKCoverage)
	}

	validSDK := base
	validSDK.SDKVersions = []SDKVersion{{Name: "lib", Version: "1.0.0"}}
	if score := CalculateScore(validSDK); score.SDKCoverage != 100 {
		t.Fatalf("sdk coverage = %d with valid sdk version, want 100", score.SDKCoverage)
	}
}
