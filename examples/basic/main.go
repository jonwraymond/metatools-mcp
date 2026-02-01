// Package main provides a minimal metatools-mcp server example.
package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/jonwraymond/metatools-mcp/internal/adapters"
	"github.com/jonwraymond/metatools-mcp/internal/server"
	"github.com/jonwraymond/tooldiscovery/index"
	"github.com/jonwraymond/tooldiscovery/tooldoc"
	"github.com/jonwraymond/toolfoundation/model"
	"github.com/jonwraymond/toolexec/run"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// mapLocalRegistry is a minimal LocalRegistry for local tool handlers.
type mapLocalRegistry struct {
	handlers map[string]run.LocalHandler
}

func newMapLocalRegistry() *mapLocalRegistry {
	return &mapLocalRegistry{handlers: make(map[string]run.LocalHandler)}
}

func (r *mapLocalRegistry) Get(name string) (run.LocalHandler, bool) {
	h, ok := r.handlers[name]
	return h, ok
}

func (r *mapLocalRegistry) Register(name string, h run.LocalHandler) {
	r.handlers[name] = h
}

func main() {
	// Wire the core libraries.
	idx := index.NewInMemoryIndex()
	docs := tooldoc.NewInMemoryStore(tooldoc.StoreOptions{Index: idx})
	locals := newMapLocalRegistry()
	runner := run.NewRunner(
		run.WithIndex(idx),
		run.WithLocalRegistry(locals),
	)

	// Register a simple local tool.
	locals.Register("echo", func(ctx context.Context, args map[string]any) (any, error) {
		_ = ctx
		msg, _ := args["message"].(string)
		return map[string]any{"echo": msg}, nil
	})

	tool := model.Tool{
		Tool: mcp.Tool{
			Name:        "echo",
			Title:       "Echo",
			Description: "Echo a message back to the caller",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"message": map[string]any{"type": "string"},
				},
				"required":             []string{"message"},
				"additionalProperties": false,
			},
			OutputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"echo": map[string]any{"type": "string"},
				},
				"required":             []string{"echo"},
				"additionalProperties": false,
			},
		},
		Namespace: "demo",
		Tags:      []string{"example", "echo"},
	}

	backend := model.ToolBackend{
		Kind: model.BackendKindLocal,
		Local: &model.LocalBackend{
			Name: "echo",
		},
	}

	if err := idx.RegisterTool(tool, backend); err != nil {
		log.Fatalf("register tool: %v", err)
	}

	if err := docs.RegisterDoc(tool.ToolID(), tooldoc.DocEntry{
		Summary: "Echo a string for testing MCP flows",
		Notes:   "This is a local demo tool wired through toolrun.",
		Examples: []tooldoc.ToolExample{
			{
				Title:       "Echo hello",
				Description: "Return the same message you send.",
				Args:        map[string]any{"message": "hello"},
				ResultHint:  "structured.echo == \"hello\"",
			},
		},
	}); err != nil {
		log.Fatalf("register doc: %v", err)
	}

	cfg := adapters.NewConfig(idx, docs, runner, nil)
	srv, err := server.New(cfg)
	if err != nil {
		log.Fatalf("new server: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	log.Printf("metatools-mcp example starting with %d tools", len(srv.ListTools()))
	if err := srv.Run(ctx, &mcp.StdioTransport{}); err != nil && ctx.Err() == nil {
		log.Fatalf("server run: %v", err)
	}
}
