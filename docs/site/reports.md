# Reports

[日本語版](reports.ja.md)

Mockport exposes a run report at:

```bash
curl http://localhost:43101/_mockport/report
```

The report includes safety status, enabled adapters, request metadata, scenario coverage, behavior matrix entries, and unsupported endpoint attempts.

## JSON payload fields

`/_mockport/report` and `mockport report --format json` return the same JSON object. Top-level fields:

| Field | Description |
| --- | --- |
| `mode` | Configured Mockport run mode (for example `ai-safe` or `strict`). |
| `safety` | AI-safe summary: run mode, overall safe status, counts of real-looking secrets and external URLs, and public-env safety. |
| `adapters` | Enabled adapters with base paths, scenarios, maturity, and capabilities. |
| `requests` | Replay-safe request metadata for handled traffic. Bodies and secret headers are not stored. |
| `safety_warnings` | Individual safety warnings that contributed to the `safety` summary. |
| `scenario_coverage` | Supported scenarios per adapter. |
| `behavior_matrix` | Supported endpoints and the scenarios that exercise them. |
| `compatibility` | Measured compatibility level, score, provider version, SDK/client evidence, and known gaps. Omitted when empty. |
| `state_coverage` | Stateful resource coverage, idempotency, and reset support per adapter. Omitted when empty. |
| `unsupported_endpoints` | Requests that returned unsupported-endpoint classifications, derived from `requests`. |

## Request history

Request history keeps metadata for the most recent 1000 requests recorded during a run. When that limit is exceeded, older entries are pruned from the front so the report always returns the newest requests in chronological order. The same bounded history feeds `unsupported_endpoints` in the report payload.

For CLI output:

```bash
mockport report --format text
mockport report --format json
```
