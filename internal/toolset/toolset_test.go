package toolset

import (
	"testing"

	"github.com/jonwraymond/tooldiscovery/index"
	"github.com/jonwraymond/toolfoundation/model"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestBuildRegistryFromConfig_FiltersAndPolicy(t *testing.T) {
	idx := index.NewInMemoryIndex(index.IndexOptions{})

	registerTool(t, idx, model.Tool{Tool: mcp.Tool{
		Name:        "alpha",
		Description: "alpha tool",
		InputSchema: map[string]any{"type": "object"},
	}, Namespace: "ns1", Tags: []string{"a"}}, backend("alpha"))

	registerTool(t, idx, model.Tool{Tool: mcp.Tool{
		Name:        "beta",
		Description: "beta tool",
		InputSchema: map[string]any{"type": "object"},
	}, Namespace: "ns1", Tags: []string{"b"}}, backend("beta"))

	registerTool(t, idx, model.Tool{Tool: mcp.Tool{
		Name:        "gamma",
		Description: "gamma tool",
		InputSchema: map[string]any{"type": "object"},
	}, Namespace: "ns2", Tags: []string{"a"}}, backend("gamma"))

	specs := []Spec{
		{
			Name:             "OnlyAlpha",
			NamespaceFilters: []string{"ns1"},
			TagFilters:       []string{"a"},
			AllowIDs:         []string{"ns1:alpha"},
			Policy:           "allow_all",
		},
		{
			Name:   "DenyAll",
			Policy: "deny_all",
		},
	}

	reg, err := BuildRegistry(idx, specs)
	require.NoError(t, err)

	toolsets := reg.List()
	require.Len(t, toolsets, 2)

	first := toolsets[0]
	if first.Name == "OnlyAlpha" {
		require.Equal(t, []string{"ns1:alpha"}, first.ToolIDs())
	} else {
		require.Empty(t, first.ToolIDs())
	}
}

func TestBuildRegistryFromConfig_DeterministicID(t *testing.T) {
	idx := index.NewInMemoryIndex(index.IndexOptions{})
	registerTool(t, idx, model.Tool{Tool: mcp.Tool{
		Name:        "alpha",
		Description: "alpha tool",
		InputSchema: map[string]any{"type": "object"},
	}}, backend("alpha"))

	specs := []Spec{{Name: "My Toolset"}}
	reg, err := BuildRegistry(idx, specs)
	require.NoError(t, err)

	toolsets := reg.List()
	require.Len(t, toolsets, 1)
	require.Equal(t, "toolset:my-toolset", toolsets[0].ID)
}

func registerTool(t *testing.T, idx index.Index, tool model.Tool, backend model.ToolBackend) {
	t.Helper()
	require.NoError(t, idx.RegisterTool(tool, backend))
}

func backend(name string) model.ToolBackend {
	return model.ToolBackend{
		Kind: model.BackendKindLocal,
		Local: &model.LocalBackend{
			Name: name,
		},
	}
}
