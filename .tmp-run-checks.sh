#!/usr/bin/env bash
set +e
cd /Volumes/Satechi/Developer/mockport

echo "=== COMMAND 1 ==="
test -f docs/site/scenarios.md && echo "scenarios.md: OK" || echo "scenarios.md: MISSING"
echo "EXIT_CODE: $?"

echo "=== COMMAND 2 ==="
grep -F "scenarios.md" docs/site/index.md && echo "index link: OK" || echo "index link: MISSING"
echo "EXIT_CODE: $?"

echo "=== COMMAND 3 ==="
bash scripts/check-doc-links.sh
echo "EXIT_CODE: $?"

echo "=== COMMAND 4 ==="
bash scripts/check-public-trust.sh
echo "EXIT_CODE: $?"
