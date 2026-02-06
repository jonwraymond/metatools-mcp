//go:build toolsearch

package bootstrap

import (
	"testing"

	"github.com/jonwraymond/metatools-mcp/internal/config"
	"github.com/jonwraymond/tooldiscovery/search"
	"github.com/jonwraymond/toolfoundation/model"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearcherFromConfig_BM25Strategy(t *testing.T) {
	cfg := config.SearchConfig{
		Strategy:           "bm25",
		BM25NameBoost:      3,
		BM25NamespaceBoost: 2,
		BM25TagsBoost:      2,
		BM25MaxDocs:        0,
		BM25MaxDocTextLen:  0,
	}

	searcher, err := SearcherFromConfig(cfg)
	require.NoError(t, err)
	assert.NotNil(t, searcher)

	// Verify it's the correct type
	_, ok := searcher.(*search.BM25Searcher)
	assert.True(t, ok, "expected *search.BM25Searcher")
}

func TestSearcherFromConfig_LexicalStrategy(t *testing.T) {
	cfg := config.SearchConfig{
		Strategy: "lexical",
	}

	searcher, err := SearcherFromConfig(cfg)
	require.NoError(t, err)
	assert.Nil(t, searcher, "lexical strategy should return nil")
}

func TestSearcherFromConfig_SemanticRequiresBuildTag(t *testing.T) {
	cfg := config.SearchConfig{
		Strategy:         "semantic",
		SemanticEmbedder: "stub",
	}

	searcher, err := SearcherFromConfig(cfg)
	require.Error(t, err)
	assert.Nil(t, searcher)
}

func TestSearcherFromConfig_CustomBM25Values(t *testing.T) {
	cfg := config.SearchConfig{
		Strategy:           "bm25",
		BM25NameBoost:      5,
		BM25NamespaceBoost: 4,
		BM25TagsBoost:      1,
		BM25MaxDocs:        1000,
		BM25MaxDocTextLen:  500,
	}

	searcher, err := SearcherFromConfig(cfg)
	require.NoError(t, err)
	assert.NotNil(t, searcher)

	// We can't easily verify the internal config values, but we can
	// verify it was created without error
	_, ok := searcher.(*search.BM25Searcher)
	assert.True(t, ok)
}

func TestNewIndexFromConfig_BM25Strategy(t *testing.T) {
	cfg := config.EnvConfig{
		Search: config.SearchConfig{
			Strategy:           "bm25",
			BM25NameBoost:      3,
			BM25NamespaceBoost: 2,
			BM25TagsBoost:      2,
		},
	}

	idx, err := NewIndexFromConfig(cfg)
	require.NoError(t, err)
	require.NotNil(t, idx)

	// Verify the index works by registering and searching for a tool
	tool := model.Tool{
		Tool: mcp.Tool{
			Name:        "bm25_test_tool",
			Description: "A test tool for BM25 verification",
			InputSchema: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
		},
		Namespace: "testing",
	}

	backend := model.ToolBackend{
		Kind: model.BackendKindLocal,
		Local: &model.LocalBackend{
			Name: "test_handler",
		},
	}

	err = idx.RegisterTool(tool, backend)
	require.NoError(t, err)

	// Search should return the tool using BM25 ranking
	results, err := idx.Search("bm25", 10)
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "bm25_test_tool", results[0].Name)
}
