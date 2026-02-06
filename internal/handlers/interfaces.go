package handlers

import (
	"context"
	"time"

	"github.com/jonwraymond/metatools-mcp/internal/skills"
	"github.com/jonwraymond/metatools-mcp/internal/toolset"
	"github.com/jonwraymond/metatools-mcp/pkg/metatools"
	"github.com/jonwraymond/toolfoundation/model"
)

// Index provides tool search and discovery.
//
// Contract:
// - Concurrency: implementations must be safe for concurrent use.
// - Context: methods must honor cancellation/deadlines.
// - Errors: use errors.Is with toolindex/tooldocs/toolrun equivalents where applicable.
type Index interface {
	SearchPage(ctx context.Context, query string, limit int, cursor string) ([]metatools.ToolSummary, string, error)
	ListNamespacesPage(ctx context.Context, limit int, cursor string) ([]string, string, error)
	GetAllBackends(ctx context.Context, id string) ([]model.ToolBackend, error)
}

// Refresher optionally triggers backend refreshes before discovery operations.
type Refresher interface {
	MaybeRefresh(ctx context.Context) error
}

// ToolDoc represents a tool documentation result
type ToolDoc struct {
	Tool         any
	Summary      string
	SchemaInfo   any
	Notes        *string
	Examples     []metatools.ToolExample
	ExternalRefs []string
}

// Store provides tool documentation.
//
// Contract:
// - Concurrency: implementations must be safe for concurrent use.
// - Context: methods must honor cancellation/deadlines.
// - Errors: invalid IDs should return ErrNotFound/ErrInvalidDetail equivalents.
type Store interface {
	DescribeTool(ctx context.Context, id string, level string) (ToolDoc, error)
	ListExamples(ctx context.Context, id string, maxExamples int) ([]metatools.ToolExample, error)
}

// RunResult represents a tool execution result
type RunResult struct {
	Structured any
	Backend    any
	Tool       any
	MCPResult  any
	DurationMs int
}

// ChainStep represents a chain step input
type ChainStep struct {
	ToolID      string
	Args        map[string]any
	UsePrevious bool
}

// ProgressEvent represents a progress update during execution.
type ProgressEvent struct {
	Progress float64
	Total    float64
	Message  string
}

// StepResult represents a step result
type StepResult struct {
	ToolID     string
	Structured any
	Backend    any
	Tool       any
	Error      error
}

// Runner provides tool execution.
//
// Contract:
// - Concurrency: implementations must be safe for concurrent use.
// - Context: methods must honor cancellation/deadlines and return ctx.Err() when canceled.
type Runner interface {
	Run(ctx context.Context, toolID string, args map[string]any) (RunResult, error)
	RunChain(ctx context.Context, steps []ChainStep) (RunResult, []StepResult, error)
}

// ProgressRunner is an optional interface for progress-aware runners.
//
// Contract:
// - Progress callbacks must be invoked in-order; nil callbacks are allowed.
type ProgressRunner interface {
	RunWithProgress(ctx context.Context, toolID string, args map[string]any, onProgress func(ProgressEvent)) (RunResult, error)
	RunChainWithProgress(ctx context.Context, steps []ChainStep, onProgress func(ProgressEvent)) (RunResult, []StepResult, error)
}

// ExecuteParams represents code execution parameters
type ExecuteParams struct {
	Language     string
	Code         string
	TimeoutMs    int
	MaxToolCalls int
}

// ExecuteResult represents code execution result
type ExecuteResult struct {
	Value      any
	Stdout     string
	Stderr     string
	DurationMs int
}

// Executor provides code execution.
//
// Contract:
// - Concurrency: implementations must be safe for concurrent use.
// - Context: methods must honor cancellation/deadlines and return ctx.Err() when canceled.
type Executor interface {
	ExecuteCode(ctx context.Context, params ExecuteParams) (ExecuteResult, error)
}

// ToolsetRegistry provides access to configured toolsets.
type ToolsetRegistry interface {
	List() []*toolset.Toolset
	Get(id string) (*toolset.Toolset, bool)
}

// SkillRegistry provides access to configured skills.
type SkillRegistry interface {
	List() []*skills.Skill
	Get(id string) (*skills.Skill, bool)
}

// SkillDefaults defines default limits for skills.
type SkillDefaults struct {
	MaxSteps     int
	MaxToolCalls int
	Timeout      time.Duration
}
