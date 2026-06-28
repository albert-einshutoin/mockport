> **⚠️ Archive notice — Not maintained, may diverge, do not cite as authoritative.**
>
> Pre-implementation design archive. This is **not** the authoritative source for current implementation.
> For current specs see [docs/site/](../../../site/index.md).

# 01. Problem Statement

[日本語版](01_problem_statement.ja.md)

## Problem

Modern applications depend on external services:

- payments
- OAuth login
- notification APIs
- AI APIs
- storage APIs
- webhook providers
- email providers

Developers want to test these integrations locally and in CI, but real credentials are risky.

AI coding tools make this worse because they can inspect files, run commands, execute test suites, and observe logs.

## Why existing measures are not enough

### `.env` ignore rules are insufficient

`.env` can be excluded from Git and AI context, but secrets can still leak through:

- application logs
- Docker environment interpolation
- generated configs
- failed tests
- error messages
- `printenv`
- `process.env`
- package scripts

### hooks and deny commands are insufficient

Blocking `cat .env` does not block:

```bash
npm test
docker compose config
node -e "console.log(process.env)"
grep -R API_KEY .
```

### secret managers are not enough

Secret managers help with rotation, audit logs, and centralized access control, but if the AI-executed process has permission to fetch the secret, the secret can still be exposed.

## Desired state

Developers should be able to test external integrations without:

- storing real secrets locally
- giving secrets to AI agents
- calling real external APIs in CI
- manually faking webhook behavior
- rewriting application code for each service

## Mockport solution

Mockport runs local Docker-based service emulators. Applications point to Mockport URLs and use fake credentials.

Mockport returns service-like responses and can simulate webhooks, OAuth flows, rate limits, failures, and timeouts.

## Non-goals

- Full behavioral clone of every provider
- Replacement for real sandbox environments
- Production proxy
- Secret manager replacement
- WAF/firewall/security agent
