# Phase 20 GitHub OAuth Provider Compatibility Implementation Plan

[日本語版](phase20_github_oauth_provider_compatibility.ja.md)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** GitHub OAuth-like adapter を OAuth client/workflow-compatible local API へ引き上げる。

**Architecture:** GitHub OAuth の authorize/token/user profile workflow を fake identity/state 上で再現する。GitHub org/enterprise の完全権限モデルや内部 policy は再現しない。

**Tech Stack:** Go GitHub OAuth adapter, OAuth client contract tests, sanitized fixtures, compatibility manifests.

---

## Files

- Create: `compat/fixtures/github/*`
- Create: `contract/sdk/github-oauth-smoke.test.js`
- Modify: `adapters/githuboauth/*`
- Modify: `docs/site/support-matrix.md`
- Modify: `tasks/status.md`

## Task P20-T01: GitHub OAuth Client Contract Baseline

**Status:** done

- [x] Write failing client smoke for authorize redirect with state.
- [x] Write failing client smoke for token exchange.
- [x] Write failing client smoke for user profile and user emails subset.
- [x] Run `bash scripts/run-sdk-contracts.sh github-oauth`.

## Task P20-T02: GitHub OAuth State And Scope Fidelity

**Status:** done

- [x] Persist auth codes, access tokens, scopes, expiry, and fake user identities.
- [x] Add invalid code, expired token, missing scope, redirect URI mismatch fixtures.
- [x] Add user emails and org membership subset if required by common OAuth workflows.
- [x] Update support matrix and compatibility score.

## Task P20-T03: GitHub OAuth Error Fidelity

**Status:** done

- [x] Match public OAuth error response shapes for token endpoint.
- [x] Match common API auth error shapes for user endpoints.
- [x] Report unsupported scopes and endpoints.
- [x] Run adapter tests, client contracts, and compatibility report.

## Phase 20 Exit

- [x] GitHub OAuth adapter is at least `workflow-compatible`.
- [x] OAuth client contract passes for supported workflows.
- [x] Scope and token state behavior are documented.
- [x] Known gaps are explicit.
