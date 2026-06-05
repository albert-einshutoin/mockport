# Comparison

[日本語版](comparison.ja.md)

## Mockport vs Provider Sandboxes

Provider sandboxes are authoritative for provider behavior. Mockport is local, Docker-first, secret-free, and deterministic. Use Mockport for fast local and CI integration tests, then validate critical paths against the real provider sandbox before production.

## Mockport vs WireMock

WireMock is a general HTTP mocking tool. Mockport is provider-shaped: adapters know about service workflows, fake credentials, safety warnings, reports, and support matrices.

## Mockport vs Hand-written Test Doubles

Hand-written doubles are quick for one codebase but drift quickly. Mockport centralizes adapter behavior, examples, reports, and compatibility evidence so multiple apps and CI jobs can share the same local provider API.

## Positioning

Mockport is best when:

- You want Docker-first local APIs.
- You want fake env values that are safe to commit.
- You want deterministic external-service scenarios in CI.
- You want adapter coverage and unsupported behavior reported explicitly.

Mockport is not a full clone of provider internals.
