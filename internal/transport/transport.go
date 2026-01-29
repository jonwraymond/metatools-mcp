package transport

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Server provides the MCP server contract needed by transports.
type Server interface {
	Run(ctx context.Context, transport mcp.Transport) error
	MCPServer() *mcp.Server
}

// TransportInfo describes a transport instance.
type TransportInfo struct {
	Name string
	Addr string
	Path string
}

// Transport defines the interface for MCP protocol transports.
type Transport interface {
	Name() string
	Info() TransportInfo
	Serve(ctx context.Context, server Server) error
	Close() error
}
