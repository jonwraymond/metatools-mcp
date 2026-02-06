package middleware

import (
	"fmt"
	"time"

	"github.com/jonwraymond/toolops/cache"
	"github.com/jonwraymond/toolops/observe"
	"github.com/jonwraymond/toolops/resilience"
)

// Config is the top-level middleware configuration.
type Config struct {
	Chain      []string         `koanf:"chain"`
	Configs    map[string]Entry `koanf:"configs"`
	Observe    ObserveConfig    `koanf:"observe"`
	Cache      CacheConfig      `koanf:"cache"`
	Resilience ResilienceConfig `koanf:"resilience"`
}

// Entry configures a single middleware.
type Entry struct {
	Config map[string]any `koanf:"config"`
}

// ObserveConfig configures observability middleware.
type ObserveConfig struct {
	Enabled bool           `koanf:"enabled"`
	Config  observe.Config `koanf:"config"`
}

// CacheConfig configures cache middleware.
type CacheConfig struct {
	Enabled bool         `koanf:"enabled"`
	Policy  cache.Policy `koanf:"policy"`
}

// ResilienceConfig configures resilience middleware.
type ResilienceConfig struct {
	Enabled bool                    `koanf:"enabled"`
	Retry   ResilienceRetryConfig   `koanf:"retry"`
	Circuit ResilienceCircuitConfig `koanf:"circuit"`
	Timeout time.Duration           `koanf:"timeout"`
}

// ResilienceRetryConfig toggles retry behavior.
type ResilienceRetryConfig struct {
	Enabled bool                   `koanf:"enabled"`
	Config  resilience.RetryConfig `koanf:"config"`
}

// ResilienceCircuitConfig toggles circuit breaker behavior.
type ResilienceCircuitConfig struct {
	Enabled bool                            `koanf:"enabled"`
	Config  resilience.CircuitBreakerConfig `koanf:"config"`
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
