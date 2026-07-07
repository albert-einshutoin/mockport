# Compatibility Manifests

Checked-in compatibility manifests are versioned evidence for selected provider workflows. They record adapter metadata claims—maturity, levels, endpoints, scenarios, SDK versions, and known gaps—so compatibility changes produce reviewable diffs instead of hiding behind runtime-only state.

Manifests are evidence for selected workflows, not a claim that Mockport fully clones a provider. Provider internals, undocumented behavior, production network effects, and broad surface-area parity are out of scope. See [Compatibility Model](../docs/compatibility-model.md) for the full boundary, levels, scoring rules, and evidence workflow.

## Manifest Index

Manifests live under `compat/manifests/{adapter}.json`. Each file is the JSON output of `compat.FromMetadata()` for the adapter's current `Metadata()` claim.

| Manifest | Adapter |
| --- | --- |
| `github-oauth.json` | GitHub OAuth |
| `line.json` | LINE |
| `openai.json` | OpenAI |
| `slack.json` | Slack |
| `stripe.json` | Stripe |
| `zoho-oauth.json` | Zoho OAuth |

List the current files:

```bash
ls compat/manifests
```

## Maturity Promotion Gate

Adapter maturity is promoted only when implementation, metadata, tests, fixtures or SDK evidence, known-gap documentation, and generated reports agree. Docs alone cannot raise maturity.

| Maturity | Minimum evidence |
| --- | --- |
| `experimental` | Adapter exists with explicit metadata and known gaps. |
| `partial` | Common scenario-compatible paths are implemented and reported. |
| `sdk-compatible` | SDK level evidence exists and selected SDK contracts pass. |
| `workflow-compatible` | Workflow, state, and error evidence exists for selected workflows. |
| `provider-compatible` | Contract level evidence exists with manifest, fixtures, SDK/client contracts, workflow/state/error coverage, score, and known gaps. |

`workflow-compatible` promotion requires full state and error coverage, not level declarations alone. `provider-compatible` promotion additionally requires `contract_evidence` with fixture, SDK/client contract, and known-gap publication evidence.

CI runs `scripts/check-compat-manifests.sh` to fail on drift between checked-in manifests and regenerated output. The compatibility release gate also requires report maturity to match checked-in manifest maturity for every published adapter.

## Verification Commands

Use these commands to validate public trust surfaces, SDK contracts, and the Go runtime:

```bash
ls compat/manifests
bash scripts/check-public-trust.sh
bash scripts/run-sdk-contracts.sh all
go test ./...
```

`bash scripts/run-sdk-contracts.sh all` runs SDK/client smoke tests for `stripe`, `openai`, `github-oauth`, and `slack`. Other adapters may have manifest and fixture evidence without a harness entry in that script yet.
