# Roadmap

[日本語版](ROADMAP.ja.md)

Mockport is a Docker-first local API environment for AI-native development and CI. This roadmap is intentionally scoped to public preview work and provider-compatible direction without promising full provider internals.

## Current Release

- `v0.3.0-preview`: six workflow-compatible adapters and selected Stripe, OpenAI, and Slack app flows, including fake-key checks and an unmodified LLM 0.36 technical connection. See the [release notes](docs/releases/v0.3.0-preview.md).

## Current Mainline

- Workflow-compatible local adapters for Stripe-like payments, OpenAI-compatible API, GitHub OAuth-like API, Slack-like messaging API, LINE-like platform APIs, and Zoho OAuth-like API.
- Compatibility reports are generated from runtime metadata and known-gap mappings.
- Shared deterministic state, idempotency primitives, report hooks, and Go engineering hardening are in place.

## Near Term

1. P0 complete on main: empty-directory source smoke runs the built checkout image ([#315](https://github.com/albert-einshutoin/mockport/issues/315), [PR #384](https://github.com/albert-einshutoin/mockport/pull/384)); ordinary Stripe/OpenAI requests check fake keys ([#82](https://github.com/albert-einshutoin/mockport/issues/82), [PR #385](https://github.com/albert-einshutoin/mockport/pull/385)); Responses streaming is rejected explicitly ([#386](https://github.com/albert-einshutoin/mockport/issues/386), [PR #387](https://github.com/albert-einshutoin/mockport/pull/387)).
2. P1 is included in `v0.3.0-preview`: [#86](https://github.com/albert-einshutoin/mockport/issues/86) has [Stripe PR #389](https://github.com/albert-einshutoin/mockport/pull/389) and [OpenAI PR #390](https://github.com/albert-einshutoin/mockport/pull/390): Checkout create/retrieve → signed webhook → one business-state update; Python app HTTP request → official SDK non-streaming/Chat streaming → bounded error, retry, and timeout responses. Both run against one Mockport process locally and in CI.
3. The selected [#84](https://github.com/albert-einshutoin/mockport/issues/84) Slack source flow and [#395](https://github.com/albert-einshutoin/mockport/issues/395) LLM 0.36 technical connection are included in `v0.3.0-preview`. A [published-image trial](docs/site/app-trial.md) records the versioned image and sample source. First-time user adoption and continued use remain untested.

## Public Preview Follow-up

- Expand Block Kit, interactions, LINE features, and additional OpenAI APIs only for a demonstrated app workflow; investigate regressions against published contracts sooner when their impact warrants it.
- Implement Responses SSE only after a concrete app need and minimal event contract are recorded. Current main rejects `stream:true` with 501 rather than claiming success.
- Seek a first-time external app trial after the reproducible local/CI flows exist; until then, adoption is an untested hypothesis.

## Adapter Direction

Current adapters:

- Stripe-like payments.
- OpenAI-compatible API.
- GitHub OAuth-like API.
- Slack-like messaging API.
- LINE-like platform APIs.
- Zoho OAuth-like API.

SendGrid-like email remains a candidate after a specific user need, a narrow first mail flow, official SDK/reference evidence, and a maintainable verification plan are documented.

The broader candidate catalog, priority tiers, adapter-family strategy, and exploratory sequencing live in [Adapter Candidate Priorities](docs/planning/adapter-candidate-priorities.md). That document is planning input, not a commitment or a current-support claim.

## Compatibility Direction

Mockport aims for provider-compatible local APIs for selected workflows. Compatibility is measured by documented endpoint behavior, SDK contract tests, fake state, error shape, and reportable gaps.

Mockport does not reproduce provider internal logic, undocumented behavior, or production network effects.

## Non-Goals

- Proxying real provider traffic.
- Accepting real provider secrets in public examples.
- Claiming full provider compatibility before SDK and workflow contract evidence exists.
- Publishing npm or Homebrew as primary channels before Docker and Go binary release paths are stable.
