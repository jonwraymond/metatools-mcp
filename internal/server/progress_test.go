package server

import (
	"context"
	"testing"
	"time"

	"github.com/jonwraymond/metatools-mcp/internal/config"
	"github.com/jonwraymond/metatools-mcp/internal/handlers"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type progressRunner struct{}

func (p *progressRunner) Run(_ context.Context, _ string, _ map[string]any) (handlers.RunResult, error) {
	return handlers.RunResult{Structured: "ok"}, nil
}

func (p *progressRunner) RunChain(_ context.Context, _ []handlers.ChainStep) (handlers.RunResult, []handlers.StepResult, error) {
	return handlers.RunResult{Structured: "ok"}, nil, nil
}

func (p *progressRunner) RunWithProgress(_ context.Context, _ string, _ map[string]any, onProgress func(handlers.ProgressEvent)) (handlers.RunResult, error) {
	onProgress(handlers.ProgressEvent{Progress: 0, Total: 1, Message: "start"})
	onProgress(handlers.ProgressEvent{Progress: 1, Total: 1, Message: "done"})
	return handlers.RunResult{Structured: "ok"}, nil
}

func (p *progressRunner) RunChainWithProgress(_ context.Context, _ []handlers.ChainStep, onProgress func(handlers.ProgressEvent)) (handlers.RunResult, []handlers.StepResult, error) {
	onProgress(handlers.ProgressEvent{Progress: 0, Total: 2, Message: "start"})
	onProgress(handlers.ProgressEvent{Progress: 1, Total: 2, Message: "step1"})
	onProgress(handlers.ProgressEvent{Progress: 2, Total: 2, Message: "done"})
	return handlers.RunResult{Structured: "ok"}, nil, nil
}

func TestServer_ProgressNotifications_RunTool(t *testing.T) {
	cfg := config.Config{
		Index:     &mockIndex{},
		Docs:      &mockStore{},
		Runner:    &progressRunner{},
		Providers: config.DefaultAppConfig().Providers,
	}

	srv, err := New(cfg)
	require.NoError(t, err)

	ctx := context.Background()
	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	serverSession, err := srv.MCPServer().Connect(ctx, serverTransport, nil)
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, serverSession.Close())
	}()

	progressCh := make(chan struct{}, 4)
	client := mcp.NewClient(&mcp.Implementation{Name: "metatools-progress-client"}, &mcp.ClientOptions{
		ProgressNotificationHandler: func(_ context.Context, _ *mcp.ProgressNotificationClientRequest) {
			progressCh <- struct{}{}
		},
	})
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, clientSession.Close())
	}()

	params := &mcp.CallToolParams{
		Name:      "run_tool",
		Arguments: map[string]any{"tool_id": "test.tool"},
	}
	params.Meta = mcp.Meta{}
	params.SetProgressToken("token-1")

	_, err = clientSession.CallTool(ctx, params)
	require.NoError(t, err)

	// Expect at least 2 progress notifications (start/end).
	timeout := time.After(500 * time.Millisecond)
	received := 0
	for received < 2 {
		select {
		case <-progressCh:
			received++
		case <-timeout:
			t.Fatalf("expected progress notifications, got %d", received)
		}
	}
}
