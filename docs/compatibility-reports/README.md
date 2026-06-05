# Compatibility Reports

[日本語版](README.ja.md)

Compatibility reports are generated trust artifacts for the provider-compatible release track.

- `latest.md` is the human-readable report.
- `latest.json` is the machine-readable report consumed by release checks.
- Reports are generated from Mockport runtime compatibility metadata, SDK/client contract evidence, fixture checks, and known gaps.

Maturity labels:

| Label | Meaning |
| --- | --- |
| `experimental` | Early adapter coverage for selected workflows. Expect gaps. |
| `sdk-compatible` | Selected SDK or client contract calls pass against local Mockport. |
| `workflow-compatible` | Selected workflows include fake state, errors, and replayable behavior. |
| `provider-compatible` | Selected provider workflows are backed by manifests, SDK contracts, fixtures, scores, and known-gap reports. |

Generate locally:

```sh
bash scripts/generate-compatibility-report.sh
bash scripts/check-compatibility-release.sh
```
