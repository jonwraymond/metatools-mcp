package provider

import (
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ErrProviderNotFound is returned when unregistering a missing provider.
var ErrProviderNotFound = errors.New("provider not found")

// Registry stores tool providers by name.
type Registry struct {
	mu        sync.RWMutex
	providers map[string]ToolProvider
}

// NewRegistry creates a new ProviderRegistry.
func NewRegistry() *Registry {
	return &Registry{providers: make(map[string]ToolProvider)}
}

// Register registers a tool provider.
func (r *Registry) Register(p ToolProvider) error {
	if p == nil {
		return fmt.Errorf("provider is nil")
	}
	name := p.Name()
	if name == "" {
		return fmt.Errorf("provider name is required")
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.providers[name]; exists {
		return fmt.Errorf("provider %q already registered", name)
	}
	r.providers[name] = p
	return nil
}

// Unregister removes a provider by name.
func (r *Registry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.providers[name]; !ok {
		return ErrProviderNotFound
	}
	delete(r.providers, name)
	return nil
}

// Get returns a provider by name.
func (r *Registry) Get(name string) (ToolProvider, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.providers[name]
	return p, ok
}

// List returns all providers.
func (r *Registry) List() []ToolProvider {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]ToolProvider, 0, len(r.providers))
	for _, p := range r.providers {
		out = append(out, p)
	}
	return out
}

// ListEnabled returns enabled providers.
func (r *Registry) ListEnabled() []ToolProvider {
	all := r.List()
	out := make([]ToolProvider, 0, len(all))
	for _, p := range all {
		if p.Enabled() {
			out = append(out, p)
		}
	}
	return out
}

// Tools returns the MCP tool definitions for enabled providers.
func (r *Registry) Tools() []mcp.Tool {
	providers := r.ListEnabled()
	out := make([]mcp.Tool, 0, len(providers))
	for _, p := range providers {
		out = append(out, p.Tool())
	}
	return out
}

// Names returns provider names sorted for deterministic output.
func (r *Registry) Names() []string {
	all := r.List()
	out := make([]string, 0, len(all))
	for _, p := range all {
		out = append(out, p.Name())
	}
	sort.Strings(out)
	return out
}
