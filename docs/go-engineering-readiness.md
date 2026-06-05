# Go Engineering Readiness

[日本語版](go-engineering-readiness.ja.md)

Last updated: 2026-05-26

## Required Gates

| Gate | Command | Status |
| --- | --- | --- |
| Unit and integration tests | `go test ./...` | required |
| Vet | `go vet ./...` | required |
| Race detection | `go test -race ./...` | required in CI |
| Staticcheck | `staticcheck ./...` | required in CI |
| Vulnerability scan | `govulncheck ./...` | required in CI |
| Go engineering static policy | `bash scripts/check-go-engineering.sh` | required in CI |

## Current Accepted Gaps

- `map[string]any` remains allowed for dynamic provider payload storage and provider-specific fixture-like fields.
- Provider internals are not reproduced; Mockport focuses on public API and SDK/client workflow compatibility.
- Benchmarks are lightweight regression sentinels, not formal performance guarantees.

## Ignored Error Policy

Ignored errors are allowed only for best-effort response writes, best-effort response body closes in tests or outbound HTTP cleanup, and test helpers where the next assertion observes the failure. Critical setup, listener, registry, metadata, JSON encode, config, state, and outbound request errors must be returned or asserted.

## Readiness Definition

Mockport is Go 10/10 ready when the required gates pass, compatibility metadata validates without silent level loss, core response builders use typed structs, server runtime supports graceful shutdown, request bodies are bounded, shared state deep-clones nested data, and the generated compatibility reports match documented claims.
