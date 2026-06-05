package security

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLooksLikeSecret(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{"stripe live", "sk_live_123", true},
		{"stripe test", "sk_test_123", true},
		{"aws access key", "AKIAIOSFODNN7EXAMPLE", true},
		{"github token", "github_pat_abc", true},
		{"quoted live secret", " 'sk_live_123' ", true},
		{"mockport fake", "mockport_stripe_secret", false},
		{"quoted mockport fake", " \"mockport_stripe_secret\" ", false},
		{"mockport webhook fake", "whsec_mockport", false},
		{"local fake", "local_openai_key", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LooksLikeSecret(tt.value)
			if got != tt.want {
				t.Fatalf("LooksLikeSecret(%q) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestRedactSecret(t *testing.T) {
	if got := RedactSecret("mockport_stripe_secret"); got != "mockport_...cret" {
		t.Fatalf("RedactSecret fake = %q", got)
	}
	if got := RedactSecret("sk_live_123456789"); got != "[real-looking secret redacted]" {
		t.Fatalf("RedactSecret real-looking = %q", got)
	}
}

func TestRedactValue(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{"short", "short", "[redacted]"},
		{"long fake", "mockport_stripe_secret", "mockport_...cret"},
		{"real secret", "sk_live_123456789", "[real-looking secret redacted]"},
		{"env assignment", "STRIPE_SECRET_KEY=sk_live_123456789", "STRIPE_SECRET_KEY=[real-looking secret redacted]"},
		{"provider url", "https://api.stripe.com/v1/checkout/sessions", "[external service URL redacted]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RedactValue(tt.value); got != tt.want {
				t.Fatalf("RedactValue(%q) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}

func TestRedactMessage(t *testing.T) {
	got := RedactMessage("bad sk_live_123 and https://api.stripe.com")
	if strings.Contains(got, "sk_live_123") || strings.Contains(got, "https://api.stripe.com") {
		t.Fatalf("message leaked unsafe value: %q", got)
	}
}

func TestLooksLikeExternalServiceURL(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{"stripe path", "https://api.stripe.com/v1/checkout/sessions", true},
		{"stripe uppercase host", " HTTPS://API.STRIPE.COM/v1/checkout/sessions?expand[]=payment_intent ", true},
		{"openai quoted", "'https://api.openai.com/v1/chat/completions?timeout=1'", true},
		{"github query", "https://api.github.com/repos/acme/project?per_page=1", true},
		{"line path", "https://api.line.me/v2/bot/message/push", true},
		{"slack api path", "https://slack.com/api/chat.postMessage?token=x", true},
		{"slack webhook host", "https://hooks.slack.com/services/T000/B000/XXX", true},
		{"localhost", "http://localhost:43101/stripe", false},
		{"lookalike host", "https://example.com/api.stripe.com/v1", false},
		{"slack non api path", "https://slack.com/help", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LooksLikeExternalServiceURL(tt.value)
			if got != tt.want {
				t.Fatalf("LooksLikeExternalServiceURL(%q) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestSafeWebhookTargetURLAllowsOnlyLocalTargets(t *testing.T) {
	for _, value := range []string{
		"http://localhost:3000/webhook",
		"http://127.0.0.1:3000/webhook",
		"http://[::1]:3000/webhook",
		"http://host.docker.internal:3000/webhook",
		"http://app:3000/webhook",
	} {
		if !IsSafeWebhookTargetURL(value) {
			t.Fatalf("target %q was rejected, want allowed", value)
		}
	}
	for _, value := range []string{
		"https://example.com/webhook",
		"http://169.254.169.254/latest/meta-data",
		"ftp://localhost/webhook",
		"http://user:pass@localhost/webhook",
	} {
		if IsSafeWebhookTargetURL(value) {
			t.Fatalf("target %q was allowed, want rejected", value)
		}
	}
}

func TestSafeOAuthRedirectURLAllowsOnlyLoopbackTargets(t *testing.T) {
	for _, value := range []string{
		"http://localhost/callback",
		"http://127.0.0.1:3000/callback",
		"http://[::1]:3000/callback",
	} {
		if !IsSafeOAuthRedirectURL(value) {
			t.Fatalf("redirect %q was rejected, want allowed", value)
		}
	}
	for _, value := range []string{
		"http://app/callback",
		"https://example.com/callback",
		"javascript:alert(1)",
		"http://user:pass@localhost/callback",
	} {
		if IsSafeOAuthRedirectURL(value) {
			t.Fatalf("redirect %q was allowed, want rejected", value)
		}
	}
}

func TestIsLoopbackRemoteAddr(t *testing.T) {
	for _, value := range []string{"127.0.0.1:43101", "[::1]:43101", "localhost:43101"} {
		if !IsLoopbackRemoteAddr(value) {
			t.Fatalf("remote addr %q was rejected, want loopback", value)
		}
	}
	for _, value := range []string{"192.168.1.10:43101", "10.0.0.5:43101", "example.com:43101"} {
		if IsLoopbackRemoteAddr(value) {
			t.Fatalf("remote addr %q was allowed, want rejected", value)
		}
	}
}

func TestScanPublicEnvRejectsRealProviderSecretsAndURLs(t *testing.T) {
	env := strings.Join([]string{
		"STRIPE_SECRET_KEY=sk_live_123",
		"OPENAI_BASE_URL=https://api.openai.com/v1",
		"SLACK_BOT_TOKEN=xoxb-real-token",
		"GITHUB_OAUTH_CLIENT_SECRET=github_pat_real",
	}, "\n")

	findings := ScanPublicEnv(env)
	if len(findings) != 4 {
		t.Fatalf("finding count = %d, want 4: %#v", len(findings), findings)
	}
	for _, finding := range findings {
		if finding.Line == 0 || finding.Key == "" || finding.Reason == "" {
			t.Fatalf("incomplete finding: %#v", finding)
		}
	}
}

func TestScanPublicEnvAllowsMockportExamples(t *testing.T) {
	env := strings.Join([]string{
		"STRIPE_API_URL=http://localhost:43101/stripe",
		"STRIPE_SECRET_KEY=mockport_stripe_secret",
		"STRIPE_WEBHOOK_SECRET=whsec_mockport",
		"OPENAI_BASE_URL=http://localhost:43101/openai/v1",
		"OPENAI_API_KEY=mockport_openai_key",
		"GITHUB_OAUTH_CLIENT_SECRET=mockport_github_secret",
		"SLACK_BOT_TOKEN=mockport_slack_token",
	}, "\n")

	if findings := ScanPublicEnv(env); len(findings) != 0 {
		t.Fatalf("ScanPublicEnv() findings = %#v, want none", findings)
	}
}

func TestScanFixtureContentRequiresSourceMetadata(t *testing.T) {
	fixture := `{
  "provider": "stripe",
  "provider_version": "2026-05-26",
  "sdk": {"name": "stripe-go", "version": "v83.0.0"},
  "request": {"method": "POST", "path": "/v1/checkout/sessions"},
  "response": {"status": 200}
}`

	findings := ScanFixtureContent("compat/fixtures/schema.example.json", fixture)
	if len(findings) != 1 {
		t.Fatalf("finding count = %d, want 1: %#v", len(findings), findings)
	}
	if findings[0].Field != "source" || findings[0].Reason != "missing source metadata" {
		t.Fatalf("finding = %#v, want missing source metadata", findings[0])
	}
}

func TestScanFixtureContentRequiresSourceMetadataFields(t *testing.T) {
	fixture := `{
  "provider": "stripe",
  "provider_version": "2026-05-26",
  "sdk": {"name": "stripe-go", "version": "v83.0.0"},
  "source": {"type": "docs"},
  "request": {"method": "POST", "path": "/v1/checkout/sessions"},
  "response": {"status": 200}
}`

	findings := ScanFixtureContent("compat/fixtures/stripe/incomplete-source.json", fixture)
	reasons := map[string]bool{}
	for _, finding := range findings {
		reasons[finding.Field+":"+finding.Reason] = true
	}
	for _, key := range []string{
		"source.title:missing source metadata",
		"source.url_or_path:missing source metadata",
		"source.retrieved_at:missing source metadata",
	} {
		if !reasons[key] {
			t.Fatalf("missing finding %q in %#v", key, findings)
		}
	}
}

func TestScanFixtureContentRejectsUnsafeValues(t *testing.T) {
	fixture := `{
  "provider": "stripe",
  "provider_version": "2026-05-26",
  "sdk": {"name": "stripe-go", "version": "v83.0.0"},
  "source": {
    "type": "docs",
    "title": "Stripe Checkout Sessions",
    "url_or_path": "https://docs.stripe.com/api/checkout/sessions",
    "retrieved_at": "2026-05-26"
  },
  "request": {
    "method": "POST",
    "path": "/v1/checkout/sessions",
    "headers": {"Authorization": "Bearer sk_live_real"}
  },
  "response": {
    "status": 200,
    "body": {"livemode": false, "url": "https://api.stripe.com/v1/checkout/sessions"}
  }
}`

	findings := ScanFixtureContent("compat/fixtures/stripe/unsafe.json", fixture)
	if len(findings) != 2 {
		t.Fatalf("finding count = %d, want 2: %#v", len(findings), findings)
	}
	reasons := map[string]bool{}
	for _, finding := range findings {
		reasons[finding.Reason] = true
	}
	if !reasons["real-looking provider secret"] || !reasons["production provider URL"] {
		t.Fatalf("findings = %#v, want secret and production URL findings", findings)
	}
}

func TestScanFixtureContentAllowsSanitizedFixture(t *testing.T) {
	fixture := `{
  "provider": "stripe",
  "provider_version": "2026-05-26",
  "sdk": {"name": "stripe-go", "version": "v83.0.0"},
  "source": {
    "type": "docs",
    "title": "Stripe Checkout Sessions",
    "url_or_path": "https://docs.stripe.com/api/checkout/sessions",
    "retrieved_at": "2026-05-26"
  },
  "scenario": "payment_success",
  "request": {
    "method": "POST",
    "path": "/v1/checkout/sessions",
    "headers": {"Authorization": "Bearer mockport_stripe_secret"},
    "body": {"mode": "payment"}
  },
  "response": {
    "status": 200,
    "headers": {"Content-Type": "application/json"},
    "body": {"id": "cs_mockport_123", "livemode": false}
  }
}`

	if findings := ScanFixtureContent("compat/fixtures/schema.example.json", fixture); len(findings) != 0 {
		t.Fatalf("ScanFixtureContent() findings = %#v, want none", findings)
	}
}

func TestSchemaExampleFixtureIsPublicSafe(t *testing.T) {
	path := "../../compat/fixtures/schema.example.json"
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q): %v", path, err)
	}
	if findings := ScanFixtureContent(path, string(content)); len(findings) != 0 {
		t.Fatalf("schema example findings = %#v, want none", findings)
	}
}

func TestCompatibilityFixtureFilesArePublicSafe(t *testing.T) {
	root := "../../compat/fixtures"
	var checked int
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || filepath.Ext(path) != ".json" {
			return nil
		}
		checked++
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if findings := ScanFixtureContent(path, string(content)); len(findings) != 0 {
			t.Fatalf("%s findings = %#v, want none", path, findings)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("WalkDir(%q): %v", root, err)
	}
	if checked == 0 {
		t.Fatalf("checked fixture count = 0, want at least one")
	}
}
