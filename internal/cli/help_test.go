package cli

import (
	"strings"
	"testing"
)

func TestHelpServiceShowsAdapterImplementationAndSpec(t *testing.T) {
	cmd, out := newTestCommand(t, "help", "stripe")

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute service help: %v", err)
	}

	got := out.String()
	for _, want := range []string{
		"Mockport service: stripe",
		"base_path: /stripe",
		"default_scenario: payment_success",
		"maturity: workflow-compatible",
		"capabilities: checkout_sessions",
		"stateful_resources: checkout_session",
		"STRIPE_API_URL=http://localhost:43101/stripe",
		"payment_failed (supported)",
		"POST /stripe/v1/checkout/sessions",
		"mockport add stripe",
		"docs/adapters/stripe.md",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("service help missing %q:\n%s", want, got)
		}
	}
}

func TestHelpServiceSupportsEveryBuiltInService(t *testing.T) {
	for _, service := range supportedServiceNames() {
		t.Run(service, func(t *testing.T) {
			spec, ok := adapterSpecFor(service)
			if !ok {
				t.Fatalf("adapterSpecFor(%q) = false", service)
			}
			adapterImpl, ok := builtinAdapterFor(service)
			if !ok {
				t.Fatalf("builtinAdapterFor(%q) = false", service)
			}
			meta := adapterImpl.Metadata()

			cmd, out := newTestCommand(t, "help", service)

			if err := cmd.Execute(); err != nil {
				t.Fatalf("execute service help: %v", err)
			}
			got := out.String()

			assertHelpContains(t, service, got, "Mockport service: "+service)
			if spec.BasePath != "" {
				assertHelpContains(t, service, got, "base_path: "+spec.BasePath)
			}
			if len(meta.Scenarios) == 0 {
				t.Fatalf("service %s metadata has no scenarios", service)
			}
			for _, scenario := range meta.Scenarios {
				// Match Scenarios list items, not endpoint SupportedScenarios mentions.
				assertHelpContains(t, service, got, "  - "+scenario.Name+" (")
			}
		})
	}
}

func assertHelpContains(t *testing.T, service, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Fatalf("service help for %s missing %q:\n%s", service, want, got)
	}
}

func TestHelpCommandStillShowsCommandHelp(t *testing.T) {
	cmd, out := newTestCommand(t, "help", "add")

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute command help: %v", err)
	}

	got := out.String()
	for _, want := range []string{
		"Add adapter config to mockport.yml",
		"add [adapter...]",
		"--config",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("command help missing %q:\n%s", want, got)
		}
	}
}

func TestHelpServiceRejectsUnsupportedService(t *testing.T) {
	cmd, _ := newTestCommand(t, "help", "unknown")

	err := cmd.Execute()
	if err == nil {
		t.Fatal("help returned nil error for unsupported service")
	}
	errText := err.Error()
	if !strings.Contains(errText, `unsupported service "unknown"`) {
		t.Fatalf("error = %q, want unsupported service", errText)
	}
	if !strings.Contains(errText, "github-oauth, line, openai, slack, stripe, zoho-oauth") {
		t.Fatalf("error missing supported services: %q", errText)
	}
}
