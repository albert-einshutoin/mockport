package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/albert-einshutoin/mockport/internal/adapter"
	"github.com/albert-einshutoin/mockport/internal/config"
	"github.com/albert-einshutoin/mockport/internal/report"
)

func TestHealthReturnsOK(t *testing.T) {
	handler, err := NewConfiguredHandler(config.Config{}, adapter.NewRegistry(), report.NewRecorder())
	if err != nil {
		t.Fatalf("configure handler: %v", err)
	}
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode health body: %v", err)
	}
	if body["status"] != "ok" {
		t.Fatalf("status body = %q, want ok", body["status"])
	}
}

func TestHealthRejectsNonGETMethod(t *testing.T) {
	handler, err := NewConfiguredHandler(config.Config{}, adapter.NewRegistry(), report.NewRecorder())
	if err != nil {
		t.Fatalf("configure handler: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
	if got := strings.TrimSpace(rec.Body.String()); got != "Method Not Allowed" {
		t.Fatalf("body = %q, want Method Not Allowed", got)
	}
	// Go 1.22 ServeMux auto-includes HEAD for GET routes in the Allow header.
	if allow := rec.Header().Get("Allow"); !strings.Contains(allow, http.MethodGet) {
		t.Fatalf("Allow = %q, want contains %q", allow, http.MethodGet)
	}
}

func TestReportRejectsNonGETMethod(t *testing.T) {
	handler, err := NewConfiguredHandler(config.Config{}, adapter.NewRegistry(), report.NewRecorder())
	if err != nil {
		t.Fatalf("configure handler: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/_mockport/report", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
	if got := strings.TrimSpace(rec.Body.String()); got != "Method Not Allowed" {
		t.Fatalf("body = %q, want Method Not Allowed", got)
	}
	if allow := rec.Header().Get("Allow"); !strings.Contains(allow, http.MethodGet) {
		t.Fatalf("Allow = %q, want contains %q", allow, http.MethodGet)
	}
}

func TestNewConfiguredHandlerUnregisteredAdapterReturnsErrAdapterNotRegistered(t *testing.T) {
	cfg := config.Config{
		Adapters: map[string]config.AdapterConfig{
			"missing": {Enabled: true, BasePath: "/missing"},
		},
	}

	_, err := NewConfiguredHandler(cfg, adapter.NewRegistry(), report.NewRecorder())
	if err == nil {
		t.Fatal("NewConfiguredHandler returned nil error for unregistered adapter")
	}
	if !errors.Is(err, ErrAdapterNotRegistered) {
		t.Fatalf("error = %v, want ErrAdapterNotRegistered", err)
	}
	errText := err.Error()
	if !strings.Contains(errText, "missing") {
		t.Fatalf("error = %q, want adapter name in message", errText)
	}
}

func TestRecordMiddlewareAllowsResponseControllerFlush(t *testing.T) {
	handler := recordMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := http.NewResponseController(w).Flush(); err != nil {
			http.Error(w, "flush unsupported: "+err.Error(), http.StatusInternalServerError)
			return
		}
		_, _ = w.Write([]byte("flushed"))
	}), report.NewRecorder(), nil)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/stream", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if rec.Body.String() != "flushed" {
		t.Fatalf("body = %q, want flushed", rec.Body.String())
	}
}

func TestRecordMiddlewareRecordsStreamingRequestAfterFlush(t *testing.T) {
	recorder := report.NewRecorder()
	handler := recordMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("data: ok\n\n"))
		_ = http.NewResponseController(w).Flush()
	}), recorder, []report.AdapterStatus{{Name: "openai", BasePath: "/openai", Scenario: "stream_success"}})

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/openai/v1/chat/completions", nil))

	snapshot := recorder.Snapshot()
	if len(snapshot.Requests) != 1 {
		t.Fatalf("requests = %#v, want one", snapshot.Requests)
	}
	got := snapshot.Requests[0]
	if got.Status != http.StatusOK || got.Adapter != "openai" || got.Scenario != "stream_success" {
		t.Fatalf("request = %#v", got)
	}
}

func TestDelayMiddlewareNoHeader(t *testing.T) {
	handled := false
	handler := delayMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handled = true
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !handled {
		t.Fatalf("expected handler to run without delay header")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestDelayMiddlewareAppliesDelayHeader(t *testing.T) {
	handled := make(chan struct{}, 1)
	handler := delayMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(handled)
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(delayHeader, "40")

	start := time.Now()
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	elapsed := time.Since(start)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if elapsed < 35*time.Millisecond {
		t.Fatalf("elapsed = %s, want at least 35ms", elapsed)
	}
	if elapsed > 200*time.Millisecond {
		t.Fatalf("elapsed = %s, want not exceed 200ms", elapsed)
	}
	select {
	case <-handled:
	default:
		t.Fatal("expected handler to be called")
	}
}

func TestRecordMiddlewareSkipsCancelledDelayRequestFromReport(t *testing.T) {
	recorder := report.NewRecorder()
	handler := recordMiddleware(
		delayMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})),
		recorder,
		nil,
	)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	req := httptest.NewRequest(http.MethodGet, "/timeout", nil).WithContext(ctx)
	req.Header.Set(delayHeader, "120")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	snapshot := recorder.Snapshot()
	if len(snapshot.Requests) != 0 {
		t.Fatalf("requests = %#v, want empty", snapshot.Requests)
	}
}

func TestDelayMiddlewareRejectsBadDelay(t *testing.T) {
	for _, tc := range []struct {
		name    string
		delay   string
		wantMsg string
	}{
		{
			name:  "non-integer",
			delay: "abc",
		},
		{
			name:  "negative",
			delay: "-1",
		},
		{
			name:  "too-large",
			delay: "60000",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			handled := false
			handler := delayMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handled = true
			}))

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set(delayHeader, tc.delay)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
			}
			if handled {
				t.Fatalf("expected handler not to run for delay=%q", tc.delay)
			}
		})
	}
}

func TestDelayMiddlewareHonorsCancelledContext(t *testing.T) {
	handled := make(chan struct{}, 1)
	handler := delayMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handled <- struct{}{}
		w.WriteHeader(http.StatusOK)
	}))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
	req.Header.Set(delayHeader, "2000")

	rec := httptest.NewRecorder()
	start := time.Now()
	handler.ServeHTTP(rec, req)
	elapsed := time.Since(start)

	if elapsed > 100*time.Millisecond {
		t.Fatalf("elapsed = %s, want cancel-aware quick return", elapsed)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	select {
	case <-handled:
		t.Fatalf("expected handler not to run on canceled context")
	default:
	}
}
