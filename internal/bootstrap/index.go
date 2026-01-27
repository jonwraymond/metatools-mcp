package bootstrap

import (
	"fmt"

	"github.com/jonwraymond/metatools-mcp/internal/config"
	"github.com/jonwraymond/toolindex"
)

// NewIndexFromConfig creates a toolindex.Index configured based on EnvConfig.
// When the searcher is nil (lexical strategy or default build), toolindex
// uses its default lexical search behavior.
func NewIndexFromConfig(cfg config.EnvConfig) (toolindex.Index, error) {
	searcher, err := SearcherFromConfig(cfg.Search)
	if err != nil {
		return nil, fmt.Errorf("selecting searcher: %w", err)
	}
	return toolindex.NewInMemoryIndex(toolindex.IndexOptions{
		Searcher: searcher,
	}), nil
}
