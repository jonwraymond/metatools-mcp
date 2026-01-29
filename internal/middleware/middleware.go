package middleware

import (
	"fmt"

	"github.com/jonwraymond/metatools-mcp/internal/provider"
)

// Middleware wraps a ToolProvider to add cross-cutting concerns.
// Middleware functions receive the next provider in the chain and return
// a wrapped provider that adds behavior before/after the next provider.
type Middleware func(provider.ToolProvider) provider.ToolProvider

// Chain holds an ordered list of middleware to apply.
type Chain struct {
	middleware []Middleware
}

// NewChain creates a new middleware chain.
func NewChain(middleware ...Middleware) *Chain {
	return &Chain{middleware: middleware}
}

// Use adds middleware to the chain.
func (c *Chain) Use(mw Middleware) *Chain {
	c.middleware = append(c.middleware, mw)
	return c
}

// Apply wraps a provider with all middleware in the chain.
// Middleware is applied in order: first middleware wraps outermost.
func (c *Chain) Apply(p provider.ToolProvider) provider.ToolProvider {
	wrapped := p
	for i := len(c.middleware) - 1; i >= 0; i-- {
		wrapped = c.middleware[i](wrapped)
	}
	return wrapped
}

// ApplyToRegistry wraps all providers in a registry with the chain.
func (c *Chain) ApplyToRegistry(registry *provider.Registry) error {
	for _, p := range registry.List() {
		wrapped := c.Apply(p)
		if err := registry.Unregister(p.Name()); err != nil {
			return fmt.Errorf("unregister %q: %w", p.Name(), err)
		}
		if err := registry.Register(wrapped); err != nil {
			return fmt.Errorf("register %q: %w", p.Name(), err)
		}
	}
	return nil
}

// Len returns the number of middleware in the chain.
func (c *Chain) Len() int {
	return len(c.middleware)
}

// Clear removes all middleware from the chain.
func (c *Chain) Clear() {
	c.middleware = nil
}
