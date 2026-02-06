package handlers

import (
	"context"
	"errors"
	"testing"

	"github.com/jonwraymond/metatools-mcp/pkg/metatools"
	"github.com/jonwraymond/tooldiscovery/index"
	"github.com/jonwraymond/toolfoundation/model"
	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockListIndex struct {
	searchFunc      func(_ context.Context, query string, limit int, cursor string) ([]metatools.ToolSummary, string, error)
	getBackendsFunc func(_ context.Context, id string) ([]model.ToolBackend, error)
}

func (m *mockListIndex) SearchPage(ctx context.Context, query string, limit int, cursor string) ([]metatools.ToolSummary, string, error) {
	if m.searchFunc != nil {
		return m.searchFunc(ctx, query, limit, cursor)
	}
	return []metatools.ToolSummary{}, "", nil
}

func (m *mockListIndex) ListNamespacesPage(_ context.Context, _ int, _ string) ([]string, string, error) {
	return []string{}, "", nil
}

func (m *mockListIndex) GetAllBackends(ctx context.Context, id string) ([]model.ToolBackend, error) {
	if m.getBackendsFunc != nil {
		return m.getBackendsFunc(ctx, id)
	}
	return []model.ToolBackend{}, nil
}

func TestListTools_NoFilter(t *testing.T) {
	idx := &mockListIndex{
		searchFunc: func(_ context.Context, query string, limit int, cursor string) ([]metatools.ToolSummary, string, error) {
			assert.Equal(t, "", query)
			assert.Equal(t, 2, limit)
			assert.Equal(t, "", cursor)
			return []metatools.ToolSummary{
				{ID: "a", Name: "tool-a"},
				{ID: "b", Name: "tool-b"},
			}, "next", nil
		},
	}

	handler := NewListToolsHandler(idx)
	limit := 2
	out, err := handler.Handle(context.Background(), metatools.ListToolsInput{Limit: &limit})
	require.NoError(t, err)
	assert.Len(t, out.Tools, 2)
	require.NotNil(t, out.NextCursor)
	assert.Equal(t, "next", *out.NextCursor)
}

func TestListTools_FilterByBackend(t *testing.T) {
	idx := &mockListIndex{
		searchFunc: func(_ context.Context, _ string, _ int, _ string) ([]metatools.ToolSummary, string, error) {
			return []metatools.ToolSummary{
				{ID: "mcp.deepwiki:search", Name: "search"},
				{ID: "local.ping", Name: "ping"},
			}, "", nil
		},
		getBackendsFunc: func(_ context.Context, id string) ([]model.ToolBackend, error) {
			switch id {
			case "mcp.deepwiki:search":
				return []model.ToolBackend{{Kind: model.BackendKindMCP, MCP: &model.MCPBackend{ServerName: "deepwiki"}}}, nil
			case "local.ping":
				return []model.ToolBackend{{Kind: model.BackendKindLocal, Local: &model.LocalBackend{Name: "ping"}}}, nil
			default:
				return nil, index.ErrNotFound
			}
		},
	}

	handler := NewListToolsHandler(idx)
	kind := "mcp"
	name := "deepwiki"
	out, err := handler.Handle(context.Background(), metatools.ListToolsInput{BackendKind: &kind, BackendName: &name})
	require.NoError(t, err)
	require.Len(t, out.Tools, 1)
	assert.Equal(t, "mcp.deepwiki:search", out.Tools[0].ID)
}

func TestListTools_InvalidCursor(t *testing.T) {
	idx := &mockListIndex{
		searchFunc: func(_ context.Context, _ string, _ int, _ string) ([]metatools.ToolSummary, string, error) {
			return nil, "", index.ErrInvalidCursor
		},
	}

	handler := NewListToolsHandler(idx)
	_, err := handler.Handle(context.Background(), metatools.ListToolsInput{})
	require.Error(t, err)

	var rpcErr *jsonrpc.Error
	if errors.As(err, &rpcErr) {
		assert.EqualValues(t, jsonrpc.CodeInvalidParams, rpcErr.Code)
	} else {
		t.Fatalf("expected jsonrpc.Error, got %T", err)
	}
}
