# Phase 4 Trust Reports And Adapter Contracts Implementation Plan

[日本語版](phase4_trust_reports.ja.md)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [x]`) syntax for tracking.

**Goal:** adapter を増やす前に、Mockport が「何をサポートし、何をサポートしないか」を report で説明できる trust foundation を作る。

**Architecture:** Adapter metadata と request recorder を統合し、scenario coverage、unsupported endpoint、request replay metadata、behavior matrix、maturity levels を JSON と text CLI の両方で出す。Phase 5 の追加 adapter はこの contract に必ず乗せる。

**Tech Stack:** Go 1.26.3, JSON report schema, text renderer tests, adapter metadata contract.

---

## Files

- Modify: `internal/adapter/adapter.go`
- Modify: `internal/adapter/registry.go`
- Modify: `internal/adapter/registry_test.go`
- Modify: `adapters/stripe/adapter.go`
- Modify: `adapters/stripe/adapter_test.go`
- Modify: `internal/report/report.go`
- Modify: `internal/report/recorder.go`
- Modify: `internal/report/recorder_test.go`
- Modify: `internal/server/server.go`
- Modify: `internal/server/report_test.go`
- Create: `internal/report/render.go`
- Create: `internal/report/render_test.go`
- Modify: `internal/cli/report.go`
- Modify: `internal/cli/report_test.go`
- Create: `docs/reporting.md`

## Task P4-T01: Adapter Metadata Contract

**Status:** done

- [x] Write failing adapter tests asserting each adapter exposes name, maturity, capabilities, scenarios, and supported endpoints.
- [x] Extend adapter types with metadata without introducing dynamic plugins.
- [x] Add Stripe metadata: maturity `partial`, scenarios `payment_success`, `payment_failed`, `auth_error`, `rate_limited`, `timeout`.
- [x] Add endpoint metadata for Stripe checkout session, payment intent, and webhook send.
- [x] Run `/usr/local/go/bin/go test ./internal/adapter ./adapters/stripe -v`.

## Task P4-T02: Scenario Coverage Report

**Status:** done

- [x] Write failing report tests asserting Stripe reports scenario names and supported status.
- [x] Add report schema field `scenario_coverage`.
- [x] Generate scenario coverage from adapter metadata.
- [x] Run `/usr/local/go/bin/go test ./internal/report ./internal/server -run Scenario -v`.

## Task P4-T03: Unsupported Endpoint Report

**Status:** done

- [x] Write failing HTTP test: request unknown adapter path, then report includes method/path/status and reason `unsupported_endpoint`.
- [x] Update middleware to record 404 and 405 responses with classification.
- [x] Avoid recording `/_mockport/report` itself.
- [x] Run `/usr/local/go/bin/go test ./internal/server -run Unsupported -v`.

## Task P4-T04: Request Replay Log Metadata

**Status:** done

- [x] Write failing recorder tests for stable request id, timestamp, method, path, status, adapter, scenario.
- [x] Implement monotonic request ids and replay-safe metadata.
- [x] Ensure no request body or secret header is stored by default.
- [x] Run `/usr/local/go/bin/go test ./internal/report -run Replay -v`.

## Task P4-T05: Behavior Matrix And Maturity Levels

**Status:** done

- [x] Write failing report tests for `behavior_matrix` with endpoint, method, supported scenarios, notes, and maturity.
- [x] Validate allowed maturity values: `experimental`, `partial`, `sdk-compatible`, `workflow-compatible`, `provider-compatible`.
- [x] Include maturity in report and README adapter table.
- [x] Run `/usr/local/go/bin/go test ./internal/adapter ./internal/report ./internal/server -run 'Behavior|Maturity' -v`.

## Task P4-T06: JSON And Text Report Modes

**Status:** done

- [x] Write failing CLI tests for `mockport report --format json` and `--format text`.
- [x] Implement `internal/report/render.go`.
- [x] Text output must include adapters, requests, safety, coverage, unsupported endpoints, and maturity.
- [x] Create `docs/reporting.md` with interpretation examples.
- [x] Run `/usr/local/go/bin/go test ./internal/cli ./internal/report -run Report -v`.

## Phase 4 Exit

- [x] Report explains supported and unsupported behavior.
- [x] Report includes scenario coverage and adapter maturity.
- [x] Unsupported endpoints are visible.
- [x] Text and JSON report outputs are tested.
- [x] Docs explain how to interpret report output.
- [x] Phase 5 adapters can be added without inventing a new report contract.
