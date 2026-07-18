package cli

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/spf13/cobra"
)

func newAgentContextCommand() *cobra.Command {
	var adapterNames []string
	cmd := &cobra.Command{
		Use:   "agent-context",
		Short: "Print safe Mockport context for AI coding agents",
		RunE: func(cmd *cobra.Command, _ []string) error {
			specs, err := specsFor(adapterNames)
			if err != nil {
				return err
			}
			fmt.Fprint(cmd.OutOrStdout(), renderAgentContext(specs))
			return nil
		},
	}
	cmd.Flags().StringArrayVar(&adapterNames, "adapter", nil, "Adapter to include; repeat for multiple adapters")
	return cmd
}

func renderAgentContext(specs []adapterSpec) string {
	var out strings.Builder
	out.WriteString("## Mockport (local API emulator)\n\n")
	out.WriteString("This project uses Mockport to emulate external APIs locally. Do NOT ask for real API keys.\n\n")
	out.WriteString("- Start: `docker compose -f docker-compose.mockport.yml up`\n")
	out.WriteString("- Error-path testing: send `X-Mockport-Scenario: <scenario>`\n")
	out.WriteString("- Verify coverage: `curl http://localhost:43101/_mockport/report`\n")
	out.WriteString("- Keep the generated fake values below unchanged; they are intentionally safe to commit.\n")
	out.WriteString("- Never replace fake values with real credentials or provider URLs.\n")

	for _, spec := range specs {
		fmt.Fprintf(&out, "\n### %s\n\n", spec.Name)
		// Sort environment keys so generated context is stable in reviews and can
		// be safely regenerated without map iteration noise.
		for _, key := range slices.Sorted(maps.Keys(spec.Env)) {
			fmt.Fprintf(&out, "- `%s=%s`\n", key, spec.Env[key])
		}
	}
	return out.String()
}
