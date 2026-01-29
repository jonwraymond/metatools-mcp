package handlers

import (
	"context"
	"errors"
	"testing"

	"github.com/jonwraymond/metatools-mcp/pkg/metatools"
	"github.com/jonwraymond/toolindex"
	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockIndex struct {
	searchFunc         func(_ context.Context, query string, limit int, cursor string) ([]metatools.ToolSummary, string, error)
	listNamespacesFunc func(_ context.Context, limit int, cursor string) ([]string, string, error)
}

func (m *mockIndex) SearchPage(ctx context.Context, query string, limit int, cursor string) ([]metatools.ToolSummary, string, error) {
	if m.searchFunc != nil {
		return m.searchFunc(ctx, query, limit, cursor)
	}
	return []metatools.ToolSummary{}, "", nil
}

func (m *mockIndex) ListNamespacesPage(ctx context.Context, limit int, cursor string) ([]string, string, error) {
	if m.listNamespacesFunc != nil {
		return m.listNamespacesFunc(ctx, limit, cursor)
	}
	return []string{}, "", nil
}

func TestSearchTools_EmptyQuery(t *testing.T) {
	idx := &mockIndex{
		searchFunc: func(_ context.Context, _ string, _ int, _ string) ([]metatools.ToolSummary, string, error) {
			// Empty query should still work, returns all tools
			return []metatools.ToolSummary{
				{ID: "test.tool1", Name: "tool1"},
				{ID: "test.tool2", Name: "tool2"},
			}, "", nil
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
		searchFunc: func(_ context.Context, query string, _ int, _ string) ([]metatools.ToolSummary, string, error) {
			assert.Equal(t, "test query", query)
			return []metatools.ToolSummary{
				{ID: "test.matching", Name: "matching", ShortDescription: "test query match"},
			}, "", nil
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
		searchFunc: func(_ context.Context, _ string, limit int, _ string) ([]metatools.ToolSummary, string, error) {
			// Verify limit is passed through
			assert.Equal(t, 5, limit)
			return []metatools.ToolSummary{
				{ID: "a"}, {ID: "b"}, {ID: "c"}, {ID: "d"}, {ID: "e"},
			}, "", nil
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
	idx := &mockIndex{
		searchFunc: func(_ context.Context, _ string, _ int, cursor string) ([]metatools.ToolSummary, string, error) {
			if cursor == "" {
				return []metatools.ToolSummary{{ID: "a"}, {ID: "b"}}, "cursor-2", nil
			}
			assert.Equal(t, "cursor-2", cursor)
			return []metatools.ToolSummary{{ID: "c"}, {ID: "d"}}, "", nil
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
	idx := &mockIndex{
		searchFunc: func(_ context.Context, _ string, _ int, _ string) ([]metatools.ToolSummary, string, error) {
			return []metatools.ToolSummary{{ID: "a"}, {ID: "b"}}, "cursor-2", nil
		},
	}

	handler := NewSearchHandler(idx)
	limit := 2
	input := metatools.SearchToolsInput{Query: "", Limit: &limit}

	result, err := handler.Handle(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, result.NextCursor)
	assert.Equal(t, "cursor-2", *result.NextCursor)
}

func TestSearchTools_ReturnsSummariesNotFullSchemas(t *testing.T) {
	idx := &mockIndex{
		searchFunc: func(_ context.Context, _ string, _ int, _ string) ([]metatools.ToolSummary, string, error) {
			return []metatools.ToolSummary{
				{
					ID:               "test.tool",
					Name:             "tool",
					Namespace:        "test",
					ShortDescription: "A test tool",
					Tags:             []string{"test"},
				},
			}, "", nil
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
		searchFunc: func(_ context.Context, _ string, _ int, _ string) ([]metatools.ToolSummary, string, error) {
			return nil, "", errors.New("index error")
		},
	}

	handler := NewSearchHandler(idx)
	input := metatools.SearchToolsInput{Query: "test"}

	_, err := handler.Handle(context.Background(), input)
	assert.Error(t, err)
}

func TestSearchTools_InvalidCursor(t *testing.T) {
	idx := &mockIndex{
		searchFunc: func(_ context.Context, _ string, _ int, _ string) ([]metatools.ToolSummary, string, error) {
			return nil, "", toolindex.ErrInvalidCursor
		},
	}

	handler := NewSearchHandler(idx)
	input := metatools.SearchToolsInput{Query: "test", Cursor: strPtr("bad")}

	_, err := handler.Handle(context.Background(), input)
	require.Error(t, err)

	rpcErr, ok := err.(*jsonrpc.Error)
	require.True(t, ok, "expected jsonrpc.Error")
	assert.EqualValues(t, jsonrpc.CodeInvalidParams, rpcErr.Code)
}

func strPtr(s string) *string {
	return &s
}
