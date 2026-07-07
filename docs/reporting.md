# Reporting

[日本語版](reporting.ja.md)

Mockport reports are trust artifacts. They explain what ran, what was safe, what is supported, and what was not supported.

## JSON Report

```bash
mockport report --format json
```

Field-level reference: [Report schema](site/report-schema.md).

Important fields:

- `safety`: AI-safe summary, including real-looking secret and external URL counts.
- `adapters`: enabled adapters, base paths, capabilities, and maturity.
- `requests`: replay-safe request metadata. Request bodies and secret headers are not stored by default.
- `request_history`: bounded request buffer summary with `limit`, `retained`, `evicted`, and `truncated`.
- `scenario_coverage`: supported scenarios per adapter.
- `behavior_matrix`: supported endpoints and their scenarios.
- `compatibility`: measured compatibility level, score, provider version, SDK/client evidence, and unsupported endpoint ids.
- `unsupported_endpoints`: requests that returned unsupported endpoint classifications.

Generated compatibility reports under `docs/compatibility-reports/` add release-facing known gaps for each adapter.

## Text Report

```bash
mockport report --format text
```

Text output is meant for local development and CI logs. JSON output is better for tools.

Request history keeps the most recent 500 requests by default. Override with `MOCKPORT_REQUEST_HISTORY` (positive integer only). When history is truncated, text output includes a truncation line and JSON includes `request_history.truncated=true`.

## Maturity Levels

Allowed adapter maturity values:

```txt
experimental
partial
sdk-compatible
workflow-compatible
provider-compatible
```

Current built-in adapters on the mainline runtime are `workflow-compatible` for selected local workflows: Stripe-like payments, OpenAI-compatible API, GitHub OAuth-like API, Slack-like messaging API, and LINE-like platform APIs.

All adapters are scenario-driven today, not full provider clones. Use the behavior matrix, compatibility section, and unsupported endpoint list as the source of truth for what a test run actually exercised.
