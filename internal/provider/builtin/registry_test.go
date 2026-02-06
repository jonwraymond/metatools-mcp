package builtin

import (
	"context"
	"testing"

	"github.com/jonwraymond/metatools-mcp/internal/config"
	"github.com/jonwraymond/metatools-mcp/internal/handlers"
	"github.com/jonwraymond/metatools-mcp/internal/skills"
	"github.com/jonwraymond/metatools-mcp/internal/toolset"
	"github.com/jonwraymond/metatools-mcp/pkg/metatools"
	"github.com/jonwraymond/toolfoundation/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockIndex struct{}

func (m *mockIndex) SearchPage(_ context.Context, _ string, _ int, _ string) ([]metatools.ToolSummary, string, error) {
	return nil, "", nil
}
func (m *mockIndex) ListNamespacesPage(_ context.Context, _ int, _ string) ([]string, string, error) {
	return nil, "", nil
}
func (m *mockIndex) GetAllBackends(_ context.Context, _ string) ([]model.ToolBackend, error) {
	return nil, nil
}

type mockStore struct{}

func (m *mockStore) DescribeTool(_ context.Context, _ string, _ string) (handlers.ToolDoc, error) {
	return handlers.ToolDoc{}, nil
}
func (m *mockStore) ListExamples(_ context.Context, _ string, _ int) ([]metatools.ToolExample, error) {
	return nil, nil
}

type mockRunner struct{}

func (m *mockRunner) Run(_ context.Context, _ string, _ map[string]any) (handlers.RunResult, error) {
	return handlers.RunResult{}, nil
}
func (m *mockRunner) RunChain(_ context.Context, _ []handlers.ChainStep) (handlers.RunResult, []handlers.StepResult, error) {
	return handlers.RunResult{}, nil, nil
}

func TestNewRegistry_Defaults(t *testing.T) {
	runner := &mockRunner{}
	toolsets := toolset.NewRegistry(nil)
	skillsRegistry := skills.NewRegistry(nil)
	defaults := config.DefaultAppConfig().SkillDefaults
	deps := Deps{
		Search:     handlers.NewSearchHandler(&mockIndex{}),
		ListTools:  handlers.NewListToolsHandler(&mockIndex{}),
		Namespaces: handlers.NewNamespacesHandler(&mockIndex{}),
		Describe:   handlers.NewDescribeHandler(&mockStore{}),
		Examples:   handlers.NewExamplesHandler(&mockStore{}),
		Run:        handlers.NewRunHandler(runner),
		Chain:      handlers.NewChainHandler(runner),
		Code:       nil,
		Toolsets:   handlers.NewToolsetsHandler(toolsets),
		Skills: handlers.NewSkillsHandler(
			skillsRegistry,
			toolsets,
			runner,
			handlers.SkillDefaults{
				MaxSteps:     defaults.MaxSteps,
				MaxToolCalls: defaults.MaxToolCalls,
				Timeout:      defaults.Timeout,
			},
		),
	}

	cfg := config.DefaultAppConfig().Providers

	registry, err := NewRegistry(deps, RegistryOptions{Providers: cfg})
	require.NoError(t, err)

	names := registry.Names()
	assert.ElementsMatch(t, []string{
		"search_tools",
		"list_tools",
		"list_namespaces",
		"describe_tool",
		"list_tool_examples",
		"run_tool",
		"run_chain",
		"list_toolsets",
		"describe_toolset",
		"list_skills",
		"describe_skill",
		"plan_skill",
		"run_skill",
	}, names)
}

func TestNewRegistry_ExecuteCodeEnabledMissingHandler(t *testing.T) {
	runner := &mockRunner{}
	toolsets := toolset.NewRegistry(nil)
	skillsRegistry := skills.NewRegistry(nil)
	defaults := config.DefaultAppConfig().SkillDefaults
	deps := Deps{
		Search:     handlers.NewSearchHandler(&mockIndex{}),
		ListTools:  handlers.NewListToolsHandler(&mockIndex{}),
		Namespaces: handlers.NewNamespacesHandler(&mockIndex{}),
		Describe:   handlers.NewDescribeHandler(&mockStore{}),
		Examples:   handlers.NewExamplesHandler(&mockStore{}),
		Run:        handlers.NewRunHandler(runner),
		Chain:      handlers.NewChainHandler(runner),
		Code:       nil,
		Toolsets:   handlers.NewToolsetsHandler(toolsets),
		Skills: handlers.NewSkillsHandler(
			skillsRegistry,
			toolsets,
			runner,
			handlers.SkillDefaults{
				MaxSteps:     defaults.MaxSteps,
				MaxToolCalls: defaults.MaxToolCalls,
				Timeout:      defaults.Timeout,
			},
		),
	}

	cfg := config.DefaultAppConfig().Providers
	cfg.ExecuteCode.Enabled = true

	_, err := NewRegistry(deps, RegistryOptions{Providers: cfg})
	require.Error(t, err)
}

func TestNewRegistry_DisabledProviders(t *testing.T) {
	runner := &mockRunner{}
	toolsets := toolset.NewRegistry(nil)
	skillsRegistry := skills.NewRegistry(nil)
	defaults := config.DefaultAppConfig().SkillDefaults
	deps := Deps{
		Search:     handlers.NewSearchHandler(&mockIndex{}),
		ListTools:  handlers.NewListToolsHandler(&mockIndex{}),
		Namespaces: handlers.NewNamespacesHandler(&mockIndex{}),
		Describe:   handlers.NewDescribeHandler(&mockStore{}),
		Examples:   handlers.NewExamplesHandler(&mockStore{}),
		Run:        handlers.NewRunHandler(runner),
		Chain:      handlers.NewChainHandler(runner),
		Code:       nil,
		Toolsets:   handlers.NewToolsetsHandler(toolsets),
		Skills: handlers.NewSkillsHandler(
			skillsRegistry,
			toolsets,
			runner,
			handlers.SkillDefaults{
				MaxSteps:     defaults.MaxSteps,
				MaxToolCalls: defaults.MaxToolCalls,
				Timeout:      defaults.Timeout,
			},
		),
	}

	cfg := config.DefaultAppConfig().Providers
	cfg.SearchTools.Enabled = false
	cfg.ListToolExamples.Enabled = false

	registry, err := NewRegistry(deps, RegistryOptions{Providers: cfg})
	require.NoError(t, err)

	_, ok := registry.Get("search_tools")
	assert.False(t, ok)
	_, ok = registry.Get("list_tool_examples")
	assert.False(t, ok)
}
