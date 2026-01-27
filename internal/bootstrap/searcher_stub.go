//go:build !toolsearch

package bootstrap

import (
	"fmt"

	"github.com/jonwraymond/metatools-mcp/internal/config"
	"github.com/jonwraymond/toolindex"
)

// SearcherFromConfig returns nil in the default build, delegating to
// toolindex's default lexical search behavior. If BM25 is requested without
// the toolsearch build tag, fail fast to avoid silent lexical fallback.
func SearcherFromConfig(cfg config.SearchConfig) (toolindex.Searcher, error) {
	if cfg.Strategy == "bm25" {
		return nil, fmt.Errorf("bm25 strategy requested but toolsearch build tag is not enabled")
	}
	return nil, nil
}
