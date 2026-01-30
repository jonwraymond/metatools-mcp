package wasm

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"

	wasmbackend "github.com/jonwraymond/toolruntime/backend/wasm"
)

// Client implements toolruntime's WasmRunner interface using wazero.
type Client struct {
	runtime wazero.Runtime
	config  ClientConfig
	closed  bool
}

// ClientConfig holds configuration for the wazero client.
type ClientConfig struct {
	// MaxMemoryPages is the maximum memory in 64KB pages.
	// Default: 256 (16MB)
	MaxMemoryPages uint32

	// EnableCompilationCache enables module compilation caching.
	// Default: false
	EnableCompilationCache bool
}

// NewClient creates a new wazero-based WASM client.
func NewClient(cfg ClientConfig) (*Client, error) {
	maxPages := cfg.MaxMemoryPages
	if maxPages == 0 {
		maxPages = 256 // 16MB default
	}

	runtimeConfig := wazero.NewRuntimeConfig().
		WithMemoryLimitPages(maxPages).
		WithCloseOnContextDone(true) // Critical for timeout/cancellation

	if cfg.EnableCompilationCache {
		cache := wazero.NewCompilationCache()
		runtimeConfig = runtimeConfig.WithCompilationCache(cache)
	}

	runtime := wazero.NewRuntimeWithConfig(context.Background(), runtimeConfig)

	return &Client{
		runtime: runtime,
		config:  cfg,
	}, nil
}

// Run implements WasmRunner.Run.
// It compiles, instantiates, executes, and cleans up a WASM module atomically.
func (c *Client) Run(ctx context.Context, spec wasmbackend.WasmSpec) (wasmbackend.WasmResult, error) {
	if c.closed {
		return wasmbackend.WasmResult{}, fmt.Errorf("client is closed")
	}

	start := time.Now()

	// Validate spec
	if len(spec.Module) == 0 {
		return wasmbackend.WasmResult{}, fmt.Errorf("module bytes required")
	}

	// Apply timeout if specified
	if spec.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, spec.Timeout)
		defer cancel()
	}

	// Initialize WASI if enabled (must be done before module instantiation)
	if spec.Security.EnableWASI {
		if _, err := wasi_snapshot_preview1.Instantiate(ctx, c.runtime); err != nil {
			// If already instantiated, ignore the error
			if !strings.Contains(err.Error(), "module[wasi_snapshot_preview1] has already been instantiated") {
				return wasmbackend.WasmResult{}, fmt.Errorf("instantiate WASI: %w", err)
			}
		}
	}

	// Compile module
	compiled, err := c.runtime.CompileModule(ctx, spec.Module)
	if err != nil {
		return wasmbackend.WasmResult{}, fmt.Errorf("compile module: %w", err)
	}
	defer func() { _ = compiled.Close(ctx) }()

	// Setup I/O buffers
	var stdout, stderr bytes.Buffer

	// Build module config
	moduleConfig := wazero.NewModuleConfig().
		WithStdout(&stdout).
		WithStderr(&stderr).
		WithName("main")

	// Add command-line arguments
	if len(spec.Args) > 0 {
		moduleConfig = moduleConfig.WithArgs(spec.Args...)
	}

	// Add environment variables
	for _, env := range spec.Env {
		if key, value, ok := parseEnv(env); ok {
			moduleConfig = moduleConfig.WithEnv(key, value)
		}
	}

	// Add stdin if provided
	if len(spec.Stdin) > 0 {
		moduleConfig = moduleConfig.WithStdin(bytes.NewReader(spec.Stdin))
	}

	// Configure WASI filesystem mounts
	if spec.Security.EnableWASI && len(spec.Mounts) > 0 {
		fsConfig := wazero.NewFSConfig()
		for _, mount := range spec.Mounts {
			if mount.ReadOnly {
				fsConfig = fsConfig.WithReadOnlyDirMount(mount.HostPath, mount.GuestPath)
			} else {
				fsConfig = fsConfig.WithDirMount(mount.HostPath, mount.GuestPath)
			}
		}
		moduleConfig = moduleConfig.WithFSConfig(fsConfig)
	}

	// Instantiate and run module
	// For WASI modules, instantiation runs _start automatically
	mod, err := c.runtime.InstantiateModule(ctx, compiled, moduleConfig)
	if err != nil {
		// Check if it's a context error
		if ctx.Err() != nil {
			return wasmbackend.WasmResult{
				Stdout:   stdout.String(),
				Stderr:   stderr.String(),
				Duration: time.Since(start),
			}, ctx.Err()
		}

		// Try to extract exit code from error
		exitCode := extractExitCode(err)
		if exitCode != 0 {
			return wasmbackend.WasmResult{
				ExitCode: exitCode,
				Stdout:   stdout.String(),
				Stderr:   stderr.String(),
				Duration: time.Since(start),
			}, nil
		}

		return wasmbackend.WasmResult{
			ExitCode: 1,
			Stdout:   stdout.String(),
			Stderr:   stderr.String() + "\n" + err.Error(),
			Duration: time.Since(start),
		}, fmt.Errorf("instantiate module: %w", err)
	}

	// Module ran successfully, close it
	if mod != nil {
		_ = mod.Close(ctx)
	}

	return wasmbackend.WasmResult{
		ExitCode: 0,
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		Duration: time.Since(start),
	}, nil
}

// Close closes the wazero runtime and releases all resources.
func (c *Client) Close(ctx context.Context) error {
	c.closed = true
	return c.runtime.Close(ctx)
}

// Runtime returns the underlying wazero runtime for advanced use.
func (c *Client) Runtime() wazero.Runtime {
	return c.runtime
}

// parseEnv parses an environment variable string in KEY=value format.
func parseEnv(env string) (key, value string, ok bool) {
	for i := 0; i < len(env); i++ {
		if env[i] == '=' {
			return env[:i], env[i+1:], true
		}
	}
	return "", "", false
}

// extractExitCode tries to extract an exit code from an error.
// wazero wraps exit codes in sys.ExitError.
func extractExitCode(err error) int {
	if err == nil {
		return 0
	}

	// wazero uses sys.ExitError for non-zero exits
	errStr := err.Error()
	if strings.Contains(errStr, "exit_code(") {
		// Parse "exit_code(42)" pattern
		start := strings.Index(errStr, "exit_code(")
		if start >= 0 {
			start += len("exit_code(")
			end := strings.Index(errStr[start:], ")")
			if end > 0 {
				var code int
				if _, parseErr := fmt.Sscanf(errStr[start:start+end], "%d", &code); parseErr == nil {
					return code
				}
			}
		}
	}

	return 0
}

// Ensure interface compliance at compile time.
var _ wasmbackend.WasmRunner = (*Client)(nil)
