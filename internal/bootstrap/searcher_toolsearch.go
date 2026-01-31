//go:build toolsearch

package bootstrap

import (
	"github.com/jonwraymond/metatools-mcp/internal/config"
	"github.com/jonwraymond/tooldiscovery/index"
	"github.com/jonwraymond/tooldiscovery/search"
)

// SearcherFromConfig returns a BM25 searcher when strategy is "bm25",
// otherwise returns nil to use toolindex's default lexical search.
func SearcherFromConfig(cfg config.SearchConfig) (index.Searcher, error) {
	if cfg.Strategy != "bm25" {
		return nil, nil
	}
	return search.NewBM25Searcher(search.BM25Config{
		NameBoost:      cfg.BM25NameBoost,
		NamespaceBoost: cfg.BM25NamespaceBoost,
		TagsBoost:      cfg.BM25TagsBoost,
		MaxDocs:        cfg.BM25MaxDocs,
		MaxDocTextLen:  cfg.BM25MaxDocTextLen,
	}), nil
}
