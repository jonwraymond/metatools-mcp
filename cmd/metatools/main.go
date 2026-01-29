package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	cmdpkg "github.com/jonwraymond/metatools-mcp/cmd/metatools/cmd"
	"github.com/jonwraymond/metatools-mcp/internal/adapters"
	"github.com/jonwraymond/metatools-mcp/internal/bootstrap"
	"github.com/jonwraymond/metatools-mcp/internal/config"
	"github.com/jonwraymond/metatools-mcp/internal/server"
	"github.com/jonwraymond/tooldocs"
	"github.com/jonwraymond/toolrun"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	if err := cmdpkg.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func createServer() (*server.Server, error) {
	envCfg, err := config.LoadEnv()
	if err != nil {
		return nil, fmt.Errorf("loading env config: %w", err)
	}
	if err := envCfg.ValidateEnv(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	idx, err := bootstrap.NewIndexFromConfig(envCfg)
	if err != nil {
		return nil, fmt.Errorf("creating index: %w", err)
	}
	docs := tooldocs.NewInMemoryStore(tooldocs.StoreOptions{Index: idx})
	runner := toolrun.NewRunner(toolrun.WithIndex(idx))

	exec, err := maybeCreateExecutor(idx, docs, runner)
	if err != nil {
		return nil, err
	}

	cfg := adapters.NewConfig(idx, docs, runner, exec)
	cfg.NotifyToolListChanged = envCfg.NotifyToolListChanged
	cfg.NotifyToolListChangedDebounceMs = envCfg.NotifyToolListChangedDebounceMs
	return server.New(cfg)
}

// runLegacy is retained for compatibility testing of the stdio server.
func runLegacy() error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	srv, err := createServer()
	if err != nil {
		return fmt.Errorf("create server: %w", err)
	}

	tools := srv.ListTools()
	log.Printf("metatools-mcp server starting with %d tools", len(tools))

	transport := &mcp.StdioTransport{}
	if err := srv.Run(ctx, transport); err != nil && ctx.Err() == nil {
		return fmt.Errorf("server error: %w", err)
	}
	log.Println("Server stopped")
	return nil
}
