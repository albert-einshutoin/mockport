package adapter

import (
	"fmt"
	"net/http"
	"strings"
)

// Config is the runtime configuration passed to a built-in adapter.
type Config struct {
	BasePath             string
	Scenario             string
	FakeSecret           string
	WebhookTargetURL     string
	WebhookSigningSecret string
}

// Adapter describes the minimal contract every built-in provider mock implements.
type Adapter interface {
	Name() string
	Register(mux *http.ServeMux, cfg Config) error
	FakeEnv(cfg Config) map[string]string
	Metadata() Metadata
}

// Metadata describes an adapter's compatibility, scenarios, and supported endpoint surface.
type Metadata struct {
	Name              string
	Maturity          Maturity
	ProviderVersion   string
	SDKVersions       []SDKVersion
	ClientEvidence    []string
	Levels            []Level
	Capabilities      []string
	Scenarios         []Scenario
	Endpoints         []Endpoint
	StatefulResources []string
	Idempotency       bool
	Reset             bool
}

// SDKVersion records a client SDK version used as compatibility evidence.
type SDKVersion struct {
	Name    string
	Version string
}

// Scenario describes a built-in deterministic behavior mode for an adapter.
type Scenario struct {
	Name      string
	Supported bool
	// Category optionally classifies the scenario. The "error" category marks a
	// scenario as concrete error-behavior evidence for compatibility scoring.
	Category string
}

// ScenarioCategoryError marks a scenario as concrete error-behavior evidence.
const ScenarioCategoryError = "error"

// Endpoint describes one provider-like HTTP endpoint exposed by an adapter.
type Endpoint struct {
	Method             string
	Path               string
	SupportedScenarios []string
	Notes              string
}

// Maturity is the public compatibility maturity claim for an adapter.
type Maturity string

const (
	MaturityExperimental       Maturity = "experimental"
	MaturityPartial            Maturity = "partial"
	MaturitySDKCompatible      Maturity = "sdk-compatible"
	MaturityWorkflowCompatible Maturity = "workflow-compatible"
	MaturityProviderCompatible Maturity = "provider-compatible"
)

// Level identifies the compatibility evidence dimension covered by an adapter.
type Level string

const (
	LevelWire     Level = "wire"
	LevelSDK      Level = "sdk"
	LevelClient   Level = "client"
	LevelWorkflow Level = "workflow"
	LevelState    Level = "state"
	LevelError    Level = "error"
	LevelContract Level = "contract"
)

// ValidateMaturity reports whether maturity is one of Mockport's known maturity values.
func ValidateMaturity(maturity Maturity) bool {
	switch maturity {
	case MaturityExperimental, MaturityPartial, MaturitySDKCompatible, MaturityWorkflowCompatible, MaturityProviderCompatible:
		return true
	default:
		return false
	}
}

// ValidateLevel reports whether level is one of Mockport's known compatibility dimensions.
func ValidateLevel(level Level) bool {
	switch level {
	case LevelWire, LevelSDK, LevelClient, LevelWorkflow, LevelState, LevelError, LevelContract:
		return true
	default:
		return false
	}
}

// ValidateMetadata validates the adapter contract before it is exposed through reports.
func ValidateMetadata(meta Metadata) error {
	if strings.TrimSpace(meta.Name) == "" {
		return fmt.Errorf("adapter name is required")
	}
	if !ValidateMaturity(meta.Maturity) {
		return fmt.Errorf("invalid maturity: %s", meta.Maturity)
	}
	if strings.TrimSpace(meta.ProviderVersion) == "" {
		return fmt.Errorf("provider version is required")
	}
	for _, level := range meta.Levels {
		if !ValidateLevel(level) {
			return fmt.Errorf("invalid compatibility level: %s", level)
		}
	}
	scenarios := map[string]bool{}
	for _, scenario := range meta.Scenarios {
		if strings.TrimSpace(scenario.Name) == "" {
			return fmt.Errorf("scenario name is required")
		}
		if scenarios[scenario.Name] {
			return fmt.Errorf("duplicate scenario: %s", scenario.Name)
		}
		scenarios[scenario.Name] = true
	}
	endpoints := map[string]bool{}
	for _, endpoint := range meta.Endpoints {
		key := endpoint.Method + " " + endpoint.Path
		if strings.TrimSpace(endpoint.Method) == "" || strings.TrimSpace(endpoint.Path) == "" {
			return fmt.Errorf("endpoint method and path are required")
		}
		if endpoints[key] {
			return fmt.Errorf("duplicate endpoint: %s", key)
		}
		endpoints[key] = true
	}
	return nil
}
