#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
WORK_DIR="$(mktemp -d)"
trap 'rm -rf "$WORK_DIR"' EXIT
python3 - "$ROOT_DIR/examples/app-e2e/stripe-app/server.mjs" "$WORK_DIR/server.mjs" <<'PY'
from pathlib import Path
import sys

source = Path(sys.argv[1]).read_text()
old_dedupe = 'if (processedEvents.has(event.id)) return reply(response, 200, { duplicate: true });'
assert source.count(old_dedupe) == 1
source = source.replace(old_dedupe, '// mutation: event replay is processed again', 1)
Path(sys.argv[2]).write_text(source)
PY
ln -s "$ROOT_DIR/examples/app-e2e/stripe-app/node_modules" "$WORK_DIR/node_modules"
if STRIPE_APP_ENTRY="$WORK_DIR/server.mjs" bash "$ROOT_DIR/scripts/run-app-e2e.sh" stripe >"$WORK_DIR/mutated.log" 2>&1; then
  echo "mutated app unexpectedly passed" >&2
  exit 1
fi
if ! grep -q 'processed_events' "$WORK_DIR/mutated.log"; then
  cat "$WORK_DIR/mutated.log" >&2
  echo "mutated app failed for a reason other than replay processing" >&2
  exit 1
fi
grep 'processed_events' "$WORK_DIR/mutated.log"
bash "$ROOT_DIR/scripts/run-app-e2e.sh" stripe
