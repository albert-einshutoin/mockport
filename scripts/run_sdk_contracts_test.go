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

	// Pin TMPDIR so we can assert the script does not start harness setup
	// (mktemp -d) before rejecting unsupported providers.
	tmpBase := t.TempDir()

	cmd := exec.Command("bash", filepath.Join(repoRoot, "scripts", "run-sdk-contracts.sh"), "line")
	cmd.Dir = repoRoot
	cmd.Env = append(os.Environ(), "TMPDIR="+tmpBase)

	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected non-zero exit for unsupported provider line, got success\noutput: %s", output)
	}
	if exitErr, ok := err.(*exec.ExitError); !ok || exitErr.ExitCode() != 1 {
		t.Fatalf("expected exit code 1, got %v\noutput: %s", err, output)
	}

	combined := string(output)
	if !strings.Contains(combined, "unsupported provider: line") {
		t.Fatalf("expected clear unsupported provider message, got:\n%s", combined)
	}

	entries, err := os.ReadDir(tmpBase)
	if err != nil {
		t.Fatalf("ReadDir TMPDIR: %v", err)
	}
	for _, entry := range entries {
		t.Errorf("unexpected temp entry after rejected provider (harness setup ran too early): %s", entry.Name())
	}
}
