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
	"github.com/jonwraymond/metatools-mcp/internal/handlers"
	"github.com/jonwraymond/metatools-mcp/internal/mcpbackend"
	"github.com/jonwraymond/metatools-mcp/internal/middleware"
	"github.com/jonwraymond/metatools-mcp/internal/server"
	"github.com/jonwraymond/metatools-mcp/internal/skills"
	"github.com/jonwraymond/metatools-mcp/internal/toolset"
	transportpkg "github.com/jonwraymond/metatools-mcp/internal/transport"
	"github.com/jonwraymond/tooldiscovery/tooldoc"
	"github.com/jonwraymond/toolexec/run"
	bwssecret "github.com/jonwraymond/toolops-integrations/secret/bws"
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

var validTransports = []string{"stdio", "sse", "streamable"}

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
  stdio      - Standard input/output (default, for MCP clients like Claude Desktop)
  sse        - Server-Sent Events over HTTP (deprecated, for legacy web clients)
	streamable - Streamable HTTP (MCP spec 2025-11-25, recommended for HTTP clients)

Examples:
  metatools serve                                    # stdio mode (default)
  metatools serve --transport=streamable --port=8080 # HTTP mode
  metatools serve --config=metatools.yaml`,
		PreRunE: func(_ *cobra.Command, _ []string) error {
			return validateTransport(cfg.Transport)
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runServe(cmd.Context(), cfg)
		},
	}

	cmd.Flags().StringVarP(&cfg.Transport, "transport", "t", "stdio", "Transport type (stdio, sse, streamable)")
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

// loadServeConfig loads config with CLI overrides.
func loadServeConfig(configPath string, cli *ServeConfig) (config.AppConfig, error) {
	overrides := map[string]any{}

	if cli.Transport != "" && cli.Transport != "stdio" {
		overrides["transport.type"] = cli.Transport
	}
	if cli.Port != 0 && cli.Port != 8080 {
		overrides["transport.http.port"] = cli.Port
	}
	if cli.Host != "" && cli.Host != "0.0.0.0" {
		overrides["transport.http.host"] = cli.Host
	}

	return config.LoadWithOverrides(configPath, overrides)
}

func buildServerConfig(_ *ServeConfig) (config.Config, error) {
	appCfg, err := config.Load("")
	if err != nil {
		return config.Config{}, fmt.Errorf("loading config: %w", err)
	}
	return buildServerConfigFromConfig(appCfg)
}

func buildServerConfigFromConfig(appCfg config.AppConfig) (config.Config, error) {
	idx, err := bootstrap.NewIndexFromAppConfig(appCfg)
	if err != nil {
		return config.Config{}, fmt.Errorf("creating index: %w", err)
	}
	docs := tooldoc.NewInMemoryStore(tooldoc.StoreOptions{Index: idx})

	mcpBackendCfgs := make([]mcpbackend.Config, len(appCfg.Backends.MCP))
	for i, backend := range appCfg.Backends.MCP {
		mcpBackendCfgs[i] = mcpbackend.Config{
			Name:       backend.Name,
			URL:        backend.URL,
			Headers:    backend.Headers,
			MaxRetries: backend.MaxRetries,
		}
	}
	mcpManager, err := mcpbackend.NewManager(mcpBackendCfgs)
	if err != nil {
		return config.Config{}, fmt.Errorf("mcp backends: %w", err)
	}
	if mcpManager.HasBackends() {
		if err := mcpManager.ConnectAll(context.Background()); err != nil {
			return config.Config{}, fmt.Errorf("connect mcp backends: %w", err)
		}
		if err := mcpManager.RegisterTools(idx); err != nil {
			return config.Config{}, fmt.Errorf("register mcp tools: %w", err)
		}
	}

	runner := run.NewRunner(run.WithIndex(idx))
	if mcpManager.HasBackends() {
		runner = run.NewRunner(run.WithIndex(idx), run.WithMCPExecutor(mcpManager))
	}

	exec, err := maybeCreateExecutor(appCfg.Execution, idx, docs, runner)
	if err != nil {
		return config.Config{}, fmt.Errorf("create executor: %w", err)
	}

	toolsetSpecs := make([]toolset.Spec, len(appCfg.Toolsets))
	for i, spec := range appCfg.Toolsets {
		toolsetSpecs[i] = toolset.Spec{
			Name:             spec.Name,
			Description:      spec.Description,
			NamespaceFilters: spec.NamespaceFilters,
			TagFilters:       spec.TagFilters,
			AllowIDs:         spec.AllowIDs,
			DenyIDs:          spec.DenyIDs,
			Policy:           spec.Policy,
		}
	}
	toolsetsRegistry, err := toolset.BuildRegistry(idx, toolsetSpecs)
	if err != nil {
		return config.Config{}, fmt.Errorf("build toolsets: %w", err)
	}
	skillSpecs := make([]skills.Spec, len(appCfg.Skills))
	for i, spec := range appCfg.Skills {
		steps := make([]skills.StepSpec, len(spec.Steps))
		for j, step := range spec.Steps {
			steps[j] = skills.StepSpec{
				ID:     step.ID,
				ToolID: step.ToolID,
				Inputs: step.Inputs,
			}
		}
		skillSpecs[i] = skills.Spec{
			Name:        spec.Name,
			Description: spec.Description,
			ToolsetID:   spec.ToolsetID,
			Steps:       steps,
			Guards: skills.GuardSpec{
				MaxSteps: spec.Guards.MaxSteps,
				AllowIDs: spec.Guards.AllowIDs,
			},
		}
	}
	skillsRegistry, err := skills.BuildRegistry(toolsetsRegistry, skillSpecs)
	if err != nil {
		return config.Config{}, fmt.Errorf("build skills: %w", err)
	}

	cfg := adapters.NewConfig(idx, docs, runner, exec)
	cfg.Providers = appCfg.Providers
	cfg.Middleware = appCfg.Middleware
	cfg.Toolsets = toolsetsRegistry
	cfg.Skills = skillsRegistry
	if mcpManager.HasBackends() {
		refreshPolicy := mcpbackend.RefreshPolicy{
			Interval:   appCfg.Backends.MCPRefresh.Interval,
			Jitter:     appCfg.Backends.MCPRefresh.Jitter,
			StaleAfter: appCfg.Backends.MCPRefresh.StaleAfter,
			OnDemand:   appCfg.Backends.MCPRefresh.OnDemand,
		}
		cfg.Refresher = mcpbackend.NewRefresher(mcpManager, idx, refreshPolicy)
	}
	cfg.SkillDefaults = handlers.SkillDefaults{
		MaxSteps:     appCfg.SkillDefaults.MaxSteps,
		MaxToolCalls: appCfg.SkillDefaults.MaxToolCalls,
		Timeout:      appCfg.SkillDefaults.Timeout,
	}

	wrappedRunner, err := middleware.WrapRunner(cfg.Runner, idx, appCfg.Middleware)
	if err != nil {
		return config.Config{}, fmt.Errorf("wrap runner: %w", err)
	}
	if wrappedRunner != nil {
		cfg.Runner = wrappedRunner
	}
	if cfg.Executor != nil {
		wrappedExec, err := middleware.WrapExecutor(cfg.Executor, appCfg.Middleware)
		if err != nil {
			return config.Config{}, fmt.Errorf("wrap executor: %w", err)
		}
		if wrappedExec != nil {
			cfg.Executor = wrappedExec
		}
	}

	// Preserve notify settings from env config for now.
	envCfg, err := config.LoadEnv()
	if err != nil {
		return config.Config{}, fmt.Errorf("loading env config: %w", err)
	}
	if err := envCfg.ValidateEnv(); err != nil {
		return config.Config{}, fmt.Errorf("invalid config: %w", err)
	}
	cfg.NotifyToolListChanged = envCfg.NotifyToolListChanged
	cfg.NotifyToolListChangedDebounceMs = envCfg.NotifyToolListChangedDebounceMs

	return cfg, nil
}

func runServe(ctx context.Context, cfg *ServeConfig) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	appCfg, err := loadServeConfig(cfg.Config, cfg)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	if err := appCfg.ApplyRuntimeLimitsStore(ctx); err != nil {
		return fmt.Errorf("apply runtime limits: %w", err)
	}

	secretResolver, closeSecrets, err := bootstrap.NewSecretResolver(appCfg.Secrets, bwssecret.Register)
	if err != nil {
		return fmt.Errorf("secrets: %w", err)
	}
	if closeSecrets != nil {
		defer func() { _ = closeSecrets() }()
	}
	if secretResolver != nil {
		resolved, err := bootstrap.ResolveMCPBackendConfigs(ctx, secretResolver, appCfg.Backends.MCP)
		if err != nil {
			return fmt.Errorf("resolve mcp backend secrets: %w", err)
		}
		appCfg.Backends.MCP = resolved
	}

	serverCfg, err := buildServerConfigFromConfig(appCfg)
	if err != nil {
		return fmt.Errorf("build server config: %w", err)
	}

	srv, err := server.New(serverCfg)
	if err != nil {
		return fmt.Errorf("create server: %w", err)
	}

	if refresher, ok := serverCfg.Refresher.(*mcpbackend.Refresher); ok && refresher != nil {
		refresher.StartLoop(ctx)
	}

	var transport transportpkg.Transport
	switch appCfg.Transport.Type {
	case "stdio":
		transport = &transportpkg.StdioTransport{}
	case "sse":
		transport = &transportpkg.SSETransport{Config: transportpkg.SSEConfig{
			Host:          appCfg.Transport.HTTP.Host,
			Port:          appCfg.Transport.HTTP.Port,
			Path:          "/mcp",
			HealthEnabled: appCfg.Health.Enabled,
			HealthPath:    appCfg.Health.Path,
		}}
	case "streamable":
		transport = &transportpkg.StreamableHTTPTransport{Config: transportpkg.StreamableHTTPConfig{
			Host:           appCfg.Transport.HTTP.Host,
			Port:           appCfg.Transport.HTTP.Port,
			Path:           "/mcp",
			Stateless:      appCfg.Transport.Streamable.Stateless,
			JSONResponse:   appCfg.Transport.Streamable.JSONResponse,
			SessionTimeout: appCfg.Transport.Streamable.SessionTimeout,
			HealthEnabled:  appCfg.Health.Enabled,
			HealthPath:     appCfg.Health.Path,
			TLS: transportpkg.TLSConfig{
				Enabled:  appCfg.Transport.HTTP.TLS.Enabled,
				CertFile: appCfg.Transport.HTTP.TLS.CertFile,
				KeyFile:  appCfg.Transport.HTTP.TLS.KeyFile,
			},
		}}
	default:
		return fmt.Errorf("unknown transport: %s", appCfg.Transport.Type)
	}

	fmt.Fprintf(os.Stderr, "Starting metatools server (transport=%s)\n", appCfg.Transport.Type)
	return transport.Serve(ctx, srv)
}

// Ensure compile-time interface usage for default transport.
var _ mcp.Transport = (*mcp.StdioTransport)(nil)
