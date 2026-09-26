#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
WORK_DIR="$(mktemp -d)"
MOCKPORT_PORT="${MOCKPORT_APP_E2E_PORT:-43101}"
STRIPE_PORT="${STRIPE_APP_E2E_PORT:-33001}"
STRIPE_PROXY_PORT="${STRIPE_TEST_PROXY_PORT:-43103}"
MOCKPORT_URL="http://127.0.0.1:${MOCKPORT_PORT}"
STRIPE_URL="http://127.0.0.1:${STRIPE_PORT}"
STRIPE_PROXY_URL="http://127.0.0.1:${STRIPE_PROXY_PORT}"

cleanup() {
  for pid in "${STRIPE_PID:-}" "${STRIPE_PROXY_PID:-}" "${MOCKPORT_PID:-}"; do
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
node - "$MOCKPORT_PORT" "$STRIPE_PORT" "$STRIPE_PROXY_PORT" <<'NODE'
const net = require("node:net");
(async () => {
  for (const value of process.argv.slice(2)) {
    const port = Number(value);
    await new Promise((resolve, reject) => {
      const server = net.createServer();
      server.once("error", () => reject(new Error(`port ${port} is already in use`)));
      server.listen(port, "127.0.0.1", () => server.close(resolve));
    });
  }
})().catch((error) => { console.error(error.message); process.exitCode = 1; });
NODE
"${GO_BIN:-go}" build -o "$WORK_DIR/mockport" ./cmd/mockport
sed -e "s/port: 43101/port: ${MOCKPORT_PORT}/" \
  -e "s/127.0.0.1:33001/127.0.0.1:${STRIPE_PORT}/" \
  examples/app-e2e/mockport.yml > "$WORK_DIR/mockport.yml"
"$WORK_DIR/mockport" run --config "$WORK_DIR/mockport.yml" >"$WORK_DIR/mockport.log" 2>&1 &
MOCKPORT_PID="$!"

cd "$ROOT_DIR/examples/app-e2e/stripe-app"
npm ci --ignore-scripts
MOCKPORT_UPSTREAM_URL="$MOCKPORT_URL" PORT="$STRIPE_PROXY_PORT" node test-proxy.mjs >"$WORK_DIR/stripe-proxy.log" 2>&1 &
STRIPE_PROXY_PID="$!"
MOCKPORT_BASE_URL="$STRIPE_PROXY_URL" STRIPE_SECRET_KEY=mockport_stripe_secret \
  STRIPE_WEBHOOK_SECRET=whsec_mockport PORT="$STRIPE_PORT" \
  node "${STRIPE_APP_ENTRY:-$ROOT_DIR/examples/app-e2e/stripe-app/server.mjs}" >"$WORK_DIR/stripe.log" 2>&1 &
STRIPE_PID="$!"

for _ in {1..30}; do
  if ! kill -0 "$MOCKPORT_PID" 2>/dev/null || ! kill -0 "$STRIPE_PROXY_PID" 2>/dev/null || ! kill -0 "$STRIPE_PID" 2>/dev/null; then
    cat "$WORK_DIR/mockport.log" "$WORK_DIR/stripe-proxy.log" "$WORK_DIR/stripe.log" >&2
    exit 1
  fi
  if curl -fsS "$MOCKPORT_URL/health" >/dev/null 2>&1 && curl -fsS "$STRIPE_PROXY_URL/health" >/dev/null 2>&1 && curl -fsS "$STRIPE_URL/health" >/dev/null 2>&1; then
    break
  fi
  sleep 1
done
curl -fsS "$MOCKPORT_URL/health" >/dev/null
curl -fsS "$STRIPE_URL/health" >/dev/null
kill -0 "$MOCKPORT_PID" 2>/dev/null && kill -0 "$STRIPE_PROXY_PID" 2>/dev/null && kill -0 "$STRIPE_PID" 2>/dev/null
echo "source=$(git rev-parse HEAD) binary=$WORK_DIR/mockport stripe=22.3.1"
MOCKPORT_BASE_URL="$MOCKPORT_URL" APP_BASE_URL="$STRIPE_URL" STRIPE_TEST_PROXY_URL="$STRIPE_PROXY_URL" node e2e.mjs
