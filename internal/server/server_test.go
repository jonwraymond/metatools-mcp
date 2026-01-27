package server

import (
	"context"
	"testing"

	"github.com/jonwraymond/metatools-mcp/internal/config"
	"github.com/jonwraymond/metatools-mcp/internal/handlers"
	"github.com/jonwraymond/metatools-mcp/pkg/metatools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock implementations
type mockIndex struct{}

func (m *mockIndex) Search(ctx context.Context, query string, limit int) ([]metatools.ToolSummary, error) {
	return nil, nil
}
func (m *mockIndex) ListNamespaces(ctx context.Context) ([]string, error) {
	return nil, nil
}

type mockStore struct{}

func (m *mockStore) DescribeTool(ctx context.Context, id string, level string) (handlers.ToolDoc, error) {
	return handlers.ToolDoc{}, nil
}
func (m *mockStore) ListExamples(ctx context.Context, id string, max int) ([]metatools.ToolExample, error) {
	return nil, nil
}

type mockRunner struct{}

func (m *mockRunner) Run(ctx context.Context, toolID string, args map[string]any) (handlers.RunResult, error) {
	return handlers.RunResult{}, nil
}
func (m *mockRunner) RunChain(ctx context.Context, steps []handlers.ChainStep) (handlers.RunResult, []handlers.StepResult, error) {
	return handlers.RunResult{}, nil, nil
}

type mockExecutor struct{}

func (m *mockExecutor) ExecuteCode(ctx context.Context, params handlers.ExecuteParams) (handlers.ExecuteResult, error) {
	return handlers.ExecuteResult{}, nil
}

func TestNewServer_RegistersAllTools(t *testing.T) {
	cfg := config.Config{
		Index:    &mockIndex{},
		Docs:     &mockStore{},
		Runner:   &mockRunner{},
		Executor: &mockExecutor{},
	}

	srv, err := New(cfg)
	require.NoError(t, err)
	require.NotNil(t, srv)

	// Verify all 7 tools are registered
	tools := srv.ListTools()
	assert.Len(t, tools, 7)

	// Verify tool names
	toolNames := make(map[string]bool)
	for _, tool := range tools {
		toolNames[tool.Name] = true
	}

	assert.True(t, toolNames["search_tools"])
	assert.True(t, toolNames["list_namespaces"])
	assert.True(t, toolNames["describe_tool"])
	assert.True(t, toolNames["list_tool_examples"])
	assert.True(t, toolNames["run_tool"])
	assert.True(t, toolNames["run_chain"])
	assert.True(t, toolNames["execute_code"])
}

func TestNewServer_ToolsListReturns7Tools(t *testing.T) {
	cfg := config.Config{
		Index:    &mockIndex{},
		Docs:     &mockStore{},
		Runner:   &mockRunner{},
		Executor: &mockExecutor{},
	}

	srv, err := New(cfg)
	require.NoError(t, err)

	tools := srv.ListTools()
	assert.Equal(t, 7, len(tools))
}

func TestNewServer_WithoutExecutor(t *testing.T) {
	cfg := config.Config{
		Index:  &mockIndex{},
		Docs:   &mockStore{},
		Runner: &mockRunner{},
		// No executor
	}

	srv, err := New(cfg)
	require.NoError(t, err)
	require.NotNil(t, srv)

	// Should have 6 tools (no execute_code)
	tools := srv.ListTools()
	assert.Len(t, tools, 6)

	// Verify execute_code is NOT present
	for _, tool := range tools {
		assert.NotEqual(t, "execute_code", tool.Name)
	}
}

func TestServer_DeclaresToolsCapability(t *testing.T) {
	cfg := config.Config{
		Index:    &mockIndex{},
		Docs:     &mockStore{},
		Runner:   &mockRunner{},
		Executor: &mockExecutor{},
	}

	srv, err := New(cfg)
	require.NoError(t, err)

	caps := srv.Capabilities()
	assert.True(t, caps.Tools)
}

func TestNewServer_InvalidConfig(t *testing.T) {
	cfg := config.Config{
		// Missing required fields
	}

	_, err := New(cfg)
	assert.Error(t, err)
}
