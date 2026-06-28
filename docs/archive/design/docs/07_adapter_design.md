# 07. Adapter Design

[日本語版](07_adapter_design.ja.md)

## Adapter goal

An adapter emulates a specific external service or a subset of that service.

Examples:

- stripe
- github-oauth
- openai
- line-messaging
- slack
- sendgrid

## Adapter philosophy

Adapters should focus on developer workflows, not complete cloning.

Each adapter should support:

- common success path
- common failure path
- auth error
- rate limit
- timeout
- webhook/callback if applicable
- report metadata

## Minimal Go interface

```go
type Adapter interface {
    Name() string
    Register(mux *http.ServeMux, cfg AdapterConfig) error
    FakeEnv(cfg AdapterConfig) map[string]string
}
```

## Future interface

After several adapters exist, consider:

```go
type Adapter interface {
    Name() string
    Version() string
    Capabilities() []Capability
    Register(router Router, cfg AdapterConfig) error
    FakeEnv(cfg AdapterConfig) map[string]string
    Scenarios() []ScenarioDefinition
}
```

## Built-in adapter registration

```go
func RegisterDefaults(r *adapter.Registry) {
    r.Register(stripe.New())
}
```

## Stripe minimal adapter

### Endpoints

```txt
POST /stripe/v1/checkout/sessions
GET  /stripe/v1/checkout/sessions/{id}
POST /stripe/v1/payment_intents
GET  /stripe/v1/payment_intents/{id}
POST /stripe/test/webhook/send
```

### Scenarios

```txt
payment_success
payment_failed
auth_error
rate_limited
timeout
```

### Fake env

```env
STRIPE_API_URL=http://localhost:43101/stripe
STRIPE_SECRET_KEY=mockport_stripe_secret
STRIPE_WEBHOOK_SECRET=whsec_mockport
```

## OpenAI adapter target

Useful later because many AI-native apps need API-compatible local responses.

Potential endpoints:

```txt
POST /openai/v1/chat/completions
POST /openai/v1/responses
GET  /openai/v1/models
```

Scenarios:

```txt
chat_success
stream_success
rate_limited
context_length_exceeded
auth_error
```

## GitHub OAuth adapter target

Potential endpoints:

```txt
GET  /github/login/oauth/authorize
POST /github/login/oauth/access_token
GET  /github/user
```

Scenarios:

```txt
oauth_success
invalid_code
expired_token
scope_missing
```

## LINE/Slack adapter target

Potential features:

- message push/reply
- webhook signature verification
- delivery failure
- rate limit
- retry
