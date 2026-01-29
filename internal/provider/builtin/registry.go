package builtin

import (
	"fmt"

	"github.com/jonwraymond/metatools-mcp/internal/config"
	"github.com/jonwraymond/metatools-mcp/internal/handlers"
	"github.com/jonwraymond/metatools-mcp/internal/provider"
)

// Deps bundles the handler dependencies for built-in providers.
type Deps struct {
	Search     *handlers.SearchHandler
	Namespaces *handlers.NamespacesHandler
	Describe   *handlers.DescribeHandler
	Examples   *handlers.ExamplesHandler
	Run        *handlers.RunHandler
	Chain      *handlers.ChainHandler
	Code       *handlers.CodeHandler
}

// RegistryOptions configures built-in provider registration.
type RegistryOptions struct {
	Providers config.ProvidersConfig
}

// NewRegistry creates a provider registry populated with built-in providers.
func NewRegistry(deps Deps, opts RegistryOptions) (*provider.Registry, error) {
	registry := provider.NewRegistry()

	if opts.Providers.SearchTools.Enabled {
		if deps.Search == nil {
			return nil, fmt.Errorf("search_tools provider enabled but handler is nil")
		}
		if err := registry.Register(NewSearchToolsProvider(deps.Search, true)); err != nil {
			return nil, err
		}
	}

	if opts.Providers.ListNamespaces.Enabled {
		if deps.Namespaces == nil {
			return nil, fmt.Errorf("list_namespaces provider enabled but handler is nil")
		}
		if err := registry.Register(NewListNamespacesProvider(deps.Namespaces, true)); err != nil {
			return nil, err
		}
	}

	if opts.Providers.DescribeTool.Enabled {
		if deps.Describe == nil {
			return nil, fmt.Errorf("describe_tool provider enabled but handler is nil")
		}
		if err := registry.Register(NewDescribeToolProvider(deps.Describe, true)); err != nil {
			return nil, err
		}
	}

	if opts.Providers.ListToolExamples.Enabled {
		if deps.Examples == nil {
			return nil, fmt.Errorf("list_tool_examples provider enabled but handler is nil")
		}
		if err := registry.Register(NewListToolExamplesProvider(deps.Examples, true)); err != nil {
			return nil, err
		}
	}

	if opts.Providers.RunTool.Enabled {
		if deps.Run == nil {
			return nil, fmt.Errorf("run_tool provider enabled but handler is nil")
		}
		if err := registry.Register(NewRunToolProvider(deps.Run, true)); err != nil {
			return nil, err
		}
	}

	if opts.Providers.RunChain.Enabled {
		if deps.Chain == nil {
			return nil, fmt.Errorf("run_chain provider enabled but handler is nil")
		}
		if err := registry.Register(NewRunChainProvider(deps.Chain, true)); err != nil {
			return nil, err
		}
	}

	if opts.Providers.ExecuteCode.Enabled {
		if deps.Code == nil {
			return nil, fmt.Errorf("execute_code provider enabled but handler is nil")
		}
		if err := registry.Register(NewExecuteCodeProvider(deps.Code, true)); err != nil {
			return nil, err
		}
	}

	return registry, nil
}

// BuiltinProviders lists the built-in provider names.
func BuiltinProviders() []string {
	return []string{
		"search_tools",
		"list_namespaces",
		"describe_tool",
		"list_tool_examples",
		"run_tool",
		"run_chain",
		"execute_code",
	}
}
