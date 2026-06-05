# Phase 24 GitHub Actions Execution Recovery Implementation Plan

[日本語版](phase24_github_actions_execution_recovery.ja.md)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** `git push` 後に GitHub Actions run が作成されない状態を調査し、CI / smoke / compatibility workflow が実際に実行される公開OSS運用に戻す。

**Architecture:** ローカル推測ではなく、GitHub repository settings、workflow file state、default branch、Actions permissions、workflow trigger events、GitHub API results を evidence として記録する。原因を直せる場合は workflow/settings を修正し、設定権限が必要な場合は明確な手順を docs に残す。

**Tech Stack:** GitHub CLI, GitHub Actions YAML, shell checks, docs.

---

## Files

- Modify: `.github/workflows/ci.yml`
- Modify: `.github/workflows/compatibility.yml`
- Modify: `.github/workflows/smoke.yml`
- Modify: `docs/maintainer-guide.md`
- Modify: `tasks/status.md`
- Create: `docs/operations/github-actions-troubleshooting.md`

## Task P24-T01: Evidence Collection

**Status:** pending

- [ ] Run `gh repo view --json nameWithOwner,defaultBranchRef,visibility`.
- [ ] Run `gh api repos/albert-einshutoin/mockport/actions/permissions`.
- [ ] Run `gh workflow list --all`.
- [ ] Run `gh run list --limit 20 --json databaseId,workflowName,headSha,event,status,conclusion,createdAt`.
- [ ] Record the exact output summary in `docs/operations/github-actions-troubleshooting.md`.

## Task P24-T02: Trigger And Workflow Static Audit

**Status:** pending

- [ ] Verify `.github/workflows/ci.yml` includes `on: push` and `on: pull_request`.
- [ ] Verify `.github/workflows/compatibility.yml` includes `workflow_dispatch` and `schedule`.
- [ ] Verify workflow filenames are committed on remote `main` with `git ls-remote origin refs/heads/main` and `git show origin/main:.github/workflows/ci.yml`.
- [ ] Check whether YAML parsing is valid by running `ruby -e 'require "yaml"; Dir[".github/workflows/*.yml"].each { |f| YAML.load_file(f); puts f }'` or an equivalent local parser.

## Task P24-T03: Recovery Implementation

**Status:** pending

- [ ] If Actions are disabled at repository/org level, document the exact setting path and required owner action.
- [ ] If workflow files are disabled or not recognized, re-enable with `gh workflow enable <workflow>` where possible.
- [ ] If branch protection or default branch mismatch is the issue, update docs and workflow trigger expectations.
- [ ] If YAML shape is the issue, fix workflow YAML and add a static check to `scripts/check-maintenance-policy.sh`.

## Task P24-T04: Prove Actions Run

**Status:** pending

- [ ] Trigger CI with a harmless docs commit or `gh workflow run ci.yml --ref main`.
- [ ] Trigger compatibility with `gh workflow run compatibility.yml --ref main`.
- [ ] Run `gh run watch <run-id> --exit-status` for at least one CI run and one compatibility run.
- [ ] Update `docs/operations/github-actions-troubleshooting.md` with the successful run URLs.

## Phase 24 Exit

- [ ] `gh run list --commit <latest-main-sha>` shows workflow runs or the documented reason why push-triggered runs are impossible.
- [ ] Manual `workflow_dispatch` works for compatibility workflow.
- [ ] Maintainer docs explain how to diagnose a future no-run state.
- [ ] Maintenance checks cover any workflow syntax or policy issue found.
