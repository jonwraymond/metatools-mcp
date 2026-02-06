package transport

import (
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

type healthMockServer struct {
	mcp *mcp.Server
}

func newHealthMockServer() *healthMockServer {
	return &healthMockServer{
		mcp: mcp.NewServer(&mcp.Implementation{Name: "test", Version: "dev"}, nil),
	}
}

func (m *healthMockServer) Run(ctx context.Context, transport mcp.Transport) error {
	return m.mcp.Run(ctx, transport)
}

func (m *healthMockServer) MCPServer() *mcp.Server {
	return m.mcp
}

func TestHealthEndpoint_SSE(t *testing.T) {
	srv := newHealthMockServer()
	tr := &SSETransport{Config: SSEConfig{
		Host:          "127.0.0.1",
		Port:          0,
		Path:          "/mcp",
		HealthEnabled: true,
		HealthPath:    "/healthz",
	}}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() { errCh <- tr.Serve(ctx, srv) }()

	time.Sleep(100 * time.Millisecond)

	addr := tr.Info().Addr
	require.NotEmpty(t, addr)

	resp, err := http.Get("http://" + addr + "/healthz")
	require.NoError(t, err)
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, "OK", string(body))

	cancel()
	<-errCh
}

func TestHealthEndpoint_Streamable(t *testing.T) {
	srv := newHealthMockServer()
	tr := &StreamableHTTPTransport{Config: StreamableHTTPConfig{
		Host:          "127.0.0.1",
		Port:          0,
		Path:          "/mcp",
		HealthEnabled: true,
		HealthPath:    "/healthz",
	}}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() { errCh <- tr.Serve(ctx, srv) }()

	time.Sleep(100 * time.Millisecond)

	addr := tr.Info().Addr
	require.NotEmpty(t, addr)

	resp, err := http.Get("http://" + addr + "/healthz")
	require.NoError(t, err)
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, "OK", string(body))

	cancel()
	<-errCh
}
