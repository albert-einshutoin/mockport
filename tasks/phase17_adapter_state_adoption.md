# Phase 17 Adapter State Adoption Implementation Plan

[日本語版](phase17_adapter_state_adoption.ja.md)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Phase 16 の state foundation を主要 adapter に適用し、provider-specific SDK compatibility tracks の前提を作る。

**Architecture:** Stripe/OpenAI/GitHub OAuth/Slack の既存 scenario behavior を壊さず、create/retrieve/list/update などの公開 workflow に見える範囲で fake state を導入する。state adoption は provider internals の再現ではなく、SDK/client が期待する resource lifecycle を deterministic に成立させるための準備に限定する。

**Tech Stack:** Go adapters, internal state package, httptest, report metadata.

---

## Files

- Modify: `adapters/stripe/*`
- Modify: `adapters/openai/*`
- Modify: `adapters/githuboauth/*`
- Modify: `adapters/slack/*`
- Modify: `docs/site/support-matrix.md`
- Modify: `tasks/status.md`

## Task P17-T01: Stripe State Adoption

**Status:** done

- [x] Make checkout session and payment intent create/retrieve/list stateful.
- [x] Add idempotency key handling for create endpoints.
- [x] Keep existing scenario tests green while adding stateful tests first.
- [x] Run `/usr/local/go/bin/go test ./adapters/stripe -v`.

## Task P17-T02: OpenAI State Adoption

**Status:** done

- [x] Persist response IDs and chat completion IDs for retrieve-compatible fake workflows where applicable.
- [x] Preserve Phase 13 SSE behavior for `stream_success`.
- [x] Add validation for model, messages/input, and unsupported parameters where stateful workflows need it.
- [x] Run `/usr/local/go/bin/go test ./adapters/openai -v`.

## Task P17-T03: OAuth And Messaging State Adoption

**Status:** done

- [x] Persist GitHub OAuth codes, tokens, scopes, and fake user identities.
- [x] Persist Slack messages by channel and timestamp.
- [x] Add report metadata for stateful adapter coverage.
- [x] Run `/usr/local/go/bin/go test ./adapters/githuboauth ./adapters/slack -v`.

## Phase 17 Exit

- [x] Major adapters use the shared state foundation where supported.
- [x] Existing scenario behavior remains backward compatible.
- [x] Idempotency and validation behavior is adapter-visible where implemented.
- [x] Provider-specific compatibility tracks can build on stateful workflows.
