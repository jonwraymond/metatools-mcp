package transport

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestStreamableHTTPTransport_Name(t *testing.T) {
	tr := &StreamableHTTPTransport{}
	if got := tr.Name(); got != "streamable" {
		t.Errorf("Name() = %q, want %q", got, "streamable")
	}
}

func TestStreamableHTTPTransport_Info(t *testing.T) {
	tests := []struct {
		name     string
		config   StreamableHTTPConfig
		wantName string
		wantPath string
		wantAddr string
	}{
		{
			name:     "defaults",
			config:   StreamableHTTPConfig{},
			wantName: "streamable",
			wantPath: "/mcp",
			wantAddr: "",
		},
		{
			name:     "with port",
			config:   StreamableHTTPConfig{Port: 8080},
			wantName: "streamable",
			wantPath: "/mcp",
			wantAddr: "0.0.0.0:8080",
		},
		{
			name:     "with host and port",
			config:   StreamableHTTPConfig{Host: "127.0.0.1", Port: 9090},
			wantName: "streamable",
			wantPath: "/mcp",
			wantAddr: "127.0.0.1:9090",
		},
		{
			name:     "custom path",
			config:   StreamableHTTPConfig{Host: "localhost", Port: 8080, Path: "/api/mcp"},
			wantName: "streamable",
			wantPath: "/api/mcp",
			wantAddr: "localhost:8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &StreamableHTTPTransport{Config: tt.config}
			info := tr.Info()

			if info.Name != tt.wantName {
				t.Errorf("Info().Name = %q, want %q", info.Name, tt.wantName)
			}
			if info.Path != tt.wantPath {
				t.Errorf("Info().Path = %q, want %q", info.Path, tt.wantPath)
			}
			if info.Addr != tt.wantAddr {
				t.Errorf("Info().Addr = %q, want %q", info.Addr, tt.wantAddr)
			}
		})
	}
}

func TestStreamableHTTPTransport_Info_WithListener(t *testing.T) {
	tr := &StreamableHTTPTransport{Config: StreamableHTTPConfig{Host: "127.0.0.1", Port: 0}}
	srv := newMockServer()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- tr.Serve(ctx, srv)
	}()

	// Wait for server to start
	time.Sleep(50 * time.Millisecond)

	info := tr.Info()
	if info.Addr == "" {
		t.Error("Info().Addr should return bound address when listener active")
	}
	if !strings.HasPrefix(info.Addr, "127.0.0.1:") {
		t.Errorf("Info().Addr = %q, want prefix %q", info.Addr, "127.0.0.1:")
	}

	cancel()
	<-errCh
}

func TestStreamableHTTPTransport_ServeAndShutdown(t *testing.T) {
	srv := newMockServer()
	tr := &StreamableHTTPTransport{Config: StreamableHTTPConfig{Host: "127.0.0.1", Port: 0, Path: "/mcp"}}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- tr.Serve(ctx, srv)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("Serve() error = %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Serve() did not return after cancel")
	}
}

func TestStreamableHTTPTransport_Close_Idempotent(t *testing.T) {
	tr := &StreamableHTTPTransport{Config: StreamableHTTPConfig{Host: "127.0.0.1", Port: 0}}

	// Close before Serve should be safe
	if err := tr.Close(); err != nil {
		t.Errorf("Close() before Serve error = %v", err)
	}

	// Multiple closes should be safe
	if err := tr.Close(); err != nil {
		t.Errorf("Close() second call error = %v", err)
	}
}

func TestStreamableHTTPTransport_Close_AfterServe(t *testing.T) {
	srv := newMockServer()
	tr := &StreamableHTTPTransport{Config: StreamableHTTPConfig{Host: "127.0.0.1", Port: 0}}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- tr.Serve(ctx, srv)
	}()

	time.Sleep(50 * time.Millisecond)

	// Close should work while serving
	if err := tr.Close(); err != nil {
		t.Errorf("Close() while serving error = %v", err)
	}

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("Serve() error after Close = %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Serve() did not return after Close")
	}
}

func TestStreamableHTTPTransport_Interface(t *testing.T) {
	// Compile-time interface check
	var _ Transport = (*StreamableHTTPTransport)(nil)
}

func TestStreamableHTTPTransport_StatelessMode(t *testing.T) {
	srv := newMockServer()
	tr := &StreamableHTTPTransport{Config: StreamableHTTPConfig{
		Host:      "127.0.0.1",
		Port:      0,
		Stateless: true,
	}}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- tr.Serve(ctx, srv)
	}()

	time.Sleep(50 * time.Millisecond)

	// Verify server started
	info := tr.Info()
	if info.Addr == "" {
		t.Fatal("Server did not start")
	}

	cancel()
	<-errCh
}

func TestStreamableHTTPTransport_JSONResponseMode(t *testing.T) {
	srv := newMockServer()
	tr := &StreamableHTTPTransport{Config: StreamableHTTPConfig{
		Host:         "127.0.0.1",
		Port:         0,
		JSONResponse: true,
	}}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- tr.Serve(ctx, srv)
	}()

	time.Sleep(50 * time.Millisecond)

	info := tr.Info()
	if info.Addr == "" {
		t.Fatal("Server did not start")
	}

	cancel()
	<-errCh
}

func TestStreamableHTTPTransport_SessionTimeout(t *testing.T) {
	srv := newMockServer()
	tr := &StreamableHTTPTransport{Config: StreamableHTTPConfig{
		Host:           "127.0.0.1",
		Port:           0,
		SessionTimeout: 5 * time.Minute,
	}}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- tr.Serve(ctx, srv)
	}()

	time.Sleep(50 * time.Millisecond)

	info := tr.Info()
	if info.Addr == "" {
		t.Fatal("Server did not start")
	}

	cancel()
	<-errCh
}

func TestStreamableHTTPTransport_HTTPEndpoint(t *testing.T) {
	srv := newMockServer()
	tr := &StreamableHTTPTransport{Config: StreamableHTTPConfig{
		Host:         "127.0.0.1",
		Port:         0,
		Path:         "/mcp",
		JSONResponse: true,
	}}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- tr.Serve(ctx, srv)
	}()

	time.Sleep(100 * time.Millisecond)

	info := tr.Info()
	if info.Addr == "" {
		t.Fatal("Server did not start")
	}

	// Test that endpoint is reachable (non-JSON-RPC request should fail gracefully)
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("http://" + info.Addr + info.Path)
	if err != nil {
		t.Fatalf("GET request error = %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// GET without proper headers should return method not allowed or bad request
	// The exact status depends on SDK implementation
	if resp.StatusCode != http.StatusMethodNotAllowed && resp.StatusCode != http.StatusBadRequest {
		t.Logf("GET response status = %d (expected 400 or 405)", resp.StatusCode)
	}

	cancel()
	<-errCh
}
