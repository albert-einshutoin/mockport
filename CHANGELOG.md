# Changelog

[日本語版](CHANGELOG.ja.md)

## Unreleased

## v0.3.0-preview - 2026-09-27

- Added opt-in fake-key validation for ordinary Stripe and OpenAI requests. Unsupported OpenAI Responses `stream:true` now returns 501.
- Added a Stripe app flow with valid inline Checkout items, unpaid Session creation, signed completion webhook, and one order update.
- Added OpenAI nonstreaming and Chat streaming app flows, bounded failures, and the response fields needed by unmodified LLM 0.36. Simulated usage counts are fixed values.
- Added a signed Slack message event flow with official SDK signature verification and a reply in the same thread.
- Pinned the Docker, binary, CLI init, and experimental npm wrapper entrypoints to this preview. See [release notes](docs/releases/v0.3.0-preview.md) for migration, reproducible samples, and limits.

## v0.2.0-preview - 2026-07-19

### Compatibility release track

- Added scheduled/manual compatibility CI for Stripe, OpenAI, GitHub OAuth, Slack, and the official LINE SDK contract.
- Added generated compatibility reports with compatibility scores, provider API versions, SDK/client evidence, and known gaps.
- Added release checks for maturity labels: `experimental`, `sdk-compatible`, `workflow-compatible`, and `provider-compatible`.
- Compared with the v0.1.0-alpha scope, current mainline now classifies all six built-in adapters—Stripe, OpenAI, GitHub OAuth, Slack, LINE, and Zoho OAuth—as `workflow-compatible` for documented selected workflows. This does not claim `provider-compatible` parity.

### Runtime and OSS experience

- Added deterministic state, reset endpoints, bounded request history, compatibility metadata, and broader selected workflow coverage across all six adapters.
- Separated deployment warnings from the narrower `public_env_safe` environment-commit signal.
- Added runnable official Stripe, OpenAI, and LINE Node.js SDK examples plus `llms.txt` guidance for AI coding agents.
- Refreshed Docker, binary archive, CLI init, and npm fallback paths to the `v0.2.0-preview` release contract.

## v0.1.0-alpha - 2026-05-26

Initial public preview release.

### Included

- Docker-first Mockport runtime with AI-safe configuration checks.
- Stripe-like payment adapter for checkout sessions, payment intents, webhook sending, and common error scenarios.
- Experimental OpenAI-compatible, GitHub OAuth-like, and Slack-like adapters.
- `/_mockport/report` for request history, scenario coverage, behavior matrix, and safety findings.
- GitHub Release archives for Linux and macOS on amd64 and arm64.
- GHCR image published as `ghcr.io/albert-einshutoin/mockport:0.1.0-alpha`.

### Known Limits

- This is scenario-compatible, not full provider-compatible.
- Provider SDK contract coverage starts in later phases.
- Homebrew and npm are not published distribution channels yet.
