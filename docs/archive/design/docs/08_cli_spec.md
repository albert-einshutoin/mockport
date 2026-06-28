# 08. CLI Spec

[日本語版](08_cli_spec.ja.md)

## CLI name

```bash
mockport
```

## Commands

```txt
mockport init
mockport run
mockport detect
mockport add
mockport up
mockport report
mockport version
```

## Minimal MVP commands

```txt
mockport init
mockport run
mockport version
```

## `mockport init`

Generates:

```txt
mockport.yml
.env.mockport
docker-compose.mockport.yml
```

Example:

```bash
mockport init --adapter stripe
```

## `mockport run`

Runs Mockport directly as a local process.

```bash
mockport run --config mockport.yml
```

## `mockport detect`

Scans local project files:

- `.env.example`
- `.env.local.example`
- `package.json`
- `docker-compose.yml`
- source code strings
- framework config files

Example output:

```txt
Detected integrations:

- Stripe
  Evidence:
    - STRIPE_SECRET_KEY in .env.example
    - stripe package in package.json
  Suggested adapter:
    stripe

- OpenAI
  Evidence:
    - OPENAI_API_KEY in .env.example
  Suggested adapter:
    openai
```

## `mockport add`

Adds adapter config:

```bash
mockport add stripe openai
```

## `mockport up`

Runs generated Docker Compose.

```bash
mockport up
```

Initial implementation may simply call:

```bash
docker compose -f docker-compose.mockport.yml up
```

## `mockport report`

Displays request and safety report:

```txt
Mockport Report

Mode: ai-safe

Adapters:
- stripe enabled at /stripe

Requests:
- POST /stripe/v1/checkout/sessions -> 200

Safety:
- Real-looking secrets detected: 0
- External live URLs detected: 0
```

## CLI implementation

Use Cobra.

Package:

```txt
internal/cli
```

`cmd/mockport/main.go` must stay thin.
