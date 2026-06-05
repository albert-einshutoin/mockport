# Phase 23 Roadmap And Documentation Alignment Implementation Plan

[日本語版](phase23_roadmap_docs_alignment.ja.md)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Phase 22 完了後の実態と公開 docs / roadmap / release notes のズレをなくし、次の contributor が現在地と次の優先順位を誤読しない状態にする。

**Architecture:** `tasks/status.md` を正本として、`ROADMAP.md`、`README.md`、`docs/site/*`、`CHANGELOG.md`、compatibility report の説明を同期する。新機能は追加せず、公開文書の期待値を実装済みの evidence に合わせる。

**Tech Stack:** Markdown docs, shell static checks, existing doc link checker.

---

## Files

- Modify: `ROADMAP.md`
- Modify: `README.md`
- Modify: `docs/site/index.md`
- Modify: `docs/site/support-matrix.md`
- Modify: `docs/site/limitations.md`
- Modify: `CHANGELOG.md`
- Modify: `tasks/status.md`

## Task P23-T01: Current State Audit

**Status:** pending

- [ ] Run `git status --short` and confirm the worktree is clean before editing.
- [ ] Read `tasks/status.md`, `docs/compatibility-reports/latest.md`, `docs/site/support-matrix.md`, and `ROADMAP.md`.
- [ ] Write down mismatches in `tasks/status.md` under Phase 23 verification notes before fixing them.
- [ ] Confirm Phase 0-22 are `done` and Phase 23-30 are the only `pending` future phases.

## Task P23-T02: Roadmap Refresh

**Status:** pending

- [ ] Replace stale Near Term entries in `ROADMAP.md` that still mention Phase 12-16 as future work.
- [ ] Add a new "Current State" section describing Phase 22 completion: workflow-compatible adapters, compatibility report, and Docker-first runtime.
- [ ] Add a "Next Work" section listing Phase 24-30 in priority order.
- [ ] Keep non-goals explicit: no provider internals, no undocumented behavior, no real secret/proxy mode.

## Task P23-T03: Public Docs Consistency

**Status:** pending

- [ ] Update `README.md` so service maturity, compatibility report links, and release status match Phase 22.
- [ ] Update `docs/site/index.md` to point first-time users to support matrix and compatibility report.
- [ ] Update `docs/site/limitations.md` so known gaps match `docs/compatibility-reports/latest.md`.
- [ ] Update `docs/site/support-matrix.md` only if wording diverges from generated compatibility reports.

## Task P23-T04: Documentation Verification

**Status:** pending

- [ ] Run `bash scripts/check-doc-links.sh`.
- [ ] Run `bash scripts/check-public-trust.sh`.
- [ ] Run `bash scripts/check-compatibility-release.sh`.
- [ ] Run `rg -n "Phase 12|Phase 13|Phase 14|Phase 15|Phase 16" ROADMAP.md README.md docs` and verify any hits are historical, not future-looking.

## Phase 23 Exit

- [ ] Roadmap reflects Phase 22 completion and Phase 23-30 future direction.
- [ ] Public docs do not claim stale adapter maturity or stale near-term phases.
- [ ] Compatibility report and support matrix tell the same story.
- [ ] Documentation checks pass locally.
