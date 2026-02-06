package mcpbackend

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/jonwraymond/tooldiscovery/index"
	"github.com/jonwraymond/toolexec/run"
	"github.com/stretchr/testify/require"
)

func TestNewManagerValidation(t *testing.T) {
	_, err := NewManager([]Config{{Name: "", URL: "https://example.com/mcp"}})
	require.Error(t, err)

	_, err = NewManager([]Config{{Name: "test", URL: ""}})
	require.Error(t, err)

	_, err = NewManager([]Config{
		{Name: "dup", URL: "https://example.com/mcp"},
		{Name: "dup", URL: "https://example.com/mcp"},
	})
	require.Error(t, err)
}

func TestManagerConnectAndCall(t *testing.T) {
	ctx := context.Background()

	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0.0.0"}, nil)
	tool := &mcp.Tool{
		Name:        "ping",
		Description: "ping",
		InputSchema: map[string]any{"type": "object"},
	}
	mcp.AddTool[map[string]any, any](server, tool, func(_ context.Context, _ *mcp.CallToolRequest, _ map[string]any) (*mcp.CallToolResult, any, error) {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "pong"}},
		}, "pong", nil
	})

	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	serverSession, err := server.Connect(ctx, serverTransport, nil)
	require.NoError(t, err)
	defer func() { _ = serverSession.Close() }()

	manager, err := NewManager([]Config{{Name: "backend", Transport: clientTransport}})
	require.NoError(t, err)

	require.NoError(t, manager.ConnectAll(ctx))

	tools := manager.ToolsSnapshot()
	require.Len(t, tools["backend"], 1)
	require.Equal(t, "ping", tools["backend"][0].Name)

	res, err := manager.CallTool(ctx, "backend", &mcp.CallToolParams{Name: "ping"})
	require.NoError(t, err)
	require.Len(t, res.Content, 1)
	require.Equal(t, "pong", res.Content[0].(*mcp.TextContent).Text)

	_, err = manager.CallToolStream(ctx, "backend", &mcp.CallToolParams{Name: "ping"})
	require.True(t, errors.Is(err, run.ErrStreamNotSupported))
}

func TestManagerRefreshAllContinuesOnError(t *testing.T) {
	ctx := context.Background()

	goodServer := mcp.NewServer(&mcp.Implementation{Name: "good", Version: "0.0.0"}, nil)
	goodTool := &mcp.Tool{
		Name:        "ping",
		Description: "ping",
		InputSchema: map[string]any{"type": "object"},
	}
	mcp.AddTool[map[string]any, any](goodServer, goodTool, func(_ context.Context, _ *mcp.CallToolRequest, _ map[string]any) (*mcp.CallToolResult, any, error) {
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "pong"}}}, "pong", nil
	})

	badServer := mcp.NewServer(&mcp.Implementation{Name: "bad", Version: "0.0.0"}, nil)
	badTool := &mcp.Tool{
		Name:        "noop",
		Description: "noop",
		InputSchema: map[string]any{"type": "object"},
	}
	mcp.AddTool[map[string]any, any](badServer, badTool, func(_ context.Context, _ *mcp.CallToolRequest, _ map[string]any) (*mcp.CallToolResult, any, error) {
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "ok"}}}, "ok", nil
	})

	goodServerTransport, goodClientTransport := mcp.NewInMemoryTransports()
	goodServerSession, err := goodServer.Connect(ctx, goodServerTransport, nil)
	require.NoError(t, err)
	defer func() { _ = goodServerSession.Close() }()

	badServerTransport, badClientTransport := mcp.NewInMemoryTransports()
	badServerSession, err := badServer.Connect(ctx, badServerTransport, nil)
	require.NoError(t, err)
	defer func() { _ = badServerSession.Close() }()

	manager, err := NewManager([]Config{
		{Name: "good", Transport: goodClientTransport},
		{Name: "bad", Transport: badClientTransport},
	})
	require.NoError(t, err)
	require.NoError(t, manager.ConnectAll(ctx))

	idx := index.NewInMemoryIndex()
	require.NoError(t, manager.RegisterTools(idx))

	// Add a new tool to the good backend after initial registration.
	newTool := &mcp.Tool{
		Name:        "newtool",
		Description: "newtool",
		InputSchema: map[string]any{"type": "object"},
	}
	mcp.AddTool[map[string]any, any](goodServer, newTool, func(_ context.Context, _ *mcp.CallToolRequest, _ map[string]any) (*mcp.CallToolResult, any, error) {
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "new"}}}, "new", nil
	})

	// Simulate a broken backend session so refresh fails for "bad".
	b, err := manager.lookupBackend("bad")
	require.NoError(t, err)
	b.mu.Lock()
	b.session = nil
	b.mu.Unlock()

	err = manager.RefreshAll(ctx, idx)
	require.Error(t, err)

	// Ensure the good backend still refreshed and its new tool is registered.
	_, _, err = idx.GetTool("mcp.good:newtool")
	require.NoError(t, err)
}

func TestRefresherMaybeRefreshStale(t *testing.T) {
	ctx := context.Background()

	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0.0.0"}, nil)
	tool := &mcp.Tool{
		Name:        "ping",
		Description: "ping",
		InputSchema: map[string]any{"type": "object"},
	}
	mcp.AddTool[map[string]any, any](server, tool, func(_ context.Context, _ *mcp.CallToolRequest, _ map[string]any) (*mcp.CallToolResult, any, error) {
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "pong"}}}, "pong", nil
	})

	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	serverSession, err := server.Connect(ctx, serverTransport, nil)
	require.NoError(t, err)
	defer func() { _ = serverSession.Close() }()

	manager, err := NewManager([]Config{{Name: "backend", Transport: clientTransport}})
	require.NoError(t, err)
	require.NoError(t, manager.ConnectAll(ctx))

	idx := index.NewInMemoryIndex()
	require.NoError(t, manager.RegisterTools(idx))

	// Add a new tool after initial registration.
	newTool := &mcp.Tool{
		Name:        "newtool",
		Description: "newtool",
		InputSchema: map[string]any{"type": "object"},
	}
	mcp.AddTool[map[string]any, any](server, newTool, func(_ context.Context, _ *mcp.CallToolRequest, _ map[string]any) (*mcp.CallToolResult, any, error) {
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "new"}}}, "new", nil
	})

	// Make the backend stale and verify on-demand refresh registers the new tool.
	b, err := manager.lookupBackend("backend")
	require.NoError(t, err)
	b.mu.Lock()
	b.lastRefresh = time.Now().Add(-2 * time.Second)
	b.mu.Unlock()

	refresher := NewRefresher(manager, idx, RefreshPolicy{
		OnDemand:   true,
		StaleAfter: 1 * time.Second,
	})
	require.NotNil(t, refresher)
	require.NoError(t, refresher.MaybeRefresh(ctx))

	_, _, err = idx.GetTool("mcp.backend:newtool")
	require.NoError(t, err)
}
