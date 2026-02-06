package wasm

import (
	"context"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"

	wasmbackend "github.com/jonwraymond/toolexec/runtime/backend/wasm"
)

// StreamClient extends Client with streaming capabilities.
// It implements both Runner and StreamRunner interfaces.
type StreamClient struct {
	*Client
}

// NewStreamClient creates a client with streaming support.
func NewStreamClient(cfg ClientConfig) (*StreamClient, error) {
	client, err := NewClient(cfg)
	if err != nil {
		return nil, err
	}
	return &StreamClient{Client: client}, nil
}

// RunStream implements StreamRunner.RunStream with real-time output streaming.
// It returns a channel that receives streaming events for stdout, stderr, and exit.
func (c *StreamClient) RunStream(ctx context.Context, spec wasmbackend.Spec) (<-chan wasmbackend.StreamEvent, error) {
	if c.closed {
		return nil, wasmbackend.ErrWASMRuntimeNotAvailable
	}

	events := make(chan wasmbackend.StreamEvent, 100)

	go c.runStreamInternal(ctx, spec, events)

	return events, nil
}

func (c *StreamClient) runStreamInternal(ctx context.Context, spec wasmbackend.Spec, events chan<- wasmbackend.StreamEvent) {
	start := time.Now()

	// Validate spec
	if len(spec.Module) == 0 {
		events <- wasmbackend.StreamEvent{
			Type:  wasmbackend.StreamEventError,
			Error: wasmbackend.ErrInvalidModule,
		}
		close(events)
		return
	}

	// Apply timeout
	if spec.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, spec.Timeout)
		defer cancel()
	}

	// Initialize WASI if enabled
	if spec.Security.EnableWASI {
		if _, err := wasi_snapshot_preview1.Instantiate(ctx, c.runtime); err != nil {
			if !strings.Contains(err.Error(), "module[wasi_snapshot_preview1] has already been instantiated") {
				events <- wasmbackend.StreamEvent{
					Type:  wasmbackend.StreamEventError,
					Error: err,
				}
				close(events)
				return
			}
		}
	}

	// Compile module
	compiled, err := c.runtime.CompileModule(ctx, spec.Module)
	if err != nil {
		events <- wasmbackend.StreamEvent{
			Type:  wasmbackend.StreamEventError,
			Error: err,
		}
		close(events)
		return
	}
	defer func() { _ = compiled.Close(ctx) }()

	// Create streaming pipes for stdout and stderr
	stdoutReader, stdoutWriter := io.Pipe()
	stderrReader, stderrWriter := io.Pipe()

	// Build module config with piped outputs
	moduleConfig := wazero.NewModuleConfig().
		WithStdout(stdoutWriter).
		WithStderr(stderrWriter).
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

	var wg sync.WaitGroup
	wg.Add(2)

	// Stream stdout in a goroutine
	go func() {
		defer wg.Done()
		buf := make([]byte, 1024)
		for {
			n, readErr := stdoutReader.Read(buf)
			if n > 0 {
				data := make([]byte, n)
				copy(data, buf[:n])
				events <- wasmbackend.StreamEvent{
					Type: wasmbackend.StreamEventStdout,
					Data: data,
				}
			}
			if readErr != nil {
				break
			}
		}
	}()

	// Stream stderr in a goroutine
	go func() {
		defer wg.Done()
		buf := make([]byte, 1024)
		for {
			n, readErr := stderrReader.Read(buf)
			if n > 0 {
				data := make([]byte, n)
				copy(data, buf[:n])
				events <- wasmbackend.StreamEvent{
					Type: wasmbackend.StreamEventStderr,
					Data: data,
				}
			}
			if readErr != nil {
				break
			}
		}
	}()

	// Instantiate and run module
	mod, err := c.runtime.InstantiateModule(ctx, compiled, moduleConfig)

	// Close writers to signal EOF to readers
	_ = stdoutWriter.Close()
	_ = stderrWriter.Close()

	// Ensure stdout/stderr stream goroutines finish before we emit the terminal event and close channel.
	wg.Wait()

	if err != nil {
		if mod != nil {
			_ = mod.Close(ctx)
		}

		// Check if it's a context error
		if ctx.Err() != nil {
			events <- wasmbackend.StreamEvent{
				Type:  wasmbackend.StreamEventError,
				Error: ctx.Err(),
			}
			close(events)
			return
		}

		// Try to extract exit code from error
		exitCode := extractExitCode(err)
		if exitCode != 0 {
			events <- wasmbackend.StreamEvent{
				Type:     wasmbackend.StreamEventExit,
				ExitCode: exitCode,
			}
			close(events)
			return
		}

		events <- wasmbackend.StreamEvent{
			Type:  wasmbackend.StreamEventError,
			Error: err,
		}
		close(events)
		return
	}

	// Module ran successfully
	if mod != nil {
		_ = mod.Close(ctx)
	}

	_ = start // Used for potential duration tracking

	events <- wasmbackend.StreamEvent{
		Type:     wasmbackend.StreamEventExit,
		ExitCode: 0,
	}
	close(events)
}

// Ensure interface compliance at compile time.
var _ wasmbackend.StreamRunner = (*StreamClient)(nil)
