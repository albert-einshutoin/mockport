> **⚠️ Archive notice — Not maintained, may diverge, do not cite as authoritative.**
>
> Pre-implementation design archive. This is **not** the authoritative source for current implementation.
> For current specs see [docs/site/](../../../site/index.md).

# 12. Roadmap

[日本語版](12_roadmap.ja.md)

## Phase 0: Baseline

Goal: repository skeleton.

Tasks:

- create Go module
- create cmd/mockport
- add Cobra CLI
- add config loader
- add `/health`
- add Dockerfile
- add CI

Exit evidence:

- `go test ./...` passes
- Docker image builds
- `/health` returns 200

## Phase 1: Stripe minimal adapter

Goal: prove service emulation value.

Tasks:

- Stripe adapter
- checkout session response
- payment intent response
- success/failure scenarios
- webhook sender
- fake env generator

Exit evidence:

- local app can point `STRIPE_API_URL` to Mockport
- success/failure responses work
- webhook can be sent to local target

## Phase 2: CLI UX

Goal: make the project usable.

Tasks:

- `mockport init`
- `.env.mockport` generation
- `mockport.yml` generation
- `docker-compose.mockport.yml` generation
- `mockport run`

Exit evidence:

- user can start from empty directory and run Mockport in under 2 minutes

## Phase 3: AI-safe mode

Goal: differentiate Mockport.

Tasks:

- real-looking secret detection
- external URL detection
- strict mode
- report
- redaction
- AI-safe documentation

Exit evidence:

- startup warns/fails on real-looking secrets
- report marks safe/unsafe state

## Phase 4: Additional adapters

Goal: become broadly useful.

Adapters:

- OpenAI-compatible
- GitHub OAuth
- LINE or Slack

Exit evidence:

- examples exist for each adapter
- scenarios cover success/failure/rate limit/auth error

## Phase 5: Compatibility and reports

Goal: become trustworthy.

Tasks:

- scenario coverage report
- unsupported endpoint report
- request replay log
- behavior matrix
- adapter maturity levels

Exit evidence:

- report explains what is supported and unsupported

## Phase 6: Distribution

Goal: improve adoption.

Tasks:

- GHCR Docker image
- Homebrew tap
- npm wrapper
- GitHub release binaries
- documentation site

Exit evidence:

- install options documented and tested

## Long-term

- adapter-specific Docker images
- contract testing against provider sandbox
- recording/replay mode
- OpenAPI/HAR import
- compatibility score
- plugin SDK
