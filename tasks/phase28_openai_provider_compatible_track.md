# Phase 28 OpenAI Provider-compatible Track Implementation Plan

[日本語版](phase28_openai_provider_compatible_track.ja.md)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** OpenAI-compatible adapter を選定workflow単位で provider-compatible 候補に近づけ、real model inference をしない前提でも API / SDK / streaming / state の互換証跡を強化する。

**Architecture:** Chat Completions、Responses、SSE streaming、Embeddings、Files、Batches を対象にする。model quality、tokenization parity、hosted tools、vector stores、provider scheduling は再現しない known gaps として残す。

**Tech Stack:** Go OpenAI adapter, OpenAI SDK `6.39.0`, SSE contract tests, compatibility manifests.

---

## Files

- Modify: `adapters/openai/*`
- Modify: `contract/sdk/openai-smoke.test.js`
- Modify: `compat/manifests/openai.json`
- Modify: `compat/fixtures/openai/*`
- Modify: `docs/site/support-matrix.md`
- Modify: `docs/compatibility-reports/latest.md`
- Modify: `tasks/status.md`

## Task P28-T01: OpenAI Scope Definition

**Status:** pending

- [ ] Define selected workflows in `compat/manifests/openai.json`: chat, responses, streaming, embeddings, files, batches.
- [ ] Exclude model quality, tokenization parity, hosted tools, vector stores, and provider scheduling.
- [ ] Link every selected workflow to at least one fixture and one contract test.
- [ ] Run `node scripts/check-compat-manifests.mjs`.

## Task P28-T02: SDK Streaming And Error Contracts

**Status:** pending

- [ ] Add failing SDK tests for SSE chunk shape, terminal `[DONE]`, and content accumulation.
- [ ] Add failing SDK tests for malformed messages/input, unsupported parameters, invalid model, context length, auth, and rate limit.
- [ ] Add failing SDK tests for responses retrieve consistency and batch retrieve consistency.
- [ ] Run `bash scripts/run-sdk-contracts.sh openai` and confirm RED.

## Task P28-T03: Adapter Fidelity Improvements

**Status:** pending

- [ ] Implement only the minimal response shape fixes needed by SDK tests.
- [ ] Ensure deterministic fake IDs and state survive retrieve/list workflows.
- [ ] Ensure streaming remains SSE-compatible and does not return plain JSON when `stream=true`.
- [ ] Keep fake inference deterministic and documented.

## Task P28-T04: Promotion Decision

**Status:** pending

- [ ] Regenerate compatibility reports.
- [ ] If OpenAI satisfies contract-level gate for selected workflows, promote to `provider-compatible`; otherwise record blockers.
- [ ] Update known gaps to distinguish "not implemented" from "intentionally not reproducible".
- [ ] Run full provider contract checks.

## Phase 28 Exit

- [ ] OpenAI selected workflows have contract evidence and known gaps.
- [ ] Streaming and error behavior are SDK-tested.
- [ ] OpenAI maturity decision is automated by manifest gate.
- [ ] `bash scripts/run-sdk-contracts.sh openai` and `all` pass.
