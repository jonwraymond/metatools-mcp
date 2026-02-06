//go:build toolsemantic

package semantic

import (
	"fmt"
	"strings"
	"sync"

	"github.com/jonwraymond/tooldiscovery/semantic"
)

// EmbedderFactory constructs an embedder from configuration.
type EmbedderFactory func(config map[string]any) (semantic.Embedder, error)

var (
	registryMu sync.RWMutex
	registry   = map[string]EmbedderFactory{}
)

// RegisterEmbedder registers an embedder factory under a name.
func RegisterEmbedder(name string, factory EmbedderFactory) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("embedder name is required")
	}
	if factory == nil {
		return fmt.Errorf("embedder factory is required")
	}
	key := strings.ToLower(strings.TrimSpace(name))
	registryMu.Lock()
	defer registryMu.Unlock()
	registry[key] = factory
	return nil
}

// NewEmbedder creates an embedder by name.
func NewEmbedder(name string, config map[string]any) (semantic.Embedder, error) {
	if strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("embedder name is required")
	}
	key := strings.ToLower(strings.TrimSpace(name))
	registryMu.RLock()
	factory, ok := registry[key]
	registryMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("embedder %q not registered", key)
	}
	return factory(config)
}
