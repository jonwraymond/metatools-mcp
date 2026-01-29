package builtin

import (
	"context"
	"errors"
	"testing"

	"github.com/jonwraymond/metatools-mcp/internal/handlers"
	"github.com/jonwraymond/metatools-mcp/pkg/metatools"
	"github.com/stretchr/testify/require"
)

type errorRunner struct{}

func (e *errorRunner) Run(_ context.Context, _ string, _ map[string]any) (handlers.RunResult, error) {
	return handlers.RunResult{}, errors.New("boom")
}

func (e *errorRunner) RunChain(_ context.Context, _ []handlers.ChainStep) (handlers.RunResult, []handlers.StepResult, error) {
	return handlers.RunResult{}, nil, errors.New("boom")
}

func TestRunToolProvider_IsError(t *testing.T) {
	handler := handlers.NewRunHandler(&errorRunner{})
	provider := NewRunToolProvider(handler, true)

	res, out, err := provider.Handle(context.Background(), nil, map[string]any{
		"tool_id": "test.tool",
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.True(t, res.IsError)

	output, ok := out.(metatools.RunToolOutput)
	require.True(t, ok)
	require.NotNil(t, output.Error)
}
