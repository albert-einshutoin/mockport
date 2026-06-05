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
