# Phase 2 CLI UX Implementation Plan

[日本語版](phase2_cli_ux.ja.md)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [x]`) syntax for tracking.

**Goal:** 空ディレクトリから `mockport init` と `mockport up/run` で 2 分以内に Stripe adapter を起動できる CLI UX にする。

**Architecture:** 既存の `internal/cli` を拡張し、生成ファイル保護、`--force`、`up`、`report` を追加する。Docker 実行は `docker compose -f docker-compose.mockport.yml up` の薄い wrapper に留め、複雑な Docker SDK は入れない。

**Tech Stack:** Go 1.26.3, Cobra, `os/exec`, `httptest`, temp directories.

---

## Files

- Modify: `internal/cli/init.go`
- Modify: `internal/cli/init_test.go`
- Create: `internal/cli/up.go`
- Create: `internal/cli/up_test.go`
- Create: `internal/cli/report.go`
- Create: `internal/cli/report_test.go`
- Modify: `internal/cli/root.go`
- Create: `scripts/smoke-empty-dir.sh`
- Modify: `README.md`
- Modify: `tasks/status.md`

## Task P2-T01: Protect Generated Files From Accidental Overwrite

**Status:** done

- [x] Write failing test: in `internal/cli/init_test.go`, create `mockport.yml` with custom content, run `mockport init --adapter stripe`, assert command returns error and file content is unchanged.
- [x] Run: `/usr/local/go/bin/go test ./internal/cli -run Init -v`; expected failure because overwrite protection is missing.
- [x] Implement: make `init` check `mockport.yml`, `.env.mockport`, `docker-compose.mockport.yml` before writing.
- [x] Run: `/usr/local/go/bin/go test ./internal/cli -run Init -v`; expected PASS.
- [x] Update `tasks/status.md`: P2-T01 `done`.

## Task P2-T02: Add `mockport init --force`

**Status:** done

- [x] Write failing test: existing generated files are overwritten when `mockport init --adapter stripe --force` is used.
- [x] Run targeted CLI test and confirm failure.
- [x] Implement `--force` bool flag in `internal/cli/init.go`.
- [x] Verify generated files contain deterministic Stripe config after force.
- [x] Run `/usr/local/go/bin/go test ./internal/cli -run Init -v`.

## Task P2-T03: Add `mockport up`

**Status:** done

- [x] Write failing test in `internal/cli/up_test.go`: inject command runner, run `mockport up`, assert command is `docker compose -f docker-compose.mockport.yml up`.
- [x] Verify RED: `up` command is missing.
- [x] Implement `internal/cli/up.go` with a small runner abstraction for tests.
- [x] Register `up` in `root.go`.
- [x] Run `/usr/local/go/bin/go test ./internal/cli -run Up -v`.

## Task P2-T04: Add `mockport report --url`

**Status:** done

- [x] Write failing test in `internal/cli/report_test.go`: `httptest.Server` returns report JSON, CLI prints `Mockport Report`, mode, adapter, request, safety counts.
- [x] Verify RED: `report` command is missing.
- [x] Implement `internal/cli/report.go` with default URL `http://localhost:43101/_mockport/report`.
- [x] Register `report` in `root.go`.
- [x] Run `/usr/local/go/bin/go test ./internal/cli -run Report -v`.

## Task P2-T05: Add User-facing Success Output

**Status:** done

- [x] Write failing tests asserting `mockport init` output includes exact next commands: `docker compose -f docker-compose.mockport.yml up` and `curl http://localhost:43101/health`.
- [x] Implement concise next-step output.
- [x] Run `/usr/local/go/bin/go test ./internal/cli -run Init -v`.
- [x] Update README quickstart to match output.

## Task P2-T06: Empty-directory E2E Smoke

**Status:** done

- [x] Create `scripts/smoke-empty-dir.sh` that builds `mockport:local`, runs binary `init --adapter stripe`, starts Docker compose, curls `/health`, curls checkout, curls report, then cleans up.
- [x] Add README section documenting the smoke command.
- [x] Run: `bash scripts/smoke-empty-dir.sh`.
- [x] Run full verification: `/usr/local/go/bin/go test ./...`, `/usr/local/go/bin/go vet ./...`, `docker build -t mockport:local -f docker/Dockerfile .`.

## Phase 2 Exit

- [x] Empty directory flow completes in under 2 minutes.
- [x] Existing files are protected unless `--force` is used.
- [x] `mockport up` delegates to Docker Compose.
- [x] `mockport report --url` renders report output.
- [x] README commands match implementation.
- [x] `tasks/status.md` Phase 2 is `done`.
