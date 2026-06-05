# Phase 32 Service Baseline Execution

[日本語版](phase32_service_baseline_execution.ja.md)

**Goal:** Close the minimum useful integration-test baseline for every committed adapter, then add SendGrid only after the current service baselines are explicit and verifiable.

**Current registered adapters:** `stripe`, `openai`, `github-oauth`, `slack`, and `line`.

**Planned adapter:** `sendgrid`.

**Source of truth:** Development rules live in `docs/maintainer-guide.md` and compatibility promotion rules live in `docs/compatibility-model.md`. Adapter-specific contracts and gaps live in `docs/adapters/*.md`.

## Cross-Adapter Baseline

Each committed adapter should satisfy these requirements before any maturity promotion:

| Requirement | Acceptance criteria |
| --- | --- |
| Official reference map | `docs/adapters/{adapter}.md` links to official API docs for every supported feature group. |
| Metadata truthfulness | `Metadata()` capabilities, endpoints, state, and scenarios match implemented routes and docs. |
| Local fake env | `FakeEnv()` and init examples provide enough variables for local app integration. |
| Core happy path | Main workflow succeeds through real HTTP handler tests, not helper-only tests. |
| Auth failure | Token, key, or client-secret failure scenario is implemented and tested. |
| Rate-limit or quota failure | If provider exposes rate limits, a deterministic scenario or response header exists. |
| Validation error shape | Common malformed input returns provider-shaped errors with useful field or property details. |
| Stateful lifecycle | Created resources can be listed, retrieved, updated, or deleted when the provider workflow depends on it. |
| Webhook/callback support | Providers with callbacks can send or receive signed local callback payloads. |
| Report compatibility | Support matrix, adapter docs, and generated report metadata describe the same surface. |
| Tests and gate | Adapter tests, relevant server/config/cli tests, `go test ./...`, and `bash scripts/check-go-engineering.sh` pass. |

## Service Baselines

### Stripe

Minimum surface:

- Checkout Sessions: create, retrieve, list, `payment_status`, URL, and customer/client reference propagation.
- Payment Intents: create, retrieve, list, required field validation, success path, and decline path.
- Customers, Products, Prices, Subscriptions, Invoices, Refunds: create/list/retrieve with provider-shaped IDs and objects.
- Idempotency: replay matching requests and reject mismatched request fingerprints.
- Webhooks: signed local webhook sender for checkout success and payment failure.
- Errors: auth, card failure, rate limit, timeout, validation, and missing resource.
- SDK contract: Stripe SDK smoke for supported workflows.

Execution:

1. Audit `adapters/stripe` against this baseline and update `docs/adapters/stripe.md`.
2. Add failing tests for missing lifecycle or error paths.
3. Fill missing route and state behavior.
4. Verify SDK contract and update support matrix.

### OpenAI

Minimum surface:

- Models: list/retrieve with deterministic model inventory.
- Chat Completions: non-streaming and streaming, tool-call-shaped response, auth/rate/context errors.
- Responses API: non-streaming and streaming, retrieve/list where applicable, and tool-output surface where applicable.
- Embeddings: deterministic vector dimensions and batch input handling.
- Files: upload/list/retrieve/content/delete.
- Batches: create/retrieve/list/cancel with state transitions.
- Errors: auth, rate limit, context length, and invalid request.
- SDK contract: current OpenAI JS/Python SDK smoke for supported workflows.

Execution:

1. Reconfirm current OpenAI API docs before touching implementation.
2. Update `docs/adapters/openai.md` with any missing official reference map entries.
3. Add failing tests for SDK-critical missing endpoints and error shapes.
4. Implement missing behavior with deterministic responses only.
5. Verify SDK contract, `go test ./...`, and support docs.

### GitHub OAuth

Minimum surface:

- OAuth authorize redirect with state preservation.
- Access token exchange with invalid code, expired code, redirect URI mismatch, and scope handling.
- User profile: `/user` with deterministic ID, login, avatar, and email fields.
- Emails: `/user/emails` with primary/verified behavior.
- Orgs: `/user/orgs` with deterministic org list.
- Scope enforcement: missing scope scenario returns provider-shaped failure.
- Docs: explicit out-of-scope boundary for GitHub Apps, installations, repo permissions, SSO, and enterprise policy.

Execution:

1. Update `docs/adapters/github-oauth.md` with any missing boundaries.
2. Add failing tests for scope and token exchange edge cases not currently covered.
3. Implement missing scope/error behavior.
4. Update metadata and support matrix.

### Slack

Minimum surface:

- `auth.test`.
- Conversations: list, history, open or deterministic direct-message equivalent if app tests need bot DMs.
- Chat: postMessage, update, delete with stateful message lifecycle.
- Events API: URL verification and message callbacks.
- Signed requests: verify Slack signing secret for inbound Events API.
- Interactions: baseline `block_actions` or slash-command callback if app integration commonly depends on it.
- Block Kit: shallow validation for `blocks` array shape, unknown block type errors, and pass-through storage.
- Errors: invalid auth, rate limit with `Retry-After`, delivery failure, channel not found, and not in channel.

Execution:

1. Update `docs/adapters/slack.md` with any missing reference map entries.
2. Write failing tests for interaction callback and Block Kit shallow validation.
3. Implement minimal interaction route and validation.
4. Update metadata, support matrix, and multi-adapter examples.

### LINE

Minimum surface:

- Send messages: push, reply, multicast, broadcast, narrowcast, and progress.
- Webhook settings: endpoint set/get/test.
- Signed webhook delivery helper: local `x-line-signature` sender.
- Message validation: LINE-style `details[].property` for common message failures; expand beyond text over time.
- Content: content, preview, and transcoding placeholder responses.
- User/bot/group/room helpers: profile, followers, bot info, and group/room members.
- Rich menu: create/validate/list/get/delete, image upload/download, default/user link, alias, and batch ack.
- Channel access tokens: v2.1, v3 stateless, short-lived issue/verify/revoke helpers.
- LINE Login: authorize/token/profile.
- LIFF and MINI App helpers: profile/context and service messages.
- Errors: auth, rate limit, invalid request, and pay failure.

Execution:

1. Keep `docs/adapters/line.md` as the canonical baseline spec.
2. Expand message validation in small TDD slices: image/video/audio/location/sticker/template/flex.
3. Add webhook redelivery scenario and common webhook event fixture catalog.
4. Add official SDK or client-contract smoke when practical.

### SendGrid

Minimum surface:

- Mail send: `POST /v3/mail/send` accepts personalizations, from, subject, content, templates, and returns `202`.
- Validation: missing `from`, missing recipients, missing content/template returns SendGrid-shaped errors.
- Auth: API key failure scenario.
- Rate limit: deterministic `429` with rate-limit headers.
- Templates: create/list/retrieve minimal transactional template and version helpers if app tests need templates.
- Suppressions: global suppressions add/list/delete for apps that test unsubscribe behavior.
- Event webhook: local signed event sender for delivered, processed, dropped, bounce, open, click, spamreport, and unsubscribe.
- Docs and examples: add `docs/adapters/sendgrid.md`, then promote README from planned to supported only after baseline tests pass.

Execution:

1. Create `adapters/sendgrid/adapter.go`, `models.go`, and `adapter_test.go`.
2. Add config registration and `mockport init --adapter sendgrid`.
3. Implement mail send happy path first with RED/GREEN.
4. Add auth/rate/validation scenarios.
5. Add signed event webhook sender.
6. Add templates and suppressions only after core mail/webhook passes.
7. Update README, support matrix, examples, and metadata conformance.

## Execution Order

1. Documentation parity for existing adapters: Stripe, OpenAI, GitHub OAuth, Slack.
2. Slack baseline gaps: interactions and Block Kit shallow validation.
3. LINE validation expansion and webhook fixture catalog.
4. OpenAI current SDK contract refresh.
5. Stripe final audit against docs and idempotency/webhook evidence.
6. GitHub OAuth scope/error hardening.
7. SendGrid adapter from scratch.
8. Generate/update compatibility reports and support matrix.

## Verification

Run after each adapter slice:

```bash
/usr/local/go/bin/go test ./adapters/<adapter>
```

Run before every commit:

```bash
/usr/local/go/bin/go test ./...
bash scripts/check-go-engineering.sh
```

For docs-only changes:

```bash
rg -n "https?://" docs/adapters docs/site README.md
```

## Commit Discipline

Use one commit per independently working adapter slice. Never mix unrelated existing workspace changes into these commits.
