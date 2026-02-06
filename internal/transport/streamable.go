package transport

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/jonwraymond/toolops/auth"
	"github.com/jonwraymond/toolops/health"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// StreamableHTTPConfig holds configuration for the Streamable HTTP transport.
//
// Streamable HTTP is the recommended HTTP transport per MCP spec 2025-11-25,
// replacing the deprecated SSE transport. It uses a single endpoint that handles:
//   - POST: JSON-RPC requests (returns JSON or SSE stream)
//   - GET: Opens server-to-client notification stream
//   - DELETE: Terminates session
//
// # Session Management
//
// By default, sessions are stateful with Mcp-Session-Id header tracking.
// Set Stateless=true for simpler deployments where session persistence
// is not required (e.g., single-request tools, serverless functions).
//
// # Response Modes
//
// By default, responses use SSE streaming (text/event-stream) for
// compatibility with long-running operations. Set JSONResponse=true
// to prefer application/json responses (per MCP spec §2.1.5).
type StreamableHTTPConfig struct {
	// Host is the network interface to bind (default: "0.0.0.0").
	Host string

	// Port is the TCP port to listen on (required for HTTP transports).
	Port int

	// Path is the HTTP endpoint path (default: "/mcp").
	Path string

	// ReadHeaderTimeout limits how long to wait for request headers.
	// Prevents slowloris attacks. Default: 10 seconds.
	ReadHeaderTimeout time.Duration

	// TLS enables HTTPS with certificate-based encryption.
	TLS TLSConfig

	// Stateless disables session management when true.
	// In stateless mode, no Mcp-Session-Id validation occurs and each
	// request uses temporary session parameters. Server-to-client requests
	// are rejected since clients cannot respond without a persistent session.
	// Notifications may still reach clients if sent during request handling.
	Stateless bool

	// JSONResponse causes responses to use application/json instead of
	// text/event-stream (SSE). This is useful for simple request/response
	// patterns without streaming, per MCP spec §2.1.5.
	JSONResponse bool

	// SessionTimeout configures idle session cleanup duration.
	// Sessions with no HTTP activity for this duration are automatically
	// closed. Zero means sessions never expire from inactivity.
	SessionTimeout time.Duration

	// HealthEnabled exposes the liveness endpoint when true.
	HealthEnabled bool

	// HealthPath is the HTTP path for the liveness endpoint.
	HealthPath string
}

// TLSConfig holds TLS/HTTPS configuration for secure transport.
//
// When Enabled is true, the transport serves HTTPS using the specified
// certificate and key files. TLS 1.2 is the minimum supported version.
type TLSConfig struct {
	// Enabled activates TLS encryption for the transport.
	Enabled bool

	// CertFile is the path to the PEM-encoded certificate file.
	CertFile string

	// KeyFile is the path to the PEM-encoded private key file.
	KeyFile string
}

// StreamableHTTPTransport implements the Transport interface for MCP's
// Streamable HTTP protocol (spec version 2025-11-25).
//
// This transport replaces the deprecated SSE transport and provides:
//   - Single endpoint handling POST/GET/DELETE methods
//   - Session management via Mcp-Session-Id header
//   - Bidirectional communication support
//   - Stream resumability (when EventStore is configured)
//   - Optional stateless mode for simpler deployments
//
// # Usage
//
//	transport := &StreamableHTTPTransport{
//	    Config: StreamableHTTPConfig{
//	        Host: "0.0.0.0",
//	        Port: 8080,
//	        Path: "/mcp",
//	    },
//	}
//	err := transport.Serve(ctx, server)
//
// # Protocol Flow
//
// 1. Client sends POST to /mcp with JSON-RPC initialize request
// 2. Server responds with session ID in Mcp-Session-Id header
// 3. Client includes session ID in subsequent requests
// 4. Client may open GET stream for server notifications
// 5. Client sends DELETE to terminate session
//
// # Concurrency
//
// StreamableHTTPTransport is safe for concurrent use. The underlying
// MCP SDK handler manages session state thread-safely.
type StreamableHTTPTransport struct {
	// Config holds the transport configuration.
	Config StreamableHTTPConfig

	mu       sync.Mutex
	listener net.Listener
	server   *http.Server
}

// Name returns "streamable" as the transport identifier.
func (t *StreamableHTTPTransport) Name() string {
	return "streamable"
}

// Info returns runtime information about the transport including
// the bound address and endpoint path. If the server is running,
// Info returns the actual bound address (useful when Port=0 for
// OS-assigned ports).
func (t *StreamableHTTPTransport) Info() Info {
	path := t.Config.Path
	if path == "" {
		path = "/mcp"
	}

	addr := ""
	t.mu.Lock()
	if t.listener != nil {
		addr = t.listener.Addr().String()
	}
	t.mu.Unlock()

	if addr == "" && t.Config.Port != 0 {
		host := t.Config.Host
		if host == "" {
			host = "0.0.0.0"
		}
		addr = fmt.Sprintf("%s:%d", host, t.Config.Port)
	}

	return Info{Name: "streamable", Addr: addr, Path: path}
}

// Serve starts the HTTP server and blocks until ctx is cancelled or
// an unrecoverable error occurs. The server handles MCP protocol
// messages via the configured endpoint path.
//
// When ctx is cancelled, Serve initiates graceful shutdown with a
// 5-second timeout for in-flight requests to complete.
//
// Serve returns nil on clean shutdown (context cancellation) or
// an error if the server fails to start or encounters a fatal error.
func (t *StreamableHTTPTransport) Serve(ctx context.Context, server Server) error {
	host := t.Config.Host
	if host == "" {
		host = "0.0.0.0"
	}
	path := t.Config.Path
	if path == "" {
		path = "/mcp"
	}
	addr := fmt.Sprintf("%s:%d", host, t.Config.Port)

	// Build SDK options from config
	opts := &mcp.StreamableHTTPOptions{
		Stateless:    t.Config.Stateless,
		JSONResponse: t.Config.JSONResponse,
	}
	if t.Config.SessionTimeout > 0 {
		opts.SessionTimeout = t.Config.SessionTimeout
	}

	// Create the Streamable HTTP handler from the MCP SDK.
	// The handler manages session lifecycle and protocol compliance.
	handler := mcp.NewStreamableHTTPHandler(func(_ *http.Request) *mcp.Server {
		return server.MCPServer()
	}, opts)

	mux := http.NewServeMux()
	// Wrap handler with auth headers middleware to extract HTTP headers into context
	mux.Handle(path, auth.WithAuthHeaders(handler))
	if t.Config.HealthEnabled {
		healthPath := t.Config.HealthPath
		if healthPath == "" {
			healthPath = "/healthz"
		}
		mux.HandleFunc(healthPath, health.LivenessHandler())
	}

	readHeaderTimeout := t.Config.ReadHeaderTimeout
	if readHeaderTimeout == 0 {
		readHeaderTimeout = 10 * time.Second
	}

	httpServer := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	// Configure TLS if enabled
	if t.Config.TLS.Enabled {
		cert, err := tls.LoadX509KeyPair(t.Config.TLS.CertFile, t.Config.TLS.KeyFile)
		if err != nil {
			return fmt.Errorf("load TLS certificate: %w", err)
		}
		httpServer.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
			MinVersion:   tls.VersionTLS12,
		}
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", addr, err)
	}

	t.mu.Lock()
	t.listener = ln
	t.server = httpServer
	t.mu.Unlock()

	errCh := make(chan error, 1)
	go func() {
		var serveErr error
		if t.Config.TLS.Enabled {
			// Use ServeTLS with empty cert/key since TLSConfig is pre-configured
			serveErr = httpServer.ServeTLS(ln, "", "")
		} else {
			serveErr = httpServer.Serve(ln)
		}
		if serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			errCh <- serveErr
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		_ = t.Close()
		return nil
	case err := <-errCh:
		if err == nil {
			return nil
		}
		return err
	}
}

// Close gracefully shuts down the HTTP server with a 5-second timeout.
// In-flight requests are given time to complete before forced termination.
//
// Close is idempotent and safe to call multiple times or before Serve.
func (t *StreamableHTTPTransport) Close() error {
	t.mu.Lock()
	srv := t.server
	ln := t.listener
	t.mu.Unlock()

	if srv == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := srv.Shutdown(ctx)
	if ln != nil {
		_ = ln.Close()
	}
	return err
}
