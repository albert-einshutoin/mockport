package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var Version = "dev"

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print Mockport version",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(cmd.OutOrStdout(), "mockport %s\n", Version)
			return nil
		},
	}
}
