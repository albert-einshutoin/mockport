> **⚠️ Archive notice — Not maintained, may diverge, do not cite as authoritative.**
>
> Pre-implementation design archive. This is **not** the authoritative source for current implementation.
> For current specs see [docs/site/](../../../site/index.md).

# 02. MVP Scope

[日本語版](02_mvp_scope.ja.md)

Mockport should be built in two MVP layers:

1. Minimal MVP
2. Product Goal MVP

## Minimal MVP

The minimal MVP proves the core idea with one service adapter.

### Target adapter

Stripe-like adapter.

### Minimal MVP features

- Go HTTP server
- Docker image
- YAML config
- `/health` endpoint
- Stripe-like checkout session success response
- Stripe-like payment failure response
- webhook send endpoint
- fake secret generation
- AI-safe warning for real-looking secrets
- request report endpoint
- README quickstart
- GitHub Actions test/build/docker build

### Minimal endpoints

```txt
GET  /health
GET  /_mockport/report

POST /stripe/v1/checkout/sessions
GET  /stripe/v1/checkout/sessions/{id}
POST /stripe/v1/payment_intents
GET  /stripe/v1/payment_intents/{id}
POST /stripe/test/webhook/send
```

### Minimal scenarios

```txt
payment_success
payment_failed
auth_error
rate_limited
timeout
```

### Minimal exit criteria

- `go test ./...` passes
- `go build ./cmd/mockport` passes
- Docker image builds
- `docker run -p 127.0.0.1:43101:43101 ...` starts the server
- `/health` returns 200
- Stripe-like success scenario returns 200
- Stripe-like failure scenario returns 402
- webhook sender can POST to a configured target URL
- report shows requests and safety warnings
- documentation includes quickstart

## Product Goal MVP

The product goal MVP should feel useful as an OSS project.

### Target adapters

- Stripe-like payment adapter
- GitHub OAuth-like adapter
- OpenAI-compatible API adapter
- LINE Messaging or Slack-like notification adapter

### Product Goal MVP features

- `mockport init`
- `mockport run`
- `mockport detect`
- `mockport add`
- `mockport report`
- `.env.mockport` generation
- `docker-compose.mockport.yml` generation
- AI-safe mode
- real-looking secret detection
- adapter development guide
- compatibility report
- scenario coverage report
- examples for Node/NestJS/Express

## Important MVP rule

Do not implement a complex plugin system in the minimal MVP.

Start with built-in adapters:

```go
registry.Register(stripe.New())
registry.Register(openai.New())
```

Adapter splitting and dynamic plugin loading can be postponed.
