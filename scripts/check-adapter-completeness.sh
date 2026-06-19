#!/usr/bin/env bash
set -euo pipefail

MISSING=0
ADAPTERS=()

require_file() {
  local path="$1"
  if [[ ! -f "$path" ]]; then
    echo "missing required file: $path" >&2
    MISSING=1
  fi
}

has_adapter() {
  local name="$1"
  local adapter
  for adapter in "${ADAPTERS[@]-}"; do
    if [[ "$adapter" == "$name" ]]; then
      return 0
    fi
  done
  return 1
}

# Discover built-in adapter names from `builtinAdapters()` registration using POSIX tools
# (CI runner environments do not guarantee GNU ripgrep availability).
while IFS= read -r package_name; do
  adapter_file="adapters/${package_name}/adapter.go"
  adapter_name=""

  if ! [[ -f "$adapter_file" ]]; then
    echo "adapters/${package_name}/adapter.go is missing" >&2
    MISSING=1
    continue
  fi

  adapter_name=$(awk '/func \(a Adapter\) Name\(\) string/{in_name=1} in_name && /return[[:space:]]*"/ {line=$0; sub(/.*return[[:space:]]*"/, "", line); sub(/".*/, "", line); print line; exit} in_name && /^\}/ {in_name=0}' "$adapter_file")

  if [[ -z "$adapter_name" ]]; then
    echo "$adapter_file missing Name() return string" >&2
    MISSING=1
    continue
  fi

  if ! has_adapter "$adapter_name"; then
    ADAPTERS+=("$adapter_name")
  fi
done < <(grep -oE '[A-Za-z_][A-Za-z0-9_]*\.New\(\)' internal/cli/builtin.go | sed 's/\.New()$//' | sort -u)

if [[ ${#ADAPTERS[@]} -eq 0 ]]; then
  echo "failed to resolve built-in adapters" >&2
  exit 1
fi

require_file "configs/mockport.example.yml"
require_file "docs/site/support-matrix.md"

for adapter in "${ADAPTERS[@]}"; do
  if ! grep -Fq "  ${adapter}:" configs/mockport.example.yml; then
    echo "skipping completeness check for ${adapter}: not listed in sample config"
    continue
  fi

  require_file "docs/adapters/${adapter}.md"

  if ! grep -Fq "\`${adapter}\`" docs/site/support-matrix.md; then
    echo "missing support-matrix entry for adapter: ${adapter}" >&2
    MISSING=1
  fi

  if ! grep -Fq "  ${adapter}:" configs/mockport.example.yml; then
    echo "configs/mockport.example.yml missing adapter entry: ${adapter}" >&2
    MISSING=1
  fi

  if [[ ! -d "examples/${adapter}" ]]; then
    found_example=0
    while IFS= read -r example_file; do
      if grep -q "^  ${adapter}:" "$example_file"; then
        found_example=1
        break
      fi
    done < <(find examples -type f -name mockport.yml)

    if [[ $found_example -eq 0 ]]; then
      echo "missing example config containing adapter name: ${adapter}" >&2
      MISSING=1
    fi
  fi
done

if [[ $MISSING -ne 0 ]]; then
  exit 1
fi

echo "check-adapter-completeness passed for: ${ADAPTERS[*]}"
