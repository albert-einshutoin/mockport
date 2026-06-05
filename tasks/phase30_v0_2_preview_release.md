# Phase 30 v0.2.0-preview Release Implementation Plan

[日本語版](phase30_v0_2_preview_release.ja.md)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Phase 23-29 の成果を `v0.2.0-preview` として公開し、Docker/GitHub Release/compatibility report の証跡付きで導入前PoCに使えるOSS状態へ進める。

**Architecture:** Release は機能完了宣言ではなく、測定可能な compatibility evidence と known gaps を含む preview とする。release artifacts、GHCR image、checksums、docs、changelog、post-release smoke を一連の手順で検証する。

**Tech Stack:** GitHub Release, GHCR, Go release archives, Docker, compatibility report, docs.

---

## Files

- Create: `docs/releases/v0.2.0-preview.md`
- Modify: `CHANGELOG.md`
- Modify: `README.md`
- Modify: `docs/site/index.md`
- Modify: `docs/site/distribution.md`
- Modify: `docs/compatibility-reports/latest.md`
- Modify: `tasks/status.md`

## Task P30-T01: Release Readiness Gate

**Status:** pending

- [ ] Confirm Phase 23-29 are `done` in `tasks/status.md`.
- [ ] Run `go test ./...`, `go vet ./...`, and `go build ./cmd/mockport`.
- [ ] Run `bash scripts/run-sdk-contracts.sh all`.
- [ ] Run `bash scripts/check-compatibility-release.sh`, `bash scripts/check-distribution.sh`, and `bash scripts/check-public-trust.sh`.

## Task P30-T02: Release Notes

**Status:** pending

- [ ] Create `docs/releases/v0.2.0-preview.md` with supported services, maturity labels, compatibility scores, known gaps, and verification commands.
- [ ] Update `CHANGELOG.md` with the release date, major changes, compatibility report link, and migration notes from `v0.1.0-alpha`.
- [ ] Update README badges/links only if release artifacts are actually published.
- [ ] Keep "not full provider internals" and "no real provider traffic" explicit.

## Task P30-T03: Artifact Publication

**Status:** pending

- [ ] Create tag `v0.2.0-preview` only after readiness checks pass.
- [ ] Confirm release workflow runs and uploads Linux/macOS archives plus checksums.
- [ ] Confirm Docker/GHCR image is published with `v0.2.0-preview` tag if Docker workflow is configured for tags.
- [ ] If a workflow does not run, stop and resolve Phase 24-style Actions issue before announcing release.

## Task P30-T04: Post-release Smoke

**Status:** pending

- [ ] Download release archive into a temp directory and run `mockport version`.
- [ ] Run Docker image from GHCR and verify `GET /health`.
- [ ] Run at least one Stripe, OpenAI, GitHub OAuth, and Slack smoke request against the release artifact.
- [ ] Record commands and results in `docs/releases/v0.2.0-preview.md`.

## Phase 30 Exit

- [ ] `v0.2.0-preview` GitHub Release exists with checksums and release notes.
- [ ] GHCR image is available or the reason it is unavailable is documented.
- [ ] Compatibility report is linked from release notes.
- [ ] Post-release install and smoke checks pass from a clean environment.
