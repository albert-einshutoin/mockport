# Adapter Helper Policy

[日本語版](adapter-helper-policy.ja.md)

The six current built-in adapters—Stripe, OpenAI, GitHub OAuth, Slack, LINE, and Zoho OAuth—keep provider-shaped helpers local to their adapter packages. Examples shared by some packages include `bearerToken`, `clientIDMatches`, and `writeError`; the duplicated names do not imply identical behavior.

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

[`scripts/check-adapter-helpers.sh`](../scripts/check-adapter-helpers.sh) scans built-in adapter packages for repeated unexported helper names such as `bearerToken`, `clientIDMatches`, and `writeError`.

The script is a tracking aid, not a mandate to consolidate immediately:

- Duplicate helper names are evidence for review, not proof that adapters should be refactored now.
- Name duplication does not prove identical behavior across adapters.
- Provider-specific response shape, headers, status codes, and scenario defaults still require adapter regression tests before any shared helper is introduced. Do not hide provider-specific behavior behind a shared helper without those tests.
- The script reports duplicates and exits successfully by default.

### Sample output

Each duplicated helper name is printed on its own line. A summary line follows, then a final pass or fail line:

```text
duplicate helper: bearerToken (2 adapters: githuboauth line)
duplicate helper: clientIDMatches (3 adapters: githuboauth line zohooauth)
duplicate helper: decodePayload (2 adapters: line openai)
duplicate helper: firstNonEmpty (2 adapters: line zohooauth)
duplicate helper: newWithWebhookTimeout (2 adapters: line stripe)
duplicate helper: redirectWithQuery (3 adapters: githuboauth line zohooauth)
duplicate helper: writeError (2 adapters: openai zohooauth)
duplicate helper: writeOAuthError (2 adapters: githuboauth line)
check-adapter-helpers: 8 duplicated helper name(s) tracked (threshold=6)
check-adapter-helpers passed
```

When a helper appears in more adapter packages than `DUPLICATE_ADAPTER_THRESHOLD`, the script prints a threshold warning on stderr and exits with status 1:

```text
duplicate helper: clientIDMatches (3 adapters: githuboauth line zohooauth)
  exceeds DUPLICATE_ADAPTER_THRESHOLD=2
...
duplicate helper: redirectWithQuery (3 adapters: githuboauth line zohooauth)
  exceeds DUPLICATE_ADAPTER_THRESHOLD=2
...
check-adapter-helpers: 8 duplicated helper name(s) tracked (threshold=2)
check-adapter-helpers failed: one or more helpers exceed DUPLICATE_ADAPTER_THRESHOLD=2
```

### `DUPLICATE_ADAPTER_THRESHOLD`

The script fails only when one helper name appears in more adapter packages than `DUPLICATE_ADAPTER_THRESHOLD`. The default equals the current built-in adapter package count, so routine duplicates are reported without blocking CI.

- Leave the default in place for normal development and CI.
- Lower the threshold when you want the check to fail sooner and surface broader duplication during review.
- Raise the threshold only temporarily, with a documented reason, when broader duplication is expected and you still want the script to pass.

Example:

```bash
DUPLICATE_ADAPTER_THRESHOLD=4 bash scripts/check-adapter-helpers.sh
```
