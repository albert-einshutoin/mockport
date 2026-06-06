package compat

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/albert-einshutoin/mockport/internal/adapter"
)

var endpointIDInvalidChars = regexp.MustCompile(`[^a-z0-9_]+`)

type Level = adapter.Level

const (
	LevelWire     = adapter.LevelWire
	LevelSDK      = adapter.LevelSDK
	LevelClient   = adapter.LevelClient
	LevelWorkflow = adapter.LevelWorkflow
	LevelError    = adapter.LevelError
	LevelState    = adapter.LevelState
	LevelContract = adapter.LevelContract
)

// ScenarioCategoryError is the scenario category that marks a manifest scenario
// as concrete error-behavior evidence for compatibility scoring.
const ScenarioCategoryError = adapter.ScenarioCategoryError

type SDKVersion struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type Manifest struct {
	Adapter          string                `json:"adapter"`
	ProviderVersion  string                `json:"provider_version"`
	Maturity         string                `json:"maturity,omitempty"`
	SDKVersions      []SDKVersion          `json:"sdk_versions,omitempty"`
	ClientEvidence   []string              `json:"client_evidence,omitempty"`
	Levels           []Level               `json:"levels,omitempty"`
	Endpoints        []Endpoint            `json:"endpoints,omitempty"`
	Scenarios        []Scenario            `json:"scenarios,omitempty"`
	StateEvidence    *StateEvidence        `json:"state_evidence,omitempty"`
	ContractEvidence *ContractEvidence     `json:"contract_evidence,omitempty"`
	Unsupported      []UnsupportedBehavior `json:"unsupported_behavior,omitempty"`
}

// StateEvidence captures the concrete fake-state surface an adapter exposes.
// It backs the state coverage score so a declared state level alone does not
// inflate the score without observable stateful behavior.
type StateEvidence struct {
	StatefulResources []string `json:"stateful_resources,omitempty"`
	Idempotency       bool     `json:"idempotency,omitempty"`
	Reset             bool     `json:"reset,omitempty"`
}

// HasEvidence reports whether any concrete fake-state behavior is present.
func (s *StateEvidence) HasEvidence() bool {
	if s == nil {
		return false
	}
	return len(s.StatefulResources) > 0 || s.Idempotency || s.Reset
}

// ContractEvidence captures the concrete release artifacts required for a
// contract-level provider-compatible claim. It must include fixture coverage,
// SDK/client contract evidence, and known-gap publication evidence.
type ContractEvidence struct {
	Fixtures     []string `json:"fixtures,omitempty"`
	SDKContracts []string `json:"sdk_contracts,omitempty"`
	KnownGaps    []string `json:"known_gaps,omitempty"`
}

// HasEvidence reports whether every contract-evidence dimension is present.
func (c *ContractEvidence) HasEvidence() bool {
	if c == nil {
		return false
	}
	return hasStringEvidence(c.Fixtures) &&
		hasStringEvidence(c.SDKContracts) &&
		hasStringEvidence(c.KnownGaps)
}

func (c *ContractEvidence) hasAnyEvidence() bool {
	if c == nil {
		return false
	}
	return len(c.Fixtures) > 0 || len(c.SDKContracts) > 0 || len(c.KnownGaps) > 0
}

type Endpoint struct {
	ID        string  `json:"id"`
	Method    string  `json:"method"`
	Path      string  `json:"path"`
	Supported bool    `json:"supported"`
	Levels    []Level `json:"levels,omitempty"`
}

type Scenario struct {
	Name      string  `json:"name"`
	BuiltIn   bool    `json:"built_in"`
	Supported bool    `json:"supported"`
	Category  string  `json:"category,omitempty"`
	Levels    []Level `json:"levels,omitempty"`
}

type UnsupportedBehavior struct {
	ID     string `json:"id"`
	Reason string `json:"reason"`
}

func (m Manifest) Validate() error {
	if strings.TrimSpace(m.Adapter) == "" {
		return fmt.Errorf("adapter is required")
	}
	if strings.TrimSpace(m.ProviderVersion) == "" {
		return fmt.Errorf("provider_version is required")
	}
	for _, level := range m.Levels {
		if !ValidLevel(level) {
			return fmt.Errorf("invalid compatibility level: %s", level)
		}
	}
	ids := map[string]bool{}
	for _, endpoint := range m.Endpoints {
		if strings.TrimSpace(endpoint.ID) == "" {
			return fmt.Errorf("endpoint id is required")
		}
		if ids[endpoint.ID] {
			return fmt.Errorf("duplicate endpoint id: %s", endpoint.ID)
		}
		ids[endpoint.ID] = true
		for _, level := range endpoint.Levels {
			if !ValidLevel(level) {
				return fmt.Errorf("invalid compatibility level: %s", level)
			}
		}
	}
	scenarios := map[string]bool{}
	for _, scenario := range m.Scenarios {
		if strings.TrimSpace(scenario.Name) == "" {
			return fmt.Errorf("scenario name is required")
		}
		if scenarios[scenario.Name] {
			return fmt.Errorf("duplicate scenario name: %s", scenario.Name)
		}
		scenarios[scenario.Name] = true
		for _, level := range scenario.Levels {
			if !ValidLevel(level) {
				return fmt.Errorf("invalid compatibility level: %s", level)
			}
		}
	}
	return nil
}

func ValidLevel(level Level) bool {
	return adapter.ValidateLevel(level)
}

func FromMetadata(meta adapter.Metadata) Manifest {
	providerVersion := meta.ProviderVersion
	if providerVersion == "" {
		providerVersion = "unspecified"
	}
	manifest := Manifest{
		Adapter:         meta.Name,
		ProviderVersion: providerVersion,
		Maturity:        string(meta.Maturity),
		Levels:          metadataLevels(meta.Levels),
	}
	for _, sdk := range meta.SDKVersions {
		manifest.SDKVersions = append(manifest.SDKVersions, SDKVersion{Name: sdk.Name, Version: sdk.Version})
	}
	manifest.ClientEvidence = append(manifest.ClientEvidence, meta.ClientEvidence...)
	// Adapter-wide levels live only on manifest.Levels. adapter.Metadata cannot
	// express per-endpoint/scenario evidence, so backfilling Levels here would let a
	// bare declaration masquerade as endpoint/scenario evidence and overstate the
	// state/error score (#21). Per-item levels come only from explicit manifests.
	for _, endpoint := range meta.Endpoints {
		manifest.Endpoints = append(manifest.Endpoints, Endpoint{
			ID:        endpointID(endpoint.Method, endpoint.Path),
			Method:    endpoint.Method,
			Path:      endpoint.Path,
			Supported: len(endpoint.SupportedScenarios) > 0 || endpoint.Method == http.MethodGet,
		})
	}
	for _, scenario := range meta.Scenarios {
		manifest.Scenarios = append(manifest.Scenarios, Scenario{
			Name:      scenario.Name,
			BuiltIn:   true,
			Supported: scenario.Supported,
			Category:  scenario.Category,
		})
	}
	if len(meta.StatefulResources) > 0 || meta.Idempotency || meta.Reset {
		manifest.StateEvidence = &StateEvidence{
			StatefulResources: append([]string(nil), meta.StatefulResources...),
			Idempotency:       meta.Idempotency,
			Reset:             meta.Reset,
		}
	}
	if evidence := contractEvidenceFromMetadata(meta.ContractEvidence); evidence.hasAnyEvidence() {
		manifest.ContractEvidence = evidence
	}
	return manifest
}

func contractEvidenceFromMetadata(evidence adapter.ContractEvidence) *ContractEvidence {
	return &ContractEvidence{
		Fixtures:     compactEvidence(evidence.Fixtures),
		SDKContracts: compactEvidence(evidence.SDKContracts),
		KnownGaps:    compactEvidence(evidence.KnownGaps),
	}
}

func compactEvidence(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			out = append(out, value)
		}
	}
	return out
}

func hasStringEvidence(values []string) bool {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return true
		}
	}
	return false
}

func metadataLevels(values []adapter.Level) []Level {
	if len(values) == 0 {
		return []Level{LevelWire}
	}
	levels := make([]Level, 0, len(values))
	for _, value := range values {
		level := Level(value)
		if ValidLevel(level) {
			levels = append(levels, level)
		}
	}
	if len(levels) == 0 {
		return []Level{LevelWire}
	}
	return levels
}

func endpointID(method, path string) string {
	value := strings.ToLower(method + "_" + strings.Trim(path, "/"))
	value = strings.ReplaceAll(value, "/", "_")
	value = strings.ReplaceAll(value, "-", "_")
	value = endpointIDInvalidChars.ReplaceAllString(value, "_")
	value = strings.Trim(value, "_")
	if value == "" {
		return "endpoint"
	}
	return value
}
