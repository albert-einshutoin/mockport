# Contributing

[日本語版](CONTRIBUTING.ja.md)

## Setup

Use Go 1.26.4. Install a Go toolchain (mise, asdf, Homebrew, or the official installer) so `go` is available on your PATH.

```bash
go test ./...
go vet ./...
go build ./cmd/mockport
```

## Spec-First TDD

Production changes must follow spec-first TDD. Start from the written Mockport contract, then prove the gap with a failing test before changing production code.

1. Update the relevant spec first. Adapter behavior belongs in `docs/adapters/<adapter>.md`; compatibility evidence belongs in fixtures, manifests, SDK/client contracts, and reports.
2. Write the failing test for the smallest useful slice.
3. Run the narrow test and confirm it fails for the expected reason.
4. Implement the smallest change that passes.
5. Run the narrow test again.
6. Update metadata, docs, fixtures, reports, and known gaps so public claims match runtime behavior.
7. Run the full verification for the touched Phase.

Do not widen provider surface just to make an example pass. Mockport supports selected deterministic local workflows, not full provider internals.

### New adapter onboarding

For adapter PRs, follow the onboarding guide before implementation:

- [`docs/adding-an-adapter.md`](docs/adding-an-adapter.md)
- [`docs/scenario-policy.md`](docs/scenario-policy.md)
- [`docs/compatibility-model.md`](docs/compatibility-model.md)

## Public Trust Checks

Run these before opening a pull request:

```bash
bash scripts/check-public-trust.sh
bash scripts/check-distribution.sh
go test ./...
go test -race ./...
go vet ./...
```

Race tests are expected for adapter changes or shared state modifications.

## Adapter Changes

Adapter acceptance criteria:

Adapter pull requests must include:

- Spec-first TDD evidence: spec change, failing regression or contract test, implementation, and final verification.
- HTTP tests for success and failure scenarios.
- Metadata/report coverage that matches adapter `Metadata()`, `mockport add <adapter>`, and `mockport help <service>`.
- Example config or docs when user-facing.
- AI-safe behavior for fake credentials and local base URLs.
- Clear unsupported behavior in reports or docs.
- No real provider secrets, production URLs, customer payloads, or unsanitized fixtures.

## Pull Requests

Include test evidence in the PR body. Do not include real provider secrets, production URLs, customer payloads, or unsanitized fixtures.
