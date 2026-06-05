# Phase 19 OpenAI Provider Compatibility Implementation Plan

[日本語版](phase19_openai_provider_compatibility.ja.md)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** OpenAI-compatible adapter を OpenAI SDK/workflow-compatible local API へ引き上げる。

**Architecture:** OpenAI の主要 developer workflow を deterministic fake inference として再現する。実 model quality、tokenization 完全一致、provider 内部 scheduling は再現しない。

**Tech Stack:** Go OpenAI adapter, OpenAI SDK contract tests, streaming fixtures, compatibility manifests.

---

## Files

- Create: `compat/fixtures/openai/*`
- Create: `contract/sdk/openai-smoke.test.js`
- Modify: `adapters/openai/*`
- Modify: `docs/site/support-matrix.md`
- Modify: `tasks/status.md`

## Task P19-T01: OpenAI SDK Contract Baseline

**Status:** done

- [x] Write failing SDK smoke for models list.
- [x] Write failing SDK smoke for chat completions and responses.
- [x] Add streaming smoke if SDK supports base URL override cleanly.
- [x] Run `bash scripts/run-sdk-contracts.sh openai`.

## Task P19-T02: OpenAI Major Surface Expansion

**Status:** done

- [x] Add embeddings, files, batches, and responses tool-call subset coverage backlog.
- [x] Implement endpoint groups with TDD and SDK contracts.
- [x] Keep inference fake, deterministic, and documented.
- [x] Update support matrix and compatibility score.

## Task P19-T03: OpenAI Error And Streaming Fidelity

**Status:** done

- [x] Add fixtures for auth errors, rate limits, context length, invalid model, malformed messages/input.
- [x] Verify Phase 13 SSE-compatible `stream_success` against SDK contracts and streaming fixtures.
- [x] Add unsupported parameter reporting.
- [x] Run adapter tests, SDK contracts, and compatibility report.

## Phase 19 Exit

- [x] OpenAI adapter is at least `workflow-compatible`.
- [x] OpenAI SDK contracts pass for supported workflows.
- [x] Streaming and error behavior are documented by fixtures.
- [x] Known gaps are explicit.
