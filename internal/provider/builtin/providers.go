package builtin

import (
	"context"

	"github.com/jonwraymond/metatools-mcp/internal/handlers"
	"github.com/jonwraymond/metatools-mcp/pkg/metatools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type SearchToolsProvider struct {
	handler *handlers.SearchHandler
	enabled bool
}

func NewSearchToolsProvider(handler *handlers.SearchHandler, enabled bool) *SearchToolsProvider {
	return &SearchToolsProvider{handler: handler, enabled: enabled}
}

func (p *SearchToolsProvider) Name() string   { return "search_tools" }
func (p *SearchToolsProvider) Enabled() bool  { return p.enabled }
func (p *SearchToolsProvider) Tool() mcp.Tool { return searchToolsTool() }

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

type ListNamespacesProvider struct {
	handler *handlers.NamespacesHandler
	enabled bool
}

func NewListNamespacesProvider(handler *handlers.NamespacesHandler, enabled bool) *ListNamespacesProvider {
	return &ListNamespacesProvider{handler: handler, enabled: enabled}
}

func (p *ListNamespacesProvider) Name() string   { return "list_namespaces" }
func (p *ListNamespacesProvider) Enabled() bool  { return p.enabled }
func (p *ListNamespacesProvider) Tool() mcp.Tool { return listNamespacesTool() }

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

type DescribeToolProvider struct {
	handler *handlers.DescribeHandler
	enabled bool
}

func NewDescribeToolProvider(handler *handlers.DescribeHandler, enabled bool) *DescribeToolProvider {
	return &DescribeToolProvider{handler: handler, enabled: enabled}
}

func (p *DescribeToolProvider) Name() string   { return "describe_tool" }
func (p *DescribeToolProvider) Enabled() bool  { return p.enabled }
func (p *DescribeToolProvider) Tool() mcp.Tool { return describeToolTool() }

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

type ListToolExamplesProvider struct {
	handler *handlers.ExamplesHandler
	enabled bool
}

func NewListToolExamplesProvider(handler *handlers.ExamplesHandler, enabled bool) *ListToolExamplesProvider {
	return &ListToolExamplesProvider{handler: handler, enabled: enabled}
}

func (p *ListToolExamplesProvider) Name() string   { return "list_tool_examples" }
func (p *ListToolExamplesProvider) Enabled() bool  { return p.enabled }
func (p *ListToolExamplesProvider) Tool() mcp.Tool { return listToolExamplesTool() }

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

type RunToolProvider struct {
	handler *handlers.RunHandler
	enabled bool
}

func NewRunToolProvider(handler *handlers.RunHandler, enabled bool) *RunToolProvider {
	return &RunToolProvider{handler: handler, enabled: enabled}
}

func (p *RunToolProvider) Name() string   { return "run_tool" }
func (p *RunToolProvider) Enabled() bool  { return p.enabled }
func (p *RunToolProvider) Tool() mcp.Tool { return runToolTool() }

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

type RunChainProvider struct {
	handler *handlers.ChainHandler
	enabled bool
}

func NewRunChainProvider(handler *handlers.ChainHandler, enabled bool) *RunChainProvider {
	return &RunChainProvider{handler: handler, enabled: enabled}
}

func (p *RunChainProvider) Name() string   { return "run_chain" }
func (p *RunChainProvider) Enabled() bool  { return p.enabled }
func (p *RunChainProvider) Tool() mcp.Tool { return runChainTool() }

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

type ExecuteCodeProvider struct {
	handler *handlers.CodeHandler
	enabled bool
}

func NewExecuteCodeProvider(handler *handlers.CodeHandler, enabled bool) *ExecuteCodeProvider {
	return &ExecuteCodeProvider{handler: handler, enabled: enabled}
}

func (p *ExecuteCodeProvider) Name() string   { return "execute_code" }
func (p *ExecuteCodeProvider) Enabled() bool  { return p.enabled }
func (p *ExecuteCodeProvider) Tool() mcp.Tool { return executeCodeTool() }

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
