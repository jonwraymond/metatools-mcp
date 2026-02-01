// Package handlers contains MCP tool handlers.
package handlers

import (
	"context"
	"testing"
	"time"

	merrors "github.com/jonwraymond/metatools-mcp/internal/errors"
	"github.com/jonwraymond/metatools-mcp/pkg/metatools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockIndex, mockRunner, and mockStore are defined in search_test.go, run_test.go, and describe_test.go respectively

func TestRunTool_ContextCancellationDuringExecution(t *testing.T) {
	runner := &mockRunner{
		runFunc: func(ctx context.Context, _ string, _ map[string]any) (RunResult, error) {
			select {
			case <-ctx.Done():
				return RunResult{}, ctx.Err()
			case <-time.After(200 * time.Millisecond):
				return RunResult{Structured: map[string]any{"result": "completed"}}, nil
			}
		},
	}

	handler := NewRunHandler(runner)
	input := metatools.RunToolInput{ToolID: "test.tool"}

	ctx, cancel := context.WithCancel(context.Background())

	// Start handler
	resultChan := make(chan struct {
		res   *metatools.RunToolOutput
		isErr bool
		err   error
	})

	go func() {
		res, isErr, err := handler.Handle(ctx, input)
		resultChan <- struct {
			res   *metatools.RunToolOutput
			isErr bool
			err   error
		}{res, isErr, err}
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case r := <-resultChan:
		require.NoError(t, r.err)
		assert.True(t, r.isErr)
		require.NotNil(t, r.res.Error)
		// Expect CodeCancelled for context cancellation
		assert.Equal(t, string(merrors.CodeCancelled), r.res.Error.Code)
	case <-time.After(1 * time.Second):
		t.Fatal("Handler did not return after cancellation")
	}
}

func TestSearchHandler_ContextCancelled(t *testing.T) {
	index := &mockIndex{
		searchFunc: func(_ context.Context, _ string, _ int, _ string) ([]metatools.ToolSummary, string, error) {
			return []metatools.ToolSummary{{Name: "test"}}, "", nil
		},
	}

	handler := NewSearchHandler(index)
	input := metatools.SearchToolsInput{Query: "test"}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := handler.Handle(ctx, input)
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestNamespacesHandler_ContextCancelled(t *testing.T) {
	index := &mockIndex{
		listNamespacesFunc: func(_ context.Context, _ int, _ string) ([]string, string, error) {
			return []string{"ns1"}, "", nil
		},
	}

	handler := NewNamespacesHandler(index)
	input := metatools.ListNamespacesInput{}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := handler.Handle(ctx, input)
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestDescribeHandler_ContextCancelled(t *testing.T) {
	store := &mockStore{
		describeToolFunc: func(_ context.Context, _ string, _ string) (ToolDoc, error) {
			return ToolDoc{}, nil
		},
	}

	handler := NewDescribeHandler(store)
	input := metatools.DescribeToolInput{ToolID: "test.tool", DetailLevel: "summary"}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := handler.Handle(ctx, input)
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}
