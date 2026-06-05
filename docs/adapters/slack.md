# Slack Adapter Specification

[日本語版](slack.ja.md)

This document describes the Mockport `slack` adapter contract. It is not a copy of Slack's documentation and does not claim full Slack platform compatibility.

## Scope

The `slack` adapter provides deterministic local behavior for selected Slack Web API and Events API workflows:

- `auth.test`.
- `chat.postMessage`, `chat.update`, and `chat.delete`.
- `conversations.list` and `conversations.history`.
- Events API URL verification and message callback subset.
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
| Events API message callbacks | `https://api.slack.com/events/message` |
| Slack request signing | `https://api.slack.com/docs/verifying-requests-from-slack` |

## Supported Endpoints

| Method | Path | Purpose |
| --- | --- | --- |
| `POST` | `/slack/api/auth.test` | Returns deterministic Slack identity/auth status. |
| `POST` | `/slack/api/chat.postMessage` | Creates a local message. |
| `POST` | `/slack/api/chat.update` | Updates a local message. |
| `POST` | `/slack/api/chat.delete` | Deletes a local message. |
| `POST` | `/slack/api/conversations.list` | Returns deterministic channels. |
| `GET` | `/slack/api/conversations.history` | Returns deterministic channel history. |
| `POST` | `/slack/api/conversations.history` | Returns deterministic channel history. |
| `POST` | `/slack/events` | Handles URL verification and a message callback subset. |

## Scenarios

| Scenario | Behavior |
| --- | --- |
| `message_success` | Default successful local messaging workflow. |
| `auth_error` | Returns Slack-like `invalid_auth` behavior. |
| `rate_limited` | Returns Slack-like rate limit behavior. |
| `delivery_failed` | Returns Slack-like delivery failure behavior. |
| `channel_not_found` | Returns Slack-like channel lookup failure behavior. |
| `not_in_channel` | Returns Slack-like channel membership failure behavior. |

## Current Gaps And Tasks

| Priority | Task | Current source of truth |
| --- | --- | --- |
| P1 | Test whether the official `@slack/web-api` client can be pinned and pointed at Mockport. If not, record the exact blocker in the manifest. | `tasks/phase29_oauth_slack_client_evidence.md` |
| P1 | Deepen client contract coverage for `invalid_auth`, `channel_not_found`, `not_in_channel`, `rate_limited` plus `Retry-After`, and `invalid_signature`. | `contract/sdk/slack-smoke.test.js` and `compat/fixtures/slack/` |
| P1 | Add message lifecycle assertions for post, update, history, delete, and deleted-message visibility. | `tasks/phase29_oauth_slack_client_evidence.md` |
| P1 | Add Events API URL verification and message callback contract evidence when official SDK coverage is not enough. | `compat/fixtures/slack/` |
| P2 | Keep real delivery, Events API completeness, Block Kit validation, files, app scopes, enterprise policy, and workspace directory as known gaps. | `docs/site/support-matrix.md` |

## Verification

Run the adapter tests and client contract:

```bash
/usr/local/go/bin/go test ./adapters/slack
bash scripts/run-sdk-contracts.sh slack
```
