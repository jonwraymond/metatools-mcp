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
func (p *SearchToolsProvider) Name() string   { return "search_tools" }
// Enabled reports whether the provider is enabled.
func (p *SearchToolsProvider) Enabled() bool  { return p.enabled }
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
func (p *ListNamespacesProvider) Name() string   { return "list_namespaces" }
// Enabled reports whether the provider is enabled.
func (p *ListNamespacesProvider) Enabled() bool  { return p.enabled }
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
func (p *DescribeToolProvider) Name() string   { return "describe_tool" }
// Enabled reports whether the provider is enabled.
func (p *DescribeToolProvider) Enabled() bool  { return p.enabled }
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
func (p *ListToolExamplesProvider) Name() string   { return "list_tool_examples" }
// Enabled reports whether the provider is enabled.
func (p *ListToolExamplesProvider) Enabled() bool  { return p.enabled }
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
func (p *RunToolProvider) Name() string   { return "run_tool" }
// Enabled reports whether the provider is enabled.
func (p *RunToolProvider) Enabled() bool  { return p.enabled }
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
func (p *RunChainProvider) Name() string   { return "run_chain" }
// Enabled reports whether the provider is enabled.
func (p *RunChainProvider) Enabled() bool  { return p.enabled }
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
func (p *ExecuteCodeProvider) Name() string   { return "execute_code" }
// Enabled reports whether the provider is enabled.
func (p *ExecuteCodeProvider) Enabled() bool  { return p.enabled }
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
