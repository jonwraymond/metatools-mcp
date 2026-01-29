package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/jonwraymond/metatools-mcp/internal/adapters"
	"github.com/jonwraymond/metatools-mcp/internal/bootstrap"
	"github.com/jonwraymond/metatools-mcp/internal/config"
	"github.com/jonwraymond/metatools-mcp/internal/server"
	"github.com/jonwraymond/tooldocs"
	"github.com/jonwraymond/toolrun"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
)

// ServeConfig holds serve command configuration.
type ServeConfig struct {
	Transport string
	Port      int
	Host      string
	Config    string
}

var validTransports = []string{"stdio", "sse", "http"}

func validateTransport(transport string) error {
	for _, valid := range validTransports {
		if transport == valid {
			return nil
		}
	}
	return fmt.Errorf("invalid transport %q, must be one of: %v", transport, validTransports)
}

func newServeCmd() *cobra.Command {
	cfg := &ServeConfig{}

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the MCP server",
		Long: `Start the metatools MCP server with the specified transport.

Transports:
  stdio  - Standard input/output (default, for MCP clients like Claude Desktop)
  sse    - Server-Sent Events over HTTP (for web clients)
  http   - Simple HTTP request/response (for REST clients)

Examples:
  metatools serve                           # stdio mode (default)
  metatools serve --transport=sse --port=8080
  metatools serve --config=metatools.yaml`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return validateTransport(cfg.Transport)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServe(cmd.Context(), cfg)
		},
	}

	cmd.Flags().StringVarP(&cfg.Transport, "transport", "t", "stdio", "Transport type (stdio, sse, http)")
	cmd.Flags().IntVarP(&cfg.Port, "port", "p", 8080, "Port for HTTP transports")
	cmd.Flags().StringVar(&cfg.Host, "host", "0.0.0.0", "Host to bind for HTTP transports")
	cmd.Flags().StringVarP(&cfg.Config, "config", "c", "", "Path to config file")

	applyServeEnvDefaults(cmd, cfg)

	return cmd
}

func applyServeEnvDefaults(cmd *cobra.Command, cfg *ServeConfig) {
	if !cmd.Flags().Changed("transport") {
		if v := os.Getenv("METATOOLS_TRANSPORT"); v != "" {
			_ = cmd.Flags().Set("transport", v)
			cfg.Transport = v
		}
	}
	if !cmd.Flags().Changed("port") {
		if v := os.Getenv("METATOOLS_PORT"); v != "" {
			if port, err := strconv.Atoi(v); err == nil {
				_ = cmd.Flags().Set("port", v)
				cfg.Port = port
			}
		}
	}
	if !cmd.Flags().Changed("host") {
		if v := os.Getenv("METATOOLS_HOST"); v != "" {
			_ = cmd.Flags().Set("host", v)
			cfg.Host = v
		}
	}
	if !cmd.Flags().Changed("config") {
		if v := os.Getenv("METATOOLS_CONFIG"); v != "" {
			_ = cmd.Flags().Set("config", v)
			cfg.Config = v
		}
	}
}

func buildServerConfig(_ *ServeConfig) (config.Config, error) {
	envCfg, err := config.LoadEnv()
	if err != nil {
		return config.Config{}, fmt.Errorf("loading env config: %w", err)
	}
	if err := envCfg.ValidateEnv(); err != nil {
		return config.Config{}, fmt.Errorf("invalid config: %w", err)
	}

	idx, err := bootstrap.NewIndexFromConfig(envCfg)
	if err != nil {
		return config.Config{}, fmt.Errorf("creating index: %w", err)
	}
	docs := tooldocs.NewInMemoryStore(tooldocs.StoreOptions{Index: idx})
	runner := toolrun.NewRunner(toolrun.WithIndex(idx))

	return adapters.NewConfig(idx, docs, runner, nil), nil
}

func runServe(ctx context.Context, cfg *ServeConfig) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	serverCfg, err := buildServerConfig(cfg)
	if err != nil {
		return fmt.Errorf("build server config: %w", err)
	}

	srv, err := server.New(serverCfg)
	if err != nil {
		return fmt.Errorf("create server: %w", err)
	}

	var transport mcp.Transport
	switch cfg.Transport {
	case "stdio":
		transport = &mcp.StdioTransport{}
	case "sse", "http":
		return fmt.Errorf("transport %q not yet implemented", cfg.Transport)
	default:
		return fmt.Errorf("unknown transport: %s", cfg.Transport)
	}

	fmt.Fprintf(os.Stderr, "Starting metatools server (transport=%s)\n", cfg.Transport)
	return srv.Run(ctx, transport)
}

// Ensure compile-time interface usage for default transport.
var _ mcp.Transport = (*mcp.StdioTransport)(nil)
