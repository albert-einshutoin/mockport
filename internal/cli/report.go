package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/albert-einshutoin/mockport/internal/config"
	"github.com/albert-einshutoin/mockport/internal/report"
	"github.com/spf13/cobra"
)

func newReportCommand() *cobra.Command {
	var reportURL string
	var format string
	cmd := &cobra.Command{
		Use:   "report",
		Short: "Print Mockport request and safety report",
		RunE: func(cmd *cobra.Command, args []string) error {
			silenceUsageForRuntimeError(cmd)
			client := &http.Client{Timeout: 5 * time.Second}
			resp, err := client.Get(reportURL)
			if err != nil {
				return fmt.Errorf("fetch report: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				return fmt.Errorf("fetch report: status %d", resp.StatusCode)
			}

			var snapshot report.Snapshot
			if err := json.NewDecoder(resp.Body).Decode(&snapshot); err != nil {
				return fmt.Errorf("decode report: %w", err)
			}
			switch format {
			case "text":
				fmt.Fprint(cmd.OutOrStdout(), report.RenderText(snapshot))
			case "json":
				encoder := json.NewEncoder(cmd.OutOrStdout())
				encoder.SetIndent("", "  ")
				if err := encoder.Encode(snapshot); err != nil {
					return fmt.Errorf("encode report: %w", err)
				}
			default:
				return fmt.Errorf("unsupported report format %q", format)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&reportURL, "url", fmt.Sprintf("http://localhost:%d/_mockport/report", config.DefaultPort), "Mockport report endpoint URL")
	cmd.Flags().StringVar(&format, "format", "text", "Report format: text or json")
	return cmd
}
