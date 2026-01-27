package handlers

import (
	"context"
	"errors"
	"testing"

	"github.com/jonwraymond/metatools-mcp/pkg/metatools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockExecutor struct {
	executeCodeFunc func(ctx context.Context, params ExecuteParams) (ExecuteResult, error)
}

func (m *mockExecutor) ExecuteCode(ctx context.Context, params ExecuteParams) (ExecuteResult, error) {
	if m.executeCodeFunc != nil {
		return m.executeCodeFunc(ctx, params)
	}
	return ExecuteResult{}, nil
}

func TestExecuteCode_Success(t *testing.T) {
	executor := &mockExecutor{
		executeCodeFunc: func(_ context.Context, params ExecuteParams) (ExecuteResult, error) {
			assert.Equal(t, "javascript", params.Language)
			assert.Equal(t, "console.log('hello')", params.Code)
			return ExecuteResult{
				Value:      nil,
				DurationMs: 50,
			}, nil
		},
	}

	handler := NewCodeHandler(executor)
	input := metatools.ExecuteCodeInput{
		Language: "javascript",
		Code:     "console.log('hello')",
	}

	result, err := handler.Handle(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 50, result.DurationMs)
}

func TestExecuteCode_WithValue(t *testing.T) {
	executor := &mockExecutor{
		executeCodeFunc: func(_ context.Context, _ ExecuteParams) (ExecuteResult, error) {
			return ExecuteResult{
				Value:      map[string]any{"result": 42},
				DurationMs: 100,
			}, nil
		},
	}

	handler := NewCodeHandler(executor)
	input := metatools.ExecuteCodeInput{
		Language: "javascript",
		Code:     "return {result: 42}",
	}

	result, err := handler.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, result.Value)
	valueMap, ok := result.Value.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, 42, valueMap["result"])
}

func TestExecuteCode_WithStdout(t *testing.T) {
	executor := &mockExecutor{
		executeCodeFunc: func(_ context.Context, _ ExecuteParams) (ExecuteResult, error) {
			return ExecuteResult{
				Stdout:     "Hello, World!\n",
				DurationMs: 30,
			}, nil
		},
	}

	handler := NewCodeHandler(executor)
	input := metatools.ExecuteCodeInput{
		Language: "javascript",
		Code:     "console.log('Hello, World!')",
	}

	result, err := handler.Handle(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, "Hello, World!\n", result.Stdout)
}

func TestExecuteCode_WithStderr(t *testing.T) {
	executor := &mockExecutor{
		executeCodeFunc: func(_ context.Context, _ ExecuteParams) (ExecuteResult, error) {
			return ExecuteResult{
				Stderr:     "Warning: deprecated function\n",
				DurationMs: 25,
			}, nil
		},
	}

	handler := NewCodeHandler(executor)
	input := metatools.ExecuteCodeInput{
		Language: "javascript",
		Code:     "console.error('Warning: deprecated function')",
	}

	result, err := handler.Handle(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, "Warning: deprecated function\n", result.Stderr)
}

func TestExecuteCode_Timeout(t *testing.T) {
	executor := &mockExecutor{
		executeCodeFunc: func(_ context.Context, params ExecuteParams) (ExecuteResult, error) {
			assert.Equal(t, 5000, params.TimeoutMs)
			return ExecuteResult{
				Value:      "completed",
				DurationMs: 4500,
			}, nil
		},
	}

	handler := NewCodeHandler(executor)
	timeout := 5000
	input := metatools.ExecuteCodeInput{
		Language:  "javascript",
		Code:      "slowOperation()",
		TimeoutMs: &timeout,
	}

	result, err := handler.Handle(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 4500, result.DurationMs)
}

func TestExecuteCode_MaxToolCalls(t *testing.T) {
	executor := &mockExecutor{
		executeCodeFunc: func(_ context.Context, params ExecuteParams) (ExecuteResult, error) {
			assert.Equal(t, 10, params.MaxToolCalls)
			return ExecuteResult{DurationMs: 100}, nil
		},
	}

	handler := NewCodeHandler(executor)
	maxCalls := 10
	input := metatools.ExecuteCodeInput{
		Language:     "javascript",
		Code:         "callTools()",
		MaxToolCalls: &maxCalls,
	}

	result, err := handler.Handle(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, 100, result.DurationMs)
}

func TestExecuteCode_ExecutionError(t *testing.T) {
	executor := &mockExecutor{
		executeCodeFunc: func(_ context.Context, _ ExecuteParams) (ExecuteResult, error) {
			return ExecuteResult{}, errors.New("syntax error: unexpected token")
		},
	}

	handler := NewCodeHandler(executor)
	input := metatools.ExecuteCodeInput{
		Language: "javascript",
		Code:     "invalid code {{{{",
	}

	_, err := handler.Handle(context.Background(), input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "syntax error")
}

func TestExecuteCode_MissingLanguage(t *testing.T) {
	handler := NewCodeHandler(&mockExecutor{})
	input := metatools.ExecuteCodeInput{
		Language: "",
		Code:     "some code",
	}

	_, err := handler.Handle(context.Background(), input)
	assert.Error(t, err)
}

func TestExecuteCode_MissingCode(t *testing.T) {
	handler := NewCodeHandler(&mockExecutor{})
	input := metatools.ExecuteCodeInput{
		Language: "javascript",
		Code:     "",
	}

	_, err := handler.Handle(context.Background(), input)
	assert.Error(t, err)
}
