package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"

	"github.com/albert-einshutoin/mockport/internal/adapter"
	"github.com/albert-einshutoin/mockport/internal/compat"
	"github.com/albert-einshutoin/mockport/internal/config"
	"github.com/albert-einshutoin/mockport/internal/report"
)

func NewHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	return mux
}

func NewConfiguredHandler(cfg config.Config, reg *adapter.Registry, rec *report.Recorder) (http.Handler, error) {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	if rec == nil {
		rec = report.NewRecorder()
	}
	rec.SetMode(cfg.Mode)

	var adapterStatuses []report.AdapterStatus
	var coverage []report.ScenarioCoverage
	var matrix []report.BehaviorMatrixEntry
	var compatibility []report.CompatibilityStatus
	var state []report.StateCoverageStatus
	for warningIdx := range cfg.SafetyWarnings {
		warning := cfg.SafetyWarnings[warningIdx]
		rec.RecordSafetyWarning(warning.Field, warning.Category, warning.Message)
	}

	adapterNames := make([]string, 0, len(cfg.Adapters))
	for name := range cfg.Adapters {
		adapterNames = append(adapterNames, name)
	}
	sort.Strings(adapterNames)

	for _, name := range adapterNames {
		adapterCfg := cfg.Adapters[name]
		if !adapterCfg.Enabled {
			continue
		}
		registered, ok := reg.Get(name)
		if !ok {
			return nil, fmt.Errorf("adapter %s is enabled but not registered", name)
		}
		meta := registered.Metadata()
		if err := adapter.ValidateMetadata(meta); err != nil {
			return nil, fmt.Errorf("adapter %s metadata: %w", name, err)
		}
		adapterStatuses = append(adapterStatuses, report.AdapterStatus{
			Name:         name,
			BasePath:     adapterCfg.BasePath,
			Enabled:      true,
			Scenario:     adapterCfg.Scenario,
			Maturity:     string(meta.Maturity),
			Capabilities: append([]string(nil), meta.Capabilities...),
		})
		coverage = append(coverage, scenarioCoverage(meta))
		matrix = append(matrix, behaviorMatrix(meta)...)
		compatibility = append(compatibility, compatibilityStatus(compat.FromMetadata(meta)))
		if stateStatus, ok := stateCoverage(meta); ok {
			state = append(state, stateStatus)
		}
		if err := registered.Register(mux, adapter.Config{
			BasePath:             adapterCfg.BasePath,
			Scenario:             adapterCfg.Scenario,
			FakeSecret:           adapterCfg.FakeSecret,
			WebhookTargetURL:     adapterCfg.Webhook.TargetURL,
			WebhookSigningSecret: adapterCfg.Webhook.SigningSecret,
		}); err != nil {
			return nil, err
		}
	}
	rec.SetAdapters(adapterStatuses)
	rec.SetScenarioCoverage(coverage)
	rec.SetBehaviorMatrix(matrix)
	rec.SetCompatibility(compatibility)
	rec.SetStateCoverage(state)
	mux.HandleFunc("/_mockport/report", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(rec.Snapshot())
	})

	return recordMiddleware(mux, rec, adapterStatuses), nil
}

func compatibilityStatus(manifest compat.Manifest) report.CompatibilityStatus {
	score := compat.CalculateScore(manifest)
	status := report.CompatibilityStatus{
		Adapter:          manifest.Adapter,
		Level:            score.Level,
		Score:            score.Total,
		EndpointCoverage: score.EndpointCoverage,
		ScenarioCoverage: score.ScenarioCoverage,
		SDKCoverage:      score.SDKCoverage,
		StateCoverage:    score.StateCoverage,
		ErrorCoverage:    score.ErrorCoverage,
		// Single source of truth for the published report: compute on the Go side
		// whether the declared maturity satisfies CanPromote. The validator trusts
		// this value and does not re-implement the scoring logic.
		PromotionEligible: compat.CanPromote(manifest, score, manifest.Maturity),
		ProviderVersion:   manifest.ProviderVersion,
	}
	for _, sdk := range manifest.SDKVersions {
		status.SDKVersions = append(status.SDKVersions, sdk.Name+"@"+sdk.Version)
	}
	status.ClientEvidence = append(status.ClientEvidence, manifest.ClientEvidence...)
	if manifest.ContractEvidence != nil {
		status.ContractEvidence = &report.ContractEvidence{
			Fixtures:     append([]string(nil), manifest.ContractEvidence.Fixtures...),
			SDKContracts: append([]string(nil), manifest.ContractEvidence.SDKContracts...),
			KnownGaps:    append([]string(nil), manifest.ContractEvidence.KnownGaps...),
		}
	}
	for _, unsupported := range manifest.Unsupported {
		status.UnsupportedEndpoints = append(status.UnsupportedEndpoints, unsupported.ID)
	}
	return status
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}

func recordMiddleware(next http.Handler, rec *report.Recorder, adapters []report.AdapterStatus) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sr := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sr, r)
		if r.URL.Path != "/_mockport/report" {
			adapterName, scenario := classifyAdapter(r.URL.Path, adapters)
			reason := ""
			if sr.status == http.StatusNotFound || sr.status == http.StatusMethodNotAllowed {
				reason = "unsupported_endpoint"
			}
			rec.RecordRequestWithDetails(r.Method, r.URL.Path, sr.status, adapterName, scenario, reason)
		}
	})
}

func scenarioCoverage(meta adapter.Metadata) report.ScenarioCoverage {
	coverage := report.ScenarioCoverage{Adapter: meta.Name}
	for _, scenario := range meta.Scenarios {
		coverage.Scenarios = append(coverage.Scenarios, report.ScenarioSupport{Name: scenario.Name, Supported: scenario.Supported})
	}
	return coverage
}

func behaviorMatrix(meta adapter.Metadata) []report.BehaviorMatrixEntry {
	matrix := make([]report.BehaviorMatrixEntry, 0, len(meta.Endpoints))
	for _, endpoint := range meta.Endpoints {
		matrix = append(matrix, report.BehaviorMatrixEntry{
			Adapter:            meta.Name,
			Maturity:           string(meta.Maturity),
			Method:             endpoint.Method,
			Path:               endpoint.Path,
			SupportedScenarios: append([]string(nil), endpoint.SupportedScenarios...),
			Notes:              endpoint.Notes,
		})
	}
	return matrix
}

func stateCoverage(meta adapter.Metadata) (report.StateCoverageStatus, bool) {
	if len(meta.StatefulResources) == 0 && !meta.Idempotency && !meta.Reset {
		return report.StateCoverageStatus{}, false
	}
	return report.StateCoverageStatus{
		Adapter:           meta.Name,
		StatefulResources: append([]string(nil), meta.StatefulResources...),
		Idempotency:       meta.Idempotency,
		Reset:             meta.Reset,
	}, true
}

func classifyAdapter(path string, adapters []report.AdapterStatus) (string, string) {
	for _, adapter := range adapters {
		if adapter.BasePath != "" && (path == adapter.BasePath || len(path) > len(adapter.BasePath) && path[:len(adapter.BasePath)+1] == adapter.BasePath+"/") {
			return adapter.Name, adapter.Scenario
		}
	}
	return "", ""
}
