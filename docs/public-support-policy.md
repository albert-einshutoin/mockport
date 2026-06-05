# Public Support Policy

[日本語版](public-support-policy.ja.md)

Mockport public preview support is best-effort.

## Compatibility Terms

- `scenario-compatible`: deterministic local responses cover selected success and failure scenarios.
- `sdk-compatible`: an official SDK or common client works against Mockport for selected workflows.
- `workflow-compatible`: create, retrieve, list, update, error, and state behavior work for selected workflows.
- `provider-compatible`: measured compatibility for selected provider workflows, backed by manifests, SDK contracts, fixtures, and known-gap reports.

Mockport targets provider-compatible local APIs for selected workflows. It does not reproduce provider internals or undocumented behavior.

## Public Issue Scope

Good public issues include:

- Adapter name and scenario.
- Mockport version or commit.
- Redacted config.
- Expected behavior.
- Actual behavior.
- Reproduction command.

Do not include real secrets, production URLs, customer data, or unsanitized provider payloads.
