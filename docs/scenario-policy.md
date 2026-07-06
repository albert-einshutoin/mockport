# Scenario Policy

[日本語版](scenario-policy.ja.md)

Mockport supports built-in scenarios first. User-defined scenarios are future work and must stay separate from provider compatibility claims until they are promoted through the fixture and compatibility process.

## Built-in Scenarios

Built-in scenarios are maintained by Mockport. They must have:

- A stable scenario name.
- Adapter metadata.
- Tests.
- Public-safe examples.
- Fixture or documentation evidence when used for compatibility scoring.

Built-in scenarios can contribute to compatibility scores only when they are backed by source metadata and visible in reports.

Header override (`X-Mockport-Scenario`) is available for built-in scenarios only; user-defined scenarios are out of scope for per-request switching.

## User-defined Scenarios

User-defined scenarios are local project behavior. They may be useful for app-specific tests, but they do not prove provider compatibility by themselves.

Until a full user-defined scenario system exists, adapters should prefer explicit built-in scenarios over partial custom behavior. If a user-defined scenario is later promoted, it must receive a built-in scenario name, tests, docs, and sanitized fixture evidence.

> **Current status:** The `scenarios:` block in `mockport.yml` is parsed but not yet implemented at runtime. Mockport emits a warning when this block is present. See [limitations](site/limitations.md#unimplemented-configuration-blocks) for details.

## Compatibility Boundary

Compatibility scoring must distinguish:

- Mockport built-in scenario coverage.
- SDK contract coverage.
- Workflow state coverage.
- User-defined local behavior.

Only the first three can raise provider compatibility maturity. User-defined local behavior can be reported, but it must not hide unsupported provider behavior.
