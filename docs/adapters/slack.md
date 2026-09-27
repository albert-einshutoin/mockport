# Slack Adapter Specification

[日本語版](slack.ja.md)

Shared rules: [Scenario Policy](../scenario-policy.md).

This document describes the Mockport `slack` adapter contract. It is not a copy of Slack's documentation and does not claim full Slack platform compatibility.

## Scope

The `slack` adapter provides deterministic local behavior for selected Slack Web API and Events API workflows:

- `auth.test`.
- `chat.postMessage`, `chat.update`, and `chat.delete`.
- `conversations.list` and `conversations.history`.
- Events API URL verification, a message callback subset, and one local signed message delivery helper.
- Slack-like `ok:false` error bodies for auth, rate limit, delivery failure, channel membership, channel lookup, and signature failures.

## Base Path

Default base path:

```text
/slack
```

Example config:

```yaml
adapters:
  slack:
    enabled: true
    base_path: /slack
    scenario: message_success
    fake_secret: mockport_slack_token
    webhook:
      signing_secret: mockport_slack_signing_secret
```

## Official Reference Map

Use this table to jump from Mockport's supported local surface to the closest official Slack documentation. These links are references for behavior shape only; Mockport remains a deterministic local emulator.

| Mockport surface | Official reference |
| --- | --- |
| `auth.test` | `https://api.slack.com/methods/auth.test` |
| `chat.postMessage` | `https://api.slack.com/methods/chat.postMessage` |
| `chat.update` | `https://api.slack.com/methods/chat.update` |
| `chat.delete` | `https://api.slack.com/methods/chat.delete` |
| `conversations.list` | `https://api.slack.com/methods/conversations.list` |
| `conversations.history` | `https://api.slack.com/methods/conversations.history` |
| Events API URL verification | `https://api.slack.com/events/url_verification` |
| Events API message callbacks | `https://docs.slack.dev/apis/events-api/` |
| Slack request signing | `https://docs.slack.dev/authentication/verifying-requests-from-slack/` |
| Python `SignatureVerifier` | `https://docs.slack.dev/tools/python-slack-sdk/reference/signature/index.html` |
| Thread reply with `chat.postMessage` | `https://docs.slack.dev/reference/methods/chat.postMessage/` |
| Thread retrieval (not implemented) | `https://docs.slack.dev/reference/methods/conversations.replies/` |

## Supported Endpoints

| Method | Path | Purpose |
| --- | --- | --- |
| `POST` | `/slack/api/auth.test` | Returns deterministic Slack identity/auth status. |
| `POST` | `/slack/api/chat.postMessage` | Creates a local message; retains optional `thread_ts` in response and store. Thread replies are excluded from ordinary channel history. |
| `POST` | `/slack/api/chat.update` | Updates a local message. |
| `POST` | `/slack/api/chat.delete` | Deletes a local message. |
| `POST` | `/slack/api/conversations.list` | Returns deterministic channels. |
| `GET` | `/slack/api/conversations.history` | Returns deterministic channel history. |
| `POST` | `/slack/api/conversations.history` | Returns deterministic channel history. |
| `POST` | `/slack/events` | Handles URL verification and a message callback subset. |
| `POST` | `/slack/test/event/send` | Loopback-only helper that signs and sends the fixed `event_callback` message fixture to the configured safe local target. Requires `webhook.target_url` and `webhook.signing_secret`; returns delivery status, not an SDK reply claim. |
| `POST` | `/slack/test/reset` | Clears local state and idempotency records for test isolation. |

## Scenarios

| Scenario | Behavior |
| --- | --- |
| `message_success` | Default successful local messaging workflow. |
| `auth_error` | Returns Slack-like `invalid_auth` behavior. |
| `rate_limited` | Returns HTTP 429 with `Retry-After: 1` and Slack-like `{"ok":false,"error":"ratelimited"}` body. |
| `delivery_failed` | Returns Slack-like delivery failure behavior. |
| `channel_not_found` | Returns Slack-like channel lookup failure behavior. |
| `not_in_channel` | Returns Slack-like channel membership failure behavior. |

## Selected event and reply contract

`fake_secret` is the Web API token. Only `webhook.signing_secret` signs or verifies Events API requests. Existing configurations that used `fake_secret` as the signing key must set a separate `webhook.signing_secret`; `/slack/events` returns `503 missing_signing_secret` when it is absent. The send helper rejects a missing target or signing secret with a Slack-shaped error. No implicit signing-key fallback exists.

The helper sends the raw JSON body represented by [`events_message_callback.json`](../../compat/fixtures/slack/events_message_callback.json), with a fresh `X-Slack-Request-Timestamp` and `v0` HMAC signature. The fixed envelope includes `team_id`, `api_app_id`, `event_id`, and `event_time`; its inner ordinary message includes `channel`, `user`, `text`, `ts`, `event_ts`, and `channel_type`. Its fixture timestamp is example data; the signed delivery timestamp is current Unix seconds. Only `message_success` is supported by this helper. A non-2xx target response is a `502 event_target_non_2xx` with the target status, and a send timeout is `504 event_send_timeout`. Delivery is bounded and redirects are not followed.

The [Python sample app](../../examples/app-e2e/slack-app/server.py) validates the unmodified body with the official `SignatureVerifier`, then sends the fixed text `Mockport thread reply` through pinned `slack-sdk==3.44.1` to `channel=event.channel` and `thread_ts=event.ts`. Its one-second SDK timeout bounds the synchronous example. The app E2E checks the SDK response and one `chat.postMessage` report entry; the HTTP-handler adapter test checks the saved reply's channel, thread timestamp, and text. The parent event is injected from the selected fixture; parent-message retrieval and the full thread are not emulated. `conversations.history` does not expose thread replies and `conversations.replies` is not implemented.

## Current Gaps And Tasks

| Priority | Task | Current source of truth |
| --- | --- | --- |
| P2 | Keep real workspace delivery, Events API completeness, parent/thread retrieval, Block Kit validation, files, app scopes, enterprise policy, and workspace directory as known gaps. | `docs/site/support-matrix.md` |

## Verification

Run the adapter tests and client contract:

```bash
go test ./adapters/slack
bash scripts/run-sdk-contracts.sh slack
bash scripts/run-app-e2e.sh slack
bash scripts/check-slack-app-mutation.sh
```
