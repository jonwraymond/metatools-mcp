package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoad_DefaultsWhenNoFile(t *testing.T) {
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Transport.Type != "stdio" {
		t.Errorf("Transport.Type = %q, want %q", cfg.Transport.Type, "stdio")
	}
}

func TestLoad_FromYAMLFile(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "metatools.yaml")

	yaml := `
server:
  name: test-server
transport:
  type: sse
  http:
    port: 9090
search:
  strategy: bm25
  bm25:
    name_boost: 5
execution:
  timeout: 60s
`
	if err := os.WriteFile(configPath, []byte(yaml), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Server.Name != "test-server" {
		t.Errorf("Server.Name = %q, want %q", cfg.Server.Name, "test-server")
	}
	if cfg.Transport.Type != "sse" {
		t.Errorf("Transport.Type = %q, want %q", cfg.Transport.Type, "sse")
	}
	if cfg.Transport.HTTP.Port != 9090 {
		t.Errorf("Transport.HTTP.Port = %d, want %d", cfg.Transport.HTTP.Port, 9090)
	}
	if cfg.Search.BM25.NameBoost != 5 {
		t.Errorf("Search.BM25.NameBoost = %d, want %d", cfg.Search.BM25.NameBoost, 5)
	}
	if cfg.Execution.Timeout != 60*time.Second {
		t.Errorf("Execution.Timeout = %v, want %v", cfg.Execution.Timeout, 60*time.Second)
	}
}

func TestLoad_EnvOverrides(t *testing.T) {
	t.Setenv("METATOOLS_TRANSPORT_TYPE", "streamable")
	t.Setenv("METATOOLS_TRANSPORT_HTTP_PORT", "3000")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Transport.Type != "streamable" {
		t.Errorf("Transport.Type = %q, want %q from env", cfg.Transport.Type, "streamable")
	}
	if cfg.Transport.HTTP.Port != 3000 {
		t.Errorf("Transport.HTTP.Port = %d, want %d from env", cfg.Transport.HTTP.Port, 3000)
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "bad.yaml")

	if err := os.WriteFile(configPath, []byte("invalid: yaml: ["), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Error("Load() should fail with invalid YAML")
	}
}

func TestLoad_ExpandsEnvVarsInFile(t *testing.T) {
	t.Setenv("MCP_BACKEND_TOKEN", "secret-token")

	dir := t.TempDir()
	configPath := filepath.Join(dir, "metatools.yaml")

	yaml := `
backends:
  mcp:
    - name: deepwiki
      url: https://mcp.deepwiki.com/mcp
      headers:
        Authorization: "Bearer ${MCP_BACKEND_TOKEN}"
`
	if err := os.WriteFile(configPath, []byte(yaml), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if got := cfg.Backends.MCP[0].Headers["Authorization"]; got != "Bearer secret-token" {
		t.Errorf("Backends.MCP[0].Headers[\"Authorization\"] = %q, want %q", got, "Bearer secret-token")
	}
}

func TestLoad_EnvVarMissingInFile(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "metatools.yaml")

	yaml := `
backends:
  mcp:
    - name: deepwiki
      url: https://mcp.deepwiki.com/mcp
      headers:
        Authorization: "Bearer ${MISSING_TOKEN}"
`
	if err := os.WriteFile(configPath, []byte(yaml), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Fatal("Load() should fail when an env var referenced in the file is missing")
	}
}
