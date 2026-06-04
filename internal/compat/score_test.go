package compat

import "testing"

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
