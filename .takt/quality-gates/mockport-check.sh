#!/usr/bin/env bash
set -euo pipefail

MODE="${MOCKPORT_TAKT_GATE_MODE:-standard}"

if [[ -n "${GO_BIN:-}" ]]; then
  GO="$GO_BIN"
elif command -v go >/dev/null 2>&1; then
  GO="$(command -v go)"
else
  GO="/usr/local/go/bin/go"
fi

# TAKT sandboxes may not inherit Go module paths; go test requires at least one.
# Some sandboxes expose read-only /go; fall back to a workspace-local cache.
ensure_go_cache_dirs() {
  local root
  root="$(pwd)"

  if [[ -z "${GOPATH:-}" ]]; then
    GOPATH="$("$GO" env GOPATH 2>/dev/null || true)"
    if [[ -z "$GOPATH" ]]; then
      GOPATH="${HOME:-}/go"
    fi
    export GOPATH
  fi

  if [[ -z "${GOMODCACHE:-}" ]]; then
    GOMODCACHE="$("$GO" env GOMODCACHE 2>/dev/null || true)"
    if [[ -z "$GOMODCACHE" ]]; then
      GOMODCACHE="${GOPATH}/pkg/mod"
    fi
  fi
  if ! mkdir -p "$GOMODCACHE" 2>/dev/null || [[ ! -w "$GOMODCACHE" ]]; then
    GOMODCACHE="${root}/.takt/.gomodcache"
    mkdir -p "$GOMODCACHE"
  fi
  export GOMODCACHE

  # Sandboxes may set GOCACHE=off; Go requires a writable cache directory.
  if [[ -z "${GOCACHE:-}" || "${GOCACHE:-}" == "off" ]]; then
    GOCACHE="$("$GO" env GOCACHE 2>/dev/null || true)"
    if [[ -z "$GOCACHE" || "$GOCACHE" == "off" ]]; then
      GOCACHE="${GOPATH}/pkg/cache"
    fi
  fi
  if [[ "${GOCACHE:-}" == "off" ]] || ! mkdir -p "$GOCACHE" 2>/dev/null || [[ ! -w "$GOCACHE" ]]; then
    GOCACHE="${root}/.takt/.gocache"
    mkdir -p "$GOCACHE"
  fi
  export GOCACHE
}
ensure_go_cache_dirs

run() {
  printf '\n==> %s\n' "$*"
  "$@"
}

check_gofmt() {
  # Keep formatting as a read-only gate so failed agent runs do not rewrite files.
  local output
  output="$(find . \
    \( -path './.git' -o -path './.devloop' -o -path './.takt' -o -path './node_modules' \) -prune \
    -o -name '*.go' -print0 | xargs -0 gofmt -l)"
  if [[ -n "$output" ]]; then
    printf 'gofmt required:\n%s\n' "$output" >&2
    return 1
  fi
}

install_go_tool() {
  local module="$1"
  local binary="$2"
  if command -v "$binary" >/dev/null 2>&1; then
    command -v "$binary"
    return
  fi
  run "$GO" install "$module"
  printf '%s/bin/%s\n' "$("$GO" env GOPATH)" "$binary"
}

case "$MODE" in
  standard|full) ;;
  *)
    printf 'unsupported MOCKPORT_TAKT_GATE_MODE: %s\n' "$MODE" >&2
    exit 2
    ;;
esac

run check_gofmt
run "$GO" test ./...
run "$GO" vet ./...
run bash scripts/check-public-trust.sh
run bash scripts/check-public-env.sh

if [[ "$MODE" == "full" ]]; then
  run "$GO" test -race ./...
  staticcheck_bin="$(install_go_tool honnef.co/go/tools/cmd/staticcheck@v0.7.0 staticcheck)"
  run "$staticcheck_bin" ./...
  govulncheck_bin="$(install_go_tool golang.org/x/vuln/cmd/govulncheck@v1.3.0 govulncheck)"
  run "$govulncheck_bin" ./...
  run bash scripts/run-sdk-contracts.sh all
  run bash scripts/check-distribution.sh
  run bash scripts/check-maintenance-policy.sh
fi
