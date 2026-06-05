package cli

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/albert-einshutoin/mockport/internal/adapter"
	"github.com/spf13/cobra"
)

func newHelpCommand(root *cobra.Command) *cobra.Command {
	return &cobra.Command{
		Use:   "help [command|service]",
		Short: "Help about any command or built-in service",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				root.SetOut(cmd.OutOrStdout())
				root.SetErr(cmd.ErrOrStderr())
				return root.Help()
			}
			if target := findCommand(root, args); target != nil {
				target.SetOut(cmd.OutOrStdout())
				target.SetErr(cmd.ErrOrStderr())
				return target.Help()
			}
			if len(args) == 1 {
				wrote, err := writeServiceHelp(cmd.OutOrStdout(), args[0])
				if wrote || err != nil {
					return err
				}
				return fmt.Errorf("unsupported service %q; supported services: %s", args[0], strings.Join(supportedServiceNames(), ", "))
			}
			return fmt.Errorf("unknown help topic %q", strings.Join(args, " "))
		},
	}
}

func findCommand(root *cobra.Command, args []string) *cobra.Command {
	current := root
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			return nil
		}
		var next *cobra.Command
		for _, child := range current.Commands() {
			if commandMatches(child, arg) {
				next = child
				break
			}
		}
		if next == nil {
			return nil
		}
		current = next
	}
	return current
}

func commandMatches(cmd *cobra.Command, name string) bool {
	if cmd.Name() == name {
		return true
	}
	for _, alias := range cmd.Aliases {
		if alias == name {
			return true
		}
	}
	return false
}

func writeServiceHelp(w io.Writer, name string) (bool, error) {
	spec, ok := adapterSpecFor(name)
	if !ok {
		return false, nil
	}
	adapterImpl, ok := builtinAdapterFor(name)
	if !ok {
		return true, fmt.Errorf("service %q has config defaults but no built-in adapter", name)
	}
	meta := adapterImpl.Metadata()
	if err := adapter.ValidateMetadata(meta); err != nil {
		return true, fmt.Errorf("service %s metadata: %w", name, err)
	}

	fmt.Fprintf(w, "Mockport service: %s\n\n", spec.Name)
	fmt.Fprintln(w, "Default config:")
	fmt.Fprintf(w, "  adapter: %s\n", spec.Name)
	fmt.Fprintf(w, "  base_path: %s\n", spec.BasePath)
	fmt.Fprintf(w, "  default_scenario: %s\n", spec.Scenario)
	fmt.Fprintf(w, "  fake_secret: %s\n", spec.FakeSecret)
	if spec.Webhook.TargetURL != "" || spec.Webhook.SigningSecret != "" {
		fmt.Fprintln(w, "  webhook:")
		if spec.Webhook.TargetURL != "" {
			fmt.Fprintf(w, "    target_url: %s\n", spec.Webhook.TargetURL)
		}
		if spec.Webhook.SigningSecret != "" {
			fmt.Fprintf(w, "    signing_secret: %s\n", spec.Webhook.SigningSecret)
		}
	}

	fmt.Fprintln(w)
	fmt.Fprintln(w, "Implementation:")
	fmt.Fprintf(w, "  maturity: %s\n", meta.Maturity)
	fmt.Fprintf(w, "  provider_version: %s\n", meta.ProviderVersion)
	writeInlineList(w, "  compatibility_levels", levelNames(meta.Levels))
	writeInlineList(w, "  capabilities", meta.Capabilities)
	writeInlineList(w, "  stateful_resources", meta.StatefulResources)
	fmt.Fprintf(w, "  idempotency: %t\n", meta.Idempotency)
	fmt.Fprintf(w, "  reset: %t\n", meta.Reset)
	writeInlineList(w, "  sdk_evidence", sdkVersionNames(meta.SDKVersions))
	writeInlineList(w, "  client_evidence", meta.ClientEvidence)

	fmt.Fprintln(w)
	fmt.Fprintln(w, "Environment:")
	for _, key := range sortedMapKeys(spec.Env) {
		fmt.Fprintf(w, "  %s=%s\n", key, spec.Env[key])
	}

	fmt.Fprintln(w)
	fmt.Fprintln(w, "Scenarios:")
	for _, scenario := range meta.Scenarios {
		status := "unsupported"
		if scenario.Supported {
			status = "supported"
		}
		fmt.Fprintf(w, "  - %s (%s)\n", scenario.Name, status)
	}

	fmt.Fprintln(w)
	fmt.Fprintln(w, "Endpoints:")
	for _, endpoint := range meta.Endpoints {
		fmt.Fprintf(w, "  - %s %s", endpoint.Method, endpoint.Path)
		if len(endpoint.SupportedScenarios) > 0 {
			fmt.Fprintf(w, " [%s]", strings.Join(endpoint.SupportedScenarios, ", "))
		}
		if endpoint.Notes != "" {
			fmt.Fprintf(w, " - %s", endpoint.Notes)
		}
		fmt.Fprintln(w)
	}

	fmt.Fprintln(w)
	fmt.Fprintln(w, "Related:")
	fmt.Fprintf(w, "  mockport add %s\n", spec.Name)
	fmt.Fprintf(w, "  docs/adapters/%s.md\n", spec.Name)
	return true, nil
}

func writeInlineList(w io.Writer, label string, values []string) {
	if len(values) == 0 {
		fmt.Fprintf(w, "%s: none\n", label)
		return
	}
	fmt.Fprintf(w, "%s: %s\n", label, strings.Join(values, ", "))
}

func levelNames(levels []adapter.Level) []string {
	names := make([]string, 0, len(levels))
	for _, level := range levels {
		names = append(names, string(level))
	}
	return names
}

func sdkVersionNames(versions []adapter.SDKVersion) []string {
	names := make([]string, 0, len(versions))
	for _, version := range versions {
		names = append(names, version.Name+"@"+version.Version)
	}
	return names
}

func sortedMapKeys(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func supportedServiceNames() []string {
	names := make([]string, 0, len(builtinAdapters()))
	for _, adapterImpl := range builtinAdapters() {
		names = append(names, adapterImpl.Name())
	}
	sort.Strings(names)
	return names
}
