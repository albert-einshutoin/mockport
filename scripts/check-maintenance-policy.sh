#!/usr/bin/env bash
set -euo pipefail

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

require_absent_text() {
  local path="$1"
  local text="$2"
  if grep -Fq "$text" "$path"; then
    echo "unexpected text in $path: $text" >&2
    exit 1
  fi
}

require_file ".github/dependabot.yml"
require_text ".github/dependabot.yml" 'package-ecosystem: "github-actions"'
require_text ".github/dependabot.yml" 'package-ecosystem: "gomod"'
require_text ".github/dependabot.yml" 'package-ecosystem: "npm"'
require_text ".github/dependabot.yml" 'directory: "/packaging/npm"'
require_text ".github/dependabot.yml" 'directory: "/contract/sdk"'

require_file "ROADMAP.md"
require_text "ROADMAP.md" "Near Term"
require_text "ROADMAP.md" "Compatibility Direction"
require_text "ROADMAP.md" "Non-Goals"
require_text "ROADMAP.md" "Mockport does not reproduce provider internal logic"

require_file "docs/maintainer-guide.md"
require_file "docs/adapter-helper-policy.md"
require_file "docs/compatibility-model.md"
require_text "docs/maintainer-guide.md" "Do not auto-close stale issues"
require_text "docs/maintainer-guide.md" "GitHub Actions should use Node.js 24-compatible action releases"
require_text "docs/maintainer-guide.md" "Adapter Contribution Quality Bar"
require_text "docs/maintainer-guide.md" 'Test-only SDK dependencies are pinned in `contract/sdk`'
require_text "docs/adapter-helper-policy.md" "Until then, adapters should prefer clear local helpers over broad abstraction"
require_text "docs/compatibility-model.md" "Mockport compatibility is measured against public provider API behavior"
require_text "docs/compatibility-model.md" "SDK contract tests live under"
require_text "docs/compatibility-model.md" "provider-compatible"

require_file ".github/workflows/ci.yml"
require_file ".github/workflows/docker.yml"
require_file ".github/workflows/release.yml"
require_file ".github/workflows/smoke.yml"
require_file ".github/workflows/compatibility.yml"
require_file ".github/workflows/security.yml"

node24_workflows=(.github/workflows/ci.yml .github/workflows/docker.yml .github/workflows/release.yml .github/workflows/smoke.yml .github/workflows/compatibility.yml)
checkout_workflows=("${node24_workflows[@]}" .github/workflows/security.yml)

for workflow in "${node24_workflows[@]}"; do
  require_text "$workflow" "FORCE_JAVASCRIPT_ACTIONS_TO_NODE24"
  require_absent_text "$workflow" "actions/setup-go@v5"
done

# Keep every repository checkout on the Node.js 24-compatible major so the policy gate follows Dependabot action bumps.
for workflow in "${checkout_workflows[@]}"; do
  require_text "$workflow" "actions/checkout@v7"
  require_absent_text "$workflow" "actions/checkout@v4"
  require_absent_text "$workflow" "actions/checkout@v5"
  require_absent_text "$workflow" "actions/checkout@v6"
done

require_text ".github/workflows/ci.yml" "actions/setup-go@v6"
require_text ".github/workflows/ci.yml" "actions/setup-node@v6"
require_text ".github/workflows/ci.yml" "python3 ci/impact.py run"
require_text "ci/run-full.sh" "bash scripts/run-sdk-contracts.sh all"
require_text "ci/run-full.sh" "bash scripts/check-maintenance-policy.sh"

require_text ".github/workflows/docker.yml" "docker/setup-buildx-action@v4"
require_text ".github/workflows/docker.yml" "docker/login-action@v4"
require_text ".github/workflows/docker.yml" "docker/metadata-action@v6"
require_text ".github/workflows/docker.yml" "docker/build-push-action@v7"

require_text ".github/workflows/release.yml" "actions/upload-artifact@v7"
require_text ".github/workflows/release.yml" "actions/download-artifact@v8"
require_text ".github/workflows/release.yml" "softprops/action-gh-release@v3"

require_text ".github/workflows/smoke.yml" "schedule:"
require_text ".github/workflows/smoke.yml" "bash scripts/smoke-multi-adapter.sh"

require_text ".github/workflows/compatibility.yml" "schedule:"
require_text ".github/workflows/compatibility.yml" "workflow_dispatch"
require_text ".github/workflows/compatibility.yml" "actions/upload-artifact@v7"
require_text ".github/workflows/compatibility.yml" "bash scripts/check-compatibility-release.sh"

require_text "CONTRIBUTING.md" "Adapter acceptance criteria"
require_text "CONTRIBUTING.md" "No real provider secrets"
require_text "README.md" "Roadmap"
require_text "README.md" "Maintainer Guide"
