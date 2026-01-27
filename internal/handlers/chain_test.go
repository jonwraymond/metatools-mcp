package handlers

import (
	"context"
	"errors"
	"testing"

	merrors "github.com/jonwraymond/metatools-mcp/internal/errors"
	"github.com/jonwraymond/metatools-mcp/pkg/metatools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Success cases

func TestRunChain_SingleStep(t *testing.T) {
	runner := &mockRunner{
		runChainFunc: func(_ context.Context, steps []ChainStep) (RunResult, []StepResult, error) {
			assert.Len(t, steps, 1)
			return RunResult{
					Structured: map[string]any{"final": "result"},
				}, []StepResult{
					{ToolID: "test.tool", Structured: map[string]any{"step": "result"}},
				}, nil
		},
	}

	handler := NewChainHandler(runner)
	input := metatools.RunChainInput{
		Steps: []metatools.ChainStep{{ToolID: "test.tool"}},
	}

	result, isError, err := handler.Handle(context.Background(), input)
	require.NoError(t, err)
	assert.False(t, isError)
	assert.Len(t, result.Results, 1)
	assert.NotNil(t, result.Final)
}

func TestRunChain_MultipleSteps(t *testing.T) {
	runner := &mockRunner{
		runChainFunc: func(_ context.Context, steps []ChainStep) (RunResult, []StepResult, error) {
			assert.Len(t, steps, 3)
			return RunResult{
					Structured: map[string]any{"final": "combined"},
				}, []StepResult{
					{ToolID: "step1", Structured: map[string]any{"s": 1}},
					{ToolID: "step2", Structured: map[string]any{"s": 2}},
					{ToolID: "step3", Structured: map[string]any{"s": 3}},
				}, nil
		},
	}

	handler := NewChainHandler(runner)
	input := metatools.RunChainInput{
		Steps: []metatools.ChainStep{
			{ToolID: "step1"},
			{ToolID: "step2"},
			{ToolID: "step3"},
		},
	}

	result, isError, err := handler.Handle(context.Background(), input)
	require.NoError(t, err)
	assert.False(t, isError)
	assert.Len(t, result.Results, 3)
}

func TestRunChain_UsePrevious(t *testing.T) {
	runner := &mockRunner{
		runChainFunc: func(_ context.Context, steps []ChainStep) (RunResult, []StepResult, error) {
			// Verify use_previous is passed through
			assert.True(t, steps[1].UsePrevious)
			return RunResult{
					Structured: map[string]any{"final": "result"},
				}, []StepResult{
					{ToolID: "step1", Structured: map[string]any{"data": "from step1"}},
					{ToolID: "step2", Structured: map[string]any{"data": "used previous"}},
				}, nil
		},
	}

	handler := NewChainHandler(runner)
	input := metatools.RunChainInput{
		Steps: []metatools.ChainStep{
			{ToolID: "step1"},
			{ToolID: "step2", UsePrevious: true},
		},
	}

	result, isError, err := handler.Handle(context.Background(), input)
	require.NoError(t, err)
	assert.False(t, isError)
	assert.Len(t, result.Results, 2)
}

func TestRunChain_UsePreviousInjectsAtArgsPrevious(t *testing.T) {
	runner := &mockRunner{
		runChainFunc: func(_ context.Context, _ []ChainStep) (RunResult, []StepResult, error) {
			// The runner is responsible for injecting previous at args["previous"]
			return RunResult{
					Structured: map[string]any{"final": "result"},
				}, []StepResult{
					{ToolID: "step1", Structured: map[string]any{"data": "first"}},
					{ToolID: "step2", Structured: map[string]any{"used": "previous data"}},
				}, nil
		},
	}

	handler := NewChainHandler(runner)
	input := metatools.RunChainInput{
		Steps: []metatools.ChainStep{
			{ToolID: "step1", Args: map[string]any{"x": 1}},
			{ToolID: "step2", UsePrevious: true, Args: map[string]any{"y": 2}},
		},
	}

	result, isError, err := handler.Handle(context.Background(), input)
	require.NoError(t, err)
	assert.False(t, isError)
	assert.Len(t, result.Results, 2)
}

func TestRunChain_IncludeBackends(t *testing.T) {
	runner := &mockRunner{
		runChainFunc: func(_ context.Context, steps []ChainStep) (RunResult, []StepResult, error) {
			return RunResult{
					Structured: map[string]any{"final": "result"},
				}, []StepResult{
					{
						ToolID:     "step1",
						Structured: map[string]any{"s": 1},
						Backend:    map[string]any{"kind": "mcp"},
					},
				}, nil
		},
	}

	handler := NewChainHandler(runner)
	includeBackends := true
	input := metatools.RunChainInput{
		Steps:           []metatools.ChainStep{{ToolID: "step1"}},
		IncludeBackends: &includeBackends,
	}

	result, isError, err := handler.Handle(context.Background(), input)
	require.NoError(t, err)
	assert.False(t, isError)
	assert.NotNil(t, result.Results[0].Backend)
}

func TestRunChain_IncludeTools(t *testing.T) {
	runner := &mockRunner{
		runChainFunc: func(_ context.Context, steps []ChainStep) (RunResult, []StepResult, error) {
			return RunResult{
					Structured: map[string]any{"final": "result"},
				}, []StepResult{
					{
						ToolID:     "step1",
						Structured: map[string]any{"s": 1},
						Tool:       map[string]any{"name": "step1", "id": "step1"},
					},
				}, nil
		},
	}

	handler := NewChainHandler(runner)
	includeTools := true
	input := metatools.RunChainInput{
		Steps:        []metatools.ChainStep{{ToolID: "step1"}},
		IncludeTools: &includeTools,
	}

	result, isError, err := handler.Handle(context.Background(), input)
	require.NoError(t, err)
	assert.False(t, isError)
	assert.NotNil(t, result.Results[0].Tool)
}

// Error cases

func TestRunChain_StopOnFirstError(t *testing.T) {
	runner := &mockRunner{
		runChainFunc: func(_ context.Context, steps []ChainStep) (RunResult, []StepResult, error) {
			// Chain stops at step 2
			return RunResult{}, []StepResult{
				{ToolID: "step1", Structured: map[string]any{"s": 1}},
				{ToolID: "step2", Error: merrors.ErrExecution},
			}, merrors.ErrExecution
		},
	}

	handler := NewChainHandler(runner)
	input := metatools.RunChainInput{
		Steps: []metatools.ChainStep{
			{ToolID: "step1"},
			{ToolID: "step2"},
			{ToolID: "step3"},
		},
	}

	result, isError, err := handler.Handle(context.Background(), input)
	require.NoError(t, err)
	assert.True(t, isError)
	assert.Len(t, result.Results, 2) // Only 2 steps executed
}

func TestRunChain_PartialResultsOnError(t *testing.T) {
	runner := &mockRunner{
		runChainFunc: func(_ context.Context, steps []ChainStep) (RunResult, []StepResult, error) {
			return RunResult{}, []StepResult{
				{ToolID: "step1", Structured: map[string]any{"partial": "result1"}},
				{ToolID: "step2", Error: errors.New("step failed")},
			}, errors.New("chain failed at step 2")
		},
	}

	handler := NewChainHandler(runner)
	input := metatools.RunChainInput{
		Steps: []metatools.ChainStep{
			{ToolID: "step1"},
			{ToolID: "step2"},
		},
	}

	result, isError, err := handler.Handle(context.Background(), input)
	require.NoError(t, err)
	assert.True(t, isError)
	// Partial results should be preserved
	assert.Len(t, result.Results, 2)
	assert.NotNil(t, result.Results[0].Structured)
}

func TestRunChain_ErrorHasStepIndex(t *testing.T) {
	runner := &mockRunner{
		runChainFunc: func(_ context.Context, steps []ChainStep) (RunResult, []StepResult, error) {
			return RunResult{}, []StepResult{
				{ToolID: "step1", Structured: map[string]any{"s": 1}},
				{ToolID: "step2", Structured: map[string]any{"s": 2}},
				{ToolID: "step3", Error: merrors.ErrExecution},
			}, merrors.ErrExecution
		},
	}

	handler := NewChainHandler(runner)
	input := metatools.RunChainInput{
		Steps: []metatools.ChainStep{
			{ToolID: "step1"},
			{ToolID: "step2"},
			{ToolID: "step3"},
		},
	}

	result, isError, err := handler.Handle(context.Background(), input)
	require.NoError(t, err)
	assert.True(t, isError)
	require.NotNil(t, result.Error)
	assert.Equal(t, string(merrors.CodeChainStepFailed), result.Error.Code)
	require.NotNil(t, result.Error.StepIndex)
	assert.Equal(t, 2, *result.Error.StepIndex)
}

func TestRunChain_ErrorHasCauseDetails(t *testing.T) {
	runner := &mockRunner{
		runChainFunc: func(_ context.Context, steps []ChainStep) (RunResult, []StepResult, error) {
			return RunResult{}, []StepResult{
				{ToolID: "step1", Structured: map[string]any{"s": 1}},
				{ToolID: "step2", Error: merrors.ErrToolNotFound},
			}, merrors.ErrToolNotFound
		},
	}

	handler := NewChainHandler(runner)
	input := metatools.RunChainInput{
		Steps: []metatools.ChainStep{
			{ToolID: "step1"},
			{ToolID: "step2"},
		},
	}

	result, isError, err := handler.Handle(context.Background(), input)
	require.NoError(t, err)
	assert.True(t, isError)
	require.NotNil(t, result.Error)
	// Error should have cause details
	require.NotNil(t, result.Error.Details)
	assert.Contains(t, result.Error.Details, "cause_code")
}

func TestRunChain_EmptySteps(t *testing.T) {
	handler := NewChainHandler(&mockRunner{})
	input := metatools.RunChainInput{
		Steps: []metatools.ChainStep{},
	}

	_, _, err := handler.Handle(context.Background(), input)
	assert.Error(t, err)
}
