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

require_file "LICENSE"
require_text "LICENSE" "MIT License"

require_file "SECURITY.md"
require_text "SECURITY.md" "Supported Versions"
require_text "SECURITY.md" "AI-safe"
require_text "SECURITY.md" "Do not include real secrets"

require_file "CONTRIBUTING.md"
require_text "CONTRIBUTING.md" "TDD"
require_text "CONTRIBUTING.md" "go test ./..."
require_text "CONTRIBUTING.md" "bash scripts/check-public-trust.sh"
require_text "CONTRIBUTING.md" "bash scripts/check-adapter-completeness.sh"
require_text "CONTRIBUTING.md" "make verify"

require_file "CODE_OF_CONDUCT.md"
require_text "CODE_OF_CONDUCT.md" "Contributor Covenant"

require_file "docs/public-support-policy.md"
require_text "docs/public-support-policy.md" "Support Policy"
require_text "docs/public-support-policy.md" "scenario-compatible"
require_text "docs/public-support-policy.md" "provider-compatible"

require_file ".github/ISSUE_TEMPLATE/bug_report.yml"
require_text ".github/ISSUE_TEMPLATE/bug_report.yml" "adapter"
require_text ".github/ISSUE_TEMPLATE/bug_report.yml" "redacted"

require_file ".github/ISSUE_TEMPLATE/feature_request.yml"
require_text ".github/ISSUE_TEMPLATE/feature_request.yml" "target adapter"
require_text ".github/ISSUE_TEMPLATE/feature_request.yml" "safety impact"

require_file ".github/pull_request_template.md"
require_text ".github/pull_request_template.md" "Test evidence"
require_text ".github/pull_request_template.md" "Public env safety"

require_text "README.md" "No local install required"
require_text "README.md" "npm wrapper is experimental"
require_file "llms.txt"
require_text "llms.txt" "Security boundary"
require_file "examples/node-sdk-clients/package.json"
require_file "examples/node-sdk-clients/openai.mjs"
require_file "examples/node-sdk-clients/stripe.mjs"
require_file "examples/node-sdk-clients/line.mjs"
require_file "examples/python-openai/example.py"
require_file "examples/python-openai/requirements.txt"
require_file "scripts/smoke-node-sdk-examples.sh"
require_text "README.md" "Node SDK examples"
require_text ".github/workflows/compatibility.yml" "bash scripts/smoke-node-sdk-examples.sh"
require_text "docs/public-env-safety.md" "mockport-public-safety"
require_text ".github/workflows/ci.yml" "bash scripts/check-public-trust.sh"
require_text ".github/workflows/ci.yml" "bash scripts/check-adapter-completeness.sh"
require_text ".github/workflows/ci.yml" "bash scripts/check-distribution.sh"
require_file "scripts/check-adapter-helpers.sh"
bash scripts/check-adapter-helpers.sh
require_file "scripts/check-support-surfaces.mjs"
node scripts/check-support-surfaces.mjs
