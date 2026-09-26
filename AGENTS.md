# Repository verification commands

Follow the user's task instructions for scope and approvals. These commands are defined in the current repository.

| Command | Coverage |
| --- | --- |
| `make verify` | Go vet, tests, race tests, public trust, and adapter completeness. |
| `bash scripts/check-go-engineering.sh` | Go engineering checks. |
| `bash scripts/run-sdk-contracts.sh all` | Pinned official SDK contracts against a local Mockport binary. |
| `bash scripts/run-app-e2e.sh stripe` | Node Stripe app HTTP, official SDK, webhook, and business state. |
| `bash scripts/run-app-e2e.sh all` | One Mockport process for Node Stripe and Python OpenAI app HTTP E2E; includes key, retry, timeout, and cancellation cases. |
| `bash scripts/check-stripe-app-mutation.sh` | Proves the Stripe E2E detects duplicate event processing; the separate order guard keeps paid updates at one. |
| `bash scripts/check-openai-app-mutation.sh` | Proves the OpenAI E2E rejects an app that converts persistent 429 to success. |
| `bash scripts/check-public-env.sh` | Public example secret and URL safety. |
| `bash scripts/check-compat-manifests.sh` | Compatibility manifest checks. |
| `bash scripts/check-distribution.sh` | Distribution checks. |
| `bash scripts/check-maintenance-policy.sh` | Maintenance policy checks. |
| `$(go env GOPATH)/bin/govulncheck ./...` | Go vulnerability scan after installing `golang.org/x/vuln/cmd/govulncheck@v1.3.0`. |
| `docker build -f docker/Dockerfile -t mockport:local .` | Build the pinned Go toolchain image. |
| `docker compose -f examples/app-e2e/compose.yml -p mockport-app-e2e up --build --abort-on-container-exit --exit-code-from runner` | Network-isolated Compose app smoke. Run `docker compose -f examples/app-e2e/compose.yml -p mockport-app-e2e down --remove-orphans` afterward. |

The `CI` workflow runs the full Go, SDK, public safety, compatibility, distribution, and maintenance checks on pushes and pull requests. The Stripe app E2E is an additional step in that job. The planned `ci-pr` selector belongs to draft PR #366 and is not yet on main.
