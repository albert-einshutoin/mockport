# Phase 14 Compatibility Engine Implementation Plan

[日本語版](phase14_compatibility_engine.ja.md)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Mockport の互換性を「感覚」ではなく、adapter ごとの contract、coverage、score として測れる基盤にする。

**Architecture:** Phase 12 の fixture/spec/scenario policy と既存の Phase 4 metadata/report contract を拡張し、wire/schema/sdk/workflow/state/error の互換レベルを adapter ごとに表現する。Full-compatible は provider 内部再現ではなく、公開 API 契約、SDK、主要 workflow、fake state、主要 error を local Docker API として高忠実度に再現することとして定義する。

**Tech Stack:** Go 1.26.3, JSON/YAML compatibility manifests, report model, CLI report output, shell/static checks.

---

## Files

- Create: `internal/compat/manifest.go`
- Create: `internal/compat/manifest_test.go`
- Create: `internal/compat/score.go`
- Create: `internal/compat/score_test.go`
- Create: `docs/compatibility-model.md`
- Modify: `internal/adapter/adapter.go`
- Modify: `internal/report/report.go`
- Modify: `internal/report/render.go`
- Modify: `tasks/status.md`

## Task P14-T01: Compatibility Level Model

**Status:** done

- [x] Write failing tests for levels: `wire`, `sdk`, `workflow`, `error`, `state`, `contract`.
- [x] Implement compatibility manifest structs with adapter, provider version, SDK versions, endpoints, scenarios, unsupported behavior.
- [x] Consume the Phase 12 built-in/user-defined scenario policy when scoring scenario coverage.
- [x] Add validation for required fields and duplicate endpoint ids.
- [x] Run `go test ./internal/compat -v`.

## Task P14-T02: Compatibility Score

**Status:** done

- [x] Write failing score tests for endpoint coverage, scenario coverage, SDK smoke status, state support, and error support.
- [x] Implement deterministic score calculation without network access.
- [x] Expose score in report JSON.
- [x] Run `go test ./internal/compat ./internal/report -v`.

## Task P14-T03: Compatibility Report Output

**Status:** done

- [x] Add report fields for compatibility level, score, SDK versions tested, provider API version, unsupported endpoints.
- [x] Add text report rendering.
- [x] Update docs to define what Mockport means by provider-compatible local API.
- [x] Run `go test ./...`.

## Task P14-T04: Provisional Promotion Rule

**Status:** done

- [x] Define provisional maturity labels before provider-specific tracks start.
- [x] Require fixture policy, report visibility, and minimum contract evidence before raising maturity.
- [x] Add tests or static checks that prevent undocumented compatibility claims.
- [x] Run docs/static checks.

## Phase 14 Exit

- [x] Compatibility is represented as a structured manifest.
- [x] Report shows compatibility score and levels.
- [x] Full-compatible is explicitly defined as external API/SDK/workflow compatibility, not provider internals.
- [x] Provisional promotion rules exist before provider-specific compatibility work starts.
- [x] Unsupported and approximate behavior remains visible.
