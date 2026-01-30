package transport

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Server provides the MCP server contract needed by transports.
//
// Contract:
// - Concurrency: implementations must be safe for concurrent use.
// - Context: Run must honor cancellation/deadlines.
type Server interface {
	Run(ctx context.Context, transport mcp.Transport) error
	MCPServer() *mcp.Server
}

// Info describes a transport instance.
type Info struct {
	Name string
	Addr string
	Path string
}

// Transport defines the interface for MCP protocol transports.
//
// Contract:
// - Concurrency: implementations must be safe for concurrent use.
// - Context: Serve must honor cancellation/deadlines.
// - Errors: Close must be idempotent.
type Transport interface {
	Name() string
	Info() Info
	Serve(ctx context.Context, server Server) error
	Close() error
}
