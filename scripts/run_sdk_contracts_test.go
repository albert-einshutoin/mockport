package scripts_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/albert-einshutoin/mockport/internal/compat"
)

func TestRunSDKContractsRejectsUnsupportedProvider(t *testing.T) {
	repoRoot, err := compat.FindRepoRoot(".")
	if err != nil {
		t.Fatalf("FindRepoRoot: %v", err)
	}

	// Reject before mktemp so unsupported providers never start or leave behind
	// contract-harness resources.
	tmpBase := t.TempDir()

	const unsupportedProvider = "sendgrid"
	cmd := exec.Command("bash", filepath.Join(repoRoot, "scripts", "run-sdk-contracts.sh"), unsupportedProvider)
	cmd.Dir = repoRoot
	cmd.Env = append(os.Environ(), "TMPDIR="+tmpBase)

	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected non-zero exit for unsupported provider %s, got success\noutput: %s", unsupportedProvider, output)
	}
	if exitErr, ok := err.(*exec.ExitError); !ok || exitErr.ExitCode() != 1 {
		t.Fatalf("expected exit code 1, got %v\noutput: %s", err, output)
	}

	if !strings.Contains(string(output), "unsupported provider: "+unsupportedProvider) {
		t.Fatalf("expected clear unsupported provider message, got:\n%s", output)
	}

	entries, err := os.ReadDir(tmpBase)
	if err != nil {
		t.Fatalf("ReadDir TMPDIR: %v", err)
	}
	for _, entry := range entries {
		t.Errorf("unexpected temp entry after rejected provider: %s", entry.Name())
	}
}
