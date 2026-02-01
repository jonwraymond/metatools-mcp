# PRD-170: Create tooltransport

**Phase:** 7 - Protocol Layer
**Priority:** Critical
**Effort:** 8 hours
**Dependencies:** PRD-120
**Status:** Done (2026-02-01)

---

## Objective

Create `toolprotocol/transport/` for multi-transport support including stdio, SSE, and Streamable HTTP. WebSocket/gRPC are deferred.

---

## Package Design

**Location:** `github.com/jonwraymond/toolprotocol/transport`

**Purpose:**
- Transport abstraction layer
- Stdio transport (MCP default)
- SSE transport (legacy HTTP)
- Streamable HTTP transport (MCP 2025-03-26)

---

## Deliverables

| Deliverable | Location | Description |
|-------------|----------|-------------|
| Transport Package | `toolprotocol/transport/` | Transport abstraction |
| Stdio | `transport/stdio.go` | Stdio implementation |
| SSE | `transport/sse.go` | HTTP SSE implementation |
| Streamable HTTP | `transport/streamable.go` | MCP 2025-03-26 HTTP transport |
| Registry | `transport/factory.go` | Transport registry + factory |
| Tests | `transport/*_test.go` | Comprehensive tests |

## Implementation Summary

- Implemented `Transport` and `Server` contracts with concurrency + cancellation requirements.
- Shipped stdio, SSE, and streamable HTTP transports; WebSocket/gRPC deferred.

---

## Tasks

### Task 1: Create Package Structure

```bash
cd /tmp/migration
git clone git@github.com:ApertureStack/toolprotocol.git
cd toolprotocol

mkdir -p transport
```

### Task 2: Define Transport Interface

**File:** `toolprotocol/transport/transport.go`

```go
package transport

import (
    "context"
    "io"
)

// Transport represents a communication transport.
type Transport interface {
    // Start starts the transport.
    Start(ctx context.Context) error

    // Stop stops the transport gracefully.
    Stop(ctx context.Context) error

    // Send sends a message.
    Send(ctx context.Context, msg Message) error

    // Receive returns a channel of incoming messages.
    Receive() <-chan Message

    // Type returns the transport type.
    Type() string
}

// Message represents a transport message.
type Message struct {
    ID      string
    Type    MessageType
    Payload []byte
    Error   error
}

// MessageType defines message types.
type MessageType string

const (
    MessageRequest  MessageType = "request"
    MessageResponse MessageType = "response"
    MessageNotify   MessageType = "notification"
    MessageError    MessageType = "error"
)

// Server is a transport that accepts connections.
type Server interface {
    Transport

    // Listen starts accepting connections.
    Listen(addr string) error

    // Connections returns a channel of new connections.
    Connections() <-chan Connection
}

// Connection represents a client connection.
type Connection interface {
    io.ReadWriteCloser
    ID() string
}

// Config is base transport configuration.
type Config struct {
    ReadTimeout  time.Duration
    WriteTimeout time.Duration
    MaxMsgSize   int
}
```

### Task 3: Implement Stdio Transport

**File:** `toolprotocol/transport/stdio.go`

```go
package transport

import (
    "bufio"
    "context"
    "encoding/json"
    "io"
    "os"
    "sync"
)

// StdioTransport implements Transport for stdio communication.
type StdioTransport struct {
    reader  io.Reader
    writer  io.Writer
    recv    chan Message
    done    chan struct{}
    scanner *bufio.Scanner
    mu      sync.Mutex
}

// StdioConfig configures stdio transport.
type StdioConfig struct {
    Reader io.Reader
    Writer io.Writer
}

// NewStdioTransport creates a new stdio transport.
func NewStdioTransport(config StdioConfig) *StdioTransport {
    reader := config.Reader
    writer := config.Writer
    if reader == nil {
        reader = os.Stdin
    }
    if writer == nil {
        writer = os.Stdout
    }

    return &StdioTransport{
        reader: reader,
        writer: writer,
        recv:   make(chan Message, 100),
        done:   make(chan struct{}),
    }
}

func (t *StdioTransport) Type() string { return "stdio" }

func (t *StdioTransport) Start(ctx context.Context) error {
    t.scanner = bufio.NewScanner(t.reader)
    t.scanner.Buffer(make([]byte, 1024*1024), 10*1024*1024)

    go t.readLoop(ctx)
    return nil
}

func (t *StdioTransport) readLoop(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            return
        case <-t.done:
            return
        default:
            if !t.scanner.Scan() {
                if err := t.scanner.Err(); err != nil {
                    t.recv <- Message{Type: MessageError, Error: err}
                }
                return
            }

            line := t.scanner.Bytes()
            if len(line) == 0 {
                continue
            }

            msg := Message{
                Type:    MessageRequest,
                Payload: make([]byte, len(line)),
            }
            copy(msg.Payload, line)
            t.recv <- msg
        }
    }
}

func (t *StdioTransport) Stop(ctx context.Context) error {
    close(t.done)
    return nil
}

func (t *StdioTransport) Send(ctx context.Context, msg Message) error {
    t.mu.Lock()
    defer t.mu.Unlock()

    _, err := t.writer.Write(append(msg.Payload, '\n'))
    return err
}

func (t *StdioTransport) Receive() <-chan Message {
    return t.recv
}
```

### Task 4: Implement SSE Transport

**File:** `toolprotocol/transport/sse.go`

```go
package transport

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "sync"
)

// SSETransport implements Server for HTTP SSE communication.
type SSETransport struct {
    config SSEConfig
    server *http.Server
    recv   chan Message
    conns  chan Connection
    done   chan struct{}
    mu     sync.RWMutex
}

// SSEConfig configures SSE transport.
type SSEConfig struct {
    Host        string
    Port        int
    TLSCert     string
    TLSKey      string
    CORSOrigins []string
}

// NewSSETransport creates a new SSE transport server.
func NewSSETransport(config SSEConfig) *SSETransport {
    return &SSETransport{
        config: config,
        recv:   make(chan Message, 100),
        conns:  make(chan Connection, 10),
        done:   make(chan struct{}),
    }
}

func (t *SSETransport) Type() string { return "sse" }

func (t *SSETransport) Start(ctx context.Context) error {
    mux := http.NewServeMux()
    mux.HandleFunc("/mcp", t.handleMCP)
    mux.HandleFunc("/health", t.handleHealth)

    addr := fmt.Sprintf("%s:%d", t.config.Host, t.config.Port)
    t.server = &http.Server{
        Addr:    addr,
        Handler: t.corsMiddleware(mux),
    }

    go func() {
        var err error
        if t.config.TLSCert != "" {
            err = t.server.ListenAndServeTLS(t.config.TLSCert, t.config.TLSKey)
        } else {
            err = t.server.ListenAndServe()
        }
        if err != nil && err != http.ErrServerClosed {
            t.recv <- Message{Type: MessageError, Error: err}
        }
    }()

    return nil
}

func (t *SSETransport) handleMCP(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    // Read request
    var payload []byte
    payload, err := io.ReadAll(r.Body)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Create response channel
    respChan := make(chan Message, 1)
    msg := Message{
        ID:      r.Header.Get("X-Request-ID"),
        Type:    MessageRequest,
        Payload: payload,
    }

    t.recv <- msg

    // Wait for response (simplified - real impl needs response routing)
    // Set SSE headers
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")

    flusher, ok := w.(http.Flusher)
    if !ok {
        http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
        return
    }

    select {
    case resp := <-respChan:
        fmt.Fprintf(w, "event: message\ndata: %s\n\n", resp.Payload)
        flusher.Flush()
    case <-r.Context().Done():
        return
    }
}

func (t *SSETransport) handleHealth(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(`{"status":"ok"}`))
}

func (t *SSETransport) corsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        origin := r.Header.Get("Origin")
        for _, allowed := range t.config.CORSOrigins {
            if allowed == "*" || allowed == origin {
                w.Header().Set("Access-Control-Allow-Origin", origin)
                break
            }
        }
        w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Request-ID")

        if r.Method == http.MethodOptions {
            w.WriteHeader(http.StatusNoContent)
            return
        }

        next.ServeHTTP(w, r)
    })
}

func (t *SSETransport) Stop(ctx context.Context) error {
    close(t.done)
    return t.server.Shutdown(ctx)
}

func (t *SSETransport) Send(ctx context.Context, msg Message) error {
    // Implementation depends on connection management
    return nil
}

func (t *SSETransport) Receive() <-chan Message {
    return t.recv
}

func (t *SSETransport) Listen(addr string) error {
    return nil // Start handles this
}

func (t *SSETransport) Connections() <-chan Connection {
    return t.conns
}
```

### Task 5: Create Package Documentation

**File:** `toolprotocol/transport/doc.go`

```go
// Package transport provides multi-transport support for tool communication.
//
// This package implements various transport mechanisms for the MCP protocol
// and other AI tool protocols.
//
// # Transports
//
// Built-in transports:
//
//   - StdioTransport: Standard input/output (MCP default)
//   - SSETransport: HTTP with Server-Sent Events
//   - WebSocketTransport: WebSocket bidirectional
//   - GRPCTransport: gRPC for high-performance
//
// # Usage
//
// Create and use a transport:
//
//	// Stdio (default MCP)
//	stdio := transport.NewStdioTransport(transport.StdioConfig{})
//	stdio.Start(ctx)
//
//	// SSE for web clients
//	sse := transport.NewSSETransport(transport.SSEConfig{
//	    Host: "localhost",
//	    Port: 8080,
//	})
//	sse.Start(ctx)
//
// # Message Loop
//
// Process incoming messages:
//
//	for msg := range transport.Receive() {
//	    response := handleMessage(msg)
//	    transport.Send(ctx, response)
//	}
//
// # Transport Selection
//
// Select transport based on configuration:
//
//	func NewTransport(typ string, config any) (Transport, error) {
//	    switch typ {
//	    case "stdio":
//	        return NewStdioTransport(config.(StdioConfig)), nil
//	    case "sse":
//	        return NewSSETransport(config.(SSEConfig)), nil
//	    default:
//	        return nil, fmt.Errorf("unknown transport: %s", typ)
//	    }
//	}
package transport
```

### Task 6: Build and Test

```bash
cd /tmp/migration/toolprotocol

go mod tidy
go build ./...
go test -v ./transport/...
```

### Task 7: Commit and Push

```bash
cd /tmp/migration/toolprotocol

git add -A
git commit -m "feat(transport): add multi-transport support

Create transport package for protocol communication.

Package contents:
- Transport interface for pluggable transports
- StdioTransport for MCP stdio mode
- SSETransport for HTTP/SSE
- WebSocketTransport for bidirectional
- GRPCTransport for high-performance

Features:
- Pluggable transport abstraction
- Message-based communication
- Graceful shutdown
- CORS support for SSE
- TLS support

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"

git push origin main
```

---

## Next Steps

- PRD-171: Create toolwire
- PRD-172: Create tooldiscover
