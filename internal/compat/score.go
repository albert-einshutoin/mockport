package compat

import (
	"slices"
	"strings"
)

type Score struct {
	Adapter          string `json:"adapter"`
	Level            string `json:"level"`
	Total            int    `json:"total"`
	EndpointCoverage int    `json:"endpoint_coverage"`
	ScenarioCoverage int    `json:"scenario_coverage"`
	SDKCoverage      int    `json:"sdk_coverage"`
	StateCoverage    int    `json:"state_coverage"`
	ErrorCoverage    int    `json:"error_coverage"`
}

func CalculateScore(manifest Manifest) Score {
	score := Score{
		Adapter:          manifest.Adapter,
		Level:            string(highestLevel(manifest.Levels)),
		EndpointCoverage: endpointCoverage(manifest.Endpoints),
		ScenarioCoverage: scenarioCoverage(manifest.Scenarios),
	}
	if hasLevel(manifest.Levels, LevelSDK) && hasSDKEvidence(manifest.SDKVersions) {
		score.SDKCoverage = 100
	}
	if hasLevel(manifest.Levels, LevelClient) && hasClientEvidence(manifest.ClientEvidence) {
		score.SDKCoverage = 100
	}
	if hasLevel(manifest.Levels, LevelState) && hasStateEvidence(manifest) {
		score.StateCoverage = 100
	}
	if hasLevel(manifest.Levels, LevelError) && hasErrorEvidence(manifest) {
		score.ErrorCoverage = 100
	}
	score.Total = (score.EndpointCoverage + score.ScenarioCoverage + score.SDKCoverage + score.StateCoverage + score.ErrorCoverage) / 5
	return score
}

// hasStateEvidence reports whether the manifest carries concrete fake-state
// evidence. A declared state level alone is not enough: there must be at least
// one stateful resource, idempotency/reset support, an endpoint that explicitly
// claims the state level, or a supported built-in scenario that does. Only
// built-in scenarios count, matching scenarioCoverage, so a user-defined local
// scenario cannot self-inflate state coverage.
func hasStateEvidence(manifest Manifest) bool {
	if manifest.StateEvidence.HasEvidence() {
		return true
	}
	for _, endpoint := range manifest.Endpoints {
		if endpoint.Supported && hasLevel(endpoint.Levels, LevelState) {
			return true
		}
	}
	for _, scenario := range manifest.Scenarios {
		if scenario.Supported && scenario.BuiltIn && hasLevel(scenario.Levels, LevelState) {
			return true
		}
	}
	return false
}

// hasErrorEvidence reports whether the manifest carries concrete error-behavior
// evidence. A declared error level alone is not enough: there must be at least
// one supported built-in error scenario, or an endpoint that explicitly claims
// the error level. Only built-in scenarios count (matching scenarioCoverage), so
// a user-defined local scenario with an error-like name cannot self-inflate error
// coverage or unlock workflow promotion.
func hasErrorEvidence(manifest Manifest) bool {
	for _, scenario := range manifest.Scenarios {
		if !scenario.Supported || !scenario.BuiltIn {
			continue
		}
		if isErrorScenario(scenario) {
			return true
		}
		if hasLevel(scenario.Levels, LevelError) {
			return true
		}
	}
	for _, endpoint := range manifest.Endpoints {
		if endpoint.Supported && hasLevel(endpoint.Levels, LevelError) {
			return true
		}
	}
	return false
}

// errorScenarioMarkers are name fragments that identify a scenario as concrete
// error-behavior evidence when an explicit category is not provided.
var errorScenarioMarkers = []string{"error", "fail", "denied", "rate_limit", "rate_limited", "timeout", "unauthorized", "forbidden", "invalid", "conflict"}

// isErrorScenario reports whether a scenario represents error-behavior evidence.
// It prefers the explicit category and falls back to name-based heuristics so
// existing adapters keep contributing error evidence without metadata changes.
func isErrorScenario(scenario Scenario) bool {
	if scenario.Category == ScenarioCategoryError {
		return true
	}
	name := strings.ToLower(scenario.Name)
	return slices.ContainsFunc(errorScenarioMarkers, func(marker string) bool {
		return strings.Contains(name, marker)
	})
}

// hasSDKEvidence reports whether at least one SDK version carries a concrete,
// non-empty name and version. Empty entries must not inflate SDK coverage.
func hasSDKEvidence(versions []SDKVersion) bool {
	for _, v := range versions {
		if strings.TrimSpace(v.Name) != "" && strings.TrimSpace(v.Version) != "" {
			return true
		}
	}
	return false
}

// hasClientEvidence reports whether at least one client-evidence entry is a
// concrete, non-empty value.
func hasClientEvidence(evidence []string) bool {
	for _, e := range evidence {
		if strings.TrimSpace(e) != "" {
			return true
		}
	}
	return false
}

// meetsSDKCompatible reports whether the manifest clears the SDK/client evidence
// bar shared by sdk-compatible and every higher maturity.
func meetsSDKCompatible(manifest Manifest, score Score) bool {
	return (hasLevel(manifest.Levels, LevelSDK) || hasLevel(manifest.Levels, LevelClient)) &&
		score.SDKCoverage == 100
}

// meetsWorkflowCompatible reports whether the manifest clears the state/error
// evidence bar shared by workflow-compatible and provider-compatible.
func meetsWorkflowCompatible(manifest Manifest, score Score) bool {
	return hasLevel(manifest.Levels, LevelWorkflow) &&
		hasLevel(manifest.Levels, LevelState) && score.StateCoverage == 100 &&
		hasLevel(manifest.Levels, LevelError) && score.ErrorCoverage == 100
}

func CanPromote(manifest Manifest, score Score, target string) bool {
	switch target {
	case "experimental":
		return true
	case "sdk-compatible":
		return meetsSDKCompatible(manifest, score) && score.Total >= 40
	case "workflow-compatible":
		return meetsWorkflowCompatible(manifest, score) && score.Total >= 60
	case "provider-compatible":
		// Top maturity: subsumes the lower (sdk / workflow) evidence bars and also
		// requires concrete contract evidence and total>=80, so promotion cannot skip
		// the hierarchy or self-promote through a bare contract-level declaration.
		return meetsSDKCompatible(manifest, score) &&
			meetsWorkflowCompatible(manifest, score) &&
			hasLevel(manifest.Levels, LevelContract) &&
			manifest.ContractEvidence.HasEvidence() &&
			score.Total >= 80
	default:
		return false
	}
}

func endpointCoverage(endpoints []Endpoint) int {
	if len(endpoints) == 0 {
		return 0
	}
	supported := 0
	for _, endpoint := range endpoints {
		if endpoint.Supported {
			supported++
		}
	}
	return supported * 100 / len(endpoints)
}

func scenarioCoverage(scenarios []Scenario) int {
	total := 0
	supported := 0
	for _, scenario := range scenarios {
		if !scenario.BuiltIn {
			continue
		}
		total++
		if scenario.Supported {
			supported++
		}
	}
	if total == 0 {
		return 0
	}
	return supported * 100 / total
}

func highestLevel(levels []Level) Level {
	for _, level := range []Level{LevelContract, LevelWorkflow, LevelState, LevelSDK, LevelWire, LevelError} {
		if hasLevel(levels, level) {
			return level
		}
	}
	return LevelWire
}

func hasLevel(levels []Level, want Level) bool {
	return slices.Contains(levels, want)
}
