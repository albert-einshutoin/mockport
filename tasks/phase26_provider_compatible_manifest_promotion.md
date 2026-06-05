# Phase 26 Provider-compatible Manifest Promotion Implementation Plan

[日本語版](phase26_provider_compatible_manifest_promotion.ja.md)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** `provider-compatible` 昇格に必要な manifest / contract evidence / fixture coverage / known gaps を adapter ごとに明示し、maturity promotion が手作業の主観にならないようにする。

**Architecture:** Runtime metadata からの簡易scoreに加えて、versioned compatibility manifest を `compat/manifests/` に置く。release check は manifest、SDK/client contract、fixture evidence、known gaps を検査し、`provider-compatible` は contract level evidence がある adapter だけ許可する。

**Tech Stack:** JSON manifests, Node.js validation script, existing `internal/compat` model, shell release checks.

---

## Files

- Create: `compat/manifests/README.md`
- Create: `compat/manifests/schema.example.json`
- Create: `compat/manifests/stripe.json`
- Create: `compat/manifests/openai.json`
- Create: `compat/manifests/github-oauth.json`
- Create: `compat/manifests/slack.json`
- Create: `compat/manifests/line.json`
- Create: `scripts/check-compat-manifests.mjs`
- Modify: `scripts/check-compatibility-release.sh`
- Modify: `scripts/render-compatibility-report.mjs`
- Modify: `docs/compatibility-model.md`
- Modify: `tasks/status.md`

## Task P26-T01: Manifest Schema RED

**Status:** pending

- [ ] Add `scripts/check-compat-manifests.mjs` with validation for required fields: `adapter`, `provider_version`, `maturity`, `levels`, `workflows`, `evidence`, `known_gaps`.
- [ ] Run `node scripts/check-compat-manifests.mjs`.
- [ ] Confirm it fails because `compat/manifests/*.json` do not exist yet.
- [ ] Keep the checker strict: empty `known_gaps` is invalid, and `provider-compatible` requires `contract` in `levels`.

## Task P26-T02: Create Initial Workflow Manifests

**Status:** pending

- [ ] Create `stripe.json` with current workflow-compatible evidence and selected provider-compatible candidate workflows.
- [ ] Create `openai.json` with current workflow-compatible evidence and explicit known gaps for model quality, tokenization, tools, and vector stores.
- [ ] Create `github-oauth.json` with OAuth/client contract evidence and known gaps for repository/org policy.
- [ ] Create `slack.json` with Web API/client contract evidence and known gaps for delivery, files, Block Kit, and enterprise policy.
- [ ] Create `line.json` with current workflow-compatible evidence and explicit known gaps for LINE SDK contract coverage, LIFF runtime, provider webhook redelivery, quota/rate buckets, schema validation, regional policy, and Mini Dapp helper boundaries.

## Task P26-T03: Report Integration

**Status:** pending

- [ ] Update `scripts/render-compatibility-report.mjs` to merge runtime report data with `compat/manifests/*.json`.
- [ ] Add manifest evidence fields to `docs/compatibility-reports/latest.json`.
- [ ] Add a "Manifest Evidence" section to `docs/compatibility-reports/latest.md`.
- [ ] Ensure generated reports still include runtime score and known gaps.

## Task P26-T04: Promotion Gate

**Status:** pending

- [ ] Update `scripts/check-compatibility-release.sh` to call `node scripts/check-compat-manifests.mjs`.
- [ ] Enforce that no adapter can be marked `provider-compatible` unless manifest levels include `contract`, score is at least 80, and evidence includes SDK/client contract plus fixtures.
- [ ] Add docs to `docs/compatibility-model.md` explaining manifest-based promotion.
- [ ] Run full compatibility checks.

## Phase 26 Exit

- [ ] Versioned manifests exist for every current registered adapter.
- [ ] Release checks fail on invalid maturity promotion.
- [ ] Compatibility report includes manifest evidence.
- [ ] `provider-compatible` has a concrete, automated gate.
