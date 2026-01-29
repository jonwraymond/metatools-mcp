package handlers

import (
	"context"
	"testing"

	merrors "github.com/jonwraymond/metatools-mcp/internal/errors"
	"github.com/jonwraymond/metatools-mcp/pkg/metatools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRunner struct {
	runFunc      func(_ context.Context, _ string, _ map[string]any) (RunResult, error)
	runChainFunc func(_ context.Context, steps []ChainStep) (RunResult, []StepResult, error)
}

func (m *mockRunner) Run(ctx context.Context, toolID string, args map[string]any) (RunResult, error) {
	if m.runFunc != nil {
		return m.runFunc(ctx, toolID, args)
	}
	return RunResult{}, nil
}

func (m *mockRunner) RunChain(ctx context.Context, steps []ChainStep) (RunResult, []StepResult, error) {
	if m.runChainFunc != nil {
		return m.runChainFunc(ctx, steps)
	}
	return RunResult{}, []StepResult{}, nil
}

// Success cases

func TestRunTool_Success(t *testing.T) {
	runner := &mockRunner{
		runFunc: func(_ context.Context, _ string, _ map[string]any) (RunResult, error) {
			return RunResult{
				Structured: map[string]any{"result": "success"},
				DurationMs: 100,
			}, nil
		},
	}

	handler := NewRunHandler(runner)
	input := metatools.RunToolInput{
		ToolID: "test.tool",
		Args:   map[string]any{"param": "value"},
	}

	result, isError, err := handler.Handle(context.Background(), input)
	require.NoError(t, err)
	assert.False(t, isError)
	require.NotNil(t, result.Structured)
}

func TestRunTool_SuccessWithStructured(t *testing.T) {
	runner := &mockRunner{
		runFunc: func(_ context.Context, _ string, _ map[string]any) (RunResult, error) {
			return RunResult{
				Structured: map[string]any{"key": "value", "count": 42},
			}, nil
		},
	}

	handler := NewRunHandler(runner)
	input := metatools.RunToolInput{ToolID: "test.tool"}

	result, isError, err := handler.Handle(context.Background(), input)
	require.NoError(t, err)
	assert.False(t, isError)

	structured, ok := result.Structured.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "value", structured["key"])
	assert.Equal(t, 42, structured["count"])
}

func TestRunTool_IncludeTool(t *testing.T) {
	runner := &mockRunner{
		runFunc: func(_ context.Context, _ string, _ map[string]any) (RunResult, error) {
			return RunResult{
				Structured: map[string]any{"result": "ok"},
				Tool:       map[string]any{"name": "tool", "id": "test.tool"},
			}, nil
		},
	}

	handler := NewRunHandler(runner)
	input := metatools.RunToolInput{
		ToolID:      "test.tool",
		IncludeTool: true,
	}

	result, isError, err := handler.Handle(context.Background(), input)
	require.NoError(t, err)
	assert.False(t, isError)
	assert.NotNil(t, result.Tool)
}

func TestRunTool_IncludeBackend(t *testing.T) {
	runner := &mockRunner{
		runFunc: func(_ context.Context, _ string, _ map[string]any) (RunResult, error) {
			return RunResult{
				Structured: map[string]any{"result": "ok"},
				Backend:    map[string]any{"kind": "mcp", "serverName": "test-server"},
			}, nil
		},
	}

	handler := NewRunHandler(runner)
	input := metatools.RunToolInput{
		ToolID:         "test.tool",
		IncludeBackend: true,
	}

	result, isError, err := handler.Handle(context.Background(), input)
	require.NoError(t, err)
	assert.False(t, isError)
	assert.NotNil(t, result.Backend)
}

func TestRunTool_IncludeMCPResult(t *testing.T) {
	runner := &mockRunner{
		runFunc: func(_ context.Context, _ string, _ map[string]any) (RunResult, error) {
			return RunResult{
				Structured: map[string]any{"result": "ok"},
				MCPResult:  map[string]any{"structuredContent": map[string]any{"result": "ok"}},
			}, nil
		},
	}

	handler := NewRunHandler(runner)
	input := metatools.RunToolInput{
		ToolID:           "test.tool",
		IncludeMCPResult: true,
	}

	// MCPResult would be populated by the server layer, not the handler
	result, isError, err := handler.Handle(context.Background(), input)
	require.NoError(t, err)
	assert.False(t, isError)
	assert.NotNil(t, result.MCPResult)
}

// Error cases - tool errors (isError: true)

func TestRunTool_ToolNotFound_ReturnsToolError(t *testing.T) {
	runner := &mockRunner{
		runFunc: func(_ context.Context, _ string, _ map[string]any) (RunResult, error) {
			return RunResult{}, merrors.ErrToolNotFound
		},
	}

	handler := NewRunHandler(runner)
	input := metatools.RunToolInput{ToolID: "nonexistent.tool"}

	result, isError, err := handler.Handle(context.Background(), input)
	require.NoError(t, err) // No protocol error
	assert.True(t, isError)
	require.NotNil(t, result.Error)
	assert.Equal(t, string(merrors.CodeToolNotFound), result.Error.Code)
}

func TestRunTool_NoBackends_ReturnsToolError(t *testing.T) {
	runner := &mockRunner{
		runFunc: func(_ context.Context, _ string, _ map[string]any) (RunResult, error) {
			return RunResult{}, merrors.ErrNoBackends
		},
	}

	handler := NewRunHandler(runner)
	input := metatools.RunToolInput{ToolID: "test.tool"}

	result, isError, err := handler.Handle(context.Background(), input)
	require.NoError(t, err)
	assert.True(t, isError)
	require.NotNil(t, result.Error)
	assert.Equal(t, string(merrors.CodeNoBackends), result.Error.Code)
}

func TestRunTool_ValidationInput_ReturnsToolError(t *testing.T) {
	runner := &mockRunner{
		runFunc: func(_ context.Context, _ string, _ map[string]any) (RunResult, error) {
			return RunResult{}, merrors.ErrValidationInput
		},
	}

	handler := NewRunHandler(runner)
	input := metatools.RunToolInput{ToolID: "test.tool", Args: map[string]any{"bad": "input"}}

	result, isError, err := handler.Handle(context.Background(), input)
	require.NoError(t, err)
	assert.True(t, isError)
	require.NotNil(t, result.Error)
	assert.Equal(t, string(merrors.CodeValidationInput), result.Error.Code)
}

func TestRunTool_ValidationOutput_ReturnsToolError(t *testing.T) {
	runner := &mockRunner{
		runFunc: func(_ context.Context, _ string, _ map[string]any) (RunResult, error) {
			return RunResult{}, merrors.ErrValidationOutput
		},
	}

	handler := NewRunHandler(runner)
	input := metatools.RunToolInput{ToolID: "test.tool"}

	result, isError, err := handler.Handle(context.Background(), input)
	require.NoError(t, err)
	assert.True(t, isError)
	require.NotNil(t, result.Error)
	assert.Equal(t, string(merrors.CodeValidationOutput), result.Error.Code)
}

func TestRunTool_ExecutionFailed_ReturnsToolError(t *testing.T) {
	runner := &mockRunner{
		runFunc: func(_ context.Context, _ string, _ map[string]any) (RunResult, error) {
			return RunResult{}, merrors.ErrExecution
		},
	}

	handler := NewRunHandler(runner)
	input := metatools.RunToolInput{ToolID: "test.tool"}

	result, isError, err := handler.Handle(context.Background(), input)
	require.NoError(t, err)
	assert.True(t, isError)
	require.NotNil(t, result.Error)
	assert.Equal(t, string(merrors.CodeExecutionFailed), result.Error.Code)
	assert.True(t, result.Error.Retryable)
}

func TestRunTool_ContextCancelled_ReturnsToolError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	runner := &mockRunner{
		runFunc: func(ctx context.Context, _ string, _ map[string]any) (RunResult, error) {
			return RunResult{}, ctx.Err()
		},
	}

	handler := NewRunHandler(runner)
	input := metatools.RunToolInput{ToolID: "test.tool"}

	result, isError, err := handler.Handle(ctx, input)
	require.NoError(t, err)
	assert.True(t, isError)
	require.NotNil(t, result.Error)
	assert.Equal(t, string(merrors.CodeInternal), result.Error.Code)
}

// Backend override

func TestRunTool_BackendOverride_Valid(t *testing.T) {
	called := false
	runner := &mockRunner{
		runFunc: func(_ context.Context, _ string, _ map[string]any) (RunResult, error) {
			called = true
			return RunResult{Structured: map[string]any{"result": "ok"}}, nil
		},
	}

	handler := NewRunHandler(runner)
	input := metatools.RunToolInput{
		ToolID: "test.tool",
		BackendOverride: &metatools.BackendOverride{
			Kind:       "mcp",
			ServerName: "override-server",
		},
	}

	result, isError, err := handler.Handle(context.Background(), input)
	require.NoError(t, err)
	assert.False(t, called)
	assert.True(t, isError)
	require.NotNil(t, result.Error)
	assert.Equal(t, string(merrors.CodeBackendOverrideInvalid), result.Error.Code)
}

func TestRunTool_BackendOverride_Invalid(t *testing.T) {
	called := false
	runner := &mockRunner{
		runFunc: func(_ context.Context, _ string, _ map[string]any) (RunResult, error) {
			called = true
			return RunResult{}, nil
		},
	}

	handler := NewRunHandler(runner)
	input := metatools.RunToolInput{
		ToolID: "test.tool",
		BackendOverride: &metatools.BackendOverride{
			Kind: "invalid",
		},
	}

	result, isError, err := handler.Handle(context.Background(), input)
	require.NoError(t, err)
	assert.False(t, called)
	assert.True(t, isError)
	require.NotNil(t, result.Error)
	assert.Equal(t, string(merrors.CodeBackendOverrideInvalid), result.Error.Code)
}

func TestRunTool_BackendOverride_NoMatch(t *testing.T) {
	called := false
	runner := &mockRunner{
		runFunc: func(_ context.Context, _ string, _ map[string]any) (RunResult, error) {
			called = true
			return RunResult{}, nil
		},
	}

	handler := NewRunHandler(runner)
	input := metatools.RunToolInput{
		ToolID: "test.tool",
		BackendOverride: &metatools.BackendOverride{
			Kind:       "mcp",
			ServerName: "nonexistent-server",
		},
	}

	result, isError, err := handler.Handle(context.Background(), input)
	require.NoError(t, err)
	assert.False(t, called)
	assert.True(t, isError)
	require.NotNil(t, result.Error)
	assert.Equal(t, string(merrors.CodeBackendOverrideInvalid), result.Error.Code)
}

// Streaming

func TestRunTool_Stream_NotSupported(t *testing.T) {
	called := false
	runner := &mockRunner{
		runFunc: func(_ context.Context, _ string, _ map[string]any) (RunResult, error) {
			called = true
			return RunResult{}, nil
		},
	}

	handler := NewRunHandler(runner)
	input := metatools.RunToolInput{
		ToolID: "test.tool",
		Stream: true,
	}

	result, isError, err := handler.Handle(context.Background(), input)
	require.NoError(t, err)
	assert.False(t, called)
	assert.True(t, isError)
	require.NotNil(t, result.Error)
	assert.Equal(t, string(merrors.CodeStreamNotSupported), result.Error.Code)
}

func TestRunTool_MissingToolID(t *testing.T) {
	handler := NewRunHandler(&mockRunner{})
	input := metatools.RunToolInput{ToolID: ""}

	_, _, err := handler.Handle(context.Background(), input)
	assert.Error(t, err)
}
