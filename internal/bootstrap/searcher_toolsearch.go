//go:build toolsearch

package bootstrap

import (
	"github.com/jonwraymond/metatools-mcp/internal/config"
	"github.com/jonwraymond/tooldiscovery/index"
	"github.com/jonwraymond/tooldiscovery/search"
)

// bm25SearcherFromConfig returns a BM25 searcher.
func bm25SearcherFromConfig(cfg config.SearchConfig) (index.Searcher, error) {
	return search.NewBM25Searcher(search.BM25Config{
		NameBoost:      cfg.BM25NameBoost,
		NamespaceBoost: cfg.BM25NamespaceBoost,
		TagsBoost:      cfg.BM25TagsBoost,
		MaxDocs:        cfg.BM25MaxDocs,
		MaxDocTextLen:  cfg.BM25MaxDocTextLen,
	}), nil
}
