package server

import (
	"context"
	"testing"

	"github.com/jonwraymond/metatools-mcp/internal/provider"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

type stubProvider struct {
	name string
	tool mcp.Tool
}

func (s *stubProvider) Name() string  { return s.name }
func (s *stubProvider) Enabled() bool { return true }
func (s *stubProvider) Tool() mcp.Tool {
	return s.tool
}
func (s *stubProvider) Handle(ctx context.Context, _ *mcp.CallToolRequest, _ map[string]any) (*mcp.CallToolResult, any, error) {
	return nil, map[string]any{"ok": true}, nil
}

func TestProviderAdapter_RegisterTools(t *testing.T) {
	registry := provider.NewRegistry()
	require.NoError(t, registry.Register(&stubProvider{
		name: "test_tool",
		tool: mcp.Tool{
			Name:        "test_tool",
			Description: "test",
			InputSchema: map[string]any{"type": "object"},
		},
	}))

	mcpServer := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "dev"}, &mcp.ServerOptions{})
	srv := &Server{mcp: mcpServer}

	adapter := NewProviderAdapter(registry)
	err := adapter.RegisterTools(srv)
	require.NoError(t, err)

	tools := srv.ListTools()
	require.Len(t, tools, 1)
	require.Equal(t, "test_tool", tools[0].Name)
}
