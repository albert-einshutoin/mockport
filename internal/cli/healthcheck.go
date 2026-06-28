package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/albert-einshutoin/mockport/internal/config"
	"github.com/spf13/cobra"
)

const defaultHealthcheckHost = "127.0.0.1"

func newHealthcheckCommand() *cobra.Command {
	var configPath string
	var healthURL string
	var timeout time.Duration
	cmd := &cobra.Command{
		Use:   "healthcheck",
		Short: "Check Mockport health endpoint",
		RunE: func(cmd *cobra.Command, args []string) error {
			silenceUsageForRuntimeError(cmd)
			resolvedURL, err := resolveHealthcheckURL(configPath, healthURL)
			if err != nil {
				return err
			}
			client := &http.Client{Timeout: timeout}
			resp, err := client.Get(resolvedURL)
			if err != nil {
				return fmt.Errorf("healthcheck request: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				return fmt.Errorf("healthcheck status: %s", resp.Status)
			}
			var payload map[string]string
			if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
				return fmt.Errorf("healthcheck decode: %w", err)
			}
			status, ok := payload["status"]
			if !ok {
				return fmt.Errorf("healthcheck response missing status field")
			}
			if strings.ToLower(status) != "ok" {
				return fmt.Errorf("healthcheck status value %q (expected ok)", status)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "mockport healthcheck passed: %s\n", resolvedURL)
			return nil
		},
	}
	cmd.Flags().StringVar(&healthURL, "url", "", "Health endpoint URL; optional when config is available")
	cmd.Flags().StringVar(&configPath, "config", filepath.Clean("mockport.yml"), "Path to mockport.yml")
	cmd.Flags().DurationVar(&timeout, "timeout", 5*time.Second, "HTTP client timeout (e.g. 5s, 1m)")
	return cmd
}

func resolveHealthcheckURL(configPath, configured string) (string, error) {
	if configured != "" {
		return configured, nil
	}
	cfg, err := loadHealthcheckConfig(configPath)
	if err != nil {
		return "", err
	}
	if cfg.Server.Host == "0.0.0.0" {
		cfg.Server.Host = "127.0.0.1"
	}
	return fmt.Sprintf("http://%s:%d/health", cfg.Server.Host, cfg.Server.Port), nil
}

func loadHealthcheckConfig(configPath string) (config.Config, error) {
	if _, err := os.Stat(configPath); err != nil {
		if os.IsNotExist(err) {
			return config.Config{Server: config.ServerConfig{
				Host: defaultHealthcheckHost,
				Port: config.DefaultPort,
			}}, nil
		}
		return config.Config{}, fmt.Errorf("load config %s: %w", configPath, err)
	}
	cfg, err := config.LoadFile(configPath)
	if err != nil {
		return config.Config{}, err
	}
	return cfg, nil
}
