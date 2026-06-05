# 00. Product Vision

[日本語版](00_product_vision.ja.md)

## Project name

Mockport

## One-line description

Mockport is a Docker-first service emulator for testing external integrations without real secrets.

## Tagline candidates

- Run fake ports for real-world services.
- Test external integrations locally. No real secrets. No real external calls.
- Secret-free service emulation for AI-native development.

## Why now

AI coding tools can read files, run commands, inspect logs, run tests, and execute scripts. Even if `.env` is ignored or direct commands like `cat .env` are denied, secrets can still leak through:

- `process.env`
- test logs
- docker compose expansion
- generated artifacts
- stack traces
- package scripts
- local integration commands
- accidental `.env.example` misuse

Mockport addresses this by making real secrets unnecessary for local, CI, and AI-assisted integration testing.

## Product philosophy

Mockport follows three principles:

1. **Secret-free by default**
   - Local and CI tests must work with fake credentials.
   - Real-looking credentials should trigger warnings.
   - The default mode should never call real external services.

2. **Docker-first, language-agnostic**
   - Applications should integrate through URLs and environment variables, not SDK-specific test hooks.
   - The app should point to Mockport instead of the real service.

3. **Scenario-driven compatibility, not full cloning**
   - Mockport does not aim to perfectly clone every external service.
   - It focuses on the most important developer workflows: success, failure, webhook, auth error, rate limit, timeout, and retry.

## Primary user

- Backend engineers
- Full-stack developers
- AI-assisted coding users
- OSS maintainers
- CI/CD maintainers
- Teams that want safer external integration testing

## Core promise

With Mockport, a developer should be able to change:

```env
STRIPE_API_URL=https://api.stripe.com
STRIPE_SECRET_KEY=sk_live_xxx
```

to:

```env
STRIPE_API_URL=http://localhost:43101/stripe
STRIPE_SECRET_KEY=mockport_stripe_secret
```

and run integration tests without exposing real credentials.
