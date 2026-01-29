package server

import (
	"github.com/jonwraymond/metatools-mcp/internal/middleware"
	"github.com/jonwraymond/metatools-mcp/internal/provider"
)

// MiddlewareAdapter applies middleware to provider registries.
type MiddlewareAdapter struct {
	chain    *middleware.Chain
	registry *middleware.Registry
}

// NewMiddlewareAdapter creates a new adapter with the provided chain.
func NewMiddlewareAdapter(chain *middleware.Chain) *MiddlewareAdapter {
	if chain == nil {
		chain = middleware.NewChain()
	}
	return &MiddlewareAdapter{
		chain:    chain,
		registry: middleware.DefaultRegistry(),
	}
}

// NewMiddlewareAdapterFromConfig builds a middleware chain from configuration.
func NewMiddlewareAdapterFromConfig(cfg *middleware.Config) (*MiddlewareAdapter, error) {
	registry := middleware.DefaultRegistry()
	chain, err := middleware.BuildChainFromConfig(registry, cfg)
	if err != nil {
		return nil, err
	}
	return &MiddlewareAdapter{chain: chain, registry: registry}, nil
}

// ApplyToProviders wraps all providers in a registry with middleware.
func (a *MiddlewareAdapter) ApplyToProviders(registry *provider.Registry) error {
	if a.chain == nil {
		return nil
	}
	return a.chain.ApplyToRegistry(registry)
}

// Chain returns the middleware chain.
func (a *MiddlewareAdapter) Chain() *middleware.Chain {
	return a.chain
}

// Registry returns the middleware registry.
func (a *MiddlewareAdapter) Registry() *middleware.Registry {
	return a.registry
}
