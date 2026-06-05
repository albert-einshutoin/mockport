# Phase 29 GitHub OAuth And Slack Client Evidence Implementation Plan

[日本語版](phase29_oauth_slack_client_evidence.ja.md)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** GitHub OAuth と Slack の client evidence を強化し、SDK/client contract の弱さによる compatibility score の不透明さを解消する。

**Architecture:** GitHub OAuth は標準OAuth client flowとGitHub REST subsetを強化する。Slack は可能なら公式 `@slack/web-api` を契約に採用し、採用しない場合も理由と代替client evidenceをmanifestに記録する。

**Tech Stack:** Node.js 24 contract runner, OAuth HTTP client, optional Slack Web API SDK, Go adapters, manifests.

---

## Files

- Modify: `contract/sdk/github-oauth-smoke.test.js`
- Modify: `contract/sdk/slack-smoke.test.js`
- Modify: `contract/sdk/package.json`
- Modify: `contract/sdk/package-lock.json`
- Modify: `adapters/githuboauth/*`
- Modify: `adapters/slack/*`
- Modify: `compat/manifests/github-oauth.json`
- Modify: `compat/manifests/slack.json`
- Modify: `tasks/status.md`

## Task P29-T01: GitHub OAuth Evidence Strengthening

**Status:** pending

- [ ] Add client contract assertions for `state`, redirect URI mismatch, invalid code, missing scope, and bad credentials.
- [ ] Add REST subset contract for `/user`, `/user/emails`, and `/user/orgs` using Bearer token.
- [ ] Ensure token scope is propagated and enforced in every protected endpoint.
- [ ] Run `bash scripts/run-sdk-contracts.sh github-oauth`.

## Task P29-T02: Slack Official SDK Feasibility

**Status:** pending

- [ ] Test adding `@slack/web-api` as a pinned dev dependency in `contract/sdk`.
- [ ] Attempt official WebClient calls against `SLACK_API_URL` or equivalent base URL override.
- [ ] If feasible, replace or augment `slack-smoke.test.js` with official SDK calls.
- [ ] If not feasible, document the exact blocker in `compat/manifests/slack.json` and keep HTTP client evidence explicit.

## Task P29-T03: Slack Client Contract Deepening

**Status:** pending

- [ ] Add client contract assertions for invalid_auth, channel_not_found, not_in_channel, rate_limited with Retry-After, and invalid_signature.
- [ ] Add message lifecycle assertions: post, update, history, delete, history no longer includes deleted message.
- [ ] Add Events API URL verification and message callback contract if official SDK does not cover it.
- [ ] Run `bash scripts/run-sdk-contracts.sh slack`.

## Task P29-T04: Score And Report Update

**Status:** pending

- [ ] Update manifests so GitHub OAuth and Slack evidence describes client/SDK status honestly.
- [ ] Regenerate compatibility report.
- [ ] Confirm GitHub OAuth and Slack score changes are explainable.
- [ ] Update support matrix known gaps if evidence changes maturity expectations.

## Phase 29 Exit

- [ ] GitHub OAuth client contract covers success and common failure paths.
- [ ] Slack has official SDK evidence or a documented reason for HTTP client evidence.
- [ ] GitHub OAuth and Slack manifests reflect stronger evidence.
- [ ] `run-sdk-contracts.sh all` passes with strengthened contracts.
