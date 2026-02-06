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
	if cfg.State.RuntimeLimitsDB != "" {
		t.Errorf("State.RuntimeLimitsDB = %q, want empty", cfg.State.RuntimeLimitsDB)
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

func TestAppConfig_ValidateExecutionLimits(t *testing.T) {
	cfg := DefaultAppConfig()
	cfg.Execution.MaxToolCalls = -1
	if err := cfg.Validate(); err == nil {
		t.Fatalf("Validate() should fail for negative max tool calls")
	}

	cfg = DefaultAppConfig()
	cfg.Execution.MaxChainSteps = -1
	if err := cfg.Validate(); err == nil {
		t.Fatalf("Validate() should fail for negative max chain steps")
	}
}

func TestAppConfig_ValidateMCPBackends(t *testing.T) {
	cfg := DefaultAppConfig()
	cfg.Backends.MCP = []MCPBackendConfig{{Name: "", URL: "https://example.com/mcp"}}
	if err := cfg.Validate(); err == nil {
		t.Fatalf("Validate() should fail for empty mcp backend name")
	}

	cfg = DefaultAppConfig()
	cfg.Backends.MCP = []MCPBackendConfig{{Name: "test", URL: ""}}
	if err := cfg.Validate(); err == nil {
		t.Fatalf("Validate() should fail for empty mcp backend url")
	}

	cfg = DefaultAppConfig()
	cfg.Backends.MCP = []MCPBackendConfig{
		{Name: "dup", URL: "https://example.com/mcp"},
		{Name: "dup", URL: "https://example.com/mcp"},
	}
	if err := cfg.Validate(); err == nil {
		t.Fatalf("Validate() should fail for duplicate mcp backend names")
	}
}

func TestAppConfig_ValidateMCPRefresh(t *testing.T) {
	cfg := DefaultAppConfig()
	cfg.Backends.MCPRefresh.Interval = -1
	if err := cfg.Validate(); err == nil {
		t.Fatalf("Validate() should fail for negative mcp refresh interval")
	}

	cfg = DefaultAppConfig()
	cfg.Backends.MCPRefresh.Jitter = -1
	if err := cfg.Validate(); err == nil {
		t.Fatalf("Validate() should fail for negative mcp refresh jitter")
	}

	cfg = DefaultAppConfig()
	cfg.Backends.MCPRefresh.StaleAfter = -1
	if err := cfg.Validate(); err == nil {
		t.Fatalf("Validate() should fail for negative mcp refresh stale_after")
	}
}
