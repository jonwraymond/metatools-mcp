package middleware

import "fmt"

// Config is the top-level middleware configuration.
type Config struct {
	Chain   []string         `koanf:"chain"`
	Configs map[string]Entry `koanf:"configs"`
}

// Entry configures a single middleware.
type Entry struct {
	Config map[string]any `koanf:"config"`
}

// BuildChainFromConfig creates a middleware chain from configuration.
func BuildChainFromConfig(registry *Registry, cfg *Config) (*Chain, error) {
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
	_ = registry.Register("auth", AuthMiddlewareFactory)
	_ = registry.Register("logging", LoggingMiddlewareFactory)
	_ = registry.Register("metrics", MetricsMiddlewareFactory)
	_ = registry.Register("ratelimit", RateLimitMiddlewareFactory)
	_ = registry.Register("audit", AuditLoggingMiddlewareFactory)
	return registry
}
