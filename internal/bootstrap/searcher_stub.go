//go:build !toolsearch

package bootstrap

import (
	"fmt"

	"github.com/jonwraymond/metatools-mcp/internal/config"
	"github.com/jonwraymond/tooldiscovery/index"
)

// bm25SearcherFromConfig returns an error when bm25 is requested without toolsearch.
func bm25SearcherFromConfig(cfg config.SearchConfig) (index.Searcher, error) {
	if cfg.Strategy == "bm25" {
		return nil, fmt.Errorf("bm25 strategy requested but toolsearch build tag is not enabled")
	}
	return nil, nil
}
