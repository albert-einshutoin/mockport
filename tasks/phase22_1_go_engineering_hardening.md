# Phase 22.1: Go Engineering Hardening Before Docs Alignment

[日本語版](phase22_1_go_engineering_hardening.ja.md)

## Goal

Phase 23 の roadmap/docs alignment に入る前に、Phase 22 までの実装レビューで見つかった Go 固有の品質課題を解消する。

この Phase は新 provider surface を増やさない。対象は、既存機能の振る舞いを保ったまま、Go の標準 HTTP 特性、型安全性、deterministic output、軽量 helper 境界を改善することに限定する。

## Non-goals

- Provider-compatible のスコアや maturity を上げる。
- 新しい adapter endpoint を追加する。
- npm/Rust/外部 plugin system を進める。
- adapter 固有レスポンス shape を抽象化しすぎる。

## Tasks

| ID | Task | Status | Test First |
| --- | --- | --- | --- |
| P22.1-T01 | Preserve streaming through report middleware | done | Server test fails until `recordMiddleware` allows `http.ResponseController` flush through the wrapped writer |
| P22.1-T02 | Flush OpenAI SSE chunks | done | OpenAI adapter test fails until `stream_success` calls a flush-capable response path |
| P22.1-T03 | Add typed adapter metadata constants | done | Adapter tests fail/compile-fail until maturity and compatibility levels have typed constants |
| P22.1-T04 | Make report adapter ordering deterministic | done | Server report test fails until adapters, coverage, compatibility, and state report entries are sorted by adapter name |
| P22.1-T05 | Extract minimal adapter JSON helper | done | Existing adapter tests must pass after removing duplicated JSON writer implementations |
| P22.1-T06 | Move repeated regexp compilation to package-level values | done | Existing state and compat tests must pass after package-level regexp cleanup |

## Acceptance Criteria

- `go test ./...` passes.
- `go vet ./...` passes.
- `go test -race ./...` passes.
- OpenAI `stream_success` remains SSE-compatible and explicitly flushes chunks when the writer supports it.
- `/_mockport/report` produces deterministic adapter ordering.
- Adapter metadata uses typed constants for known maturity and compatibility level values.
- JSON response helper extraction does not change provider response shape.

## Implementation Notes

- For streaming, prefer `http.NewResponseController(w).Flush()` so wrapped writers can expose optional behavior through `Unwrap`.
- Keep the recorder middleware focused on status capture and request recording; do not couple it to provider-specific streaming logic.
- Typed metadata should live in `internal/adapter` to avoid import cycles with `internal/compat`.
- Shared JSON helpers may live under `internal/adapter/httpx`; keep them tiny and response-shape neutral.
- Sorting should happen before report slices are recorded, not when rendering only.
