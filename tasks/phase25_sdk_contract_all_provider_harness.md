# Phase 25 SDK Contract All-provider Harness Implementation Plan

[日本語版](phase25_sdk_contract_all_provider_harness.ja.md)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** `bash scripts/run-sdk-contracts.sh all` を placeholder ではなく、Stripe / OpenAI / GitHub OAuth / Slack の実 provider-specific contracts をまとめて実行する信頼できるCI入口にする。

**Architecture:** 既存の provider-specific smoke を保持しつつ、`all` は同じ Mockport server 起動に対して全 provider contract を順番に実行する。各 provider の結果を JSON array として出し、失敗 provider が分かるようにする。

**Tech Stack:** Node.js 24 contract runner, Go Mockport server, shell harness.

---

## Files

- Modify: `contract/sdk/test-runner.js`
- Modify: `contract/sdk/smoke-placeholder.test.js`
- Modify: `scripts/run-sdk-contracts.sh`
- Modify: `.github/workflows/ci.yml`
- Modify: `.github/workflows/compatibility.yml`
- Modify: `docs/compatibility-reports/README.md`
- Modify: `tasks/status.md`

## Task P25-T01: RED Test For `all`

**Status:** pending

- [ ] Write a failing assertion in `contract/sdk/test-runner.js` or a new `contract/sdk/all-smoke.test.js` requiring `--provider all` live mode to run all four providers.
- [ ] Run `bash scripts/run-sdk-contracts.sh all`.
- [ ] Confirm it fails because the current output is only `{"provider":"all","status":"ok"}` and does not include `stripe`, `openai`, `github-oauth`, and `slack`.
- [ ] Capture the failure text in the Phase 25 task notes.

## Task P25-T02: Implement All-provider Runner

**Status:** pending

- [ ] Refactor `contract/sdk/test-runner.js` so provider dispatch is a function like `runProvider(options, provider)`.
- [ ] Implement `runAll(options)` to execute `stripe`, `openai`, `github-oauth`, and `slack` sequentially against the same `baseURL`.
- [ ] Return JSON shaped as `{ "provider": "all", "status": "sdk-ok", "results": [...] }`.
- [ ] Ensure any failed provider exits non-zero and includes the provider name in the error message.

## Task P25-T03: CI Integration

**Status:** pending

- [ ] Keep provider-specific lines in compatibility workflow for readable failure boundaries.
- [ ] Update `ci.yml` so `bash scripts/run-sdk-contracts.sh all` is the real all-provider gate.
- [ ] Update `scripts/check-maintenance-policy.sh` to require the all-provider gate in CI.
- [ ] Update compatibility docs to explain when to run `all` versus a single provider.

## Task P25-T04: Verification

**Status:** pending

- [ ] Run `(cd contract/sdk && npm test)`.
- [ ] Run `bash scripts/run-sdk-contracts.sh all`.
- [ ] Run every single-provider contract: `stripe`, `openai`, `github-oauth`, `slack`.
- [ ] Run `go test ./...`, `go vet ./...`, `bash scripts/check-maintenance-policy.sh`, and `bash scripts/check-compatibility-release.sh`.

## Phase 25 Exit

- [ ] `run-sdk-contracts.sh all` executes all supported provider contracts, not a placeholder.
- [ ] JSON output contains individual provider results.
- [ ] CI uses the real all-provider gate.
- [ ] Provider-specific runs still work for debugging.
