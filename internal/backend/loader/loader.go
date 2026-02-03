// Package loader constructs backend registries from configuration.
package loader

import (
	"github.com/jonwraymond/metatools-mcp/internal/config"
	"github.com/jonwraymond/toolexec/backend"
	"github.com/jonwraymond/toolexec/backend/local"
)

const defaultLocalBackendName = "local"

// LoadFromConfig creates a backend registry from configuration.
func LoadFromConfig(cfg config.BackendsConfig) (*backend.Registry, error) {
	registry := backend.NewRegistry()

	if cfg.Local.Enabled {
		localBackend := local.New(defaultLocalBackendName)
		if err := registry.Register(localBackend); err != nil {
			return nil, err
		}
	}

	return registry, nil
}
