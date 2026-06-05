# Phase 18 Stripe Provider Compatibility Implementation Plan

[日本語版](phase18_stripe_provider_compatibility.ja.md)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Stripe adapter を scenario-compatible から Stripe SDK/workflow-compatible local API へ引き上げる。

**Architecture:** Stripe の主要決済 workflow を fixture/spec/SDK contract に基づいて拡張する。内部決済処理、fraud、billing network は再現せず、公開 API の request/response/error/state/idempotency を高忠実度にする。

**Tech Stack:** Go Stripe adapter, Stripe SDK contract tests, compatibility manifests, sanitized fixtures.

---

## Files

- Create: `compat/fixtures/stripe/*`
- Create: `contract/sdk/stripe-smoke.test.js`
- Modify: `adapters/stripe/*`
- Modify: `docs/site/support-matrix.md`
- Modify: `tasks/status.md`

## Task P18-T01: Stripe SDK Contract Baseline

**Status:** done

- [x] Write failing SDK smoke for checkout session create/retrieve/list.
- [x] Write failing SDK smoke for payment intent create/retrieve/list.
- [x] Record Stripe SDK version in compatibility manifest.
- [x] Run `bash scripts/run-sdk-contracts.sh stripe`.

## Task P18-T02: Stripe Major Surface Expansion

**Status:** done

- [x] Add customers, products, prices, subscriptions, invoices, refunds, and webhooks coverage backlog.
- [x] Implement one endpoint group at a time with failing tests first.
- [x] Add SDK contract coverage for each endpoint group.
- [x] Update support matrix and compatibility score.

## Task P18-T03: Stripe Error And Idempotency Fidelity

**Status:** done

- [x] Add validation error fixtures for missing required fields and malformed IDs.
- [x] Add auth, rate limit, idempotency replay, and conflict error fixtures.
- [x] Implement response shape and headers close to Stripe public API.
- [x] Run adapter tests, SDK contracts, and compatibility report.

## Phase 18 Exit

- [x] Stripe adapter is at least `workflow-compatible`.
- [x] Stripe SDK contracts pass for supported workflows.
- [x] Stripe support matrix shows endpoint and scenario coverage.
- [x] Known gaps are explicit.
