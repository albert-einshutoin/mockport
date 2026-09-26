#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
WORK_DIR="$(mktemp -d)"
trap 'rm -rf "$WORK_DIR"' EXIT
python3 - "$ROOT_DIR/examples/app-e2e/openai-app/server.py" "$WORK_DIR/server.py" <<'PY'
from pathlib import Path
import sys

source = Path(sys.argv[1]).read_text()
old = 'self.reply(429, {"error": "rate_limited", "max_attempts": 2})'
assert source.count(old) == 1
Path(sys.argv[2]).write_text(source.replace(old, 'self.reply(200, {"text": "Mockport response", "completed": True})', 1))
PY
if OPENAI_APP_ENTRY="$WORK_DIR/server.py" bash "$ROOT_DIR/scripts/run-app-e2e.sh" openai >"$WORK_DIR/mutated.log" 2>&1; then
  echo "mutated OpenAI app unexpectedly passed" >&2
  exit 1
fi
if ! grep -q 'AssertionError: (200,' "$WORK_DIR/mutated.log"; then
  cat "$WORK_DIR/mutated.log" >&2
  echo "mutated app failed for a reason other than treating 429 as success" >&2
  exit 1
fi
grep 'AssertionError: (200,' "$WORK_DIR/mutated.log"
bash "$ROOT_DIR/scripts/run-app-e2e.sh" openai
