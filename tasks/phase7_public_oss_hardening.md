# Phase 7 Public OSS Hardening Implementation Plan

[日本語版](phase7_public_oss_hardening.ja.md)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Mockport を公開前に最低限信頼できる OSS repository にする。

**Architecture:** 機能追加ではなく public trust artifacts、repository governance、初見 install 導線、security disclosure、CI enforcement を固める。公開前に README の主張と実行可能なコマンドを一致させる。

**Tech Stack:** Markdown, GitHub issue templates, GitHub Actions, shell checks, Go 1.26.3.

---

## Files

- Create: `LICENSE`
- Create: `SECURITY.md`
- Create: `CONTRIBUTING.md`
- Create: `CODE_OF_CONDUCT.md`
- Create: `.github/ISSUE_TEMPLATE/bug_report.yml`
- Create: `.github/ISSUE_TEMPLATE/feature_request.yml`
- Create: `.github/pull_request_template.md`
- Create: `docs/public-support-policy.md`
- Create: `scripts/check-public-trust.sh`
- Modify: `README.md`
- Modify: `.github/workflows/ci.yml`
- Modify: `tasks/status.md`

## Task P7-T01: Public Trust Artifacts

**Status:** done

- [x] Write failing `scripts/check-public-trust.sh` check for `LICENSE`, `SECURITY.md`, `CONTRIBUTING.md`, `CODE_OF_CONDUCT.md`, issue templates, and PR template.
- [x] Add MIT `LICENSE`.
- [x] Add `SECURITY.md` with supported versions, vulnerability report path, and AI-safe scope boundaries.
- [x] Add `CONTRIBUTING.md` with setup, TDD flow, verification commands, and PR expectations.
- [x] Add `CODE_OF_CONDUCT.md`.
- [x] Run `bash scripts/check-public-trust.sh`.

## Task P7-T02: GitHub Collaboration Surface

**Status:** done

- [x] Add bug report issue template with reproduction, adapter, config redaction, expected behavior, actual behavior.
- [x] Add feature request template with target adapter, scenario, API surface, and safety impact.
- [x] Add PR template with test evidence checklist.
- [x] Run `bash scripts/check-public-trust.sh`.

## Task P7-T03: README First-run Install Path

**Status:** done

- [x] Rewrite README quickstart so a new user can start from Docker without preinstalled `mockport`.
- [x] Add binary install placeholder for first release artifacts.
- [x] Add clear note that npm is experimental wrapper only.
- [x] Run README command audit: Docker build/run, `/health`, one adapter request, report.

## Task P7-T04: CI Public Gate

**Status:** done

- [x] Add public trust check to `.github/workflows/ci.yml`.
- [x] Add distribution static check to CI.
- [x] Keep Docker smoke optional or scheduled if it is too slow for every PR.
- [x] Run `/usr/local/go/bin/go test ./...`, `/usr/local/go/bin/go vet ./...`, `bash scripts/check-public-trust.sh`, and `bash scripts/check-distribution.sh`.

## Phase 7 Exit

- [x] Public trust files exist and are checked by CI.
- [x] README first-run path does not require unpublished install assumptions.
- [x] Security and contribution policies are explicit.
- [x] Issue/PR templates guide useful external contributions.
- [x] CI gates public trust and distribution static checks.
