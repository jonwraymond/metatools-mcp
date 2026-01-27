package handlers

import (
	"context"
	"errors"
	"testing"

	"github.com/jonwraymond/metatools-mcp/pkg/metatools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListToolExamples_Default(t *testing.T) {
	store := &mockStore{
		listExamplesFunc: func(_ context.Context, id string, _ int) ([]metatools.ToolExample, error) {
			assert.Equal(t, "test.tool", id)
			return []metatools.ToolExample{
				{Title: "Example 1", Description: "First", Args: map[string]any{}},
				{Title: "Example 2", Description: "Second", Args: map[string]any{}},
			}, nil
		},
	}

	handler := NewExamplesHandler(store)
	input := metatools.ListToolExamplesInput{
		ToolID: "test.tool",
	}

	result, err := handler.Handle(context.Background(), input)
	require.NoError(t, err)
	assert.Len(t, result.Examples, 2)
}

func TestListToolExamples_WithMax(t *testing.T) {
	store := &mockStore{
		listExamplesFunc: func(_ context.Context, _ string, maxExamples int) ([]metatools.ToolExample, error) {
			assert.Equal(t, 3, maxExamples)
			return []metatools.ToolExample{
				{Title: "Example 1", Description: "First", Args: map[string]any{}},
				{Title: "Example 2", Description: "Second", Args: map[string]any{}},
				{Title: "Example 3", Description: "Third", Args: map[string]any{}},
			}, nil
		},
	}

	handler := NewExamplesHandler(store)
	maxVal := 3
	input := metatools.ListToolExamplesInput{
		ToolID: "test.tool",
		Max:    &maxVal,
	}

	result, err := handler.Handle(context.Background(), input)
	require.NoError(t, err)
	assert.Len(t, result.Examples, 3)
}

func TestListToolExamples_CappedByStoreMax(t *testing.T) {
	// Store has its own max cap (e.g., 5)
	store := &mockStore{
		listExamplesFunc: func(_ context.Context, _ string, maxExamples int) ([]metatools.ToolExample, error) {
			// If max is 10 but store caps at 5, return 5
			examples := []metatools.ToolExample{
				{Title: "Ex1", Description: "D1", Args: map[string]any{}},
				{Title: "Ex2", Description: "D2", Args: map[string]any{}},
				{Title: "Ex3", Description: "D3", Args: map[string]any{}},
				{Title: "Ex4", Description: "D4", Args: map[string]any{}},
				{Title: "Ex5", Description: "D5", Args: map[string]any{}},
			}
			if maxExamples < len(examples) {
				return examples[:maxExamples], nil
			}
			return examples, nil
		},
	}

	handler := NewExamplesHandler(store)
	maxVal := 10
	input := metatools.ListToolExamplesInput{
		ToolID: "test.tool",
		Max:    &maxVal,
	}

	result, err := handler.Handle(context.Background(), input)
	require.NoError(t, err)
	assert.Len(t, result.Examples, 5) // Capped by store's max
}

func TestListToolExamples_NotFound(t *testing.T) {
	store := &mockStore{
		listExamplesFunc: func(_ context.Context, _ string, _ int) ([]metatools.ToolExample, error) {
			return nil, errors.New("tool not found")
		},
	}

	handler := NewExamplesHandler(store)
	input := metatools.ListToolExamplesInput{
		ToolID: "nonexistent.tool",
	}

	_, err := handler.Handle(context.Background(), input)
	assert.Error(t, err)
}

func TestListToolExamples_EmptyExamples(t *testing.T) {
	store := &mockStore{
		listExamplesFunc: func(_ context.Context, _ string, _ int) ([]metatools.ToolExample, error) {
			return []metatools.ToolExample{}, nil
		},
	}

	handler := NewExamplesHandler(store)
	input := metatools.ListToolExamplesInput{
		ToolID: "test.tool",
	}

	result, err := handler.Handle(context.Background(), input)
	require.NoError(t, err)
	assert.Empty(t, result.Examples)
}
