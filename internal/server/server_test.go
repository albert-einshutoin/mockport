package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/albert-einshutoin/mockport/adapters/stripe"
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

func immediateDelayTimer() delayTimerFunc {
	return func(time.Duration) <-chan time.Time {
		ch := make(chan time.Time, 1)
		ch <- time.Now()
		return ch
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
	var gotDelay time.Duration
	timer := func(d time.Duration) <-chan time.Time {
		gotDelay = d
		return immediateDelayTimer()(d)
	}
	handled := make(chan struct{}, 1)
	handler := delayMiddlewareWithTimer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(handled)
		w.WriteHeader(http.StatusOK)
	}), timer)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(delayHeader, "40")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if gotDelay != 40*time.Millisecond {
		t.Fatalf("delay = %s, want %s", gotDelay, 40*time.Millisecond)
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

func TestDelayMiddlewareBoundaryValues(t *testing.T) {
	for _, tc := range []struct {
		name        string
		delay       string
		missing     bool
		empty       bool
		timer       delayTimerFunc
		wantStatus  int
		wantBody    string
		wantHandled bool
		wantDelay   time.Duration
		maxElapsed  time.Duration
	}{
		{
			name:        "missing header",
			missing:     true,
			wantStatus:  http.StatusOK,
			wantHandled: true,
		},
		{
			name:        "zero",
			delay:       "0",
			wantStatus:  http.StatusOK,
			wantHandled: true,
			wantDelay:   0,
		},
		{
			name:        "max boundary 30000",
			delay:       "30000",
			wantStatus:  http.StatusOK,
			wantHandled: true,
			wantDelay:   30 * time.Second,
		},
		{
			name:        "over max 30001",
			delay:       "30001",
			wantStatus:  http.StatusBadRequest,
			wantBody:    invalidDelayMessage,
			wantHandled: false,
			maxElapsed:  50 * time.Millisecond,
		},
		{
			name:        "negative one",
			delay:       "-1",
			wantStatus:  http.StatusBadRequest,
			wantBody:    invalidDelayMessage,
			wantHandled: false,
			maxElapsed:  50 * time.Millisecond,
		},
		{
			name:        "non-integer not-a-number",
			delay:       "not-a-number",
			wantStatus:  http.StatusBadRequest,
			wantBody:    invalidDelayMessage,
			wantHandled: false,
			maxElapsed:  50 * time.Millisecond,
		},
		{
			name:        "empty value",
			empty:       true,
			wantStatus:  http.StatusBadRequest,
			wantBody:    invalidDelayMessage,
			wantHandled: false,
			maxElapsed:  50 * time.Millisecond,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var gotDelay time.Duration
			timer := tc.timer
			if timer == nil {
				timer = func(d time.Duration) <-chan time.Time {
					gotDelay = d
					return immediateDelayTimer()(d)
				}
			}
			handled := make(chan struct{}, 1)
			handler := delayMiddlewareWithTimer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				select {
				case handled <- struct{}{}:
				default:
				}
				w.WriteHeader(http.StatusOK)
			}), timer)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			switch {
			case tc.missing:
			case tc.empty:
				req.Header[http.CanonicalHeaderKey(delayHeader)] = []string{""}
			default:
				req.Header.Set(delayHeader, tc.delay)
			}

			start := time.Now()
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			elapsed := time.Since(start)

			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tc.wantStatus)
			}
			if tc.wantBody != "" && rec.Body.String() != tc.wantBody {
				t.Fatalf("body = %q, want %q", rec.Body.String(), tc.wantBody)
			}
			if tc.maxElapsed > 0 && elapsed > tc.maxElapsed {
				t.Fatalf("elapsed = %s, want <= %s for invalid delay", elapsed, tc.maxElapsed)
			}
			if tc.wantHandled && !tc.missing && gotDelay != tc.wantDelay {
				t.Fatalf("delay = %s, want %s", gotDelay, tc.wantDelay)
			}

			gotHandled := false
			select {
			case <-handled:
				gotHandled = true
			default:
			}
			if gotHandled != tc.wantHandled {
				t.Fatalf("handled = %v, want %v", gotHandled, tc.wantHandled)
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

func TestRequestBodySizeMiddleware(t *testing.T) {
	for _, tc := range []struct {
		name        string
		body        io.Reader
		contentLen  int64
		wantStatus  int
		wantBody    string
		wantPayload string
	}{
		{
			name:       "content length over limit",
			body:       http.NoBody,
			contentLen: maxRequestBodyBytes + 1,
			wantStatus: http.StatusRequestEntityTooLarge,
			wantBody:   requestBodyTooLargeMessage,
		},
		{
			name:       "unknown length oversized body",
			body:       bytes.NewReader(bytes.Repeat([]byte("x"), int(maxRequestBodyBytes+1))),
			contentLen: -1,
			wantStatus: http.StatusRequestEntityTooLarge,
			wantBody:   requestBodyTooLargeMessage,
		},
		{
			name:        "body within limit",
			body:        strings.NewReader(`{"ok":true}`),
			contentLen:  11,
			wantStatus:  http.StatusOK,
			wantPayload: `{"ok":true}`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var gotPayload string
			handler := requestBodySizeMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tc.wantPayload == "" {
					t.Fatal("inner handler should not run")
				}
				payload, err := io.ReadAll(r.Body)
				if err != nil {
					t.Fatalf("read body: %v", err)
				}
				gotPayload = string(payload)
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(http.MethodPost, "/", tc.body)
			req.ContentLength = tc.contentLen
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tc.wantStatus)
			}
			if tc.wantBody != "" && rec.Body.String() != tc.wantBody {
				t.Fatalf("body = %q, want %q", rec.Body.String(), tc.wantBody)
			}
			if tc.wantPayload != "" && gotPayload != tc.wantPayload {
				t.Fatalf("handler payload = %q, want %q", gotPayload, tc.wantPayload)
			}
		})
	}
}

func TestConfiguredHandlerRejectsOversizedRequestBody(t *testing.T) {
	cfg := config.Config{
		Mode:   "ai-safe",
		Server: config.ServerConfig{Host: "127.0.0.1", Port: 43101},
		Adapters: map[string]config.AdapterConfig{
			"stripe": {Enabled: true, BasePath: "/stripe", Scenario: "payment_success", FakeSecret: "mockport_stripe_secret"},
		},
	}
	if err := config.Validate(&cfg); err != nil {
		t.Fatalf("validate config: %v", err)
	}
	reg := adapter.NewRegistry()
	if err := reg.Register(stripe.New()); err != nil {
		t.Fatalf("register stripe: %v", err)
	}
	handler, err := NewConfiguredHandler(cfg, reg, report.NewRecorder())
	if err != nil {
		t.Fatalf("new handler: %v", err)
	}

	body := bytes.Repeat([]byte("x"), int(maxRequestBodyBytes+1))
	req := httptest.NewRequest(http.MethodPost, "/stripe/v1/customers", bytes.NewReader(body))
	req.ContentLength = -1

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusRequestEntityTooLarge, rec.Body.String())
	}
	if rec.Body.String() != requestBodyTooLargeMessage {
		t.Fatalf("body = %q, want %q", rec.Body.String(), requestBodyTooLargeMessage)
	}
	if strings.HasPrefix(strings.TrimSpace(rec.Body.String()), "{") {
		t.Fatalf("adapter handler ran, got JSON body %q", rec.Body.String())
	}
}
