package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/jonwraymond/metatools-mcp/internal/adapters"
	"github.com/jonwraymond/metatools-mcp/internal/server"
	"github.com/jonwraymond/tooldocs"
	"github.com/jonwraymond/toolindex"
	"github.com/jonwraymond/toolrun"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	srv, err := createServer()
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	tools := srv.ListTools()
	log.Printf("metatools-mcp server starting with %d tools", len(tools))

	transport := &mcp.StdioTransport{}
	if err := srv.Run(ctx, transport); err != nil && ctx.Err() == nil {
		log.Fatalf("Server error: %v", err)
	}
	log.Println("Server stopped")
}

// createServer creates a new metatools server with default dependencies
func createServer() (*server.Server, error) {
	idx := toolindex.NewInMemoryIndex()
	docs := tooldocs.NewInMemoryStore(tooldocs.StoreOptions{Index: idx})
	runner := toolrun.NewRunner(toolrun.WithIndex(idx))

	cfg := adapters.NewConfig(idx, docs, runner, nil)
	return server.New(cfg)
}
