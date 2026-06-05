# Compatibility Fixtures

[日本語版](README.ja.md)

Compatibility fixtures are sanitized evidence used to implement and review provider-compatible behavior. They are not recordings of production traffic.

## Required Fields

Each fixture must include:

- `provider`: adapter/provider name, such as `stripe`.
- `provider_version`: provider API or documentation version used for the fixture.
- `sdk.name` and `sdk.version`: SDK used for contract evidence, or `none` when no SDK was used.
- `source.type`: `docs`, `openapi`, `sdk`, or `manual`.
- `source.title`: human-readable source name.
- `source.url_or_path`: public docs/spec URL or repository path.
- `source.retrieved_at`: date the source was checked.
- `scenario`: built-in scenario name when the fixture backs built-in behavior.
- `request`: method, path, headers, and body.
- `response`: status, headers, and body.
- `normalization`: fields changed during sanitization.

## Safety Rules

- Use only fake credentials with the `mockport_`, `local_`, `fake_`, or `dummy_` namespace.
- Do not include real provider secrets, customer payloads, webhook signatures, or production API URLs.
- Do not include raw provider traffic unless it has been reduced and sanitized into the fixture format.
- Keep source metadata even when the fixture is manually authored from public docs.

## Scenario Policy

Fixtures may support built-in scenarios or future user-defined scenarios. Built-in scenarios are maintained by Mockport and can be used for compatibility scoring. User-defined scenarios are local project behavior and must not raise provider compatibility scores unless they are promoted into a documented built-in scenario with source evidence.
