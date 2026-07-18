package cli

import (
	"strings"
	"testing"
)

func TestAgentContextRendersSelectedAdapterEnvironment(t *testing.T) {
	cmd, out := newTestCommand(t, "agent-context", "--adapter", "stripe", "--adapter", "openai")

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute agent-context: %v", err)
	}

	got := out.String()
	for _, want := range []string{
		"Do NOT ask for real API keys",
		"STRIPE_API_URL=http://localhost:43101/stripe",
		"OPENAI_BASE_URL=http://localhost:43101/openai/v1",
		"X-Mockport-Scenario",
		"/_mockport/report",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("agent context missing %q: %s", want, got)
		}
	}
}

func TestAgentContextRejectsUnknownAdapter(t *testing.T) {
	cmd, _ := newTestCommand(t, "agent-context", "--adapter", "unknown")

	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "unsupported adapter") {
		t.Fatalf("error = %v, want unsupported adapter", err)
	}
}
