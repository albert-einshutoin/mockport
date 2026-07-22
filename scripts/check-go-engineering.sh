#!/usr/bin/env bash
set -euo pipefail

missing=0

require_file() {
  if [[ ! -f "$1" ]]; then
    echo "missing required file: $1" >&2
    missing=1
  fi
}

require_pattern() {
  local pattern="$1"
  local path="$2"
  if ! grep -qE -- "$pattern" "$path"; then
    echo "missing pattern '$pattern' in $path" >&2
    missing=1
  fi
}

require_file internal/adapter/httpx/json.go
require_file internal/adapter/httpx/benchmark_test.go
require_file internal/state/benchmark_test.go
require_file internal/report/benchmark_test.go
require_file internal/compat/benchmark_test.go
require_file docs/go-engineering-readiness.md

require_pattern "func ValidateMetadata" internal/adapter/adapter.go
require_pattern "LevelClient" internal/compat/manifest.go
require_pattern "http.Server" internal/cli/run.go
require_pattern "signal.NotifyContext" internal/cli/run.go
require_pattern "MaxBytesReader" internal/adapter/httpx/json.go
require_pattern "func cloneValue" internal/state/store.go
require_pattern "SetClock" internal/report/recorder.go
require_pattern "go test -race ./..." ci/run-full.sh
require_pattern "staticcheck ./..." .github/workflows/ci.yml
require_pattern "govulncheck ./..." .github/workflows/ci.yml
require_pattern "Ignored Error Policy" docs/go-engineering-readiness.md

if grep -rnE --include='*.go' --exclude='*_test.go' "func .*map\\[string\\](any|interface\\{\\})" adapters | grep -vE "completionBody|messageBody|stripeDataFromStruct|dataFromStruct|fallbackCheckoutSession|fallbackPaymentIntent|decodePayload|writeGenericResource|createStatefulResource|writeResource"; then
  echo "raw map-returning adapter builders must be explicitly allowlisted as dynamic boundaries" >&2
  missing=1
fi

exit "$missing"
