# 03. Architecture

[日本語版](03_architecture.ja.md)

## High-level architecture

```txt
Application
  |
  | HTTP request
  v
Mockport Docker container
  |
  +-- adapter registry
  +-- scenario runtime
  +-- fake secret validator
  +-- webhook sender
  +-- request recorder
  +-- report generator
```

## Runtime model

Mockport is a Docker-first HTTP service.

The application does not import Mockport as a library. Instead, it changes service URLs:

```env
STRIPE_API_URL=http://mockport:43101/stripe
STRIPE_SECRET_KEY=mockport_stripe_secret
```

This keeps Mockport language-agnostic.

## Core components

### CLI

Responsible for:

- init
- config generation
- detect
- add adapter config
- docker compose generation
- local run

### Server

Responsible for:

- HTTP server lifecycle
- route registration
- middleware
- health check
- report endpoint

### Adapter registry

Responsible for:

- registering built-in adapters
- enabling/disabling adapters based on config
- mapping base paths to adapters

### Scenario runtime

Responsible for:

- selecting response scenario
- returning configured responses
- applying latency/timeouts
- generating failure responses

### AI-safe guard

Responsible for:

- detecting real-looking secrets
- warning or blocking in strict mode
- preventing accidental real service URLs in ai-safe mode

### Request recorder

Responsible for:

- recording incoming requests
- capturing response status
- producing report data

### Webhook sender

Responsible for:

- generating provider-like webhook payloads
- signing payloads with fake local signing secrets
- sending to configured local target URL

## Minimal request flow

```txt
POST /stripe/v1/checkout/sessions
  -> server middleware
  -> request recorder
  -> stripe adapter route
  -> scenario runtime
  -> response builder
  -> report updated
```

## Adapter registration model

Minimal interface:

```go
type Adapter interface {
    Name() string
    Register(mux *http.ServeMux, cfg AdapterConfig) error
    FakeEnv(cfg AdapterConfig) map[string]string
}
```

This is intentionally small.

## Why not dynamic plugins first?

Go dynamic plugins create distribution and OS compatibility complexity. The MVP should prioritize:

- clear behavior
- small Docker image
- simple code
- easy contribution
- reliable CI

Dynamic plugins can be revisited after product-market validation.
