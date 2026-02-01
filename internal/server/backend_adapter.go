// Package server implements the MCP server wiring.
package server

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jonwraymond/metatools-mcp/internal/backend"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// BackendAdapter adapts the backend registry for the MCP server.
type BackendAdapter struct {
	registry   *backend.Registry
	aggregator *backend.Aggregator
}

// NewBackendAdapter creates a new adapter.
func NewBackendAdapter(registry *backend.Registry) *BackendAdapter {
	return &BackendAdapter{
		registry:   registry,
		aggregator: backend.NewAggregator(registry),
	}
}

// GetTools returns all tools from all enabled backends.
func (a *BackendAdapter) GetTools(ctx context.Context) ([]mcp.Tool, error) {
	tools, err := a.aggregator.ListAllTools(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]mcp.Tool, 0, len(tools))
	for _, t := range tools {
		tool := t.Tool
		tool.Name = backend.FormatToolID(t.Namespace, t.Name)
		out = append(out, tool)
	}
	return out, nil
}

// Execute invokes a tool through the backend registry.
func (a *BackendAdapter) Execute(ctx context.Context, toolID string, args map[string]any) (any, error) {
	return a.aggregator.Execute(ctx, toolID, args)
}

// CreateToolHandler creates an MCP tool handler for backend tools.
func (a *BackendAdapter) CreateToolHandler() mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var args map[string]any
		if req != nil && req.Params != nil && len(req.Params.Arguments) > 0 {
			if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
				return nil, err
			}
		}

		result, err := a.Execute(ctx, req.Params.Name, args)
		if err != nil {
			return nil, err
		}

		text, err := formatAsText(result)
		if err != nil {
			return nil, err
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: text},
			},
			StructuredContent: result,
		}, nil
	}
}

func formatAsText(result any) (string, error) {
	switch v := result.(type) {
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	default:
		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return fmt.Sprintf("%v", result), nil
		}
		return string(data), nil
	}
}
