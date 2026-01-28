# PRD-004: SSE Transport Layer

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement the Server-Sent Events (SSE) transport layer enabling metatools-mcp to serve web clients and support concurrent connections.

**Architecture:** Create a Transport interface that abstracts protocol details. Implement SSETransport using net/http with proper SSE headers, event streaming, and connection management. The existing stdio behavior becomes StdioTransport implementing the same interface.

**Tech Stack:** net/http, SSE protocol (text/event-stream), existing MCP handler logic

**Priority:** P1 - Stream A, Phase 2 (enables web clients)

**Scope:** Transport abstraction + SSE implementation only - HTTP/REST deferred

**Dependencies:** PRD-002 (CLI), PRD-003 (Config)

---

## Context

With CLI and config foundation in place, we need HTTP transport for web clients. The MCP protocol supports SSE for server-to-client streaming. This enables:
- Web-based chat interfaces
- Multiple concurrent clients
- Load-balanced deployments
- Streaming responses

**Current State:** Only stdio transport, hardcoded

**Target State:**
```bash
metatools-mcp serve --transport=sse --port=8080
# Client connects: POST /mcp → SSE response stream
```

**MCP SSE Protocol:**
1. Client POSTs JSON-RPC request to `/mcp`
2. Server responds with `Content-Type: text/event-stream`
3. Server sends SSE events: `event: message\ndata: {...}\n\n`
4. Connection closes after response (or stays open for streaming tools)

---

## Tasks

### Task 1: Define Transport Interface

**Files:**
- Create: `internal/transport/transport.go`
- Test: `internal/transport/transport_test.go`

**Step 1: Write failing test for Transport interface**

```go
// internal/transport/transport_test.go
package transport

import (
    "context"
    "testing"
)

func TestTransportInterface(t *testing.T) {
    // Verify interface is defined correctly
    var _ Transport = (*StdioTransport)(nil)
    var _ Transport = (*SSETransport)(nil)
}

func TestTransportInfo(t *testing.T) {
    tests := []struct {
        name      string
        transport Transport
        wantName  string
    }{
        {
            name:      "stdio transport",
            transport: &StdioTransport{},
            wantName:  "stdio",
        },
        {
            name:      "sse transport",
            transport: &SSETransport{Config: SSEConfig{Port: 8080}},
            wantName:  "sse",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            info := tt.transport.Info()
            if info.Name != tt.wantName {
                t.Errorf("Info().Name = %q, want %q", info.Name, tt.wantName)
            }
        })
    }
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/jraymond/Documents/Projects/metatools-mcp && go test ./internal/transport/... -v`
Expected: FAIL - Transport type doesn't exist

**Step 3: Implement Transport interface**

```go
// internal/transport/transport.go
package transport

import (
    "context"
)

// Transport defines the interface for MCP protocol transports.
type Transport interface {
    // Name returns the transport identifier.
    Name() string

    // Serve starts the transport and blocks until context is cancelled.
    Serve(ctx context.Context, handler Handler) error

    // Close gracefully shuts down the transport.
    Close() error

    // Info returns transport metadata.
    Info() TransportInfo
}

// TransportInfo provides metadata about a transport.
type TransportInfo struct {
    Name     string
    Address  string // e.g., "localhost:8080" for HTTP, "stdio" for stdio
    Protocol string // e.g., "MCP/1.0"
}

// Handler processes MCP requests.
type Handler interface {
    // HandleRequest processes a single MCP request and returns the response.
    HandleRequest(ctx context.Context, request []byte) ([]byte, error)

    // HandleStream processes a streaming MCP request.
    // Returns a channel that will receive response events.
    HandleStream(ctx context.Context, request []byte) (<-chan []byte, error)
}

// TransportFactory creates transports from configuration.
type TransportFactory func(cfg interface{}) (Transport, error)

// Registry holds registered transport factories.
type Registry struct {
    factories map[string]TransportFactory
}

// NewRegistry creates a new transport registry.
func NewRegistry() *Registry {
    return &Registry{
        factories: make(map[string]TransportFactory),
    }
}

// Register adds a transport factory to the registry.
func (r *Registry) Register(name string, factory TransportFactory) {
    r.factories[name] = factory
}

// Create creates a transport by name.
func (r *Registry) Create(name string, cfg interface{}) (Transport, error) {
    factory, ok := r.factories[name]
    if !ok {
        return nil, ErrUnknownTransport{Name: name}
    }
    return factory(cfg)
}

// ErrUnknownTransport is returned when a transport is not found.
type ErrUnknownTransport struct {
    Name string
}

func (e ErrUnknownTransport) Error() string {
    return "unknown transport: " + e.Name
}
```

**Step 4: Run test to verify it passes**

Run: `cd /Users/jraymond/Documents/Projects/metatools-mcp && go test ./internal/transport/... -v`
Expected: PASS (after adding stub types in next tasks)

**Step 5: Commit**

```bash
git add internal/transport/transport.go internal/transport/transport_test.go
git commit -m "$(cat <<'EOF'
feat(transport): define Transport interface and registry

- Transport interface with Serve, Close, Info methods
- Handler interface for MCP request processing
- TransportFactory and Registry for pluggable transports

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 2: Implement StdioTransport

**Files:**
- Create: `internal/transport/stdio.go`
- Test: `internal/transport/stdio_test.go`

**Step 1: Write failing test for StdioTransport**

```go
// internal/transport/stdio_test.go
package transport

import (
    "bytes"
    "context"
    "io"
    "testing"
)

type mockHandler struct {
    response []byte
}

func (m *mockHandler) HandleRequest(ctx context.Context, request []byte) ([]byte, error) {
    return m.response, nil
}

func (m *mockHandler) HandleStream(ctx context.Context, request []byte) (<-chan []byte, error) {
    ch := make(chan []byte, 1)
    ch <- m.response
    close(ch)
    return ch, nil
}

func TestStdioTransport_Name(t *testing.T) {
    transport := &StdioTransport{}
    if got := transport.Name(); got != "stdio" {
        t.Errorf("Name() = %q, want %q", got, "stdio")
    }
}

func TestStdioTransport_Info(t *testing.T) {
    transport := &StdioTransport{}
    info := transport.Info()

    if info.Name != "stdio" {
        t.Errorf("Info().Name = %q, want %q", info.Name, "stdio")
    }
    if info.Address != "stdio" {
        t.Errorf("Info().Address = %q, want %q", info.Address, "stdio")
    }
}

func TestStdioTransport_Serve(t *testing.T) {
    // Create pipes for testing
    inputRead, inputWrite := io.Pipe()
    outputRead, outputWrite := io.Pipe()

    transport := &StdioTransport{
        Input:  inputRead,
        Output: outputWrite,
    }

    handler := &mockHandler{
        response: []byte(`{"jsonrpc":"2.0","result":"ok","id":1}`),
    }

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Start server in goroutine
    errCh := make(chan error, 1)
    go func() {
        errCh <- transport.Serve(ctx, handler)
    }()

    // Send a request
    request := []byte(`{"jsonrpc":"2.0","method":"test","id":1}`)
    go func() {
        inputWrite.Write(request)
        inputWrite.Close()
    }()

    // Read response
    response, err := io.ReadAll(outputRead)
    if err != nil {
        t.Fatalf("ReadAll() error = %v", err)
    }

    if !bytes.Contains(response, []byte("ok")) {
        t.Errorf("Response should contain 'ok', got: %s", response)
    }

    // Stop server
    cancel()
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/jraymond/Documents/Projects/metatools-mcp && go test ./internal/transport/... -run TestStdioTransport -v`
Expected: FAIL - StdioTransport not implemented

**Step 3: Implement StdioTransport**

```go
// internal/transport/stdio.go
package transport

import (
    "bufio"
    "context"
    "io"
    "os"
    "sync"
)

// StdioTransport implements Transport for stdio-based MCP communication.
type StdioTransport struct {
    Input  io.Reader
    Output io.Writer

    mu     sync.Mutex
    closed bool
}

// NewStdioTransport creates a stdio transport using os.Stdin and os.Stdout.
func NewStdioTransport() *StdioTransport {
    return &StdioTransport{
        Input:  os.Stdin,
        Output: os.Stdout,
    }
}

// Name returns "stdio".
func (t *StdioTransport) Name() string {
    return "stdio"
}

// Info returns transport metadata.
func (t *StdioTransport) Info() TransportInfo {
    return TransportInfo{
        Name:     "stdio",
        Address:  "stdio",
        Protocol: "MCP/1.0",
    }
}

// Serve reads requests from stdin and writes responses to stdout.
func (t *StdioTransport) Serve(ctx context.Context, handler Handler) error {
    reader := bufio.NewReader(t.Input)

    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
        }

        // Read line (JSON-RPC messages are newline-delimited)
        line, err := reader.ReadBytes('\n')
        if err != nil {
            if err == io.EOF {
                return nil // Normal termination
            }
            return err
        }

        // Skip empty lines
        if len(line) <= 1 {
            continue
        }

        // Process request
        response, err := handler.HandleRequest(ctx, line)
        if err != nil {
            // Log error but continue processing
            continue
        }

        // Write response
        t.mu.Lock()
        _, err = t.Output.Write(response)
        if err == nil {
            _, err = t.Output.Write([]byte("\n"))
        }
        t.mu.Unlock()

        if err != nil {
            return err
        }
    }
}

// Close marks the transport as closed.
func (t *StdioTransport) Close() error {
    t.mu.Lock()
    defer t.mu.Unlock()
    t.closed = true
    return nil
}
```

**Step 4: Run test to verify it passes**

Run: `cd /Users/jraymond/Documents/Projects/metatools-mcp && go test ./internal/transport/... -run TestStdioTransport -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/transport/stdio.go internal/transport/stdio_test.go
git commit -m "$(cat <<'EOF'
feat(transport): implement StdioTransport

- Read JSON-RPC messages from stdin
- Write responses to stdout
- Support context cancellation

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 3: Implement SSETransport Configuration

**Files:**
- Create: `internal/transport/sse_config.go`
- Test: `internal/transport/sse_config_test.go`

**Step 1: Write failing test for SSE config**

```go
// internal/transport/sse_config_test.go
package transport

import (
    "testing"
    "time"
)

func TestSSEConfig_Validate(t *testing.T) {
    tests := []struct {
        name    string
        config  SSEConfig
        wantErr bool
    }{
        {
            name:    "valid config",
            config:  SSEConfig{Host: "0.0.0.0", Port: 8080},
            wantErr: false,
        },
        {
            name:    "zero port invalid",
            config:  SSEConfig{Host: "0.0.0.0", Port: 0},
            wantErr: true,
        },
        {
            name:    "port too high",
            config:  SSEConfig{Host: "0.0.0.0", Port: 70000},
            wantErr: true,
        },
        {
            name:    "negative timeout",
            config:  SSEConfig{Host: "0.0.0.0", Port: 8080, ReadTimeout: -1 * time.Second},
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.config.Validate()
            if (err != nil) != tt.wantErr {
                t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}

func TestSSEConfig_Defaults(t *testing.T) {
    cfg := DefaultSSEConfig()

    if cfg.Port != 8080 {
        t.Errorf("Port = %d, want %d", cfg.Port, 8080)
    }
    if cfg.ReadTimeout != 30*time.Second {
        t.Errorf("ReadTimeout = %v, want %v", cfg.ReadTimeout, 30*time.Second)
    }
    if cfg.WriteTimeout != 30*time.Second {
        t.Errorf("WriteTimeout = %v, want %v", cfg.WriteTimeout, 30*time.Second)
    }
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/jraymond/Documents/Projects/metatools-mcp && go test ./internal/transport/... -run TestSSEConfig -v`
Expected: FAIL - SSEConfig doesn't exist

**Step 3: Implement SSEConfig**

```go
// internal/transport/sse_config.go
package transport

import (
    "errors"
    "fmt"
    "time"
)

// SSEConfig holds SSE transport configuration.
type SSEConfig struct {
    // Host to bind to (default: "0.0.0.0")
    Host string

    // Port to listen on (default: 8080)
    Port int

    // ReadTimeout for incoming requests
    ReadTimeout time.Duration

    // WriteTimeout for outgoing responses
    WriteTimeout time.Duration

    // IdleTimeout for keep-alive connections
    IdleTimeout time.Duration

    // CORSOrigins allowed for cross-origin requests
    CORSOrigins []string

    // TLS configuration
    TLSCert string
    TLSKey  string

    // MaxRequestSize in bytes (default: 1MB)
    MaxRequestSize int64
}

// DefaultSSEConfig returns sensible defaults for SSE transport.
func DefaultSSEConfig() SSEConfig {
    return SSEConfig{
        Host:           "0.0.0.0",
        Port:           8080,
        ReadTimeout:    30 * time.Second,
        WriteTimeout:   30 * time.Second,
        IdleTimeout:    120 * time.Second,
        CORSOrigins:    []string{"*"},
        MaxRequestSize: 1 << 20, // 1MB
    }
}

// Validate checks the configuration for errors.
func (c *SSEConfig) Validate() error {
    if c.Port <= 0 || c.Port > 65535 {
        return fmt.Errorf("invalid port %d, must be 1-65535", c.Port)
    }
    if c.ReadTimeout < 0 {
        return errors.New("read timeout cannot be negative")
    }
    if c.WriteTimeout < 0 {
        return errors.New("write timeout cannot be negative")
    }
    if c.IdleTimeout < 0 {
        return errors.New("idle timeout cannot be negative")
    }
    return nil
}

// Address returns the listen address as "host:port".
func (c *SSEConfig) Address() string {
    return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// WithDefaults fills zero values with defaults.
func (c SSEConfig) WithDefaults() SSEConfig {
    defaults := DefaultSSEConfig()

    if c.Host == "" {
        c.Host = defaults.Host
    }
    if c.Port == 0 {
        c.Port = defaults.Port
    }
    if c.ReadTimeout == 0 {
        c.ReadTimeout = defaults.ReadTimeout
    }
    if c.WriteTimeout == 0 {
        c.WriteTimeout = defaults.WriteTimeout
    }
    if c.IdleTimeout == 0 {
        c.IdleTimeout = defaults.IdleTimeout
    }
    if len(c.CORSOrigins) == 0 {
        c.CORSOrigins = defaults.CORSOrigins
    }
    if c.MaxRequestSize == 0 {
        c.MaxRequestSize = defaults.MaxRequestSize
    }

    return c
}
```

**Step 4: Run test to verify it passes**

Run: `cd /Users/jraymond/Documents/Projects/metatools-mcp && go test ./internal/transport/... -run TestSSEConfig -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/transport/sse_config.go internal/transport/sse_config_test.go
git commit -m "$(cat <<'EOF'
feat(transport): add SSEConfig with validation and defaults

- Host, Port, Timeouts, CORS, TLS settings
- Validation for required fields
- WithDefaults() for zero-value handling

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 4: Implement SSETransport Server

**Files:**
- Create: `internal/transport/sse.go`
- Test: `internal/transport/sse_test.go`

**Step 1: Write failing test for SSE transport**

```go
// internal/transport/sse_test.go
package transport

import (
    "bytes"
    "context"
    "io"
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"
    "time"
)

func TestSSETransport_Name(t *testing.T) {
    transport := NewSSETransport(SSEConfig{Port: 8080})
    if got := transport.Name(); got != "sse" {
        t.Errorf("Name() = %q, want %q", got, "sse")
    }
}

func TestSSETransport_Info(t *testing.T) {
    transport := NewSSETransport(SSEConfig{Host: "localhost", Port: 9090})
    info := transport.Info()

    if info.Name != "sse" {
        t.Errorf("Info().Name = %q, want %q", info.Name, "sse")
    }
    if info.Address != "localhost:9090" {
        t.Errorf("Info().Address = %q, want %q", info.Address, "localhost:9090")
    }
}

func TestSSETransport_HandleRequest(t *testing.T) {
    handler := &mockHandler{
        response: []byte(`{"jsonrpc":"2.0","result":"hello","id":1}`),
    }

    transport := NewSSETransport(SSEConfig{Port: 0}) // Port 0 for testing

    // Create test request
    body := bytes.NewReader([]byte(`{"jsonrpc":"2.0","method":"test","id":1}`))
    req := httptest.NewRequest(http.MethodPost, "/mcp", body)
    req.Header.Set("Content-Type", "application/json")

    rec := httptest.NewRecorder()

    // Handle request directly (without starting server)
    transport.handleMCP(rec, req, handler)

    resp := rec.Result()
    defer resp.Body.Close()

    // Check SSE headers
    if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "text/event-stream") {
        t.Errorf("Content-Type = %q, want text/event-stream", ct)
    }

    // Check response body
    respBody, _ := io.ReadAll(resp.Body)
    if !bytes.Contains(respBody, []byte("hello")) {
        t.Errorf("Response should contain 'hello', got: %s", respBody)
    }
    if !bytes.Contains(respBody, []byte("event: message")) {
        t.Errorf("Response should contain SSE event, got: %s", respBody)
    }
}

func TestSSETransport_CORS(t *testing.T) {
    handler := &mockHandler{
        response: []byte(`{"jsonrpc":"2.0","result":"ok","id":1}`),
    }

    transport := NewSSETransport(SSEConfig{
        Port:        0,
        CORSOrigins: []string{"https://example.com"},
    })

    // OPTIONS request (CORS preflight)
    req := httptest.NewRequest(http.MethodOptions, "/mcp", nil)
    req.Header.Set("Origin", "https://example.com")
    rec := httptest.NewRecorder()

    transport.handleMCP(rec, req, handler)

    resp := rec.Result()
    if resp.Header.Get("Access-Control-Allow-Origin") != "https://example.com" {
        t.Errorf("CORS header not set correctly")
    }
}

func TestSSETransport_Serve(t *testing.T) {
    handler := &mockHandler{
        response: []byte(`{"jsonrpc":"2.0","result":"test","id":1}`),
    }

    transport := NewSSETransport(SSEConfig{
        Host: "127.0.0.1",
        Port: 0, // Random available port
    })

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Start server in goroutine
    errCh := make(chan error, 1)
    go func() {
        errCh <- transport.Serve(ctx, handler)
    }()

    // Wait for server to start
    time.Sleep(100 * time.Millisecond)

    // Server should be running (we can't easily test the full flow here)
    // Just verify it starts without error
    cancel()

    select {
    case err := <-errCh:
        if err != nil && err != context.Canceled && err != http.ErrServerClosed {
            t.Errorf("Serve() error = %v", err)
        }
    case <-time.After(time.Second):
        t.Error("Serve() did not return after cancel")
    }
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/jraymond/Documents/Projects/metatools-mcp && go test ./internal/transport/... -run TestSSETransport -v`
Expected: FAIL - SSETransport not implemented

**Step 3: Implement SSETransport**

```go
// internal/transport/sse.go
package transport

import (
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net"
    "net/http"
    "sync"
    "time"
)

// SSETransport implements Transport for HTTP/SSE-based MCP communication.
type SSETransport struct {
    Config SSEConfig

    server   *http.Server
    listener net.Listener
    mu       sync.Mutex
}

// NewSSETransport creates a new SSE transport with the given configuration.
func NewSSETransport(cfg SSEConfig) *SSETransport {
    return &SSETransport{
        Config: cfg.WithDefaults(),
    }
}

// Name returns "sse".
func (t *SSETransport) Name() string {
    return "sse"
}

// Info returns transport metadata.
func (t *SSETransport) Info() TransportInfo {
    return TransportInfo{
        Name:     "sse",
        Address:  t.Config.Address(),
        Protocol: "MCP/1.0",
    }
}

// Serve starts the HTTP server and blocks until context is cancelled.
func (t *SSETransport) Serve(ctx context.Context, handler Handler) error {
    mux := http.NewServeMux()

    // MCP endpoint
    mux.HandleFunc("/mcp", func(w http.ResponseWriter, r *http.Request) {
        t.handleMCP(w, r, handler)
    })

    // Health check
    mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
    })

    t.server = &http.Server{
        Addr:         t.Config.Address(),
        Handler:      mux,
        ReadTimeout:  t.Config.ReadTimeout,
        WriteTimeout: t.Config.WriteTimeout,
        IdleTimeout:  t.Config.IdleTimeout,
    }

    // Start listening
    var err error
    t.listener, err = net.Listen("tcp", t.Config.Address())
    if err != nil {
        return fmt.Errorf("listen: %w", err)
    }

    // Handle graceful shutdown
    go func() {
        <-ctx.Done()
        shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        t.server.Shutdown(shutdownCtx)
    }()

    // Serve requests
    if t.Config.TLSCert != "" && t.Config.TLSKey != "" {
        err = t.server.ServeTLS(t.listener, t.Config.TLSCert, t.Config.TLSKey)
    } else {
        err = t.server.Serve(t.listener)
    }

    if err == http.ErrServerClosed {
        return nil
    }
    return err
}

// Close shuts down the server.
func (t *SSETransport) Close() error {
    t.mu.Lock()
    defer t.mu.Unlock()

    if t.server != nil {
        return t.server.Close()
    }
    return nil
}

// handleMCP processes MCP requests with SSE response.
func (t *SSETransport) handleMCP(w http.ResponseWriter, r *http.Request, handler Handler) {
    // Handle CORS
    t.setCORSHeaders(w, r)
    if r.Method == http.MethodOptions {
        w.WriteHeader(http.StatusOK)
        return
    }

    // Only accept POST
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    // Read request body
    body, err := io.ReadAll(io.LimitReader(r.Body, t.Config.MaxRequestSize))
    if err != nil {
        http.Error(w, "Failed to read request", http.StatusBadRequest)
        return
    }
    defer r.Body.Close()

    // Set SSE headers
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")
    w.Header().Set("X-Accel-Buffering", "no") // Disable nginx buffering

    // Flush headers
    if f, ok := w.(http.Flusher); ok {
        f.Flush()
    }

    // Process request
    response, err := handler.HandleRequest(r.Context(), body)
    if err != nil {
        t.writeSSEEvent(w, "error", []byte(fmt.Sprintf(`{"error":"%s"}`, err.Error())))
        return
    }

    // Write response as SSE event
    t.writeSSEEvent(w, "message", response)

    // Write done event
    t.writeSSEEvent(w, "done", []byte("{}"))
}

// writeSSEEvent writes a single SSE event.
func (t *SSETransport) writeSSEEvent(w http.ResponseWriter, event string, data []byte) {
    fmt.Fprintf(w, "event: %s\n", event)
    fmt.Fprintf(w, "data: %s\n\n", data)

    if f, ok := w.(http.Flusher); ok {
        f.Flush()
    }
}

// setCORSHeaders sets CORS headers based on configuration.
func (t *SSETransport) setCORSHeaders(w http.ResponseWriter, r *http.Request) {
    origin := r.Header.Get("Origin")
    if origin == "" {
        return
    }

    // Check if origin is allowed
    allowed := false
    for _, o := range t.Config.CORSOrigins {
        if o == "*" || o == origin {
            allowed = true
            break
        }
    }

    if allowed {
        w.Header().Set("Access-Control-Allow-Origin", origin)
        w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        w.Header().Set("Access-Control-Max-Age", "86400")
    }
}
```

**Step 4: Run test to verify it passes**

Run: `cd /Users/jraymond/Documents/Projects/metatools-mcp && go test ./internal/transport/... -run TestSSETransport -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/transport/sse.go internal/transport/sse_test.go
git commit -m "$(cat <<'EOF'
feat(transport): implement SSETransport with HTTP server

- POST /mcp endpoint with SSE response
- CORS support with configurable origins
- Health check endpoint at /health
- TLS support
- Graceful shutdown

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 5: Integrate Transports with Serve Command

**Files:**
- Modify: `cmd/metatools-mcp/cmd/serve.go`
- Test: `cmd/metatools-mcp/cmd/serve_test.go`

**Step 1: Update serve.go to use Transport interface**

```go
// Update serve.go imports
import (
    "github.com/your-org/metatools-mcp/internal/transport"
    // ... existing imports
)

// Update runServe to use Transport
func runServe(ctx context.Context, cliCfg *ServeConfig) error {
    ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
    defer cancel()

    // Load configuration
    cfg, err := loadServeConfig(cliCfg.Config, cliCfg)
    if err != nil {
        return fmt.Errorf("load config: %w", err)
    }

    // Build server
    serverCfg, err := buildServerConfigFromConfig(cfg)
    if err != nil {
        return fmt.Errorf("build server config: %w", err)
    }

    srv, err := server.New(serverCfg)
    if err != nil {
        return fmt.Errorf("create server: %w", err)
    }

    // Create transport based on config
    var trans transport.Transport
    switch cfg.Transport.Type {
    case "stdio":
        trans = transport.NewStdioTransport()
    case "sse":
        trans = transport.NewSSETransport(transport.SSEConfig{
            Host:        cfg.Transport.HTTP.Host,
            Port:        cfg.Transport.HTTP.Port,
            TLSCert:     cfg.Transport.HTTP.TLS.CertFile,
            TLSKey:      cfg.Transport.HTTP.TLS.KeyFile,
            CORSOrigins: []string{"*"}, // TODO: Add to config
        })
    case "http":
        return fmt.Errorf("transport %q not yet implemented", cfg.Transport.Type)
    default:
        return fmt.Errorf("unknown transport: %s", cfg.Transport.Type)
    }

    // Log startup
    info := trans.Info()
    fmt.Fprintf(os.Stderr, "Starting %s\n", cfg.Server.Name)
    fmt.Fprintf(os.Stderr, "  Transport: %s\n", info.Name)
    fmt.Fprintf(os.Stderr, "  Address:   %s\n", info.Address)

    // Create handler adapter
    handler := &serverHandler{server: srv}

    // Start server
    return trans.Serve(ctx, handler)
}

// serverHandler adapts the MCP server to the Transport Handler interface.
type serverHandler struct {
    server *server.Server
}

func (h *serverHandler) HandleRequest(ctx context.Context, request []byte) ([]byte, error) {
    // TODO: Forward to actual MCP server
    // This is a placeholder that will be connected to the real server
    return request, nil
}

func (h *serverHandler) HandleStream(ctx context.Context, request []byte) (<-chan []byte, error) {
    ch := make(chan []byte, 1)
    response, err := h.HandleRequest(ctx, request)
    if err != nil {
        close(ch)
        return ch, err
    }
    ch <- response
    close(ch)
    return ch, nil
}
```

**Step 2: Run test to verify it works**

Run: `cd /Users/jraymond/Documents/Projects/metatools-mcp && go build ./cmd/metatools-mcp/... && ./cmd/metatools-mcp/metatools-mcp serve --help`
Expected: Shows help with transport options

**Step 3: Commit**

```bash
git add cmd/metatools-mcp/cmd/serve.go
git commit -m "$(cat <<'EOF'
feat(cli): integrate Transport interface with serve command

- Create stdio or SSE transport based on config
- Add serverHandler adapter for MCP server
- Log transport info on startup

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 6: Add Integration Test

**Files:**
- Create: `internal/transport/integration_test.go`

**Step 1: Write integration test**

```go
//go:build integration

// internal/transport/integration_test.go
package transport

import (
    "bytes"
    "context"
    "encoding/json"
    "io"
    "net/http"
    "testing"
    "time"
)

func TestIntegration_SSETransport(t *testing.T) {
    handler := &mockHandler{
        response: []byte(`{"jsonrpc":"2.0","result":{"message":"hello"},"id":1}`),
    }

    transport := NewSSETransport(SSEConfig{
        Host: "127.0.0.1",
        Port: 19090, // Use unusual port for testing
    })

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Start server
    errCh := make(chan error, 1)
    go func() {
        errCh <- transport.Serve(ctx, handler)
    }()

    // Wait for server to start
    time.Sleep(200 * time.Millisecond)

    // Send request
    client := &http.Client{Timeout: 5 * time.Second}

    reqBody := []byte(`{"jsonrpc":"2.0","method":"test","params":{},"id":1}`)
    resp, err := client.Post(
        "http://127.0.0.1:19090/mcp",
        "application/json",
        bytes.NewReader(reqBody),
    )
    if err != nil {
        t.Fatalf("POST error = %v", err)
    }
    defer resp.Body.Close()

    // Check response
    if resp.StatusCode != http.StatusOK {
        t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusOK)
    }

    if ct := resp.Header.Get("Content-Type"); ct != "text/event-stream" {
        t.Errorf("Content-Type = %q, want text/event-stream", ct)
    }

    // Read SSE response
    body, _ := io.ReadAll(resp.Body)
    if !bytes.Contains(body, []byte("hello")) {
        t.Errorf("Response should contain 'hello', got: %s", body)
    }

    // Test health endpoint
    healthResp, err := client.Get("http://127.0.0.1:19090/health")
    if err != nil {
        t.Fatalf("GET /health error = %v", err)
    }
    defer healthResp.Body.Close()

    var health map[string]string
    json.NewDecoder(healthResp.Body).Decode(&health)
    if health["status"] != "ok" {
        t.Errorf("Health status = %q, want 'ok'", health["status"])
    }

    // Shutdown
    cancel()
}
```

**Step 2: Run integration test**

Run: `cd /Users/jraymond/Documents/Projects/metatools-mcp && go test ./internal/transport/... -tags=integration -run TestIntegration -v`
Expected: PASS

**Step 3: Commit**

```bash
git add internal/transport/integration_test.go
git commit -m "$(cat <<'EOF'
test(transport): add SSE transport integration test

- Test full HTTP request/response cycle
- Test SSE headers and event format
- Test health endpoint

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Verification Checklist

- [ ] Transport interface defined
- [ ] StdioTransport implemented
- [ ] SSEConfig with validation
- [ ] SSETransport with HTTP server
- [ ] CORS support
- [ ] Health endpoint
- [ ] Serve command integration
- [ ] Integration tests pass

## Definition of Done

1. All tests pass: `go test ./internal/transport/...`
2. Integration tests pass: `go test ./internal/transport/... -tags=integration`
3. `metatools-mcp serve --transport=stdio` works as before
4. `metatools-mcp serve --transport=sse --port=8080` starts HTTP server
5. `curl -X POST http://localhost:8080/mcp -d '{...}'` returns SSE response
6. `curl http://localhost:8080/health` returns OK

## Next PRD

PRD-005 will implement the Tool Provider Registry, enabling plug-and-play tool registration.
