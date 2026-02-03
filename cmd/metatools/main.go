package main

import (
	"context"
	"fmt"
	"os"

	cmdpkg "github.com/jonwraymond/metatools-mcp/cmd/metatools/cmd"
	"github.com/jonwraymond/metatools-mcp/internal/adapters"
	"github.com/jonwraymond/metatools-mcp/internal/bootstrap"
	"github.com/jonwraymond/metatools-mcp/internal/config"
	"github.com/jonwraymond/metatools-mcp/internal/server"
	"github.com/jonwraymond/tooldiscovery/tooldoc"
	"github.com/jonwraymond/toolexec/run"
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

	appCfg, err := config.Load("")
	if err != nil {
		return nil, fmt.Errorf("loading app config: %w", err)
	}
	if err := appCfg.ApplyRuntimeLimitsStore(context.Background()); err != nil {
		return nil, fmt.Errorf("apply runtime limits: %w", err)
	}

	idx, err := bootstrap.NewIndexFromConfig(envCfg)
	if err != nil {
		return nil, fmt.Errorf("creating index: %w", err)
	}
	docs := tooldoc.NewInMemoryStore(tooldoc.StoreOptions{Index: idx})
	runner := run.NewRunner(run.WithIndex(idx))

	exec, err := maybeCreateExecutor(appCfg.Execution, idx, docs, runner)
	if err != nil {
		return nil, err
	}

	cfg := adapters.NewConfig(idx, docs, runner, exec)
	cfg.NotifyToolListChanged = envCfg.NotifyToolListChanged
	cfg.NotifyToolListChangedDebounceMs = envCfg.NotifyToolListChangedDebounceMs
	return server.New(cfg)
}

// runLegacy is retained for compatibility testing of the stdio server.
