# Phase 15 SDK Contract Harness Foundation Implementation Plan

[日本語版](phase15_sdk_contract_harness_foundation.ja.md)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 公式 SDK または実運用に近い client を使った contract harness の土台を作る。

**Architecture:** この Phase では SDK contract runner、dependency pinning、Mockport 起動管理、result reporting までに限定する。Provider ごとの深い SDK compatibility は Phase18-21 の provider-specific track で実装する。

**Tech Stack:** Go server, Node SDK smoke workspace, shell scripts, Docker Compose, fixture snapshots.

---

## Files

- Create: `contract/sdk/README.md`
- Create: `contract/sdk/package.json`
- Create: `contract/sdk/package-lock.json`
- Create: `contract/sdk/test-runner.js`
- Create: `contract/sdk/smoke-placeholder.test.js`
- Create: `scripts/run-sdk-contracts.sh`
- Modify: `.github/workflows/ci.yml`
- Modify: `docs/compatibility-model.md`
- Modify: `tasks/status.md`

## Task P15-T01: SDK Contract Workspace

**Status:** done

- [x] Add a dedicated SDK contract package outside runtime code.
- [x] Pin package versions with `package-lock.json`.
- [x] Add placeholder smoke that proves the runner can reach local Mockport health endpoint.
- [x] Run `(cd contract/sdk && npm test)`.

## Task P15-T02: Mockport Contract Runner

**Status:** done

- [x] Add `scripts/run-sdk-contracts.sh` that builds Mockport, starts it with multi-adapter config, runs selected SDK tests, and cleans up.
- [x] Support selecting provider names: `stripe`, `openai`, `github-oauth`, `slack`, or `all`.
- [x] Emit JSON or text result summary for compatibility report ingestion.
- [x] Run `bash scripts/run-sdk-contracts.sh all`.

## Task P15-T03: CI Integration

**Status:** done

- [x] Add SDK contract foundation check to CI without requiring external provider network calls.
- [x] Keep SDK dependencies test-only and out of the Go runtime.
- [x] Document how provider-specific tracks add tests later.
- [x] Run full local verification.

## Phase 15 Exit

- [x] SDK contract workspace exists with pinned dependencies.
- [x] Contract runner starts local Mockport and executes tests.
- [x] No external provider network calls are required.
- [x] Provider-specific SDK contracts can be added incrementally.
