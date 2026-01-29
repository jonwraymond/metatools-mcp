package backend

import (
	"context"
	"errors"

	"github.com/jonwraymond/toolmodel"
)

// Common errors for backend operations.
var (
	ErrBackendNotFound    = errors.New("backend not found")
	ErrBackendDisabled    = errors.New("backend disabled")
	ErrToolNotFound       = errors.New("tool not found in backend")
	ErrBackendUnavailable = errors.New("backend unavailable")
)

// Backend defines a source of tools.
// Backends can be local handlers, MCP servers, HTTP APIs, or custom implementations.
type Backend interface {
	// Kind returns the backend type (e.g., "local", "mcp", "http").
	Kind() string

	// Name returns the unique instance name for this backend.
	Name() string

	// Enabled returns whether this backend is currently enabled.
	Enabled() bool

	// ListTools returns all tools available from this backend.
	ListTools(ctx context.Context) ([]toolmodel.Tool, error)

	// Execute invokes a tool on this backend.
	Execute(ctx context.Context, tool string, args map[string]any) (any, error)

	// Start initializes the backend (connect to remote, start subprocess, etc.).
	Start(ctx context.Context) error

	// Stop gracefully shuts down the backend.
	Stop() error
}

// ConfigurableBackend can be configured from raw bytes (YAML/JSON).
type ConfigurableBackend interface {
	Backend

	Configure(raw []byte) error
}

// StreamingBackend supports streaming responses.
type StreamingBackend interface {
	Backend

	ExecuteStream(ctx context.Context, tool string, args map[string]any) (<-chan any, error)
}

// BackendFactory creates backend instances.
type BackendFactory func(name string) (Backend, error)

// BackendInfo contains metadata about a backend.
type BackendInfo struct {
	Kind        string
	Name        string
	Enabled     bool
	Description string
	Version     string
}
