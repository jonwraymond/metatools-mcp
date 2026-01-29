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
	t.Setenv("METATOOLS_TRANSPORT_TYPE", "http")
	t.Setenv("METATOOLS_TRANSPORT_HTTP_PORT", "3000")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Transport.Type != "http" {
		t.Errorf("Transport.Type = %q, want %q from env", cfg.Transport.Type, "http")
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
