package compat

import (
	"net/http"
	"strings"
	"testing"

	"github.com/albert-einshutoin/mockport/internal/adapter"
)

func TestManifestValidation(t *testing.T) {
	manifest := Manifest{
		Adapter:         "stripe",
		ProviderVersion: "2026-05-26",
		SDKVersions:     []SDKVersion{{Name: "stripe-go", Version: "v83.0.0"}},
		Levels:          []Level{LevelWire, LevelSDK, LevelError},
		Endpoints: []Endpoint{
			{ID: "checkout_sessions_create", Method: http.MethodPost, Path: "/v1/checkout/sessions", Supported: true, Levels: []Level{LevelWire}},
		},
		Scenarios: []Scenario{
			{Name: "payment_success", BuiltIn: true, Supported: true, Levels: []Level{LevelWire}},
		},
	}

	if err := manifest.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestManifestValidationRejectsDuplicateEndpointIDs(t *testing.T) {
	manifest := Manifest{
		Adapter:         "stripe",
		ProviderVersion: "2026-05-26",
		Endpoints: []Endpoint{
			{ID: "duplicate", Method: http.MethodGet, Path: "/v1/a"},
			{ID: "duplicate", Method: http.MethodGet, Path: "/v1/b"},
		},
	}

	err := manifest.Validate()
	if err == nil || !strings.Contains(err.Error(), "duplicate endpoint id") {
		t.Fatalf("Validate() error = %v, want duplicate endpoint id", err)
	}
}

func TestManifestValidationRejectsInvalidLevels(t *testing.T) {
	manifest := Manifest{
		Adapter:         "stripe",
		ProviderVersion: "2026-05-26",
		Levels:          []Level{"complete"},
	}

	err := manifest.Validate()
	if err == nil || !strings.Contains(err.Error(), "invalid compatibility level") {
		t.Fatalf("Validate() error = %v, want invalid compatibility level", err)
	}
}

func TestManifestFromAdapterMetadata(t *testing.T) {
	meta := adapter.Metadata{
		Name:            "openai",
		Maturity:        adapter.MaturityExperimental,
		ProviderVersion: "2025-10-29.clover",
		SDKVersions:     []adapter.SDKVersion{{Name: "openai", Version: "6.0.0"}},
		Levels:          []adapter.Level{adapter.LevelWire, adapter.LevelSDK, adapter.LevelState},
		Capabilities:    []string{"chat_completions"},
		Scenarios:       []adapter.Scenario{{Name: "chat_success", Supported: true}},
		Endpoints:       []adapter.Endpoint{{Method: http.MethodPost, Path: "/openai/v1/chat/completions", SupportedScenarios: []string{"chat_success"}}},
	}

	manifest := FromMetadata(meta)
	if manifest.Adapter != "openai" {
		t.Fatalf("adapter = %q, want openai", manifest.Adapter)
	}
	if manifest.ProviderVersion != "2025-10-29.clover" {
		t.Fatalf("provider version = %q", manifest.ProviderVersion)
	}
	if len(manifest.SDKVersions) != 1 || manifest.SDKVersions[0].Name != "openai" {
		t.Fatalf("sdk versions = %#v", manifest.SDKVersions)
	}
	if len(manifest.Levels) != 3 {
		t.Fatalf("levels = %#v", manifest.Levels)
	}
	if len(manifest.Endpoints) != 1 || manifest.Endpoints[0].ID == "" {
		t.Fatalf("endpoints = %#v, want generated id", manifest.Endpoints)
	}
	if len(manifest.Scenarios) != 1 || !manifest.Scenarios[0].BuiltIn {
		t.Fatalf("scenarios = %#v, want built-in scenario", manifest.Scenarios)
	}
}

func TestManifestFromAdapterMetadataPreservesClientLevel(t *testing.T) {
	meta := adapter.Metadata{
		Name:            "slack",
		Maturity:        adapter.MaturityWorkflowCompatible,
		ProviderVersion: "2025-02-01",
		ClientEvidence:  []string{"slack-client-contract"},
		Levels:          []adapter.Level{adapter.LevelWire, adapter.LevelClient, adapter.LevelWorkflow},
		Scenarios:       []adapter.Scenario{{Name: "message_success", Supported: true}},
		Endpoints:       []adapter.Endpoint{{Method: http.MethodPost, Path: "/slack/api/chat.postMessage", SupportedScenarios: []string{"message_success"}}},
	}

	manifest := FromMetadata(meta)
	if len(manifest.Levels) != 3 {
		t.Fatalf("levels = %#v, want client level preserved", manifest.Levels)
	}
	if manifest.Levels[1] != LevelClient {
		t.Fatalf("levels = %#v, want client at index 1", manifest.Levels)
	}
	if len(manifest.ClientEvidence) != 1 || manifest.ClientEvidence[0] != "slack-client-contract" {
		t.Fatalf("client evidence = %#v", manifest.ClientEvidence)
	}
	if err := manifest.Validate(); err != nil {
		t.Fatalf("Validate() with client level error = %v", err)
	}
}
