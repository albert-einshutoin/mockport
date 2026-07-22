#!/usr/bin/env bash
set -euo pipefail

# This runner is deliberately independent from the impact code and config. If
# planning or config parsing breaks, CI still exercises every product suite.
go test ./...
go test -race ./...
bash scripts/run-sdk-contracts.sh all
bash scripts/smoke-multi-adapter.sh
bash scripts/check-public-trust.sh
bash scripts/check-public-env.sh
bash scripts/check-doc-links.sh
bash scripts/check-adapter-completeness.sh
bash scripts/check-compat-manifests.sh
bash scripts/check-distribution.sh
bash scripts/check-maintenance-policy.sh
