# Phase 3 AI-safe Mode Implementation Plan

[日本語版](phase3_ai_safe_mode.ja.md)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [x]`) syntax for tracking.

**Goal:** AI-safe mode を Mockport の明確な差別化として、警告、strict fail、redaction、report、docs まで一貫させる。

**Architecture:** `internal/security` が検出・分類・redaction を担当し、`internal/config` が mode に応じて警告または error に変換する。`internal/report` は safe/unsafe の summary を持ち、CLI/server は full secret を表示しない。

**Tech Stack:** Go 1.26.3, table-driven tests, JSON report, Cobra output tests.

---

## Files

- Modify: `internal/security/secrets.go`
- Modify: `internal/security/secrets_test.go`
- Create: `internal/security/urls.go`
- Create: `internal/security/redactor.go`
- Modify: `internal/config/validate.go`
- Modify: `internal/config/config_test.go`
- Modify: `internal/report/report.go`
- Modify: `internal/server/report_test.go`
- Modify: `internal/cli/run.go`
- Modify: `internal/cli/run_test.go`
- Create: `docs/ai-safe-development.md`
- Create: `examples/unsafe-config/mockport.yml`

## Task P3-T01: Safety Status Model

**Status:** done

- [x] Write failing report test: unsafe config produces report fields `safety.safe=false`, `real_looking_secrets=1`, `external_urls=1`, `mode=ai-safe`.
- [x] Implement `report.SafetySummary`.
- [x] Wire config warnings into recorder summary.
- [x] Run `/usr/local/go/bin/go test ./internal/report ./internal/server -run Safety -v`.

## Task P3-T02: Startup Warnings In `ai-safe`

**Status:** done

- [x] Write failing CLI test: `mockport run --config unsafe.yml --check` prints warning names and categories but not secret values.
- [x] Add `--check` to `run` so tests avoid long-running server startup.
- [x] Implement warning output for `ai-safe`.
- [x] Run `/usr/local/go/bin/go test ./internal/cli -run Run -v`.

## Task P3-T03: Strict Mode Exit Behavior

**Status:** done

- [x] Write failing CLI test: strict unsafe config returns non-nil error containing `strict mode rejected unsafe config fields`.
- [x] Ensure `config.Validate` returns category-rich error without secret values.
- [x] Run `/usr/local/go/bin/go test ./internal/config ./internal/cli -run Strict -v`.

## Task P3-T04: Redaction Coverage

**Status:** done

- [x] Add table tests for short secrets, long secrets, fake secrets, env assignment strings, webhook secrets, provider URLs.
- [x] Implement `security.RedactValue` and `security.RedactMessage`.
- [x] Replace ad hoc redaction call sites with redactor functions.
- [x] Run `/usr/local/go/bin/go test ./internal/security -v`.

## Task P3-T05: External URL Guard

**Status:** done

- [x] Add config tests for Stripe/OpenAI/GitHub/LINE/Slack real URLs.
- [x] Move URL detection to `internal/security/urls.go`.
- [x] In `strict`, reject all known provider URLs.
- [x] In `ai-safe`, record warning categories.
- [x] Run `/usr/local/go/bin/go test ./internal/config ./internal/security -v`.

## Task P3-T06: AI-safe Docs And Unsafe Example

**Status:** done

- [x] Create `docs/ai-safe-development.md` with safe and unsafe config examples, strict mode behavior, and redaction examples.
- [x] Create `examples/unsafe-config/mockport.yml`.
- [x] Run documented `mockport run --config examples/unsafe-config/mockport.yml --check` and confirm warning output.
- [x] Run full verification: `/usr/local/go/bin/go test ./...`, `/usr/local/go/bin/go vet ./...`.

## Phase 3 Exit

- [x] `ai-safe` warns but starts.
- [x] `strict` fails before startup.
- [x] Report marks safe/unsafe state.
- [x] No full secret values appear in CLI output, logs, or report.
- [x] AI-safe docs match implemented behavior.
