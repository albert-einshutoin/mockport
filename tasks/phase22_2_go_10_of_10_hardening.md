# Phase 22.2: Go 10/10 Hardening Track

[日本語版](phase22_2_go_10_of_10_hardening.ja.md)

## Goal

Phase 23 の docs alignment に入る前に、Mockport の Go 実装を「堅い Go 実装」から「Go の言語特性と標準 ecosystem を最大限活かした実装」へ引き上げるための作業を設計する。

Phase 22.2 は Go engineering の最終 hardening track とする。Provider surface を広げるのではなく、既存 behavior を壊さずに、型安全性、context、error handling、HTTP runtime、テスト、静的解析、保守性を底上げする。

## 10/10 Criteria

Go 実装として 10/10 と言える状態は、次を満たすこと。

- Public/internal package boundary が明確で、`internal/` が責務ごとに小さく保たれている。
- Provider adapter contract が typed const / typed struct / small interface で表現され、string typo が compile/test で早く検出される。
- Compatibility level と adapter metadata level が一つの意味体系として扱われ、変換時に情報が silently dropped されない。
- 主要 provider response/error/request model が typed struct で表現され、`map[string]any` は dynamic payload や fixture-like data に限定される。
- HTTP handlers は `context.Context`、request-scoped cancellation、timeout、body size limit、method/path handling を明示的に扱う。
- Streaming は `http.ResponseController` / `http.Flusher` / middleware unwrap まで含めて検証される。
- JSON helpers は encode failure、header write ordering、provider-specific headers を壊さない設計になっている。
- Error は sentinel/type/wrapping を使い分け、handler layer で HTTP response に変換される。
- CLI runtime は `http.Server`、signal handling、graceful shutdown、port conflict error を扱う。
- Shared state は race-free で、mutation boundary と clone depth が明確である。
- Tests は table-driven、race、parallel-safe、contract/smoke/unit の責務が分かれている。
- Static analysis は `go test`, `go vet`, `go test -race`, `staticcheck`, `govulncheck` を CI で実行できる。
- Benchmarks または lightweight performance checks が hot path の regression を検出する。
- Public docs の主張が code/test/report の実測と一致する。

## Non-goals

- Provider API の完全内部再現。
- Stripe/OpenAI/GitHub/Slack の endpoint surface 拡張。
- npm wrapper / Rust component / plugin system の実装。
- 過度な generic abstraction。
- 既存 adapter response shape を理由なく変更する refactor。

## Tasks

| ID | Task | Status | Test First |
| --- | --- | --- | --- |
| P22.2-T01 | Align adapter and compatibility level types | done | Compat test fails until `adapter.LevelClient` is either accepted by compat or removed from adapter metadata without silent data loss |
| P22.2-T02 | Add metadata validation gate | done | Adapter metadata validation test fails on invalid maturity, invalid level, empty provider version, duplicate scenarios, and duplicate endpoints |
| P22.2-T03 | Add typed provider response models for core Stripe paths | done | Stripe adapter tests fail until checkout session, payment intent, and Stripe error responses are encoded from typed structs |
| P22.2-T04 | Add typed provider response models for core OpenAI paths | done | OpenAI adapter tests fail until chat completion, response, embedding, and OpenAI error responses are encoded from typed structs |
| P22.2-T05 | Add typed provider response models for OAuth and Slack core paths | done | GitHub OAuth and Slack tests fail until token/user/message/error responses are encoded from typed structs |
| P22.2-T06 | Restrict `map[string]any` to explicit dynamic boundaries | done | Static Go test/check fails if new core response builders return raw `map[string]any` without an allowlist comment or helper |
| P22.2-T07 | Introduce adapter HTTP error mapping boundary | done | Handler tests fail until validation/state/idempotency errors are converted through typed error helpers instead of ad hoc response writes |
| P22.2-T08 | Make JSON helper error-aware | done | HTTP helper tests fail until JSON encode errors are observable and provider-specific headers remain intact |
| P22.2-T09 | Add request body size limits | done | Handler tests fail until oversized JSON/form/multipart requests return controlled 413/400 responses without unbounded reads |
| P22.2-T10 | Add context-aware request handling guidelines and tests | done | Webhook/client tests fail until request cancellation is propagated and long outbound work uses request context |
| P22.2-T11 | Add graceful server shutdown to `mockport run` | done | CLI/runtime test fails until `mockport run` uses `http.Server`, signal handling, and bounded shutdown |
| P22.2-T12 | Improve listener and port error UX | done | CLI test fails until bind failures return actionable errors with host/port context |
| P22.2-T13 | Harden streaming behavior through middleware | done | Server/OpenAI tests fail until SSE works through recorder middleware with flush, status recording, and report capture |
| P22.2-T14 | Add deep clone guarantees for shared state | done | State tests fail until nested maps/slices in resources cannot be mutated after create/get/list/update |
| P22.2-T15 | Add registry duplicate and nil adapter protections | done | Registry tests fail until duplicate names and nil/invalid adapters are rejected with clear errors |
| P22.2-T16 | Replace large route switches where they reduce clarity | done | Route registration tests fail until large adapters expose deterministic route tables or small handler groups without changing paths |
| P22.2-T17 | Add table-driven conformance tests for every adapter metadata endpoint | done | Metadata/report test fails until every declared endpoint has at least one matching handler test or documented unsupported behavior |
| P22.2-T18 | Add package-level godoc for exported internal contracts | done | Static doc check fails until exported adapter/compat/state/report contracts have concise comments |
| P22.2-T19 | Add benchmark coverage for hot helpers | done | Benchmark target fails until state store, report snapshot, compatibility conversion, and JSON helper benchmarks exist |
| P22.2-T20 | Add staticcheck gate | done | CI/static script fails until `staticcheck ./...` is installed or explicitly documented as unavailable in local-only fallback |
| P22.2-T21 | Add govulncheck gate | done | CI/static script fails until `govulncheck ./...` runs in CI or release readiness checks |
| P22.2-T22 | Add race test gate to CI | done | Workflow/static check fails until `go test -race ./...` is part of scheduled or pre-release CI |
| P22.2-T23 | Add lint policy for ignored errors | done | Static check fails until intentionally ignored errors are limited to documented, low-risk writes/closes or helper-level best-effort paths |
| P22.2-T24 | Add deterministic test mode for timestamps and request IDs | done | Report tests fail until time/request IDs can be injected or asserted deterministically without sleep/flaky behavior |
| P22.2-T25 | Add Go engineering readiness report | done | Script/test fails until it summarizes test/vet/race/staticcheck/govulncheck status and remaining accepted gaps |

## Acceptance Criteria

- `go test ./...` passes.
- `go vet ./...` passes.
- `go test -race ./...` passes.
- `staticcheck ./...` passes in CI or the repository documents the exact installation/runtime blocker.
- `govulncheck ./...` passes in CI or release readiness.
- Compatibility reports no longer drop declared adapter levels silently.
- Core adapter responses are typed at the builder boundary.
- CLI runtime supports graceful shutdown.
- Streaming behavior is verified through middleware, not only direct adapter tests.
- State clone semantics are safe for nested data.
- A Go engineering readiness report makes remaining non-10/10 gaps explicit.

## Suggested Execution Order

1. Fix semantic consistency first: P22.2-T01 through P22.2-T02.
2. Strengthen typed boundaries: P22.2-T03 through P22.2-T08.
3. Harden runtime and HTTP behavior: P22.2-T09 through P22.2-T13.
4. Harden shared infrastructure: P22.2-T14 through P22.2-T18.
5. Add quality gates and evidence: P22.2-T19 through P22.2-T25.

## Implementation Notes

- Keep `map[string]any` where provider payloads are genuinely dynamic, but require the boundary to be named and tested.
- Prefer small concrete types over broad interfaces. Add interfaces only where tests or package boundaries need substitution.
- Use standard library first: `context`, `http.Server`, `http.ResponseController`, `errors.As`, `errors.Is`, `signal.NotifyContext`, `io.LimitReader`, `testing`, `httptest`, `testing/quick` if useful.
- Do not make adapter packages depend on `internal/compat` if that introduces cycles. Shared semantic constants should live at the lowest stable layer or be converted explicitly with validation.
- Do not chase 10/10 by adding framework complexity. The goal is stronger contracts, clearer boundaries, and better automated evidence.
