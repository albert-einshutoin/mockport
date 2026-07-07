#!/usr/bin/env bash
set -u
cd /Volumes/Satechi/Developer/mockport

run() {
  echo "===== COMMAND: $1 ====="
  bash -c "$1"
  local ec=$?
  echo "EXIT_CODE=$ec"
  echo
  return $ec
}

run "bash scripts/check-doc-links.sh"
run "bash scripts/check-public-trust.sh"
