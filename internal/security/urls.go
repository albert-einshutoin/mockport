package security

import (
	"net"
	"net/url"
	"strings"
)

func LooksLikeExternalServiceURL(value string) bool {
	value = NormalizePublicSafetyValue(value)
	parsed, err := url.Parse(value)
	if err != nil || parsed.Host == "" {
		return false
	}
	switch strings.ToLower(parsed.Scheme) {
	case "http", "https":
	default:
		return false
	}

	host := normalizedURLHost(parsed)
	switch host {
	case "api.stripe.com", "api.openai.com", "api.github.com", "api.line.me", "hooks.slack.com":
		return true
	case "slack.com":
		return parsed.Path == "/api" || strings.HasPrefix(parsed.Path, "/api/")
	}
	return false
}

func IsLoopbackRemoteAddr(remoteAddr string) bool {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		host = remoteAddr
	}
	return IsLoopbackHost(host)
}

func IsLoopbackHost(host string) bool {
	host = strings.Trim(strings.ToLower(strings.TrimSpace(host)), "[]")
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func IsSafeWebhookTargetURL(value string) bool {
	parsed, err := url.Parse(value)
	if err != nil || parsed.Host == "" {
		return false
	}
	switch parsed.Scheme {
	case "http", "https":
	default:
		return false
	}
	if parsed.User != nil {
		return false
	}
	host := normalizedURLHost(parsed)
	switch host {
	case "localhost", "host.docker.internal", "app":
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func IsSafeOAuthRedirectURL(value string) bool {
	parsed, err := url.Parse(value)
	if err != nil || parsed.Host == "" {
		return false
	}
	switch parsed.Scheme {
	case "http", "https":
	default:
		return false
	}
	if parsed.User != nil {
		return false
	}
	host := normalizedURLHost(parsed)
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func normalizedURLHost(parsed *url.URL) string {
	host := parsed.Hostname()
	if host == "" {
		host = parsed.Host
	}
	return strings.Trim(strings.ToLower(strings.TrimSpace(host)), "[]")
}
