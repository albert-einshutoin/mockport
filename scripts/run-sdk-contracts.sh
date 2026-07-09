#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PROVIDER="${1:-all}"
PORT="${MOCKPORT_CONTRACT_PORT:-43101}"
BASE_URL="http://127.0.0.1:${PORT}"

case "$PROVIDER" in
  all|stripe|openai|github-oauth|slack) ;;
  *)
    echo "unsupported provider: $PROVIDER" >&2
    exit 1
    ;;
esac

WORK_DIR="$(mktemp -d)"

if [[ -z "${GO_BIN:-}" ]]; then
  if command -v go >/dev/null 2>&1; then
    GO_BIN="go"
  else
    GO_BIN="/usr/local/go/bin/go"
  fi
fi

cleanup() {
  if [[ -n "${SERVER_PID:-}" ]]; then
    kill "$SERVER_PID" >/dev/null 2>&1 || true
    wait "$SERVER_PID" >/dev/null 2>&1 || true
    SERVER_PID=""
  fi
  rm -rf "$WORK_DIR"
}
trap cleanup EXIT

start_mockport() {
  cleanup_server
  "$WORK_DIR/mockport" run --config "$WORK_DIR/mockport.yml" >"$WORK_DIR/mockport.log" 2>&1 &
  SERVER_PID="$!"

  for _ in $(seq 1 30); do
    if curl -fsS "$BASE_URL/health" >/dev/null 2>&1; then
      return 0
    fi
    sleep 1
  done

  echo "mockport failed to become healthy" >&2
  return 1
}

cleanup_server() {
  if [[ -n "${SERVER_PID:-}" ]]; then
    kill "$SERVER_PID" >/dev/null 2>&1 || true
    wait "$SERVER_PID" >/dev/null 2>&1 || true
    SERVER_PID=""
  fi
}

set_slack_scenario() {
  local scenario="$1"
  python3 - "$WORK_DIR/mockport.yml" "$scenario" <<'PY'
import pathlib
import sys

path = pathlib.Path(sys.argv[1])
scenario = sys.argv[2]
text = path.read_text()
text = text.replace("    scenario: message_success", f"    scenario: {scenario}", 1)
path.write_text(text)
PY
}

cd "$ROOT_DIR"
"$GO_BIN" build -o "$WORK_DIR/mockport" ./cmd/mockport
cp examples/multi-adapter/mockport.yml "$WORK_DIR/mockport.yml"
python3 - "$WORK_DIR/mockport.yml" "$PORT" <<'PY'
import pathlib
import sys

path = pathlib.Path(sys.argv[1])
port = sys.argv[2]
text = path.read_text()
text = text.replace("  port: 43101", f"  port: {port}")
path.write_text(text)
PY

start_mockport

cd "$ROOT_DIR/contract/sdk"
npm ci
npm run test:live -- --provider "$PROVIDER" --base-url "$BASE_URL" --json

if [[ "$PROVIDER" == "slack" || "$PROVIDER" == "all" ]]; then
  set_slack_scenario "rate_limited"
  start_mockport
  npm run test:live -- --provider slack-rate-limited --base-url "$BASE_URL" --json
fi
