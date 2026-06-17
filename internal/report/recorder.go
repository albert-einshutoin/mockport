package report

import (
	"slices"
	"strings"
	"sync"
	"time"
)

type Recorder struct {
	mu             sync.Mutex
	now            func() time.Time
	mode           string
	adapters       []AdapterStatus
	requests       []Request
	safetyWarnings []SafetyWarning
	coverage       []ScenarioCoverage
	matrix         []BehaviorMatrixEntry
	compatibility  []CompatibilityStatus
	stateCoverage  []StateCoverageStatus
	nextID         int64
}

const MaxRecordedRequests = 1000

func NewRecorder() *Recorder {
	return &Recorder{now: time.Now}
}

func (r *Recorder) SetClock(now func() time.Time) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if now == nil {
		r.now = time.Now
		return
	}
	r.now = now
}

func (r *Recorder) SetMode(mode string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.mode = mode
}

func (r *Recorder) SetAdapters(adapters []AdapterStatus) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.adapters = slices.Clone(adapters)
}

func (r *Recorder) SetScenarioCoverage(coverage []ScenarioCoverage) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.coverage = cloneScenarioCoverage(coverage)
}

func (r *Recorder) SetBehaviorMatrix(matrix []BehaviorMatrixEntry) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.matrix = cloneBehaviorMatrix(matrix)
}

func (r *Recorder) SetCompatibility(compatibility []CompatibilityStatus) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.compatibility = cloneCompatibility(compatibility)
}

func (r *Recorder) SetStateCoverage(coverage []StateCoverageStatus) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.stateCoverage = cloneStateCoverage(coverage)
}

func (r *Recorder) RecordRequest(method, path string, status int) {
	r.RecordRequestWithDetails(method, path, status, "", "", "")
}

func (r *Recorder) RecordRequestWithDetails(method, path string, status int, adapter, scenario, reason string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.nextID++
	now := r.now
	r.requests = append(r.requests, Request{
		ID:        r.nextID,
		Timestamp: now().UTC().Format(time.RFC3339),
		Method:    method,
		Path:      path,
		Status:    status,
		Adapter:   adapter,
		Scenario:  scenario,
		Reason:    reason,
	})
	if len(r.requests) > MaxRecordedRequests {
		r.requests = r.requests[len(r.requests)-MaxRecordedRequests:]
	}
}

func (r *Recorder) RecordSafetyWarning(field, category, message string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.safetyWarnings = append(r.safetyWarnings, SafetyWarning{Field: field, Category: category, Message: message})
}

func (r *Recorder) Snapshot() Snapshot {
	r.mu.Lock()
	defer r.mu.Unlock()
	return Snapshot{
		Mode:                 r.mode,
		Safety:               safetySummary(r.mode, r.safetyWarnings),
		Adapters:             slices.Clone(r.adapters),
		Requests:             slices.Clone(r.requests),
		SafetyWarnings:       slices.Clone(r.safetyWarnings),
		ScenarioCoverage:     cloneScenarioCoverage(r.coverage),
		BehaviorMatrix:       cloneBehaviorMatrix(r.matrix),
		Compatibility:        cloneCompatibility(r.compatibility),
		StateCoverage:        cloneStateCoverage(r.stateCoverage),
		UnsupportedEndpoints: unsupportedEndpoints(r.requests),
	}
}

func safetySummary(mode string, warnings []SafetyWarning) SafetySummary {
	summary := SafetySummary{Mode: mode, Safe: len(warnings) == 0, PublicEnvSafe: len(warnings) == 0}
	for _, warning := range warnings {
		switch warning.Category {
		case "real_looking_secret":
			summary.RealLookingSecrets++
		case "external_url":
			summary.ExternalURLs++
		}
	}
	return summary
}

func unsupportedEndpoints(requests []Request) []UnsupportedEndpoint {
	var unsupported []UnsupportedEndpoint
	for _, request := range requests {
		if strings.TrimSpace(request.Reason) == "" {
			continue
		}
		unsupported = append(unsupported, UnsupportedEndpoint{
			Method: request.Method,
			Path:   request.Path,
			Status: request.Status,
			Reason: request.Reason,
		})
	}
	return unsupported
}

func cloneScenarioCoverage(in []ScenarioCoverage) []ScenarioCoverage {
	out := slices.Clone(in)
	for i := range out {
		out[i].Scenarios = slices.Clone(out[i].Scenarios)
	}
	return out
}

func cloneBehaviorMatrix(in []BehaviorMatrixEntry) []BehaviorMatrixEntry {
	out := slices.Clone(in)
	for i := range out {
		out[i].SupportedScenarios = slices.Clone(out[i].SupportedScenarios)
	}
	return out
}

func cloneCompatibility(in []CompatibilityStatus) []CompatibilityStatus {
	out := slices.Clone(in)
	for i := range out {
		out[i].SDKVersions = slices.Clone(out[i].SDKVersions)
		out[i].ClientEvidence = slices.Clone(out[i].ClientEvidence)
		if out[i].ContractEvidence != nil {
			out[i].ContractEvidence = &ContractEvidence{
				Fixtures:     slices.Clone(out[i].ContractEvidence.Fixtures),
				SDKContracts: slices.Clone(out[i].ContractEvidence.SDKContracts),
				KnownGaps:    slices.Clone(out[i].ContractEvidence.KnownGaps),
			}
		}
		out[i].UnsupportedEndpoints = slices.Clone(out[i].UnsupportedEndpoints)
	}
	return out
}

func cloneStateCoverage(in []StateCoverageStatus) []StateCoverageStatus {
	out := slices.Clone(in)
	for i := range out {
		out[i].StatefulResources = slices.Clone(out[i].StatefulResources)
	}
	return out
}
