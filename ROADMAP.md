# Roadmap

[日本語版](ROADMAP.ja.md)

Mockport is a Docker-first local API environment for AI-native development and CI. This roadmap is intentionally scoped to public preview work and provider-compatible direction without promising full provider internals.

## Current Release

- `v0.2.0-preview`: six workflow-compatible adapters, generated compatibility evidence, official Stripe/OpenAI/LINE SDK contracts, bounded reports, and AI-safe environment checks.

## Current Mainline

- Workflow-compatible local adapters for Stripe-like payments, OpenAI-compatible API, GitHub OAuth-like API, Slack-like messaging API, LINE-like platform APIs, and Zoho OAuth-like API.
- Compatibility reports are generated from runtime metadata and known-gap mappings.
- Shared deterministic state, idempotency primitives, report hooks, and Go engineering hardening are in place.

## Near Term

- Phase 24: recover observable GitHub Actions execution if compatibility workflow runs are missing.
- Phase 25: expand the SDK/client contract harness beyond provider-specific smoke coverage.
- Phase 26: add versioned compatibility manifests and automated provider-compatible promotion gates.
- Phase 27-29: deepen provider-specific contract evidence for Stripe, OpenAI, GitHub OAuth, Slack, and LINE where applicable.
- Publish post-release artifact, checksum, GHCR, and smoke evidence for `v0.2.0-preview`.

## Public Preview Follow-up

- Expand official SDK contracts beyond the selected workflows currently covered.
- Turn verified compatibility gaps into a small milestone-driven backlog.
- Improve runnable application examples and release adoption evidence.

## Adapter Direction

Current adapters:

- Stripe-like payments.
- OpenAI-compatible API.
- GitHub OAuth-like API.
- Slack-like messaging API.
- LINE-like platform APIs.
- Zoho OAuth-like API.

Candidate adapters after the compatibility foundation:

- SendGrid-like email API.

The broader candidate catalog, priority tiers, adapter-family strategy, and exploratory sequencing live in [Adapter Candidate Priorities](docs/planning/adapter-candidate-priorities.md). That document is planning input, not a commitment or a current-support claim.

## Compatibility Direction

Mockport aims for provider-compatible local APIs for selected workflows. Compatibility is measured by documented endpoint behavior, SDK contract tests, fake state, error shape, and reportable gaps.

Mockport does not reproduce provider internal logic, undocumented behavior, or production network effects.

## Non-Goals

- Proxying real provider traffic.
- Accepting real provider secrets in public examples.
- Claiming full provider compatibility before SDK and workflow contract evidence exists.
- Publishing npm or Homebrew as primary channels before Docker and Go binary release paths are stable.
