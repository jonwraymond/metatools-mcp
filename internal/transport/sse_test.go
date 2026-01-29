package transport

import (
	"context"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type mockServer struct {
	mcp *mcp.Server
}

func newMockServer() *mockServer {
	return &mockServer{
		mcp: mcp.NewServer(&mcp.Implementation{Name: "test", Version: "dev"}, nil),
	}
}

func (m *mockServer) Run(ctx context.Context, transport mcp.Transport) error {
	return m.mcp.Run(ctx, transport)
}

func (m *mockServer) MCPServer() *mcp.Server {
	return m.mcp
}

func TestSSETransport_ServeAndShutdown(t *testing.T) {
	srv := newMockServer()
	tr := &SSETransport{Config: SSEConfig{Host: "127.0.0.1", Port: 0, Path: "/mcp"}}

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
