# Stripe and OpenAI app E2E

This example uses one local Mockport process. It exercises an application's HTTP entry point, the pinned official provider SDKs, Mockport, and the application's result. The keys below are fake values. It does not evaluate AI answer quality or process real payments.

## Selected Stripe contract

| Step | Order state | Event | Allowed business effect |
| --- | --- | --- | --- |
| Create and retrieve Checkout Session | `pending` | none | Store one Session ID and the order's `client_reference_id`; Session `payment_status=unpaid`, no grant. |
| Paid delivery | `paid` | `checkout.session.completed` with `payment_status=paid` | The app retrieves the matching paid Session through the official SDK inside the webhook handler, then marks the order paid and increments `paid_updates` once. |
| Replay of the same event ID | unchanged | same event ID and body, freshly signed at delivery time | No further payment update. |
| Unpaid delivery | `failed` | `checkout.session.async_payment_failed` with `payment_status=unpaid` | No payment update or grant. |
| Invalid signature or changed body | unchanged | rejected before event processing | No state change. |
| Late failure after payment | `paid` | `checkout.session.async_payment_failed` | The saved Session and order stay paid. |

The selected Session ID and `client_reference_id` are read from Mockport's created resource. Creation uses one inline price item, so no real key or pre-created Price is needed. The app verifies the raw request body with `stripe.webhooks.constructEvent`, then checks both identifiers. Mockport stores paid before delivering the completed event; the app verifies that state through an SDK retrieve during webhook handling, and the E2E retrieves it again afterward. The test setup calls Mockport's loopback-only send helper; the app does not expose scenario or reset controls. Event IDs are explicit and reproducible. Signature timestamps are fresh for each delivery, including a replay, because the official SDK enforces a time tolerance. Mockport does not deduplicate deliveries.

The Node app uses `stripe@22.3.1`, matching the existing [Node SDK example](../node-sdk-clients/stripe.mjs). Its API host is restricted to local Mockport names; it has no real-provider fallback. `auth_required` is enabled in both E2E configs.

## Selected OpenAI contract

| Case | App HTTP result | Provider calls | Limit |
| --- | --- | --- | --- |
| Chat nonstreaming | `200`, text `Mockport response` | One successful completion | No answer-quality claim. |
| Chat streaming | `200`, assembled `Mockport simulated streaming response.`, multiple chunks, completed iteration | One SSE completion | The official Python SDK consumes the terminal marker. |
| Wrong fake key | `401`, `invalid_api_key` | One Mockport 401 | Separate from the forced `auth_error` scenario. |
| Missing key | `503`, `missing_key_before_request` | Zero calls; SDK refuses construction | This is local configuration failure, not adapter authentication evidence. |
| Persistent 429 | `429`, `rate_limited` | Two Mockport 429 responses | SDK `max_retries=1`; finite failure within four seconds. |
| Slow response and cancelled caller | `504`, `upstream_timeout`, or an abandoned HTTP response | SDK timeout `0.4s`, at most two attempts | Handler active count returns to zero within four seconds. |

The Python app uses the pinned official `openai==2.46.0` SDK, as in the [existing Python SDK example](../python-openai/example.py), and its public `max_retries` and `timeout` options. A test-only local forwarding proxy injects built-in scenario or delay headers for the 429 and timeout cases. The app receives ordinary HTTP and has no scenario endpoint or real-provider fallback. The test asserts the app response and Mockport report request count; fixed text only checks transport and control behavior.

## Run locally

Prerequisites: Go from `PATH`, Node 24, npm, curl. From the repository root:

```sh
bash scripts/run-app-e2e.sh stripe
bash scripts/check-stripe-app-mutation.sh
bash scripts/run-app-e2e.sh all
bash scripts/check-openai-app-mutation.sh
```

The Stripe command builds the current checkout into a temporary binary, installs the pinned SDK from `package-lock.json`, starts Mockport, a local test proxy, and the app, runs the assertions, and stops them. It uses ports 43101, 43103, and 33001 by default; `MOCKPORT_APP_E2E_PORT`, `STRIPE_TEST_PROXY_PORT`, and `STRIPE_APP_E2E_PORT` override them. The proxy delays creation to expose parallel requests and injects one failed creation to verify retry of the same order ID. The runner fails if a chosen port is already occupied. The Stripe mutation command removes the event ID replay guard from a temporary source copy. The assertion must fail with `processed_events:2` while `paid_updates` stays one; the original app must then pass.

The `all` command adds the Python app using the pinned dependency lock and runs both apps against the same Mockport process. Its local test proxy uses port 43102 and the Python app uses 33002 by default. The OpenAI mutation command changes the Python app's 429 handling in a temporary copy; the same E2E must reject the false 200 and then pass with original source. No faulty app source is kept.

## Run with Compose

Prerequisites: Docker with Compose. From the repository root:

```sh
docker compose -f examples/app-e2e/compose.yml -p mockport-app-e2e up --build --abort-on-container-exit --exit-code-from runner
docker compose -f examples/app-e2e/compose.yml -p mockport-app-e2e down --remove-orphans
```

Compose builds Mockport from this checkout and installs the pinned Node and Python SDKs while building the app images. Both apps call `http://mockport:43101`; Mockport sends Stripe webhooks to `http://app:33001`. The test runner shares Mockport's network namespace and calls its management helper over `127.0.0.1`, preserving the loopback guard. One runner tests Stripe and OpenAI success/streaming against the same Mockport process. The Compose network has `internal: true`, so containers cannot route to external services during the test after image build and dependency retrieval. Host ports are loopback-bound. The test checks `request_history.truncated=false` with fewer than 500 requests and does not infer unsupported coverage from an empty truncated history. Run the local `all` command for the OpenAI error, retry, timeout, and cancelled-caller cases.

CI runs the local `all` command on every PR and push. The Docker Compose run is an additional network-isolated smoke check. These selected flows are example contracts, not full provider compatibility claims.
