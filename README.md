# Mockport

Secret-free service emulation for AI-native development.

[![CI](https://github.com/albert-einshutoin/mockport/actions/workflows/ci.yml/badge.svg)](https://github.com/albert-einshutoin/mockport/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/albert-einshutoin/mockport)](https://github.com/albert-einshutoin/mockport/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

[日本語版](README.ja.md)

## Why Mockport

Run Stripe, OpenAI, Slack, GitHub OAuth, LINE, and Zoho OAuth integration code locally without giving real API keys to AI coding agents or CI. Point your SDK or HTTP client at `localhost` instead of the provider.

Unlike single-service mocks such as [stripe-mock](https://github.com/stripe/stripe-mock), Mockport runs multiple SaaS adapters in one Docker process with fake state, built-in error scenarios, webhook helpers, and secret-safe defaults. See [Comparison](docs/site/comparison.md) for a feature-by-feature view.

## 30-Second Quickstart

No local install required:

```bash
docker run --rm -p 127.0.0.1:43101:43101 \
  -v $(pwd)/examples/stripe-checkout/mockport.yml:/etc/mockport/mockport.yml \
  ghcr.io/albert-einshutoin/mockport:0.1.0-alpha \
  run --config /etc/mockport/mockport.yml --host 0.0.0.0
```

In another terminal:

```bash
curl http://localhost:43101/health
curl -X POST http://localhost:43101/stripe/v1/checkout/sessions
curl http://localhost:43101/_mockport/report
```

### From source

Build the binary and image from a checkout of this repository:

```bash
make build
docker build -t mockport:local -f docker/Dockerfile .
docker run --rm -p 127.0.0.1:43101:43101 \
  -v $(pwd)/examples/stripe-checkout/mockport.yml:/etc/mockport/mockport.yml \
  mockport:local run --config /etc/mockport/mockport.yml --host 0.0.0.0
```

Or run the empty-directory smoke test:

```bash
bash scripts/smoke-empty-dir.sh
```

## What It Looks Like

Health check:

```bash
$ curl http://localhost:43101/health
{"status":"ok"}
```

Stripe-like checkout session:

```bash
$ curl -X POST http://localhost:43101/stripe/v1/checkout/sessions
{"id":"stripe_checkout_session_000001","object":"checkout.session","payment_status":"paid"}
```

With the CLI installed, the same request and safety report renders as:

```bash
$ mockport report --url http://localhost:43101/_mockport/report
Mockport Report

Mode: ai-safe
Safety: safe=true real-looking-secrets=0 external-urls=0
Public env safe-to-commit: true

Adapters:
- stripe enabled at /stripe maturity=workflow-compatible

Requests:
- #1 POST /stripe/v1/checkout/sessions -> 200
```

## Supported Services

| Service | Adapter | Base path | Maturity | Supported workflows |
| --- | --- | --- | --- | --- |
| Stripe-like payments | `stripe` | `/stripe` plus SDK-compatible `/v1` alias | `workflow-compatible` | checkout sessions, payment intents, fake signed webhooks, … ([spec](docs/adapters/stripe.md)) |
| OpenAI-compatible API | `openai` | `/openai` | `workflow-compatible` | models, chat completions, streaming, embeddings, … ([spec](docs/adapters/openai.md)) |
| GitHub OAuth-like API | `github-oauth` | `/github` | `workflow-compatible` | authorize redirect, token exchange, user profile, … ([spec](docs/adapters/github-oauth.md)) |
| Slack-like messaging API | `slack` | `/slack` | `workflow-compatible` | auth test, conversations, message post/update/delete, … ([spec](docs/adapters/slack.md)) |
| LINE-like platform APIs | `line` | `/line` | `workflow-compatible` | Messaging API, LINE Login, LINE Pay, … ([spec](docs/adapters/line.md)) |
| Zoho OAuth-like API | `zoho-oauth` | `/zoho` | `workflow-compatible` | authorize redirect, token exchange, user info, … ([spec](docs/adapters/zoho-oauth.md)) |

Planned:

| Service | Planned adapter | Target workflows | Status |
| --- | --- | --- | --- |
| SendGrid-like email API | `sendgrid` | email send success/failure, auth error, rate limit, webhook events | Later candidate |

## SDK Connection

Point the official SDK at Mockport instead of the provider. See [Examples](docs/site/examples.md) and [OpenAI Chat example](examples/openai-chat/README.md).

```javascript
import OpenAI from "openai";

const client = new OpenAI({
  apiKey: "mockport_openai_key",
  baseURL: "http://localhost:43101/openai/v1",
});
```

Application env for Stripe (safe to commit when fake values stay unchanged):

```env
STRIPE_API_URL=http://localhost:43101/stripe
STRIPE_SECRET_KEY=mockport_stripe_secret
STRIPE_WEBHOOK_SECRET=whsec_mockport
```

See [Public Env Safety](docs/public-env-safety.md).

## AI-safe By Default

Mockport warns when real-looking credentials or real external service URLs are detected. In `strict` mode, unsafe configuration fails startup.

```bash
mockport run --config examples/unsafe-config/mockport.yml --check
```

See [AI-safe Development](docs/ai-safe-development.md) for warning categories, strict mode, and redaction behavior.

## Reports And Compatibility

Every run exposes `/_mockport/report` and `mockport report` with request history, scenario coverage, behavior matrix, and safety summary.

See [Reports](docs/site/reports.md) and [Support matrix](docs/site/support-matrix.md).

## Docs And Distribution

Full docs, install paths, and release verification live under [docs/site/](docs/site/index.md). Current preview: `v0.1.0-alpha` via [Docker / GHCR](docs/site/distribution.md) and [GitHub release archives](docs/site/distribution.md). The npm wrapper is experimental; Go binary and Docker remain primary.

> **⚠️ Archive**: Pre-implementation design documents (May 2026) are preserved in [docs/archive/design/](docs/archive/design/README.md). They are not maintained and may diverge from the current implementation.

## Contribute

Contributions follow spec-first TDD. See [CONTRIBUTING.md](CONTRIBUTING.md), [Adapter onboarding guide](docs/adding-an-adapter.md), [Maintainer Guide](docs/maintainer-guide.md), [Roadmap](ROADMAP.md), and [Support Policy](docs/public-support-policy.md).

Mockport is not a full clone of external services. It focuses on local integration testing scenarios: success, failure, auth error, rate limit, timeout, and webhook/callback.

---

[Quickstart](docs/site/quickstart.md) · [Docs](docs/site/index.md) · [Examples](docs/site/examples.md) · [Roadmap](ROADMAP.md) · [日本語版](README.ja.md)
