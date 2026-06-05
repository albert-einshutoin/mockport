# Phase 11 Community And Maintenance Implementation Plan

[日本語版](phase11_community_and_maintenance.ja.md)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 公開OSSとして継続メンテできる運用面を整え、外部 contributor と adopter の負担を下げる。

**Architecture:** Repository automation、dependency update、roadmap、triage labels、GitHub Actions runtime upkeep を用意する。Compatibility policy は Phase 12 の compatibility model 後に詳細化するため、この Phase では一般的な maintenance policy に絞る。

**Tech Stack:** GitHub Actions, Dependabot, Markdown, GitHub labels/issues.

---

## Files

- Create: `.github/dependabot.yml`
- Create: `ROADMAP.md`
- Create: `docs/maintainer-guide.md`
- Create: `scripts/check-maintenance-policy.sh`
- Modify: `README.md`
- Modify: `tasks/status.md`

## Task P11-T01: Maintenance Policy

**Status:** done

- [x] Add maintainer guide for release process, issue triage, and security reports.
- [x] Add roadmap with near-term adapters, provider-compatible direction, and non-goals.
- [x] Document no stale auto-close policy for early external users.
- [x] Run maintenance policy check.

## Task P11-T02: Dependency And CI Maintenance

**Status:** done

- [x] Add Dependabot config for GitHub Actions, Go modules, and npm wrapper.
- [x] Update GitHub Actions versions/runtime configuration to remove Node.js 20 deprecation warnings and prefer Node.js 24-compatible actions.
- [x] Add scheduled CI smoke for Docker path if feasible.
- [x] Document that test-only SDK dependencies are pinned in the SDK contract harness phase.
- [x] Run CI static checks locally.

## Task P11-T03: Contribution Quality Bar

**Status:** done

- [x] Document adapter acceptance criteria: metadata, scenarios, examples, report coverage, AI-safe behavior.
- [x] Add checklist for new adapter PRs.
- [x] Reference future compatibility levels without defining final thresholds yet.
- [x] Run public trust and docs checks.

## Phase 11 Exit

- [x] Maintenance policy is documented.
- [x] Dependabot is configured.
- [x] GitHub Actions run without Node.js 20 deprecation warnings, or any unavoidable warning is tracked with a dated upstream reason.
- [x] Adapter contribution quality bar is explicit.
- [x] Public roadmap exists and avoids overpromising.
