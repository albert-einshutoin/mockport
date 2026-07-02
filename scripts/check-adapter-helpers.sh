#!/usr/bin/env bash
set -euo pipefail

# Reports unexported helper names duplicated across adapter packages.
# Duplication is a tracking signal only; identical behavior is not inferred.
# Fails when any helper name appears in more adapter packages than
# DUPLICATE_ADAPTER_THRESHOLD (default: built-in adapter package count).

ADAPTERS_DIR="adapters"
EXCLUDED_HELPERS='^(New|Name|Register|FakeEnv|Metadata|handle|handleReset)$'

if [[ ! -d "$ADAPTERS_DIR" ]]; then
  echo "missing directory: $ADAPTERS_DIR" >&2
  exit 1
fi

adapter_packages=()
for package_dir in "$ADAPTERS_DIR"/*/; do
  [[ -d "$package_dir" ]] || continue
  adapter_packages+=("$(basename "$package_dir")")
done

if [[ ${#adapter_packages[@]} -eq 0 ]]; then
  echo "no adapter packages found under $ADAPTERS_DIR" >&2
  exit 1
fi

threshold="${DUPLICATE_ADAPTER_THRESHOLD:-${#adapter_packages[@]}}"

helper_map="$(mktemp)"
trap 'rm -f "$helper_map" "${helper_map}.uniq"' EXIT

for package_name in "${adapter_packages[@]}"; do
  package_dir="$ADAPTERS_DIR/$package_name"
  while IFS= read -r go_file; do
    while IFS= read -r helper_name; do
      [[ -z "$helper_name" ]] && continue
      if [[ "$helper_name" =~ $EXCLUDED_HELPERS ]]; then
        continue
      fi
      printf '%s|%s\n' "$helper_name" "$package_name"
    done < <(
      sed -nE \
        -e 's/^func ([a-z][a-zA-Z0-9_]*)\(.*/\1/p' \
        -e 's/^func ([^)]*) ([a-z][a-zA-Z0-9_]*)\(.*/\2/p' \
        "$go_file"
    )
  done < <(find "$package_dir" -maxdepth 1 -type f -name '*.go' ! -name '*_test.go' | sort)
done >"$helper_map"

sort -u "$helper_map" >"${helper_map}.uniq"

duplicate_count=0
threshold_exceeded=0

while IFS= read -r helper_name; do
  adapter_list=$(
    grep -F "${helper_name}|" "${helper_map}.uniq" |
      cut -d'|' -f2 |
      sort -u |
      tr '\n' ' ' |
      sed 's/ $//'
  )
  adapter_total=$(
    grep -F "${helper_name}|" "${helper_map}.uniq" |
      cut -d'|' -f2 |
      sort -u |
      wc -l |
      tr -d ' '
  )

  if [[ "$adapter_total" -lt 2 ]]; then
    continue
  fi

  duplicate_count=$((duplicate_count + 1))
  echo "duplicate helper: ${helper_name} (${adapter_total} adapters: ${adapter_list})"

  if [[ "$adapter_total" -gt "$threshold" ]]; then
    threshold_exceeded=1
    echo "  exceeds DUPLICATE_ADAPTER_THRESHOLD=${threshold}" >&2
  fi
done < <(cut -d'|' -f1 "${helper_map}.uniq" | sort -u)

if [[ "$duplicate_count" -eq 0 ]]; then
  echo "check-adapter-helpers: no duplicated unexported helper names found"
else
  echo "check-adapter-helpers: ${duplicate_count} duplicated helper name(s) tracked (threshold=${threshold})"
fi

if [[ "$threshold_exceeded" -ne 0 ]]; then
  echo "check-adapter-helpers failed: one or more helpers exceed DUPLICATE_ADAPTER_THRESHOLD=${threshold}" >&2
  exit 1
fi

echo "check-adapter-helpers passed"
