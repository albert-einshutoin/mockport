# Stripe Checkout Example

[日本語版](README.ja.md)

This example runs the Stripe-like Mockport adapter with fake local credentials.

```bash
docker build -t mockport:local -f docker/Dockerfile .
docker compose -f examples/stripe-checkout/docker-compose.yml up
```

Use these values in the application under test:

```env
STRIPE_API_URL=http://localhost:43101/stripe
STRIPE_SECRET_KEY=mockport_stripe_secret
STRIPE_WEBHOOK_SECRET=whsec_mockport
```

Smoke test:

```bash
curl -X POST http://localhost:43101/stripe/v1/checkout/sessions
curl http://localhost:43101/_mockport/report
```
