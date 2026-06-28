package cli

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/albert-einshutoin/mockport/internal/config"
)

func TestHealthcheckCommandChecksConfiguredHealthURL(t *testing.T) {
	server := newJSONHTTPServer(t, map[string]string{
		"status": "ok",
	})
	defer server.Close()

	configPath := createHealthcheckConfigForServer(t, server, false)
	cmd, out := newTestCommand(t, "healthcheck", "--config", configPath)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute healthcheck: %v", err)
	}
	if !strings.Contains(out.String(), "/health") {
		t.Fatalf("missing healthcheck output detail: %s", out.String())
	}
}

func TestHealthcheckCommandRejectsBadResponse(t *testing.T) {
	server := newJSONHTTPServer(t, map[string]string{
		"status": "error",
	})
	defer server.Close()

	configPath := createHealthcheckConfigForServer(t, server, false)
	cmd, _ := newTestCommand(t, "healthcheck", "--config", configPath)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("healthcheck succeeded for non-ok status payload")
	}
	errText := err.Error()
	if !strings.Contains(errText, "healthcheck status value") {
		t.Fatalf("error=%v", err)
	}
}

func TestHealthcheckCommandSupportsExplicitURL(t *testing.T) {
	server := newJSONHTTPServer(t, map[string]string{
		"status": "ok",
	})
	defer server.Close()

	cmd, out := newTestCommand(t, "healthcheck", "--url", server.URL+"/health")
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute healthcheck --url: %v", err)
	}
	if !strings.Contains(out.String(), "mockport healthcheck passed") {
		t.Fatalf("output=%s", out.String())
	}
}

func TestLoadHealthcheckConfigFallsBackToDefaultWithoutConfig(t *testing.T) {
	cfg, err := loadHealthcheckConfig(filepath.Join(t.TempDir(), "missing.yml"))
	if err != nil {
		t.Fatalf("loadHealthcheckConfig fallback: %v", err)
	}
	if cfg.Server.Host != "127.0.0.1" || cfg.Server.Port != config.DefaultPort {
		t.Fatalf("fallback config = %+v, want host=127.0.0.1 port=%d", cfg.Server, config.DefaultPort)
	}
}

func TestHealthcheckCommandNormalizesPublicBindHost(t *testing.T) {
	server := newJSONHTTPServer(t, map[string]string{"status": "ok"})
	defer server.Close()

	configPath := createHealthcheckConfigForServer(t, server, true)
	cmd, out := newTestCommand(t, "healthcheck", "--config", configPath)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute healthcheck with 0.0.0.0 host: %v", err)
	}
	if !strings.Contains(out.String(), "/health") {
		t.Fatalf("missing healthcheck output detail: %s", out.String())
	}
}

func newJSONHTTPServer(t *testing.T, payload map[string]string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			t.Fatalf("path = %q, want /health", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(payload); err != nil {
			t.Fatalf("encode payload: %v", err)
		}
	}))
}

func createHealthcheckConfigForServer(t *testing.T, server *httptest.Server, hostOverride bool) string {
	t.Helper()
	parsed, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("parse server URL: %v", err)
	}
	host, portStr, err := net.SplitHostPort(parsed.Host)
	if err != nil {
		t.Fatalf("split host/port: %v", err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		t.Fatalf("atoi port: %v", err)
	}
	actualHost := host
	if hostOverride {
		actualHost = "0.0.0.0"
	}
	configPath := filepath.Join(t.TempDir(), "mockport.yml")
	content := fmt.Sprintf(`version: "0.1"
server:
  host: %s
  port: %d
mode: ai-safe
`, actualHost, port)
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return configPath
}
