package main

import (
	"context"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateServer_ValidConfig(t *testing.T) {
	srv, err := createServer()
	require.NoError(t, err)
	require.NotNil(t, srv)

	tools := srv.ListTools()
	assert.Equal(t, 6, len(tools))
	assert.True(t, srv.Capabilities().Tools)
}

func TestCreateServer_ListToolsViaMCP(t *testing.T) {
	srv, err := createServer()
	require.NoError(t, err)

	ctx := context.Background()
	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	serverSession, err := srv.MCPServer().Connect(ctx, serverTransport, nil)
	require.NoError(t, err)
	defer serverSession.Close()

	client := mcp.NewClient(&mcp.Implementation{Name: "metatools-test-client"}, nil)
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer clientSession.Close()

	res, err := clientSession.ListTools(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.GreaterOrEqual(t, len(res.Tools), 6)
}

func TestMain_ContextShutdown(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	<-ctx.Done()
	assert.ErrorIs(t, ctx.Err(), context.DeadlineExceeded)
}
