// Package limits persists execution limits for metatools-mcp.
package limits

import (
	"context"
	"time"
)

// RuntimeLimits captures configurable execution limits that can be persisted.
type RuntimeLimits struct {
	Timeout       time.Duration
	MaxToolCalls  int
	MaxChainSteps int
}

// Store persists runtime limits.
type Store interface {
	// Load returns the stored limits. ok=false when no record exists.
	Load(ctx context.Context) (limits RuntimeLimits, ok bool, err error)
	// Save persists the given limits.
	Save(ctx context.Context, limits RuntimeLimits) error
}
