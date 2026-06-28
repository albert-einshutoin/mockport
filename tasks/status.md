# Mockport Implementation Status

[日本語版](status.ja.md)

最終更新: 2026-05-26

## Decisions

| Decision | Status | Note |
| --- | --- | --- |
| First adapter | done | Stripe first |
| Runtime priority | done | Docker first, Go binary second |
| Repository model | done | Single repository for MVP |
| Adapter model | done | Built-in adapters first |
| npm wrapper | done | Later, not Minimal MVP |
| Rust component | done | Later, not Minimal MVP |
| Dynamic plugins | done | Later, not Minimal MVP |

## Phase Summary

| Phase | Goal | Status | Exit Evidence |
| --- | --- | --- | --- |
| Phase 0 | Repository skeleton, CLI, config, server, health, Docker, CI | done | Go 1.26.3 test/vet/build, Docker build, `/health` 200 |
| Phase 1 | Stripe-like Minimal MVP | done | Stripe scenarios, webhook tests, report, AI-safe tests, Docker run |
| Phase 2 | CLI UX | done | Empty directory init/up/run flow works in under 2 minutes |
| Phase 3 | AI-safe mode | done | Warn/fail/redact/report/docs are explicit and tested |
| Phase 4 | Trust reports and adapter contracts | done | Report explains supported/unsupported behavior before adding more adapters |
| Phase 5 | Additional adapters | done | OpenAI, GitHub OAuth, Slack-like adapters, examples, multi-adapter CLI, Docker smoke |
| Phase 6 | Distribution | done | GHCR/release/Homebrew/npm/docs distribution paths are documented and tested where local |
| Phase 7 | Public OSS hardening | done | Public trust files, contribution surface, first-run install path, public CI gates |
| Phase 8 | Public env safety | done | `.env.mockport.example` is safe-to-commit by policy, scanner, docs, and CI before preview release |
| Phase 9 | Public docs and discovery | done | Support matrix, limitations, examples, and positioning docs are public-ready before preview release |
| Phase 10 | Public preview release | done | `v0.1.0-alpha` GitHub Release and GHCR image are published and install-verified |
| Phase 11 | Community and maintenance | done | Maintenance policy, Dependabot, roadmap, Node.js 24 Actions, and adapter contribution quality bar |
| Phase 12 | Fixture, spec, and scenario policy | done | Sanitized fixtures, source metadata, provider spec snapshot rules, and scenario policy exist |
| Phase 13 | Public preview contract cleanup | done | `mockport up`, OpenAI streaming, and adapter helper boundaries no longer create expectation gaps |
| Phase 14 | Compatibility engine | done | Compatibility manifests, scores, reports, and provisional promotion rules define provider-compatible local API |
| Phase 15 | SDK contract harness foundation | done | Pinned SDK contract runner reaches local Mockport without external provider calls |
| Phase 16 | State foundation | done | Shared deterministic state, idempotency primitives, and state coverage report hooks exist |
| Phase 17 | Adapter state adoption | done | Major adapters adopt fake state without breaking scenario-compatible behavior |
| Phase 18 | Stripe provider compatibility | done | Stripe reaches workflow-compatible status with SDK contracts and support matrix |
| Phase 19 | OpenAI provider compatibility | done | OpenAI reaches workflow-compatible status with SDK contracts and support matrix |
| Phase 20 | GitHub OAuth provider compatibility | done | GitHub OAuth reaches workflow-compatible status with client contracts and support matrix |
| Phase 21 | Slack provider compatibility | done | Slack reaches workflow-compatible status with client contracts and support matrix |
| Phase 22 | Provider-compatible release track | done | Compatibility CI and release reports publish scores, SDK versions, and known gaps |
| Phase 22.1 | Go engineering hardening before docs alignment | done | Streaming middleware, typed metadata, deterministic reports, helper boundaries, and regexp cleanup are hardened before Phase 23 |
| Phase 22.2 | Go 10/10 hardening track | done | Typed responses, context/error/runtime hardening, state clone safety, static analysis gates, and Go readiness evidence close remaining Go-specific gaps |
| Phase 23 | Roadmap and docs alignment | pending | Roadmap, README, docs, changelog, and compatibility report explain the same current state |
| Phase 24 | GitHub Actions execution recovery | pending | CI and compatibility workflows create observable GitHub Actions runs or the blocker is documented |
| Phase 25 | SDK contract all-provider harness | pending | `run-sdk-contracts.sh all` runs every provider-specific contract, not a placeholder |
| Phase 26 | Provider-compatible manifest promotion | done | Versioned manifests and release checks gate manifest drift and maturity claims |
| Phase 27 | Stripe provider-compatible track | pending | Stripe selected workflows have contract-level evidence or explicit blockers |
| Phase 28 | OpenAI provider-compatible track | pending | OpenAI selected workflows have contract-level evidence or explicit blockers |
| Phase 29 | GitHub OAuth and Slack client evidence | pending | GitHub OAuth and Slack client/SDK evidence is strong enough for honest scoring |
| Phase 30 | v0.2.0-preview release | pending | Release artifacts, GHCR image, compatibility report, and post-release smoke are verified |
| Phase 31 | Adapter reference docs | done | Registered adapter docs include official reference maps and implementation boundaries for `stripe`, `openai`, `github-oauth`, `slack`, `line`, and `zoho-oauth` |
| Phase 32 | Service baseline execution | pending | Registered adapter baselines and SendGrid execution order are explicit and verifiable |

## Phase 0 Tasks

| ID | Task | Status | Test First |
| --- | --- | --- | --- |
| P0-T01 | Create repository skeleton and Go module | done | `go test ./...` compiles empty packages |
| P0-T02 | Add Cobra root command and version command | done | CLI test asserts command output |
| P0-T03 | Add config defaults and YAML loader | done | Config tests cover valid config, default host, invalid port |
| P0-T04 | Add security detector primitives | done | Security tests cover real-looking and fake secrets |
| P0-T05 | Add HTTP server and `/health` | done | `httptest` checks 200 JSON response |
| P0-T06 | Add `mockport run` | done | CLI/server test starts with config and serves health |
| P0-T07 | Add Dockerfile, Makefile, CI workflow | done | Docker build passes with `golang:1.26-alpine` |
| P0-T08 | Add root README from draft | done | README quickstart matches current commands |

## Phase 1 Tasks

| ID | Task | Status | Test First |
| --- | --- | --- | --- |
| P1-T01 | Add adapter registry | done | Registry test resolves enabled adapter |
| P1-T02 | Add request recorder and report model | done | Report test records request, status, safety fields |
| P1-T03 | Add Stripe checkout session success/failure | done | HTTP tests assert 200 and 402 responses |
| P1-T04 | Add Stripe payment intent scenarios | done | HTTP tests assert success, failure, auth, rate limit |
| P1-T05 | Add timeout scenario behavior | done | HTTP test uses 504 simulated timeout response |
| P1-T06 | Add webhook sender endpoint | done | `httptest` target receives signed fake event |
| P1-T07 | Add AI-safe config validation and strict mode | done | Config/security tests warn or fail on real-looking values |
| P1-T08 | Add `/_mockport/report` endpoint | done | HTTP test asserts request and safety report JSON |
| P1-T09 | Add `mockport init` generated files | done | CLI test in temp dir asserts `mockport.yml`, `.env.mockport`, compose file |
| P1-T10 | Add Minimal MVP verification docs | done | Docs/examples created; Docker run verified |

## Minimal MVP Exit Checklist

- [x] `go test ./...` passes with Go 1.26.3.
- [x] `go vet ./...` passes with Go 1.26.3.
- [x] `go build ./cmd/mockport` passes with Go 1.26.3.
- [x] `docker build -t mockport:local -f docker/Dockerfile .` passes.
- [x] `docker run -p 43101:43101 ... mockport:local` starts the server.
- [x] `curl http://localhost:43101/health` returns 200.
- [x] `POST /stripe/v1/checkout/sessions` returns a success response in `payment_success` in `httptest`.
- [x] Stripe failure scenario returns 402 in `httptest`.
- [x] Webhook sender posts to configured target URL in `httptest`.
- [x] `/_mockport/report` shows requests and safety warnings in `httptest`.
- [x] README includes a working quickstart.

## Phase 2 Tasks

| ID | Task | Status | Test First |
| --- | --- | --- | --- |
| P2-T01 | Protect generated files from accidental overwrite | done | CLI tests assert existing files are not overwritten without `--force` |
| P2-T02 | Add `mockport init --force` | done | CLI tests assert `--force` overwrites generated files |
| P2-T03 | Add `mockport up` command | done | CLI tests assert compose command construction |
| P2-T04 | Add `mockport report --url` command | done | CLI tests use `httptest` report endpoint |
| P2-T05 | Add user-facing init/run success output | done | CLI tests assert next-step output |
| P2-T06 | Add empty-directory E2E script/doc | done | Shell smoke test validates init + Docker run path |

## Phase 3 Tasks

| ID | Task | Status | Test First |
| --- | --- | --- | --- |
| P3-T01 | Add safety status model | done | Report tests assert `safe`, warning counts, and categories |
| P3-T02 | Emit startup warnings in ai-safe mode | done | CLI run tests assert warning text without full secrets |
| P3-T03 | Harden strict mode exit behavior | done | CLI run tests assert non-zero error for unsafe config |
| P3-T04 | Expand redaction coverage | done | Security tests cover URLs, short values, env-like values |
| P3-T05 | Block proxy-like external URLs | done | Config tests assert known provider URLs fail in strict mode |
| P3-T06 | Add AI-safe docs and examples | done | Docs command examples match implemented behavior |

## Phase 4 Tasks

| ID | Task | Status | Test First |
| --- | --- | --- | --- |
| P4-T01 | Add adapter metadata contract | done | Adapter tests assert capabilities, scenarios, endpoints, maturity are discoverable |
| P4-T02 | Add scenario coverage report | done | Report tests assert per-adapter scenario matrix |
| P4-T03 | Record unsupported endpoint attempts | done | HTTP tests assert 404/405 entries appear in report |
| P4-T04 | Add request replay log metadata | done | Recorder tests assert stable request ids and replay-safe data |
| P4-T05 | Add behavior matrix and maturity levels | done | Report tests assert endpoints, scenarios, maturity, and support status |
| P4-T06 | Add machine-readable and text report modes | done | CLI tests assert JSON and text report output |

## Phase 5 Tasks

| ID | Task | Status | Test First |
| --- | --- | --- | --- |
| P5-T01 | Add OpenAI-compatible adapter | done | HTTP tests cover models, chat success, auth, rate limit |
| P5-T02 | Add GitHub OAuth-like adapter | done | HTTP tests cover authorize, token, user, invalid code |
| P5-T03 | Add Slack-like messaging adapter | done | HTTP tests cover auth.test, chat.postMessage, rate limit |
| P5-T04 | Extend `mockport init/add` for multiple adapters | done | CLI tests assert multi-adapter config/env generation |
| P5-T05 | Add examples for each adapter | done | Example configs load and adapter routes respond |
| P5-T06 | Add cross-adapter smoke coverage | done | Docker smoke validates multiple adapters in one config |

## Phase 6 Tasks

| ID | Task | Status | Test First |
| --- | --- | --- | --- |
| P6-T01 | Add release build workflow | done | Workflow lint/static checks assert expected matrix targets |
| P6-T02 | Add GHCR image workflow | done | Workflow checks assert docker metadata and tags |
| P6-T03 | Add release archives and checksums | done | Local script test asserts archive names and checksum file |
| P6-T04 | Add Homebrew formula template | done | Template test asserts version/url/sha placeholders |
| P6-T05 | Add npm wrapper design scaffold | done | Package tests assert wrapper delegates to binary or Docker |
| P6-T06 | Add docs site scaffold | done | Docs build check renders quickstart and adapter pages |

## Phase 7 Tasks

| ID | Task | Status | Test First |
| --- | --- | --- | --- |
| P7-T01 | Add public trust artifacts | done | Static trust check fails until LICENSE/SECURITY/CONTRIBUTING/CODE_OF_CONDUCT exist |
| P7-T02 | Add GitHub collaboration surface | done | Static trust check asserts issue and PR templates exist |
| P7-T03 | Fix README first-run install path | done | README command audit starts from no preinstalled `mockport` |
| P7-T04 | Add public CI gates | done | CI runs trust and distribution static checks |

## Phase 8 Tasks

| ID | Task | Status | Test First |
| --- | --- | --- | --- |
| P8-T01 | Define public-safe env policy | done | Static docs check asserts allowed prefixes and forbidden real provider patterns |
| P8-T02 | Add public env scanner | done | Security tests catch real-looking secrets in env files and docs snippets |
| P8-T03 | Add public env UX | done | CLI/report tests assert generated env is fake, local, and safe-to-commit |

## Phase 9 Tasks

| ID | Task | Status | Test First |
| --- | --- | --- | --- |
| P9-T01 | Build public docs information architecture | done | Docs link/path check asserts expected pages |
| P9-T02 | Add example-driven onboarding docs | done | Example configs load and smoke path is covered |
| P9-T03 | Add public positioning and limitations | done | Markdown/link checks cover comparison and limitations pages |

## Phase 10 Tasks

| ID | Task | Status | Test First |
| --- | --- | --- | --- |
| P10-T01 | Add preview release readiness audit | done | Release verification script checks archive names, checksums, version, image |
| P10-T02 | Publish `v0.1.0-alpha` GitHub Release | done | Release workflow produces archives and `checksums.txt` |
| P10-T03 | Verify GHCR preview publish | done | Docker pull and smoke use GHCR preview image, not local build |
| P10-T04 | Update preview install docs | done | README install audit passes from a temporary directory |

## Phase 11 Tasks

| ID | Task | Status | Test First |
| --- | --- | --- | --- |
| P11-T01 | Add maintenance policy | done | Static maintenance check asserts roadmap and maintainer guide |
| P11-T02 | Add dependency and CI maintenance | done | Dependabot config, scheduled checks, and Node.js 24-compatible Actions are validated |
| P11-T03 | Define adapter contribution quality bar | done | Public trust/docs checks cover adapter PR criteria |

## Public Preview Follow-up Backlog

| Issue | Priority | Destination | Status |
| --- | --- | --- | --- |
| [#6](https://github.com/albert-einshutoin/mockport/issues/6) Add SSE-compatible streaming response for OpenAI `stream_success` scenario | high | Phase 13 | done |
| [#8](https://github.com/albert-einshutoin/mockport/issues/8) Improve `mockport up` Docker Compose UX | high | Phase 13 | done |
| [#5](https://github.com/albert-einshutoin/mockport/issues/5) Clarify scenario policy: built-in scenarios vs user-defined scenarios | medium | Phase 12 | done in docs/scenario-policy.md |
| [#7](https://github.com/albert-einshutoin/mockport/issues/7) Track adapter helper duplication before adding more adapters | low | Phase 13 | done in docs/adapter-helper-policy.md |

## Phase 12 Tasks

| ID | Task | Status | Test First |
| --- | --- | --- | --- |
| P12-T01 | Define fixture format and scenario policy | done | Fixture check asserts source metadata, fake local credentials, and built-in/user-defined scenario rules |
| P12-T02 | Add fixture safety check | done | Scanner rejects real secrets and production provider URLs in fixtures |
| P12-T03 | Add spec snapshot policy | done | Docs/static checks cover source and update policy |

## Phase 13 Tasks

| ID | Task | Status | Test First |
| --- | --- | --- | --- |
| P13-T01 | Improve `mockport up` UX | done | CLI tests cover missing Docker, missing compose file, `--detach` / `-d`, and `--build` |
| P13-T02 | Add OpenAI SSE `stream_success` | done | HTTP tests assert `text/event-stream`, streaming chunks, and terminal `[DONE]` |
| P13-T03 | Decide adapter helper duplication boundary | done | Adapter regression tests preserve response shape if helpers are extracted |

## Phase 14 Tasks

| ID | Task | Status | Test First |
| --- | --- | --- | --- |
| P14-T01 | Add compatibility level model | done | Manifest tests cover wire/sdk/workflow/error/state/contract levels |
| P14-T02 | Add compatibility scoring | done | Score tests cover endpoint, scenario, SDK, state, and error coverage |
| P14-T03 | Add compatibility report output | done | Report tests assert score, levels, SDK versions, provider API version, gaps |
| P14-T04 | Define provisional promotion rule | done | Static checks prevent undocumented compatibility claims |

## Phase 15 Tasks

| ID | Task | Status | Test First |
| --- | --- | --- | --- |
| P15-T01 | Add SDK contract workspace | done | SDK contract package reaches local Mockport health endpoint |
| P15-T02 | Add Mockport contract runner | done | Runner starts Mockport, executes selected tests, and cleans up |
| P15-T03 | Add CI integration | done | CI runs SDK contract foundation without external provider calls |

## Phase 16 Tasks

| ID | Task | Status | Test First |
| --- | --- | --- | --- |
| P16-T01 | Add deterministic state store | done | Store tests cover create/retrieve/list/update/delete and reset |
| P16-T02 | Add idempotency and validation primitives | done | State tests cover replay, conflict detection, and missing required field errors |
| P16-T03 | Add state coverage report hooks | done | Report tests assert stateful resources, idempotency support, and reset behavior |

## Phase 17 Tasks

| ID | Task | Status | Test First |
| --- | --- | --- | --- |
| P17-T01 | Adopt state in Stripe workflows | done | Adapter tests cover stateful checkout/payment intent and idempotency |
| P17-T02 | Adopt state in OpenAI workflows | done | Adapter tests cover IDs, retrieve-compatible fake workflows, validation, and preserved streaming |
| P17-T03 | Adopt state in OAuth and messaging workflows | done | Adapter tests cover codes, tokens, scopes, users, channels, messages |

## Phase 18 Tasks

| ID | Task | Status | Test First |
| --- | --- | --- | --- |
| P18-T01 | Add Stripe SDK contract baseline | done | Official SDK smoke covers checkout sessions and payment intents |
| P18-T02 | Expand Stripe major surface | done | Endpoint group tests and SDK contracts cover customers/prices/products/subscriptions/invoices/refunds |
| P18-T03 | Add Stripe error and idempotency fidelity | done | Adapter tests cover validation, auth, rate limit, idempotency replay, and conflict errors |

## Phase 19 Tasks

| ID | Task | Status | Test First |
| --- | --- | --- | --- |
| P19-T01 | Add OpenAI SDK contract baseline | done | Official SDK smoke covers models, chat, responses, streaming where feasible |
| P19-T02 | Expand OpenAI major surface | done | Endpoint group tests and SDK contracts cover embeddings/files/batches/tool-call subset |
| P19-T03 | Add OpenAI error and streaming fidelity | done | Adapter tests cover auth, rate limit, context length, invalid model, malformed input, and Phase 13 streaming fixtures |

## Phase 20 Tasks

| ID | Task | Status | Test First |
| --- | --- | --- | --- |
| P20-T01 | Add GitHub OAuth client contract baseline | done | Client smoke covers authorize redirect, token exchange, user profile, and user emails subset |
| P20-T02 | Add GitHub OAuth state and scope fidelity | done | Tests cover codes, tokens, scopes, expiry, fake identities, and scope errors |
| P20-T03 | Add GitHub OAuth error fidelity | done | Tests cover token endpoint errors, API auth errors, unsupported scopes, and unsupported endpoints |

## Phase 21 Tasks

| ID | Task | Status | Test First |
| --- | --- | --- | --- |
| P21-T01 | Add Slack client contract baseline | done | Client smoke covers auth.test, chat.postMessage, conversations list/history where needed |
| P21-T02 | Add Slack messaging and conversation state | done | Tests cover channels, users, messages, timestamps, update/delete, and history |
| P21-T03 | Add Slack error and rate limit fidelity | done | Tests cover invalid_auth, channel_not_found, not_in_channel, rate_limited, delivery_failed |

## Phase 22 Tasks

| ID | Task | Status | Test First |
| --- | --- | --- | --- |
| P22-T01 | Add compatibility CI | done | Workflow static checks assert SDK contracts, fixture checks, report artifact |
| P22-T02 | Generate compatibility reports | done | Report generation tests assert adapter scores, SDK versions, known gaps |
| P22-T03 | Define provider-compatible release criteria | done | Release check enforces minimum score and passing contracts before maturity promotion |

## Phase 22.1 Tasks

| ID | Task | Status | Test First |
| --- | --- | --- | --- |
| P22.1-T01 | Preserve streaming through report middleware | done | Server test fails until `recordMiddleware` allows `http.ResponseController` flush through the wrapped writer |
| P22.1-T02 | Flush OpenAI SSE chunks | done | OpenAI adapter test fails until `stream_success` calls a flush-capable response path |
| P22.1-T03 | Add typed adapter metadata constants | done | Adapter tests fail/compile-fail until maturity and compatibility levels have typed constants |
| P22.1-T04 | Make report adapter ordering deterministic | done | Server report test fails until adapters, coverage, compatibility, and state report entries are sorted by adapter name |
| P22.1-T05 | Extract minimal adapter JSON helper | done | Existing adapter tests must pass after removing duplicated JSON writer implementations |
| P22.1-T06 | Move repeated regexp compilation to package-level values | done | Existing state and compat tests must pass after package-level regexp cleanup |

## Phase 22.2 Tasks

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

## Phase 23 Tasks

| ID | Task | Status | Test First |
| --- | --- | --- | --- |
| P23-T01 | Audit current docs state | pending | Static review records mismatches between tasks, roadmap, support matrix, and compatibility report |
| P23-T02 | Refresh ROADMAP current and next work | pending | `rg` checks prove stale near-term Phase 12-16 language is no longer future-looking |
| P23-T03 | Align README and docs site | pending | Doc link and public trust checks cover updated support matrix, limitations, and report links |
| P23-T04 | Verify documentation consistency | pending | `check-doc-links`, `check-public-trust`, and `check-compatibility-release` pass |

## Phase 24 Tasks

| ID | Task | Status | Test First |
| --- | --- | --- | --- |
| P24-T01 | Collect GitHub Actions evidence | pending | `gh` commands capture repo Actions permissions, workflow list, and run list |
| P24-T02 | Audit workflow triggers and YAML | pending | Local parser/static checks fail on invalid workflow YAML or missing triggers |
| P24-T03 | Recover workflow execution | pending | Workflow enable/settings/docs changes address the discovered no-run cause |
| P24-T04 | Prove CI and compatibility runs | pending | `gh workflow run` plus `gh run watch --exit-status` succeeds or blocker is documented |

## Phase 25 Tasks

| ID | Task | Status | Test First |
| --- | --- | --- | --- |
| P25-T01 | Add RED coverage for `--provider all` | pending | Live `all` contract fails until every provider result is present |
| P25-T02 | Implement all-provider contract runner | pending | `test-runner.js --provider all --json` returns results for Stripe, OpenAI, GitHub OAuth, and Slack |
| P25-T03 | Wire real all-provider gate into CI | pending | CI requires the real `run-sdk-contracts.sh all` gate |
| P25-T04 | Verify all provider contracts | pending | Single-provider contracts and `all` pass against the same local Mockport runtime |

## Phase 26 Tasks

| ID | Task | Status | Test First |
| --- | --- | --- | --- |
| P26-T01 | Add manifest generator and drift check | done | `scripts/gen-compat-manifests` and `scripts/check-compat-manifests.sh` gate checked-in manifests |
| P26-T02 | Create adapter manifests | done | Stripe, OpenAI, GitHub OAuth, Slack, and LINE manifests committed under `compat/manifests/` |
| P26-T03 | Connect promotion gate to manifests | done | Report validator requires report maturity to match checked-in manifest maturity |
| P26-T04 | Wire release and CI checks | done | `check-compatibility-release.sh` and CI run manifest drift checks |

## Phase 27 Tasks

| ID | Task | Status | Test First |
| --- | --- | --- | --- |
| P27-T01 | Define Stripe provider-compatible scope | pending | Manifest checker validates selected Stripe workflows and explicit non-goals |
| P27-T02 | Deepen Stripe SDK contracts | pending | Stripe SDK tests fail before list/retrieve/idempotency/error evidence is implemented |
| P27-T03 | Improve Stripe adapter fidelity | pending | Adapter tests and SDK contracts cover selected workflows without external Stripe calls |
| P27-T04 | Make Stripe maturity decision | pending | Compatibility gate promotes Stripe or records exact blockers |

## Phase 28 Tasks

| ID | Task | Status | Test First |
| --- | --- | --- | --- |
| P28-T01 | Define OpenAI provider-compatible scope | pending | Manifest checker validates selected OpenAI workflows and explicit non-goals |
| P28-T02 | Deepen OpenAI SDK streaming and error contracts | pending | OpenAI SDK tests fail before SSE/error/retrieve consistency is implemented |
| P28-T03 | Improve OpenAI adapter fidelity | pending | SDK contracts verify selected workflows while fake inference remains deterministic |
| P28-T04 | Make OpenAI maturity decision | pending | Compatibility gate promotes OpenAI or records exact blockers |

## Phase 29 Tasks

| ID | Task | Status | Test First |
| --- | --- | --- | --- |
| P29-T01 | Strengthen GitHub OAuth client contract | pending | Client contract fails before success and common failure paths are covered |
| P29-T02 | Test Slack official SDK feasibility | pending | Pinned SDK experiment proves official SDK support or records exact blocker |
| P29-T03 | Deepen Slack client contract | pending | Slack contract fails before errors, rate limits, lifecycle, and Events API paths are covered |
| P29-T04 | Update evidence scores and reports | pending | Generated report explains GitHub OAuth and Slack evidence changes |

## Phase 30 Tasks

| ID | Task | Status | Test First |
| --- | --- | --- | --- |
| P30-T01 | Run release readiness gate | pending | Full test, contract, compatibility, distribution, and public trust checks pass before tagging |
| P30-T02 | Write v0.2.0-preview release notes | pending | Release docs include compatibility scores, known gaps, and verification commands |
| P30-T03 | Publish release artifacts | pending | GitHub Release, checksums, and GHCR image exist or blockers are documented |
| P30-T04 | Run post-release smoke | pending | Clean download/container smoke verifies health plus all major adapters |

## Phase 31 Tasks

| ID | Task | Status | Test First |
| --- | --- | --- | --- |
| P31-T01 | Inventory registered adapters and open adapter work | done | Task audit fails if `internal/cli/run.go` has an adapter missing from docs/task inventory; current inventory covers `stripe`, `openai`, `github-oauth`, `slack`, `line`, and `zoho-oauth` |
| P31-T02 | Add official reference maps for adapter docs | done | Docs link check fails on missing per-adapter spec links |
| P31-T03 | Keep adapter task inventory current | done | `scripts/check-support-surfaces.mjs` fails when Phase 31/32 registered-adapter lists drift from built-ins |
| P31-T04 | Freeze adapter implementation line | done | Compatibility model and Phase 31 define in-scope/out-of-scope boundaries before adapter implementation continues |

## Verification Notes

- `/usr/local/go/bin/go version`: `go version go1.26.3 darwin/arm64`.
- Passed: `/usr/local/go/bin/go test ./...`.
- Passed: `/usr/local/go/bin/go vet ./...`.
- Passed: `/usr/local/go/bin/go build ./cmd/mockport`.
- Passed: `docker build -t mockport:local -f docker/Dockerfile .`.
- Passed with `mockport:local`: `GET /health`, `POST /stripe/v1/checkout/sessions`, `GET /_mockport/report`.
- Passed: `bash scripts/smoke-empty-dir.sh`.
- Passed: `bash scripts/smoke-multi-adapter.sh` with Stripe, OpenAI, GitHub OAuth, Slack, LINE, and Zoho OAuth endpoints.
- Passed: `bash scripts/check-distribution.sh`.
- Passed: `bash scripts/test-release-archives.sh`.
- Passed: `(cd packaging/npm && npm test)`.
- Passed: `docker build -t mockport:local -f docker/Dockerfile .`.
- Passed: `bash scripts/check-public-trust.sh`.
- Passed: README first-run Docker audit: build `mockport:local`, run with `configs/mockport.example.yml`, `GET /health`, `POST /stripe/v1/checkout/sessions`, `GET /_mockport/report`.
- Passed: `bash scripts/check-public-env.sh`.
- Passed: `bash scripts/smoke-empty-dir.sh` with public env safe-to-commit report field.
- Passed: `bash scripts/check-doc-links.sh`.
- Passed: `bash scripts/smoke-multi-adapter.sh`.
- Passed: Phase 22.1 RED checks for streaming middleware, OpenAI SSE flush, typed metadata constants, and deterministic report order.
- Passed: `go test ./...` after Phase 22.1.
- Passed: `go vet ./...` after Phase 22.1.
- Passed: `go test -race ./...` after Phase 22.1.
- Passed: Phase 22.2 RED checks for client level preservation, metadata validation, registry guards, JSON helper errors, deep state clone, deterministic recorder clock, graceful shutdown, and streaming report capture.
- Passed: `go test ./...` after Phase 22.2.
- Passed: `go vet ./...` after Phase 22.2.
- Passed: `go test -race ./...` after Phase 22.2.
- Passed: `bash scripts/check-go-engineering.sh`.
- Passed: benchmark smoke for state store, report snapshot, compatibility score, and JSON helper.
- Passed: `staticcheck ./...` after installing `honnef.co/go/tools/cmd/staticcheck`.
- Passed: `govulncheck ./...` after installing `golang.org/x/vuln/cmd/govulncheck`.
