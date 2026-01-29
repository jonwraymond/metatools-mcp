package cmd

import (
	"os"
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
		{"http", false},
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
	cmd.ParseFlags([]string{})

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
