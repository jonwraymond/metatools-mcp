# Pluggable Architecture Implementation Phases

**Status:** Draft (Revised)
**Date:** 2026-01-28
**Related:** [Pluggable Architecture Proposal](./pluggable-architecture.md), [Component Library Analysis](./component-library-analysis.md)

> **Revised Timeline:** Architecture discovery revealed that 13 extension points already exist as Go interfaces. The work is primarily **configuration and exposure**, not architecture redesign. Timeline reduced from 9 weeks to **6-7 weeks** (25% reduction).

## Overview

This document breaks the pluggable architecture proposal into manageable implementation phases, following Go best practices and layered architecture principles. Each phase is designed to be:

- **Independently deliverable** - Can be merged and deployed after completion
- **Backward compatible** - Existing functionality continues to work
- **Testable** - Includes clear verification criteria
- **Incremental** - Builds on previous phases

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    REVISED IMPLEMENTATION ROADMAP (6-7 weeks)               │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                               │
│   Phase 1 ─────────► Phase 2 ─────────► Phase 3 ─────────► Phase 4          │
│   CLI + Config       Transport          Public APIs        Backend           │
│   (Expose 13 pts)    Abstraction        & Documentation    Integration       │
│                                                                               │
│   ┌──────────────┐   ┌──────────────┐   ┌──────────────┐   ┌──────────────┐ │
│   │   2 weeks    │   │  1-2 weeks   │   │   1 week     │   │   2 weeks    │ │
│   │              │   │              │   │              │   │              │ │
│   │ Cobra CLI    │   │ Transport    │   │ Export       │   │ Docker/WASM  │ │
│   │ Koanf config │   │ Interface    │   │ internal pkg │   │ integration  │ │
│   │ 13 ext points│   │ Stdio + SSE  │   │ 13 ext docs  │   │ Backend reg  │ │
│   │ config schema│   │              │   │ Examples     │   │ config       │ │
│   └──────────────┘   └──────────────┘   └──────────────┘   └──────────────┘ │
│                                                                               │
│   MVP ◄───────────────────────────────► │                                    │
│   (Phases 1-2: ~3-4 weeks)              │                                    │
│                                                                               │
│   Note: Original Phase 5 (Middleware) deferred - 13 extension points        │
│         already provide middleware integration via ExecutionHook             │
│                                                                               │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Phase 1: CLI Framework & Configuration (Foundation)

**Duration:** ~2 weeks
**Priority:** Critical (foundation for all other phases)
**Risk:** Low (additive changes, no breaking changes)

### Objective

Introduce Cobra CLI framework and Koanf configuration library while maintaining backward compatibility with current environment variable configuration.

### Directory Structure Changes

```
cmd/metatools/
├── main.go              # Entry point (simplified)
├── root.go              # Root command
├── stdio.go             # `metatools stdio` (current behavior)
├── serve.go             # `metatools serve` (HTTP/SSE) - placeholder
├── version.go           # `metatools version`
├── validate.go          # `metatools validate` (config validation)
├── executor_toolruntime.go  # (unchanged)
└── executor_stub.go         # (unchanged)

internal/
├── config/
│   ├── env.go           # (unchanged - backward compat)
│   ├── loader.go        # NEW: Koanf-based config loader
│   ├── schema.go        # NEW: Config struct definitions
│   └── defaults.go      # NEW: Default values
└── ...
```

### Implementation Tasks

#### 1.1 Add Dependencies

```go
// go.mod additions
require (
    github.com/spf13/cobra v1.8.1
    github.com/knadh/koanf/v2 v2.1.2
    github.com/knadh/koanf/parsers/yaml v0.1.0
    github.com/knadh/koanf/providers/env v1.0.0
    github.com/knadh/koanf/providers/file v1.1.0
)
```

#### 1.2 Configuration Schema (`internal/config/schema.go`)

```go
package config

// ServerConfig is the root configuration structure
type ServerConfig struct {
    Server    ServerSettings    `koanf:"server"`
    Transport TransportConfig   `koanf:"transport"`
    Search    SearchConfig      `koanf:"search"`
    Execution ExecutionConfig   `koanf:"execution"`
    Providers ProvidersConfig   `koanf:"providers"`
    Backends  BackendsConfig    `koanf:"backends"`
    Middleware MiddlewareConfig `koanf:"middleware"`
}

type ServerSettings struct {
    Name    string `koanf:"name"`
    Version string `koanf:"version"`
}

type TransportConfig struct {
    Type string       `koanf:"type"` // stdio, sse, http
    HTTP HTTPConfig   `koanf:"http"`
}

type HTTPConfig struct {
    Host     string        `koanf:"host"`
    Port     int           `koanf:"port"`
    TLS      TLSConfig     `koanf:"tls"`
    Timeouts TimeoutConfig `koanf:"timeouts"`
    CORS     CORSConfig    `koanf:"cors"`
    Health   HealthConfig  `koanf:"health"`
}

// SearchConfig matches current EnvConfig.SearchConfig
type SearchConfig struct {
    Strategy       string `koanf:"strategy"`
    NameBoost      int    `koanf:"name_boost"`
    NamespaceBoost int    `koanf:"namespace_boost"`
    TagsBoost      int    `koanf:"tags_boost"`
    MaxDocs        int    `koanf:"max_docs"`
    MaxDocTextLen  int    `koanf:"max_doctext_len"`
}

// ... (additional config types)
```

#### 1.3 Configuration Loader (`internal/config/loader.go`)

```go
package config

import (
    "github.com/knadh/koanf/v2"
    "github.com/knadh/koanf/parsers/yaml"
    "github.com/knadh/koanf/providers/env"
    "github.com/knadh/koanf/providers/file"
)

type Loader struct {
    k *koanf.Koanf
}

// Load loads configuration with precedence:
// CLI flags > Environment variables > Config file > Defaults
func (l *Loader) Load(configPath string) (*ServerConfig, error) {
    l.k = koanf.New(".")

    // 1. Load defaults
    if err := l.loadDefaults(); err != nil {
        return nil, err
    }

    // 2. Load config file (if exists)
    if configPath != "" {
        if err := l.k.Load(file.Provider(configPath), yaml.Parser()); err != nil {
            return nil, fmt.Errorf("loading config file: %w", err)
        }
    }

    // 3. Load environment variables (METATOOLS_ prefix)
    if err := l.k.Load(env.Provider("METATOOLS_", ".", func(s string) string {
        return strings.Replace(strings.ToLower(
            strings.TrimPrefix(s, "METATOOLS_")), "_", ".", -1)
    }), nil); err != nil {
        return nil, fmt.Errorf("loading env vars: %w", err)
    }

    var cfg ServerConfig
    if err := l.k.Unmarshal("", &cfg); err != nil {
        return nil, fmt.Errorf("unmarshaling config: %w", err)
    }

    return &cfg, nil
}
```

#### 1.4 CLI Root Command (`cmd/metatools/root.go`)

```go
package main

import (
    "os"

    "github.com/spf13/cobra"
)

var (
    cfgFile string
    verbose bool
)

var rootCmd = &cobra.Command{
    Use:   "metatools",
    Short: "MCP server for tool discovery and execution",
    Long: `metatools-mcp is a Model Context Protocol server that provides
unified tool discovery, documentation, and execution capabilities.`,
}

func init() {
    rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "",
        "config file (default: metatools.yaml)")
    rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false,
        "verbose output")

    rootCmd.AddCommand(stdioCmd)
    rootCmd.AddCommand(serveCmd)
    rootCmd.AddCommand(versionCmd)
    rootCmd.AddCommand(validateCmd)
}

func Execute() {
    if err := rootCmd.Execute(); err != nil {
        os.Exit(1)
    }
}
```

#### 1.5 Stdio Command (`cmd/metatools/stdio.go`)

```go
package main

import (
    "context"
    "log"
    "os/signal"
    "syscall"

    "github.com/spf13/cobra"
    "github.com/modelcontextprotocol/go-sdk/mcp"
)

var stdioCmd = &cobra.Command{
    Use:   "stdio",
    Short: "Run as stdio MCP server (default mode)",
    Long:  `Runs the MCP server using stdin/stdout transport for MCP clients like Claude Desktop.`,
    RunE:  runStdio,
}

func init() {
    // Set as default command when no subcommand provided
    rootCmd.Run = stdioCmd.Run
}

func runStdio(cmd *cobra.Command, args []string) error {
    ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer cancel()

    srv, err := createServer()
    if err != nil {
        return fmt.Errorf("failed to create server: %w", err)
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
```

### Verification Criteria

- [ ] `metatools stdio` works identically to current behavior
- [ ] `metatools version` prints version info
- [ ] `metatools validate -c config.yaml` validates configuration files
- [ ] Existing environment variables (`METATOOLS_SEARCH_*`) continue to work
- [ ] New YAML config file loading works
- [ ] Config precedence: CLI > Env > File > Defaults
- [ ] All existing tests pass
- [ ] New unit tests for config loader (>80% coverage)

### Migration Notes

- No breaking changes
- `metatools` (no args) defaults to `metatools stdio`
- Environment variable prefix remains `METATOOLS_`

---

## Phase 2: Transport Layer Abstraction

**Duration:** ~2 weeks
**Priority:** High (enables multi-modal deployment)
**Risk:** Medium (touches core server logic)
**Depends on:** Phase 1

### Objective

Abstract the transport layer so the same server logic can run over stdio, SSE, or HTTP.

### Directory Structure Changes

```
internal/
├── transport/
│   ├── transport.go     # NEW: Transport interface
│   ├── registry.go      # NEW: Transport registry
│   ├── stdio.go         # NEW: Stdio transport wrapper
│   ├── sse.go           # NEW: SSE transport
│   └── http.go          # NEW: HTTP transport (optional)
├── server/
│   ├── server.go        # MODIFIED: Use Transport interface
│   └── handler.go       # NEW: Shared request handler
└── ...
```

### Implementation Tasks

#### 2.1 Transport Interface (`internal/transport/transport.go`)

```go
package transport

import (
    "context"
)

// Transport defines how MCP clients connect to the server
type Transport interface {
    // Name returns the transport identifier
    Name() string

    // Serve starts the transport and blocks until ctx is cancelled
    Serve(ctx context.Context, handler RequestHandler) error

    // Close gracefully shuts down the transport
    Close() error

    // Info returns runtime information about the transport
    Info() Info
}

// RequestHandler processes incoming MCP requests
type RequestHandler interface {
    HandleRequest(ctx context.Context, req *Request) (*Response, error)
}

// Info provides runtime details about a transport
type Info struct {
    Name      string
    Listening bool
    Address   string
    Metadata  map[string]string
}

// Config holds transport-specific configuration
type Config struct {
    Type      string
    HTTP      HTTPConfig
    WebSocket WebSocketConfig
    GRPC      GRPCConfig
}
```

#### 2.2 Transport Registry (`internal/transport/registry.go`)

```go
package transport

import (
    "fmt"
    "sync"
)

// Factory creates a configured transport instance
type Factory func(cfg Config) (Transport, error)

// Registry manages available transports
type Registry struct {
    mu         sync.RWMutex
    transports map[string]Factory
}

// NewRegistry creates a new transport registry with built-in transports
func NewRegistry() *Registry {
    r := &Registry{
        transports: make(map[string]Factory),
    }

    // Register built-in transports
    r.Register("stdio", NewStdioTransport)
    r.Register("sse", NewSSETransport)
    r.Register("http", NewHTTPTransport)

    return r
}

// Register adds a transport factory
func (r *Registry) Register(name string, factory Factory) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.transports[name] = factory
}

// Create instantiates a transport from config
func (r *Registry) Create(cfg Config) (Transport, error) {
    r.mu.RLock()
    factory, ok := r.transports[cfg.Type]
    r.mu.RUnlock()

    if !ok {
        return nil, fmt.Errorf("unknown transport type: %s", cfg.Type)
    }

    return factory(cfg)
}
```

#### 2.3 Stdio Transport (`internal/transport/stdio.go`)

```go
package transport

import (
    "context"

    "github.com/modelcontextprotocol/go-sdk/mcp"
)

// StdioTransport wraps the MCP SDK's stdio transport
type StdioTransport struct {
    sdk *mcp.StdioTransport
}

// NewStdioTransport creates a stdio transport
func NewStdioTransport(cfg Config) (Transport, error) {
    return &StdioTransport{
        sdk: &mcp.StdioTransport{},
    }, nil
}

func (t *StdioTransport) Name() string { return "stdio" }

func (t *StdioTransport) Serve(ctx context.Context, handler RequestHandler) error {
    // Adapt to MCP SDK's transport interface
    return t.sdk.Run(ctx, adaptHandler(handler))
}

func (t *StdioTransport) Close() error { return nil }

func (t *StdioTransport) Info() Info {
    return Info{
        Name:      "stdio",
        Listening: true,
        Address:   "stdin/stdout",
    }
}
```

#### 2.4 SSE Transport (`internal/transport/sse.go`)

```go
package transport

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "sync"
    "time"
)

// SSETransport implements Server-Sent Events transport
type SSETransport struct {
    cfg      HTTPConfig
    server   *http.Server
    handler  RequestHandler
    mu       sync.RWMutex
    clients  map[string]chan []byte
}

// NewSSETransport creates an SSE transport
func NewSSETransport(cfg Config) (Transport, error) {
    return &SSETransport{
        cfg:     cfg.HTTP,
        clients: make(map[string]chan []byte),
    }, nil
}

func (t *SSETransport) Name() string { return "sse" }

func (t *SSETransport) Serve(ctx context.Context, handler RequestHandler) error {
    t.handler = handler

    mux := http.NewServeMux()
    mux.HandleFunc("/mcp", t.handleMCP)
    mux.HandleFunc("/health", t.handleHealth)
    mux.HandleFunc("/ready", t.handleReady)

    addr := fmt.Sprintf("%s:%d", t.cfg.Host, t.cfg.Port)
    t.server = &http.Server{
        Addr:         addr,
        Handler:      t.applyCORS(mux),
        ReadTimeout:  t.cfg.Timeouts.Read,
        WriteTimeout: t.cfg.Timeouts.Write,
        IdleTimeout:  t.cfg.Timeouts.Idle,
    }

    // Start server in goroutine
    errCh := make(chan error, 1)
    go func() {
        if t.cfg.TLS.Enabled {
            errCh <- t.server.ListenAndServeTLS(t.cfg.TLS.Cert, t.cfg.TLS.Key)
        } else {
            errCh <- t.server.ListenAndServe()
        }
    }()

    // Wait for context cancellation or error
    select {
    case <-ctx.Done():
        shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()
        return t.server.Shutdown(shutdownCtx)
    case err := <-errCh:
        return err
    }
}

func (t *SSETransport) handleMCP(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    // Set SSE headers
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")

    // Parse request
    var req Request
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        t.sendError(w, "Invalid request", http.StatusBadRequest)
        return
    }

    // Handle request
    resp, err := t.handler.HandleRequest(r.Context(), &req)
    if err != nil {
        t.sendError(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Send response as SSE event
    t.sendEvent(w, "message", resp)
    t.sendEvent(w, "done", struct{}{})
}

func (t *SSETransport) sendEvent(w http.ResponseWriter, event string, data any) {
    jsonData, _ := json.Marshal(data)
    fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, jsonData)
    if f, ok := w.(http.Flusher); ok {
        f.Flush()
    }
}

// ... additional helper methods
```

#### 2.5 Server Modifications (`internal/server/server.go`)

```go
// Run now accepts our Transport interface
func (s *Server) Run(ctx context.Context, t transport.Transport) error {
    log.Printf("Starting server with %s transport", t.Name())
    return t.Serve(ctx, s)
}

// HandleRequest implements transport.RequestHandler
func (s *Server) HandleRequest(ctx context.Context, req *transport.Request) (*transport.Response, error) {
    // Delegate to MCP SDK's request handling
    return s.mcp.HandleRequest(ctx, req)
}
```

### Verification Criteria

- [ ] `metatools stdio` works identically to Phase 1
- [ ] `metatools serve --port 8080` starts SSE server
- [ ] SSE endpoint accepts POST requests at `/mcp`
- [ ] Health check endpoints work (`/health`, `/ready`)
- [ ] CORS headers applied correctly
- [ ] Graceful shutdown on SIGTERM
- [ ] TLS works when configured
- [ ] Integration tests for SSE transport
- [ ] Load test: 100 concurrent SSE connections

### CLI Changes

```bash
# New serve command
metatools serve [flags]

Flags:
  --port int          HTTP port (default 8080)
  --host string       Bind address (default "0.0.0.0")
  --tls               Enable TLS
  --cert string       TLS certificate path
  --key string        TLS key path
```

---

## Phase 3: Tool Provider Registry

**Duration:** ~1 week
**Priority:** High (enables plug-and-play tools)
**Risk:** Medium (refactors core registration logic)
**Depends on:** Phase 1

### Objective

Replace hardcoded tool registration with a registry pattern that allows dynamic tool registration.

### Directory Structure Changes

```
internal/
├── provider/
│   ├── provider.go      # NEW: ToolProvider interface
│   ├── registry.go      # NEW: Tool registry
│   ├── search.go        # NEW: search_tools provider
│   ├── describe.go      # NEW: describe_tool provider
│   ├── run.go           # NEW: run_tool provider
│   ├── chain.go         # NEW: run_chain provider
│   ├── namespaces.go    # NEW: list_namespaces provider
│   ├── examples.go      # NEW: list_tool_examples provider
│   └── code.go          # NEW: execute_code provider (optional)
├── server/
│   └── server.go        # MODIFIED: Use provider registry
└── ...
```

### Implementation Tasks

#### 3.1 Tool Provider Interface (`internal/provider/provider.go`)

```go
package provider

import (
    "context"

    "github.com/modelcontextprotocol/go-sdk/mcp"
)

// ToolProvider defines a pluggable MCP tool
type ToolProvider interface {
    // Name returns the tool name (e.g., "search_tools")
    Name() string

    // Tool returns the MCP tool definition with JSON schema
    Tool() *mcp.Tool

    // Handle executes the tool with the given input
    Handle(ctx context.Context, input map[string]any) (*mcp.CallToolResult, error)
}

// Option configures a tool provider
type Option func(ToolProvider)
```

#### 3.2 Tool Registry (`internal/provider/registry.go`)

```go
package provider

import (
    "fmt"
    "sync"

    "github.com/modelcontextprotocol/go-sdk/mcp"
)

// Registry manages registered tool providers
type Registry struct {
    mu        sync.RWMutex
    providers map[string]ToolProvider
    order     []string // Maintains registration order
}

// NewRegistry creates a new tool provider registry
func NewRegistry() *Registry {
    return &Registry{
        providers: make(map[string]ToolProvider),
        order:     make([]string, 0),
    }
}

// Register adds a tool provider
func (r *Registry) Register(p ToolProvider) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    name := p.Name()
    if _, exists := r.providers[name]; exists {
        return fmt.Errorf("provider already registered: %s", name)
    }

    r.providers[name] = p
    r.order = append(r.order, name)
    return nil
}

// Get retrieves a provider by name
func (r *Registry) Get(name string) (ToolProvider, bool) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    p, ok := r.providers[name]
    return p, ok
}

// All returns all registered providers in registration order
func (r *Registry) All() []ToolProvider {
    r.mu.RLock()
    defer r.mu.RUnlock()

    result := make([]ToolProvider, 0, len(r.order))
    for _, name := range r.order {
        result = append(result, r.providers[name])
    }
    return result
}

// Tools returns MCP tool definitions for all providers
func (r *Registry) Tools() []*mcp.Tool {
    providers := r.All()
    tools := make([]*mcp.Tool, 0, len(providers))
    for _, p := range providers {
        tools = append(tools, p.Tool())
    }
    return tools
}
```

#### 3.3 Example Provider (`internal/provider/search.go`)

```go
package provider

import (
    "context"

    "github.com/jonwraymond/metatools-mcp/internal/handlers"
    "github.com/modelcontextprotocol/go-sdk/mcp"
)

// SearchProvider implements the search_tools tool
type SearchProvider struct {
    handler *handlers.SearchHandler
}

// NewSearchProvider creates a search_tools provider
func NewSearchProvider(h *handlers.SearchHandler) *SearchProvider {
    return &SearchProvider{handler: h}
}

func (p *SearchProvider) Name() string { return "search_tools" }

func (p *SearchProvider) Tool() *mcp.Tool {
    return &mcp.Tool{
        Name:        "search_tools",
        Description: "Search for tools by query string. Returns ranked list of matching tools.",
        InputSchema: mcp.ToolInputSchema{
            Type: "object",
            Properties: map[string]any{
                "query": map[string]any{
                    "type":        "string",
                    "description": "Search query to find relevant tools",
                },
                "limit": map[string]any{
                    "type":        "integer",
                    "description": "Maximum number of results (default 10)",
                    "default":     10,
                },
            },
            Required: []string{"query"},
        },
    }
}

func (p *SearchProvider) Handle(ctx context.Context, input map[string]any) (*mcp.CallToolResult, error) {
    query, _ := input["query"].(string)
    limit := 10
    if l, ok := input["limit"].(float64); ok {
        limit = int(l)
    }

    return p.handler.Handle(ctx, query, limit)
}
```

#### 3.4 Server Integration

```go
// internal/server/server.go

func New(cfg config.Config, registry *provider.Registry) (*Server, error) {
    // ... existing validation

    mcpServer := mcp.NewServer(&mcp.Implementation{
        Name:    implementationName,
        Version: implementationVersion,
    }, &mcp.ServerOptions{
        PageSize: defaultPageSize,
    })

    srv := &Server{
        config:   cfg,
        mcp:      mcpServer,
        registry: registry,
    }

    // Register tools from provider registry
    for _, p := range registry.All() {
        srv.registerProvider(p)
    }

    return srv, nil
}

func (s *Server) registerProvider(p provider.ToolProvider) {
    tool := p.Tool()
    s.tools = append(s.tools, tool)

    s.mcp.AddTool(tool, func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        return p.Handle(ctx, req.Params.Arguments)
    })
}
```

### Verification Criteria

- [ ] All existing tools work via provider registry
- [ ] Provider registration order matches tool list order
- [ ] Custom providers can be registered programmatically
- [ ] Unit tests for registry (>90% coverage)
- [ ] Benchmark: Registration of 100 providers < 1ms

---

## Phase 4: Backend Registry

**Duration:** ~2 weeks
**Priority:** Medium (enables multi-source tool aggregation)
**Risk:** Medium (new subsystem)
**Depends on:** Phase 3

### Objective

Implement a backend registry that aggregates tools from multiple sources (local, MCP servers, HTTP APIs).

### Directory Structure Changes

```
internal/
├── backend/
│   ├── backend.go       # NEW: Backend interface
│   ├── registry.go      # NEW: Backend registry
│   ├── local.go         # NEW: Local file backend
│   ├── mcp.go           # NEW: MCP subprocess backend
│   ├── http.go          # NEW: HTTP API backend
│   ├── aggregator.go    # NEW: Tool aggregation logic
│   └── errors.go        # NEW: Backend-specific errors
└── ...
```

### Implementation Tasks

#### 4.1 Backend Interface (`internal/backend/backend.go`)

```go
package backend

import (
    "context"

    "github.com/jonwraymond/toolmodel"
)

// Backend defines a source of tools
type Backend interface {
    // Identity
    Kind() string  // e.g., "local", "mcp", "http"
    Name() string  // Instance name

    // Lifecycle
    Start(ctx context.Context) error
    Stop() error

    // Discovery
    ListTools(ctx context.Context) ([]toolmodel.Tool, error)

    // Execution
    Execute(ctx context.Context, tool string, args map[string]any) (any, error)
}

// Configurable backends support dynamic configuration
type Configurable interface {
    Configure(raw []byte) error
}
```

#### 4.2 Backend Registry (`internal/backend/registry.go`)

```go
package backend

import (
    "context"
    "fmt"
    "sync"
)

// Registry manages backend instances
type Registry struct {
    mu       sync.RWMutex
    backends map[string]Backend
    order    []string
}

// NewRegistry creates a backend registry
func NewRegistry() *Registry {
    return &Registry{
        backends: make(map[string]Backend),
        order:    make([]string, 0),
    }
}

// Register adds a backend
func (r *Registry) Register(name string, b Backend) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    if _, exists := r.backends[name]; exists {
        return fmt.Errorf("backend already registered: %s", name)
    }

    r.backends[name] = b
    r.order = append(r.order, name)
    return nil
}

// StartAll starts all registered backends
func (r *Registry) StartAll(ctx context.Context) error {
    r.mu.RLock()
    defer r.mu.RUnlock()

    for name, b := range r.backends {
        if err := b.Start(ctx); err != nil {
            return fmt.Errorf("starting backend %s: %w", name, err)
        }
    }
    return nil
}

// StopAll stops all registered backends
func (r *Registry) StopAll() error {
    r.mu.RLock()
    defer r.mu.RUnlock()

    var errs []error
    for name, b := range r.backends {
        if err := b.Stop(); err != nil {
            errs = append(errs, fmt.Errorf("stopping backend %s: %w", name, err))
        }
    }

    if len(errs) > 0 {
        return fmt.Errorf("errors stopping backends: %v", errs)
    }
    return nil
}
```

#### 4.3 MCP Backend (`internal/backend/mcp.go`)

```go
package backend

import (
    "context"
    "os/exec"

    "github.com/modelcontextprotocol/go-sdk/mcp"
)

// MCPBackend connects to an MCP subprocess
type MCPBackend struct {
    name    string
    command string
    args    []string
    env     map[string]string

    cmd    *exec.Cmd
    client *mcp.Client
}

// NewMCPBackend creates an MCP subprocess backend
func NewMCPBackend(name, command string, args []string, env map[string]string) *MCPBackend {
    return &MCPBackend{
        name:    name,
        command: command,
        args:    args,
        env:     env,
    }
}

func (b *MCPBackend) Kind() string { return "mcp" }
func (b *MCPBackend) Name() string { return b.name }

func (b *MCPBackend) Start(ctx context.Context) error {
    // Start subprocess and establish MCP connection
    // ... implementation
}

func (b *MCPBackend) ListTools(ctx context.Context) ([]toolmodel.Tool, error) {
    resp, err := b.client.ListTools(ctx)
    if err != nil {
        return nil, err
    }

    // Convert MCP tools to toolmodel.Tool
    tools := make([]toolmodel.Tool, 0, len(resp.Tools))
    for _, t := range resp.Tools {
        tools = append(tools, convertMCPTool(t, b.name))
    }
    return tools, nil
}

func (b *MCPBackend) Execute(ctx context.Context, tool string, args map[string]any) (any, error) {
    return b.client.CallTool(ctx, tool, args)
}
```

### Verification Criteria

- [ ] Local backend loads tools from file paths
- [ ] MCP backend spawns subprocess and communicates via stdio
- [ ] HTTP backend calls remote API endpoints
- [ ] Tool aggregation merges tools from all backends
- [ ] Backend failures don't crash the server (graceful degradation)
- [ ] Integration tests for each backend type
- [ ] E2E test: GitHub MCP server integration

---

## Phase 5: Middleware Chain

**Duration:** ~2 weeks
**Priority:** Low (nice-to-have for MVP)
**Risk:** Low (additive feature)
**Depends on:** Phase 3

### Objective

Implement a middleware chain for cross-cutting concerns (logging, auth, rate limiting, metrics).

### Directory Structure Changes

```
internal/
├── middleware/
│   ├── middleware.go    # NEW: Middleware interface
│   ├── chain.go         # NEW: Middleware chain builder
│   ├── registry.go      # NEW: Middleware registry
│   ├── logging.go       # NEW: Logging middleware
│   ├── auth.go          # NEW: Auth middleware
│   ├── ratelimit.go     # NEW: Rate limiting middleware
│   ├── metrics.go       # NEW: Metrics middleware
│   ├── cache.go         # NEW: Caching middleware
│   └── validation.go    # NEW: Input validation middleware
└── ...
```

### Implementation Tasks

#### 5.1 Middleware Interface (`internal/middleware/middleware.go`)

```go
package middleware

import (
    "context"

    "github.com/jonwraymond/metatools-mcp/internal/provider"
)

// Middleware wraps a ToolProvider with additional behavior
type Middleware func(provider.ToolProvider) provider.ToolProvider

// Factory creates a configured middleware instance
type Factory func(cfg Config) (Middleware, error)

// Config holds middleware-specific configuration
type Config struct {
    Name    string
    Enabled bool
    Raw     map[string]any
}
```

#### 5.2 Middleware Chain (`internal/middleware/chain.go`)

```go
package middleware

import (
    "github.com/jonwraymond/metatools-mcp/internal/provider"
)

// Chain applies middleware to providers in order
func Chain(middlewares []Middleware, p provider.ToolProvider) provider.ToolProvider {
    if len(middlewares) == 0 {
        return p
    }

    // Apply in reverse order so first middleware wraps outermost
    wrapped := p
    for i := len(middlewares) - 1; i >= 0; i-- {
        wrapped = middlewares[i](wrapped)
    }
    return wrapped
}

// ApplyToRegistry wraps all providers in a registry with middleware
func ApplyToRegistry(registry *provider.Registry, middlewares []Middleware) {
    for _, p := range registry.All() {
        wrapped := Chain(middlewares, p)
        // Replace in registry
        // ...
    }
}
```

#### 5.3 Logging Middleware (`internal/middleware/logging.go`)

```go
package middleware

import (
    "context"
    "log/slog"
    "time"

    "github.com/jonwraymond/metatools-mcp/internal/provider"
    "github.com/modelcontextprotocol/go-sdk/mcp"
)

type loggingMiddleware struct {
    next   provider.ToolProvider
    logger *slog.Logger
    level  slog.Level
}

// NewLoggingMiddleware creates a logging middleware
func NewLoggingMiddleware(logger *slog.Logger, level slog.Level) Middleware {
    return func(next provider.ToolProvider) provider.ToolProvider {
        return &loggingMiddleware{
            next:   next,
            logger: logger,
            level:  level,
        }
    }
}

func (m *loggingMiddleware) Name() string        { return m.next.Name() }
func (m *loggingMiddleware) Tool() *mcp.Tool     { return m.next.Tool() }

func (m *loggingMiddleware) Handle(ctx context.Context, input map[string]any) (*mcp.CallToolResult, error) {
    start := time.Now()

    m.logger.Log(ctx, m.level, "tool call started",
        "tool", m.next.Name(),
        "input", input,
    )

    result, err := m.next.Handle(ctx, input)

    m.logger.Log(ctx, m.level, "tool call completed",
        "tool", m.next.Name(),
        "duration", time.Since(start),
        "error", err,
    )

    return result, err
}
```

### Verification Criteria

- [ ] Logging middleware logs all tool calls
- [ ] Auth middleware validates tokens when configured
- [ ] Rate limiting middleware enforces limits
- [ ] Metrics middleware exposes Prometheus metrics
- [ ] Middleware chain order matches config order
- [ ] Unit tests for each middleware type
- [ ] Integration test: Full middleware chain

---

## Summary: Implementation Priority Matrix

| Phase | Priority | Risk | Duration | MVP? |
|-------|----------|------|----------|------|
| Phase 1: CLI + Config | Critical | Low | 2 weeks | ✓ |
| Phase 2: Transport | High | Medium | 2 weeks | ✓ |
| Phase 3: Tool Provider Registry | High | Medium | 1 week | ✓ |
| Phase 4: Backend Registry | Medium | Medium | 2 weeks | |
| Phase 5: Middleware Chain | Low | Low | 2 weeks | |

**MVP Timeline:** ~5 weeks (Phases 1-3)
**Full Implementation:** ~9 weeks (All phases)

---

## Appendix: Architectural Decisions

### Decision 1: Koanf over Viper

**Decision:** Use Koanf for configuration loading instead of Viper.

**Rationale:**
- Lighter dependency footprint
- Modular provider architecture
- Cleaner API for layered configuration
- Better suited for our use case

### Decision 2: Interface-based Transports over MCP SDK Extension

**Decision:** Define our own Transport interface that wraps MCP SDK transports.

**Rationale:**
- Decouples from MCP SDK implementation details
- Enables custom transports (Unix socket, gRPC)
- Allows consistent lifecycle management
- Facilitates testing with mock transports

### Decision 3: Provider Pattern over Direct Handler Registration

**Decision:** Use ToolProvider interface instead of direct MCP tool registration.

**Rationale:**
- Enables middleware wrapping
- Supports dynamic registration
- Cleaner separation of concerns
- Facilitates testing

### Decision 4: Decorator Pattern for Middleware

**Decision:** Use decorator pattern (wrapping) for middleware chain.

**Rationale:**
- Matches go-chi's proven approach
- Simple mental model
- Easy to compose and test
- Preserves ToolProvider interface

---

## Changelog

| Date | Change |
|------|--------|
| 2026-01-27 | Initial draft with 5 phases |
