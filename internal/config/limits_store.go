package config

import (
	"context"
	"fmt"

	"github.com/jonwraymond/metatools-mcp/internal/state/limits"
)

// ApplyRuntimeLimitsStore loads or seeds persisted runtime limits if configured.
func (c *AppConfig) ApplyRuntimeLimitsStore(ctx context.Context) error {
	if c == nil {
		return nil
	}
	if c.State.RuntimeLimitsDB == "" {
		return nil
	}

	store, closeFn, err := limits.OpenSQLite(c.State.RuntimeLimitsDB)
	if err != nil {
		return fmt.Errorf("open runtime limits store: %w", err)
	}
	defer func() {
		if closeFn != nil {
			_ = closeFn()
		}
	}()

	loaded, ok, err := store.Load(ctx)
	if err != nil {
		return fmt.Errorf("load runtime limits: %w", err)
	}
	if ok {
		c.Execution.Timeout = loaded.Timeout
		c.Execution.MaxToolCalls = loaded.MaxToolCalls
		c.Execution.MaxChainSteps = loaded.MaxChainSteps
		return nil
	}

	return store.Save(ctx, limits.RuntimeLimits{
		Timeout:       c.Execution.Timeout,
		MaxToolCalls:  c.Execution.MaxToolCalls,
		MaxChainSteps: c.Execution.MaxChainSteps,
	})
}
