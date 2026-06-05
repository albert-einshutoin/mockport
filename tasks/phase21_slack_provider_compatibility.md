# Phase 21 Slack Provider Compatibility Implementation Plan

[日本語版](phase21_slack_provider_compatibility.ja.md)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Slack-like adapter を Slack client/workflow-compatible local API へ引き上げる。

**Architecture:** Slack messaging workflow、conversation state、message lifecycle、basic events を fake workspace 上で再現する。Slack の enterprise policy、real delivery、workspace directory 全体は再現しない。

**Tech Stack:** Go Slack adapter, Slack SDK/client contract tests, sanitized fixtures, compatibility manifests.

---

## Files

- Create: `compat/fixtures/slack/*`
- Create: `contract/sdk/slack-smoke.test.js`
- Modify: `adapters/slack/*`
- Modify: `docs/site/support-matrix.md`
- Modify: `tasks/status.md`

## Task P21-T01: Slack Client Contract Baseline

**Status:** done

- [x] Write failing client smoke for auth.test.
- [x] Write failing client smoke for chat.postMessage.
- [x] Write failing client smoke for conversations list/history if client workflow needs it.
- [x] Run `bash scripts/run-sdk-contracts.sh slack`.

## Task P21-T02: Slack Messaging And Conversation State

**Status:** done

- [x] Persist channels, users, messages, timestamps, and bot identity.
- [x] Add message update/delete and conversation history subset.
- [x] Add event callback subset with signature verification where useful.
- [x] Update support matrix and compatibility score.

## Task P21-T03: Slack Error And Rate Limit Fidelity

**Status:** done

- [x] Add fixtures for invalid_auth, channel_not_found, not_in_channel, rate_limited, delivery_failed.
- [x] Implement response shape and retry headers close to Slack public API.
- [x] Report unsupported methods and scopes.
- [x] Run adapter tests, client contracts, and compatibility report.

## Phase 21 Exit

- [x] Slack adapter is at least `workflow-compatible`.
- [x] Slack client contract passes for supported workflows.
- [x] Messaging state and errors are documented by fixtures.
- [x] Known gaps are explicit.
