# Maintainer Guide

[日本語版](maintainer-guide.ja.md)

This guide keeps the public OSS workflow explicit while Mockport is in preview.

## Release Process

1. Confirm `tasks/status.md` has the current phase ready for release.
2. Run local verification:

```bash
bash scripts/check-public-trust.sh
bash scripts/check-public-env.sh
bash scripts/check-distribution.sh
bash scripts/check-maintenance-policy.sh
/usr/local/go/bin/go test ./...
/usr/local/go/bin/go vet ./...
```

3. Build and verify release archives:

```bash
tmpdir="$(mktemp -d)"
scripts/build-release-archives.sh 0.1.0-alpha "$tmpdir"
scripts/verify-release-artifacts.sh 0.1.0-alpha "$tmpdir"
```

4. Sync the public preview image tag everywhere it appears before tagging (see [Release version update checklist](#release-version-update-checklist) below).

5. Create an annotated tag and let GitHub Actions publish release archives and GHCR images.
6. Download the GitHub Release artifacts and run `scripts/verify-release-artifacts.sh`.
7. Pull the GHCR image and run Docker smoke from the published image.
8. Record release URL, image tag, digest, known limitations, and verification evidence in `docs/releases/`.

## Release Version Update Checklist

On each release, update `0.1.0-alpha` (or the new semver) consistently in every file below. Run:

```bash
grep -rln "0.1.0-alpha" --include="*.js" --include="*.json" --include="*.md" --include="*.yml" . \
  | grep -v ".git/" | grep -v docs/releases/
```

| File | Role |
| --- | --- |
| `README.md` | 30-second quickstart `docker run` and distribution mentions |
| `README.ja.md` | Japanese quickstart and distribution mentions |
| `docs/site/quickstart.md` | Option A install command |
| `docs/site/distribution.md` | Docker pull/run, archive download URLs, verification examples |
| `packaging/npm/bin/mockport.js` | `MOCKPORT_IMAGE` default when env var is unset |
| `packaging/npm/test/wrapper.test.js` | Asserts npm wrapper default image tag |
| `packaging/npm/README.md` | Documents Docker fallback image tag |
| `packaging/npm/README.ja.md` | Japanese Docker fallback image tag |
| `internal/cli/init.go` | `defaultDockerImage` in generated `docker-compose.mockport.yml` |
| `internal/cli/init_test.go` | Asserts generated compose pins the preview image |
| `scripts/check-distribution.sh` | CI guard that npm wrapper default matches preview tag |
| `scripts/verify-release-artifacts.sh` | Default `VERSION` argument in usage examples |
| `CHANGELOG.md` | Release notes entry for the new tag |
| `ROADMAP.md` | Current preview version mention |
| `.github/workflows/docker.yml` | Workflow input description example tag |

Also update release-process command examples in this guide (`build-release-archives.sh`, `verify-release-artifacts.sh`) and add a new `docs/releases/vX.Y.Z.md` entry. Historical task notes under `tasks/phase10_public_preview_release.md` and `tasks/status.md` are records only; update them only when documenting a milestone change.

## Issue Triage

- Do not auto-close stale issues while Mockport is in early public preview.
- Ask for adapter name, Mockport version or commit, sanitized config, expected behavior, actual behavior, and test evidence.
- Remove or redact any real provider secret, production URL, customer payload, or unsanitized fixture.
- Treat safety issues involving leaked credentials or unsafe examples as security-sensitive until ruled out.

## Security Reports

Follow `SECURITY.md`. Do not request real secrets or production provider traffic to reproduce a report.

## Spec-First TDD Development

Mockport development must start from a written contract, not from an implementation shortcut. For adapter work, the human-facing contract lives in `docs/adapters/<adapter>.md`; compatibility evidence lives in `compat/fixtures/<adapter>/`, `compat/manifests/`, SDK/client contract tests, and generated compatibility reports.

Use this loop for new behavior and bug fixes:

1. Define the selected public provider surface in the adapter spec. Include supported paths, request/response shape, status codes, headers, state behavior, scenarios, known gaps, and explicit non-goals.
2. Add or update sanitized fixtures when public documentation, official schema examples, or SDK/client behavior is needed to explain the expected contract.
3. Write the failing test first. Prefer the narrowest useful RED test: adapter package tests for provider-shaped responses, server tests for routing and mounting, CLI tests for discovery/help behavior, SDK contract tests for client compatibility, and report tests for compatibility metadata.
4. Run the targeted test and confirm it fails for the expected reason before changing production code.
5. Implement the smallest provider-scoped behavior needed to pass the test. Do not add broad provider surface only to make examples pass.
6. Refactor only while keeping the new test and existing package tests green.
7. Synchronize the public claim: adapter `Metadata()`, `mockport add <adapter>`, `mockport help <service>`, docs, fixtures, manifests, support matrix, and generated compatibility reports must describe the same surface.
8. Run the relevant package tests, then the repo gates listed in the release process before merging.

For the full onboarding checklist, see [`docs/adding-an-adapter.md`](adding-an-adapter.md).

Recommended TDD command sequence:

```bash
/usr/local/go/bin/go test ./adapters/<adapter>
/usr/local/go/bin/go test ./internal/server ./internal/cli ./internal/config
/usr/local/go/bin/go test ./...
/usr/local/go/bin/go vet ./...
go test -race ./...
bash scripts/check-go-engineering.sh
```

When SDK/client compatibility is part of the claim, also run the relevant checks under `contract/sdk` and regenerate compatibility reports before updating checked-in report artifacts.

## Dependency And CI Maintenance

- Dependabot covers GitHub Actions, Go modules, the experimental npm wrapper, and test-only SDK contracts.
- GitHub Actions should use Node.js 24-compatible action releases where available.
- If a Node.js 20 deprecation warning remains because an upstream action has no Node.js 24-compatible release, document the action, warning date, and upstream tracking reason in the relevant phase notes.
- Scheduled smoke checks should validate the Docker-first path without publishing images.

Test-only SDK dependencies are pinned in `contract/sdk` and must stay out of the Go runtime.

## Adapter Contribution Quality Bar

Adapter pull requests must include:

- A spec-first TDD trail: adapter spec change, failing regression or contract test, implementation, and final verification evidence.
- Adapter metadata: name, base path, maturity, capabilities, endpoints, and scenarios.
- HTTP tests for success, auth error, rate limit, and at least one meaningful failure scenario.
- Example config and user-facing docs when the adapter is exposed publicly.
- Report coverage for supported and unsupported behavior.
- AI-safe fake credentials and local-only URLs.
- No real provider secrets, production URLs, customer payloads, or unsanitized fixtures.

Keep provider-shaped helpers local until `docs/adapter-helper-policy.md` says the abstraction threshold has been met.

Future compatibility levels will be defined in Phase 14. Until then, adapter claims must stay at scenario-compatible or experimental unless stronger evidence exists.

## Mockport Development Invariants

- Adapter specs are Mockport contracts, not copies of provider documentation. Rewrite official behavior into the selected local surface and keep provider internals, undocumented behavior, console policy, billing networks, delivery guarantees, fraud checks, and real model quality out of scope.
- Fake state must be deterministic, local-only, restart-resettable unless explicitly documented, and safe under concurrent HTTP handlers. `go test -race` is required, but a green race detector does not prove workflow invariants; add tests for ordering-sensitive state transitions when set/delete/default or create/retrieve/list behavior can interleave.
- Public examples, fixtures, configs, and tests must use AI-safe fake credentials, localhost URLs, deterministic IDs, and sanitized payloads only.
- Compatibility maturity must never be raised by docs alone. Promotion requires matching implementation, metadata, tests, fixtures or SDK evidence, known-gap documentation, and generated reports.
- Runtime registration and docs must not drift. `internal/cli/builtin.go`, adapter `Metadata()`, generated config/env output, service help, support matrix, and compatibility reports should be checked together whenever an adapter surface changes.
- Keep abstractions provider-shaped until duplication is proven. Shared helpers are acceptable only when they preserve provider-specific response shape, headers, errors, and scenario defaults through regression tests.
