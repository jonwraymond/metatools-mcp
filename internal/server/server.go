package server

import (
	"context"
	"sync"
	"time"

	"github.com/jonwraymond/metatools-mcp/internal/config"
	"github.com/jonwraymond/metatools-mcp/internal/handlers"
	"github.com/jonwraymond/metatools-mcp/internal/provider/builtin"
	"github.com/jonwraymond/tooldiscovery/index"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	implementationName    = "metatools-mcp"
	implementationVersion = "0.1.0"
	defaultPageSize       = 50
)

// Capabilities represents the capabilities this server supports.
// This mirrors the MCP tools capability for simple checks in tests and callers.
type Capabilities struct {
	Tools bool
}

// Server is the metatools MCP server backed by the official MCP Go SDK.
type Server struct {
	config   config.Config
	mcp      *mcp.Server
	tools    []*mcp.Tool
	handlers *Handlers

	toolRegistrations []func()
	toolListMu        sync.Mutex
	toolListTimer     *time.Timer
	toolListUnsub     func()
}

// Handlers holds all the metatool handlers.
type Handlers struct {
	Search     *handlers.SearchHandler
	Namespaces *handlers.NamespacesHandler
	Describe   *handlers.DescribeHandler
	Examples   *handlers.ExamplesHandler
	Run        *handlers.RunHandler
	Chain      *handlers.ChainHandler
	Code       *handlers.CodeHandler
}

// New creates a new metatools server.
func New(cfg config.Config) (*Server, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	h := &Handlers{
		Search:     handlers.NewSearchHandler(cfg.Index),
		Namespaces: handlers.NewNamespacesHandler(cfg.Index),
		Describe:   handlers.NewDescribeHandler(cfg.Docs),
		Examples:   handlers.NewExamplesHandler(cfg.Docs),
		Run:        handlers.NewRunHandler(cfg.Runner),
		Chain:      handlers.NewChainHandler(cfg.Runner),
	}
	if cfg.Executor != nil {
		h.Code = handlers.NewCodeHandler(cfg.Executor)
	}

	serverOptions := &mcp.ServerOptions{
		PageSize: defaultPageSize,
	}
	if !cfg.NotifyToolListChanged {
		serverOptions.Capabilities = &mcp.ServerCapabilities{
			Logging: &mcp.LoggingCapabilities{},
			Tools:   &mcp.ToolCapabilities{},
		}
	}

	mcpServer := mcp.NewServer(&mcp.Implementation{
		Name:    implementationName,
		Version: implementationVersion,
	}, serverOptions)

	srv := &Server{
		config:   cfg,
		mcp:      mcpServer,
		handlers: h,
	}
	registry := cfg.ProviderRegistry
	if registry == nil {
		builtinRegistry, err := builtin.NewRegistry(builtin.Deps{
			Search:     h.Search,
			Namespaces: h.Namespaces,
			Describe:   h.Describe,
			Examples:   h.Examples,
			Run:        h.Run,
			Chain:      h.Chain,
			Code:       h.Code,
		}, builtin.RegistryOptions{Providers: cfg.Providers})
		if err != nil {
			return nil, err
		}
		registry = builtinRegistry
	}
	mwAdapter, err := NewMiddlewareAdapterFromConfig(&cfg.Middleware)
	if err != nil {
		return nil, err
	}
	if err := mwAdapter.ApplyToProviders(registry); err != nil {
		return nil, err
	}
	adapter := NewProviderAdapter(registry)
	if err := adapter.RegisterTools(srv); err != nil {
		return nil, err
	}
	srv.registerToolListNotifications()
	return srv, nil
}

// MCPServer returns the underlying MCP SDK server.
func (s *Server) MCPServer() *mcp.Server {
	return s.mcp
}

// Run starts handling MCP requests over the provided transport.
func (s *Server) Run(ctx context.Context, transport mcp.Transport) error {
	return s.mcp.Run(ctx, transport)
}

// ListTools returns the registered MCP tools.
func (s *Server) ListTools() []*mcp.Tool {
	out := make([]*mcp.Tool, len(s.tools))
	copy(out, s.tools)
	return out
}

// Capabilities returns the server capabilities.
func (s *Server) Capabilities() Capabilities {
	return Capabilities{Tools: len(s.tools) > 0}
}

// Handlers returns the server's handlers for tool execution.
func (s *Server) Handlers() *Handlers {
	return s.handlers
}

func registerTool[In, Out any](s *Server, tool *mcp.Tool, handler mcp.ToolHandlerFor[In, Out]) {
	mcp.AddTool(s.mcp, tool, handler)
	s.tools = append(s.tools, tool)
	s.toolRegistrations = append(s.toolRegistrations, func() {
		mcp.AddTool(s.mcp, tool, handler)
	})
}

func (s *Server) registerToolListNotifications() {
	if !s.config.NotifyToolListChanged {
		return
	}
	changeNotifier, ok := s.config.Index.(index.ChangeNotifier)
	if !ok {
		return
	}
	debounce := time.Duration(s.config.NotifyToolListChangedDebounceMs) * time.Millisecond
	if debounce <= 0 {
		debounce = 150 * time.Millisecond
	}

	s.toolListUnsub = changeNotifier.OnChange(func(_ index.ChangeEvent) {
		s.toolListMu.Lock()
		defer s.toolListMu.Unlock()
		if s.toolListTimer == nil {
			s.toolListTimer = time.AfterFunc(debounce, s.reregisterTools)
			return
		}
		s.toolListTimer.Reset(debounce)
	})
}

func (s *Server) reregisterTools() {
	if len(s.toolRegistrations) == 0 {
		return
	}
	// Re-register a single tool to trigger one list_changed notification.
	s.toolRegistrations[0]()
}

// Close releases resources associated with the server.
func (s *Server) Close() error {
	s.toolListMu.Lock()
	if s.toolListTimer != nil {
		s.toolListTimer.Stop()
		s.toolListTimer = nil
	}
	s.toolListMu.Unlock()

	if s.toolListUnsub != nil {
		s.toolListUnsub()
		s.toolListUnsub = nil
	}
	return nil
}
