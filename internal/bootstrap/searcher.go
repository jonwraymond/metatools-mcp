package bootstrap

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/jonwraymond/metatools-mcp/internal/config"
	"github.com/jonwraymond/tooldiscovery/index"
)

// SearcherFromConfig selects a searcher based on configuration.
// Returns nil to use the default lexical search.
func SearcherFromConfig(cfg config.SearchConfig) (index.Searcher, error) {
	switch strings.ToLower(strings.TrimSpace(cfg.Strategy)) {
	case "", "lexical":
		return nil, nil
	case "bm25":
		return bm25SearcherFromConfig(cfg)
	case "semantic":
		if strings.TrimSpace(cfg.SemanticEmbedder) == "" {
			slog.Default().Warn("semantic search requested without embedder; falling back to lexical")
			return nil, nil
		}
		return semanticSearcherFromConfig(cfg)
	case "hybrid":
		if strings.TrimSpace(cfg.SemanticEmbedder) == "" {
			slog.Default().Warn("hybrid search requested without embedder; falling back to bm25/lexical")
			searcher, err := bm25SearcherFromConfig(cfg)
			if err != nil {
				// If BM25 isn't available, fall back to lexical.
				return nil, nil
			}
			return searcher, nil
		}
		return semanticSearcherFromConfig(cfg)
	default:
		return nil, fmt.Errorf("unknown search strategy %q", cfg.Strategy)
	}
}
