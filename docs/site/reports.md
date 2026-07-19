# Reports

[日本語版](reports.ja.md)

Mockport exposes a run report at:

```bash
curl http://localhost:43101/_mockport/report
```

The report includes safety status, enabled adapters, request metadata, scenario coverage, behavior matrix entries, and unsupported endpoint attempts. For the full JSON field reference, see [Report schema](report-schema.md).

## Request history

Request history keeps metadata for the most recent 500 requests recorded during a run. When that limit is exceeded, older entries are pruned from the front so the report always returns the newest requests in chronological order. The same bounded history feeds `unsupported_endpoints` in the report payload.

Set `MOCKPORT_REQUEST_HISTORY` to a positive integer to override the default cap. Invalid values (empty, zero, negative, non-numeric, or overflow) fall back to 500.

When history has been truncated, the report includes a `request_history` summary with `limit`, `retained`, `evicted`, and `truncated`. Text reports include a truncation line only when `truncated` is true.

For CLI output:

```bash
mockport report --format text
mockport report --json
```

`mockport report --format json` remains available for backward compatibility. JSON output contains only the report payload, without text headings or ANSI escape sequences.
