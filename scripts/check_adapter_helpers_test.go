package scripts_test

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/albert-einshutoin/mockport/internal/compat"
)

func TestCheckAdapterHelpersHelpPointsToPolicy(t *testing.T) {
	repoRoot, err := compat.FindRepoRoot(".")
	if err != nil {
		t.Fatalf("FindRepoRoot: %v", err)
	}

	for _, arg := range []string{"--help", "-h"} {
		t.Run(arg, func(t *testing.T) {
			cmd := exec.Command("bash", filepath.Join(repoRoot, "scripts", "check-adapter-helpers.sh"), arg)
			cmd.Dir = repoRoot

			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("help command failed: %v\noutput: %s", err, output)
			}
			if !strings.Contains(string(output), "docs/adapter-helper-policy.md") {
				t.Fatalf("help output does not point to the helper policy:\n%s", output)
			}
			if !strings.Contains(string(output), "Usage:") {
				t.Fatalf("help output does not include usage:\n%s", output)
			}
			if strings.Contains(string(output), "duplicate helper:") {
				t.Fatalf("help should not run the duplicate scan:\n%s", output)
			}
		})
	}
}
