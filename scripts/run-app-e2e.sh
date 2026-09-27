#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
WORK_DIR="$(mktemp -d)"
KIND="${1:-all}"
MOCKPORT_PORT="${MOCKPORT_APP_E2E_PORT:-43101}"
STRIPE_PORT="${STRIPE_APP_E2E_PORT:-33001}"
STRIPE_PROXY_PORT="${STRIPE_TEST_PROXY_PORT:-43103}"
OPENAI_PORT="${OPENAI_APP_E2E_PORT:-33002}"
PROXY_PORT="${OPENAI_TEST_PROXY_PORT:-43102}"
SLACK_PORT="${SLACK_APP_E2E_PORT:-33003}"
MOCKPORT_URL="http://127.0.0.1:${MOCKPORT_PORT}"
STRIPE_URL="http://127.0.0.1:${STRIPE_PORT}"
STRIPE_PROXY_URL="http://127.0.0.1:${STRIPE_PROXY_PORT}"
OPENAI_URL="http://127.0.0.1:${OPENAI_PORT}"
PROXY_URL="http://127.0.0.1:${PROXY_PORT}"
SLACK_URL="http://127.0.0.1:${SLACK_PORT}"

stop_pid() {
  if [[ -n "${1:-}" ]]; then
    kill "$1" >/dev/null 2>&1 || true
    for _ in {1..10}; do
      if ! kill -0 "$1" 2>/dev/null; then
        break
      fi
      sleep 0.1
    done
    kill -KILL "$1" >/dev/null 2>&1 || true
    wait "$1" >/dev/null 2>&1 || true
  fi
}
cleanup() {
  local result="$?"
  if [[ "$result" != 0 ]]; then
    for log in "$WORK_DIR"/*.log; do
      [[ -f "$log" ]] && { echo "--- $log" >&2; tail -30 "$log" >&2; }
    done
  fi
  stop_pid "${OPENAI_PID:-}"
  stop_pid "${PROXY_PID:-}"
  stop_pid "${STRIPE_PID:-}"
  stop_pid "${STRIPE_PROXY_PID:-}"
  stop_pid "${SLACK_PID:-}"
  stop_pid "${MOCKPORT_PID:-}"
  rm -rf "$WORK_DIR"
}
trap cleanup EXIT

if [[ "$KIND" != "stripe" && "$KIND" != "openai" && "$KIND" != "slack" && "$KIND" != "all" ]]; then
  echo "usage: bash scripts/run-app-e2e.sh [stripe|openai|slack|all]" >&2
  exit 2
fi

wait_for() {
  local url="$1" pid="$2" log="$3"
  for _ in {1..30}; do
    if ! kill -0 "$pid" 2>/dev/null; then
      cat "$log" >&2
      return 1
    fi
    if curl -fsS --max-time 1 "$url/health" >/dev/null 2>&1; then
      return 0
    fi
    sleep 0.2
  done
  cat "$log" >&2
  echo "$url did not become ready" >&2
  return 1
}

start_openai() {
  local key="$1" base="$2"
  if [[ "$key" == "__missing__" ]]; then
    env -u OPENAI_API_KEY OPENAI_BASE_URL="$base/openai/v1" PORT="$OPENAI_PORT" \
      "$PYTHON" "${OPENAI_APP_ENTRY:-$ROOT_DIR/examples/app-e2e/openai-app/server.py}" >"$WORK_DIR/openai.log" 2>&1 &
  else
    OPENAI_API_KEY="$key" OPENAI_BASE_URL="$base/openai/v1" PORT="$OPENAI_PORT" \
      "$PYTHON" "${OPENAI_APP_ENTRY:-$ROOT_DIR/examples/app-e2e/openai-app/server.py}" >"$WORK_DIR/openai.log" 2>&1 &
  fi
  OPENAI_PID="$!"
  wait_for "$OPENAI_URL" "$OPENAI_PID" "$WORK_DIR/openai.log"
}

restart_openai() {
  stop_pid "$OPENAI_PID"
  OPENAI_PID=""
  start_openai "$1" "$2"
}

start_proxy() {
  TEST_SCENARIO="$1" TEST_DELAY_MS="$2" MOCKPORT_PORT="$MOCKPORT_PORT" PORT="$PROXY_PORT" \
    "$PYTHON" "$ROOT_DIR/examples/app-e2e/openai-app/test_proxy.py" >"$WORK_DIR/proxy.log" 2>&1 &
  PROXY_PID="$!"
  sleep 0.2
  kill -0 "$PROXY_PID"
}

cd "$ROOT_DIR"
ports=("$MOCKPORT_PORT")
if [[ "$KIND" == "stripe" || "$KIND" == "all" ]]; then
  ports+=("$STRIPE_PORT" "$STRIPE_PROXY_PORT")
fi
if [[ "$KIND" == "openai" || "$KIND" == "all" ]]; then
  ports+=("$OPENAI_PORT" "$PROXY_PORT")
fi
if [[ "$KIND" == "slack" || "$KIND" == "all" ]]; then
  ports+=("$SLACK_PORT")
fi
node - "${ports[@]}" <<'NODE'
const net = require("node:net");
(async () => {
  const seen = new Set();
  for (const value of process.argv.slice(2)) {
    const port = Number(value);
    if (seen.has(port)) throw new Error(`port ${port} is configured for multiple services`);
    seen.add(port);
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
  -e "s/127.0.0.1:33003/127.0.0.1:${SLACK_PORT}/" \
  examples/app-e2e/mockport.yml > "$WORK_DIR/mockport.yml"
"$WORK_DIR/mockport" run --config "$WORK_DIR/mockport.yml" >"$WORK_DIR/mockport.log" 2>&1 &
MOCKPORT_PID="$!"
wait_for "$MOCKPORT_URL" "$MOCKPORT_PID" "$WORK_DIR/mockport.log"

if [[ "$KIND" == "stripe" || "$KIND" == "all" ]]; then
  cd "$ROOT_DIR/examples/app-e2e/stripe-app"
  npm ci --ignore-scripts
  MOCKPORT_UPSTREAM_URL="$MOCKPORT_URL" PORT="$STRIPE_PROXY_PORT" node test-proxy.mjs >"$WORK_DIR/stripe-proxy.log" 2>&1 &
  STRIPE_PROXY_PID="$!"
  wait_for "$STRIPE_PROXY_URL" "$STRIPE_PROXY_PID" "$WORK_DIR/stripe-proxy.log"
  MOCKPORT_BASE_URL="$STRIPE_PROXY_URL" STRIPE_SECRET_KEY=mockport_stripe_secret \
    STRIPE_WEBHOOK_SECRET=whsec_mockport PORT="$STRIPE_PORT" \
    node "${STRIPE_APP_ENTRY:-$ROOT_DIR/examples/app-e2e/stripe-app/server.mjs}" >"$WORK_DIR/stripe.log" 2>&1 &
  STRIPE_PID="$!"
  wait_for "$STRIPE_URL" "$STRIPE_PID" "$WORK_DIR/stripe.log"
fi

if [[ "$KIND" == "openai" || "$KIND" == "slack" || "$KIND" == "all" ]]; then
  python3 -m venv "$WORK_DIR/venv"
  PYTHON="$WORK_DIR/venv/bin/python"
fi
if [[ "$KIND" == "openai" || "$KIND" == "all" ]]; then
  "$PYTHON" -m pip install -r "$ROOT_DIR/examples/app-e2e/openai-app/requirements.lock" --quiet
  start_openai mockport_openai_key "$MOCKPORT_URL"
fi
if [[ "$KIND" == "slack" || "$KIND" == "all" ]]; then
  "$PYTHON" -m pip install -r "$ROOT_DIR/examples/app-e2e/slack-app/requirements.lock" --quiet
  SLACK_BASE_URL="$MOCKPORT_URL/slack/api/" SLACK_BOT_TOKEN=mockport_slack_token \
    SLACK_SIGNING_SECRET=mockport_slack_signing_secret PORT="$SLACK_PORT" \
    "$PYTHON" "${SLACK_APP_ENTRY:-$ROOT_DIR/examples/app-e2e/slack-app/server.py}" >"$WORK_DIR/slack.log" 2>&1 &
  SLACK_PID="$!"
  wait_for "$SLACK_URL" "$SLACK_PID" "$WORK_DIR/slack.log"
fi

echo "head=$(git rev-parse HEAD) dirty=$(if [[ -n "$(git status --porcelain)" ]]; then echo true; else echo false; fi) binary=$WORK_DIR/mockport stripe=22.3.1 openai=2.46.0 slack=3.44.1"
if [[ "$KIND" == "stripe" || "$KIND" == "all" ]]; then
  cd "$ROOT_DIR/examples/app-e2e/stripe-app"
  MOCKPORT_BASE_URL="$MOCKPORT_URL" APP_BASE_URL="$STRIPE_URL" STRIPE_TEST_PROXY_URL="$STRIPE_PROXY_URL" node e2e.mjs
fi
if [[ "$KIND" == "openai" || "$KIND" == "all" ]]; then
  cd "$ROOT_DIR"
  MOCKPORT_BASE_URL="$MOCKPORT_URL" APP_BASE_URL="$OPENAI_URL" "$PYTHON" examples/app-e2e/openai-app/e2e.py success
  restart_openai mockport_wrong "$MOCKPORT_URL"
  MOCKPORT_BASE_URL="$MOCKPORT_URL" APP_BASE_URL="$OPENAI_URL" "$PYTHON" examples/app-e2e/openai-app/e2e.py wrong_key
  restart_openai __missing__ "$MOCKPORT_URL"
  MOCKPORT_BASE_URL="$MOCKPORT_URL" APP_BASE_URL="$OPENAI_URL" "$PYTHON" examples/app-e2e/openai-app/e2e.py missing_key
  start_proxy rate_limited ""
  restart_openai mockport_openai_key "$PROXY_URL"
  MOCKPORT_BASE_URL="$MOCKPORT_URL" APP_BASE_URL="$OPENAI_URL" "$PYTHON" examples/app-e2e/openai-app/e2e.py rate_limited
  stop_pid "$PROXY_PID"
  PROXY_PID=""
  start_proxy "" 1500
  restart_openai mockport_openai_key "$PROXY_URL"
  MOCKPORT_BASE_URL="$MOCKPORT_URL" APP_BASE_URL="$OPENAI_URL" "$PYTHON" examples/app-e2e/openai-app/e2e.py timeout
  MOCKPORT_BASE_URL="$MOCKPORT_URL" APP_BASE_URL="$OPENAI_URL" "$PYTHON" examples/app-e2e/openai-app/e2e.py cancel
fi
if [[ "$KIND" == "slack" || "$KIND" == "all" ]]; then
  cd "$ROOT_DIR"
  MOCKPORT_BASE_URL="$MOCKPORT_URL" APP_BASE_URL="$SLACK_URL" \
    SLACK_EVENT_FIXTURE="$ROOT_DIR/compat/fixtures/slack/events_message_callback.json" \
    "$PYTHON" examples/app-e2e/slack-app/e2e.py
fi
