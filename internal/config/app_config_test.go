package config

import (
	"testing"
	"time"
)

func TestDefaultAppConfig(t *testing.T) {
	cfg := DefaultAppConfig()

	if cfg.Server.Name != "metatools-mcp" {
		t.Errorf("Server.Name = %q, want %q", cfg.Server.Name, "metatools-mcp")
	}
	if cfg.Transport.Type != "stdio" {
		t.Errorf("Transport.Type = %q, want %q", cfg.Transport.Type, "stdio")
	}
	if cfg.Transport.HTTP.Port != 8080 {
		t.Errorf("Transport.HTTP.Port = %d, want %d", cfg.Transport.HTTP.Port, 8080)
	}
	if cfg.Execution.Timeout != 30*time.Second {
		t.Errorf("Execution.Timeout = %v, want %v", cfg.Execution.Timeout, 30*time.Second)
	}
}

func TestAppConfig_Validate(t *testing.T) {
	cfg := DefaultAppConfig()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestAppConfig_ValidateTransport(t *testing.T) {
	cfg := DefaultAppConfig()
	cfg.Transport.Type = "invalid"
	if err := cfg.Validate(); err == nil {
		t.Fatalf("Validate() should fail for invalid transport")
	}
}

func TestAppConfig_ValidateSearchStrategy(t *testing.T) {
	cfg := DefaultAppConfig()
	cfg.Search.Strategy = "invalid"
	if err := cfg.Validate(); err == nil {
		t.Fatalf("Validate() should fail for invalid search strategy")
	}
}
