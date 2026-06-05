# Phase 10 Public Preview Release Implementation Plan

[日本語版](phase10_public_preview_release.ja.md)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Mockport の最初の public preview release を GitHub Release と GHCR image として公開する。

**Architecture:** Phase 6 の scaffold を dry-run から実 publish 証跡へ進める。ただし provider-compatible release とは分け、`v0.1.0-alpha` または `v0.1.0-preview` として current scenario-compatible scope と known gaps を明示する。

**Tech Stack:** GitHub Actions, GHCR, Git tags, shell verification, Docker.

---

## Files

- Create: `CHANGELOG.md`
- Create: `docs/releases/v0.1.0-alpha.md`
- Create: `scripts/verify-release-artifacts.sh`
- Modify: `README.md`
- Modify: `docs/site/distribution.md`
- Modify: `tasks/status.md`

## Task P10-T01: Preview Release Readiness Audit

**Status:** done

- [x] Write `scripts/verify-release-artifacts.sh` to verify release archive names, checksums, binary version, and GHCR image availability.
- [x] Add `CHANGELOG.md` with `v0.1.0-alpha` initial preview release notes.
- [x] Add `docs/releases/v0.1.0-alpha.md` with scope, supported adapters, known limitations, public env safety, support matrix link, and verification commands.
- [x] Run full pre-release verification locally.

## Task P10-T02: Tag And GitHub Preview Release

**Status:** done

- [x] Create annotated tag `v0.1.0-alpha`.
- [x] Push tag and confirm release workflow completes on GitHub.
- [x] Confirm GitHub Release has four archives and `checksums.txt`.
- [x] Run release archive install test from downloaded artifact.

## Task P10-T03: GHCR Preview Publish Verification

**Status:** done

- [x] Confirm GHCR image `ghcr.io/albert-einshutoin/mockport:0.1.0-alpha` exists.
- [x] Confirm `latest` behavior is intentional for preview releases before enabling it.
- [x] Run Docker pull and smoke test from GHCR image, not local build.
- [x] Record image digest in `docs/releases/v0.1.0-alpha.md`.

## Task P10-T04: Preview Install Docs Update

**Status:** done

- [x] Update README install instructions with actual preview release URLs.
- [x] Update docs site distribution page with release artifacts and GHCR commands.
- [x] Keep Homebrew and npm marked not-yet-published unless actually published.
- [x] Run README install path audit from a temporary directory.

## Phase 10 Exit

- [x] Preview tag exists on GitHub.
- [x] GitHub Release contains archives and checksums.
- [x] GHCR preview image can be pulled and run.
- [x] README install commands work from a clean environment.
- [x] Known limitations and non-provider-compatible status are documented.
