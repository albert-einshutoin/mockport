# LINE Adapter Specification

[日本語版](line.ja.md)

This document describes the Mockport `line` adapter contract. It is not a copy of LINE's platform documentation and does not claim full LINE platform compatibility.

## Scope

The `line` adapter provides deterministic local behavior for selected LINE platform workflows:

- Messaging API-like push, reply, profile, and webhook test calls.
- Messaging API-like multicast, broadcast, narrowcast, message validation, signed webhook delivery helper, content retrieval, quota/delivery lookups, mark-as-read, loading animation, webhook endpoint settings, group/room lookups, rich menu operations, and channel access token helpers.
- LINE Login-like OAuth 2.0 authorization code flow, token exchange, and profile lookup.
- LIFF local helpers for profile and context.
- LINE MINI App service message-like notification token and send calls.
- LINE Pay v3-like payment request, confirmation, and status check.
- Mini Dapp SDK-like wallet session and payment helpers.

## Base Path

Default base path:

```text
/line
```

Example config:

```yaml
adapters:
  line:
    enabled: true
    base_path: /line
    scenario: line_success
    fake_secret: mockport_line_channel_token
    webhook:
      target_url: http://app:3000/webhooks/line
      signing_secret: mockport_line_secret
```

## Source References

The Messaging API surface is based on the Messaging API reference supplied for this implementation pass. Mockport implements deterministic local response shapes for common request/response workflows, but it does not enforce the complete official schema.

The LINE Login section is based on the official LINE Developers overview supplied to this repository review. The source describes LINE Login as a social login service for LINE accounts, based on OAuth 2.0 and OpenID Connect for web integrations, with native SDKs for iOS, Android, Unity, and Flutter.

Primary public references:

- LINE Developers top page: `https://developers.line.biz/en/`
- Messaging API reference: `https://developers.line.biz/en/reference/messaging-api/`
- LINE Login overview: `https://developers.line.biz/en/docs/line-login/overview/`
- LINE Login v2.1 API reference: `https://developers.line.biz/en/reference/line-login/`
- LIFF overview: `https://developers.line.biz/en/docs/liff/overview/`
- LINE MINI App API reference: `https://developers.line.biz/en/reference/line-mini-app/`
- LINE MINI App service messages: `https://developers.line.biz/en/docs/line-mini-app/develop/service-messages/`
- LINE Pay payment request: `https://developers-pay.line.me/online-api-v3/request-payment`
- LINE Pay payment confirmation: `https://developers-pay.line.me/online-api-v3/confirm-payment`
- LINE Pay payment request status: `https://developers-pay.line.me/online-api-v3/check-payment-request-status`
- Mini Dapp SDK: `https://docs.dappportal.io/mini-dapp/mini-dapp-sdk`
- Mini Dapp payment provider: `https://docs.dappportal.io/mini-dapp/mini-dapp-sdk/payment`
- OpenID Provider Configuration Document: `https://access.line.me/.well-known/openid-configuration`
- LINE Login security checklist: `https://developers.line.biz/en/docs/line-login/security-checklist/`

## Official Reference Map

Use this table to jump from Mockport's supported local surface to the closest official documentation. These links are references for behavior shape only; Mockport remains a deterministic local emulator and doesn't claim complete provider compatibility.

| Mockport surface | Official reference |
| --- | --- |
| Messaging API common response shape, domains, status codes, errors, and rate-limit concepts | `https://developers.line.biz/en/reference/messaging-api/#common-specifications` |
| Webhook request body, event objects, and signature validation | `https://developers.line.biz/en/reference/messaging-api/#webhooks` |
| Webhook endpoint settings and webhook test | `https://developers.line.biz/en/reference/messaging-api/#webhook-settings` |
| Message content, preview, and transcoding status | `https://developers.line.biz/en/reference/messaging-api/#getting-content` |
| Channel access tokens | `https://developers.line.biz/en/reference/messaging-api/#channel-access-token` |
| Reply, push, multicast, narrowcast, broadcast, mark-as-read, loading animation, quotas, delivery stats, and message validation | `https://developers.line.biz/en/reference/messaging-api/#message` |
| Users, profile, and follower IDs | `https://developers.line.biz/en/reference/messaging-api/#users` |
| LINE Official Account bot info | `https://developers.line.biz/en/reference/messaging-api/#line-official-account-bot` |
| Group chat endpoints | `https://developers.line.biz/en/reference/messaging-api/#group-chats` |
| Multi-person chat endpoints | `https://developers.line.biz/en/reference/messaging-api/#multi-person-chats` |
| Rich menu endpoints | `https://developers.line.biz/en/reference/messaging-api/#rich-menu` |
| Per-user rich menu endpoints | `https://developers.line.biz/en/reference/messaging-api/#per-user-rich-menu` |
| Rich menu alias endpoints | `https://developers.line.biz/en/reference/messaging-api/#rich-menu-alias` |
| Account link token | `https://developers.line.biz/en/reference/messaging-api/#account-link` |
| Message object and action object shapes | `https://developers.line.biz/en/reference/messaging-api/#message-objects` and `https://developers.line.biz/en/reference/messaging-api/#action-objects` |
| LINE Login OAuth and profile flow | `https://developers.line.biz/en/docs/line-login/overview/` and `https://developers.line.biz/en/reference/line-login/` |
| LIFF browser and LIFF app behavior | `https://developers.line.biz/en/docs/liff/overview/` |
| LINE MINI App service message API | `https://developers.line.biz/en/reference/line-mini-app/#service-messages` |
| LINE MINI App service message workflow and policy | `https://developers.line.biz/en/docs/line-mini-app/develop/service-messages/` |
| LINE Pay request flow | `https://developers-pay.line.me/online-api-v3/request-payment` |
| LINE Pay confirmation flow | `https://developers-pay.line.me/online-api-v3/confirm-payment` |
| LINE Pay payment request status check | `https://developers-pay.line.me/online-api-v3/check-payment-request-status` |
| Mini Dapp SDK wallet/payment context | `https://docs.dappportal.io/mini-dapp/mini-dapp-sdk` |
| Mini Dapp payment flow and payment APIs | `https://docs.dappportal.io/mini-dapp/mini-dapp-sdk/payment` |

## Minimum Required Coverage

This adapter treats the following surfaces as the minimum useful LINE baseline for local bot integration tests:

| Requirement | Status | Mockport behavior |
| --- | --- | --- |
| Send a message from app code | Implemented | Push/reply return `sentMessages`; multicast/broadcast return `{}`; narrowcast returns `202` plus progress. |
| Receive a LINE-like webhook in app code | Implemented | `POST /line/test/webhook/send` sends a webhook payload to the configured `webhook.target_url`. |
| Verify webhook signature in app code | Implemented for local delivery | The webhook sender signs the raw JSON body with HMAC-SHA256 and the `x-line-signature` header, using `webhook.signing_secret` or `mockport_line_secret`. |
| Validate common bad message payloads | Partial | Message validation returns LINE-style `details[].property` errors for missing message count, non-object messages, unsupported `type`, and empty text messages. |
| Exercise profile and account lookup paths | Implemented | Profile, bot info, follower IDs, group, and room helper endpoints return deterministic data. |
| Exercise rich menu lifecycle | Partial | Core rich menu create/list/get/delete/image/link/alias paths are stateful; deep image, area, and action validation is not complete. |
| Exercise token lifecycle | Partial | Channel token issue/verify/revoke helpers return deterministic fake tokens; JWT assertion and key registration are not cryptographically enforced. |

## Messaging API Contract

Mockport implements a workflow-compatible subset of the LINE Messaging API for local application tests.

Supported Messaging API-like endpoint groups:

| Group | Supported behavior |
| --- | --- |
| Send messages | Push/reply return `sentMessages`; multicast/broadcast return `{}`; narrowcast returns `202` and exposes a deterministic progress lookup. |
| Validate messages | `POST /v2/bot/message/validate/{type}` accepts 1 to 5 message objects and returns LINE-style details for invalid payloads. |
| Message content | Content, preview, and transcoding endpoints return deterministic local binary/status responses. No retention window is modeled. |
| Webhook settings and delivery | Webhook endpoint `PUT`/`GET` stores a valid HTTPS URL in process memory; webhook test returns a deterministic success result; `/test/webhook/send` sends a signed LINE-like webhook to the configured local target. |
| Bot/account info | Bot info, quota, quota consumption, delivery statistics, aggregation info/list, and follower ID lookups return deterministic data. |
| Chats | Mark-as-read and loading animation endpoints acknowledge valid local calls. |
| Group/room | Group summary, member IDs, member profile, room member IDs/profile, and leave acknowledgements are supported. |
| Rich menu | Create, validate, list, get, delete, image upload/download, default link, per-user link, batch acknowledgements, progress, and alias operations are supported. |
| Channel access tokens | v2.1, stateless v3, and short-lived token issue/verify/revoke helpers return deterministic fake token material. |

Response headers include a deterministic `X-Line-Request-Id` so SDK or client code that records LINE request IDs can be exercised locally.

## LINE Login Contract

Mockport implements the local test surface needed for apps that use LINE Login as an OAuth-style provider.

Supported LINE Login-like endpoints:

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/line/oauth2/v2.1/authorize` | Creates a fake authorization code and redirects to `redirect_uri`. |
| `POST` | `/line/oauth2/v2.1/token` | Exchanges a fake code for a fake access token, refresh token, and ID token. |
| `GET` | `/line/v2/profile` | Returns a deterministic local profile for the bearer token. |

Supported claims and profile fields:

| Field | Mockport behavior |
| --- | --- |
| `userId` | Deterministic fake user ID, usually `Umockport`. |
| `displayName` | Deterministic fake display name. |
| `pictureUrl` | Deterministic non-production URL. |
| `statusMessage` | Deterministic local status message. |
| `id_token` | Placeholder token string, not a signed JWT. |

Supported flow:

1. App redirects to `/line/oauth2/v2.1/authorize` with `client_id`, `redirect_uri`, `state`, and optional `scope`.
2. Mockport redirects back to `redirect_uri` with `code` and `state`.
3. App posts the code, `client_id`, and `redirect_uri` to `/line/oauth2/v2.1/token`.
4. Mockport returns fake token material.
5. App calls `/line/v2/profile` with `Authorization: Bearer <access_token>`.

The default LINE Login flow rejects missing `client_id` at authorization time and rejects token exchange when `client_id` is missing or does not match the code-producing authorization request.

## Authentication Methods

The official LINE Login overview distinguishes auto login, email/password login, QR code login, and SSO login. Mockport does not reproduce those UI-level authentication methods. The local adapter simulates the resulting authorization code flow after user authentication has succeeded.

Two-factor authentication is also treated as a provider-side policy boundary. Mockport does not emulate verification-code screens, trusted-browser lifetime, account switching, or channel console settings.

## Scenarios

| Scenario | Behavior |
| --- | --- |
| `line_success` | Default successful local workflow. |
| `auth_error` | Returns authentication failures for token-protected calls. |
| `rate_limited` | Returns rate limit behavior for Messaging API-like sends. |
| `invalid_request` | Returns request validation-style failures. |
| `pay_failed` | Returns LINE Pay or Mini Dapp payment failure behavior. |

## State

The adapter uses local deterministic state for:

| Resource | Purpose |
| --- | --- |
| `oauth_code` | Authorization codes issued by `/oauth2/v2.1/authorize`. |
| `oauth_token` | Access tokens issued by `/oauth2/v2.1/token`. |
| `message` | Messaging API-like sent message records. |
| `rich_menu` | Rich menu definitions and image upload status. |
| `rich_menu_alias` | Rich menu alias mappings. |
| `user_rich_menu` | Per-user rich menu links. |
| `notification_token` | MINI App service message notification tokens. |
| `line_pay_payment` | LINE Pay-like payment reservations and confirmations. |
| `mini_dapp_payment` | Mini Dapp-like local payment records. |

IDs are deterministic within the process and reset when the Mockport process restarts.

## Known Gaps

The `line` adapter is `workflow-compatible`, not provider-compatible.

Known gaps:

- No real LINE Login UI, QR code login, auto login, SSO login, or two-factor authentication screen.
- No signed or provider-verifiable ID token.
- No OpenID Connect discovery endpoint exposed by Mockport.
- No real LINE SDK contract harness yet.
- No real LIFF browser runtime.
- No provider-driven webhook redelivery, retry scheduler, or complete webhook event catalog. The local helper can send signed webhook payloads on demand.
- No monthly quota, free-message, rate-limit bucket, or concurrent audience operation enforcement beyond deterministic scenarios.
- No full Messaging API schema validation for every message, Flex, template, action, audience, insight, coupon, membership, or rich menu field.
- No real media storage lifecycle; content and preview endpoints return local placeholder bytes.
- No LINE Developers Console channel settings or review workflow.
- No regional policy enforcement.
- Mini Dapp endpoints are local SDK-style helpers, not a full Dapp Portal clone.

## Verification

Run the adapter and package tests:

```bash
/usr/local/go/bin/go test ./adapters/line ./internal/server ./internal/cli ./internal/config
```

Run all tests:

```bash
/usr/local/go/bin/go test ./...
```

Run the engineering gate:

```bash
bash scripts/check-go-engineering.sh
```
