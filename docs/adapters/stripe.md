# Stripe Adapter Specification

[日本語版](stripe.ja.md)

This document describes the Mockport `stripe` adapter contract. It is not a copy of Stripe's documentation and does not claim full Stripe API compatibility.

## Scope

The `stripe` adapter provides deterministic local behavior for selected Stripe-like payment workflows:

- Checkout Sessions create/list/retrieve.
- PaymentIntents create/list/retrieve.
- Customers, Products, Prices, Subscriptions, Invoices, and Refunds create/list/retrieve.
- Stripe-like error envelopes for auth, rate limit, payment failure, timeout, validation, and idempotency conflicts.
- `timeout` is an immediate timeout response shape; use server-wide `X-Mockport-Delay` (accepted range `0`–`30000` ms; see [Adapters](../site/adapters.md)) to inject realistic latency before handling.
- Fake signed webhook delivery to a configured local target.

## Base Path

Default base path:

```text
/stripe
```

The adapter also exposes an SDK-compatible `/v1` alias when mounted by the configured server.

Example config:

```yaml
adapters:
  stripe:
    enabled: true
    base_path: /stripe
    scenario: payment_success
    fake_secret: mockport_stripe_secret
    webhook:
      target_url: http://app:3000/webhooks/stripe
      signing_secret: whsec_mockport
```

## Official Reference Map

Use this table to jump from Mockport's supported local surface to the closest official Stripe documentation. These links are references for behavior shape only; Mockport remains a deterministic local emulator.

| Mockport surface | Official reference |
| --- | --- |
| Checkout Session create/list/retrieve | `https://docs.stripe.com/api/checkout/sessions` |
| PaymentIntent create/list/retrieve | `https://docs.stripe.com/api/payment_intents` |
| Customer create/list/retrieve | `https://docs.stripe.com/api/customers` |
| Product create/list/retrieve | `https://docs.stripe.com/api/products` |
| Price create/list/retrieve | `https://docs.stripe.com/api/prices` |
| Subscription create/list/retrieve | `https://docs.stripe.com/api/subscriptions` |
| Invoice create/list/retrieve | `https://docs.stripe.com/api/invoices` |
| Refund create/list/retrieve | `https://docs.stripe.com/api/refunds` |
| Webhook signature verification concepts | `https://docs.stripe.com/webhooks/signature` |

## Supported Endpoints

| Method | Path | Purpose |
| --- | --- | --- |
| `POST` | `/stripe/v1/checkout/sessions` | Creates a deterministic Checkout Session. |
| `GET` | `/stripe/v1/checkout/sessions` | Lists local Checkout Sessions. |
| `GET` | `/stripe/v1/checkout/sessions/{id}` | Retrieves a local Checkout Session. |
| `POST` | `/stripe/v1/payment_intents` | Creates a deterministic PaymentIntent. |
| `GET` | `/stripe/v1/payment_intents` | Lists local PaymentIntents. |
| `GET` | `/stripe/v1/payment_intents/{id}` | Retrieves a local PaymentIntent. |
| `POST` | `/stripe/v1/customers` | Creates a deterministic Customer. |
| `GET` | `/stripe/v1/customers` | Lists local Customers. |
| `GET` | `/stripe/v1/customers/{id}` | Retrieves a local Customer. |
| `POST` | `/stripe/v1/products` | Creates a deterministic Product. |
| `GET` | `/stripe/v1/products` | Lists local Products. |
| `GET` | `/stripe/v1/products/{id}` | Retrieves a local Product. |
| `POST` | `/stripe/v1/prices` | Creates a deterministic Price. |
| `GET` | `/stripe/v1/prices` | Lists local Prices. |
| `GET` | `/stripe/v1/prices/{id}` | Retrieves a local Price. |
| `POST` | `/stripe/v1/subscriptions` | Creates a deterministic Subscription. |
| `GET` | `/stripe/v1/subscriptions` | Lists local Subscriptions. |
| `GET` | `/stripe/v1/subscriptions/{id}` | Retrieves a local Subscription. |
| `POST` | `/stripe/v1/invoices` | Creates a deterministic Invoice. |
| `GET` | `/stripe/v1/invoices` | Lists local Invoices. |
| `GET` | `/stripe/v1/invoices/{id}` | Retrieves a local Invoice. |
| `POST` | `/stripe/v1/refunds` | Creates a deterministic Refund. |
| `GET` | `/stripe/v1/refunds` | Lists local Refunds. |
| `GET` | `/stripe/v1/refunds/{id}` | Retrieves a local Refund. |
| `POST` | `/stripe/test/webhook/send` | Sends a fake signed webhook to the configured target URL. |
| `POST` | `/stripe/test/reset` | Clears local state and idempotency records for test isolation. |

## Scenarios

| Scenario | Behavior |
| --- | --- |
| `payment_success` | Default successful local workflow. |
| `payment_failed` | Returns Stripe-like card decline behavior for payment creation and webhook helper paths. |
| `auth_error` | Returns authentication-style failures. |
| `rate_limited` | Returns rate limit behavior. |
| `timeout` | Returns deterministic timeout behavior immediately (504 response shape). For real latency, add `X-Mockport-Delay: <ms>` (`0`–`30000`; invalid values return `400`). See [Adapters](../site/adapters.md). |

## Current Gaps And Tasks

| Priority | Task | Current source of truth |
| --- | --- | --- |
| P1 | Define selected Stripe workflows in `compat/manifests/stripe.json`, including explicit non-goals for fraud, Connect, disputes, tax, settlement, and full Billing lifecycle. | `tasks/phase27_stripe_provider_compatible_track.md` |
| P1 | Deepen SDK contract coverage for list pagination shape, retrieve-after-create, idempotency replay/conflict, and validation error envelopes. | `contract/sdk/stripe-smoke.test.js` and `compat/fixtures/stripe/` |
| P1 | Keep Stripe at `workflow-compatible` until manifest evidence and the provider-compatible promotion gate pass. | `tasks/phase26_provider_compatible_manifest_promotion.md` |
| P2 | Expand fixture coverage beyond the current auth, rate limit, idempotency, validation, checkout, and SDK major-surface examples. | `compat/fixtures/stripe/` |

## Verification

Run the adapter tests and SDK contract:

```bash
/usr/local/go/bin/go test ./adapters/stripe
bash scripts/run-sdk-contracts.sh stripe
```
