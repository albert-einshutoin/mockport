# Compatibility Report

[日本語版](latest.ja.md)

Generated: 2026-06-06

Compatibility is measured from Mockport runtime metadata, SDK/client contract checks, fixture coverage, and known gaps. It is not a claim that provider internals or undocumented behavior are reproduced.

## Scores

| Adapter | Maturity | Score | Provider API | SDK/client evidence |
| --- | --- | ---: | --- | --- |
| `github-oauth` | `workflow-compatible` | 100 | 2022-11-28 | oauth-client-contract |
| `line` | `workflow-compatible` | 80 | Messaging API v2 / Login v2.1 / Pay v3 / MINI App service messages / Mini Dapp SDK | none |
| `openai` | `workflow-compatible` | 100 | 2025-02-01 | openai@6.39.1 |
| `slack` | `workflow-compatible` | 100 | 2025-02-01 | slack-client-contract |
| `stripe` | `workflow-compatible` | 100 | 2025-10-29.clover | stripe@22.2.0 |

## Coverage

| Adapter | Endpoint | Scenario | SDK/client | State | Error |
| --- | ---: | ---: | ---: | ---: | ---: |
| `github-oauth` | 100 | 100 | 100 | 100 | 100 |
| `line` | 100 | 100 | 0 | 100 | 100 |
| `openai` | 100 | 100 | 100 | 100 | 100 |
| `slack` | 100 | 100 | 100 | 100 | 100 |
| `stripe` | 100 | 100 | 100 | 100 | 100 |

## Known Gaps

### github-oauth
- No real GitHub policy, repository permissions, SSO, org/enterprise enforcement, or app installation model.

### line
- No official LINE SDK contract yet, no real LIFF browser runtime, no provider-driven webhook redelivery, no monthly quota/rate bucket enforcement, no complete Messaging API schema validation, no regional policy enforcement, and Mini Dapp endpoints are local SDK helpers rather than a full Dapp Portal clone.

### openai
- No real model quality, tokenization parity, hosted tools, vector stores, or provider scheduling.

### slack
- No real delivery, Events API completeness, Block Kit validation, files, app scopes, enterprise policy, or full workspace directory.

### stripe
- No fraud, payment network, tax, disputes, Connect, or full Billing lifecycle.

## Release Labels

- `experimental`: Early adapter coverage for selected workflows. Expect gaps.
- `sdk-compatible`: Selected SDK or client contract calls pass against local Mockport.
- `workflow-compatible`: Selected workflows include fake state, errors, and replayable behavior.
- `provider-compatible`: Selected provider workflows are backed by manifests, SDK/client contracts, fixtures, scores, contract evidence, and known-gap reports.
