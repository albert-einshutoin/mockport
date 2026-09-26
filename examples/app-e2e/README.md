# Stripe and OpenAI app E2E

This example uses one local Mockport process. It exercises an application's HTTP entry point, the pinned official provider SDKs, Mockport, and the application's result. The keys below are fake values. It does not evaluate AI answer quality or process real payments.

## Selected Stripe contract

| Step | Order state | Event | Allowed business effect |
| --- | --- | --- | --- |
| Create and retrieve Checkout Session | `pending` | none | Store one Session ID and the order's `client_reference_id`; Session `payment_status=unpaid`, no grant. |
| Paid delivery | `paid` | `checkout.session.completed` with `payment_status=paid` | Mark the matching order paid and increment `paid_updates` once. |
| Replay of the same event ID | unchanged | same event ID and body, freshly signed at delivery time | No further payment update. |
| Unpaid delivery | `failed` | `checkout.session.async_payment_failed` with `payment_status=unpaid` | No payment update or grant. |
| Invalid signature or changed body | unchanged | rejected before event processing | No state change. |

The selected Session ID and `client_reference_id` are read from Mockport's created resource. The app verifies the raw request body with `stripe.webhooks.constructEvent`, then checks both identifiers. The test setup calls Mockport's loopback-only send helper; the app does not expose scenario or reset controls. Event IDs are explicit and reproducible. Signature timestamps are fresh for each delivery, including a replay, because the official SDK enforces a time tolerance. Mockport does not deduplicate deliveries.

The Node app uses `stripe@22.3.1`, matching the existing [Node SDK example](../node-sdk-clients/stripe.mjs). Its API host is restricted to local Mockport names; it has no real-provider fallback. `auth_required` is enabled in both E2E configs.

## Run locally

Prerequisites: Go from `PATH`, Node 24, npm, curl. From the repository root:

```sh
bash scripts/run-app-e2e.sh stripe
bash scripts/check-stripe-app-mutation.sh
```

The first command builds the current checkout into a temporary binary, installs the pinned SDK from `package-lock.json`, starts Mockport and the app, runs the assertions, and stops both. It uses ports 43101 and 33001 by default; `MOCKPORT_APP_E2E_PORT` and `STRIPE_APP_E2E_PORT` override them. The second command runs the same E2E against a temporary source mutation that removes the event ID replay guard. The assertion must fail with `processed_events:2`; the original app must then pass. No faulty app source is kept.

## Run with Compose

Prerequisites: Docker with Compose. From the repository root:

```sh
docker compose -f examples/app-e2e/compose.yml -p mockport-app-e2e up --build --abort-on-container-exit --exit-code-from runner
docker compose -f examples/app-e2e/compose.yml -p mockport-app-e2e down --remove-orphans
```

Compose builds Mockport from this checkout and installs the pinned Node SDK while building the app image. The app calls `http://mockport:43101`; Mockport sends to `http://app:33001`. The test runner shares Mockport's network namespace and calls its management helper over `127.0.0.1`, preserving the loopback guard. The Compose network has `internal: true`, so containers cannot route to external services during the test after image build and dependency retrieval. Host ports are loopback-bound. The test checks `request_history.truncated=false` with fewer than 500 requests and does not infer unsupported coverage from an empty truncated history.

CI runs the local Stripe command on every PR and push. The Docker Compose run is an additional network-isolated smoke check. The selected flow is an example contract, not a claim of full Stripe compatibility.
