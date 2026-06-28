package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitGeneratesStripeFiles(t *testing.T) {
	dir := chdirTemp(t)

	cmd, out := newTestCommand(t, "init", "--adapter", "stripe")
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute init: %v", err)
	}

	for _, name := range []string{"mockport.yml", ".env.mockport", "docker-compose.mockport.yml"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Fatalf("expected generated file %s: %v", name, err)
		}
	}
	configData, err := os.ReadFile(filepath.Join(dir, "mockport.yml"))
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if !strings.Contains(string(configData), "host: 127.0.0.1") {
		t.Fatalf("config missing loopback host: %s", string(configData))
	}
	composeData, err := os.ReadFile(filepath.Join(dir, "docker-compose.mockport.yml"))
	if err != nil {
		t.Fatalf("read compose: %v", err)
	}
	if !strings.Contains(string(composeData), "127.0.0.1:43101:43101") {
		t.Fatalf("compose missing loopback port binding: %s", string(composeData))
	}
	if !strings.Contains(string(composeData), "ghcr.io/albert-einshutoin/mockport:0.1.0-alpha") {
		t.Fatalf("compose missing pinned preview image: %s", string(composeData))
	}
	if strings.Contains(string(composeData), "ghcr.io/albert-einshutoin/mockport:latest") {
		t.Fatalf("compose should not default to mutable latest image: %s", string(composeData))
	}
	envData, err := os.ReadFile(filepath.Join(dir, ".env.mockport"))
	if err != nil {
		t.Fatalf("read env: %v", err)
	}
	env := string(envData)
	if !strings.Contains(env, "STRIPE_API_URL=http://localhost:43101/stripe") {
		t.Fatalf("env missing Stripe URL: %s", env)
	}
	if !strings.Contains(env, "STRIPE_SECRET_KEY=mockport_stripe_secret") {
		t.Fatalf("env missing fake Stripe key: %s", env)
	}
	got := out.String()
	for _, want := range []string{
		"Generated .env.mockport uses fake local credentials and is safe to commit when unchanged.",
		"docker compose -f docker-compose.mockport.yml up",
		"curl http://localhost:43101/health",
		"mockport report",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("init output missing %q:\n%s", want, got)
		}
	}
}

func TestInitGeneratesMultipleAdapters(t *testing.T) {
	dir := chdirTemp(t)

	cmd, _ := newTestCommand(t, "init", "--adapter", "stripe", "--adapter", "openai", "--adapter", "line")
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute init: %v", err)
	}

	configData, err := os.ReadFile(filepath.Join(dir, "mockport.yml"))
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	configText := string(configData)
	for _, want := range []string{"stripe:", "openai:", "line:", "base_path: /line"} {
		if !strings.Contains(configText, want) {
			t.Fatalf("config missing %q:\n%s", want, configText)
		}
	}

	envData, err := os.ReadFile(filepath.Join(dir, ".env.mockport"))
	if err != nil {
		t.Fatalf("read env: %v", err)
	}
	env := string(envData)
	for _, want := range []string{"STRIPE_API_URL=http://localhost:43101/stripe", "OPENAI_BASE_URL=http://localhost:43101/openai/v1", "LINE_API_BASE_URL=http://localhost:43101/line"} {
		if !strings.Contains(env, want) {
			t.Fatalf("env missing %q:\n%s", want, env)
		}
	}
}

func TestInitExistingFileErrorPrefersMockportYml(t *testing.T) {
	dir := chdirTemp(t)
	for _, name := range []string{".env.mockport", "docker-compose.mockport.yml", "mockport.yml"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("keep\n"), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	cmd, _ := newTestCommand(t, "init", "--adapter", "stripe")
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when all init files exist")
	}
	if !strings.Contains(err.Error(), "mockport.yml already exists") {
		t.Fatalf("error = %q, want mockport.yml reported first", err)
	}
}

func TestInitDoesNotOverwriteExistingFiles(t *testing.T) {
	dir := chdirTemp(t)
	existingPath := filepath.Join(dir, "mockport.yml")
	existing := "custom: keep-me\n"
	if err := os.WriteFile(existingPath, []byte(existing), 0o644); err != nil {
		t.Fatalf("write existing config: %v", err)
	}

	cmd, _ := newTestCommand(t, "init", "--adapter", "stripe")
	err := cmd.Execute()
	if err == nil {
		t.Fatal("init returned nil error for existing file")
	}

	data, err := os.ReadFile(existingPath)
	if err != nil {
		t.Fatalf("read existing config: %v", err)
	}
	if string(data) != existing {
		t.Fatalf("existing file was overwritten: %q", string(data))
	}
}

func TestInitForceOverwritesExistingFiles(t *testing.T) {
	dir := chdirTemp(t)
	existingPath := filepath.Join(dir, "mockport.yml")
	if err := os.WriteFile(existingPath, []byte("custom: replace-me\n"), 0o644); err != nil {
		t.Fatalf("write existing config: %v", err)
	}

	cmd, _ := newTestCommand(t, "init", "--adapter", "stripe", "--force")
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute init --force: %v", err)
	}

	data, err := os.ReadFile(existingPath)
	if err != nil {
		t.Fatalf("read overwritten config: %v", err)
	}
	if !strings.Contains(string(data), "mode: ai-safe") {
		t.Fatalf("mockport.yml was not replaced with generated config: %s", string(data))
	}
}

func chdirTemp(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	oldwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get cwd: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(oldwd); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	})
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir temp: %v", err)
	}
	return dir
}
