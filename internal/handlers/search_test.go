package handlers

import (
	"context"
	"errors"
	"testing"

	"github.com/jonwraymond/metatools-mcp/pkg/metatools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockIndex struct {
	searchFunc         func(_ context.Context, query string, limit int) ([]metatools.ToolSummary, error)
	listNamespacesFunc func(_ context.Context) ([]string, error)
}

func (m *mockIndex) Search(ctx context.Context, query string, limit int) ([]metatools.ToolSummary, error) {
	if m.searchFunc != nil {
		return m.searchFunc(ctx, query, limit)
	}
	return []metatools.ToolSummary{}, nil
}

func (m *mockIndex) ListNamespaces(ctx context.Context) ([]string, error) {
	if m.listNamespacesFunc != nil {
		return m.listNamespacesFunc(ctx)
	}
	return []string{}, nil
}

func TestSearchTools_EmptyQuery(t *testing.T) {
	idx := &mockIndex{
		searchFunc: func(_ context.Context, query string, limit int) ([]metatools.ToolSummary, error) {
			// Empty query should still work, returns all tools
			return []metatools.ToolSummary{
				{ID: "test.tool1", Name: "tool1"},
				{ID: "test.tool2", Name: "tool2"},
			}, nil
		},
	}

	handler := NewSearchHandler(idx)
	input := metatools.SearchToolsInput{Query: ""}

	result, err := handler.Handle(context.Background(), input)
	require.NoError(t, err)
	assert.Len(t, result.Tools, 2)
}

func TestSearchTools_WithQuery(t *testing.T) {
	idx := &mockIndex{
		searchFunc: func(_ context.Context, query string, limit int) ([]metatools.ToolSummary, error) {
			assert.Equal(t, "test query", query)
			return []metatools.ToolSummary{
				{ID: "test.matching", Name: "matching", ShortDescription: "test query match"},
			}, nil
		},
	}

	handler := NewSearchHandler(idx)
	input := metatools.SearchToolsInput{Query: "test query"}

	result, err := handler.Handle(context.Background(), input)
	require.NoError(t, err)
	assert.Len(t, result.Tools, 1)
	assert.Equal(t, "test.matching", result.Tools[0].ID)
}

func TestSearchTools_WithLimit(t *testing.T) {
	idx := &mockIndex{
		searchFunc: func(_ context.Context, query string, limit int) ([]metatools.ToolSummary, error) {
			// Verify limit is passed through
			assert.Equal(t, 5, limit)
			return []metatools.ToolSummary{
				{ID: "a"}, {ID: "b"}, {ID: "c"}, {ID: "d"}, {ID: "e"},
			}, nil
		},
	}

	handler := NewSearchHandler(idx)
	limit := 5
	input := metatools.SearchToolsInput{Query: "test", Limit: &limit}

	result, err := handler.Handle(context.Background(), input)
	require.NoError(t, err)
	assert.Len(t, result.Tools, 5)
}

func TestSearchTools_WithCursor(t *testing.T) {
	allTools := []metatools.ToolSummary{
		{ID: "a"}, {ID: "b"}, {ID: "c"}, {ID: "d"}, {ID: "e"},
	}

	idx := &mockIndex{
		searchFunc: func(_ context.Context, query string, limit int) ([]metatools.ToolSummary, error) {
			return allTools, nil
		},
	}

	handler := NewSearchHandler(idx)

	// First request with limit 2
	limit := 2
	input := metatools.SearchToolsInput{Query: "", Limit: &limit}
	result, err := handler.Handle(context.Background(), input)
	require.NoError(t, err)
	assert.Len(t, result.Tools, 2)
	require.NotNil(t, result.NextCursor)

	// Second request with cursor
	cursor := *result.NextCursor
	input2 := metatools.SearchToolsInput{Query: "", Limit: &limit, Cursor: &cursor}
	result2, err := handler.Handle(context.Background(), input2)
	require.NoError(t, err)
	assert.Len(t, result2.Tools, 2)
	assert.Equal(t, "c", result2.Tools[0].ID)
}

func TestSearchTools_ReturnsNextCursor(t *testing.T) {
	allTools := []metatools.ToolSummary{
		{ID: "a"}, {ID: "b"}, {ID: "c"},
	}

	idx := &mockIndex{
		searchFunc: func(_ context.Context, query string, limit int) ([]metatools.ToolSummary, error) {
			return allTools, nil
		},
	}

	handler := NewSearchHandler(idx)
	limit := 2
	input := metatools.SearchToolsInput{Query: "", Limit: &limit}

	result, err := handler.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, result.NextCursor)
}

func TestSearchTools_ReturnsSummariesNotFullSchemas(t *testing.T) {
	idx := &mockIndex{
		searchFunc: func(_ context.Context, query string, limit int) ([]metatools.ToolSummary, error) {
			return []metatools.ToolSummary{
				{
					ID:               "test.tool",
					Name:             "tool",
					Namespace:        "test",
					ShortDescription: "A test tool",
					Tags:             []string{"test"},
				},
			}, nil
		},
	}

	handler := NewSearchHandler(idx)
	input := metatools.SearchToolsInput{Query: "test"}

	result, err := handler.Handle(context.Background(), input)
	require.NoError(t, err)

	// Verify it's a summary, not full schema
	assert.Equal(t, "test.tool", result.Tools[0].ID)
	assert.Equal(t, "A test tool", result.Tools[0].ShortDescription)
}

func TestSearchTools_IndexError(t *testing.T) {
	idx := &mockIndex{
		searchFunc: func(_ context.Context, query string, limit int) ([]metatools.ToolSummary, error) {
			return nil, errors.New("index error")
		},
	}

	handler := NewSearchHandler(idx)
	input := metatools.SearchToolsInput{Query: "test"}

	_, err := handler.Handle(context.Background(), input)
	assert.Error(t, err)
}
