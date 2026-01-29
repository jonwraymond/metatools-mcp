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

type mockNamespacesIndex struct {
	listNamespacesFunc func(_ context.Context, limit int, cursor string) ([]string, string, error)
}

func (m *mockNamespacesIndex) SearchPage(_ context.Context, _ string, _ int, _ string) ([]metatools.ToolSummary, string, error) {
	return []metatools.ToolSummary{}, "", nil
}

func (m *mockNamespacesIndex) ListNamespacesPage(ctx context.Context, limit int, cursor string) ([]string, string, error) {
	if m.listNamespacesFunc != nil {
		return m.listNamespacesFunc(ctx, limit, cursor)
	}
	return []string{}, "", nil
}

func TestListNamespaces_Empty(t *testing.T) {
	idx := &mockNamespacesIndex{
		listNamespacesFunc: func(_ context.Context, _ int, _ string) ([]string, string, error) {
			return []string{}, "", nil
		},
	}

	handler := NewNamespacesHandler(idx)
	result, err := handler.Handle(context.Background(), metatools.ListNamespacesInput{})

	require.NoError(t, err)
	assert.Empty(t, result.Namespaces)
}

func TestListNamespaces_ReturnsSorted(t *testing.T) {
	idx := &mockNamespacesIndex{
		listNamespacesFunc: func(_ context.Context, _ int, _ string) ([]string, string, error) {
			return []string{"alpha", "beta", "gamma"}, "", nil
		},
	}

	handler := NewNamespacesHandler(idx)
	result, err := handler.Handle(context.Background(), metatools.ListNamespacesInput{})

	require.NoError(t, err)
	assert.Equal(t, []string{"alpha", "beta", "gamma"}, result.Namespaces)
}

func TestListNamespaces_IndexError(t *testing.T) {
	idx := &mockNamespacesIndex{
		listNamespacesFunc: func(_ context.Context, _ int, _ string) ([]string, string, error) {
			return nil, "", errors.New("index error")
		},
	}

	handler := NewNamespacesHandler(idx)
	_, err := handler.Handle(context.Background(), metatools.ListNamespacesInput{})

	assert.Error(t, err)
}

func TestListNamespaces_WithCursor(t *testing.T) {
	idx := &mockNamespacesIndex{
		listNamespacesFunc: func(_ context.Context, limit int, cursor string) ([]string, string, error) {
			assert.Equal(t, 2, limit)
			if cursor == "" {
				return []string{"a", "b"}, "cursor-2", nil
			}
			assert.Equal(t, "cursor-2", cursor)
			return []string{"c"}, "", nil
		},
	}

	handler := NewNamespacesHandler(idx)
	limit := 2

	result, err := handler.Handle(context.Background(), metatools.ListNamespacesInput{Limit: &limit})
	require.NoError(t, err)
	assert.Equal(t, []string{"a", "b"}, result.Namespaces)
	require.NotNil(t, result.NextCursor)

	cursor := *result.NextCursor
	result2, err := handler.Handle(context.Background(), metatools.ListNamespacesInput{Limit: &limit, Cursor: &cursor})
	require.NoError(t, err)
	assert.Equal(t, []string{"c"}, result2.Namespaces)
}

func TestListNamespaces_InvalidCursor(t *testing.T) {
	idx := &mockNamespacesIndex{
		listNamespacesFunc: func(_ context.Context, _ int, _ string) ([]string, string, error) {
			return nil, "", toolindex.ErrInvalidCursor
		},
	}

	handler := NewNamespacesHandler(idx)
	cursor := "bad"
	_, err := handler.Handle(context.Background(), metatools.ListNamespacesInput{Cursor: &cursor})
	require.Error(t, err)

	rpcErr, ok := err.(*jsonrpc.Error)
	require.True(t, ok, "expected jsonrpc.Error")
	assert.EqualValues(t, jsonrpc.CodeInvalidParams, rpcErr.Code)
}
