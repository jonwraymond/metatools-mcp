package bootstrap

import (
	"fmt"

	"github.com/jonwraymond/metatools-mcp/internal/config"
	"github.com/jonwraymond/toolindex"
)

// NewIndexFromSearchConfig creates a toolindex.Index configured from a SearchConfig.
// When the searcher is nil (lexical strategy or default build), toolindex
// uses its default lexical search behavior.
func NewIndexFromSearchConfig(cfg config.SearchConfig) (toolindex.Index, error) {
	searcher, err := SearcherFromConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("selecting searcher: %w", err)
	}
	return toolindex.NewInMemoryIndex(toolindex.IndexOptions{
		Searcher: searcher,
	}), nil
}

// NewIndexFromConfig creates a toolindex.Index configured based on EnvConfig.
func NewIndexFromConfig(cfg config.EnvConfig) (toolindex.Index, error) {
	return NewIndexFromSearchConfig(cfg.Search)
}

// NewIndexFromAppConfig creates a toolindex.Index configured based on AppConfig.
func NewIndexFromAppConfig(cfg config.AppConfig) (toolindex.Index, error) {
	return NewIndexFromSearchConfig(cfg.Search.ToSearchConfig())
}
