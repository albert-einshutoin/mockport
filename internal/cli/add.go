package cli

import (
	"fmt"
	"os"

	"github.com/albert-einshutoin/mockport/internal/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func newAddCommand() *cobra.Command {
	var configPath string
	cmd := &cobra.Command{
		Use:   "add [adapter...]",
		Short: "Add adapter config to mockport.yml",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			silenceUsageForRuntimeError(cmd)
			specs, err := specsFor(args)
			if err != nil {
				return err
			}
			cfg, err := config.LoadFile(configPath)
			if err != nil {
				return err
			}
			for _, spec := range specs {
				cfg.Adapters[spec.Name] = spec.adapterConfig()
			}
			if err := config.Validate(&cfg); err != nil {
				return err
			}
			out, err := yaml.Marshal(cfg)
			if err != nil {
				return fmt.Errorf("encode config: %w", err)
			}
			if err := os.WriteFile(configPath, out, 0o644); err != nil {
				return fmt.Errorf("write %s: %w", configPath, err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Updated %s\n", configPath)
			return nil
		},
	}
	cmd.Flags().StringVar(&configPath, "config", "mockport.yml", "Path to mockport.yml")
	return cmd
}
