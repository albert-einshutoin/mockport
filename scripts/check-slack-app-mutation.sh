#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
WORK_DIR="$(mktemp -d)"
trap 'rm -rf "$WORK_DIR"' EXIT
python3 - "$ROOT_DIR/examples/app-e2e/slack-app/server.py" "$WORK_DIR/server.py" <<'PY'
from pathlib import Path
import sys

source = Path(sys.argv[1]).read_text()
old = "if not VERIFIER.is_valid_request(raw, self.headers):"
assert source.count(old) == 1
Path(sys.argv[2]).write_text(source.replace(old, "if not True:", 1))
PY
if SLACK_APP_ENTRY="$WORK_DIR/server.py" bash "$ROOT_DIR/scripts/run-app-e2e.sh" slack >"$WORK_DIR/mutated.log" 2>&1; then
  echo "mutated Slack app unexpectedly passed" >&2
  exit 1
fi
if ! grep -q 'invalid_signature_expected_rejection status=200' "$WORK_DIR/mutated.log"; then
  cat "$WORK_DIR/mutated.log" >&2
  echo "mutated app failed for a reason other than accepting an invalid signature" >&2
  exit 1
fi
grep 'invalid_signature_expected_rejection status=200' "$WORK_DIR/mutated.log"
bash "$ROOT_DIR/scripts/run-app-e2e.sh" slack
