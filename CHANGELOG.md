# Changelog

[日本語版](CHANGELOG.ja.md)

## Unreleased

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
