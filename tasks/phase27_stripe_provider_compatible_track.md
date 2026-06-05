# Phase 27 Stripe Provider-compatible Track Implementation Plan

[日本語版](phase27_stripe_provider_compatible_track.ja.md)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Stripe first 方針に従い、Stripe adapter を最初の `provider-compatible` 候補として、選定workflowのcontract evidenceを厚くする。

**Architecture:** Stripe の全内部処理は再現しない。Checkout Session、PaymentIntent、Customer/Product/Price、Subscription/Invoice、Refund、Webhook、idempotency、error shape の選定workflowを manifest と SDK contract で固定し、known gaps を残したまま maturity を昇格できるか判断する。

**Tech Stack:** Go Stripe adapter, Stripe SDK `22.1.1`, contract fixtures, compatibility manifests.

---

## Files

- Modify: `adapters/stripe/*`
- Modify: `contract/sdk/stripe-smoke.test.js`
- Modify: `compat/manifests/stripe.json`
- Modify: `compat/fixtures/stripe/*`
- Modify: `docs/site/support-matrix.md`
- Modify: `docs/compatibility-reports/latest.md`
- Modify: `tasks/status.md`

## Task P27-T01: Stripe Provider-compatible Scope Definition

**Status:** pending

- [ ] Define selected workflows in `compat/manifests/stripe.json`: checkout, payment intent, refund, subscription/invoice, webhook, idempotency.
- [ ] Explicitly exclude fraud, Connect, disputes, tax, payment network settlement, and full Billing lifecycle.
- [ ] Add fixture references for each selected workflow.
- [ ] Run `node scripts/check-compat-manifests.mjs` and confirm the scope is valid.

## Task P27-T02: SDK Contract Deepening

**Status:** pending

- [ ] Add failing Stripe SDK contract cases for list pagination shape and retrieve-after-create for every selected resource.
- [ ] Add failing SDK contract for idempotency replay and idempotency conflict.
- [ ] Add failing SDK contract for validation errors with Stripe-like error envelope.
- [ ] Run `bash scripts/run-sdk-contracts.sh stripe` and confirm RED before implementation.

## Task P27-T03: Stripe Adapter Fidelity

**Status:** pending

- [ ] Implement minimal adapter changes to pass the new SDK contracts.
- [ ] Ensure all created resources are deterministic fake state and never call Stripe.
- [ ] Ensure idempotency replay returns the same response and conflicts return a stable error.
- [ ] Ensure webhook signature remains fake but deterministic and documented.

## Task P27-T04: Promotion Decision

**Status:** pending

- [ ] Regenerate compatibility reports.
- [ ] If manifest levels and score satisfy the gate, promote Stripe maturity to `provider-compatible`.
- [ ] If not, leave Stripe as `workflow-compatible` and record missing evidence in known gaps.
- [ ] Update support matrix, changelog, and release report with the decision.

## Phase 27 Exit

- [ ] Stripe selected workflows have manifest, fixtures, SDK contract, state, error, and known-gap evidence.
- [ ] Stripe either reaches `provider-compatible` through the automated gate or has explicit blockers.
- [ ] `bash scripts/run-sdk-contracts.sh stripe` and `all` pass.
- [ ] Compatibility report explains the Stripe maturity decision.
