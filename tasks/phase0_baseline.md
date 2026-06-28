# Phase 0 Baseline Implementation Plan

[日本語版](phase0_baseline.ja.md)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [x]`) syntax for tracking.

**Goal:** Mockport の Go repository skeleton を作り、CLI から server を起動し、Docker 上で `/health` が 200 を返す状態にする。

**Architecture:** `cmd/mockport/main.go` は薄く保ち、CLI は `internal/cli`、設定は `internal/config`、HTTP server は `internal/server` に分ける。Phase 0 では Stripe adapter は作らず、server lifecycle、config、health check、Docker/CI の土台だけを完成させる。

**Tech Stack:** Go 1.26.3, `github.com/spf13/cobra`, `gopkg.in/yaml.v3`, `net/http`, `httptest`, Docker distroless image.

---

## Files

- Create: `README.md`
- Create: `go.mod`
- Create: `Makefile`
- Create: `.gitignore`
- Create: `.dockerignore`
- Create: `cmd/mockport/main.go`
- Create: `internal/cli/root.go`
- Create: `internal/cli/version.go`
- Create: `internal/cli/run.go`
- Create: `internal/config/config.go`
- Create: `internal/config/loader.go`
- Create: `internal/config/validate.go`
- Create: `internal/config/config_test.go`
- Create: `internal/security/secrets.go`
- Create: `internal/security/secrets_test.go`
- Create: `internal/server/server.go`
- Create: `internal/server/health.go`
- Create: `internal/server/server_test.go`
- Create: `configs/mockport.example.yml`
- Create: `docker/Dockerfile`
- Create: `.github/workflows/ci.yml`

## Task P0-T01: Repository Skeleton And Go Module

**Status:** done

- [x] **Step 1: Create Go module files**

Create `go.mod`:

```go
module github.com/albert-einshutoin/mockport

go 1.26

require (
	github.com/spf13/cobra v1.10.1
	gopkg.in/yaml.v3 v3.0.1
)
```

Create `cmd/mockport/main.go`:

```go
package main

import (
	"os"

	"github.com/albert-einshutoin/mockport/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
```

- [x] **Step 2: Run compile check**

Run:

```bash
go test ./...
```

Expected: fail until `internal/cli` exists. This is the RED compile boundary for the first package.

- [x] **Step 3: Add minimal CLI package**

Create `internal/cli/root.go`:

```go
package cli

import "github.com/spf13/cobra"

func Execute() error {
	return NewRootCommand().Execute()
}

func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mockport",
		Short: "Secret-free service emulation for local and CI integration tests",
	}
	return cmd
}
```

- [x] **Step 4: Verify package compiles**

Run:

```bash
go mod tidy
go test ./...
```

Expected: PASS.

## Task P0-T02: Root Command And Version Command

**Status:** done

- [x] **Step 1: Write failing CLI tests**

Create `internal/cli/root_test.go`:

```go
package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootCommandShowsHelp(t *testing.T) {
	cmd := NewRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute help: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "Secret-free service emulation") {
		t.Fatalf("help output missing product description: %q", got)
	}
}

func TestVersionCommandPrintsVersion(t *testing.T) {
	cmd := NewRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"version"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute version: %v", err)
	}

	if got := strings.TrimSpace(out.String()); got != "mockport dev" {
		t.Fatalf("version output = %q, want %q", got, "mockport dev")
	}
}
```

- [x] **Step 2: Verify RED**

Run:

```bash
go test ./internal/cli -run 'TestRootCommandShowsHelp|TestVersionCommandPrintsVersion' -v
```

Expected: `TestVersionCommandPrintsVersion` fails because `version` is not registered.

- [x] **Step 3: Implement version command**

Create `internal/cli/version.go`:

```go
package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var Version = "dev"

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print Mockport version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(cmd.OutOrStdout(), "mockport %s\n", Version)
		},
	}
}
```

Modify `internal/cli/root.go` so `NewRootCommand` registers `newVersionCommand()`.

- [x] **Step 4: Verify GREEN**

Run:

```bash
go test ./internal/cli -v
```

Expected: PASS.

## Task P0-T03: Config Defaults And YAML Loader

**Status:** done

- [x] **Step 1: Write failing config tests**

Create `internal/config/config_test.go`:

```go
package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadValidConfigAppliesDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "mockport.yml")
	err := os.WriteFile(path, []byte("version: \"0.1\"\nserver:\n  port: 43101\nmode: ai-safe\nadapters:\n  stripe:\n    enabled: true\n    base_path: /stripe\n    scenario: payment_success\n    fake_secret: mockport_stripe_secret\n"), 0o644)
	if err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := LoadFile(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.Server.Host != "0.0.0.0" {
		t.Fatalf("host = %q, want default 0.0.0.0", cfg.Server.Host)
	}
	if cfg.Server.Port != 43101 {
		t.Fatalf("port = %d, want 43101", cfg.Server.Port)
	}
}

func TestLoadRejectsInvalidPort(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "mockport.yml")
	err := os.WriteFile(path, []byte("version: \"0.1\"\nserver:\n  port: 0\nmode: ai-safe\n"), 0o644)
	if err != nil {
		t.Fatalf("write config: %v", err)
	}

	if _, err := LoadFile(path); err == nil {
		t.Fatal("LoadFile returned nil error for invalid port")
	}
}
```

- [x] **Step 2: Verify RED**

Run:

```bash
go test ./internal/config -v
```

Expected: compile failure because `LoadFile` and config types do not exist.

- [x] **Step 3: Implement config package**

Create `internal/config/config.go`, `loader.go`, and `validate.go` with `Config`, `ServerConfig`, `AdapterConfig`, `LoadFile`, `ApplyDefaults`, and `Validate`. `Validate` must reject ports outside `1..65535`, unsupported modes outside `ai-safe`, `local`, `strict`, and enabled adapters with empty `base_path`.

- [x] **Step 4: Verify GREEN**

Run:

```bash
go test ./internal/config -v
```

Expected: PASS.

## Task P0-T04: Security Detector Primitives

**Status:** done

- [x] **Step 1: Write failing security tests**

Create `internal/security/secrets_test.go`:

```go
package security

import "testing"

func TestLooksLikeSecret(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{"stripe live", "sk_live_123", true},
		{"stripe test", "sk_test_123", true},
		{"aws access key", "AKIAIOSFODNN7EXAMPLE", true},
		{"github token", "github_pat_abc", true},
		{"mockport fake", "mockport_stripe_secret", false},
		{"local fake", "local_openai_key", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LooksLikeSecret(tt.value)
			if got != tt.want {
				t.Fatalf("LooksLikeSecret(%q) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestRedactSecret(t *testing.T) {
	if got := RedactSecret("mockport_stripe_secret"); got != "mockport_...cret" {
		t.Fatalf("RedactSecret fake = %q", got)
	}
	if got := RedactSecret("sk_live_123456789"); got != "[real-looking secret redacted]" {
		t.Fatalf("RedactSecret real-looking = %q", got)
	}
}
```

- [x] **Step 2: Verify RED**

Run:

```bash
go test ./internal/security -v
```

Expected: compile failure because functions do not exist.

- [x] **Step 3: Implement security primitives**

Create `internal/security/secrets.go` with prefix-based detection for `sk_live_`, `sk_test_`, `AKIA`, `ASIA`, `ghp_`, `github_pat_`, `xoxb-`, `xoxp-`, `AIza`, `whsec_`; allow fake prefixes `mockport_`, `local_`, `fake_`, `dummy_`; redact real-looking values without printing the value.

- [x] **Step 4: Verify GREEN**

Run:

```bash
go test ./internal/security -v
```

Expected: PASS.

## Task P0-T05: HTTP Server And Health Endpoint

**Status:** done

- [x] **Step 1: Write failing server test**

Create `internal/server/server_test.go`:

```go
package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthReturnsOK(t *testing.T) {
	handler := NewHandler()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode health body: %v", err)
	}
	if body["status"] != "ok" {
		t.Fatalf("status body = %q, want ok", body["status"])
	}
}
```

- [x] **Step 2: Verify RED**

Run:

```bash
go test ./internal/server -run TestHealthReturnsOK -v
```

Expected: compile failure because `NewHandler` does not exist.

- [x] **Step 3: Implement health handler**

Create `internal/server/health.go` and `internal/server/server.go`. `NewHandler` must return an `http.Handler` with `GET /health` returning `{"status":"ok"}` and `Content-Type: application/json`.

- [x] **Step 4: Verify GREEN**

Run:

```bash
go test ./internal/server -v
```

Expected: PASS.

## Task P0-T06: `mockport run`

**Status:** done

- [x] **Step 1: Write failing run command test**

Add a CLI test that `mockport run --config missing.yml` returns a non-nil error containing `load config`. This keeps the command test short and avoids long-running server lifecycle in Phase 0 CLI tests.

- [x] **Step 2: Verify RED**

Run:

```bash
go test ./internal/cli -run TestRunCommandRejectsMissingConfig -v
```

Expected: fail because `run` is not registered.

- [x] **Step 3: Implement run command**

Create `internal/cli/run.go`. It must load config via `config.LoadFile`, create `server.NewHandler()`, and call `http.ListenAndServe` on `host:port`. Keep graceful shutdown improvement for later unless tests require it.

- [x] **Step 4: Verify GREEN**

Run:

```bash
go test ./internal/cli ./internal/config ./internal/server -v
```

Expected: PASS.

## Task P0-T07: Docker, Makefile, CI

**Status:** done

- [x] **Step 1: Add build assets**

Create `Makefile`, `docker/Dockerfile`, `.dockerignore`, and `.github/workflows/ci.yml` from the documentation templates. Keep Docker entrypoint as:

```dockerfile
ENTRYPOINT ["/mockport"]
CMD ["run", "--config", "/etc/mockport/mockport.yml"]
```

- [x] **Step 2: Verify local commands**

Run:

```bash
go test ./...
go vet ./...
go build ./cmd/mockport
docker build -t mockport:local -f docker/Dockerfile .
```

Expected: all commands pass.

## Task P0-T08: Root README

**Status:** done

- [x] **Step 1: Create README**

Create `README.md` from `docs/archive/design/docs/13_readme_draft.md`, but limit supported adapters to Stripe-like payment adapter for the current state.

- [x] **Step 2: Verify docs command consistency**

Run every command listed in the README that does not require a long-running foreground server. For server commands, verify the binary accepts the command and fails with useful errors when config is missing.

## Phase 0 Exit

- [x] `go test ./...` passes.
- [x] `go vet ./...` passes.
- [x] `go build ./cmd/mockport` passes.
- [x] `docker build -t mockport:local -f docker/Dockerfile .` passes.
- [x] `/health` returns 200 when server is running.
- [x] `tasks/status.md` Phase 0 summary is updated to `done`.
