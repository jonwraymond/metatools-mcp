package server

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jonwraymond/metatools-mcp/internal/adapters"
	"github.com/jonwraymond/metatools-mcp/internal/config"
	"github.com/jonwraymond/tooldiscovery/index"
	"github.com/jonwraymond/toolfoundation/model"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestServerWithIndex(t *testing.T, idx index.Index, notify bool, debounceMs int) *Server {
	t.Helper()

	cfg := config.Config{
		Index:                           adapters.NewIndexAdapter(idx),
		Docs:                            &mockStore{},
		Runner:                          &mockRunner{},
		Providers:                       config.DefaultAppConfig().Providers,
		NotifyToolListChanged:           notify,
		NotifyToolListChangedDebounceMs: debounceMs,
	}

	srv, err := New(cfg)
	require.NoError(t, err)
	return srv
}

func drainNotifications(ch <-chan struct{}, duration time.Duration) {
	deadline := time.After(duration)
	for {
		select {
		case <-ch:
			continue
		case <-deadline:
			return
		}
	}
}

func registerIndexTool(t *testing.T, idx index.Index, name string) {
	t.Helper()

	tool := model.Tool{
		Namespace: "test",
		Tool: mcp.Tool{
			Name: name,
			InputSchema: map[string]any{
				"type": "object",
			},
		},
	}
	backend := model.ToolBackend{
		Kind:  model.BackendKindLocal,
		Local: &model.LocalBackend{Name: name},
	}

	require.NoError(t, idx.RegisterTool(tool, backend))
}

func TestServer_ToolListChangedNotification(t *testing.T) {
	idx := index.NewInMemoryIndex()
	srv := newTestServerWithIndex(t, idx, true, 30)

	ctx := context.Background()
	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	serverSession, err := srv.MCPServer().Connect(ctx, serverTransport, nil)
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, serverSession.Close())
	}()

	notifyCh := make(chan struct{}, 1)
	client := mcp.NewClient(&mcp.Implementation{Name: "metatools-test-client"}, &mcp.ClientOptions{
		ToolListChangedHandler: func(_ context.Context, _ *mcp.ToolListChangedRequest) {
			notifyCh <- struct{}{}
		},
	})
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, clientSession.Close())
	}()

	drainNotifications(notifyCh, 50*time.Millisecond)

	registerIndexTool(t, idx, "alpha")

	select {
	case <-notifyCh:
	case <-time.After(250 * time.Millisecond):
		t.Fatal("expected tool list changed notification")
	}
}

func TestServer_ToolListChangedNotification_Debounce(t *testing.T) {
	idx := index.NewInMemoryIndex()
	srv := newTestServerWithIndex(t, idx, true, 50)

	ctx := context.Background()
	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	serverSession, err := srv.MCPServer().Connect(ctx, serverTransport, nil)
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, serverSession.Close())
	}()

	var notifyCount atomic.Int32
	client := mcp.NewClient(&mcp.Implementation{Name: "metatools-test-client"}, &mcp.ClientOptions{
		ToolListChangedHandler: func(_ context.Context, _ *mcp.ToolListChangedRequest) {
			notifyCount.Add(1)
		},
	})
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, clientSession.Close())
	}()

	time.Sleep(75 * time.Millisecond)
	notifyCount.Store(0)

	registerIndexTool(t, idx, "alpha")
	registerIndexTool(t, idx, "bravo")
	registerIndexTool(t, idx, "charlie")

	time.Sleep(300 * time.Millisecond)
	assert.Equal(t, int32(1), notifyCount.Load())
}

func TestServer_ToolListChangedNotification_SinglePerChange(t *testing.T) {
	idx := index.NewInMemoryIndex()
	srv := newTestServerWithIndex(t, idx, true, 20)

	ctx := context.Background()
	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	serverSession, err := srv.MCPServer().Connect(ctx, serverTransport, nil)
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, serverSession.Close())
	}()

	var notifyCount atomic.Int32
	client := mcp.NewClient(&mcp.Implementation{Name: "metatools-test-client"}, &mcp.ClientOptions{
		ToolListChangedHandler: func(_ context.Context, _ *mcp.ToolListChangedRequest) {
			notifyCount.Add(1)
		},
	})
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, clientSession.Close())
	}()

	time.Sleep(50 * time.Millisecond)
	notifyCount.Store(0)

	registerIndexTool(t, idx, "alpha")

	time.Sleep(200 * time.Millisecond)
	assert.Equal(t, int32(1), notifyCount.Load())
}

func TestServer_ToolListChangedNotification_Disabled(t *testing.T) {
	idx := index.NewInMemoryIndex()
	srv := newTestServerWithIndex(t, idx, false, 30)

	ctx := context.Background()
	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	serverSession, err := srv.MCPServer().Connect(ctx, serverTransport, nil)
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, serverSession.Close())
	}()

	notifyCh := make(chan struct{}, 1)
	client := mcp.NewClient(&mcp.Implementation{Name: "metatools-test-client"}, &mcp.ClientOptions{
		ToolListChangedHandler: func(_ context.Context, _ *mcp.ToolListChangedRequest) {
			notifyCh <- struct{}{}
		},
	})
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, clientSession.Close())
	}()

	drainNotifications(notifyCh, 50*time.Millisecond)

	registerIndexTool(t, idx, "alpha")

	select {
	case <-notifyCh:
		t.Fatal("unexpected tool list changed notification")
	case <-time.After(200 * time.Millisecond):
	}
}
