# 04. Go Engineering Guide

[日本語版](04_go_engineering_guide.ja.md)

## Go version

Use Go `1.26.3` for initial development.

Reason:

- It is the current latest stable Go patch release as of 2026-05-25.
- It includes security fixes and runtime/toolchain bug fixes.
- Go is a strong fit for Mockport because it is excellent for HTTP servers, CLIs, Docker images, and single-binary distribution.

## Why Go

Mockport is primarily:

- HTTP server
- reverse/service emulator
- CLI
- Docker-first runtime
- concurrent request handler
- webhook sender
- config-driven tool

Go fits these requirements well.

## Go design principles for Mockport

### 1. Prefer simple packages

Avoid over-abstracted architecture in the MVP.

Good:

```txt
internal/config
internal/server
internal/adapter
internal/scenario
adapters/stripe
```

Avoid:

```txt
domain/usecase/application/infrastructure
```

Mockport is an infrastructure tool, not an enterprise business app.

### 2. Keep interfaces small

Interfaces should emerge from usage.

Start with:

```go
type Adapter interface {
    Name() string
    Register(mux *http.ServeMux, cfg AdapterConfig) error
    FakeEnv(cfg AdapterConfig) map[string]string
}
```

Do not design a large plugin abstraction before real adapter pain appears.

### 3. Use standard library first

Recommended:

- `net/http`
- `httptest`
- `context`
- `log/slog`
- `encoding/json`
- `os`
- `os/exec`
- `time`

Small external dependencies:

- `github.com/spf13/cobra` for CLI
- `gopkg.in/yaml.v3` for YAML

### 4. Keep cmd thin

`cmd/mockport/main.go` should only call the CLI root command.

Business logic belongs in `internal`.

### 5. Make Docker the primary runtime

The binary should run locally, but product UX should center on Docker:

```bash
docker run -p 127.0.0.1:43101:43101 ghcr.io/albert-einshutoin/mockport:0.1.0-alpha
```

### 6. Build statically

Use:

```bash
CGO_ENABLED=0 go build -o mockport ./cmd/mockport
```

### 7. Context-aware server lifecycle

Use `context.Context` for cancellation and graceful shutdown.

### 8. Test with `httptest`

Adapter behavior should be testable without Docker.

## Recommended dependencies

```txt
Go:
  1.26.3

CLI:
  github.com/spf13/cobra

YAML:
  gopkg.in/yaml.v3

Logging:
  log/slog

Testing:
  testing
  httptest
```

## Avoid initially

- Go plugin system
- embedded scripting runtime
- custom DSL parser
- full OpenAPI generation
- Docker API SDK
- Kubernetes support
- complex dependency injection framework

## Coding style

- Clear names over clever names
- Return errors with context
- Keep packages small
- Prefer table-driven tests
- Avoid global mutable state except adapter registry if carefully controlled
- Do not log secrets
- Redact values in warnings and reports
