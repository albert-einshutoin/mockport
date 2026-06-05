# Fixture Policy

[日本語版](fixture-policy.ja.md)

Mockport fixtures are sanitized compatibility evidence. They help AI-driven and TDD-driven implementation reproduce public API contracts without copying production traffic or leaking credentials.

## Allowed Sources

- Public provider documentation.
- Public OpenAPI or schema references.
- Official SDK behavior observed against local Mockport or a documented sandbox.
- Manually authored examples derived from public docs.

Every fixture must record source type, title, URL or repository path, retrieval date, provider version, and SDK version when relevant.

## Sanitization

Fixtures must normalize:

- Credentials and tokens to the `mockport_`, `local_`, `fake_`, or `dummy_` namespace.
- Customer data to deterministic fake values.
- IDs to deterministic Mockport IDs.
- URLs to localhost or public documentation URLs.
- Timestamps to deterministic values when exact time is not part of the behavior.

Fixtures must not contain real provider secrets, production API URLs, customer payloads, raw webhook signatures, or unredacted recorded traffic.

## Evidence Strength

A fixture is enough for request/response shape, status codes, headers, and documented error examples. SDK contract evidence is required before claiming SDK-compatible or provider-compatible behavior. Workflow-compatible behavior requires state and error coverage in addition to fixtures.
