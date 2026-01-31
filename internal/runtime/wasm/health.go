package wasm

import (
	"context"

	"github.com/tetratelabs/wazero"

	wasmbackend "github.com/jonwraymond/toolexec/runtime/backend/wasm"
)

// wazeroVersion is the version of the wazero runtime.
// This is hardcoded as wazero doesn't expose a version constant.
const wazeroVersion = "1.11.0"

// HealthCheck implements toolruntime's HealthChecker interface for wazero.
type HealthCheck struct {
	client *Client
}

// NewHealthCheck creates a new health checker for the given client.
func NewHealthCheck(client *Client) *HealthCheck {
	return &HealthCheck{client: client}
}

// Ping checks if the wazero runtime is operational.
// Since wazero is a pure-Go library, this mainly checks that the client
// is properly initialized and not closed.
func (h *HealthCheck) Ping(ctx context.Context) error {
	if h.client == nil {
		return wasmbackend.ErrWASMRuntimeNotAvailable
	}
	if h.client.closed {
		return wasmbackend.ErrWASMRuntimeNotAvailable
	}

	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Verify runtime is accessible
	if h.client.runtime == nil {
		return wasmbackend.ErrWASMRuntimeNotAvailable
	}

	return nil
}

// Info returns runtime information about the wazero runtime.
func (h *HealthCheck) Info(_ context.Context) (wasmbackend.RuntimeInfo, error) {
	if h.client == nil || h.client.closed {
		return wasmbackend.RuntimeInfo{}, wasmbackend.ErrWASMRuntimeNotAvailable
	}

	features := []string{
		"wasi_snapshot_preview1",
		"memory_limits",
		"compilation_cache",
		"context_cancellation",
	}

	return wasmbackend.RuntimeInfo{
		Name:     "wazero",
		Version:  wazeroVersion,
		Features: features,
	}, nil
}

// Ensure interface compliance at compile time.
var _ wasmbackend.HealthChecker = (*HealthCheck)(nil)

// NewHealthCheckFromRuntime creates a health checker from a raw wazero runtime.
// This is useful when you have access to the runtime but not the Client.
func NewHealthCheckFromRuntime(runtime wazero.Runtime) *RuntimeHealthCheck {
	return &RuntimeHealthCheck{runtime: runtime}
}

// RuntimeHealthCheck implements HealthChecker for a raw wazero.Runtime.
type RuntimeHealthCheck struct {
	runtime wazero.Runtime
}

// Ping checks if the wazero runtime is operational.
func (h *RuntimeHealthCheck) Ping(ctx context.Context) error {
	if h.runtime == nil {
		return wasmbackend.ErrWASMRuntimeNotAvailable
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	return nil
}

// Info returns runtime information.
func (h *RuntimeHealthCheck) Info(_ context.Context) (wasmbackend.RuntimeInfo, error) {
	if h.runtime == nil {
		return wasmbackend.RuntimeInfo{}, wasmbackend.ErrWASMRuntimeNotAvailable
	}

	return wasmbackend.RuntimeInfo{
		Name:     "wazero",
		Version:  wazeroVersion,
		Features: []string{"wasi_snapshot_preview1", "memory_limits"},
	}, nil
}

var _ wasmbackend.HealthChecker = (*RuntimeHealthCheck)(nil)
