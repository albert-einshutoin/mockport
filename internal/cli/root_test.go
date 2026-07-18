package cli

import (
	"strings"
	"testing"
)

func TestRootCommandShowsHelp(t *testing.T) {
	cmd, out := newTestCommand(t, "--help")

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute help: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "Secret-free service emulation") {
		t.Fatalf("help output missing product description: %q", got)
	}
	if !strings.Contains(got, "healthcheck") {
		t.Fatalf("help output missing healthcheck command: %q", got)
	}
	if !strings.Contains(got, "agent-context") {
		t.Fatalf("help output missing agent-context command: %q", got)
	}
}

func TestRootCommandShowsUsageForFlagParseError(t *testing.T) {
	cmd, out := newTestCommand(t, "run", "--definitely-not-a-flag")

	err := cmd.Execute()
	if err == nil {
		t.Fatal("execute returned nil error for unknown flag")
	}
	if !strings.Contains(err.Error(), "unknown flag: --definitely-not-a-flag") {
		t.Fatalf("error = %q, want unknown flag", err.Error())
	}
	got := out.String()
	if !strings.Contains(got, "Usage:") || !strings.Contains(got, "mockport run") {
		t.Fatalf("flag parse error should show command usage, got:\n%s", got)
	}
}

func TestVersionCommandPrintsVersion(t *testing.T) {
	cmd, out := newTestCommand(t, "version")

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute version: %v", err)
	}

	if got := strings.TrimSpace(out.String()); got != "mockport dev" {
		t.Fatalf("version output = %q, want %q", got, "mockport dev")
	}
}
