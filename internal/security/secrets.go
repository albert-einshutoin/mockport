package security

import (
	"slices"
	"strings"
)

var dangerousSecretPrefixes = []string{
	"sk_live_",
	"sk_test_",
	"AKIA",
	"ASIA",
	"ghp_",
	"github_pat_",
	"xoxb-",
	"xoxp-",
	"AIza",
	"whsec_",
}

var fakeSecretPrefixes = []string{
	"mockport_",
	"local_",
	"fake_",
	"dummy_",
}

func LooksLikeSecret(value string) bool {
	value = NormalizePublicSafetyValue(value)
	if value == "whsec_mockport" {
		return false
	}
	hasPrefix := func(prefix string) bool { return strings.HasPrefix(value, prefix) }
	if slices.ContainsFunc(fakeSecretPrefixes, hasPrefix) {
		return false
	}
	if slices.ContainsFunc(dangerousSecretPrefixes, hasPrefix) {
		return true
	}
	return false
}

func NormalizePublicSafetyValue(value string) string {
	return strings.Trim(strings.TrimSpace(value), `"'`)
}

type PublicEnvFinding struct {
	Line   int
	Key    string
	Reason string
}

func ScanPublicEnv(content string) []PublicEnvFinding {
	var findings []PublicEnvFinding
	for idx, rawLine := range strings.Split(content, "\n") {
		line := strings.TrimSpace(rawLine)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = NormalizePublicSafetyValue(value)
		lower := strings.ToLower(value)
		switch {
		case LooksLikeSecret(value):
			findings = append(findings, PublicEnvFinding{Line: idx + 1, Key: key, Reason: "real-looking provider secret"})
		case LooksLikeExternalServiceURL(value):
			findings = append(findings, PublicEnvFinding{Line: idx + 1, Key: key, Reason: "production provider URL"})
		case strings.Contains(lower, "changeme") || strings.Contains(lower, "replace_me"):
			findings = append(findings, PublicEnvFinding{Line: idx + 1, Key: key, Reason: "ambiguous placeholder"})
		}
	}
	return findings
}
