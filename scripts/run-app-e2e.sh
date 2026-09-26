#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
WORK_DIR="$(mktemp -d)"
MOCKPORT_PORT="${MOCKPORT_APP_E2E_PORT:-43101}"
STRIPE_PORT="${STRIPE_APP_E2E_PORT:-33001}"
MOCKPORT_URL="http://127.0.0.1:${MOCKPORT_PORT}"
STRIPE_URL="http://127.0.0.1:${STRIPE_PORT}"

cleanup() {
  for pid in "${STRIPE_PID:-}" "${MOCKPORT_PID:-}"; do
    if [[ -n "$pid" ]]; then
      kill "$pid" >/dev/null 2>&1 || true
      wait "$pid" >/dev/null 2>&1 || true
    fi
  done
  rm -rf "$WORK_DIR"
}
trap cleanup EXIT

if [[ "${1:-stripe}" != "stripe" ]]; then
  echo "usage: bash scripts/run-app-e2e.sh stripe" >&2
  exit 2
fi

cd "$ROOT_DIR"
"${GO_BIN:-go}" build -o "$WORK_DIR/mockport" ./cmd/mockport
sed -e "s/port: 43101/port: ${MOCKPORT_PORT}/" \
  -e "s/127.0.0.1:33001/127.0.0.1:${STRIPE_PORT}/" \
  examples/app-e2e/mockport.yml > "$WORK_DIR/mockport.yml"
"$WORK_DIR/mockport" run --config "$WORK_DIR/mockport.yml" >"$WORK_DIR/mockport.log" 2>&1 &
MOCKPORT_PID="$!"

cd "$ROOT_DIR/examples/app-e2e/stripe-app"
npm ci --ignore-scripts
MOCKPORT_BASE_URL="$MOCKPORT_URL" STRIPE_SECRET_KEY=mockport_stripe_secret \
  STRIPE_WEBHOOK_SECRET=whsec_mockport PORT="$STRIPE_PORT" \
  node "${STRIPE_APP_ENTRY:-$ROOT_DIR/examples/app-e2e/stripe-app/server.mjs}" >"$WORK_DIR/stripe.log" 2>&1 &
STRIPE_PID="$!"

for _ in {1..30}; do
  if ! kill -0 "$MOCKPORT_PID" 2>/dev/null || ! kill -0 "$STRIPE_PID" 2>/dev/null; then
    cat "$WORK_DIR/mockport.log" "$WORK_DIR/stripe.log" >&2
    exit 1
  fi
  if curl -fsS "$MOCKPORT_URL/health" >/dev/null 2>&1 && curl -fsS "$STRIPE_URL/health" >/dev/null 2>&1; then
    break
  fi
  sleep 1
done
curl -fsS "$MOCKPORT_URL/health" >/dev/null
curl -fsS "$STRIPE_URL/health" >/dev/null
echo "source=$(git rev-parse HEAD) binary=$WORK_DIR/mockport stripe=22.3.1"
MOCKPORT_BASE_URL="$MOCKPORT_URL" APP_BASE_URL="$STRIPE_URL" node e2e.mjs
