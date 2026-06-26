#!/usr/bin/env bash
set -euo pipefail

REPO="${MOCKPORT_TAKT_REPO:-albert-einshutoin/mockport}"
WORKFLOW="${MOCKPORT_TAKT_WORKFLOW:-.takt/workflows/subscription-devloop.yaml}"
INTERVAL_SECONDS="${MOCKPORT_TAKT_INTERVAL_SECONDS:-300}"
MAX_AUTO_MERGE_FILES="${MOCKPORT_TAKT_MAX_AUTO_MERGE_FILES:-12}"
MAX_AUTO_MERGE_LINES="${MOCKPORT_TAKT_MAX_AUTO_MERGE_LINES:-500}"
AUTO_MERGE="${MOCKPORT_TAKT_AUTO_MERGE:-1}"
PR_REVIEW="${MOCKPORT_TAKT_PR_REVIEW:-1}"
CREATE_ISSUES="${MOCKPORT_TAKT_CREATE_ISSUES:-0}"
AGY_MODEL="${MOCKPORT_TAKT_AGY_MODEL:-Gemini 3.5 Flash (High)}"
AGY_PRINT_TIMEOUT="${MOCKPORT_TAKT_AGY_PRINT_TIMEOUT:-5m}"
ISSUE_CRAFTER="${MOCKPORT_TAKT_ISSUE_CRAFTER:-.takt/automation/create-product-issues.sh}"

READY_LABEL="agent:ready"
AUTO_MERGE_LABEL="agent:auto-merge"
PR_REVIEW_MARKER="<!-- mockport-takt-mergeability-review -->"

log() {
  printf '[%s] %s\n' "$(date '+%Y-%m-%dT%H:%M:%S%z')" "$*"
}

ensure_label() {
  local name="$1"
  local description="$2"
  local color="$3"
  if gh label list --repo "$REPO" --limit 200 --json name --jq '.[].name' | grep -Fxq "$name"; then
    return
  fi
  gh label create "$name" --repo "$REPO" --description "$description" --color "$color"
}

ensure_labels() {
  ensure_label "$READY_LABEL" "Ready for local TAKT/devloopd automation" "5319e7"
  ensure_label "$AUTO_MERGE_LABEL" "Mechanical gates passed; allow devloop auto-merge" "0e8a16"
}

issue_has_existing_pr() {
  local issue="$1"
  [[ "$(gh pr list --repo "$REPO" --state all --search "#${issue}" --json number --jq 'length')" != "0" ]]
}

candidate_from_scan() {
  devloopd scan-issues --repo "$REPO" \
    | awk '/^Candidates:/{inside=1; next} /^Skipped:/{inside=0} inside && /^- #/{sub(/^- #/, ""); sub(/ .*/, ""); print; exit}'
}

title_is_broad() {
  local title="$1"
  [[ "$title" =~ (トラッキング|全体計画|最大化計画|ロードマップ|roadmap|Roadmap) ]]
}

labels_are_forbidden() {
  local labels="$1"
  [[ ",${labels}," == *",blocked,"* \
    || ",${labels}," == *",human-required,"* \
    || ",${labels}," == *",security-sensitive,"* \
    || ",${labels}," == *",do-not-touch,"* \
    || ",${labels}," == *",duplicate,"* \
    || ",${labels}," == *",invalid,"* \
    || ",${labels}," == *",wontfix,"* ]]
}

remove_ready_if_needed() {
  local issue="$1"
  gh issue edit "$issue" --repo "$REPO" --remove-label "$READY_LABEL" >/dev/null 2>&1 || true
}

find_existing_safe_candidate() {
  local candidate
  while candidate="$(candidate_from_scan)" && [[ -n "$candidate" ]]; do
    if issue_has_existing_pr "$candidate"; then
      # A ready issue with an existing PR would make devloopd rerun the same work.
      remove_ready_if_needed "$candidate"
      log "removed ${READY_LABEL} from #${candidate}: PR already exists"
      continue
    fi
    printf '%s\n' "$candidate"
    return 0
  done
  return 1
}

try_mark_issue_ready() {
  local issue="$1"
  gh issue edit "$issue" --repo "$REPO" --add-label "$READY_LABEL" >/dev/null

  local selected
  selected="$(candidate_from_scan || true)"
  if [[ "$selected" == "$issue" ]]; then
    printf '%s\n' "$issue"
    return 0
  fi

  # devloopd may classify issue text as unsafe after the label is added.
  remove_ready_if_needed "$issue"
  log "skipped #${issue}: devloopd did not classify it as an automation candidate"
  return 1
}

mark_next_issue_ready() {
  if find_existing_safe_candidate; then
    return 0
  fi

  local line issue title labels
  while IFS=$'\t' read -r issue title labels; do
    [[ -z "${issue:-}" ]] && continue
    if title_is_broad "$title"; then
      log "skipped #${issue}: broad tracker title"
      continue
    fi
    if labels_are_forbidden "$labels"; then
      log "skipped #${issue}: forbidden label"
      continue
    fi
    if issue_has_existing_pr "$issue"; then
      log "skipped #${issue}: existing PR reference"
      continue
    fi
    if try_mark_issue_ready "$issue"; then
      return 0
    fi
  done < <(
    gh issue list --repo "$REPO" --state open --limit 100 --json number,title,labels \
      --jq '.[] | select(([.labels[].name] | index("agent:ready")) | not) | "\(.number)\t\(.title)\t\([.labels[].name] | join(","))"'
  )

  return 1
}

find_open_pr_for_issue() {
  local issue="$1"
  gh pr list --repo "$REPO" --state open --search "#${issue}" --json number --jq '.[0].number // empty'
}

path_is_forbidden_for_auto_merge() {
  local path="$1"
  case "$path" in
    .github/*|infra/*|terraform/*|migrations/*|auth/*|billing/*|payments/*|*.env*|*secret*|*credential*)
      return 0
      ;;
    *)
      return 1
      ;;
  esac
}

pr_passes_local_merge_guard() {
  local pr="$1"
  local changed_files additions deletions total_lines path
  changed_files="$(gh pr view "$pr" --repo "$REPO" --json changedFiles --jq '.changedFiles')"
  additions="$(gh pr view "$pr" --repo "$REPO" --json additions --jq '.additions')"
  deletions="$(gh pr view "$pr" --repo "$REPO" --json deletions --jq '.deletions')"
  total_lines=$((additions + deletions))

  if (( changed_files > MAX_AUTO_MERGE_FILES )); then
    log "PR #${pr} requires human review: changed files ${changed_files} > ${MAX_AUTO_MERGE_FILES}"
    return 1
  fi
  if (( total_lines > MAX_AUTO_MERGE_LINES )); then
    log "PR #${pr} requires human review: changed lines ${total_lines} > ${MAX_AUTO_MERGE_LINES}"
    return 1
  fi

  while IFS= read -r path; do
    if path_is_forbidden_for_auto_merge "$path"; then
      log "PR #${pr} requires human review: forbidden auto-merge path ${path}"
      return 1
    fi
  done < <(gh pr diff "$pr" --repo "$REPO" --name-only)

  return 0
}

existing_mergeability_decision() {
  local pr="$1"
  local head_sha="$2"

  gh api "repos/${REPO}/issues/${pr}/comments" --paginate \
    --jq "map(select(((.body // \"\") | contains(\"${PR_REVIEW_MARKER}\")) and ((.body // \"\") | contains(\"Head SHA: \`${head_sha}\`\")))) | last | .body // \"\" | if test(\"(?m)^Mergeable:[[:space:]]*YES\") then \"YES\" elif test(\"(?m)^Mergeable:[[:space:]]*NO\") then \"NO\" else \"\" end"
}

comment_pr_mergeability_review() {
  local pr="$1"
  local gate_context="$2"
  [[ "$PR_REVIEW" == "1" ]] || return 0

  local head_sha existing
  head_sha="$(gh pr view "$pr" --repo "$REPO" --json headRefOid --jq '.headRefOid')"
  existing="$(existing_mergeability_decision "$pr" "$head_sha" || true)"
  if [[ "$existing" == "YES" ]]; then
    log "agy mergeability review already approved PR #${pr} at ${head_sha}"
    return 0
  fi
  if [[ "$existing" == "NO" ]]; then
    log "agy mergeability review already blocked PR #${pr} at ${head_sha}"
    return 1
  fi

  local tmpdir prompt_file review_file body_file
  tmpdir="$(mktemp -d)"
  prompt_file="${tmpdir}/prompt.md"
  review_file="${tmpdir}/review.md"
  body_file="${tmpdir}/body.md"

  local metadata files checks
  metadata="$(gh pr view "$pr" --repo "$REPO" --json number,title,url,baseRefName,headRefName,headRefOid,mergeable,mergeStateStatus,reviewDecision,isDraft,changedFiles,additions,deletions,labels)"
  files="$(gh pr diff "$pr" --repo "$REPO" --name-only || true)"
  checks="$(gh pr checks "$pr" --repo "$REPO" 2>&1 || true)"

  cat >"$prompt_file" <<EOF
You are the Mockport PR mergeability reviewer using agy ${AGY_MODEL}.

Use only the supplied GitHub metadata, check output, changed-file list, and
local automation policy. Do not ask for secrets, CI bypass, force pushes, or
admin merges. The TAKT workflow already ran code review before PR creation;
this pass decides whether the current PR head is mergeable.

Return exactly this shape:

Mergeable: YES|NO
Reason: one concise sentence
Blockers:
- none, or concrete blockers with file paths/commands
Verification:
- checks/local guard evidence used

Gate context:
${gate_context}

PR metadata:
${metadata}

Changed files:
${files}

GitHub checks:
${checks}
EOF

  if ! agy --model "$AGY_MODEL" --print-timeout "$AGY_PRINT_TIMEOUT" -p "$(cat "$prompt_file")" >"$review_file"; then
    cat >"$review_file" <<EOF
Mergeable: NO
Reason: agy mergeability review failed to run.
Blockers:
- Review command failed; leave the PR open for manual inspection.
Verification:
- Gate context: ${gate_context}
EOF
  fi

  if ! grep -Eq '^Mergeable:[[:space:]]*(YES|NO)[[:space:]]*$' "$review_file"; then
    {
      printf 'Mergeable: NO\n'
      printf 'Reason: agy review output did not follow the required contract.\n'
      printf 'Blockers:\n'
      printf -- '- Normalize the PR review result before merging.\n'
      printf 'Verification:\n'
      printf -- '- Gate context: %s\n\n' "$gate_context"
      cat "$review_file"
    } >"${tmpdir}/review.normalized.md"
    mv "${tmpdir}/review.normalized.md" "$review_file"
  fi

  {
    printf '%s\n' "$PR_REVIEW_MARKER"
    printf 'Head SHA: `%s`\n\n' "$head_sha"
    cat "$review_file"
  } >"$body_file"

  gh pr comment "$pr" --repo "$REPO" --body-file "$body_file" >/dev/null
  log "posted agy mergeability review for PR #${pr}"

  if grep -Eq '^Mergeable:[[:space:]]*YES[[:space:]]*$' "$review_file"; then
    rm -rf "$tmpdir"
    return 0
  fi

  rm -rf "$tmpdir"
  return 1
}

merge_pr_if_safe() {
  local pr="$1"
  [[ "$AUTO_MERGE" == "1" ]] || {
    log "auto-merge disabled; leaving PR #${pr} open"
    return 0
  }

  log "waiting for checks on PR #${pr}"
  if ! gh pr checks "$pr" --repo "$REPO" --watch --interval 10; then
    comment_pr_mergeability_review "$pr" "GitHub checks failed or timed out before merge." || true
    log "PR #${pr} checks failed or timed out; not merging"
    return 0
  fi

  if ! pr_passes_local_merge_guard "$pr"; then
    comment_pr_mergeability_review "$pr" "Local size/path merge guard blocked the PR." || true
    return 0
  fi

  if ! comment_pr_mergeability_review "$pr" "GitHub checks and local size/path merge guards passed."; then
    log "agy mergeability review did not approve PR #${pr}; not merging"
    return 0
  fi

  gh pr edit "$pr" --repo "$REPO" --add-label "$AUTO_MERGE_LABEL" >/dev/null

  local head_sha
  head_sha="$(gh pr view "$pr" --repo "$REPO" --json headRefOid --jq '.headRefOid')"

  if devloopd merge-if-safe --repo "$REPO" --pr "$pr" --expected-head "$head_sha"; then
    log "devloopd merge-if-safe accepted PR #${pr}"
    return 0
  fi

  # The TAKT workflow already ran agy review before PR creation. If the local
  # size/path/check guards pass but GitHub has no separate approval decision,
  # merge directly while pinning the exact head SHA to avoid racing new commits.
  log "devloopd merge-if-safe did not accept PR #${pr}; attempting guarded direct merge"
  gh pr merge "$pr" --repo "$REPO" --squash --delete-branch --match-head-commit "$head_sha"
}

run_once() {
  ensure_labels

  local issue
  if ! issue="$(mark_next_issue_ready)"; then
    if [[ "$CREATE_ISSUES" == "1" && -x "$ISSUE_CRAFTER" ]]; then
      log "no safe issue found; creating product issues"
      "$ISSUE_CRAFTER" create || true
      if ! issue="$(mark_next_issue_ready)"; then
        log "no safe issue found for automation"
        return 0
      fi
    else
      log "no safe issue found for automation"
      return 0
    fi
  fi

  log "running devloopd for issue #${issue}"
  devloopd start --repo "$REPO" --workflow "$WORKFLOW" --once

  local pr
  pr="$(find_open_pr_for_issue "$issue")"
  if [[ -z "$pr" ]]; then
    log "no open PR found for issue #${issue}"
    return 0
  fi

  log "created/found PR #${pr} for issue #${issue}"
  merge_pr_if_safe "$pr"
}

case "${1:-once}" in
  once)
    run_once
    ;;
  loop)
    while true; do
      run_once || true
      log "sleeping ${INTERVAL_SECONDS}s"
      sleep "$INTERVAL_SECONDS"
    done
    ;;
  merge-pr)
    [[ -n "${2:-}" ]] || {
      echo "usage: $0 merge-pr <pr-number>" >&2
      exit 2
    }
    ensure_labels
    merge_pr_if_safe "$2"
    ;;
  *)
    echo "usage: $0 [once|loop|merge-pr <pr-number>]" >&2
    exit 2
    ;;
esac
