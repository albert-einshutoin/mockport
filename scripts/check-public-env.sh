#!/usr/bin/env bash
set -euo pipefail

files=()

collect_public_files() {
  local roots=(
    ".github/ISSUE_TEMPLATE"
    ".github/pull_request_template.md"
    ".github/workflows"
    "CHANGELOG.ja.md"
    "CHANGELOG.md"
    "CODE_OF_CONDUCT.ja.md"
    "CODE_OF_CONDUCT.md"
    "CONTRIBUTING.ja.md"
    "CONTRIBUTING.md"
    "README.ja.md"
    "README.md"
    "ROADMAP.ja.md"
    "ROADMAP.md"
    "SECURITY.ja.md"
    "SECURITY.md"
    "configs"
    "contract"
    "docs"
    "examples"
    "packaging"
  )

  local root
  for root in "${roots[@]}"; do
    [[ -e "$root" ]] || continue
    if [[ -f "$root" ]]; then
      files+=("$root")
      continue
    fi

    while IFS= read -r file; do
      files+=("$file")
    done < <(
      find "$root" \
        \( -path '*/node_modules' -o -path '*/node_modules/*' -o -path '*/dist' -o -path '*/dist/*' -o -path '*/bin' -o -path '*/bin/*' \) -prune -o \
        -type f \( -name '*.md' -o -name '*.yml' -o -name '*.yaml' -o -name '.env*' -o -name 'env.*' -o -name '*.env' -o -name '*.env.*' \) \
        -print
    )
  done
}

report_finding() {
  local file="$1"
  local line="$2"
  local reason="$3"
  printf '%s:%s: %s\n' "$file" "$line" "$reason" >&2
}

line_has_webhook_secret() {
  local line="$1"
  [[ "$line" == *"whsec_"* && "$line" != *"whsec_mockport"* ]]
}

line_has_ambiguous_placeholder() {
  local line="$1"
  [[ "$line" =~ (^|[^[:alnum:]])(change[-_\ ]?me|replace[-_\ ]?me|changeme)([^[:alnum:]]|$) ]]
}

collect_public_files
sorted_files=()
while IFS= read -r file; do
  [[ -n "$file" ]] && sorted_files+=("$file")
done < <(printf '%s\n' "${files[@]}" | sort -u)
files=("${sorted_files[@]}")

secret_re='sk_(live|test)_|AKIA|ASIA|ghp_|github_pat_|xox[baprs]-|AIza'
url_re='https://api\.stripe\.com|https://api\.openai\.com|https://api\.github\.com|https://api\.line\.me|https://slack\.com/api|https://hooks\.slack\.com'
failed=0
shopt -s nocasematch

for file in "${files[@]}"; do
  allow_block=""
  line_no=0
  while IFS= read -r line || [[ -n "$line" ]]; do
    line_no=$((line_no + 1))

    if [[ "$line" == *"mockport-public-safety: allow-begin"* ]]; then
      allow_block="$line"
      continue
    fi
    if [[ "$line" == *"mockport-public-safety: allow-end"* ]]; then
      if [[ -z "$allow_block" ]]; then
        report_finding "$file" "$line_no" "unmatched public-safety allow-end marker"
        failed=1
      fi
      allow_block=""
      continue
    fi
    [[ -n "$allow_block" ]] && continue

    if [[ "$line" =~ $secret_re ]]; then
      report_finding "$file" "$line_no" "real-looking provider secret"
      failed=1
    fi
    if line_has_webhook_secret "$line"; then
      report_finding "$file" "$line_no" "real-looking webhook secret"
      failed=1
    fi
    if [[ "$line" =~ $url_re ]]; then
      report_finding "$file" "$line_no" "production provider URL"
      failed=1
    fi
    if line_has_ambiguous_placeholder "$line"; then
      report_finding "$file" "$line_no" "ambiguous placeholder"
      failed=1
    fi
  done < "$file"

  if [[ -n "$allow_block" ]]; then
    report_finding "$file" "$line_no" "unclosed public-safety allow-begin marker"
    failed=1
  fi
done

exit "$failed"
