package server

import (
	"context"
	"testing"

	"github.com/jonwraymond/toolexec/backend"
	"github.com/jonwraymond/toolexec/backend/local"
	"github.com/stretchr/testify/require"
)

func TestBackendAdapter_GetTools(t *testing.T) {
	registry := backend.NewRegistry()

	localBackend := local.New("local")
	localBackend.RegisterHandler("echo", local.ToolDef{
		Name:        "echo",
		Description: "Echo input",
		Handler: func(_ context.Context, args map[string]any) (any, error) {
			return args["message"], nil
		},
	})

	require.NoError(t, registry.Register(localBackend))

	adapter := NewBackendAdapter(registry)

	tools, err := adapter.GetTools(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, tools)
}

func TestBackendAdapter_Execute(t *testing.T) {
	registry := backend.NewRegistry()

	localBackend := local.New("local")
	localBackend.RegisterHandler("echo", local.ToolDef{
		Name:        "echo",
		Description: "Echo input",
		Handler: func(_ context.Context, args map[string]any) (any, error) {
			return args["message"], nil
		},
	})

	require.NoError(t, registry.Register(localBackend))

	adapter := NewBackendAdapter(registry)

	result, err := adapter.Execute(context.Background(), "local:echo", map[string]any{
		"message": "hello world",
	})
	require.NoError(t, err)
	require.Equal(t, "hello world", result)
}
