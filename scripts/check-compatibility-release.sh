#!/usr/bin/env bash
set -euo pipefail

bash scripts/check-compat-manifests.sh

require_file() {
  local path="$1"
  if [[ ! -f "$path" ]]; then
    echo "missing file: $path" >&2
    exit 1
  fi
}

require_text() {
  local path="$1"
  local text="$2"
  if ! grep -Fq "$text" "$path"; then
    echo "missing text in $path: $text" >&2
    exit 1
  fi
}

require_file ".github/workflows/compatibility.yml"
require_text ".github/workflows/compatibility.yml" "workflow_dispatch"
require_text ".github/workflows/compatibility.yml" "schedule:"
require_text ".github/workflows/compatibility.yml" "bash scripts/run-sdk-contracts.sh stripe"
require_text ".github/workflows/compatibility.yml" "bash scripts/run-sdk-contracts.sh openai"
require_text ".github/workflows/compatibility.yml" "bash scripts/run-sdk-contracts.sh github-oauth"
require_text ".github/workflows/compatibility.yml" "bash scripts/run-sdk-contracts.sh slack"
require_text ".github/workflows/compatibility.yml" "actions/upload-artifact@v7"

require_file "scripts/generate-compatibility-report.sh"
require_file "docs/compatibility-reports/README.md"
require_file "docs/compatibility-reports/latest.md"
require_file "docs/compatibility-reports/latest.json"

require_text "docs/compatibility-reports/README.md" "provider-compatible"
require_text "docs/compatibility-reports/latest.md" "Compatibility Report"
require_text "docs/compatibility-reports/latest.md" "Known Gaps"
require_text "docs/compatibility-reports/latest.json" '"generated_by": "scripts/generate-compatibility-report.sh"'
require_text "docs/site/support-matrix.md" "provider-compatible"
require_text "CHANGELOG.md" "Compatibility release track"
require_text "CHANGELOG.md" "compatibility scores"

# Gate the published report through the shared validator, which mirrors
# internal/compat CanPromote (maturities require concrete coverage, not just a
# score threshold). Run its regression tests first so the gate logic stays covered.
node --test scripts/validate-compatibility-report.test.mjs
node scripts/validate-compatibility-report.mjs docs/compatibility-reports/latest.json

# Keep the published evidence date at least as new as every committed source
# that can change the generated compatibility snapshot. This catches content
# refreshes that accidentally preserve an older generated_at value.
LATEST_COMPATIBILITY_SOURCE_DATE="$(
  git log -1 --format=%cs -- \
    adapters \
    compat \
    contract \
    examples/multi-adapter/mockport.yml \
    internal/adapter \
    internal/compat \
    internal/report \
    scripts/generate-compatibility-report.sh \
    scripts/render-compatibility-report.mjs
)"
if [[ -z "$LATEST_COMPATIBILITY_SOURCE_DATE" ]]; then
  echo "unable to determine latest compatibility source date" >&2
  exit 1
fi
node --test scripts/check-compatibility-freshness.test.mjs
node scripts/check-compatibility-freshness.mjs \
  docs/compatibility-reports/latest.json \
  "$LATEST_COMPATIBILITY_SOURCE_DATE"

# Provenance gate: regenerate the report from the runtime endpoint and require
# the committed artifacts to match. Use the committed generated_at date so this
# check catches hand edits and stale content without failing only because the
# calendar day changed.
REPORT_WORK_DIR="$(mktemp -d)"
cleanup_report_work_dir() {
  rm -rf "$REPORT_WORK_DIR"
}
trap cleanup_report_work_dir EXIT

REPORT_DATE="$(
  node -e 'const fs = require("fs"); const report = JSON.parse(fs.readFileSync("docs/compatibility-reports/latest.json", "utf8")); if (!report.generated_at) throw new Error("latest.json missing generated_at"); process.stdout.write(report.generated_at);'
)"
MOCKPORT_COMPATIBILITY_DATE="$REPORT_DATE" bash scripts/generate-compatibility-report.sh "$REPORT_WORK_DIR"
diff -u docs/compatibility-reports/latest.json "$REPORT_WORK_DIR/latest.json"
diff -u docs/compatibility-reports/latest.md "$REPORT_WORK_DIR/latest.md"
