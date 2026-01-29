package handlers

import (
	"context"
	"testing"

	"github.com/jonwraymond/metatools-mcp/pkg/metatools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockProgressRunner struct {
	runWithProgressFunc      func(context.Context, string, map[string]any, func(ProgressEvent)) (RunResult, error)
	runChainWithProgressFunc func(context.Context, []ChainStep, func(ProgressEvent)) (RunResult, []StepResult, error)
}

func (m *mockProgressRunner) Run(context.Context, string, map[string]any) (RunResult, error) {
	return RunResult{}, nil
}

func (m *mockProgressRunner) RunChain(context.Context, []ChainStep) (RunResult, []StepResult, error) {
	return RunResult{}, nil, nil
}

func (m *mockProgressRunner) RunWithProgress(ctx context.Context, toolID string, args map[string]any, onProgress func(ProgressEvent)) (RunResult, error) {
	if m.runWithProgressFunc != nil {
		return m.runWithProgressFunc(ctx, toolID, args, onProgress)
	}
	return RunResult{}, nil
}

func (m *mockProgressRunner) RunChainWithProgress(ctx context.Context, steps []ChainStep, onProgress func(ProgressEvent)) (RunResult, []StepResult, error) {
	if m.runChainWithProgressFunc != nil {
		return m.runChainWithProgressFunc(ctx, steps, onProgress)
	}
	return RunResult{}, nil, nil
}

func TestRunHandler_ProgressFallback(t *testing.T) {
	runner := &mockRunner{
		runFunc: func(_ context.Context, _ string, _ map[string]any) (RunResult, error) {
			return RunResult{Structured: "ok"}, nil
		},
	}
	handler := NewRunHandler(runner)
	input := metatools.RunToolInput{ToolID: "test.tool"}

	var events []ProgressEvent
	result, isError, err := handler.HandleWithProgress(context.Background(), input, func(ev ProgressEvent) {
		events = append(events, ev)
	})
	require.NoError(t, err)
	assert.False(t, isError)
	assert.Equal(t, "ok", result.Structured)
	assert.Len(t, events, 2)
}

func TestChainHandler_ProgressFallback(t *testing.T) {
	runner := &mockRunner{
		runChainFunc: func(_ context.Context, _ []ChainStep) (RunResult, []StepResult, error) {
			return RunResult{Structured: "ok"}, []StepResult{}, nil
		},
	}
	handler := NewChainHandler(runner)
	input := metatools.RunChainInput{Steps: []metatools.ChainStep{{ToolID: "test.tool"}}}

	var events []ProgressEvent
	result, isError, err := handler.HandleWithProgress(context.Background(), input, func(ev ProgressEvent) {
		events = append(events, ev)
	})
	require.NoError(t, err)
	assert.False(t, isError)
	assert.Equal(t, "ok", result.Final)
	assert.Len(t, events, 2)
}

func TestRunHandler_ProgressRunner(t *testing.T) {
	runner := &mockProgressRunner{
		runWithProgressFunc: func(_ context.Context, _ string, _ map[string]any, onProgress func(ProgressEvent)) (RunResult, error) {
			onProgress(ProgressEvent{Progress: 0, Total: 1, Message: "start"})
			onProgress(ProgressEvent{Progress: 1, Total: 1, Message: "done"})
			return RunResult{Structured: "ok"}, nil
		},
	}
	handler := NewRunHandler(runner)
	input := metatools.RunToolInput{ToolID: "test.tool"}

	var events []ProgressEvent
	result, isError, err := handler.HandleWithProgress(context.Background(), input, func(ev ProgressEvent) {
		events = append(events, ev)
	})
	require.NoError(t, err)
	assert.False(t, isError)
	assert.Equal(t, "ok", result.Structured)
	assert.Len(t, events, 2)
	assert.Equal(t, "start", events[0].Message)
	assert.Equal(t, "done", events[1].Message)
}
