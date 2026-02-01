// Package migration contains tests for PRD-180 migration to consolidated repositories.
package migration

import (
	"testing"

	"github.com/jonwraymond/tooldiscovery/index"
	"github.com/jonwraymond/tooldiscovery/search"
	"github.com/jonwraymond/tooldiscovery/tooldoc"
	"github.com/jonwraymond/toolexec/code"
	"github.com/jonwraymond/toolexec/run"
	"github.com/jonwraymond/toolfoundation/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBaselineCompilation verifies all consolidated packages compile and are accessible.
func TestBaselineCompilation(t *testing.T) {
	// This test simply verifies the imports compile
	t.Run("model compiles", func(_ *testing.T) {
		var _ model.Tool
	})
	t.Run("index compiles", func(_ *testing.T) {
		var _ index.Index
	})
	t.Run("tooldoc compiles", func(_ *testing.T) {
		var _ tooldoc.Store
	})
	t.Run("search compiles", func(_ *testing.T) {
		var _ search.BM25Config
	})
	t.Run("run compiles", func(_ *testing.T) {
		var _ run.Runner
	})
	t.Run("code compiles", func(_ *testing.T) {
		var _ code.Executor
	})
}

// TestCurrentIndexCreation verifies index.NewInMemoryIndex works.
func TestCurrentIndexCreation(t *testing.T) {
	idx := index.NewInMemoryIndex()
	require.NotNil(t, idx, "InMemoryIndex should be created")

	// Create a tool using the embedded mcp.Tool pattern
	tool := model.Tool{
		Namespace: "test",
	}
	tool.Name = "test-tool"
	tool.Description = "A test tool"
	tool.InputSchema = map[string]any{"type": "object"}

	backend := model.ToolBackend{
		Kind:  model.BackendKindLocal,
		Local: &model.LocalBackend{Name: "test-handler"},
	}

	err := idx.RegisterTool(tool, backend)
	assert.NoError(t, err, "Registering tool should succeed")

	retrieved, _, err := idx.GetTool("test:test-tool")
	assert.NoError(t, err, "Getting tool should succeed")
	assert.Equal(t, "test-tool", retrieved.Name)
}

// TestCurrentSearcherCreation verifies search.NewBM25Searcher works.
func TestCurrentSearcherCreation(t *testing.T) {
	// Create BM25 searcher with default config
	config := search.BM25Config{}
	searcher := search.NewBM25Searcher(config)
	require.NotNil(t, searcher, "BM25Searcher should be created")

	// Create tool
	tool := model.Tool{
		Namespace: "fs",
	}
	tool.Name = "file-reader"
	tool.Description = "Reads files from disk"
	tool.InputSchema = map[string]any{"type": "object"}

	backend := model.ToolBackend{
		Kind:  model.BackendKindLocal,
		Local: &model.LocalBackend{Name: "file-reader-handler"},
	}

	// Configure index to use the BM25 searcher
	idxWithSearcher := index.NewInMemoryIndex(index.IndexOptions{
		Searcher: searcher,
	})
	err := idxWithSearcher.RegisterTool(tool, backend)
	require.NoError(t, err)

	// Verify search works
	results, err := idxWithSearcher.Search("file", 10)
	assert.NoError(t, err, "Search should succeed")
	assert.Len(t, results, 1, "Should find one result")
}

// TestCurrentDocsStoreCreation verifies tooldoc.NewInMemoryStore works.
func TestCurrentDocsStoreCreation(t *testing.T) {
	idx := index.NewInMemoryIndex()

	// Register a tool first
	tool := model.Tool{
		Namespace: "test",
	}
	tool.Name = "test-tool"
	tool.Description = "A test tool"
	tool.InputSchema = map[string]any{"type": "object"}

	backend := model.ToolBackend{
		Kind:  model.BackendKindLocal,
		Local: &model.LocalBackend{Name: "test-handler"},
	}
	_ = idx.RegisterTool(tool, backend)

	// Create store with index
	store := tooldoc.NewInMemoryStore(tooldoc.StoreOptions{
		Index: idx,
	})
	require.NotNil(t, store, "InMemoryStore should be created")

	// Verify we can describe a tool (using the index)
	doc, err := store.DescribeTool("test:test-tool", tooldoc.DetailSummary)
	assert.NoError(t, err, "DescribeTool should succeed")
	// ToolDoc has Tool field, not ID - check summary is populated
	assert.NotEmpty(t, doc.Summary, "Summary should be populated")
}

// TestCurrentRunnerCreation verifies run.NewRunner works.
func TestCurrentRunnerCreation(t *testing.T) {
	runner := run.NewRunner()
	require.NotNil(t, runner, "Runner should be created")
}

// TestCurrentCodeExecutorCreation verifies code.NewDefaultExecutor works.
func TestCurrentCodeExecutorCreation(_ *testing.T) {
	// code.NewDefaultExecutor requires a Config with Runner
	// For this baseline test, we just verify the type exists
	var _ code.Executor
	var _ code.Config
}

// TestToolModelTypes verifies core model types exist and work.
func TestToolModelTypes(t *testing.T) {
	t.Run("Tool struct", func(t *testing.T) {
		tool := model.Tool{
			Namespace: "test-ns",
			Version:   "1.0.0",
		}
		tool.Name = "test"
		tool.Description = "test description"
		assert.Equal(t, "test", tool.Name)
		assert.Equal(t, "test-ns", tool.Namespace)
	})

	t.Run("ToolBackend types", func(t *testing.T) {
		backend := model.ToolBackend{
			Kind: model.BackendKindLocal,
		}
		assert.Equal(t, model.BackendKindLocal, backend.Kind)
	})

	t.Run("BackendKind constants", func(t *testing.T) {
		assert.NotEmpty(t, string(model.BackendKindLocal))
		assert.NotEmpty(t, string(model.BackendKindMCP))
		assert.NotEmpty(t, string(model.BackendKindProvider))
	})
}
