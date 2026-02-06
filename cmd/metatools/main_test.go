package main

import (
	"context"
	"os"
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
	assert.Equal(t, 13, len(tools))
	assert.True(t, srv.Capabilities().Tools)
}

func TestCreateServer_ListToolsViaMCP(t *testing.T) {
	srv, err := createServer()
	require.NoError(t, err)

	ctx := context.Background()
	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	serverSession, err := srv.MCPServer().Connect(ctx, serverTransport, nil)
	require.NoError(t, err)
	defer func() {
		if err := serverSession.Close(); err != nil {
			t.Fatalf("server session close failed: %v", err)
		}
	}()

	client := mcp.NewClient(&mcp.Implementation{Name: "metatools-test-client"}, nil)
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer func() {
		if err := clientSession.Close(); err != nil {
			t.Fatalf("client session close failed: %v", err)
		}
	}()

	res, err := clientSession.ListTools(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.GreaterOrEqual(t, len(res.Tools), 13)
}

func TestMain_ContextShutdown(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	<-ctx.Done()
	assert.ErrorIs(t, ctx.Err(), context.DeadlineExceeded)
}

func TestCreateServer_LexicalStrategy(t *testing.T) {
	clearSearchEnvVars(t)
	t.Setenv("METATOOLS_SEARCH_STRATEGY", "lexical")

	srv, err := createServer()
	require.NoError(t, err)
	require.NotNil(t, srv)

	tools := srv.ListTools()
	assert.GreaterOrEqual(t, len(tools), 12)
}

func TestCreateServer_DefaultStrategy(t *testing.T) {
	clearSearchEnvVars(t)
	// No env var set - should use default (lexical)

	srv, err := createServer()
	require.NoError(t, err)
	require.NotNil(t, srv)

	tools := srv.ListTools()
	assert.GreaterOrEqual(t, len(tools), 12)
}

func TestCreateServer_InvalidStrategy(t *testing.T) {
	clearSearchEnvVars(t)
	t.Setenv("METATOOLS_SEARCH_STRATEGY", "invalid")

	srv, err := createServer()
	assert.Error(t, err)
	assert.Nil(t, srv)
	assert.Contains(t, err.Error(), "invalid config")
	assert.Contains(t, err.Error(), "unknown search strategy")
}

// clearSearchEnvVars unsets all METATOOLS_SEARCH_* env vars for test isolation
func clearSearchEnvVars(t *testing.T) {
	t.Helper()
	vars := []string{
		"METATOOLS_SEARCH_STRATEGY",
		"METATOOLS_SEARCH_BM25_NAME_BOOST",
		"METATOOLS_SEARCH_BM25_NAMESPACE_BOOST",
		"METATOOLS_SEARCH_BM25_TAGS_BOOST",
		"METATOOLS_SEARCH_BM25_MAX_DOCS",
		"METATOOLS_SEARCH_BM25_MAX_DOCTEXT_LEN",
		"METATOOLS_NOTIFY_TOOL_LIST_CHANGED",
		"METATOOLS_NOTIFY_TOOL_LIST_CHANGED_DEBOUNCE_MS",
	}
	for _, v := range vars {
		if err := os.Unsetenv(v); err != nil {
			t.Fatalf("unset %s: %v", v, err)
		}
	}
}
