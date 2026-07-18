#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PORT="${MOCKPORT_EXAMPLE_PORT:-43103}"
BASE_URL="http://127.0.0.1:${PORT}"
WORK_DIR="$(mktemp -d)"

cleanup() {
  if [[ -n "${SERVER_PID:-}" ]]; then
    kill "$SERVER_PID" >/dev/null 2>&1 || true
    wait "$SERVER_PID" >/dev/null 2>&1 || true
  fi
  rm -rf "$WORK_DIR"
}
trap cleanup EXIT

cd "$ROOT_DIR"
go build -o "$WORK_DIR/mockport" ./cmd/mockport
cp examples/multi-adapter/mockport.yml "$WORK_DIR/mockport.yml"
python3 - "$WORK_DIR/mockport.yml" "$PORT" <<'PY'
import pathlib
import sys

path = pathlib.Path(sys.argv[1])
path.write_text(path.read_text().replace("  port: 43101", f"  port: {sys.argv[2]}"))
PY

"$WORK_DIR/mockport" run --config "$WORK_DIR/mockport.yml" >"$WORK_DIR/mockport.log" 2>&1 &
SERVER_PID="$!"
for _ in $(seq 1 30); do
  if curl -fsS "$BASE_URL/health" >/dev/null 2>&1; then
    break
  fi
  sleep 1
done
curl -fsS "$BASE_URL/health" >/dev/null

npm --prefix examples/node-sdk-clients ci
for example in stripe openai line; do
  MOCKPORT_BASE_URL="$BASE_URL" npm --prefix examples/node-sdk-clients run "$example"
done

python3 -m venv "$WORK_DIR/python-venv"
"$WORK_DIR/python-venv/bin/python" -m pip install \
  --disable-pip-version-check \
  -r examples/python-openai/requirements.txt
MOCKPORT_BASE_URL="$BASE_URL" \
  "$WORK_DIR/python-venv/bin/python" examples/python-openai/example.py
