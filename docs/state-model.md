# State Model

[日本語版](state-model.ja.md)

Mockport state is deterministic, in-memory, and provider-neutral. It exists to support local API-compatible workflows without reproducing provider internals.

## Store Scope

State is scoped by adapter and resource type. IDs are deterministic within a scope:

```text
{adapter}_{resource_type}_{sequence}
```

Examples:

- `stripe_checkout_session_000001`
- `openai_chat_completion_000001`
- `slack_message_000001`

The shared store supports create, retrieve, list, update, delete, and reset. Reset clears both resources and deterministic counters for a scope so tests can start from a stable baseline.

## Idempotency

The idempotency primitive stores a provider-neutral request fingerprint and response for a scope/key pair.

- First matching request records the response.
- Repeated request with the same fingerprint replays the stored response.
- Repeated request with a different fingerprint returns an idempotency conflict error.
- Empty idempotency keys are ignored by the primitive so adapters can decide whether a provider requires them.
- Records are retained up to `MaxIdempotencyRecordsPerScope` per scope. When a scope exceeds that limit, the oldest records are evicted deterministically.

Adapters own provider-shaped error responses. The shared primitive only identifies replay vs conflict.

## Validation

Required-field validation returns a provider-neutral validation error with missing field names. Adapters map that error to Stripe/OpenAI/GitHub/Slack-shaped response bodies and status codes.

## Reporting

Adapters can expose state coverage through metadata:

- stateful resource names
- idempotency support
- reset support

The report includes these fields as state coverage hooks. Existing adapters are not marked stateful until Phase 17 adopts the shared store.

## Limits

This model does not reproduce provider databases, asynchronous jobs, fraud/risk engines, provider-side account configuration, or undocumented internal state transitions. It only models public API-visible resource lifecycle behavior needed for local development and SDK contracts.
