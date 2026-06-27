# OpenAI Adapter Specification

[日本語版](openai.ja.md)

This document describes the Mockport `openai` adapter contract. It is not a copy of OpenAI's documentation and does not claim full OpenAI API compatibility.

## Scope

The `openai` adapter provides deterministic local behavior for selected OpenAI-like API workflows:

- Model list.
- Chat Completions, including a streaming success scenario.
- Responses create and retrieve.
- Embeddings create.
- Files create for batch workflows.
- Batches create and retrieve.
- OpenAI-like error envelopes for auth, rate limit, context length, malformed input, invalid model, and unsupported parameter cases.

## Base Path

Default base path:

```text
/openai
```

Example config:

```yaml
adapters:
  openai:
    enabled: true
    base_path: /openai
    scenario: chat_success
    fake_secret: mockport_openai_key
```

## Official Reference Map

Use this table to jump from Mockport's supported local surface to the closest official OpenAI documentation. These links are references for behavior shape only; Mockport remains a deterministic local emulator.

| Mockport surface | Official reference |
| --- | --- |
| Model list | `https://platform.openai.com/docs/api-reference/models` |
| Chat Completions create | `https://platform.openai.com/docs/api-reference/chat/create` |
| Chat Completions streaming | `https://platform.openai.com/docs/guides/streaming-responses` |
| Responses create | `https://platform.openai.com/docs/api-reference/responses/create` |
| Responses retrieve | `https://platform.openai.com/docs/api-reference/responses/get` |
| Embeddings create | `https://platform.openai.com/docs/api-reference/embeddings/create` |
| Files create | `https://platform.openai.com/docs/api-reference/files/create` |
| Batches create | `https://platform.openai.com/docs/api-reference/batch/create` |
| Batches retrieve | `https://platform.openai.com/docs/api-reference/batch/retrieve` |

## Supported Endpoints

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/openai/v1/models` | Returns a deterministic model list. |
| `POST` | `/openai/v1/chat/completions` | Returns deterministic chat completion JSON or SSE chunks when streaming. |
| `POST` | `/openai/v1/responses` | Creates a deterministic response object. |
| `GET` | `/openai/v1/responses/{id}` | Retrieves a local response. |
| `POST` | `/openai/test/reset` | Clears provider state and idempotency records for test isolation. |
| `POST` | `/openai/v1/embeddings` | Creates deterministic embedding vectors. |
| `POST` | `/openai/v1/files` | Creates a local file record for batch workflows. |
| `POST` | `/openai/v1/batches` | Creates a local batch record. |
| `GET` | `/openai/v1/batches/{id}` | Retrieves a local batch record. |

## Scenarios

| Scenario | Behavior |
| --- | --- |
| `chat_success` | Default successful local workflow. |
| `stream_success` | Returns SSE-compatible chat completion chunks for streaming chat requests. |
| `rate_limited` | Returns OpenAI-like rate limit behavior. |
| `context_length_exceeded` | Returns OpenAI-like context length behavior. |
| `auth_error` | Returns authentication-style failures. |

## Current Gaps And Tasks

| Priority | Task | Current source of truth |
| --- | --- | --- |
| P1 | Responses API SSE streaming (`stream: true`) is not implemented. `/v1/responses` always returns JSON, including when the `stream_success` scenario is active. Named events such as `response.output_text.delta` are out of scope for this adapter track. | `adapters/openai/adapter_test.go` (`TestResponsesStreamSuccessReturnsJSON`) and `docs/site/limitations.md` |
| P1 | Define selected OpenAI workflows in `compat/manifests/openai.json`, including explicit non-goals for model quality, tokenization parity, hosted tools, vector stores, and provider scheduling. | `tasks/phase28_openai_provider_compatible_track.md` |
| P1 | Deepen SDK contracts for SSE chunk shape, terminal completion, content accumulation, malformed input, unsupported parameters, invalid model, context length, auth, and rate limit behavior. | `contract/sdk/openai-smoke.test.js` and `compat/fixtures/openai/` |
| P1 | Verify response retrieve and batch retrieve consistency before any maturity promotion. | `tasks/phase28_openai_provider_compatible_track.md` |
| P2 | Keep fake inference deterministic and clearly separate local API shape from real model quality. | `docs/site/support-matrix.md` |

## Verification

Run the adapter tests and SDK contract:

```bash
/usr/local/go/bin/go test ./adapters/openai
bash scripts/run-sdk-contracts.sh openai
```
