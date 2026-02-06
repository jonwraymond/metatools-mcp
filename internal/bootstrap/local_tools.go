package bootstrap

import (
	"context"
	"fmt"

	"github.com/jonwraymond/tooldiscovery/index"
	"github.com/jonwraymond/toolexec/run"
	"github.com/jonwraymond/toolfoundation/model"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// mapLocalRegistry is a minimal in-memory local handler registry.
// It intentionally lives in metatools-mcp so we can provide a deterministic,
// dependency-free set of built-in local tools for examples and smoke tests.
type mapLocalRegistry struct {
	handlers map[string]run.LocalHandler
}

func (r *mapLocalRegistry) Get(name string) (run.LocalHandler, bool) {
	h, ok := r.handlers[name]
	return h, ok
}

// RegisterDefaultLocalTools registers the built-in local tools into the index and returns a
// local registry that can execute them via toolexec/run.
func RegisterDefaultLocalTools(idx index.Index) (run.LocalRegistry, error) {
	reg := &mapLocalRegistry{handlers: map[string]run.LocalHandler{}}

	pingTool := model.Tool{
		Tool: mcp.Tool{
			Name:        "ping",
			Title:       "Ping",
			Description: "Health-check style ping. Returns a deterministic pong response.",
			InputSchema: map[string]any{
				"type":                 "object",
				"additionalProperties": false,
			},
			OutputSchema: map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"properties": map[string]any{
					"message": map[string]any{"type": "string"},
				},
				"required": []any{"message"},
			},
			Annotations: &mcp.ToolAnnotations{
				ReadOnlyHint: true,
			},
		},
		Namespace: "local",
		Tags:      []string{"safe"},
	}

	if err := idx.RegisterTool(pingTool, model.ToolBackend{
		Kind:  model.BackendKindLocal,
		Local: &model.LocalBackend{Name: "ping"},
	}); err != nil {
		return nil, fmt.Errorf("register local tool %q: %w", pingTool.ToolID(), err)
	}

	reg.handlers["ping"] = func(_ context.Context, _ map[string]any) (any, error) {
		return map[string]any{"message": "pong"}, nil
	}

	return reg, nil
}
