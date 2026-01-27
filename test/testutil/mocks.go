package testutil

import (
	"context"

	"github.com/jonwraymond/metatools-mcp/pkg/metatools"
)

// MockIndex implements a mock tool index for testing
type MockIndex struct {
	SearchFunc         func(ctx context.Context, query string, limit int) ([]metatools.ToolSummary, error)
	ListNamespacesFunc func(ctx context.Context) ([]string, error)
	GetToolFunc        func(ctx context.Context, id string) (Tool, Backend, error)
}

// Tool represents a tool definition for testing
type Tool struct {
	ID          string
	Name        string
	Namespace   string
	Description string
	InputSchema map[string]any
}

// Backend represents a backend for testing
type Backend struct {
	Kind       string
	ServerName string
	ProviderID string
}

func (m *MockIndex) Search(ctx context.Context, query string, limit int) ([]metatools.ToolSummary, error) {
	if m.SearchFunc != nil {
		return m.SearchFunc(ctx, query, limit)
	}
	return []metatools.ToolSummary{}, nil
}

func (m *MockIndex) ListNamespaces(ctx context.Context) ([]string, error) {
	if m.ListNamespacesFunc != nil {
		return m.ListNamespacesFunc(ctx)
	}
	return []string{}, nil
}

func (m *MockIndex) GetTool(ctx context.Context, id string) (Tool, Backend, error) {
	if m.GetToolFunc != nil {
		return m.GetToolFunc(ctx, id)
	}
	return Tool{}, Backend{}, nil
}

// MockStore implements a mock documentation store for testing
type MockStore struct {
	DescribeToolFunc func(ctx context.Context, id string, level string) (ToolDoc, error)
	ListExamplesFunc func(ctx context.Context, id string, max int) ([]metatools.ToolExample, error)
}

// ToolDoc represents a tool documentation for testing
type ToolDoc struct {
	Tool         any
	Summary      string
	SchemaInfo   any
	Notes        *string
	Examples     []metatools.ToolExample
	ExternalRefs []string
}

func (m *MockStore) DescribeTool(ctx context.Context, id string, level string) (ToolDoc, error) {
	if m.DescribeToolFunc != nil {
		return m.DescribeToolFunc(ctx, id, level)
	}
	return ToolDoc{}, nil
}

func (m *MockStore) ListExamples(ctx context.Context, id string, max int) ([]metatools.ToolExample, error) {
	if m.ListExamplesFunc != nil {
		return m.ListExamplesFunc(ctx, id, max)
	}
	return []metatools.ToolExample{}, nil
}

// MockRunner implements a mock tool runner for testing
type MockRunner struct {
	RunFunc      func(ctx context.Context, toolID string, args map[string]any) (RunResult, error)
	RunChainFunc func(ctx context.Context, steps []ChainStep) (RunResult, []StepResult, error)
}

// RunResult represents a tool execution result for testing
type RunResult struct {
	Structured any
	Backend    Backend
	Tool       Tool
	DurationMs int
}

// ChainStep represents a chain step for testing
type ChainStep struct {
	ToolID      string
	Args        map[string]any
	UsePrevious bool
}

// StepResult represents a step result for testing
type StepResult struct {
	ToolID     string
	Structured any
	Backend    Backend
	Tool       Tool
	Error      error
}

func (m *MockRunner) Run(ctx context.Context, toolID string, args map[string]any) (RunResult, error) {
	if m.RunFunc != nil {
		return m.RunFunc(ctx, toolID, args)
	}
	return RunResult{}, nil
}

func (m *MockRunner) RunChain(ctx context.Context, steps []ChainStep) (RunResult, []StepResult, error) {
	if m.RunChainFunc != nil {
		return m.RunChainFunc(ctx, steps)
	}
	return RunResult{}, []StepResult{}, nil
}

// MockExecutor implements a mock code executor for testing
type MockExecutor struct {
	ExecuteCodeFunc func(ctx context.Context, params ExecuteParams) (ExecuteResult, error)
}

// ExecuteParams represents code execution parameters for testing
type ExecuteParams struct {
	Language     string
	Code         string
	Timeout      int // milliseconds
	MaxToolCalls int
}

// ExecuteResult represents code execution result for testing
type ExecuteResult struct {
	Value      any
	Stdout     string
	Stderr     string
	DurationMs int
}

func (m *MockExecutor) ExecuteCode(ctx context.Context, params ExecuteParams) (ExecuteResult, error) {
	if m.ExecuteCodeFunc != nil {
		return m.ExecuteCodeFunc(ctx, params)
	}
	return ExecuteResult{}, nil
}
