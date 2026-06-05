# Mockport

Secret-free service emulation for AI-native development.

[日本語版](README.ja.md)

Mockport runs local Docker-based emulators for external services. The Minimal MVP supports a Stripe-like payment adapter so local development, CI, and AI coding workflows can test payment integration paths without real Stripe secrets.

For public preview scope, support matrix, examples, and limitations, see [Mockport Docs](docs/site/index.md).

Project planning and maintenance:

- [Roadmap](ROADMAP.md)
- [Maintainer Guide](docs/maintainer-guide.md)
- [Support Policy](docs/public-support-policy.md)

## Quickstart

No local install required:

```bash
docker build -t mockport:local -f docker/Dockerfile .
docker run --rm -p 127.0.0.1:43101:43101 \
  -v $(pwd)/configs/mockport.example.yml:/etc/mockport/mockport.yml \
  mockport:local run --config /etc/mockport/mockport.yml --host 0.0.0.0
```

Then verify the local API:

```bash
curl http://localhost:43101/health
curl -X POST http://localhost:43101/stripe/v1/checkout/sessions
curl http://localhost:43101/_mockport/report
```

CLI workflow from a built binary:

```bash
mockport init --adapter stripe
docker compose -f docker-compose.mockport.yml up
```

`mockport init` protects existing generated files by default. Use `--force` only when you intentionally want to replace `mockport.yml`, `.env.mockport`, and `docker-compose.mockport.yml`.

Application `.env`:

```env
STRIPE_API_URL=http://localhost:43101/stripe
STRIPE_SECRET_KEY=mockport_stripe_secret
STRIPE_WEBHOOK_SECRET=whsec_mockport
```

This Mockport env is safe to commit when the generated fake values remain unchanged. See [Public Env Safety](docs/public-env-safety.md).

Test a Stripe-like checkout session:

```bash
curl -X POST http://localhost:43101/stripe/v1/checkout/sessions
```

View the request and safety report:

```bash
mockport report
```

## Source Smoke Test

From a checkout of this repository:

```bash
bash scripts/smoke-empty-dir.sh
```

The smoke test builds the local Docker image, creates a temporary empty directory, runs `mockport init --adapter stripe`, starts Docker Compose, checks `/health`, posts a Stripe-like checkout request, and prints `mockport report`.

## Docker

```bash
docker run --rm -p 127.0.0.1:43101:43101 \
  -v $(pwd)/configs/mockport.example.yml:/etc/mockport/mockport.yml \
  mockport:local run --config /etc/mockport/mockport.yml --host 0.0.0.0
```

## Install And Distribution

Mockport is Docker-first. The first public preview is `v0.1.0-alpha`.

Docker preview image:

```bash
docker pull ghcr.io/albert-einshutoin/mockport:0.1.0-alpha
docker run --rm -p 127.0.0.1:43101:43101 \
  -v $(pwd)/configs/mockport.example.yml:/etc/mockport/mockport.yml \
  ghcr.io/albert-einshutoin/mockport:0.1.0-alpha run --config /etc/mockport/mockport.yml --host 0.0.0.0
```

Release archives:

```bash
curl -LO https://github.com/albert-einshutoin/mockport/releases/download/v0.1.0-alpha/mockport_0.1.0-alpha_darwin_arm64.tar.gz
curl -LO https://github.com/albert-einshutoin/mockport/releases/download/v0.1.0-alpha/checksums.txt
grep 'mockport_0.1.0-alpha_darwin_arm64.tar.gz' checksums.txt | sed 's# dist/# #' | shasum -a 256 -c -
tar -xzf mockport_0.1.0-alpha_darwin_arm64.tar.gz
./mockport_0.1.0-alpha_darwin_arm64/mockport version
```

| Channel | Status | Notes |
| --- | --- | --- |
| Docker / GHCR | Preview | `ghcr.io/albert-einshutoin/mockport:0.1.0-alpha`; `latest` follows the default branch and is not the preview release contract |
| GitHub release archives | Preview | `mockport_<version>_<os>_<arch>.tar.gz` with `checksums.txt` |
| Homebrew | Not published | Formula template is under `packaging/homebrew/` |
| npm | Not published | The npm wrapper is experimental; Go binary and Docker remain primary |

Docs site source lives under `docs/site/`.

## Services

Supported:

| Service | Adapter | Base path | Maturity | Supported workflows |
| --- | --- | --- | --- | --- |
| Stripe-like payments | `stripe` | `/stripe` plus SDK-compatible `/v1` alias | `workflow-compatible` | checkout sessions, payment intents, customers, products, prices, subscriptions, invoices, refunds, fake signed webhooks, SDK contract, state, validation, idempotency |
| OpenAI-compatible API | `openai` | `/openai` | `workflow-compatible` | models, chat completions, responses, streaming, embeddings, files, batches, SDK contract, state, validation |
| GitHub OAuth-like API | `github-oauth` | `/github` | `workflow-compatible` | authorize redirect, access token exchange, user profile, user emails, user orgs, client contract, state, scope validation |
| Slack-like messaging API | `slack` | `/slack` | `workflow-compatible` | auth test, conversations list/history, message post/update/delete, Events API URL verification/message callback subset, client contract, state, Slack-style errors |
| LINE-like platform APIs | `line` | `/line` | `workflow-compatible` | Messaging API send/content/signed webhook/rich menu/channel token workflows, LINE Login OAuth/profile, LIFF helpers, MINI App service messages, LINE Pay v3 request/confirm, Mini Dapp wallet/payment helpers |

Planned:

| Service | Planned adapter | Target workflows | Status |
| --- | --- | --- | --- |
| SendGrid-like email API | `sendgrid` | email send success/failure, auth error, rate limit, webhook events | Later candidate |

## AI-safe By Default

Mockport warns when real-looking credentials or real external service URLs are detected. In `strict` mode, unsafe configuration fails startup.

Check a config without starting the server:

```bash
mockport run --config examples/unsafe-config/mockport.yml --check
```

See [AI-safe Development](docs/ai-safe-development.md) for warning categories, strict mode, and redaction behavior.

Mockport is not a full clone of external services. It focuses on local integration testing scenarios: success, failure, auth error, rate limit, timeout, and webhook/callback.
