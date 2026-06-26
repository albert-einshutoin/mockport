# Mockport TAKT Automation

This directory configures Mockport for local subscription-only TAKT/devloopd
automation.

## One Issue

```bash
devloopd run \
  --issue 123 \
  --repo albert-einshutoin/mockport \
  --workflow .takt/workflows/subscription-devloop.yaml \
  --verbose
```

## One Scan Cycle

```bash
devloopd start \
  --repo albert-einshutoin/mockport \
  --workflow .takt/workflows/subscription-devloop.yaml \
  --once
```

`devloopd start` uses TAKT's default issue scanner. Mockport currently has the
`bug` label, but not `agent:ready`, `tests`, or `docs`. For predictable daemon
selection, add an explicit `agent:ready` label to issues that are safe for
automation, or start with `devloopd run --issue <number>`.

## Agent Routing

The subscription workflow keeps expensive reasoning on planning and arbitration
while pushing implementation loops to lower-cost coding agents:

- `codex-cli` / `gpt-5.5-extra-high`: product-safe planning and final arbitration
- `cursor-cli` / `composer-2.5`: primary TDD implementation
- `opencode-cli` / `opencode-go/minimax-m3`: cheap verification fixes and hygiene review
- `agy-cli` / `Gemini 3.5 Flash (High)`: mergeability and security review

## Full Auto Loop

Use the Mockport wrapper when you want one command to:

1. create required automation labels,
2. mark one safe issue with `agent:ready`,
3. run `devloopd start --once`,
4. wait for PR checks,
5. post an `agy` mergeability review comment for the current PR head,
6. merge only when the PR passes local size/path guards and review says
   `Mergeable: YES`.

```bash
.takt/automation/full-auto-devloop.sh once
```

For continuous operation:

```bash
.takt/automation/full-auto-devloop.sh loop
```

The merge guard intentionally refuses large or sensitive PRs. Defaults:

- max changed files: `12`
- max changed lines: `500`
- forbidden paths: `.github/**`, infra/migration/auth/billing/payment paths,
  env files, and secret/credential-like paths

Override only for a deliberate local run:

```bash
MOCKPORT_TAKT_MAX_AUTO_MERGE_FILES=20 \
MOCKPORT_TAKT_MAX_AUTO_MERGE_LINES=800 \
.takt/automation/full-auto-devloop.sh once
```

To disable merge while still auto-labeling and creating PRs:

```bash
MOCKPORT_TAKT_AUTO_MERGE=0 .takt/automation/full-auto-devloop.sh loop
```

To disable the post-PR `agy` mergeability comment:

```bash
MOCKPORT_TAKT_PR_REVIEW=0 .takt/automation/full-auto-devloop.sh loop
```

To also create new low-risk issues from product docs when no safe issue exists:

```bash
MOCKPORT_TAKT_CREATE_ISSUES=1 .takt/automation/full-auto-devloop.sh loop
```

The issue crafter uses OpenCode/MiniMax M3 for cheap candidate scouting and
Codex/gpt-5.5-extra-high for final scoping. It only creates issues marked
`ready=true` and `risk=low` by the final planner.

You can preview the planned issues without creating them:

```bash
.takt/automation/create-product-issues.sh plan
```

## Gates

The workflow runs `.takt/quality-gates/mockport-check.sh` in the OpenCode
verification step after Cursor finishes the primary implementation.
Default mode is intentionally a fast PR gate:

```bash
.takt/quality-gates/mockport-check.sh
```

Use the full mode before finalizing larger adapter, compatibility, security, or
distribution changes:

```bash
MOCKPORT_TAKT_GATE_MODE=full .takt/quality-gates/mockport-check.sh
```
