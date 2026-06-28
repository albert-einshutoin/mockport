# Mockport

[日本語版](13_readme_draft.ja.md)

Secret-free service emulation for AI-native development.

Mockport lets you run local Docker-based emulators for external services like Stripe, GitHub OAuth, LINE, Slack, SendGrid, and OpenAI.

Point your app to Mockport instead of real services, use local fake secrets, and test payments, OAuth, webhooks, notifications, and AI APIs safely in local development, CI, and AI coding environments.

## Why Mockport?

AI coding tools can accidentally inspect secrets through files, logs, tests, scripts, or environment variables.

Mockport removes the need for real secrets in local integration testing.

## Quickstart

```bash
mockport init --adapter stripe
docker compose -f docker-compose.mockport.yml up
```

Application `.env`:

```env
STRIPE_API_URL=http://localhost:43101/stripe
STRIPE_SECRET_KEY=mockport_stripe_secret
STRIPE_WEBHOOK_SECRET=whsec_mockport
```

Test a Stripe-like checkout session:

```bash
curl -X POST http://localhost:43101/stripe/v1/checkout/sessions
```

## Docker

```bash
docker run --rm -p 127.0.0.1:43101:43101 \
  -v $(pwd)/mockport.yml:/etc/mockport/mockport.yml \
  ghcr.io/albert-einshutoin/mockport:0.1.0-alpha run --config /etc/mockport/mockport.yml --host 0.0.0.0
```

## Supported adapters

MVP:

- Stripe-like payment adapter

Planned:

- OpenAI-compatible API
- GitHub OAuth
- LINE Messaging
- Slack
- SendGrid

## AI-safe by default

Mockport warns when real-looking credentials are detected:

```txt
[MOCKPORT SECURITY WARNING]
STRIPE_SECRET_KEY looks like a real Stripe key.
Use mockport_stripe_secret instead.
```

## Scope

Mockport is not a full clone of external services.

It focuses on local integration testing scenarios:

- success
- failure
- auth error
- rate limit
- timeout
- webhook/callback

## License

MIT
