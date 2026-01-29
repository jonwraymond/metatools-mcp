package transport

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// SSEConfig holds SSE transport settings.
type SSEConfig struct {
	Host              string
	Port              int
	Path              string
	ReadHeaderTimeout time.Duration
}

// SSETransport serves MCP over Server-Sent Events.
type SSETransport struct {
	Config SSEConfig

	mu       sync.Mutex
	listener net.Listener
	server   *http.Server
}

// Name returns the transport identifier.
func (t *SSETransport) Name() string {
	return "sse"
}

// Info returns descriptive information about the transport.
func (t *SSETransport) Info() Info {
	path := t.Config.Path
	if path == "" {
		path = "/mcp"
	}
	addr := ""
	if t.listener != nil {
		addr = t.listener.Addr().String()
	} else if t.Config.Port != 0 {
		host := t.Config.Host
		if host == "" {
			host = "0.0.0.0"
		}
		addr = fmt.Sprintf("%s:%d", host, t.Config.Port)
	}
	return Info{Name: "sse", Addr: addr, Path: path}
}

// Serve starts the SSE transport and blocks until the context is cancelled.
func (t *SSETransport) Serve(ctx context.Context, server Server) error {
	host := t.Config.Host
	if host == "" {
		host = "0.0.0.0"
	}
	path := t.Config.Path
	if path == "" {
		path = "/mcp"
	}
	addr := fmt.Sprintf("%s:%d", host, t.Config.Port)

	mux := http.NewServeMux()
	handler := mcp.NewSSEHandler(func(_ *http.Request) *mcp.Server {
		return server.MCPServer()
	}, nil)
	mux.Handle(path, handler)

	httpServer := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: t.Config.ReadHeaderTimeout,
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
		err := httpServer.Serve(ln)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
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

// Close shuts down the SSE HTTP server.
func (t *SSETransport) Close() error {
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
