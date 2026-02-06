package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jonwraymond/metatools-mcp/internal/handlers"
	"github.com/jonwraymond/tooldiscovery/index"
	"github.com/jonwraymond/toolops/cache"
	"github.com/jonwraymond/toolops/observe"
	"github.com/jonwraymond/toolops/resilience"
)

// WrapRunner applies toolops middleware to a runner.
func WrapRunner(base handlers.Runner, idx index.Index, cfg Config) (handlers.Runner, error) {
	if base == nil {
		return nil, nil
	}
	toolops, err := buildToolops(cfg)
	if err != nil {
		return nil, err
	}
	if toolops == nil {
		return base, nil
	}
	return &toolopsRunner{
		base:       base,
		index:      idx,
		observe:    toolops.observe,
		cache:      toolops.cache,
		resilience: toolops.resilience,
	}, nil
}

// WrapExecutor applies toolops middleware to a code executor.
func WrapExecutor(base handlers.Executor, cfg Config) (handlers.Executor, error) {
	if base == nil {
		return nil, nil
	}
	toolops, err := buildToolops(cfg)
	if err != nil {
		return nil, err
	}
	if toolops == nil || (toolops.observe == nil && toolops.resilience == nil) {
		return base, nil
	}
	return &toolopsExecutor{
		base:       base,
		observe:    toolops.observe,
		resilience: toolops.resilience,
	}, nil
}

type toolopsBundle struct {
	observe    *observe.Middleware
	cache      *cache.CacheMiddleware
	resilience *resilience.Executor
}

func buildToolops(cfg Config) (*toolopsBundle, error) {
	var obs *observe.Middleware
	if cfg.Observe.Enabled {
		observer, err := observe.NewObserver(context.Background(), cfg.Observe.Config)
		if err != nil {
			return nil, fmt.Errorf("observe: %w", err)
		}
		obs, err = observe.MiddlewareFromObserver(observer)
		if err != nil {
			return nil, fmt.Errorf("observe middleware: %w", err)
		}
	}

	var cacheMW *cache.CacheMiddleware
	if cfg.Cache.Enabled {
		policy := cfg.Cache.Policy
		if policy == (cache.Policy{}) {
			policy = cache.DefaultPolicy()
		}
		cacheMW = cache.NewCacheMiddleware(cache.NewMemoryCache(policy), cache.NewDefaultKeyer(), policy, nil)
	}

	var resil *resilience.Executor
	if cfg.Resilience.Enabled {
		opts := []resilience.ExecutorOption{}
		if cfg.Resilience.Retry.Enabled {
			opts = append(opts, resilience.WithRetry(resilience.NewRetry(cfg.Resilience.Retry.Config)))
		}
		if cfg.Resilience.Circuit.Enabled {
			opts = append(opts, resilience.WithCircuitBreaker(resilience.NewCircuitBreaker(cfg.Resilience.Circuit.Config)))
		}
		if cfg.Resilience.Timeout > 0 {
			opts = append(opts, resilience.WithTimeout(cfg.Resilience.Timeout))
		}
		resil = resilience.NewExecutor(opts...)
	}

	if obs == nil && cacheMW == nil && resil == nil {
		return nil, nil
	}
	return &toolopsBundle{observe: obs, cache: cacheMW, resilience: resil}, nil
}

type toolopsRunner struct {
	base       handlers.Runner
	index      index.Index
	observe    *observe.Middleware
	cache      *cache.CacheMiddleware
	resilience *resilience.Executor
}

func (r *toolopsRunner) Run(ctx context.Context, toolID string, args map[string]any) (handlers.RunResult, error) {
	meta, tags := resolveToolMeta(r.index, toolID)
	exec := func(ctx context.Context) (handlers.RunResult, error) {
		return r.base.Run(ctx, toolID, args)
	}

	if r.resilience != nil {
		exec = wrapResilience(exec, r.resilience)
	}
	if r.cache != nil {
		exec = wrapCache(exec, r.cache, toolID, args, tags)
	}
	if r.observe != nil {
		exec = wrapObserve(exec, r.observe, meta, args)
	}

	return exec(ctx)
}

func (r *toolopsRunner) RunChain(ctx context.Context, steps []handlers.ChainStep) (handlers.RunResult, []handlers.StepResult, error) {
	meta := observe.ToolMeta{ID: "run_chain", Name: "run_chain"}
	exec := func(ctx context.Context) (handlers.RunResult, []handlers.StepResult, error) {
		return r.base.RunChain(ctx, steps)
	}

	if r.resilience != nil {
		exec = wrapResilienceChain(exec, r.resilience)
	}
	if r.observe != nil {
		exec = wrapObserveChain(exec, r.observe, meta, steps)
	}

	return exec(ctx)
}

func (r *toolopsRunner) RunWithProgress(ctx context.Context, toolID string, args map[string]any, onProgress func(handlers.ProgressEvent)) (handlers.RunResult, error) {
	if pr, ok := r.base.(handlers.ProgressRunner); ok {
		meta, _ := resolveToolMeta(r.index, toolID)
		exec := func(ctx context.Context) (handlers.RunResult, error) {
			return pr.RunWithProgress(ctx, toolID, args, onProgress)
		}
		if r.resilience != nil {
			exec = wrapResilience(exec, r.resilience)
		}
		if r.observe != nil {
			exec = wrapObserve(exec, r.observe, meta, args)
		}
		return exec(ctx)
	}
	return r.Run(ctx, toolID, args)
}

func (r *toolopsRunner) RunChainWithProgress(ctx context.Context, steps []handlers.ChainStep, onProgress func(handlers.ProgressEvent)) (handlers.RunResult, []handlers.StepResult, error) {
	if pr, ok := r.base.(handlers.ProgressRunner); ok {
		meta := observe.ToolMeta{ID: "run_chain", Name: "run_chain"}
		exec := func(ctx context.Context) (handlers.RunResult, []handlers.StepResult, error) {
			return pr.RunChainWithProgress(ctx, steps, onProgress)
		}
		if r.resilience != nil {
			exec = wrapResilienceChain(exec, r.resilience)
		}
		if r.observe != nil {
			exec = wrapObserveChain(exec, r.observe, meta, steps)
		}
		return exec(ctx)
	}
	return r.RunChain(ctx, steps)
}

type toolopsExecutor struct {
	base       handlers.Executor
	observe    *observe.Middleware
	resilience *resilience.Executor
}

func (e *toolopsExecutor) ExecuteCode(ctx context.Context, params handlers.ExecuteParams) (handlers.ExecuteResult, error) {
	meta := observe.ToolMeta{ID: "execute_code", Name: "execute_code"}
	exec := func(ctx context.Context) (handlers.ExecuteResult, error) {
		return e.base.ExecuteCode(ctx, params)
	}
	if e.resilience != nil {
		exec = wrapResilienceExec(exec, e.resilience)
	}
	if e.observe != nil {
		exec = wrapObserveExec(exec, e.observe, meta, params)
	}
	return exec(ctx)
}

func resolveToolMeta(idx index.Index, toolID string) (observe.ToolMeta, []string) {
	if idx != nil {
		tool, _, err := idx.GetTool(toolID)
		if err == nil {
			return observe.ToolMeta{
				ID:        toolID,
				Name:      tool.Name,
				Namespace: tool.Namespace,
				Version:   tool.Version,
				Tags:      tool.Tags,
				Category:  metaString(tool.Meta, "category"),
			}, tool.Tags
		}
	}

	name := toolID
	namespace := ""
	if strings.Contains(toolID, ":") {
		parts := strings.SplitN(toolID, ":", 2)
		namespace = parts[0]
		name = parts[1]
	}
	return observe.ToolMeta{
		ID:        toolID,
		Name:      name,
		Namespace: namespace,
	}, nil
}

func metaString(meta map[string]any, key string) string {
	if meta == nil {
		return ""
	}
	if v, ok := meta[key].(string); ok {
		return v
	}
	return ""
}

func wrapResilience(next func(context.Context) (handlers.RunResult, error), resil *resilience.Executor) func(context.Context) (handlers.RunResult, error) {
	return func(ctx context.Context) (handlers.RunResult, error) {
		var result handlers.RunResult
		err := resil.Execute(ctx, func(ctx context.Context) error {
			var innerErr error
			result, innerErr = next(ctx)
			return innerErr
		})
		return result, err
	}
}

func wrapResilienceChain(next func(context.Context) (handlers.RunResult, []handlers.StepResult, error), resil *resilience.Executor) func(context.Context) (handlers.RunResult, []handlers.StepResult, error) {
	return func(ctx context.Context) (handlers.RunResult, []handlers.StepResult, error) {
		var result handlers.RunResult
		var steps []handlers.StepResult
		err := resil.Execute(ctx, func(ctx context.Context) error {
			var innerErr error
			result, steps, innerErr = next(ctx)
			return innerErr
		})
		return result, steps, err
	}
}

func wrapResilienceExec(next func(context.Context) (handlers.ExecuteResult, error), resil *resilience.Executor) func(context.Context) (handlers.ExecuteResult, error) {
	return func(ctx context.Context) (handlers.ExecuteResult, error) {
		var result handlers.ExecuteResult
		err := resil.Execute(ctx, func(ctx context.Context) error {
			var innerErr error
			result, innerErr = next(ctx)
			return innerErr
		})
		return result, err
	}
}

func wrapCache(next func(context.Context) (handlers.RunResult, error), cacheMW *cache.CacheMiddleware, toolID string, args map[string]any, tags []string) func(context.Context) (handlers.RunResult, error) {
	return func(ctx context.Context) (handlers.RunResult, error) {
		executor := func(ctx context.Context, _ string, _ any) ([]byte, error) {
			result, err := next(ctx)
			if err != nil {
				return nil, err
			}
			payload, err := json.Marshal(result)
			if err != nil {
				return nil, err
			}
			return payload, nil
		}

		payload, err := cacheMW.Execute(ctx, toolID, args, tags, executor)
		if err != nil {
			return handlers.RunResult{}, err
		}
		if len(payload) == 0 {
			return handlers.RunResult{}, errors.New("cache payload empty")
		}

		var result handlers.RunResult
		if err := json.Unmarshal(payload, &result); err != nil {
			return handlers.RunResult{}, err
		}
		return result, nil
	}
}

func wrapObserve(next func(context.Context) (handlers.RunResult, error), obs *observe.Middleware, meta observe.ToolMeta, input any) func(context.Context) (handlers.RunResult, error) {
	return func(ctx context.Context) (handlers.RunResult, error) {
		if obs == nil {
			return next(ctx)
		}
		wrapped := obs.Wrap(func(ctx context.Context, _ observe.ToolMeta, _ any) (any, error) {
			return next(ctx)
		})
		value, err := wrapped(ctx, meta, input)
		if err != nil {
			return handlers.RunResult{}, err
		}
		result, ok := value.(handlers.RunResult)
		if !ok {
			return handlers.RunResult{}, errors.New("unexpected result type from observe middleware")
		}
		return result, nil
	}
}

func wrapObserveChain(next func(context.Context) (handlers.RunResult, []handlers.StepResult, error), obs *observe.Middleware, meta observe.ToolMeta, steps []handlers.ChainStep) func(context.Context) (handlers.RunResult, []handlers.StepResult, error) {
	return func(ctx context.Context) (handlers.RunResult, []handlers.StepResult, error) {
		if obs == nil {
			return next(ctx)
		}
		wrapped := obs.Wrap(func(ctx context.Context, _ observe.ToolMeta, _ any) (any, error) {
			result, results, err := next(ctx)
			if err != nil {
				return nil, err
			}
			return struct {
				Result handlers.RunResult
				Steps  []handlers.StepResult
			}{Result: result, Steps: results}, nil
		})
		value, err := wrapped(ctx, meta, steps)
		if err != nil {
			return handlers.RunResult{}, nil, err
		}
		typed, ok := value.(struct {
			Result handlers.RunResult
			Steps  []handlers.StepResult
		})
		if !ok {
			return handlers.RunResult{}, nil, errors.New("unexpected result type from observe middleware")
		}
		return typed.Result, typed.Steps, nil
	}
}

func wrapObserveExec(next func(context.Context) (handlers.ExecuteResult, error), obs *observe.Middleware, meta observe.ToolMeta, params handlers.ExecuteParams) func(context.Context) (handlers.ExecuteResult, error) {
	return func(ctx context.Context) (handlers.ExecuteResult, error) {
		if obs == nil {
			return next(ctx)
		}
		wrapped := obs.Wrap(func(ctx context.Context, _ observe.ToolMeta, _ any) (any, error) {
			return next(ctx)
		})
		value, err := wrapped(ctx, meta, params)
		if err != nil {
			return handlers.ExecuteResult{}, err
		}
		result, ok := value.(handlers.ExecuteResult)
		if !ok {
			return handlers.ExecuteResult{}, errors.New("unexpected result type from observe middleware")
		}
		return result, nil
	}
}
