# Roadmap

[日本語版](ROADMAP.ja.md)

Mockport is a Docker-first local API environment for AI-native development and CI. This roadmap is intentionally scoped to public preview work and provider-compatible direction without promising full provider internals.

## Current Release

- `v0.1.0-alpha`: public preview with Docker/GHCR, GitHub release archives, AI-safe env policy, and scenario-compatible adapters.

## Current Mainline

- Workflow-compatible local adapters for Stripe-like payments, OpenAI-compatible API, GitHub OAuth-like API, Slack-like messaging API, and LINE-like platform APIs.
- Compatibility reports are generated from runtime metadata and known-gap mappings.
- Shared deterministic state, idempotency primitives, report hooks, and Go engineering hardening are in place.

## Near Term

- Phase 24: recover observable GitHub Actions execution if compatibility workflow runs are missing.
- Phase 25: expand the SDK/client contract harness beyond provider-specific smoke coverage.
- Phase 26: add versioned compatibility manifests and automated provider-compatible promotion gates.
- Phase 27-29: deepen provider-specific contract evidence for Stripe, OpenAI, GitHub OAuth, Slack, and LINE where applicable.
- Phase 30: publish `v0.2.0-preview` with refreshed compatibility report and post-release smoke evidence.

## Public Preview Follow-up

- [#6](https://github.com/albert-einshutoin/mockport/issues/6): Add SSE-compatible streaming response for OpenAI `stream_success`.
- [#8](https://github.com/albert-einshutoin/mockport/issues/8): Improve `mockport up` Docker Compose UX with clearer errors and `--detach` / `--build`.
- [#5](https://github.com/albert-einshutoin/mockport/issues/5): Clarify built-in scenario policy versus user-defined scenarios.
- [#7](https://github.com/albert-einshutoin/mockport/issues/7): Track adapter helper duplication before adding more adapters.

## Adapter Direction

Current adapters:

- Stripe-like payments.
- OpenAI-compatible API.
- GitHub OAuth-like API.
- Slack-like messaging API.
- LINE-like platform APIs.

Candidate adapters after the compatibility foundation:

- SendGrid-like email API.

## Compatibility Direction

Mockport aims for provider-compatible local APIs for selected workflows. Compatibility is measured by documented endpoint behavior, SDK contract tests, fake state, error shape, and reportable gaps.

Mockport does not reproduce provider internal logic, undocumented behavior, or production network effects.

## Non-Goals

- Proxying real provider traffic.
- Accepting real provider secrets in public examples.
- Claiming full provider compatibility before SDK and workflow contract evidence exists.
- Publishing npm or Homebrew as primary channels before Docker and Go binary release paths are stable.
