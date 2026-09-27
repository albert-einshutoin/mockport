#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
echo "source_sha=$(git -C "$ROOT_DIR" rev-parse HEAD) dirty=$(if [[ -n "$(git -C "$ROOT_DIR" status --porcelain)" ]]; then echo true; else echo false; fi)"
WORK_DIR="$(mktemp -d)"
trap 'rm -rf "$WORK_DIR"' EXIT

PYTHON="${PYTHON:-python3.13}"
started="$(date +%s)"
"$PYTHON" -m venv "$WORK_DIR/venv"
venv_seconds=$(( $(date +%s) - started ))
started="$(date +%s)"
"$WORK_DIR/venv/bin/python" -m pip install --disable-pip-version-check --quiet -r "$ROOT_DIR/examples/external-llm/requirements.lock"
dependency_seconds=$(( $(date +%s) - started ))
"$WORK_DIR/venv/bin/python" -m pip check
"$WORK_DIR/venv/bin/llm" --version
"$WORK_DIR/venv/bin/python" -c 'import openai; print("openai", openai.__version__)'
echo "venv_seconds=$venv_seconds dependency_seconds=$dependency_seconds"

server=(--mockport-bin "$WORK_DIR/mockport")
if [[ -n "${MOCKPORT_IMAGE:-}" ]]; then
  server=(--mockport-image "$MOCKPORT_IMAGE")
else
  started="$(date +%s)"
  (cd "$ROOT_DIR" && go build -o "$WORK_DIR/mockport" ./cmd/mockport)
  echo "source_build_seconds=$(( $(date +%s) - started ))"
fi

if [[ -n "${EXTERNAL_LLM_RESULT:-}" ]]; then
  "$WORK_DIR/venv/bin/python" "$ROOT_DIR/examples/external-llm/verify.py" \
    --llm "$WORK_DIR/venv/bin/llm" "${server[@]}" --output "$EXTERNAL_LLM_RESULT"
else
  "$WORK_DIR/venv/bin/python" "$ROOT_DIR/examples/external-llm/verify.py" \
    --llm "$WORK_DIR/venv/bin/llm" "${server[@]}"
fi
