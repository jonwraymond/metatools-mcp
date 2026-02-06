//go:build toolsemantic

package bootstrap

import (
	"context"
	"fmt"

	"github.com/jonwraymond/metatools-mcp/internal/config"
	semanticregistry "github.com/jonwraymond/metatools-mcp/internal/semantic"
	"github.com/jonwraymond/tooldiscovery/index"
	"github.com/jonwraymond/tooldiscovery/semantic"
)

func semanticSearcherFromConfig(cfg config.SearchConfig) (index.Searcher, error) {
	embedder, err := semanticregistry.NewEmbedder(cfg.SemanticEmbedder, cfg.SemanticConfig)
	if err != nil {
		return nil, err
	}

	embedding := semantic.NewEmbeddingStrategy(embedder)
	switch cfg.Strategy {
	case "semantic":
		return newSemanticIndexSearcher(embedding), nil
	case "hybrid":
		bm25 := semantic.NewBM25Strategy(nil)
		hybrid, err := semantic.NewHybridStrategy(bm25, embedding, cfg.SemanticWeight)
		if err != nil {
			return nil, err
		}
		return newSemanticIndexSearcher(hybrid), nil
	default:
		return nil, fmt.Errorf("semantic searcher does not support strategy %q", cfg.Strategy)
	}
}

type semanticIndexSearcher struct {
	strategy semantic.Strategy
}

func newSemanticIndexSearcher(strategy semantic.Strategy) index.Searcher {
	return &semanticIndexSearcher{strategy: strategy}
}

func (s *semanticIndexSearcher) Search(query string, limit int, docs []index.SearchDoc) ([]index.Summary, error) {
	if limit <= 0 {
		return []index.Summary{}, nil
	}
	if s.strategy == nil {
		return nil, semantic.ErrInvalidSearcher
	}

	semDocs := semantic.DocumentsFromSearchDocs(docs)
	idx := semantic.NewInMemoryIndex()
	for _, doc := range semDocs {
		if err := idx.Add(context.Background(), doc); err != nil {
			return nil, err
		}
	}

	searcher := semantic.NewSearcher(idx, s.strategy)
	results, err := searcher.Search(context.Background(), query)
	if err != nil {
		return nil, err
	}

	lookup := make(map[string]index.Summary, len(docs))
	for _, doc := range docs {
		lookup[doc.ID] = doc.Summary
	}

	out := make([]index.Summary, 0, minInt(limit, len(results)))
	for _, res := range results {
		summary, ok := lookup[res.Document.ID]
		if !ok {
			continue
		}
		out = append(out, summary)
		if len(out) >= limit {
			break
		}
	}
	return out, nil
}

func (s *semanticIndexSearcher) Deterministic() bool { return true }

var _ index.DeterministicSearcher = (*semanticIndexSearcher)(nil)

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
