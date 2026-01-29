package server

import (
	"context"
	"fmt"

	"github.com/jonwraymond/metatools-mcp/internal/provider"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ProviderAdapter adapts a provider registry for MCP server registration.
type ProviderAdapter struct {
	registry *provider.Registry
}

// NewProviderAdapter creates a new adapter.
func NewProviderAdapter(registry *provider.Registry) *ProviderAdapter {
	return &ProviderAdapter{registry: registry}
}

// RegisterTools registers all enabled providers as MCP tools.
func (a *ProviderAdapter) RegisterTools(server *Server) error {
	if a.registry == nil {
		return fmt.Errorf("provider registry is nil")
	}
	for _, p := range a.registry.ListEnabled() {
		tool := p.Tool()
		if tool.Name == "" {
			return fmt.Errorf("provider %q returned empty tool name", p.Name())
		}
		if p.Name() != "" && tool.Name != p.Name() {
			return fmt.Errorf("provider name mismatch: provider %q tool %q", p.Name(), tool.Name)
		}
		handler := func(ctx context.Context, req *mcp.CallToolRequest, input map[string]any) (*mcp.CallToolResult, any, error) {
			return p.Handle(ctx, req, input)
		}
		registerTool(server, &tool, handler)
	}
	return nil
}
