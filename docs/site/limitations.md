# Limitations

[日本語版](limitations.ja.md)

Mockport targets provider-compatible local APIs for selected workflows. It does not reproduce provider internals or undocumented behavior.

## Current Preview Scope

Current mainline adapters are workflow-compatible for selected local and CI integration paths. They are not a substitute for provider sandboxes or production validation.

## Concrete examples you will hit

These are symptom-based limits verified against adapter specs, compatibility reports, and runtime behavior — not abstract promises.

- **Stripe: no 3DS / SCA (`requires_action`) flow** — PaymentIntents and Checkout Sessions return success or decline shapes from built-in scenarios; card authentication UI branches and `requires_action` handling cannot be exercised locally.
- **Stripe: no billing-network math** — Amount and currency fields are echoed from your request; Mockport does not validate tax, proration, settlement, disputes, Connect, or full Billing lifecycle behavior.
- **OpenAI: Mockport does not implement `/v1/responses` streaming** — `/v1/chat/completions` supports SSE when `stream: true` or the `stream_success` scenario is active. A Responses request with `stream: true` fails with HTTP 501 and `error.code=mockport_unsupported_responses_stream`, including under `stream_success`. [OpenAI itself supports Responses streaming](https://platform.openai.com/docs/quickstart/make-your-first-api-request).

  ```bash
  curl -i -X POST http://localhost:43101/openai/v1/responses \
    -H "Authorization: Bearer $OPENAI_API_KEY" \
    -H 'Content-Type: application/json' \
    -d '{"model":"gpt-mockport","input":"hello","stream":true}'
  ```

  Mockport responds with a JSON error, records `reason=unsupported_feature`, and creates no response state. A Responses request without `stream:true` still returns deterministic JSON. Mockport does not return `text/event-stream`, `data:` chunks, or named events such as `response.output_text.delta` on this endpoint.
- **OpenAI: no real inference quality** — Responses are deterministic placeholders; model quality, tokenization parity, hosted tools, vector stores, and provider scheduling are not reproduced.
- **Slack: no real workspace delivery or full Events API** — A fixed signed message event can be delivered to a configured local app and replied to in its thread. The parent message is fixture input; parent/thread retrieval, redelivery scheduling, other event types, Block Kit validation, files, app scopes, and enterprise directory policy are not implemented.
- **LINE: no real Login UI or LIFF browser** — OAuth code/token/profile flows work locally; QR login, LIFF runtime, signed ID tokens, provider webhook redelivery, and quota enforcement beyond scenarios are not reproduced.
- **General: `scenarios:` block in `mockport.yml` is not implemented** — parsed but not applied at runtime; Mockport warns at startup, in `--check`, and in `/_mockport/report` when present.
- **General: state is in-memory only** — container or process restart clears adapter state; there is no persistence layer.

For the full gap list per adapter, see [support matrix](support-matrix.md), adapter specs under `docs/adapters/`, and [compatibility reports](../compatibility-reports/latest.md).

## What Mockport Does Not Reproduce

- Real payment processing, fraud systems, settlement, or billing networks.
- Real AI inference, model quality, provider tokenization, or private scheduling behavior.
- Real GitHub organization, enterprise, or permission policy.
- Real Slack workspace delivery, enterprise policy, or full directory state.
- Real LINE Login UI, LIFF browser runtime, provider webhook redelivery, quota/rate-bucket enforcement, regional policy, or Dapp Portal behavior.
- Undocumented provider behavior.

## Unimplemented Configuration Blocks

The `scenarios:` block in `mockport.yml` is parsed but **not implemented** and is not
applied at runtime. Mockport will emit a warning at startup (and in `--check` output and
`/_mockport/report`) when this block is present.

For response switching and error-case simulation, use:

- Built-in scenarios via the adapter's `scenario:` field in `mockport.yml`
- The `X-Mockport-Scenario` request header (see issue #80)

See [scenario-policy.md](../scenario-policy.md) for future plans on user-defined scenarios.

## Operational notes

### Port conflicts (default `43101`)

Mockport defaults to port `43101`. If another process already listens on that port, startup fails with an error like:

```text
listen on 127.0.0.1:43101: address already in use; choose another port or stop the existing process
```

Change the listen port in `mockport.yml`:

```yaml
server:
  port: 43102
```

Update your app env vars to match (for example `STRIPE_API_URL=http://localhost:43102/stripe`).

### Docker Compose networking

From another container on the same Compose network, point env vars at the service hostname — not `localhost`:

```env
STRIPE_API_URL=http://mockport:43101/stripe
```

`localhost` inside an app container refers to that container, not the Mockport service.

## How To Evaluate Support

Use:

- [Support matrix](support-matrix.md)
- `/_mockport/report`
- `mockport report --json` (`--format json` remains supported)
- Adapter examples
- Public env safety checks

Unsupported endpoints and approximations should be visible before adoption.
