# Phase 1 Stripe Minimal MVP Implementation Plan

[日本語版](phase1_stripe_minimal_mvp.ja.md)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [x]`) syntax for tracking.

**Goal:** Stripe-like adapter を built-in adapter として追加し、Minimal MVP の success/failure/webhook/report/AI-safe 検証を Docker-first で満たす。

**Architecture:** Adapter registry が config で有効化された built-in adapter を server に登録する。Stripe adapter は `/stripe` base path 配下で checkout session、payment intent、webhook send を扱い、request recorder と security report は server middleware から更新する。

**Tech Stack:** Go 1.26.3, `net/http`, `httptest`, `encoding/json`, `crypto/hmac`, `crypto/sha256`, Docker.

---

## Files

- Create: `internal/adapter/adapter.go`
- Create: `internal/adapter/registry.go`
- Create: `internal/adapter/registry_test.go`
- Create: `internal/report/recorder.go`
- Create: `internal/report/report.go`
- Create: `internal/report/recorder_test.go`
- Modify: `internal/server/server.go`
- Create: `internal/server/report_test.go`
- Create: `adapters/stripe/adapter.go`
- Create: `adapters/stripe/routes.go`
- Create: `adapters/stripe/models.go`
- Create: `adapters/stripe/scenarios.go`
- Create: `adapters/stripe/webhook.go`
- Create: `adapters/stripe/signatures.go`
- Create: `adapters/stripe/adapter_test.go`
- Modify: `internal/config/config.go`
- Modify: `internal/config/validate.go`
- Modify: `internal/cli/root.go`
- Create: `internal/cli/init.go`
- Create: `internal/cli/init_test.go`
- Create: `configs/mockport.example.yml`
- Create: `examples/stripe-checkout/README.md`
- Create: `examples/stripe-checkout/mockport.yml`
- Create: `examples/stripe-checkout/.env.mockport.example`
- Create: `examples/stripe-checkout/docker-compose.yml`

## Task P1-T01: Adapter Registry

**Status:** done

- [x] **Step 1: Write failing registry test**

Create `internal/adapter/registry_test.go`:

```go
package adapter

import (
	"net/http"
	"testing"
)

type fakeAdapter struct{ name string }

func (a fakeAdapter) Name() string { return a.name }
func (a fakeAdapter) Register(mux *http.ServeMux, cfg Config) error { return nil }
func (a fakeAdapter) FakeEnv(cfg Config) map[string]string { return map[string]string{"FAKE_URL": "http://localhost"} }

func TestRegistryReturnsRegisteredAdapter(t *testing.T) {
	reg := NewRegistry()
	reg.Register(fakeAdapter{name: "stripe"})

	got, ok := reg.Get("stripe")
	if !ok {
		t.Fatal("expected registered adapter")
	}
	if got.Name() != "stripe" {
		t.Fatalf("adapter name = %q, want stripe", got.Name())
	}
}
```

- [x] **Step 2: Verify RED**

Run:

```bash
go test ./internal/adapter -v
```

Expected: compile failure because registry does not exist.

- [x] **Step 3: Implement registry**

Create `internal/adapter/adapter.go` and `registry.go` with the documented minimal interface:

```go
type Adapter interface {
	Name() string
	Register(mux *http.ServeMux, cfg Config) error
	FakeEnv(cfg Config) map[string]string
}
```

`Config` should include `BasePath`, `Scenario`, `FakeSecret`, `WebhookTargetURL`, and `WebhookSigningSecret`.

- [x] **Step 4: Verify GREEN**

Run:

```bash
go test ./internal/adapter -v
```

Expected: PASS.

## Task P1-T02: Request Recorder And Report Model

**Status:** done

- [x] **Step 1: Write failing report test**

Create `internal/report/recorder_test.go`:

```go
package report

import (
	"net/http"
	"testing"
)

func TestRecorderStoresRequestsAndSafety(t *testing.T) {
	rec := NewRecorder()
	rec.RecordRequest(http.MethodPost, "/stripe/v1/checkout/sessions", 200)
	rec.RecordSafetyWarning("STRIPE_SECRET_KEY", "real-looking Stripe key")

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
}
```

- [x] **Step 2: Verify RED**

Run:

```bash
go test ./internal/report -v
```

Expected: compile failure because recorder does not exist.

- [x] **Step 3: Implement recorder**

Create `internal/report/report.go` and `recorder.go`. Use a mutex-protected in-memory recorder with `RecordRequest`, `RecordSafetyWarning`, and `Snapshot`. Do not store request bodies in Minimal MVP.

- [x] **Step 4: Verify GREEN**

Run:

```bash
go test ./internal/report -v
```

Expected: PASS.

## Task P1-T03: Stripe Checkout Session Scenarios

**Status:** done

- [x] **Step 1: Write failing Stripe checkout tests**

Create `adapters/stripe/adapter_test.go` with tests for:

```txt
POST /stripe/v1/checkout/sessions with scenario payment_success -> 200, object checkout.session, payment_status paid
POST /stripe/v1/checkout/sessions with scenario payment_failed -> 402, error code card_declined
```

Use `httptest.NewRecorder` and register the adapter on a fresh `http.ServeMux`.

- [x] **Step 2: Verify RED**

Run:

```bash
go test ./adapters/stripe -run Checkout -v
```

Expected: compile failure because adapter does not exist.

- [x] **Step 3: Implement checkout routes**

Create `adapters/stripe/adapter.go`, `routes.go`, `models.go`, and `scenarios.go`. Implement `Name() string` returning `stripe`, `Register`, and JSON response builders for `payment_success` and `payment_failed`.

- [x] **Step 4: Verify GREEN**

Run:

```bash
go test ./adapters/stripe -run Checkout -v
```

Expected: PASS.

## Task P1-T04: Stripe Payment Intent And Error Scenarios

**Status:** done

- [x] **Step 1: Write failing payment intent tests**

Add tests for:

```txt
POST /stripe/v1/payment_intents with payment_success -> 200, object payment_intent, status succeeded
POST /stripe/v1/payment_intents with payment_failed -> 402
POST /stripe/v1/payment_intents with auth_error -> 401
POST /stripe/v1/payment_intents with rate_limited -> 429
GET /stripe/v1/payment_intents/pi_mockport -> 200
```

- [x] **Step 2: Verify RED**

Run:

```bash
go test ./adapters/stripe -run PaymentIntent -v
```

Expected: fail because payment intent routes are missing.

- [x] **Step 3: Implement payment intent routes**

Extend Stripe routes to handle `POST /v1/payment_intents` and `GET /v1/payment_intents/{id}`. Use deterministic IDs such as `pi_mockport_success` and `pi_mockport_failed`.

- [x] **Step 4: Verify GREEN**

Run:

```bash
go test ./adapters/stripe -run PaymentIntent -v
```

Expected: PASS.

## Task P1-T05: Timeout Scenario

**Status:** done

- [x] **Step 1: Write failing timeout test**

Add a test that configures scenario `timeout`, performs `POST /stripe/v1/checkout/sessions`, and asserts the handler returns HTTP 504 with an error code `mockport_timeout`. Keep the test fast; do not sleep for multiple seconds.

- [x] **Step 2: Verify RED**

Run:

```bash
go test ./adapters/stripe -run Timeout -v
```

Expected: fail because timeout scenario is not implemented.

- [x] **Step 3: Implement timeout response**

Map scenario `timeout` to status 504 with JSON body:

```json
{"error":{"type":"api_error","code":"mockport_timeout","message":"Mockport simulated timeout"}}
```

- [x] **Step 4: Verify GREEN**

Run:

```bash
go test ./adapters/stripe -run Timeout -v
```

Expected: PASS.

## Task P1-T06: Webhook Sender Endpoint

**Status:** done

- [x] **Step 1: Write failing webhook test**

Add a test with an `httptest.Server` target URL. Register Stripe adapter with `WebhookTargetURL` and `WebhookSigningSecret`, then POST `/stripe/test/webhook/send`. Assert target receives one request with `Stripe-Signature` header and event type `checkout.session.completed`.

- [x] **Step 2: Verify RED**

Run:

```bash
go test ./adapters/stripe -run Webhook -v
```

Expected: fail because webhook endpoint is missing.

- [x] **Step 3: Implement webhook sender**

Create `webhook.go` and `signatures.go`. Generate HMAC SHA-256 signature with fake signing secret and send JSON event to configured target. Return 202 on successful send and 400 if target URL is missing.

- [x] **Step 4: Verify GREEN**

Run:

```bash
go test ./adapters/stripe -run Webhook -v
```

Expected: PASS.

## Task P1-T07: AI-safe Validation And Strict Mode

**Status:** done

- [x] **Step 1: Write failing config/security tests**

Add tests that:

```txt
mode ai-safe + sk_live_ value loads with a safety warning
mode strict + sk_live_ value fails validation
mode strict + https://api.stripe.com fails validation
mode ai-safe + mockport_stripe_secret passes without warning
```

- [x] **Step 2: Verify RED**

Run:

```bash
go test ./internal/config ./internal/security -run 'Strict|AISafe|URL' -v
```

Expected: fail because config validation does not inspect adapter secrets or URLs.

- [x] **Step 3: Implement validation**

Add URL detection for `https://api.stripe.com`, `https://api.openai.com`, `https://api.github.com`, `https://api.line.me`, and `https://slack.com/api`. In `strict`, return an error. In `ai-safe`, return config plus warnings for the report layer.

- [x] **Step 4: Verify GREEN**

Run:

```bash
go test ./internal/config ./internal/security -v
```

Expected: PASS.

## Task P1-T08: `/_mockport/report` Endpoint

**Status:** done

- [x] **Step 1: Write failing server report test**

Create `internal/server/report_test.go`. Use a recorder, serve one Stripe request, then GET `/_mockport/report`. Assert JSON contains mode, enabled adapters, request path, response status, and safety warning count.

- [x] **Step 2: Verify RED**

Run:

```bash
go test ./internal/server -run Report -v
```

Expected: fail because report endpoint is missing.

- [x] **Step 3: Implement report route and middleware**

Update server construction to accept config, registry, and recorder. Wrap handlers so response status is recorded. Register `GET /_mockport/report`.

- [x] **Step 4: Verify GREEN**

Run:

```bash
go test ./internal/server ./internal/report -v
```

Expected: PASS.

## Task P1-T09: `mockport init`

**Status:** done

- [x] **Step 1: Write failing init tests**

Create `internal/cli/init_test.go`. In `t.TempDir()`, run `mockport init --adapter stripe` and assert these files exist:

```txt
mockport.yml
.env.mockport
docker-compose.mockport.yml
```

Assert `.env.mockport` contains `STRIPE_API_URL=http://localhost:43101/stripe` and `STRIPE_SECRET_KEY=mockport_stripe_secret`.

- [x] **Step 2: Verify RED**

Run:

```bash
go test ./internal/cli -run Init -v
```

Expected: fail because `init` is not registered.

- [x] **Step 3: Implement init command**

Create `internal/cli/init.go`. Support only `--adapter stripe` in Minimal MVP. Write deterministic files matching `configs/mockport.example.yml`, `docs/archive/design/configs/env.mockport.example`, and `docs/archive/design/examples/docker-compose.mockport.yml`.

- [x] **Step 4: Verify GREEN**

Run:

```bash
go test ./internal/cli -run Init -v
```

Expected: PASS.

## Task P1-T10: Minimal MVP Documentation And Verification

**Status:** done

- [x] **Step 1: Add examples**

Create `examples/stripe-checkout/` with `README.md`, `mockport.yml`, `.env.mockport.example`, and `docker-compose.yml`.

- [x] **Step 2: Verify commands**

Run:

```bash
go test ./...
go vet ./...
go build ./cmd/mockport
docker build -t mockport:local -f docker/Dockerfile .
```

Then start Docker and verify:

```bash
docker run --rm -d --name mockport-test -p 43101:43101 -v "$(pwd)/configs/mockport.example.yml:/etc/mockport/mockport.yml" mockport:local
curl -fsS http://localhost:43101/health
curl -fsS -X POST http://localhost:43101/stripe/v1/checkout/sessions
curl -fsS http://localhost:43101/_mockport/report
docker rm -f mockport-test
```

Expected: all commands pass; health returns 200; checkout returns Stripe-like JSON; report includes the checkout request.

## Phase 1 Exit

- [x] `go test ./...` passes.
- [x] `go vet ./...` passes.
- [x] `go build ./cmd/mockport` passes.
- [x] Docker image builds.
- [x] Docker container serves `/health`.
- [x] Stripe checkout success returns 200.
- [x] Stripe failure returns 402.
- [x] Webhook sender posts to target URL.
- [x] `/_mockport/report` shows requests and safety warnings.
- [x] README quickstart is accurate.
- [x] `tasks/status.md` Phase 1 summary is updated to `done`.
