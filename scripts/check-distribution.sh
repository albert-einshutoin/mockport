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

require_file ".github/workflows/release.yml"
require_text ".github/workflows/release.yml" "linux"
require_text ".github/workflows/release.yml" "darwin"
require_text ".github/workflows/release.yml" "amd64"
require_text ".github/workflows/release.yml" "arm64"
require_text ".github/workflows/release.yml" "checksums.txt"
require_text ".github/workflows/release.yml" "actions/upload-artifact"

require_file ".github/workflows/docker.yml"
require_text ".github/workflows/docker.yml" "ghcr.io/albert-einshutoin/mockport"
require_text ".github/workflows/docker.yml" "type=semver"
require_text ".github/workflows/docker.yml" "type=raw,value=latest"
require_text ".github/workflows/docker.yml" "file: docker/Dockerfile"
require_text ".github/workflows/docker.yml" "go test ./..."

require_file "packaging/homebrew/mockport.rb.template"
require_text "packaging/homebrew/mockport.rb.template" "__VERSION__"
require_text "packaging/homebrew/mockport.rb.template" "__URL__"
require_text "packaging/homebrew/mockport.rb.template" "__SHA256__"
require_text "packaging/homebrew/mockport.rb.template" "bin.install"

require_file "packaging/npm/package.json"
require_text "packaging/npm/package.json" "\"bin\""
require_text "packaging/npm/package.json" "\"mockport\""
require_text "packaging/npm/package.json" "\"experimental\""
require_file "packaging/npm/bin/mockport.js"
require_text "packaging/npm/bin/mockport.js" "MOCKPORT_BIN"
require_text "packaging/npm/bin/mockport.js" "MOCKPORT_IMAGE"
require_text "packaging/npm/bin/mockport.js" "ghcr.io/albert-einshutoin/mockport:0.1.0-alpha"
require_text "packaging/npm/bin/mockport.js" "docker"

for page in index quickstart adapters ai-safe reports distribution; do
  require_file "docs/site/${page}.md"
done

require_text "docs/site/index.md" "quickstart.md"
require_text "docs/site/index.md" "adapters.md"
require_text "docs/site/index.md" "ai-safe.md"
require_text "docs/site/index.md" "reports.md"
require_text "docs/site/index.md" "distribution.md"
