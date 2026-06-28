package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAddRequiresAtLeastOneAdapter(t *testing.T) {
	cmd, _ := newTestCommand(t, "add")
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when add is called without adapters")
	}
	if !strings.Contains(err.Error(), "requires at least 1 arg(s)") {
		t.Fatalf("error = %q, want cobra minimum args message", err)
	}
}

func TestAddMissingConfigUsesLoadFileError(t *testing.T) {
	cmd, _ := newTestCommand(t, "add", "stripe", "--config", "missing-mockport.yml")
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing config")
	}
	if !strings.Contains(err.Error(), "load config missing-mockport.yml:") {
		t.Fatalf("error = %q, want load config prefix from LoadFile", err)
	}
}

func TestAddAdaptersUpdatesConfig(t *testing.T) {
	dir := chdirTemp(t)
	configPath := filepath.Join(dir, "mockport.yml")
	configContent, err := generatedConfig([]adapterSpec{mustSpec(t, "stripe")})
	if err != nil {
		t.Fatalf("generate config: %v", err)
	}
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cmd, _ := newTestCommand(t, "add", "openai", "github-oauth", "slack", "line", "--config", configPath)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute add: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	configText := string(data)
	for _, want := range []string{"stripe:", "openai:", "github-oauth:", "slack:", "line:"} {
		if !strings.Contains(configText, want) {
			t.Fatalf("config missing %q:\n%s", want, configText)
		}
	}
}

func mustSpec(t *testing.T, name string) adapterSpec {
	t.Helper()
	spec, ok := adapterSpecFor(name)
	if !ok {
		t.Fatalf("missing spec for %s", name)
	}
	return spec
}
