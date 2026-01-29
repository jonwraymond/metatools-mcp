package middleware

import "testing"

func TestBuildChainFromConfig_Unknown(t *testing.T) {
	registry := NewRegistry()

	cfg := &MiddlewareConfig{
		Chain: []string{"missing"},
	}

	if _, err := BuildChainFromConfig(registry, cfg); err == nil {
		t.Fatal("BuildChainFromConfig() should fail for unknown middleware")
	}
}

func TestBuildChainFromConfig_Success(t *testing.T) {
	registry := NewRegistry()
	_ = registry.Register("logging", LoggingMiddlewareFactory)

	cfg := &MiddlewareConfig{
		Chain: []string{"logging"},
		Configs: map[string]MiddlewareEntry{
			"logging": {Config: map[string]any{"include_args": true}},
		},
	}

	chain, err := BuildChainFromConfig(registry, cfg)
	if err != nil {
		t.Fatalf("BuildChainFromConfig() error = %v", err)
	}
	if chain.Len() != 1 {
		t.Fatalf("chain.Len() = %d, want 1", chain.Len())
	}
}

func TestDefaultRegistry(t *testing.T) {
	registry := DefaultRegistry()
	if !registry.Has("logging") {
		t.Fatal("DefaultRegistry missing logging middleware")
	}
	if !registry.Has("metrics") {
		t.Fatal("DefaultRegistry missing metrics middleware")
	}
}
