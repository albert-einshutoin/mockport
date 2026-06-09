package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/cobra"
)

const defaultHealthcheckURL = "http://localhost:43101/health"

func newHealthcheckCommand() *cobra.Command {
	var healthURL string
	cmd := &cobra.Command{
		Use:   "healthcheck",
		Short: "Run Mockport healthcheck endpoint",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := &http.Client{
				Timeout: 5 * time.Second,
			}
			resp, err := client.Get(healthURL)
			if err != nil {
				return fmt.Errorf("healthcheck request: %w", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("healthcheck status: %d", resp.StatusCode)
			}
			var payload map[string]string
			if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
				return fmt.Errorf("healthcheck decode: %w", err)
			}
			if payload["status"] != "ok" {
				return fmt.Errorf("healthcheck response: %v", payload)
			}
			fmt.Fprintln(cmd.OutOrStdout(), "mockport healthcheck: ok")
			return nil
		},
	}
	cmd.Flags().StringVar(&healthURL, "url", defaultHealthcheckURL, "Healthcheck endpoint URL")
	return cmd
}
