package config

import (
	"fmt"
	"slices"
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

	// loader.go の2段デコードで事前に追加された warning（unknown_config_key 等、
	// Validate 自身が生成しないカテゴリ）だけを退避しておく。Validate 管轄カテゴリは
	// 再生成して上書きするため、2回呼んでも結果が安定する（冪等）。
	var warnings []SafetyWarning
	for _, w := range cfg.SafetyWarnings {
		if !validateOwnedCategories[w.Category] {
			warnings = append(warnings, w)
		}
	}
	if warning, ok := serverHostWarning(cfg.Server.Host); ok {
		warnings = append(warnings, warning)
	}

	seenBasePaths := map[string]string{}
	for _, name := range sortedAdapterNames(cfg.Adapters) {
		adapter := cfg.Adapters[name]
		if adapter.AuthRequired && name != "stripe" && name != "openai" {
			return fmt.Errorf("adapter %s: auth_required is only supported for stripe and openai", name)
		}
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

	// scenarios: ブロックが設定されていても未実装のため警告のみ（エラーにしない）
	if len(cfg.Scenarios) > 0 {
		warnings = append(warnings, SafetyWarning{
			Field:    "scenarios",
			Category: "unsupported_config",
			Message:  "the scenarios block is not implemented yet and is ignored; see docs/scenario-policy.md",
		})
	}

	// strict モードは秘密情報・外部URL・公開バインドなどのセキュリティ系警告のみ起動拒否する。
	// unsupported_config / unknown_config_key はユーザビリティ警告であり strict 対象外。
	if cfg.Mode == "strict" {
		var strictWarnings []SafetyWarning
		for _, w := range warnings {
			if w.Category != "unsupported_config" && w.Category != "unknown_config_key" {
				strictWarnings = append(strictWarnings, w)
			}
		}
		if len(strictWarnings) > 0 {
			fields := make([]string, 0, len(strictWarnings))
			for _, w := range strictWarnings {
				fields = append(fields, w.Field)
			}
			slices.Sort(fields)
			return fmt.Errorf("strict mode rejected unsafe config fields: %s", strings.Join(fields, ", "))
		}
	}
	// 退避した loader 由来 warning + 今回生成した warning で上書きする。
	cfg.SafetyWarnings = warnings
	return nil
}

// validateOwnedCategories は Validate 自身が生成する SafetyWarning カテゴリの集合。
// Validate を複数回呼んだときの重複を防ぐため、冒頭でこれらを除去してから付け直す。
var validateOwnedCategories = map[string]bool{
	"public_bind":         true,
	"real_looking_secret": true,
	"external_url":        true,
	"unsupported_config":  true,
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
	slices.Sort(names)
	return names
}
