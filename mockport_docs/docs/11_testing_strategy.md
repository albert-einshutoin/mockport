# 11. Testing Strategy

[日本語版](11_testing_strategy.ja.md)

## Test philosophy

Mockport must be trusted as infrastructure. Tests should cover:

- config loading
- server routing
- adapter behavior
- scenario behavior
- security warnings
- report generation
- CLI output

## Test levels

### Unit tests

Use Go `testing`.

Targets:

- config validation
- secret detection
- URL detection
- scenario selection
- response builders

### HTTP tests

Use `httptest`.

Targets:

- `/health`
- adapter endpoints
- report endpoint
- error responses

### CLI tests

Use temporary directories.

Targets:

- `mockport init`
- generated files
- invalid config handling

### Docker smoke test

In CI or local release workflow:

```bash
docker build -t mockport:test -f docker/Dockerfile .
docker run -d -p 127.0.0.1:43101:43101 --name mockport-test mockport:test
curl http://localhost:43101/health
docker rm -f mockport-test
```

## Minimal MVP tests

```txt
config:
- load valid YAML
- reject invalid port
- apply default host
- reject missing adapter base path

security:
<!-- mockport-public-safety: allow-begin detector-reference -->
- detect sk_live_
- detect AKIA
<!-- mockport-public-safety: allow-end -->
- allow mockport_ prefix
- redact secret

server:
- health returns 200
- report returns JSON
- unknown route returns 404

stripe:
- checkout session success returns 200
- payment failure returns 402
- auth error returns 401
- rate limit returns 429
- webhook send builds payload
```

## Table-driven tests

Preferred style:

```go
func TestDetectSecret(t *testing.T) {
    tests := []struct {
        name string
        value string
        want bool
    }{
        // mockport-public-safety: allow-begin detector-reference
        {"stripe live", "sk_live_xxx", true},
        // mockport-public-safety: allow-end
        {"mockport fake", "mockport_stripe_secret", false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := security.LooksLikeSecret(tt.value)
            if got != tt.want {
                t.Fatalf("got %v, want %v", got, tt.want)
            }
        })
    }
}
```

## GitHub Actions

Minimal CI:

```yaml
name: CI

on:
  push:
  pull_request:

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.26.3"
      - run: go test ./...
      - run: go vet ./...
      - run: go build ./cmd/mockport
```
