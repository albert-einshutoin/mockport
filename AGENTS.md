# Repository verification commands

Follow the user's task instructions for scope and approvals. These commands are defined in the current repository.

| Command | Coverage |
| --- | --- |
| `make verify` | Go vet, tests, race tests, public trust, and adapter completeness. |
| `bash scripts/check-go-engineering.sh` | Go engineering checks. |
| `bash scripts/run-sdk-contracts.sh all` | Pinned official SDK contracts against a local Mockport binary. |
| `bash scripts/check-public-env.sh` | Public example secret and URL safety. |
| `bash scripts/check-compat-manifests.sh` | Compatibility manifest checks. |
| `bash scripts/check-distribution.sh` | Distribution checks. |
| `bash scripts/check-maintenance-policy.sh` | Maintenance policy checks. |
| `$(go env GOPATH)/bin/govulncheck ./...` | Go vulnerability scan after installing `golang.org/x/vuln/cmd/govulncheck@v1.3.0`. |
| `docker build -f docker/Dockerfile -t mockport:local .` | Build the pinned Go toolchain image. |

The `CI` workflow runs the full Go, SDK, public safety, compatibility, distribution, and maintenance checks on pushes and pull requests. The planned `ci-pr` selector belongs to draft PR #366 and is not yet on main.
