package middleware

import (
	"errors"
	"sort"
	"sync"
)

// ErrMiddlewareNotFound is returned when middleware is not registered.
var ErrMiddlewareNotFound = errors.New("middleware not found")

// Factory creates a middleware instance from configuration.
type Factory func(cfg map[string]any) (Middleware, error)

// Registry manages middleware factories.
type Registry struct {
	mu        sync.RWMutex
	factories map[string]Factory
}

// NewRegistry creates a new middleware registry.
func NewRegistry() *Registry {
	return &Registry{factories: make(map[string]Factory)}
}

// Register adds a middleware factory.
func (r *Registry) Register(name string, factory Factory) error {
	if name == "" || factory == nil {
		return errors.New("invalid middleware registration")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.factories[name]; exists {
		return errors.New("middleware already registered")
	}
	r.factories[name] = factory
	return nil
}

// Has checks if a middleware is registered.
func (r *Registry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.factories[name]
	return ok
}

// Get retrieves a middleware factory by name.
func (r *Registry) Get(name string) (Factory, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	f, ok := r.factories[name]
	return f, ok
}

// Create instantiates middleware with configuration.
func (r *Registry) Create(name string, cfg map[string]any) (Middleware, error) {
	f, ok := r.Get(name)
	if !ok {
		return nil, ErrMiddlewareNotFound
	}
	return f(cfg)
}

// List returns all registered middleware names.
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.factories))
	for name := range r.factories {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Clear removes all registered middleware.
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.factories = make(map[string]Factory)
}
