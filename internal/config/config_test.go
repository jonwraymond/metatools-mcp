package config

import (
	"context"
	"testing"

	"github.com/jonwraymond/metatools-mcp/internal/handlers"
	"github.com/jonwraymond/metatools-mcp/pkg/metatools"
	"github.com/stretchr/testify/assert"
)

// Mock implementations for testing
type mockIndex struct{}

func (m *mockIndex) Search(_ context.Context, _ string, _ int) ([]metatools.ToolSummary, error) {
	return nil, nil
}
func (m *mockIndex) ListNamespaces(_ context.Context) ([]string, error) {
	return nil, nil
}

type mockStore struct{}

func (m *mockStore) DescribeTool(_ context.Context, _ string, _ string) (handlers.ToolDoc, error) {
	return handlers.ToolDoc{}, nil
}
func (m *mockStore) ListExamples(_ context.Context, _ string, _ int) ([]metatools.ToolExample, error) {
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

func TestConfig_Validate(t *testing.T) {
	cfg := Config{
		Index:  &mockIndex{},
		Docs:   &mockStore{},
		Runner: &mockRunner{},
	}

	err := cfg.Validate()
	assert.NoError(t, err)
}

func TestConfig_RequiresIndex(t *testing.T) {
	cfg := Config{
		Docs:   &mockStore{},
		Runner: &mockRunner{},
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "index")
}

func TestConfig_RequiresDocs(t *testing.T) {
	cfg := Config{
		Index:  &mockIndex{},
		Runner: &mockRunner{},
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "docs")
}

func TestConfig_RequiresRunner(t *testing.T) {
	cfg := Config{
		Index: &mockIndex{},
		Docs:  &mockStore{},
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "runner")
}

func TestConfig_ExecutorOptional(t *testing.T) {
	cfg := Config{
		Index:  &mockIndex{},
		Docs:   &mockStore{},
		Runner: &mockRunner{},
		// Executor is nil - should be OK
	}

	err := cfg.Validate()
	assert.NoError(t, err)
}

func TestConfig_WithExecutor(t *testing.T) {
	cfg := Config{
		Index:    &mockIndex{},
		Docs:     &mockStore{},
		Runner:   &mockRunner{},
		Executor: &mockExecutor{},
	}

	err := cfg.Validate()
	assert.NoError(t, err)
}
