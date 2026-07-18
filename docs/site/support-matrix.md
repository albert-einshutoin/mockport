# Support Matrix

[日本語版](support-matrix.ja.md)

Mockport support is explicit and scenario-driven. Use reports to confirm what a test run exercised.

Generated compatibility reports live in [`docs/compatibility-reports`](../compatibility-reports/README.md). They include compatibility scores, provider API versions, SDK/client evidence, and known gaps.

## Maturity

| Maturity | Meaning |
| --- | --- |
| `experimental` | Early adapter coverage for selected workflows. Expect gaps. |
| `partial` | Common workflows are implemented with documented unsupported behavior. |
| `sdk-compatible` | Selected official SDK calls pass against local Mockport. |
| `workflow-compatible` | Selected workflows include fake state, errors, and replayable behavior. |
| `provider-compatible` | Selected provider workflows are backed by manifests, SDK contracts, fixtures, and known-gap reports. |

## Adapter Coverage

| Adapter | Maturity | Endpoints | Scenarios | Notes |
| --- | --- | --- | --- | --- |
| `stripe` | `workflow-compatible` | checkout sessions, payment intents, customers, products, prices, subscriptions, invoices, refunds, fake webhook sender | success, failure, auth error, rate limit, timeout | SDK evidence `stripe@22.3.1` passes for supported local workflows; fake state, list/retrieve, validation errors, and idempotency-key replay are available. Known gaps: no fraud, billing network, tax, disputes, Connect, or full Billing lifecycle. |
| `openai` | `workflow-compatible` | models, chat completions, responses, embeddings, files, batches | chat success, SSE stream success, rate limit, context length, auth error | SDK evidence `openai@6.46.0` passes for supported local workflows; deterministic fake inference, stateful response lookup, streaming, embeddings, files, and batches are available. Known gaps: no real model quality, tokenization parity, hosted tools, vector stores, or provider scheduling. |
| `github-oauth` | `workflow-compatible` | authorize, access token, user profile, user emails, user orgs | oauth success, invalid code, expired token, missing scope, redirect URI mismatch | OAuth client contract passes for supported local workflows; fake codes, tokens, scopes, expiry metadata, profile, email, and org subsets are available. Known gaps: no real GitHub policy, repository permissions, SSO, org/enterprise enforcement, or app installation model. |
| `slack` | `workflow-compatible` | auth.test, conversations.list, conversations.history, chat.postMessage, chat.update, chat.delete, Events API URL verification/message callback subset | message success, auth error, rate limit, delivery failure, channel not found, not in channel | Slack client contract passes for supported local workflows; fake workspace, bot, channel, message state, update/delete lifecycle, history, request signature check, and Slack-style `ok:false` errors are available. Known gaps: no real delivery, Events API completeness, Block Kit validation, files, app scopes, enterprise policy, or workspace directory. |
| `line` | `workflow-compatible` | Messaging API send/content/signed webhook/rich menu/channel token workflows, LINE Login authorize/token/profile, LIFF profile/context helpers, MINI App service messages, LINE Pay request/confirm/check, Mini Dapp wallet/payment helpers | line success, auth error, rate limit, invalid request, pay failed | Official `@line/bot-sdk@11.0.0` push contract plus deterministic local state for OAuth codes/tokens, sent messages, signed webhook delivery helper, webhook endpoint settings, rich menus, rich menu aliases, user rich menu links, notification tokens, LINE Pay payments, and Mini Dapp payments. Known gaps: other LINE SDK surfaces are unverified; there is no real LIFF browser runtime, provider-driven webhook redelivery loop, monthly quota/rate bucket enforcement beyond scenarios, full Messaging API schema validation, regional policy enforcement, or full Dapp Portal clone. |
| `zoho-oauth` | `workflow-compatible` | authorize redirect, access token exchange, user info | oauth success, invalid code, invalid token | Minimal Zoho OAuth2 authorization-code workflow: immediate authorize redirect with `state` echo, token exchange, and user info using the `Zoho-oauthtoken` auth scheme. Fake codes are one-time and client-bound; returned `Email`/`Display_Name` are configurable via env or authorize query. Known gaps: no real Zoho login UI, MFA, data-center/org routing, token refresh, scope enforcement, or full user profile fields. |

## Planned Compatibility Track

Provider-compatible work follows the model in `docs/compatibility-model.md` and is staged after public preview:

1. Compatibility manifests and scores.
2. SDK contract harness foundation.
3. Shared state foundation.
4. Adapter state adoption.
5. Provider-specific compatibility tracks.
