// Package bootstrap wires core services from configuration.
package bootstrap

import (
	"fmt"

	"github.com/jonwraymond/metatools-mcp/internal/config"
	"github.com/jonwraymond/tooldiscovery/index"
)

// NewIndexFromSearchConfig creates a index.Index configured from a SearchConfig.
// When the searcher is nil (lexical strategy or default build), toolindex
// uses its default lexical search behavior.
func NewIndexFromSearchConfig(cfg config.SearchConfig) (index.Index, error) {
	searcher, err := SearcherFromConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("selecting searcher: %w", err)
	}
	return index.NewInMemoryIndex(index.IndexOptions{
		Searcher: searcher,
	}), nil
}

// NewIndexFromConfig creates a index.Index configured based on EnvConfig.
func NewIndexFromConfig(cfg config.EnvConfig) (index.Index, error) {
	return NewIndexFromSearchConfig(cfg.Search)
}

// NewIndexFromAppConfig creates a index.Index configured based on AppConfig.
func NewIndexFromAppConfig(cfg config.AppConfig) (index.Index, error) {
	return NewIndexFromSearchConfig(cfg.Search.ToSearchConfig())
}
