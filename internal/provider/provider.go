package provider

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ToolProvider defines the interface for MCP tool providers.
// A provider encapsulates a tool's definition and execution logic.
//
// Contract:
// - Concurrency: implementations must be safe for concurrent use.
// - Context: Handle must honor cancellation/deadlines.
// - Errors: return structured *mcp.CallToolResult for protocol-level errors when possible.
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
//
// Contract:
// - Configure must validate config and return error on invalid input.
type ConfigurableProvider interface {
	ToolProvider

	// Configure applies configuration to the provider.
	Configure(cfg map[string]any) error
}

// StreamingProvider is a provider that supports streaming responses.
//
// Contract:
// - If HandleStream returns nil error, the channel must be non-nil.
type StreamingProvider interface {
	ToolProvider

	// HandleStream processes a streaming tool invocation.
	// Returns a channel that emits response parts.
	HandleStream(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (<-chan any, error)
}

// Factory creates provider instances.
type Factory func() ToolProvider

// Info contains metadata about a provider.
type Info struct {
	Name        string
	Description string
	Version     string
	Author      string
	Streaming   bool
}

// InfoOf returns provider metadata if available.
func InfoOf(p ToolProvider) Info {
	info := Info{
		Name: p.Name(),
	}

	tool := p.Tool()
	info.Description = tool.Description

	_, info.Streaming = p.(StreamingProvider)

	return info
}
