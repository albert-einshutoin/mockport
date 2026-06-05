package config

import (
	"fmt"
	"sort"
	"strings"

	"github.com/albert-einshutoin/mockport/internal/security"
)

func Validate(cfg *Config) error {
	if cfg.Server.Port < 1 || cfg.Server.Port > 65535 {
		return fmt.Errorf("invalid server port %d", cfg.Server.Port)
	}

	switch cfg.Mode {
	case "ai-safe", "local", "strict":
	default:
		return fmt.Errorf("unsupported mode %q", cfg.Mode)
	}

	var warnings []SafetyWarning
	if warning, ok := serverHostWarning(cfg.Server.Host); ok {
		warnings = append(warnings, warning)
	}

	seenBasePaths := map[string]string{}
	for _, name := range sortedAdapterNames(cfg.Adapters) {
		adapter := cfg.Adapters[name]
		checks := map[string]string{
			name + ".fake_secret":            adapter.FakeSecret,
			name + ".api_url":                adapter.APIURL,
			name + ".webhook.target_url":     adapter.Webhook.TargetURL,
			name + ".webhook.signing_secret": adapter.Webhook.SigningSecret,
		}
		for field, value := range checks {
			value = security.NormalizePublicSafetyValue(value)
			if value == "" {
				continue
			}
			if security.LooksLikeSecret(value) {
				warnings = append(warnings, SafetyWarning{Field: field, Category: "real_looking_secret", Message: "real-looking secret detected"})
			}
			if security.LooksLikeExternalServiceURL(value) {
				warnings = append(warnings, SafetyWarning{Field: field, Category: "external_url", Message: "external live service URL detected"})
			}
		}
		if !adapter.Enabled {
			continue
		}
		if err := validateBasePath(name, adapter.BasePath, seenBasePaths); err != nil {
			return err
		}
	}

	if cfg.Mode == "strict" && len(warnings) > 0 {
		fields := make([]string, 0, len(warnings))
		for _, warning := range warnings {
			fields = append(fields, warning.Field)
		}
		sort.Strings(fields)
		return fmt.Errorf("strict mode rejected unsafe config fields: %s", strings.Join(fields, ", "))
	}
	cfg.SafetyWarnings = warnings
	return nil
}

func validateBasePath(adapterName, basePath string, seen map[string]string) error {
	if basePath == "" {
		return fmt.Errorf("adapter %s missing base_path", adapterName)
	}
	if strings.TrimSpace(basePath) != basePath {
		return fmt.Errorf("adapter %s invalid base_path %q: must not contain surrounding whitespace", adapterName, basePath)
	}
	if !strings.HasPrefix(basePath, "/") {
		return fmt.Errorf("adapter %s invalid base_path %q: must start with /", adapterName, basePath)
	}
	if basePath == "/" {
		return fmt.Errorf("adapter %s invalid base_path %q: root path is reserved", adapterName, basePath)
	}
	if strings.HasSuffix(basePath, "/") {
		return fmt.Errorf("adapter %s invalid base_path %q: trailing slash is not allowed", adapterName, basePath)
	}
	if strings.ContainsAny(basePath, "?#{}*") {
		return fmt.Errorf("adapter %s invalid base_path %q: must be a literal path prefix", adapterName, basePath)
	}
	if owner, exists := seen[basePath]; exists {
		return fmt.Errorf("adapter %s base_path %q duplicates adapter %s", adapterName, basePath, owner)
	}
	seen[basePath] = adapterName
	return nil
}

func serverHostWarning(host string) (SafetyWarning, bool) {
	host = strings.TrimSpace(host)
	if host == "" || security.IsLoopbackHost(host) {
		return SafetyWarning{}, false
	}
	return SafetyWarning{
		Field:    "server.host",
		Category: "public_bind",
		Message:  "server host may expose Mockport outside loopback",
	}, true
}

func sortedAdapterNames(adapters map[string]AdapterConfig) []string {
	names := make([]string, 0, len(adapters))
	for name := range adapters {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
