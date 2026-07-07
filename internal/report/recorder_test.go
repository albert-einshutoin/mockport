package report

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestRecorderStoresRequestsAndSafety(t *testing.T) {
	rec := NewRecorder()
	rec.RecordRequest(http.MethodPost, "/stripe/v1/checkout/sessions", 200)
	rec.RecordSafetyWarning("STRIPE_SECRET_KEY", "real_looking_secret", "real-looking Stripe key")

	snapshot := rec.Snapshot()
	if len(snapshot.Requests) != 1 {
		t.Fatalf("request count = %d, want 1", len(snapshot.Requests))
	}
	if snapshot.Requests[0].Path != "/stripe/v1/checkout/sessions" {
		t.Fatalf("path = %q", snapshot.Requests[0].Path)
	}
	if len(snapshot.SafetyWarnings) != 1 {
		t.Fatalf("warning count = %d, want 1", len(snapshot.SafetyWarnings))
	}
	if snapshot.Safety.Safe {
		t.Fatal("safety summary safe = true, want false")
	}
	if snapshot.Safety.RealLookingSecrets != 1 {
		t.Fatalf("real-looking secret count = %d, want 1", snapshot.Safety.RealLookingSecrets)
	}
}

func TestRecorderStoresReplayMetadataAndUnsupportedEndpoints(t *testing.T) {
	rec := NewRecorder()
	rec.RecordRequestWithDetails(http.MethodPost, "/stripe/not-supported", 404, "stripe", "payment_success", "unsupported_endpoint")

	snapshot := rec.Snapshot()
	if len(snapshot.Requests) != 1 {
		t.Fatalf("request count = %d, want 1", len(snapshot.Requests))
	}
	req := snapshot.Requests[0]
	if req.ID != 1 {
		t.Fatalf("request id = %d, want 1", req.ID)
	}
	if req.Timestamp == "" {
		t.Fatal("request timestamp is empty")
	}
	if req.Adapter != "stripe" || req.Scenario != "payment_success" {
		t.Fatalf("request metadata = %#v", req)
	}
	if len(snapshot.UnsupportedEndpoints) != 1 {
		t.Fatalf("unsupported endpoints = %d, want 1", len(snapshot.UnsupportedEndpoints))
	}
}

func TestRecorderCapsStoredRequests(t *testing.T) {
	rec := NewRecorder()
	total := MaxRecordedRequests + 5
	for i := 0; i < total; i++ {
		path := fmt.Sprintf("/requests/%d", i+1)
		rec.RecordRequest(http.MethodGet, path, http.StatusOK)
	}

	snapshot := rec.Snapshot()
	if len(snapshot.Requests) != MaxRecordedRequests {
		t.Fatalf("request count = %d, want %d", len(snapshot.Requests), MaxRecordedRequests)
	}
	if snapshot.Requests[0].ID != 6 {
		t.Fatalf("first retained request id = %d, want 6", snapshot.Requests[0].ID)
	}
	if snapshot.Requests[0].Path != "/requests/6" {
		t.Fatalf("first retained path = %q, want /requests/6", snapshot.Requests[0].Path)
	}
	last := snapshot.Requests[len(snapshot.Requests)-1]
	if last.ID != int64(total) {
		t.Fatalf("last retained request id = %d, want %d", last.ID, total)
	}
	if wantPath := fmt.Sprintf("/requests/%d", total); last.Path != wantPath {
		t.Fatalf("last retained path = %q, want %q", last.Path, wantPath)
	}
	for i, req := range snapshot.Requests {
		wantID := int64(i + 6)
		if req.ID != wantID {
			t.Fatalf("requests[%d].ID = %d, want %d", i, req.ID, wantID)
		}
		wantPath := fmt.Sprintf("/requests/%d", i+6)
		if req.Path != wantPath {
			t.Fatalf("requests[%d].Path = %q, want %q", i, req.Path, wantPath)
		}
	}
	if !snapshot.RequestHistory.Truncated {
		t.Fatal("request history truncated = false, want true")
	}
	if snapshot.RequestHistory.Limit != MaxRecordedRequests {
		t.Fatalf("request history limit = %d, want %d", snapshot.RequestHistory.Limit, MaxRecordedRequests)
	}
	if snapshot.RequestHistory.Retained != MaxRecordedRequests {
		t.Fatalf("request history retained = %d, want %d", snapshot.RequestHistory.Retained, MaxRecordedRequests)
	}
	if snapshot.RequestHistory.Evicted != 5 {
		t.Fatalf("request history evicted = %d, want 5", snapshot.RequestHistory.Evicted)
	}
}

func TestRecorderRespectsRequestHistoryEnvOverride(t *testing.T) {
	t.Setenv("MOCKPORT_REQUEST_HISTORY", "3")

	rec := NewRecorder()
	for i := 0; i < 5; i++ {
		rec.RecordRequest(http.MethodGet, fmt.Sprintf("/requests/%d", i+1), http.StatusOK)
	}

	snapshot := rec.Snapshot()
	if len(snapshot.Requests) != 3 {
		t.Fatalf("request count = %d, want 3", len(snapshot.Requests))
	}
	if snapshot.Requests[0].ID != 3 {
		t.Fatalf("first retained request id = %d, want 3", snapshot.Requests[0].ID)
	}
	if snapshot.RequestHistory.Limit != 3 {
		t.Fatalf("request history limit = %d, want 3", snapshot.RequestHistory.Limit)
	}
	if snapshot.RequestHistory.Retained != 3 {
		t.Fatalf("request history retained = %d, want 3", snapshot.RequestHistory.Retained)
	}
	if snapshot.RequestHistory.Evicted != 2 {
		t.Fatalf("request history evicted = %d, want 2", snapshot.RequestHistory.Evicted)
	}
	if !snapshot.RequestHistory.Truncated {
		t.Fatal("request history truncated = false, want true")
	}
}

func TestRecorderFallsBackToDefaultRequestHistoryLimitForInvalidEnv(t *testing.T) {
	tests := []struct {
		name string
		env  string
	}{
		{name: "empty", env: ""},
		{name: "zero", env: "0"},
		{name: "negative", env: "-5"},
		{name: "non_numeric", env: "abc"},
		{name: "overflow", env: "999999999999999999999999"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("MOCKPORT_REQUEST_HISTORY", tc.env)

			rec := NewRecorder()
			total := MaxRecordedRequests + 5
			for i := 0; i < total; i++ {
				rec.RecordRequest(http.MethodGet, fmt.Sprintf("/requests/%d", i+1), http.StatusOK)
			}

			snapshot := rec.Snapshot()
			if len(snapshot.Requests) != MaxRecordedRequests {
				t.Fatalf("request count = %d, want %d", len(snapshot.Requests), MaxRecordedRequests)
			}
			if snapshot.RequestHistory.Limit != MaxRecordedRequests {
				t.Fatalf("request history limit = %d, want %d", snapshot.RequestHistory.Limit, MaxRecordedRequests)
			}
			if snapshot.RequestHistory.Evicted != 5 {
				t.Fatalf("request history evicted = %d, want 5", snapshot.RequestHistory.Evicted)
			}
		})
	}
}

func TestRecorderRequestHistoryNotTruncatedWithinLimit(t *testing.T) {
	rec := NewRecorder()
	rec.RecordRequest(http.MethodGet, "/health", http.StatusOK)

	snapshot := rec.Snapshot()
	if snapshot.RequestHistory.Truncated {
		t.Fatal("request history truncated = true, want false")
	}
	if snapshot.RequestHistory.Evicted != 0 {
		t.Fatalf("request history evicted = %d, want 0", snapshot.RequestHistory.Evicted)
	}
	if snapshot.RequestHistory.Retained != 1 {
		t.Fatalf("request history retained = %d, want 1", snapshot.RequestHistory.Retained)
	}
	if snapshot.RequestHistory.Limit != MaxRecordedRequests {
		t.Fatalf("request history limit = %d, want %d", snapshot.RequestHistory.Limit, MaxRecordedRequests)
	}
}

func TestRecorderStoresCompatibility(t *testing.T) {
	rec := NewRecorder()
	rec.SetCompatibility([]CompatibilityStatus{{
		Adapter:         "stripe",
		Level:           "wire",
		Score:           80,
		ProviderVersion: "2026-05-26",
		SDKVersions:     []string{"stripe-go@v83.0.0"},
		ClientEvidence:  []string{"stripe-node-contract"},
	}})

	snapshot := rec.Snapshot()
	if len(snapshot.Compatibility) != 1 {
		t.Fatalf("compatibility count = %d, want 1", len(snapshot.Compatibility))
	}
	if snapshot.Compatibility[0].Adapter != "stripe" || snapshot.Compatibility[0].Score != 80 {
		t.Fatalf("compatibility = %#v", snapshot.Compatibility[0])
	}
}

func TestRecorderSnapshotDeepCopiesNestedSlices(t *testing.T) {
	rec := NewRecorder()
	rec.SetScenarioCoverage([]ScenarioCoverage{{Adapter: "stripe", Scenarios: []ScenarioSupport{{Name: "payment_success", Supported: true}}}})
	rec.SetBehaviorMatrix([]BehaviorMatrixEntry{{Adapter: "stripe", SupportedScenarios: []string{"payment_success"}}})
	rec.SetCompatibility([]CompatibilityStatus{{
		Adapter:        "stripe",
		SDKVersions:    []string{"stripe@22.1.1"},
		ClientEvidence: []string{"stripe-node-contract"},
		ContractEvidence: &ContractEvidence{
			Fixtures:     []string{"compat/fixtures/stripe/checkout_session_create.json"},
			SDKContracts: []string{"contract/sdk/stripe"},
			KnownGaps:    []string{"docs/compatibility-reports/latest.json#stripe"},
		},
		UnsupportedEndpoints: []string{"post_v1_missing"},
	}})
	rec.SetStateCoverage([]StateCoverageStatus{{Adapter: "stripe", StatefulResources: []string{"checkout_session"}}})

	first := rec.Snapshot()
	first.ScenarioCoverage[0].Scenarios[0].Name = "mutated"
	first.BehaviorMatrix[0].SupportedScenarios[0] = "mutated"
	first.Compatibility[0].SDKVersions[0] = "mutated"
	first.Compatibility[0].ClientEvidence[0] = "mutated"
	first.Compatibility[0].ContractEvidence.Fixtures[0] = "mutated"
	first.Compatibility[0].ContractEvidence.SDKContracts[0] = "mutated"
	first.Compatibility[0].ContractEvidence.KnownGaps[0] = "mutated"
	first.Compatibility[0].UnsupportedEndpoints[0] = "mutated"
	first.StateCoverage[0].StatefulResources[0] = "mutated"

	second := rec.Snapshot()
	if second.ScenarioCoverage[0].Scenarios[0].Name != "payment_success" {
		t.Fatalf("scenario coverage was mutated through snapshot: %#v", second.ScenarioCoverage)
	}
	if second.BehaviorMatrix[0].SupportedScenarios[0] != "payment_success" {
		t.Fatalf("behavior matrix was mutated through snapshot: %#v", second.BehaviorMatrix)
	}
	if second.Compatibility[0].SDKVersions[0] != "stripe@22.1.1" {
		t.Fatalf("compatibility SDK versions were mutated through snapshot: %#v", second.Compatibility)
	}
	if second.Compatibility[0].ClientEvidence[0] != "stripe-node-contract" {
		t.Fatalf("compatibility client evidence was mutated through snapshot: %#v", second.Compatibility)
	}
	if second.Compatibility[0].ContractEvidence.Fixtures[0] != "compat/fixtures/stripe/checkout_session_create.json" {
		t.Fatalf("compatibility contract fixture evidence was mutated through snapshot: %#v", second.Compatibility)
	}
	if second.Compatibility[0].ContractEvidence.SDKContracts[0] != "contract/sdk/stripe" {
		t.Fatalf("compatibility contract SDK evidence was mutated through snapshot: %#v", second.Compatibility)
	}
	if second.Compatibility[0].ContractEvidence.KnownGaps[0] != "docs/compatibility-reports/latest.json#stripe" {
		t.Fatalf("compatibility contract known-gap evidence was mutated through snapshot: %#v", second.Compatibility)
	}
	if second.Compatibility[0].UnsupportedEndpoints[0] != "post_v1_missing" {
		t.Fatalf("compatibility unsupported endpoints were mutated through snapshot: %#v", second.Compatibility)
	}
	if second.StateCoverage[0].StatefulResources[0] != "checkout_session" {
		t.Fatalf("state coverage was mutated through snapshot: %#v", second.StateCoverage)
	}
}

func TestRecorderStoresStateCoverage(t *testing.T) {
	rec := NewRecorder()
	rec.SetStateCoverage([]StateCoverageStatus{{
		Adapter:           "stripe",
		StatefulResources: []string{"checkout_session", "payment_intent"},
		Idempotency:       true,
		Reset:             true,
	}})

	snapshot := rec.Snapshot()
	if len(snapshot.StateCoverage) != 1 {
		t.Fatalf("state coverage count = %d, want 1", len(snapshot.StateCoverage))
	}
	got := snapshot.StateCoverage[0]
	if got.Adapter != "stripe" || !got.Idempotency || !got.Reset {
		t.Fatalf("state coverage = %#v", got)
	}
	if len(got.StatefulResources) != 2 {
		t.Fatalf("stateful resources = %#v", got.StatefulResources)
	}
}

func TestRecorderUsesInjectedClockForDeterministicRequests(t *testing.T) {
	rec := NewRecorder()
	rec.SetClock(func() time.Time {
		return time.Date(2026, 5, 26, 12, 0, 0, 0, time.UTC)
	})

	rec.RecordRequest(http.MethodGet, "/health", http.StatusOK)

	snapshot := rec.Snapshot()
	if got := snapshot.Requests[0].Timestamp; got != "2026-05-26T12:00:00Z" {
		t.Fatalf("timestamp = %q", got)
	}
}
