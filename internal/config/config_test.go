package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadValidConfigAppliesDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "mockport.yml")
	err := os.WriteFile(path, []byte("version: \"0.1\"\nserver:\n  port: 43101\nmode: ai-safe\nadapters:\n  stripe:\n    enabled: true\n    base_path: /stripe\n    scenario: payment_success\n    fake_secret: mockport_stripe_secret\n"), 0o644)
	if err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := LoadFile(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.Server.Host != "127.0.0.1" {
		t.Fatalf("host = %q, want default 127.0.0.1", cfg.Server.Host)
	}
	if cfg.Server.Port != 43101 {
		t.Fatalf("port = %d, want 43101", cfg.Server.Port)
	}
}

func TestLoadRejectsInvalidPort(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "mockport.yml")
	err := os.WriteFile(path, []byte("version: \"0.1\"\nserver:\n  port: 0\nmode: ai-safe\n"), 0o644)
	if err != nil {
		t.Fatalf("write config: %v", err)
	}

	if _, err := LoadFile(path); err == nil {
		t.Fatal("LoadFile returned nil error for invalid port")
	}
}

func TestValidateRejectsInvalidBasePaths(t *testing.T) {
	tests := []struct {
		name     string
		adapters map[string]AdapterConfig
		want     string
	}{
		{
			name: "missing leading slash",
			adapters: map[string]AdapterConfig{
				"stripe": {Enabled: true, BasePath: "stripe"},
			},
			want: "must start with /",
		},
		{
			name: "root path",
			adapters: map[string]AdapterConfig{
				"stripe": {Enabled: true, BasePath: "/"},
			},
			want: "root path is reserved",
		},
		{
			name: "trailing slash",
			adapters: map[string]AdapterConfig{
				"stripe": {Enabled: true, BasePath: "/stripe/"},
			},
			want: "trailing slash is not allowed",
		},
		{
			name: "serve mux wildcard",
			adapters: map[string]AdapterConfig{
				"stripe": {Enabled: true, BasePath: "/stripe/{id}"},
			},
			want: "must be a literal path prefix",
		},
		{
			name: "duplicate enabled path",
			adapters: map[string]AdapterConfig{
				"openai": {Enabled: true, BasePath: "/api"},
				"stripe": {Enabled: true, BasePath: "/api"},
			},
			want: "duplicates adapter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{
				Server:   ServerConfig{Port: 43101},
				Mode:     "ai-safe",
				Adapters: tt.adapters,
			}
			err := Validate(&cfg)
			if err == nil {
				t.Fatal("Validate returned nil error")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %q, want %q", err.Error(), tt.want)
			}
		})
	}
}

func TestValidateAllowsDuplicateDisabledBasePath(t *testing.T) {
	cfg := Config{
		Server: ServerConfig{Port: 43101},
		Mode:   "ai-safe",
		Adapters: map[string]AdapterConfig{
			"openai": {Enabled: false, BasePath: "/api"},
			"stripe": {Enabled: true, BasePath: "/api"},
		},
	}

	if err := Validate(&cfg); err != nil {
		t.Fatalf("Validate disabled duplicate: %v", err)
	}
}

func TestAISafeScansDisabledAdaptersForUnsafeValues(t *testing.T) {
	cfg := Config{
		Server: ServerConfig{Port: 43101},
		Mode:   "ai-safe",
		Adapters: map[string]AdapterConfig{
			"stripe": {Enabled: false, FakeSecret: "sk_live_123"},
		},
	}

	if err := Validate(&cfg); err != nil {
		t.Fatalf("validate ai-safe config: %v", err)
	}
	if len(cfg.SafetyWarnings) != 1 {
		t.Fatalf("warnings = %d, want 1", len(cfg.SafetyWarnings))
	}
	if cfg.SafetyWarnings[0].Field != "stripe.fake_secret" {
		t.Fatalf("field = %q, want stripe.fake_secret", cfg.SafetyWarnings[0].Field)
	}
}

func TestStrictRejectsUnsafeDisabledAdapters(t *testing.T) {
	cfg := Config{
		Server: ServerConfig{Port: 43101},
		Mode:   "strict",
		Adapters: map[string]AdapterConfig{
			"stripe": {Enabled: false, FakeSecret: "sk_live_123"},
		},
	}

	err := Validate(&cfg)
	if err == nil {
		t.Fatal("strict config with unsafe disabled adapter returned nil error")
	}
	if !strings.Contains(err.Error(), "stripe.fake_secret") {
		t.Fatalf("error = %q, want stripe.fake_secret", err.Error())
	}
}

func TestAISafeRecordsRealLookingSecretWarning(t *testing.T) {
	cfg := Config{
		Server: ServerConfig{Port: 43101},
		Mode:   "ai-safe",
		Adapters: map[string]AdapterConfig{
			"stripe": {Enabled: true, BasePath: "/stripe", FakeSecret: "sk_live_123"},
		},
	}

	if err := Validate(&cfg); err != nil {
		t.Fatalf("validate ai-safe config: %v", err)
	}
	if len(cfg.SafetyWarnings) != 1 {
		t.Fatalf("warnings = %d, want 1", len(cfg.SafetyWarnings))
	}
}

func TestStrictRejectsRealLookingSecret(t *testing.T) {
	cfg := Config{
		Server: ServerConfig{Port: 43101},
		Mode:   "strict",
		Adapters: map[string]AdapterConfig{
			"stripe": {Enabled: true, BasePath: "/stripe", FakeSecret: "sk_live_123"},
		},
	}

	if err := Validate(&cfg); err == nil {
		t.Fatal("strict config with real-looking secret returned nil error")
	}
}

func TestStrictRejectsExternalServiceURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{"stripe", "https://api.stripe.com"},
		{"openai", "https://api.openai.com"},
		{"github", "https://api.github.com"},
		{"line", "https://api.line.me"},
		{"slack", "https://slack.com/api"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{
				Server: ServerConfig{Port: 43101},
				Mode:   "strict",
				Adapters: map[string]AdapterConfig{
					"stripe": {Enabled: true, BasePath: "/stripe", APIURL: tt.url},
				},
			}
			if err := Validate(&cfg); err == nil {
				t.Fatal("strict config with external service URL returned nil error")
			}
		})
	}
}

func TestStrictRejectsNormalizedUnsafeConfigValues(t *testing.T) {
	tests := []struct {
		name      string
		adapter   AdapterConfig
		wantField string
	}{
		{
			name:      "quoted secret",
			adapter:   AdapterConfig{Enabled: true, BasePath: "/stripe", FakeSecret: " 'sk_live_123' "},
			wantField: "stripe.fake_secret",
		},
		{
			name:      "uppercase provider host with path and query",
			adapter:   AdapterConfig{Enabled: true, BasePath: "/stripe", APIURL: " HTTPS://API.STRIPE.COM/v1/checkout/sessions?limit=1 "},
			wantField: "stripe.api_url",
		},
		{
			name:      "quoted slack api path",
			adapter:   AdapterConfig{Enabled: true, BasePath: "/stripe", Webhook: WebhookConfig{TargetURL: "\"https://slack.com/api/chat.postMessage?token=x\""}},
			wantField: "stripe.webhook.target_url",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{
				Server: ServerConfig{Port: 43101},
				Mode:   "strict",
				Adapters: map[string]AdapterConfig{
					"stripe": tt.adapter,
				},
			}

			err := Validate(&cfg)
			if err == nil {
				t.Fatal("strict config with normalized unsafe value returned nil error")
			}
			if !strings.Contains(err.Error(), tt.wantField) {
				t.Fatalf("error = %q, want %q", err.Error(), tt.wantField)
			}
		})
	}
}

func TestAISafeRecordsExternalURLWarning(t *testing.T) {
	cfg := Config{
		Server: ServerConfig{Port: 43101},
		Mode:   "ai-safe",
		Adapters: map[string]AdapterConfig{
			"stripe": {Enabled: true, BasePath: "/stripe", APIURL: "https://api.stripe.com"},
		},
	}

	if err := Validate(&cfg); err != nil {
		t.Fatalf("validate ai-safe config: %v", err)
	}
	if len(cfg.SafetyWarnings) != 1 {
		t.Fatalf("warnings = %d, want 1", len(cfg.SafetyWarnings))
	}
	if cfg.SafetyWarnings[0].Category != "external_url" {
		t.Fatalf("category = %q, want external_url", cfg.SafetyWarnings[0].Category)
	}
}

func TestAISafeRecordsPublicBindWarning(t *testing.T) {
	cfg := Config{
		Server: ServerConfig{Host: "0.0.0.0", Port: 43101},
		Mode:   "ai-safe",
		Adapters: map[string]AdapterConfig{
			"stripe": {Enabled: true, BasePath: "/stripe"},
		},
	}

	if err := Validate(&cfg); err != nil {
		t.Fatalf("validate ai-safe config: %v", err)
	}
	if len(cfg.SafetyWarnings) != 1 {
		t.Fatalf("warnings = %d, want 1", len(cfg.SafetyWarnings))
	}
	if cfg.SafetyWarnings[0].Field != "server.host" || cfg.SafetyWarnings[0].Category != "public_bind" {
		t.Fatalf("warning = %+v, want server.host public_bind", cfg.SafetyWarnings[0])
	}
}

func TestStrictRejectsPublicBindHost(t *testing.T) {
	cfg := Config{
		Server: ServerConfig{Host: "::", Port: 43101},
		Mode:   "strict",
		Adapters: map[string]AdapterConfig{
			"stripe": {Enabled: true, BasePath: "/stripe"},
		},
	}

	err := Validate(&cfg)
	if err == nil {
		t.Fatal("strict config with public bind host returned nil error")
	}
	if !strings.Contains(err.Error(), "server.host") {
		t.Fatalf("error = %q, want server.host", err.Error())
	}
}
