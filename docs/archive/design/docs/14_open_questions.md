# 14. Open Questions

[日本語版](14_open_questions.ja.md)

## Naming

Current candidate:

```txt
Mockport
```

Need to verify:

- GitHub repository availability
- npm package availability
- Docker image namespace
- domain availability
- trademark risk

## First adapter

Recommended:

```txt
stripe
```

Alternative:

```txt
openai
```

Stripe is better for proving webhook/payment/secret-free integration testing. OpenAI is better for AI-native messaging.

## Adapter granularity

Options:

1. all-in-one Docker image
2. adapter-specific Docker images
3. core image with remote plugins

Recommendation:

Start with all-in-one.

## Config DSL complexity

Start simple. Do not build a complex DSL before usage patterns are clear.

## Compatibility score

Question:

How should Mockport represent "how close" an adapter is to a real service?

Possible levels:

```txt
experimental
partial
common-path
contract-tested
sandbox-verified
```

## External network calls

Should Mockport ever proxy to real providers?

Recommendation:

- MVP: no
- later: recording mode with explicit opt-in
- strict AI-safe mode: never

## npm package

Question:

Should Mockport have an npm wrapper?

Recommendation:

Not needed for minimal MVP. Add after Docker/Go binary works.

Potential future:

```bash
npx mockport init
```

## Rust component

Possible future Rust module:

- compatibility diff
- high-performance request replay
- HAR/OpenAPI parser
- streaming response engine

Not needed in MVP.
