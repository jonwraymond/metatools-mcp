//go:build toolsemantic

package bootstrap

import (
	"context"
	"testing"

	"github.com/jonwraymond/metatools-mcp/internal/config"
	semanticregistry "github.com/jonwraymond/metatools-mcp/internal/semantic"
	"github.com/jonwraymond/tooldiscovery/index"
	"github.com/jonwraymond/tooldiscovery/semantic"
	"github.com/stretchr/testify/require"
)

type stubEmbedder struct{}

func (stubEmbedder) Embed(_ context.Context, _ string) ([]float32, error) {
	return []float32{0.1, 0.2, 0.3}, nil
}

func TestSearcherFromConfig_Semantic(t *testing.T) {
	require.NoError(t, semanticregistry.RegisterEmbedder("stub", func(_ map[string]any) (semantic.Embedder, error) {
		return stubEmbedder{}, nil
	}))

	cfg := config.SearchConfig{
		Strategy:         "semantic",
		SemanticEmbedder: "stub",
	}

	searcher, err := SearcherFromConfig(cfg)
	require.NoError(t, err)
	require.NotNil(t, searcher)

	// Ensure the searcher implements deterministic interface
	_, ok := searcher.(index.DeterministicSearcher)
	require.True(t, ok)
}

func TestSearcherFromConfig_Hybrid(t *testing.T) {
	require.NoError(t, semanticregistry.RegisterEmbedder("stub-hybrid", func(_ map[string]any) (semantic.Embedder, error) {
		return stubEmbedder{}, nil
	}))

	cfg := config.SearchConfig{
		Strategy:         "hybrid",
		SemanticEmbedder: "stub-hybrid",
		SemanticWeight:   0.7,
	}

	searcher, err := SearcherFromConfig(cfg)
	require.NoError(t, err)
	require.NotNil(t, searcher)
}
