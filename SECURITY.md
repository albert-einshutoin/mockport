# Security Policy

[日本語版](SECURITY.ja.md)

## Supported Versions

Mockport is pre-1.0. Security fixes target the current `main` branch and the latest public preview release when one exists.

## Reporting A Vulnerability

Open a private GitHub security advisory if available. If that is not available, open an issue with only non-sensitive reproduction details and ask for a private disclosure path.

Do not include real secrets, production provider URLs, customer data, tokens, webhook signing secrets, or captured provider payloads in public issues, pull requests, fixtures, screenshots, or logs.

## AI-safe Scope

Mockport is designed to keep local and CI integration tests away from real external providers. AI-safe mode warns about real-looking credentials and real provider URLs, but it is not a secret manager, sandbox boundary, or substitute for repository secret scanning.

## Non-goals

Mockport does not reproduce provider internals, fraud systems, billing networks, real AI inference, or private undocumented behavior.
