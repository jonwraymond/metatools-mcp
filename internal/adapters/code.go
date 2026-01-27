package adapters

import (
	"context"
	"math"
	"time"

	"github.com/jonwraymond/metatools-mcp/internal/handlers"
	"github.com/jonwraymond/toolcode"
)

// ExecutorAdapter bridges toolcode.Executor to the handlers.Executor interface.
type ExecutorAdapter struct {
	exec toolcode.Executor
}

// NewExecutorAdapter creates a new executor adapter.
func NewExecutorAdapter(exec toolcode.Executor) *ExecutorAdapter {
	return &ExecutorAdapter{exec: exec}
}

// ExecuteCode delegates to toolcode and converts duration units.
func (a *ExecutorAdapter) ExecuteCode(ctx context.Context, params handlers.ExecuteParams) (handlers.ExecuteResult, error) {
	req := toolcode.ExecuteParams{
		Language:     params.Language,
		Code:         params.Code,
		MaxToolCalls: params.MaxToolCalls,
	}
	if params.TimeoutMs > 0 {
		req.Timeout = time.Duration(params.TimeoutMs) * time.Millisecond
	}

	res, err := a.exec.ExecuteCode(ctx, req)
	if err != nil {
		return handlers.ExecuteResult{}, err
	}

	dur := int(res.DurationMs)
	if res.DurationMs > int64(math.MaxInt) {
		dur = math.MaxInt
	}

	return handlers.ExecuteResult{
		Value:      res.Value,
		Stdout:     res.Stdout,
		Stderr:     res.Stderr,
		DurationMs: dur,
	}, nil
}
