package middleware

import "fmt"

// MiddlewareConfig is the top-level middleware configuration.
type MiddlewareConfig struct {
	Chain   []string                   `koanf:"chain"`
	Configs map[string]MiddlewareEntry `koanf:"configs"`
}

// MiddlewareEntry configures a single middleware.
type MiddlewareEntry struct {
	Config map[string]any `koanf:"config"`
}

// BuildChainFromConfig creates a middleware chain from configuration.
func BuildChainFromConfig(registry *Registry, cfg *MiddlewareConfig) (*Chain, error) {
	chain := NewChain()
	if cfg == nil {
		return chain, nil
	}
	for _, name := range cfg.Chain {
		entry := cfg.Configs[name]
		mw, err := registry.Create(name, entry.Config)
		if err != nil {
			return nil, fmt.Errorf("middleware %s: %w", name, err)
		}
		chain.Use(mw)
	}
	return chain, nil
}

// DefaultRegistry returns a registry with built-in middleware.
func DefaultRegistry() *Registry {
	registry := NewRegistry()
	_ = registry.Register("logging", LoggingMiddlewareFactory)
	_ = registry.Register("metrics", MetricsMiddlewareFactory)
	return registry
}
