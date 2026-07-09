package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/albert-einshutoin/mockport/internal/adapter"
	"github.com/albert-einshutoin/mockport/internal/compat"
	"github.com/albert-einshutoin/mockport/internal/config"
	"github.com/albert-einshutoin/mockport/internal/report"
)

const (
	delayHeader         = "X-Mockport-Delay"
	maxDelay            = 30 * time.Second
	invalidDelayMessage = "invalid X-Mockport-Delay: must be 0-30000 (milliseconds)"
	// maxRequestBodyBytes matches httpx.MaxBodyBytes so server-wide and adapter limits stay aligned.
	maxRequestBodyBytes        int64 = 1 << 20
	requestBodyTooLargeMessage       = "request body too large"
)

// delayTimerFunc abstracts time.After so tests can inject immediate timers and avoid real sleeps in CI.
type delayTimerFunc func(time.Duration) <-chan time.Time

// ErrAdapterNotRegistered marks enabled adapters missing from the registry.
var ErrAdapterNotRegistered = errors.New("adapter is enabled but not registered")

func NewConfiguredHandler(cfg config.Config, reg *adapter.Registry, rec *report.Recorder) (http.Handler, error) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)
	if rec == nil {
		rec = report.NewRecorder()
	}
	rec.SetMode(cfg.Mode)

	var adapterStatuses []report.AdapterStatus
	var coverage []report.ScenarioCoverage
	var matrix []report.BehaviorMatrixEntry
	var compatibility []report.CompatibilityStatus
	var state []report.StateCoverageStatus
	for _, warning := range cfg.SafetyWarnings {
		rec.RecordSafetyWarning(warning.Field, warning.Category, warning.Message)
	}

	adapterNames := slices.Sorted(maps.Keys(cfg.Adapters))

	for _, name := range adapterNames {
		adapterCfg := cfg.Adapters[name]
		if !adapterCfg.Enabled {
			continue
		}
		registered, ok := reg.Get(name)
		if !ok {
			return nil, fmt.Errorf("adapter %s: %w", name, ErrAdapterNotRegistered)
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
			Capabilities: slices.Clone(meta.Capabilities),
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
	mux.HandleFunc("GET /_mockport/report", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(rec.Snapshot())
	})

	return recordMiddleware(requestBodySizeMiddleware(delayMiddleware(mux)), rec, adapterStatuses), nil
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
		evidence := manifest.ContractEvidence.Clone()
		status.ContractEvidence = &evidence
	}
	for _, unsupported := range manifest.Unsupported {
		status.UnsupportedEndpoints = append(status.UnsupportedEndpoints, unsupported.ID)
	}
	return status
}

type statusRecorder struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (r *statusRecorder) WriteHeader(status int) {
	r.wroteHeader = true
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Write(data []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}
	return r.ResponseWriter.Write(data)
}

func (r *statusRecorder) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}

func recordMiddleware(next http.Handler, rec *report.Recorder, adapters []report.AdapterStatus) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sr := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sr, r)
		if r.URL.Path != "/_mockport/report" {
			if !sr.wroteHeader {
				return
			}
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
			SupportedScenarios: slices.Clone(endpoint.SupportedScenarios),
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
		StatefulResources: slices.Clone(meta.StatefulResources),
		Idempotency:       meta.Idempotency,
		Reset:             meta.Reset,
	}, true
}

func classifyAdapter(path string, adapters []report.AdapterStatus) (string, string) {
	for _, adapter := range adapters {
		if adapter.BasePath != "" && (path == adapter.BasePath || strings.HasPrefix(path, adapter.BasePath+"/")) {
			return adapter.Name, adapter.Scenario
		}
	}
	return "", ""
}

func requestBodySizeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Body == nil {
			next.ServeHTTP(w, r)
			return
		}

		// ContentLength is honored first: a declared size above the limit must be rejected
		// even when the underlying body would be empty, so clients cannot bypass the cap.
		if r.ContentLength > maxRequestBodyBytes {
			rejectRequestBodyTooLarge(w)
			return
		}

		// Chunked or otherwise unknown-length bodies must be bounded before adapter handlers run.
		// http.NoBody and other empty readers return EOF immediately, so they pass through safely.
		if r.ContentLength < 0 {
			body, err := io.ReadAll(io.LimitReader(r.Body, maxRequestBodyBytes+1))
			_ = r.Body.Close()
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if int64(len(body)) > maxRequestBodyBytes {
				rejectRequestBodyTooLarge(w)
				return
			}
			r.Body = io.NopCloser(bytes.NewReader(body))
			r.ContentLength = int64(len(body))
		}

		next.ServeHTTP(w, r)
	})
}

func rejectRequestBodyTooLarge(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusRequestEntityTooLarge)
	_, _ = w.Write([]byte(requestBodyTooLargeMessage))
}

func delayMiddleware(next http.Handler) http.Handler {
	return delayMiddlewareWithTimer(next, time.After)
}

func delayMiddlewareWithTimer(next http.Handler, timer delayTimerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Header.Get conflates missing and empty; distinguish so explicit empty stays invalid.
		values, present := r.Header[http.CanonicalHeaderKey(delayHeader)]
		if !present {
			next.ServeHTTP(w, r)
			return
		}
		rawDelay := strings.TrimSpace(values[0])
		if rawDelay == "" {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(invalidDelayMessage))
			return
		}

		delayMs, err := strconv.ParseInt(rawDelay, 10, 64)
		if err != nil || delayMs < 0 || delayMs > int64(maxDelay/time.Millisecond) {
			// Match docs/site/adapters.md exactly; http.Error appends a newline.
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(invalidDelayMessage))
			return
		}

		delay := time.Duration(delayMs) * time.Millisecond
		select {
		case <-timer(delay):
			next.ServeHTTP(w, r)
		case <-r.Context().Done():
			return
		}
	})
}
