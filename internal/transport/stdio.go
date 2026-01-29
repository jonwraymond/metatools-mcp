package transport

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// StdioTransport wraps the MCP stdio transport.
type StdioTransport struct{}

// Name returns the transport identifier.
func (t *StdioTransport) Name() string {
	return "stdio"
}

// Info returns descriptive information about the transport.
func (t *StdioTransport) Info() Info {
	return Info{Name: "stdio"}
}

// Serve starts the transport and blocks until context is cancelled.
func (t *StdioTransport) Serve(ctx context.Context, server Server) error {
	return server.Run(ctx, &mcp.StdioTransport{})
}

// Close is a no-op for stdio.
func (t *StdioTransport) Close() error {
	return nil
}
