package adapters

import (
	"context"

	"github.com/jonwraymond/metatools-mcp/internal/handlers"
	"github.com/jonwraymond/toolmodel"
	"github.com/jonwraymond/toolrun"
)

// RunnerAdapter bridges toolrun.Runner to the handlers.Runner interface.
type RunnerAdapter struct {
	runner toolrun.Runner
}

// NewRunnerAdapter creates a new runner adapter.
func NewRunnerAdapter(runner toolrun.Runner) *RunnerAdapter {
	return &RunnerAdapter{runner: runner}
}

// Run delegates to toolrun.
func (a *RunnerAdapter) Run(ctx context.Context, toolID string, args map[string]any) (handlers.RunResult, error) {
	res, err := a.runner.Run(ctx, toolID, args)
	if err != nil {
		return handlers.RunResult{}, err
	}
	return handlers.RunResult{
		Structured: res.Structured,
		Backend:    res.Backend,
		Tool:       res.Tool,
		MCPResult:  res.MCPResult,
	}, nil
}

// RunChain delegates to toolrun and maps results.
func (a *RunnerAdapter) RunChain(ctx context.Context, steps []handlers.ChainStep) (handlers.RunResult, []handlers.StepResult, error) {
	chainSteps := make([]toolrun.ChainStep, len(steps))
	for i, s := range steps {
		chainSteps[i] = toolrun.ChainStep{
			ToolID:      s.ToolID,
			Args:        s.Args,
			UsePrevious: s.UsePrevious,
		}
	}

	final, stepResults, err := a.runner.RunChain(ctx, chainSteps)

	mappedSteps := make([]handlers.StepResult, len(stepResults))
	for i, sr := range stepResults {
		backend := pickBackend(sr)
		var backendAny any
		if backend.Kind != "" {
			backendAny = backend
		}
		mappedSteps[i] = handlers.StepResult{
			ToolID:     sr.ToolID,
			Structured: sr.Result.Structured,
			Backend:    backendAny,
			Tool:       sr.Result.Tool,
			Error:      sr.Err,
		}
	}

	return handlers.RunResult{
		Structured: final.Structured,
		Backend:    final.Backend,
		Tool:       final.Tool,
		MCPResult:  final.MCPResult,
	}, mappedSteps, err
}

// RunWithProgress delegates to toolrun when progress is supported.
func (a *RunnerAdapter) RunWithProgress(ctx context.Context, toolID string, args map[string]any, onProgress func(handlers.ProgressEvent)) (handlers.RunResult, error) {
	if pr, ok := a.runner.(toolrun.ProgressRunner); ok {
		result, err := pr.RunWithProgress(ctx, toolID, args, func(ev toolrun.ProgressEvent) {
			if onProgress != nil {
				onProgress(handlers.ProgressEvent{
					Progress: ev.Progress,
					Total:    ev.Total,
					Message:  ev.Message,
				})
			}
		})
		if err != nil {
			return handlers.RunResult{}, err
		}
		return handlers.RunResult{
			Structured: result.Structured,
			Backend:    result.Backend,
			Tool:       result.Tool,
			MCPResult:  result.MCPResult,
		}, nil
	}

	// Fallback to non-progress execution.
	res, err := a.runner.Run(ctx, toolID, args)
	if err != nil {
		return handlers.RunResult{}, err
	}
	return handlers.RunResult{
		Structured: res.Structured,
		Backend:    res.Backend,
		Tool:       res.Tool,
		MCPResult:  res.MCPResult,
	}, nil
}

// RunChainWithProgress delegates to toolrun when progress is supported.
func (a *RunnerAdapter) RunChainWithProgress(ctx context.Context, steps []handlers.ChainStep, onProgress func(handlers.ProgressEvent)) (handlers.RunResult, []handlers.StepResult, error) {
	if pr, ok := a.runner.(toolrun.ProgressRunner); ok {
		chainSteps := make([]toolrun.ChainStep, len(steps))
		for i, s := range steps {
			chainSteps[i] = toolrun.ChainStep{
				ToolID:      s.ToolID,
				Args:        s.Args,
				UsePrevious: s.UsePrevious,
			}
		}

		final, stepResults, err := pr.RunChainWithProgress(ctx, chainSteps, func(ev toolrun.ProgressEvent) {
			if onProgress != nil {
				onProgress(handlers.ProgressEvent{
					Progress: ev.Progress,
					Total:    ev.Total,
					Message:  ev.Message,
				})
			}
		})

		mappedSteps := make([]handlers.StepResult, len(stepResults))
		for i, sr := range stepResults {
			backend := pickBackend(sr)
			var backendAny any
			if backend.Kind != "" {
				backendAny = backend
			}
			mappedSteps[i] = handlers.StepResult{
				ToolID:     sr.ToolID,
				Structured: sr.Result.Structured,
				Backend:    backendAny,
				Tool:       sr.Result.Tool,
				Error:      sr.Err,
			}
		}

		return handlers.RunResult{
			Structured: final.Structured,
			Backend:    final.Backend,
			Tool:       final.Tool,
			MCPResult:  final.MCPResult,
		}, mappedSteps, err
	}

	// Fallback to non-progress chain execution.
	return a.RunChain(ctx, steps)
}

func pickBackend(sr toolrun.StepResult) toolmodel.ToolBackend {
	if sr.Backend.Kind != "" {
		return sr.Backend
	}
	return sr.Result.Backend
}
