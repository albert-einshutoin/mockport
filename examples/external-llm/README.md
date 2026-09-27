# External app trial: LLM 0.36

This exercises unmodified [simonw/llm 0.36](https://github.com/simonw/llm/releases/tag/0.36), tag commit `764dc386c58b625f3ad9d203e699715ad208455f`, through its ordinary `llm -m mockport-chat` Chat Completions path. It is an independent existing app, unlike the bundled sample apps. It measures API connection and CLI error handling, not AI answer quality or first-time human onboarding. [Issue #395](https://github.com/albert-einshutoin/mockport/issues/395) holds the final CI, main, and published-image evidence.

## Reproduce

Prerequisites: Git, Python 3.13 with `venv` and `pip`, Docker for the published-image run, an unoccupied loopback port 43101, and package-registry access. A source build additionally needs Go 1.26.8. This tag contains the runner, `requirements.lock`, `extra-openai-models.yaml`, and `mockport.yml`. [Issue #395](https://github.com/albert-einshutoin/mockport/issues/395) records the pre-release technical test; the [release record](../../docs/releases/v0.3.0-preview.md) records the versioned-image run.

```sh
git clone --branch v0.3.0-preview --depth 1 https://github.com/albert-einshutoin/mockport.git
cd mockport
git rev-parse HEAD
export MOCKPORT_IMAGE='ghcr.io/albert-einshutoin/mockport@sha256:497f6ec7dc6f1fea8afd292f9480872983bc940425a090aa8ad8580350d8e9a5'
docker pull "$MOCKPORT_IMAGE"
MOCKPORT_IMAGE="$MOCKPORT_IMAGE" bash scripts/run-external-llm.sh
```

The image selector above is the published digest of `0.3.0-preview`, also recorded in the [release record](../../docs/releases/v0.3.0-preview.md). The published-image run on macOS arm64 / Docker linux/arm64 returned CLI exits `0,0,1,0` and report statuses `200,200,401,200`, with image ID equal to the pulled digest. Run `bash scripts/run-external-llm.sh` without `MOCKPORT_IMAGE` only for a source-build check.

The runner creates and removes its own Python venv and empty `LLM_USER_PATH`. It installs the exact Python 3.13 versions in [requirements.lock](requirements.lock), copies only [extra-openai-models.yaml](extra-openai-models.yaml) into that user directory, and uses the normal `--key` option. The model file follows [LLM's documented configuration](https://llm.datasette.io/en/stable/other-models.html). Its `api_key_name` is essential: LLM 0.36 otherwise treats a model with only `api_base` as not needing a key and supplies a dummy key. No user keys or existing conversation history are read.

The script starts one owned Mockport process or digest-pinned container with [mockport.yml](mockport.yml), checks `auth_required`, and refuses an occupied port. Each CLI call has a 30-second ceiling. It compares CLI stdout, stderr, and exit code with one new `POST /openai/v1/chat/completions` and its recorded status in `/_mockport/report`; truncated history fails. The server is stopped on exit. The CLI timeout is only a test-runner limit, not a claim about LLM or SDK timeout behavior. Set `EXTERNAL_LLM_RESULT=/path/to/result.json` to retain machine-readable results.

## Observed results

The initial public baseline was `ghcr.io/albert-einshutoin/mockport@sha256:ddd52e595d472e7d706164f5a0ea8b59e7bb3d85b5c6ad5fbe615da7b929527d` (`linux/arm64`; running image ID matched the digest). On macOS arm64 with Python 3.13, LLM 0.36 and `openai==3.19.2`, the unmodified CLI and normal model configuration produced:

| Case | CLI exit / stdout / stderr | New Mockport request |
| --- | --- | --- |
| Nonstream | `1` / empty / `Error: 'NoneType' object has no attribute 'model_dump'` | one Chat Completions `200` |
| Stream | `0` / `Mockport simulated streaming response.` / empty | one Chat Completions `200` |
| Wrong fake key | `1` / empty / OpenAI error `401` and `invalid_api_key` | one Chat Completions `401` |
| Restored key | `1` / empty / same `NoneType.model_dump` error | one Chat Completions `200` |

LLM 0.36's nonstreaming Chat client dereferences `completion.usage`; Mockport's successful JSON omitted it, along with `created` and `finish_reason`. The initial streaming request sent `stream_options.include_usage=true` and completed, but Mockport omitted the requested final usage chunk. The [OpenAI Chat reference](https://platform.openai.com/docs/api-reference/chat/object) lists `usage` as optional in nonstreaming responses; LLM 0.36 requires it in this code path. For `include_usage=true`, the reference specifies a final usage chunk. The added values are fixed simulated counts (`1`, `1`, `2`), with no tokenization or inference claim.

The corrected source build passed the same four CLI cases: exits `0, 0, 1, 0`; stdout was respectively `Mockport response`, `Mockport simulated streaming response.`, empty, `Mockport response` (each nonempty value ended with a newline). Only the wrong-key case wrote a 401 error to stderr. Report statuses were `200, 200, 401, 200`, one actual request each. The adapter regression tests failed before the repair and passed after it. The runner printed venv creation 1 second, dependency installation 3 seconds, source build 1 second, server startup 0.115 seconds, first CLI success 0.505 seconds, and wrong-key check 0.376 seconds on a warm host. Python, Go, Docker, package caches, and the base image were already available; these are Codex-run phase measurements, not a cold-machine or first-time-user duration.

The existing Mockport Python and Node SDK pins are independent of this LLM venv and remain unchanged. The source run demonstrates the fix before publication; the corrected public digest and its repeat trial are recorded in Issue #395 after the main image workflow completes.
