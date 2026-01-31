// Package migration contains tests for PRD-180 migration to consolidated repositories.
package migration

import (
	"testing"

	"github.com/jonwraymond/toolcode"
	"github.com/jonwraymond/tooldocs"
	"github.com/jonwraymond/toolindex"
	"github.com/jonwraymond/toolmodel"
	"github.com/jonwraymond/toolrun"
	"github.com/jonwraymond/toolsearch"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBaselineCompilation verifies all standalone packages compile and are accessible.
func TestBaselineCompilation(t *testing.T) {
	// This test simply verifies the imports compile
	t.Run("toolmodel compiles", func(t *testing.T) {
		var _ toolmodel.Tool
	})
	t.Run("toolindex compiles", func(t *testing.T) {
		var _ toolindex.Index
	})
	t.Run("tooldocs compiles", func(t *testing.T) {
		var _ tooldocs.Store
	})
	t.Run("toolsearch compiles", func(t *testing.T) {
		var _ toolsearch.BM25Config
	})
	t.Run("toolrun compiles", func(t *testing.T) {
		var _ toolrun.Runner
	})
	t.Run("toolcode compiles", func(t *testing.T) {
		var _ toolcode.Executor
	})
}

// TestCurrentIndexCreation verifies toolindex.NewInMemoryIndex works.
func TestCurrentIndexCreation(t *testing.T) {
	idx := toolindex.NewInMemoryIndex()
	require.NotNil(t, idx, "InMemoryIndex should be created")

	// Create a tool using the embedded mcp.Tool pattern
	tool := toolmodel.Tool{
		Namespace: "test",
	}
	tool.Name = "test-tool"
	tool.Description = "A test tool"
	tool.InputSchema = map[string]any{"type": "object"}

	backend := toolmodel.ToolBackend{
		Kind:  toolmodel.BackendKindLocal,
		Local: &toolmodel.LocalBackend{Name: "test-handler"},
	}

	err := idx.RegisterTool(tool, backend)
	assert.NoError(t, err, "Registering tool should succeed")

	retrieved, _, err := idx.GetTool("test:test-tool")
	assert.NoError(t, err, "Getting tool should succeed")
	assert.Equal(t, "test-tool", retrieved.Name)
}

// TestCurrentSearcherCreation verifies toolsearch.NewBM25Searcher works.
func TestCurrentSearcherCreation(t *testing.T) {
	// Create BM25 searcher with default config
	config := toolsearch.BM25Config{}
	searcher := toolsearch.NewBM25Searcher(config)
	require.NotNil(t, searcher, "BM25Searcher should be created")

	// Create tool
	tool := toolmodel.Tool{
		Namespace: "fs",
	}
	tool.Name = "file-reader"
	tool.Description = "Reads files from disk"
	tool.InputSchema = map[string]any{"type": "object"}

	backend := toolmodel.ToolBackend{
		Kind:  toolmodel.BackendKindLocal,
		Local: &toolmodel.LocalBackend{Name: "file-reader-handler"},
	}

	// Configure index to use the BM25 searcher
	idxWithSearcher := toolindex.NewInMemoryIndex(toolindex.IndexOptions{
		Searcher: searcher,
	})
	err := idxWithSearcher.RegisterTool(tool, backend)
	require.NoError(t, err)

	// Verify search works
	results, err := idxWithSearcher.Search("file", 10)
	assert.NoError(t, err, "Search should succeed")
	assert.Len(t, results, 1, "Should find one result")
}

// TestCurrentDocsStoreCreation verifies tooldocs.NewInMemoryStore works.
func TestCurrentDocsStoreCreation(t *testing.T) {
	idx := toolindex.NewInMemoryIndex()

	// Register a tool first
	tool := toolmodel.Tool{
		Namespace: "test",
	}
	tool.Name = "test-tool"
	tool.Description = "A test tool"
	tool.InputSchema = map[string]any{"type": "object"}

	backend := toolmodel.ToolBackend{
		Kind:  toolmodel.BackendKindLocal,
		Local: &toolmodel.LocalBackend{Name: "test-handler"},
	}
	_ = idx.RegisterTool(tool, backend)

	// Create store with index
	store := tooldocs.NewInMemoryStore(tooldocs.StoreOptions{
		Index: idx,
	})
	require.NotNil(t, store, "InMemoryStore should be created")

	// Verify we can describe a tool (using the index)
	doc, err := store.DescribeTool("test:test-tool", tooldocs.DetailSummary)
	assert.NoError(t, err, "DescribeTool should succeed")
	// ToolDoc has Tool field, not ID - check summary is populated
	assert.NotEmpty(t, doc.Summary, "Summary should be populated")
}

// TestCurrentRunnerCreation verifies toolrun.NewRunner works.
func TestCurrentRunnerCreation(t *testing.T) {
	runner := toolrun.NewRunner()
	require.NotNil(t, runner, "Runner should be created")
}

// TestCurrentCodeExecutorCreation verifies toolcode.NewDefaultExecutor works.
func TestCurrentCodeExecutorCreation(t *testing.T) {
	// toolcode.NewDefaultExecutor requires a Config with Runner
	// For this baseline test, we just verify the type exists
	var _ toolcode.Executor
	var _ toolcode.Config
}

// TestToolModelTypes verifies core toolmodel types exist and work.
func TestToolModelTypes(t *testing.T) {
	t.Run("Tool struct", func(t *testing.T) {
		tool := toolmodel.Tool{
			Namespace: "test-ns",
			Version:   "1.0.0",
		}
		tool.Name = "test"
		tool.Description = "test description"
		assert.Equal(t, "test", tool.Name)
		assert.Equal(t, "test-ns", tool.Namespace)
	})

	t.Run("ToolBackend types", func(t *testing.T) {
		backend := toolmodel.ToolBackend{
			Kind: toolmodel.BackendKindLocal,
		}
		assert.Equal(t, toolmodel.BackendKindLocal, backend.Kind)
	})

	t.Run("BackendKind constants", func(t *testing.T) {
		assert.NotEmpty(t, string(toolmodel.BackendKindLocal))
		assert.NotEmpty(t, string(toolmodel.BackendKindMCP))
		assert.NotEmpty(t, string(toolmodel.BackendKindProvider))
	})
}
