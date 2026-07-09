# Adapter Scenario Matrix

Built-in adapters expose named scenarios that control deterministic success and error behavior during local tests. This page is a single overview for scope checks and reviews.

**Authoritative source:** scenario names, response shapes, and endpoint coverage live in each adapter specification under `docs/adapters/<adapter>.md`. If this matrix and an adapter spec disagree, follow the adapter spec.

Configure the active scenario per adapter in Mockport config (for example `adapters.stripe.scenario`). See [Scenario Policy](../scenario-policy.md) for built-in versus user-defined scenarios.

`timeout` scenarios return an immediate timeout response shape. To exercise client-side timeout handling, use the server-wide `X-Mockport-Delay` header; see [Adapters](adapters.md).

| Adapter | Scenario | Short use | Spec |
| --- | --- | --- | --- |
| `stripe` | `payment_success` | Stripe-like `200` success responses for supported payment workflows. | [Stripe adapter](../adapters/stripe.md) |
| `stripe` | `payment_failed` | `402` / `card_declined` payment failure envelope. | [Stripe adapter](../adapters/stripe.md) |
| `stripe` | `auth_error` | `401` / `invalid_api_key` authentication failure. | [Stripe adapter](../adapters/stripe.md) |
| `stripe` | `rate_limited` | `429` / `rate_limited` rate-limit envelope. | [Stripe adapter](../adapters/stripe.md) |
| `stripe` | `timeout` | Immediate `504` / `mockport_timeout` response shape (no scenario-induced sleep). | [Stripe adapter](../adapters/stripe.md) |
| `openai` | `chat_success` | Default successful chat, responses, embeddings, files, and batch workflows. | [OpenAI adapter](../adapters/openai.md) |
| `openai` | `stream_success` | SSE-compatible chat completion chunks for streaming requests. | [OpenAI adapter](../adapters/openai.md) |
| `openai` | `rate_limited` | HTTP `429` with `Retry-After: 1` and OpenAI-like rate limit error envelope. | [OpenAI adapter](../adapters/openai.md) |
| `openai` | `context_length_exceeded` | OpenAI-like context length exceeded behavior. | [OpenAI adapter](../adapters/openai.md) |
| `openai` | `auth_error` | Authentication-style failures for protected calls. | [OpenAI adapter](../adapters/openai.md) |
| `github-oauth` | `oauth_success` | Default successful authorize redirect, token exchange, and user lookups. | [GitHub OAuth adapter](../adapters/github-oauth.md) |
| `github-oauth` | `invalid_code` | Token exchange failure for unknown or invalid authorization codes. | [GitHub OAuth adapter](../adapters/github-oauth.md) |
| `github-oauth` | `expired_token` | Protected-resource authentication failures for expired tokens. | [GitHub OAuth adapter](../adapters/github-oauth.md) |
| `github-oauth` | `scope_missing` | Scope-related failures on protected endpoints. | [GitHub OAuth adapter](../adapters/github-oauth.md) |
| `github-oauth` | `redirect_uri_mismatch` | Token exchange failure when posted `redirect_uri` does not match authorization. | [GitHub OAuth adapter](../adapters/github-oauth.md) |
| `slack` | `message_success` | Default successful messaging, conversations, and Events API subset workflows. | [Slack adapter](../adapters/slack.md) |
| `slack` | `auth_error` | Slack-like `invalid_auth` / `ok:false` behavior. | [Slack adapter](../adapters/slack.md) |
| `slack` | `rate_limited` | HTTP `429` with `Retry-After: 1` and Slack-like `ratelimited` error body. | [Slack adapter](../adapters/slack.md) |
| `slack` | `delivery_failed` | Slack-like message delivery failure behavior. | [Slack adapter](../adapters/slack.md) |
| `slack` | `channel_not_found` | Slack-like channel lookup failure behavior. | [Slack adapter](../adapters/slack.md) |
| `slack` | `not_in_channel` | Slack-like channel membership failure behavior. | [Slack adapter](../adapters/slack.md) |
| `line` | `line_success` | Default successful Messaging API, LINE Login, LIFF, MINI App, and payment helpers. | [LINE adapter](../adapters/line.md) |
| `line` | `auth_error` | Authentication failures for token-protected calls. | [LINE adapter](../adapters/line.md) |
| `line` | `rate_limited` | Rate limit behavior for Messaging API-like sends. | [LINE adapter](../adapters/line.md) |
| `line` | `invalid_request` | Request validation-style failures. | [LINE adapter](../adapters/line.md) |
| `line` | `pay_failed` | LINE Pay or Mini Dapp payment failure behavior. | [LINE adapter](../adapters/line.md) |
| `zoho-oauth` | `oauth_success` | Default successful authorize redirect, token exchange, and user info lookup. | [Zoho OAuth adapter](../adapters/zoho-oauth.md) |
| `zoho-oauth` | `invalid_code` | Token-exchange failure for unknown or invalid authorization codes. | [Zoho OAuth adapter](../adapters/zoho-oauth.md) |
| `zoho-oauth` | `invalid_token` | User info authentication failure (`401`) for missing or unknown tokens. | [Zoho OAuth adapter](../adapters/zoho-oauth.md) |

For endpoint coverage and maturity notes, see the [support matrix](support-matrix.md).
