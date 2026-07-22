# Selective CI Test Execution

Mockport's pull request CI uses conservative change-impact analysis to shorten
feedback time without weakening failure detection. Pushes to `main` and
`release/**`, manual runs, and the daily scheduled run continue to execute the
full suite.

## Safety model

The planner follows a fail-closed policy:

1. Resolve the merge base between the current head and the latest target branch.
2. Parse additions, modifications, deletions, copies, and renames from Git.
3. Apply the centralized full-suite rules in `ci/config.json`.
4. Ask the Go adapter for changed packages and all in-repository reverse
   dependents using the import graph from `go list -json ./...`.
5. Add mapped integration and E2E checks plus the unconditional health smoke
   test.
6. Fall back to the full suite when any input is missing, unsupported, or
   unclassified.

Deleted files are used for classification but are never passed to a tool that
expects the file to exist. Changed paths are never interpolated into a shell
command; the runner executes trusted argument arrays from configuration.

Shared contracts, configuration, routing, security, state, build, dependency,
test, container, and CI changes intentionally trigger the full suite. Changing
the impact analyzer itself also triggers the full suite.

## Architecture

- `ci/impact.py`: CI-service-neutral diff handling, classification, planning,
  machine-readable output, execution, metrics, and fallback control.
- `ci/adapters/go.py`: Go project detection, package graph decoding, and reverse
  dependency traversal.
- `ci/config.json`: full-suite rules, path mappings, smoke tests, and trusted
  command definitions.
- `ci/run-full.sh`: independent emergency lane used even when the planner or its
  configuration cannot start.

The common planner does not contain Go-specific selection logic. Another
language can be added as an adapter that accepts repository paths and returns
detected projects, affected modules, and unit targets using the same JSON
contract.

## Local verification

```sh
make ci-impact-test
python3 ci/impact.py plan --base origin/main --head HEAD --output ci-plan.json
python3 ci/impact.py run --plan ci-plan.json
```

To exercise the complete lane used after merges and before releases:

```sh
make ci-full
```

`ci-plan.json` records the revisions, changed files, detected and affected
projects, affected modules, unit/integration/E2E/smoke targets, strategy, and
fallback reason. CI uploads it for 14 days so selection decisions remain
auditable.

When adding a path or project type, first add a failing planner or adapter test.
Any path that is not explicitly understood must remain a full-suite fallback.
