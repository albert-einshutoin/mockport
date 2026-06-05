# Phase 8 Public Env Safety Implementation Plan

[日本語版](phase8_public_env_safety.ja.md)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 「Mockport 用 `.env` は公開してもよい」と言える条件を、初回 public preview 前にプロダクト契約として定義し検証する。

**Architecture:** `.env.mockport.example` を public-safe artifact として扱い、fake credential namespace、base URL policy、secret scanner、README/docs の注意書きを統一する。実 provider credential と混ざった場合は strict mode と CI check で検出する。

**Tech Stack:** Go security scanner, shell checks, examples, docs, CI.

---

## Files

- Create: `docs/public-env-safety.md`
- Create: `scripts/check-public-env.sh`
- Modify: `internal/security/secrets.go`
- Modify: `internal/security/secrets_test.go`
- Modify: `examples/*/.env.mockport.example`
- Modify: `README.md`
- Modify: `.github/workflows/ci.yml`
- Modify: `tasks/status.md`

## Task P8-T01: Public-safe Env Policy

**Status:** done

- [x] Define allowed fake credential prefixes: `mockport_`, `whsec_mockport`, local-only URLs.
- [x] Define forbidden patterns for real provider secrets, production provider URLs, and ambiguous placeholders.
- [x] Add docs explaining when `.env.mockport.example` is safe to commit and when it is not.
- [x] Run docs/static check.

## Task P8-T02: Env Scanner

**Status:** done

- [x] Write failing tests for real Stripe/OpenAI/Slack/GitHub secret patterns inside env files.
- [x] Implement scanner improvements for public env examples.
- [x] Add `scripts/check-public-env.sh` to scan examples and docs snippets.
- [x] Run `/usr/local/go/bin/go test ./internal/security -v` and `bash scripts/check-public-env.sh`.

## Task P8-T03: Public Env UX

**Status:** done

- [x] Update `mockport init` output to state generated `.env.mockport` uses fake local credentials.
- [x] Add `mockport report` safety field for public env contract status.
- [x] Add README section: "This Mockport env is safe to commit".
- [x] Run CLI tests and README command audit.

## Phase 8 Exit

- [x] Public-safe env policy is documented before public preview.
- [x] Env examples are checked in CI.
- [x] Scanner catches real-looking provider secrets.
- [x] Generated fake env values are compatible with current examples and reserved provider SDK env conventions.
