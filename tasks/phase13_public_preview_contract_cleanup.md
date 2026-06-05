# Phase 13 Public Preview Contract Cleanup Implementation Plan

[日本語版](phase13_public_preview_contract_cleanup.ja.md)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** public preview で露出した実装との期待値差を compatibility engine 前に潰す。

**Architecture:** provider-compatible の深掘りに入る前に、README/docs/metadata/CLI UX が実装と一致する状態へ戻す。特に `mockport up` の Docker Compose UX、OpenAI `stream_success` の SSE contract、adapter helper 重複の判断をここで閉じる。

**Tech Stack:** Go CLI, Go adapters, httptest, Docker Compose command runner tests, docs/static checks.

---

## Files

- Modify: `internal/cli/up.go`
- Modify: `internal/cli/up_test.go`
- Modify: `adapters/openai/adapter.go`
- Modify: `adapters/openai/adapter_test.go`
- Modify: `docs/site/support-matrix.md`
- Modify: `tasks/status.md`

## Task P13-T01: Improve `mockport up` UX

**Status:** done

- [x] Write failing CLI tests for missing `docker`, missing `docker-compose.mockport.yml`, `--detach` / `-d`, and `--build`.
- [x] Return actionable errors when Docker is unavailable.
- [x] Prompt `mockport init` when the compose file is missing.
- [x] Pass `--detach` / `-d` and `--build` through to `docker compose`.
- [x] Run `go test ./internal/cli -run Up -v`.

## Task P13-T02: Add OpenAI SSE `stream_success`

**Status:** done

- [x] Write failing HTTP tests asserting `text/event-stream`, streaming chunks, and terminal `[DONE]`.
- [x] Implement minimal SSE-compatible response for OpenAI `stream_success`.
- [x] Keep normal non-stream scenarios as JSON.
- [x] Update docs/support matrix so metadata and behavior match. See [issue #6](https://github.com/albert-einshutoin/mockport/issues/6).
- [x] Run `go test ./adapters/openai -run Stream -v`.

## Task P13-T03: Decide Adapter Helper Duplication Boundary

**Status:** done

- [x] Review `writeJSON`, scenario normalization, and error helpers across OpenAI, GitHub OAuth, and Slack adapters.
- [x] Document whether helpers stay local until the next adapter wave or move into a shared package now.
- [x] If extracting helpers, write regression tests proving response shape does not change.
- [x] Run `go test ./adapters/openai ./adapters/githuboauth ./adapters/slack -v`. See [issue #7](https://github.com/albert-einshutoin/mockport/issues/7).

## Phase 13 Exit

- [x] `mockport up` errors are actionable and support `--detach` / `--build`.
- [x] OpenAI `stream_success` is actually SSE-compatible.
- [x] Adapter helper duplication has an explicit decision before more adapters are added.
- [x] Public preview docs and metadata no longer overstate implemented behavior.
