# Adapter Helper Policy

[日本語版](adapter-helper-policy.ja.md)

OpenAI, GitHub OAuth, and Slack currently keep small local helpers such as `writeJSON` and `normalizeScenario` inside each adapter package.

This is intentional for Phase 13:

- The helpers are small and still provider-shaped.
- Response formats differ across adapters.
- Premature shared helpers could hide provider-specific error and header behavior.
- Future adapter work should first add regression tests for response shape, headers, status codes, and scenario defaults.

Shared helpers may be introduced when at least one of these is true:

- Four or more adapters repeat the same helper with identical behavior.
- A shared helper preserves provider-specific response shape through regression tests.
- The helper belongs to infrastructure, such as safe JSON writing or report metadata, rather than provider semantics.

Until then, adapters should prefer clear local helpers over broad abstraction.

## Tracking duplicated helper names

[`scripts/check-adapter-helpers.sh`](../scripts/check-adapter-helpers.sh) scans built-in adapter packages for repeated unexported helper names such as `writeJSON` and `normalizeScenario`.

The script is a tracking aid, not a mandate to consolidate immediately:

- Name duplication does not prove identical behavior across adapters.
- Provider-specific response shape, headers, status codes, and scenario defaults still require adapter regression tests before any shared helper is introduced.
- The script reports duplicates and exits successfully by default.

It fails only when one helper name appears in more adapter packages than `DUPLICATE_ADAPTER_THRESHOLD`. The script uses a strict `>` comparison, so a threshold of `3` fails when the same helper name appears in four or more adapter packages.

### `DUPLICATE_ADAPTER_THRESHOLD` tuning

`DUPLICATE_ADAPTER_THRESHOLD` controls when duplicate helper names stop being informational and start failing CI. It does not decide whether a helper should be shared; duplicate names are a tracking signal only.

| Mode | Value | Effect |
| --- | --- | --- |
| Normal tracking | unset (default) | Defaults to the current built-in adapter package count (6 today). Duplicates are reported and CI passes. |
| Stricter consolidation review | `3` | CI fails when one helper name appears in four or more adapter packages. This matches the policy above that four or more adapters with identical behavior should trigger a consolidation review. |

Use normal tracking for day-to-day adapter work. Set `DUPLICATE_ADAPTER_THRESHOLD=3` only when you want CI to enforce the four-adapter consolidation review gate.

Even when the threshold is exceeded, duplicate helper names alone do not justify a shared abstraction. Consolidation still requires identical behavior across adapters and regression tests for provider-specific response shape, headers, status codes, and scenario defaults.
