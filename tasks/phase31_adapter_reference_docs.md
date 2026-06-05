# Phase 31 Adapter Reference And Task Inventory

[日本語版](phase31_adapter_reference_docs.ja.md)

> **For agentic workers:** keep this file in sync with `internal/cli/run.go`, `docs/adapters/*.md`, and the adapter entries in `docs/site/adapters.md`.

**Goal:** Find the issues and tasks for every currently registered adapter, and make the reproduced provider domains/endpoints traceable to official documentation.

**Current registered adapters:** `stripe`, `openai`, `github-oauth`, `slack`, and `line`.

**Implementation source of truth:** `internal/cli/run.go` registers adapters in `defaultRegistry()`.

## Tasks

| ID | Task | Status | Evidence |
| --- | --- | --- | --- |
| P31-T01 | Inventory adapters registered by the CLI default registry | done | `internal/cli/run.go` registers `stripe`, `openai`, `github-oauth`, `slack`, and `line` |
| P31-T02 | Find existing open adapter work | done | Phase 26-29 remain pending in `tasks/status.md` |
| P31-T03 | Add per-adapter official reference maps for missing adapter docs | done | `docs/adapters/stripe.md`, `docs/adapters/openai.md`, `docs/adapters/github-oauth.md`, `docs/adapters/slack.md` |
| P31-T04 | Link every per-adapter spec from the docs site adapter index | done | `docs/site/adapters.md` |
| P31-T05 | Keep provider-compatible promotion blocked until manifests, fixtures, SDK/client contracts, and known-gap reports are wired | pending | `tasks/phase26_provider_compatible_manifest_promotion.md` |
| P31-T06 | Freeze the implementation line for adapter work | done | `docs/compatibility-model.md` and this file define in-scope, out-of-scope, adapter-by-adapter boundaries, and work order |

## Existing Open Adapter Work

| Adapter | Priority | Open work | Official reference map |
| --- | --- | --- | --- |
| `stripe` | P1 | Manifest scope, SDK contract deepening, fixture coverage, provider-compatible promotion decision. | `docs/adapters/stripe.md` |
| `openai` | P1 | Manifest scope, SSE/error SDK contracts, response and batch retrieve consistency, fake inference boundaries. | `docs/adapters/openai.md` |
| `github-oauth` | P1 | OAuth/client contract coverage for state, redirect mismatch, invalid code, missing scope, bad credentials, and REST subset. | `docs/adapters/github-oauth.md` |
| `slack` | P1 | Official SDK feasibility, deeper client contract, lifecycle assertions, Events API evidence, request signing evidence. | `docs/adapters/slack.md` |
| `line` | P2 | Official LINE SDK contract harness and broader schema validation remain known gaps. | `docs/adapters/line.md` |

## Implementation Line

Adapter implementation should move from broad docs parity toward selected, testable local compatibility.

Global line:

- Implement only selected public API and SDK/client behavior that can run deterministically without external provider calls.
- Require official-reference grounding, tests, fixtures, and known-gap documentation before expanding any adapter surface.
- Keep all registered adapters at `workflow-compatible` unless Phase 26 manifest checks and promotion gates prove `contract` level evidence.
- Treat `provider-compatible` as a release gate outcome, not as a manual label.

In scope:

| Area | Implementation line |
| --- | --- |
| Endpoint behavior | Public path/method/status/header/response shape for selected workflows. |
| SDK/client behavior | Pinned SDK or client calls that can run against local Mockport only. |
| State | Deterministic fake create/retrieve/list/update or equivalent lifecycle state. |
| Errors | Common provider-style envelopes, status codes, retry/rate-limit hints, auth failures, and validation failures. |
| Evidence | Compatibility manifests, sanitized fixtures, contract tests, generated reports, and known gaps. |

Out of scope:

| Area | Boundary |
| --- | --- |
| Provider internals | No undocumented behavior, internal scheduling, risk engines, fraud checks, billing networks, settlement, or delivery guarantees. |
| External calls | No calls to Stripe, OpenAI, GitHub, Slack, LINE, or related production services from runtime or contract tests. |
| Provider policy | No real account policy, SSO, enterprise enforcement, review workflows, app store/console behavior, or regional policy. |
| UI/runtime surfaces | No provider login UI, QR/2FA screens, LIFF browser runtime, hosted tool runtime, or production webhook retry loops. |
| Quality parity | No real model quality, tokenization parity, Slack delivery, Block Kit completeness, or full provider schema catalog. |

Adapter-specific line:

| Adapter | Implement now | Do not implement in this track |
| --- | --- | --- |
| `stripe` | First provider-compatible candidate for Checkout Session, PaymentIntent, Customer/Product/Price, Subscription/Invoice, Refund, webhook, idempotency, and error-shape workflows. | Fraud, Connect, disputes, tax, payment network settlement, and full Billing lifecycle. |
| `openai` | API/SDK/streaming/state fidelity for Models, Chat Completions, Responses, Embeddings, Files, and Batches. | Real inference quality, tokenization parity, hosted tools, vector stores, and provider scheduling. |
| `github-oauth` | OAuth web app authorize/token flow plus `/user`, `/user/emails`, and `/user/orgs` bearer-token subset. | Repository APIs, GitHub Apps installation model, real org/enterprise policy, SSO, and permission graph enforcement. |
| `slack` | Web API messaging subset, conversations list/history, request signing, Events API URL verification and message callback evidence, plus official SDK feasibility result. | Real message delivery, full Events API, Block Kit validation completeness, files, app scopes, enterprise policy, and workspace directory. |
| `line` | Workflow-compatible manifest and known-gap evidence for the current Messaging API, Login, LIFF helper, MINI App service message, LINE Pay, and Mini Dapp helper surface. | Provider-compatible promotion until official SDK contract harness, LIFF runtime strategy, schema validation policy, and webhook redelivery boundary are defined. |

## Official Documentation Coverage

| Adapter | Covered provider domains/endpoints |
| --- | --- |
| `stripe` | Checkout Sessions, PaymentIntents, Customers, Products, Prices, Subscriptions, Invoices, Refunds, webhook signature verification. |
| `openai` | Models, Chat Completions, streaming responses, Responses create/retrieve, Embeddings, Files, Batches create/retrieve. |
| `github-oauth` | OAuth app authorize/token flow, authenticated user, authenticated emails, authenticated organizations, token request errors. |
| `slack` | `auth.test`, `chat.postMessage`, `chat.update`, `chat.delete`, `conversations.list`, `conversations.history`, `url_verification`, `message`, request signing. |
| `line` | Messaging API, LINE Login, LIFF, LINE MINI App service messages, LINE Pay, Mini Dapp SDK helper context. |

## Next Adapter Work Order

1. Finish Phase 26 manifests and promotion gate for every registered adapter, because every maturity decision depends on machine-readable evidence.
2. Continue Phase 27 for Stripe first, because it already has SDK evidence and fixture seeds.
3. Continue Phase 28 for OpenAI streaming/error fidelity.
4. Continue Phase 29 for GitHub OAuth and Slack client evidence.
5. Add LINE manifest evidence, but keep LINE as workflow-compatible until an official SDK contract harness and broader schema validation exist.

## Verification

```bash
bash scripts/check-doc-links.sh
bash scripts/check-public-trust.sh
```
