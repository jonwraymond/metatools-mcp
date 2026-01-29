package provider

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ToolProvider defines the interface for MCP tool providers.
// A provider encapsulates a tool's definition and execution logic.
type ToolProvider interface {
	// Name returns the unique identifier for this provider.
	Name() string

	// Enabled returns whether this provider is currently enabled.
	Enabled() bool

	// Tool returns the MCP tool definition.
	Tool() mcp.Tool

	// Handle processes a tool invocation.
	// Implementations may return a CallToolResult to set IsError or other fields.
	Handle(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error)
}

// ConfigurableProvider is a provider that can be configured at runtime.
type ConfigurableProvider interface {
	ToolProvider

	// Configure applies configuration to the provider.
	Configure(cfg map[string]any) error
}

// StreamingProvider is a provider that supports streaming responses.
type StreamingProvider interface {
	ToolProvider

	// HandleStream processes a streaming tool invocation.
	// Returns a channel that emits response parts.
	HandleStream(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (<-chan any, error)
}

// ProviderFactory creates provider instances.
type ProviderFactory func() ToolProvider

// ProviderInfo contains metadata about a provider.
type ProviderInfo struct {
	Name        string
	Description string
	Version     string
	Author      string
	Streaming   bool
}

// GetInfo returns provider metadata if available.
func GetInfo(p ToolProvider) ProviderInfo {
	info := ProviderInfo{
		Name: p.Name(),
	}

	tool := p.Tool()
	info.Description = tool.Description

	_, info.Streaming = p.(StreamingProvider)

	return info
}
