package cli

import (
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"

	"github.com/albert-einshutoin/mockport/internal/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const defaultDockerImage = "ghcr.io/albert-einshutoin/mockport:0.1.0-alpha"

const defaultCompose = `services:
  mockport:
    image: ` + defaultDockerImage + `
    command: ["run", "--config", "/etc/mockport/mockport.yml", "--host", "0.0.0.0"]
    ports:
      - "127.0.0.1:43101:43101"
    volumes:
      - ./mockport.yml:/etc/mockport/mockport.yml
`

type adapterSpec struct {
	Name       string
	BasePath   string
	Scenario   string
	FakeSecret string
	Webhook    config.WebhookConfig
	Env        map[string]string
}

func newInitCommand() *cobra.Command {
	var adapterNames []string
	var force bool
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Generate Mockport local files",
		RunE: func(cmd *cobra.Command, args []string) error {
			specs, err := specsFor(adapterNames)
			if err != nil {
				return err
			}
			files := map[string]string{
				"mockport.yml":                generatedConfig(specs),
				".env.mockport":               generatedEnv(specs),
				"docker-compose.mockport.yml": defaultCompose,
			}
			if !force {
				for path := range files {
					if _, err := os.Stat(path); err == nil {
						return fmt.Errorf("%s already exists; rerun with --force to overwrite", path)
					} else if !os.IsNotExist(err) {
						return fmt.Errorf("check %s: %w", path, err)
					}
				}
			}
			for path, content := range files {
				if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
					return fmt.Errorf("write %s: %w", path, err)
				}
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Generated mockport.yml, .env.mockport, docker-compose.mockport.yml")
			fmt.Fprintln(cmd.OutOrStdout(), "Generated .env.mockport uses fake local credentials and is safe to commit when unchanged.")
			fmt.Fprintln(cmd.OutOrStdout())
			fmt.Fprintln(cmd.OutOrStdout(), "Next steps:")
			fmt.Fprintln(cmd.OutOrStdout(), "  docker compose -f docker-compose.mockport.yml up")
			fmt.Fprintln(cmd.OutOrStdout(), "  curl http://localhost:43101/health")
			fmt.Fprintln(cmd.OutOrStdout(), "  mockport report")
			return nil
		},
	}
	cmd.Flags().StringArrayVar(&adapterNames, "adapter", nil, "Adapter to initialize; repeat for multiple adapters")
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite generated files if they already exist")
	return cmd
}

func specsFor(names []string) ([]adapterSpec, error) {
	if len(names) == 0 {
		names = []string{"stripe"}
	}
	seen := map[string]bool{}
	var specs []adapterSpec
	for _, name := range names {
		if seen[name] {
			continue
		}
		seen[name] = true
		spec, ok := adapterSpecFor(name)
		if !ok {
			return nil, fmt.Errorf("unsupported adapter %q", name)
		}
		specs = append(specs, spec)
	}
	return specs, nil
}

func adapterSpecFor(name string) (adapterSpec, bool) {
	switch name {
	case "stripe":
		return adapterSpec{
			Name:       "stripe",
			BasePath:   "/stripe",
			Scenario:   "payment_success",
			FakeSecret: "mockport_stripe_secret",
			Webhook:    config.WebhookConfig{TargetURL: "http://app:3000/webhooks/stripe", SigningSecret: "whsec_mockport"},
			Env: map[string]string{
				"STRIPE_API_URL":        "http://localhost:43101/stripe",
				"STRIPE_SECRET_KEY":     "mockport_stripe_secret",
				"STRIPE_WEBHOOK_SECRET": "whsec_mockport",
			},
		}, true
	case "openai":
		return adapterSpec{
			Name:       "openai",
			BasePath:   "/openai",
			Scenario:   "chat_success",
			FakeSecret: "mockport_openai_key",
			Env: map[string]string{
				"OPENAI_BASE_URL": "http://localhost:43101/openai/v1",
				"OPENAI_API_KEY":  "mockport_openai_key",
			},
		}, true
	case "github-oauth":
		return adapterSpec{
			Name:       "github-oauth",
			BasePath:   "/github",
			Scenario:   "oauth_success",
			FakeSecret: "mockport_github_secret",
			Env: map[string]string{
				"GITHUB_OAUTH_BASE_URL":      "http://localhost:43101/github",
				"GITHUB_OAUTH_CLIENT_ID":     "mockport_github_client",
				"GITHUB_OAUTH_CLIENT_SECRET": "mockport_github_secret",
			},
		}, true
	case "slack":
		return adapterSpec{
			Name:       "slack",
			BasePath:   "/slack",
			Scenario:   "message_success",
			FakeSecret: "mockport_slack_token",
			Env: map[string]string{
				"SLACK_API_URL":   "http://localhost:43101/slack/api",
				"SLACK_BOT_TOKEN": "mockport_slack_token",
			},
		}, true
	case "line":
		return adapterSpec{
			Name:       "line",
			BasePath:   "/line",
			Scenario:   "line_success",
			FakeSecret: "mockport_line_channel_token",
			Env: map[string]string{
				"LINE_API_BASE_URL":        "http://localhost:43101/line",
				"LINE_CHANNEL_ID":          "mockport_line_channel",
				"LINE_CHANNEL_SECRET":      "mockport_line_secret",
				"LINE_CHANNEL_TOKEN":       "mockport_line_channel_token",
				"LINE_LIFF_ID":             "mockport-line-liff",
				"LINE_MINI_DAPP_CLIENT_ID": "mockport_line_mini_dapp_client",
				"LINE_PAY_CHANNEL_ID":      "mockport_line_pay_channel",
				"LINE_PAY_CHANNEL_SECRET":  "mockport_line_pay_secret",
			},
		}, true
	default:
		return adapterSpec{}, false
	}
}

func generatedConfig(specs []adapterSpec) string {
	cfg := config.Config{
		Version:  "0.1",
		Server:   config.ServerConfig{Host: "127.0.0.1", Port: 43101},
		Mode:     "ai-safe",
		Adapters: map[string]config.AdapterConfig{},
	}
	for _, spec := range specs {
		cfg.Adapters[spec.Name] = config.AdapterConfig{
			Enabled:    true,
			BasePath:   spec.BasePath,
			Scenario:   spec.Scenario,
			FakeSecret: spec.FakeSecret,
			Webhook:    spec.Webhook,
		}
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func generatedEnv(specs []adapterSpec) string {
	var out strings.Builder
	out.WriteString("# Generated by Mockport. Safe for local, CI, and public example commits when unchanged.\n\n")
	for _, spec := range specs {
		keys := slices.Sorted(maps.Keys(spec.Env))
		for _, key := range keys {
			fmt.Fprintf(&out, "%s=%s\n", key, spec.Env[key])
		}
	}
	return out.String()
}
