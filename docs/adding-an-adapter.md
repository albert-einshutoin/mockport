# Adding a New Adapter

This guide defines how to add a new adapter without drifting from runtime behavior, docs, and evidence.

## 1) Scope and prerequisites

- Review the contracts before coding:
  - [`docs/scenario-policy.md`](scenario-policy.md)
  - [`docs/fixture-policy.md`](fixture-policy.md)
  - [`docs/adapter-helper-policy.md`](adapter-helper-policy.md)
  - [`docs/compatibility-model.md`](compatibility-model.md)
- Decide the supported surface first (paths, status codes, headers, state behavior, known gaps).
- Confirm no real provider credentials and only deterministic local defaults.

## 2) Adapter implementation checklist

### 2.1 Adapter interface

Implement `internal/adapter/Adapter` in `adapters/<name>/adapter.go`.

- `Name() string`
- `Register(*http.ServeMux, adapter.Config) error`
- `FakeEnv(config.AdapterConfig) map[string]string`
- `Metadata() adapter.Metadata`

Keep provider behavior as a local contract, not a full provider clone.

### 2.2 Directory structure

Recommended minimum:

- `adapters/<name>/adapter.go` — Adapter implementation (Register/FakeEnv/Metadata, handlers)
- `adapters/<name>/routes.go` — Optional when handler surface grows
- `adapters/<name>/models.go` — Structured response types
- `adapters/<name>/adapter_test.go` — HTTP scenario coverage
- `docs/adapters/<name>.md` — Public adapter contract

### 2.3 Metadata fields

Fill these fields for every adapter:

- `Name`
- `Maturity` (`experimental` unless evidence supports stronger level)
- `ProviderVersion` / `SDKVersions` / `Levels`
- `Capabilities`
- `Scenarios` and `Endpoints`
- `Endpoints` supported scenario mapping
- `Reset` only when reset behavior exists

If `Reset` is `true`, ensure `internal/state/` and CLI wiring are actually reachable.

## 3) Registration and generated defaults

Add the adapter at all required registration points:

- `internal/cli/builtin.go` (runtime registration list)
- `internal/cli/init.go` (config + env generator defaults)
- `internal/cli/help.go` (metadata/help output)
- `configs/mockport.example.yml` (base local example + at least one scenario definition if needed)
- `docs/site/adapters.md` (public adapter list and links)
- `docs/site/support-matrix.md` (adapter, maturity, endpoints, scenarios, notes)
- `docs/site/reports.md` (behavior matrix if report visibility is affected)
- `examples/<name>/` (minimal runnable example)
- `docs/adapters/<name>.md` (public contract)

## 4) Adapter completeness checks

Use [`scripts/check-adapter-completeness.sh`](scripts/check-adapter-completeness.sh) to verify that every built-in adapter has complete onboarding artifacts.

Minimum pass conditions:

- `docs/adapters/<name>.md` exists
- `docs/site/support-matrix.md` has an adapter row
- `configs/mockport.example.yml` has the adapter section
- `examples/<name>/` exists

If any check fails, fix the missing artifact in the same PR as adapter implementation.

## 5) Acceptance and verification

For adapter PRs, include:

- Failing test first (spec/API/contract test where possible)
- Minimal implementation to pass the test
- Scenario fixtures (where public claims changed)
- Sync of docs + example + support matrix + metadata artifacts
- Verification output in PR description

Required command set:

```bash
go test ./...
go test -race ./...
go test ./adapters/<name>...
```

For adapter changes with compatibility or SDK claims:

```bash
bash scripts/check-public-trust.sh
bash scripts/check-public-env.sh
bash scripts/check-adapter-completeness.sh
bash scripts/run-sdk-contracts.sh all
```

Then include what changed and what risk remains (if any).
