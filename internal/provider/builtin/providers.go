package builtin

import (
	"context"

	"github.com/jonwraymond/metatools-mcp/internal/handlers"
	"github.com/jonwraymond/metatools-mcp/pkg/metatools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// SearchToolsProvider serves the search_tools built-in tool.
type SearchToolsProvider struct {
	handler *handlers.SearchHandler
	enabled bool
}

// NewSearchToolsProvider builds a SearchToolsProvider.
func NewSearchToolsProvider(handler *handlers.SearchHandler, enabled bool) *SearchToolsProvider {
	return &SearchToolsProvider{handler: handler, enabled: enabled}
}

// Name returns the MCP tool name.
func (p *SearchToolsProvider) Name() string { return "search_tools" }

// Enabled reports whether the provider is enabled.
func (p *SearchToolsProvider) Enabled() bool { return p.enabled }

// Tool returns the MCP tool schema.
func (p *SearchToolsProvider) Tool() mcp.Tool { return searchToolsTool() }

// Handle executes the search_tools request.
func (p *SearchToolsProvider) Handle(ctx context.Context, _ *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	var input metatools.SearchToolsInput
	if err := decodeArgs(args, &input); err != nil {
		return nil, nil, err
	}
	out, err := p.handler.Handle(ctx, input)
	if err != nil {
		return nil, nil, err
	}
	if out == nil {
		out = &metatools.SearchToolsOutput{}
	}
	return nil, *out, nil
}

// ListToolsProvider serves the list_tools built-in tool.
type ListToolsProvider struct {
	handler *handlers.ListToolsHandler
	enabled bool
}

// NewListToolsProvider builds a ListToolsProvider.
func NewListToolsProvider(handler *handlers.ListToolsHandler, enabled bool) *ListToolsProvider {
	return &ListToolsProvider{handler: handler, enabled: enabled}
}

// Name returns the MCP tool name.
func (p *ListToolsProvider) Name() string { return "list_tools" }

// Enabled reports whether the provider is enabled.
func (p *ListToolsProvider) Enabled() bool { return p.enabled }

// Tool returns the MCP tool schema.
func (p *ListToolsProvider) Tool() mcp.Tool { return listToolsTool() }

// Handle executes the list_tools request.
func (p *ListToolsProvider) Handle(ctx context.Context, _ *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	var input metatools.ListToolsInput
	if err := decodeArgs(args, &input); err != nil {
		return nil, nil, err
	}
	out, err := p.handler.Handle(ctx, input)
	if err != nil {
		return nil, nil, err
	}
	if out == nil {
		out = &metatools.ListToolsOutput{}
	}
	return nil, *out, nil
}

// ListNamespacesProvider serves the list_namespaces built-in tool.
type ListNamespacesProvider struct {
	handler *handlers.NamespacesHandler
	enabled bool
}

// NewListNamespacesProvider builds a ListNamespacesProvider.
func NewListNamespacesProvider(handler *handlers.NamespacesHandler, enabled bool) *ListNamespacesProvider {
	return &ListNamespacesProvider{handler: handler, enabled: enabled}
}

// Name returns the MCP tool name.
func (p *ListNamespacesProvider) Name() string { return "list_namespaces" }

// Enabled reports whether the provider is enabled.
func (p *ListNamespacesProvider) Enabled() bool { return p.enabled }

// Tool returns the MCP tool schema.
func (p *ListNamespacesProvider) Tool() mcp.Tool { return listNamespacesTool() }

// Handle executes the list_namespaces request.
func (p *ListNamespacesProvider) Handle(ctx context.Context, _ *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	var input metatools.ListNamespacesInput
	if err := decodeArgs(args, &input); err != nil {
		return nil, nil, err
	}
	out, err := p.handler.Handle(ctx, input)
	if err != nil {
		return nil, nil, err
	}
	if out == nil {
		out = &metatools.ListNamespacesOutput{}
	}
	return nil, *out, nil
}

// DescribeToolProvider serves the describe_tool built-in tool.
type DescribeToolProvider struct {
	handler *handlers.DescribeHandler
	enabled bool
}

// NewDescribeToolProvider builds a DescribeToolProvider.
func NewDescribeToolProvider(handler *handlers.DescribeHandler, enabled bool) *DescribeToolProvider {
	return &DescribeToolProvider{handler: handler, enabled: enabled}
}

// Name returns the MCP tool name.
func (p *DescribeToolProvider) Name() string { return "describe_tool" }

// Enabled reports whether the provider is enabled.
func (p *DescribeToolProvider) Enabled() bool { return p.enabled }

// Tool returns the MCP tool schema.
func (p *DescribeToolProvider) Tool() mcp.Tool { return describeToolTool() }

// Handle executes the describe_tool request.
func (p *DescribeToolProvider) Handle(ctx context.Context, _ *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	var input metatools.DescribeToolInput
	if err := decodeArgs(args, &input); err != nil {
		return nil, nil, err
	}
	out, err := p.handler.Handle(ctx, input)
	if err != nil {
		return nil, nil, err
	}
	if out == nil {
		out = &metatools.DescribeToolOutput{}
	}
	return nil, *out, nil
}

// ListToolExamplesProvider serves the list_tool_examples built-in tool.
type ListToolExamplesProvider struct {
	handler *handlers.ExamplesHandler
	enabled bool
}

// NewListToolExamplesProvider builds a ListToolExamplesProvider.
func NewListToolExamplesProvider(handler *handlers.ExamplesHandler, enabled bool) *ListToolExamplesProvider {
	return &ListToolExamplesProvider{handler: handler, enabled: enabled}
}

// Name returns the MCP tool name.
func (p *ListToolExamplesProvider) Name() string { return "list_tool_examples" }

// Enabled reports whether the provider is enabled.
func (p *ListToolExamplesProvider) Enabled() bool { return p.enabled }

// Tool returns the MCP tool schema.
func (p *ListToolExamplesProvider) Tool() mcp.Tool { return listToolExamplesTool() }

// Handle executes the list_tool_examples request.
func (p *ListToolExamplesProvider) Handle(ctx context.Context, _ *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	var input metatools.ListToolExamplesInput
	if err := decodeArgs(args, &input); err != nil {
		return nil, nil, err
	}
	out, err := p.handler.Handle(ctx, input)
	if err != nil {
		return nil, nil, err
	}
	if out == nil {
		out = &metatools.ListToolExamplesOutput{}
	}
	return nil, *out, nil
}

// RunToolProvider serves the run_tool built-in tool.
type RunToolProvider struct {
	handler *handlers.RunHandler
	enabled bool
}

// NewRunToolProvider builds a RunToolProvider.
func NewRunToolProvider(handler *handlers.RunHandler, enabled bool) *RunToolProvider {
	return &RunToolProvider{handler: handler, enabled: enabled}
}

// Name returns the MCP tool name.
func (p *RunToolProvider) Name() string { return "run_tool" }

// Enabled reports whether the provider is enabled.
func (p *RunToolProvider) Enabled() bool { return p.enabled }

// Tool returns the MCP tool schema.
func (p *RunToolProvider) Tool() mcp.Tool { return runToolTool() }

// Handle executes the run_tool request.
func (p *RunToolProvider) Handle(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	var input metatools.RunToolInput
	if err := decodeArgs(args, &input); err != nil {
		return nil, nil, err
	}
	progress := progressNotifier(ctx, req)
	out, isError, err := p.handler.HandleWithProgress(ctx, input, progress)
	if err != nil {
		return nil, nil, err
	}
	if out == nil {
		out = &metatools.RunToolOutput{}
	}
	return &mcp.CallToolResult{IsError: isError}, *out, nil
}

// RunChainProvider serves the run_chain built-in tool.
type RunChainProvider struct {
	handler *handlers.ChainHandler
	enabled bool
}

// NewRunChainProvider builds a RunChainProvider.
func NewRunChainProvider(handler *handlers.ChainHandler, enabled bool) *RunChainProvider {
	return &RunChainProvider{handler: handler, enabled: enabled}
}

// Name returns the MCP tool name.
func (p *RunChainProvider) Name() string { return "run_chain" }

// Enabled reports whether the provider is enabled.
func (p *RunChainProvider) Enabled() bool { return p.enabled }

// Tool returns the MCP tool schema.
func (p *RunChainProvider) Tool() mcp.Tool { return runChainTool() }

// Handle executes the run_chain request.
func (p *RunChainProvider) Handle(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	var input metatools.RunChainInput
	if err := decodeArgs(args, &input); err != nil {
		return nil, nil, err
	}
	progress := progressNotifier(ctx, req)
	out, isError, err := p.handler.HandleWithProgress(ctx, input, progress)
	if err != nil {
		return nil, nil, err
	}
	if out == nil {
		out = &metatools.RunChainOutput{}
	}
	return &mcp.CallToolResult{IsError: isError}, *out, nil
}

// ExecuteCodeProvider serves the execute_code built-in tool.
type ExecuteCodeProvider struct {
	handler *handlers.CodeHandler
	enabled bool
}

// NewExecuteCodeProvider builds an ExecuteCodeProvider.
func NewExecuteCodeProvider(handler *handlers.CodeHandler, enabled bool) *ExecuteCodeProvider {
	return &ExecuteCodeProvider{handler: handler, enabled: enabled}
}

// Name returns the MCP tool name.
func (p *ExecuteCodeProvider) Name() string { return "execute_code" }

// Enabled reports whether the provider is enabled.
func (p *ExecuteCodeProvider) Enabled() bool { return p.enabled }

// Tool returns the MCP tool schema.
func (p *ExecuteCodeProvider) Tool() mcp.Tool { return executeCodeTool() }

// Handle executes the execute_code request.
func (p *ExecuteCodeProvider) Handle(ctx context.Context, req *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	var input metatools.ExecuteCodeInput
	if err := decodeArgs(args, &input); err != nil {
		return nil, nil, err
	}
	progress := progressNotifier(ctx, req)
	if progress != nil {
		progress(handlers.ProgressEvent{Progress: 0, Total: 1, Message: "started"})
	}
	out, err := p.handler.Handle(ctx, input)
	if progress != nil {
		msg := "completed"
		if err != nil {
			msg = "error"
		}
		progress(handlers.ProgressEvent{Progress: 1, Total: 1, Message: msg})
	}
	if err != nil {
		return nil, nil, err
	}
	if out == nil {
		out = &metatools.ExecuteCodeOutput{}
	}
	return nil, *out, nil
}

// ListToolsetsProvider serves the list_toolsets built-in tool.
type ListToolsetsProvider struct {
	handler *handlers.ToolsetsHandler
	enabled bool
}

// NewListToolsetsProvider builds a ListToolsetsProvider.
func NewListToolsetsProvider(handler *handlers.ToolsetsHandler, enabled bool) *ListToolsetsProvider {
	return &ListToolsetsProvider{handler: handler, enabled: enabled}
}

// Name returns the MCP tool name.
func (p *ListToolsetsProvider) Name() string { return "list_toolsets" }

// Enabled reports whether the provider is enabled.
func (p *ListToolsetsProvider) Enabled() bool { return p.enabled }

// Tool returns the MCP tool schema.
func (p *ListToolsetsProvider) Tool() mcp.Tool { return listToolsetsTool() }

// Handle executes the list_toolsets request.
func (p *ListToolsetsProvider) Handle(ctx context.Context, _ *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	var input metatools.ListToolsetsInput
	if err := decodeArgs(args, &input); err != nil {
		return nil, nil, err
	}
	out, err := p.handler.List(ctx, input)
	if err != nil {
		return nil, nil, err
	}
	if out == nil {
		out = &metatools.ListToolsetsOutput{}
	}
	return nil, *out, nil
}

// DescribeToolsetProvider serves the describe_toolset built-in tool.
type DescribeToolsetProvider struct {
	handler *handlers.ToolsetsHandler
	enabled bool
}

// NewDescribeToolsetProvider builds a DescribeToolsetProvider.
func NewDescribeToolsetProvider(handler *handlers.ToolsetsHandler, enabled bool) *DescribeToolsetProvider {
	return &DescribeToolsetProvider{handler: handler, enabled: enabled}
}

// Name returns the MCP tool name.
func (p *DescribeToolsetProvider) Name() string { return "describe_toolset" }

// Enabled reports whether the provider is enabled.
func (p *DescribeToolsetProvider) Enabled() bool { return p.enabled }

// Tool returns the MCP tool schema.
func (p *DescribeToolsetProvider) Tool() mcp.Tool { return describeToolsetTool() }

// Handle executes the describe_toolset request.
func (p *DescribeToolsetProvider) Handle(ctx context.Context, _ *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	var input metatools.DescribeToolsetInput
	if err := decodeArgs(args, &input); err != nil {
		return nil, nil, err
	}
	out, err := p.handler.Describe(ctx, input)
	if err != nil {
		return nil, nil, err
	}
	if out == nil {
		out = &metatools.DescribeToolsetOutput{}
	}
	return nil, *out, nil
}

// ListSkillsProvider serves the list_skills built-in tool.
type ListSkillsProvider struct {
	handler *handlers.SkillsHandler
	enabled bool
}

// NewListSkillsProvider builds a ListSkillsProvider.
func NewListSkillsProvider(handler *handlers.SkillsHandler, enabled bool) *ListSkillsProvider {
	return &ListSkillsProvider{handler: handler, enabled: enabled}
}

// Name returns the MCP tool name.
func (p *ListSkillsProvider) Name() string { return "list_skills" }

// Enabled reports whether the provider is enabled.
func (p *ListSkillsProvider) Enabled() bool { return p.enabled }

// Tool returns the MCP tool schema.
func (p *ListSkillsProvider) Tool() mcp.Tool { return listSkillsTool() }

// Handle executes the list_skills request.
func (p *ListSkillsProvider) Handle(ctx context.Context, _ *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	var input metatools.ListSkillsInput
	if err := decodeArgs(args, &input); err != nil {
		return nil, nil, err
	}
	out, err := p.handler.List(ctx, input)
	if err != nil {
		return nil, nil, err
	}
	if out == nil {
		out = &metatools.ListSkillsOutput{}
	}
	return nil, *out, nil
}

// DescribeSkillProvider serves the describe_skill built-in tool.
type DescribeSkillProvider struct {
	handler *handlers.SkillsHandler
	enabled bool
}

// NewDescribeSkillProvider builds a DescribeSkillProvider.
func NewDescribeSkillProvider(handler *handlers.SkillsHandler, enabled bool) *DescribeSkillProvider {
	return &DescribeSkillProvider{handler: handler, enabled: enabled}
}

// Name returns the MCP tool name.
func (p *DescribeSkillProvider) Name() string { return "describe_skill" }

// Enabled reports whether the provider is enabled.
func (p *DescribeSkillProvider) Enabled() bool { return p.enabled }

// Tool returns the MCP tool schema.
func (p *DescribeSkillProvider) Tool() mcp.Tool { return describeSkillTool() }

// Handle executes the describe_skill request.
func (p *DescribeSkillProvider) Handle(ctx context.Context, _ *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	var input metatools.DescribeSkillInput
	if err := decodeArgs(args, &input); err != nil {
		return nil, nil, err
	}
	out, err := p.handler.Describe(ctx, input)
	if err != nil {
		return nil, nil, err
	}
	if out == nil {
		out = &metatools.DescribeSkillOutput{}
	}
	return nil, *out, nil
}

// PlanSkillProvider serves the plan_skill built-in tool.
type PlanSkillProvider struct {
	handler *handlers.SkillsHandler
	enabled bool
}

// NewPlanSkillProvider builds a PlanSkillProvider.
func NewPlanSkillProvider(handler *handlers.SkillsHandler, enabled bool) *PlanSkillProvider {
	return &PlanSkillProvider{handler: handler, enabled: enabled}
}

// Name returns the MCP tool name.
func (p *PlanSkillProvider) Name() string { return "plan_skill" }

// Enabled reports whether the provider is enabled.
func (p *PlanSkillProvider) Enabled() bool { return p.enabled }

// Tool returns the MCP tool schema.
func (p *PlanSkillProvider) Tool() mcp.Tool { return planSkillTool() }

// Handle executes the plan_skill request.
func (p *PlanSkillProvider) Handle(ctx context.Context, _ *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	var input metatools.PlanSkillInput
	if err := decodeArgs(args, &input); err != nil {
		return nil, nil, err
	}
	out, err := p.handler.Plan(ctx, input)
	if err != nil {
		return nil, nil, err
	}
	if out == nil {
		out = &metatools.PlanSkillOutput{}
	}
	return nil, *out, nil
}

// RunSkillProvider serves the run_skill built-in tool.
type RunSkillProvider struct {
	handler *handlers.SkillsHandler
	enabled bool
}

// NewRunSkillProvider builds a RunSkillProvider.
func NewRunSkillProvider(handler *handlers.SkillsHandler, enabled bool) *RunSkillProvider {
	return &RunSkillProvider{handler: handler, enabled: enabled}
}

// Name returns the MCP tool name.
func (p *RunSkillProvider) Name() string { return "run_skill" }

// Enabled reports whether the provider is enabled.
func (p *RunSkillProvider) Enabled() bool { return p.enabled }

// Tool returns the MCP tool schema.
func (p *RunSkillProvider) Tool() mcp.Tool { return runSkillTool() }

// Handle executes the run_skill request.
func (p *RunSkillProvider) Handle(ctx context.Context, _ *mcp.CallToolRequest, args map[string]any) (*mcp.CallToolResult, any, error) {
	var input metatools.RunSkillInput
	if err := decodeArgs(args, &input); err != nil {
		return nil, nil, err
	}
	out, isError, err := p.handler.Run(ctx, input)
	if err != nil {
		return nil, nil, err
	}
	if out == nil {
		out = &metatools.RunSkillOutput{}
	}
	return &mcp.CallToolResult{IsError: isError}, *out, nil
}
