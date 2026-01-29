package bootstrap

import (
	"testing"

	"github.com/jonwraymond/metatools-mcp/internal/config"
	"github.com/jonwraymond/toolmodel"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewIndexFromConfig_CreatesWorkingIndex(t *testing.T) {
	cfg := config.EnvConfig{
		Search: config.SearchConfig{
			Strategy: "lexical",
		},
	}

	idx, err := NewIndexFromConfig(cfg)
	require.NoError(t, err)
	require.NotNil(t, idx)

	// Verify the index works by registering and searching for a tool
	tool := toolmodel.Tool{
		Tool: mcp.Tool{
			Name:        "test_tool",
			Description: "A test tool for verification",
			InputSchema: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
		},
		Namespace: "testing",
	}

	backend := toolmodel.ToolBackend{
		Kind: toolmodel.BackendKindLocal,
		Local: &toolmodel.LocalBackend{
			Name: "test_handler",
		},
	}

	err = idx.RegisterTool(tool, backend)
	require.NoError(t, err)

	// Search should return the tool
	results, err := idx.Search("test", 10)
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "test_tool", results[0].Name)
}

func TestNewIndexFromConfig_DefaultStrategyWorks(t *testing.T) {
	// Test with defaults (should behave like lexical)
	cfg := config.EnvConfig{}
	cfg.Search.Strategy = "lexical" // explicit default

	idx, err := NewIndexFromConfig(cfg)
	require.NoError(t, err)
	require.NotNil(t, idx)

	tool := toolmodel.Tool{
		Tool: mcp.Tool{
			Name:        "another_tool",
			Description: "Another test tool",
			InputSchema: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
		},
		Namespace: "demo",
	}

	backend := toolmodel.ToolBackend{
		Kind: toolmodel.BackendKindLocal,
		Local: &toolmodel.LocalBackend{
			Name: "test_handler",
		},
	}

	err = idx.RegisterTool(tool, backend)
	require.NoError(t, err)

	results, err := idx.Search("demo", 10)
	require.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestNewIndexFromAppConfig_CreatesWorkingIndex(t *testing.T) {
	cfg := config.DefaultAppConfig()
	cfg.Search.Strategy = "lexical"

	idx, err := NewIndexFromAppConfig(cfg)
	require.NoError(t, err)
	require.NotNil(t, idx)

	tool := toolmodel.Tool{
		Tool: mcp.Tool{
			Name:        "app_tool",
			Description: "Tool registered via app config",
			InputSchema: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
		},
		Namespace: "app",
	}

	backend := toolmodel.ToolBackend{
		Kind: toolmodel.BackendKindLocal,
		Local: &toolmodel.LocalBackend{
			Name: "test_handler",
		},
	}

	err = idx.RegisterTool(tool, backend)
	require.NoError(t, err)

	results, err := idx.Search("app", 10)
	require.NoError(t, err)
	assert.Len(t, results, 1)
}
