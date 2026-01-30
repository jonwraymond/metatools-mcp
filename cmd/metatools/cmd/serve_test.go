package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestServeCmd_Flags(t *testing.T) {
	cmd := newServeCmd()

	transportFlag := cmd.Flags().Lookup("transport")
	if transportFlag == nil {
		t.Fatal("--transport flag not found")
	}
	if transportFlag.DefValue != "stdio" {
		t.Errorf("--transport default = %q, want %q", transportFlag.DefValue, "stdio")
	}

	portFlag := cmd.Flags().Lookup("port")
	if portFlag == nil {
		t.Fatal("--port flag not found")
	}
	if portFlag.DefValue != "8080" {
		t.Errorf("--port default = %q, want %q", portFlag.DefValue, "8080")
	}

	configFlag := cmd.Flags().Lookup("config")
	if configFlag == nil {
		t.Fatal("--config flag not found")
	}
}

func TestServeCmd_TransportValidation(t *testing.T) {
	tests := []struct {
		transport string
		wantErr   bool
	}{
		{"stdio", false},
		{"sse", false},
		{"streamable", false},
		{"invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.transport, func(t *testing.T) {
			err := validateTransport(tt.transport)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateTransport(%q) error = %v, wantErr %v", tt.transport, err, tt.wantErr)
			}
		})
	}
}

func TestServeCmd_StdioIntegration(t *testing.T) {
	cfg := &ServeConfig{Transport: "stdio"}

	serverCfg, err := buildServerConfig(cfg)
	if err != nil {
		t.Fatalf("buildServerConfig() error = %v", err)
	}
	if serverCfg.Index == nil {
		t.Fatal("buildServerConfig() returned nil index")
	}
}

func TestServeCmd_EnvVars(t *testing.T) {
	oldTransport := os.Getenv("METATOOLS_TRANSPORT")
	oldPort := os.Getenv("METATOOLS_PORT")
	oldHost := os.Getenv("METATOOLS_HOST")
	oldConfig := os.Getenv("METATOOLS_CONFIG")
	defer func() {
		os.Setenv("METATOOLS_TRANSPORT", oldTransport)
		os.Setenv("METATOOLS_PORT", oldPort)
		os.Setenv("METATOOLS_HOST", oldHost)
		os.Setenv("METATOOLS_CONFIG", oldConfig)
	}()

	os.Setenv("METATOOLS_TRANSPORT", "sse")
	os.Setenv("METATOOLS_PORT", "9090")
	os.Setenv("METATOOLS_HOST", "127.0.0.1")
	os.Setenv("METATOOLS_CONFIG", "metatools.yaml")

	cmd := newServeCmd()
	if err := cmd.ParseFlags([]string{}); err != nil {
		t.Fatalf("ParseFlags() error = %v", err)
	}

	transport, _ := cmd.Flags().GetString("transport")
	port, _ := cmd.Flags().GetInt("port")
	host, _ := cmd.Flags().GetString("host")
	config, _ := cmd.Flags().GetString("config")

	if transport != "sse" {
		t.Errorf("transport = %q, want %q from env", transport, "sse")
	}
	if port != 9090 {
		t.Errorf("port = %d, want %d from env", port, 9090)
	}
	if host != "127.0.0.1" {
		t.Errorf("host = %q, want %q from env", host, "127.0.0.1")
	}
	if config != "metatools.yaml" {
		t.Errorf("config = %q, want %q from env", config, "metatools.yaml")
	}
}

func TestServeCmd_ConfigFile(t *testing.T) {
	clearServeEnv(t)
	dir := t.TempDir()
	configPath := filepath.Join(dir, "metatools.yaml")

	yaml := `
transport:
  type: sse
  http:
    port: 9999
`
	if err := os.WriteFile(configPath, []byte(yaml), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	cfg, err := loadServeConfig(configPath, &ServeConfig{})
	if err != nil {
		t.Fatalf("loadServeConfig() error = %v", err)
	}

	if cfg.Transport.Type != "sse" {
		t.Errorf("Transport.Type = %q, want %q", cfg.Transport.Type, "sse")
	}
	if cfg.Transport.HTTP.Port != 9999 {
		t.Errorf("Transport.HTTP.Port = %d, want %d", cfg.Transport.HTTP.Port, 9999)
	}
}

func TestServeCmd_CLIOverridesConfig(t *testing.T) {
	clearServeEnv(t)
	dir := t.TempDir()
	configPath := filepath.Join(dir, "metatools.yaml")

	yaml := `
transport:
  type: sse
  http:
    port: 9999
`
	if err := os.WriteFile(configPath, []byte(yaml), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	cliCfg := &ServeConfig{
		Transport: "streamable",
		Port:      3000,
	}

	cfg, err := loadServeConfig(configPath, cliCfg)
	if err != nil {
		t.Fatalf("loadServeConfig() error = %v", err)
	}

	if cfg.Transport.Type != "streamable" {
		t.Errorf("Transport.Type = %q, want %q from CLI", cfg.Transport.Type, "streamable")
	}
	if cfg.Transport.HTTP.Port != 3000 {
		t.Errorf("Transport.HTTP.Port = %d, want %d from CLI", cfg.Transport.HTTP.Port, 3000)
	}
}

func TestServeCmd_StreamableConfigFile(t *testing.T) {
	clearServeEnv(t)
	dir := t.TempDir()
	configPath := filepath.Join(dir, "metatools.yaml")

	yaml := `
transport:
  type: streamable
  http:
    host: 127.0.0.1
    port: 8080
  streamable:
    stateless: true
    json_response: true
    session_timeout: 15m
`
	if err := os.WriteFile(configPath, []byte(yaml), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	cfg, err := loadServeConfig(configPath, &ServeConfig{})
	if err != nil {
		t.Fatalf("loadServeConfig() error = %v", err)
	}

	if cfg.Transport.Type != "streamable" {
		t.Errorf("Transport.Type = %q, want %q", cfg.Transport.Type, "streamable")
	}
	if cfg.Transport.HTTP.Host != "127.0.0.1" {
		t.Errorf("Transport.HTTP.Host = %q, want %q", cfg.Transport.HTTP.Host, "127.0.0.1")
	}
	if cfg.Transport.HTTP.Port != 8080 {
		t.Errorf("Transport.HTTP.Port = %d, want %d", cfg.Transport.HTTP.Port, 8080)
	}
	if !cfg.Transport.Streamable.Stateless {
		t.Error("Transport.Streamable.Stateless = false, want true")
	}
	if !cfg.Transport.Streamable.JSONResponse {
		t.Error("Transport.Streamable.JSONResponse = false, want true")
	}
}

func clearServeEnv(t *testing.T) {
	t.Helper()
	vars := []string{
		"METATOOLS_TRANSPORT",
		"METATOOLS_TRANSPORT_TYPE",
		"METATOOLS_TRANSPORT_HTTP_PORT",
		"METATOOLS_TRANSPORT_HTTP_HOST",
	}
	for _, v := range vars {
		_ = os.Unsetenv(v)
	}
}
