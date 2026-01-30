package wasm

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	wasmbackend "github.com/jonwraymond/toolruntime/backend/wasm"
)

// Test WASM binaries are embedded at package level for use across tests.
// These are minimal WASI programs compiled from WAT source files in testdata/.
// See testdata.go for the embed directives.

// helloWasm prints "hello\n" to stdout and exits 0.
var helloWasm = helloWasmBytes

// exitCodeWasm exits with code 42.
var exitCodeWasm = exit42WasmBytes

// stderrWasm prints "error\n" to stderr.
var stderrWasm = stderrWasmBytes

// infiniteLoopWasm runs forever (for timeout testing).
var infiniteLoopWasm = infiniteLoopWasmBytes

// These modules require more complex WASI setup and are not yet implemented.
var echoArgsWasm []byte   // TODO: implement
var echoEnvWasm []byte    // TODO: implement
var echoStdinWasm []byte  // TODO: implement
var allocMemoryWasm []byte // TODO: implement

func skipIfNoTestWasm(t *testing.T, module []byte) {
	t.Helper()
	if module == nil {
		t.Skip("test WASM binary not available")
	}
}

// TestClient_Run_Success tests basic successful execution.
func TestClient_Run_Success(t *testing.T) {
	skipIfNoTestWasm(t, helloWasm)

	client, err := NewClient(ClientConfig{})
	require.NoError(t, err)
	defer func() { _ = client.Close(context.Background()) }()

	result, err := client.Run(context.Background(), wasmbackend.WasmSpec{
		Module: helloWasm,
		Security: wasmbackend.WasmSecuritySpec{
			EnableWASI: true,
		},
	})

	require.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Stdout, "hello")
	assert.Empty(t, result.Stderr)
	assert.Greater(t, result.Duration, time.Duration(0))
}

// TestClient_Run_NonZeroExit tests that non-zero exit codes are captured.
func TestClient_Run_NonZeroExit(t *testing.T) {
	skipIfNoTestWasm(t, exitCodeWasm)

	client, err := NewClient(ClientConfig{})
	require.NoError(t, err)
	defer func() { _ = client.Close(context.Background()) }()

	result, err := client.Run(context.Background(), wasmbackend.WasmSpec{
		Module: exitCodeWasm,
		Security: wasmbackend.WasmSecuritySpec{
			EnableWASI: true,
		},
	})

	require.NoError(t, err)
	assert.Equal(t, 42, result.ExitCode)
}

// TestClient_Run_Stderr tests that stderr is captured separately.
func TestClient_Run_Stderr(t *testing.T) {
	skipIfNoTestWasm(t, stderrWasm)

	client, err := NewClient(ClientConfig{})
	require.NoError(t, err)
	defer func() { _ = client.Close(context.Background()) }()

	result, err := client.Run(context.Background(), wasmbackend.WasmSpec{
		Module: stderrWasm,
		Security: wasmbackend.WasmSecuritySpec{
			EnableWASI: true,
		},
	})

	require.NoError(t, err)
	assert.Contains(t, result.Stderr, "error")
}

// TestClient_Run_Timeout tests that spec.Timeout is respected.
func TestClient_Run_Timeout(t *testing.T) {
	skipIfNoTestWasm(t, infiniteLoopWasm)

	client, err := NewClient(ClientConfig{})
	require.NoError(t, err)
	defer func() { _ = client.Close(context.Background()) }()

	ctx := context.Background()
	result, err := client.Run(ctx, wasmbackend.WasmSpec{
		Module:  infiniteLoopWasm,
		Timeout: 100 * time.Millisecond,
		Security: wasmbackend.WasmSecuritySpec{
			EnableWASI: true,
		},
	})

	// Should return an error due to timeout
	require.Error(t, err)
	assert.True(t, errors.Is(err, context.DeadlineExceeded) ||
		strings.Contains(err.Error(), "timeout") ||
		strings.Contains(err.Error(), "deadline"))
	_ = result // Result may be partial
}

// TestClient_Run_ContextCancel tests that context cancellation stops execution.
func TestClient_Run_ContextCancel(t *testing.T) {
	skipIfNoTestWasm(t, infiniteLoopWasm)

	client, err := NewClient(ClientConfig{})
	require.NoError(t, err)
	defer func() { _ = client.Close(context.Background()) }()

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel after a short delay
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	_, err = client.Run(ctx, wasmbackend.WasmSpec{
		Module: infiniteLoopWasm,
		Security: wasmbackend.WasmSecuritySpec{
			EnableWASI: true,
		},
	})

	require.Error(t, err)
	assert.True(t, errors.Is(err, context.Canceled))
}

// TestClient_Run_InvalidModule tests error handling for invalid WASM.
func TestClient_Run_InvalidModule(t *testing.T) {
	client, err := NewClient(ClientConfig{})
	require.NoError(t, err)
	defer func() { _ = client.Close(context.Background()) }()

	_, err = client.Run(context.Background(), wasmbackend.WasmSpec{
		Module: []byte("not a valid wasm module"),
		Security: wasmbackend.WasmSecuritySpec{
			EnableWASI: true,
		},
	})

	require.Error(t, err)
	// Error should indicate compilation/validation failure
	assert.True(t,
		strings.Contains(err.Error(), "compile") ||
			strings.Contains(err.Error(), "invalid") ||
			strings.Contains(err.Error(), "magic"))
}

// TestClient_Run_EmptyModule tests error handling for empty module.
func TestClient_Run_EmptyModule(t *testing.T) {
	client, err := NewClient(ClientConfig{})
	require.NoError(t, err)
	defer func() { _ = client.Close(context.Background()) }()

	_, err = client.Run(context.Background(), wasmbackend.WasmSpec{
		Module: nil,
		Security: wasmbackend.WasmSecuritySpec{
			EnableWASI: true,
		},
	})

	require.Error(t, err)
}

// TestClient_Run_Args tests that command-line arguments are passed correctly.
func TestClient_Run_Args(t *testing.T) {
	skipIfNoTestWasm(t, echoArgsWasm)

	client, err := NewClient(ClientConfig{})
	require.NoError(t, err)
	defer func() { _ = client.Close(context.Background()) }()

	result, err := client.Run(context.Background(), wasmbackend.WasmSpec{
		Module: echoArgsWasm,
		Args:   []string{"arg1", "arg2", "arg3"},
		Security: wasmbackend.WasmSecuritySpec{
			EnableWASI: true,
		},
	})

	require.NoError(t, err)
	assert.Contains(t, result.Stdout, "arg1")
	assert.Contains(t, result.Stdout, "arg2")
	assert.Contains(t, result.Stdout, "arg3")
}

// TestClient_Run_Env tests that environment variables are accessible.
func TestClient_Run_Env(t *testing.T) {
	skipIfNoTestWasm(t, echoEnvWasm)

	client, err := NewClient(ClientConfig{})
	require.NoError(t, err)
	defer func() { _ = client.Close(context.Background()) }()

	result, err := client.Run(context.Background(), wasmbackend.WasmSpec{
		Module: echoEnvWasm,
		Env:    []string{"FOO=bar", "BAZ=qux"},
		Security: wasmbackend.WasmSecuritySpec{
			EnableWASI: true,
		},
	})

	require.NoError(t, err)
	assert.Contains(t, result.Stdout, "FOO=bar")
	assert.Contains(t, result.Stdout, "BAZ=qux")
}

// TestClient_Run_Stdin tests that stdin data is available to the module.
func TestClient_Run_Stdin(t *testing.T) {
	skipIfNoTestWasm(t, echoStdinWasm)

	client, err := NewClient(ClientConfig{})
	require.NoError(t, err)
	defer func() { _ = client.Close(context.Background()) }()

	result, err := client.Run(context.Background(), wasmbackend.WasmSpec{
		Module: echoStdinWasm,
		Stdin:  []byte("input data"),
		Security: wasmbackend.WasmSecuritySpec{
			EnableWASI: true,
		},
	})

	require.NoError(t, err)
	assert.Contains(t, result.Stdout, "input data")
}

// TestClient_Run_MemoryLimit tests that memory limits are enforced.
func TestClient_Run_MemoryLimit(t *testing.T) {
	skipIfNoTestWasm(t, allocMemoryWasm)

	client, err := NewClient(ClientConfig{
		MaxMemoryPages: 16, // 1MB limit
	})
	require.NoError(t, err)
	defer func() { _ = client.Close(context.Background()) }()

	// The alloc_memory module tries to allocate more than 1MB
	_, err = client.Run(context.Background(), wasmbackend.WasmSpec{
		Module: allocMemoryWasm,
		Resources: wasmbackend.WasmResourceSpec{
			MemoryPages: 16, // 1MB
		},
		Security: wasmbackend.WasmSecuritySpec{
			EnableWASI: true,
		},
	})

	// Should fail due to memory limit
	require.Error(t, err)
}

// TestClient_Close tests that Close releases resources.
func TestClient_Close(t *testing.T) {
	client, err := NewClient(ClientConfig{})
	require.NoError(t, err)

	err = client.Close(context.Background())
	require.NoError(t, err)

	// After close, running should fail
	_, err = client.Run(context.Background(), wasmbackend.WasmSpec{
		Module: []byte{0x00, 0x61, 0x73, 0x6d}, // minimal invalid wasm
	})
	require.Error(t, err)
}

// TestNewClient_Defaults tests that NewClient applies sensible defaults.
func TestNewClient_Defaults(t *testing.T) {
	client, err := NewClient(ClientConfig{})
	require.NoError(t, err)
	defer func() { _ = client.Close(context.Background()) }()

	// Should have created a runtime
	assert.NotNil(t, client.Runtime())
}

// TestNewClient_CustomMemoryPages tests custom memory page configuration.
func TestNewClient_CustomMemoryPages(t *testing.T) {
	client, err := NewClient(ClientConfig{
		MaxMemoryPages: 512, // 32MB
	})
	require.NoError(t, err)
	defer func() { _ = client.Close(context.Background()) }()

	assert.NotNil(t, client)
}

// TestClient_Run_NoWASI tests execution without WASI.
func TestClient_Run_NoWASI(t *testing.T) {
	// A module that doesn't need WASI should still work
	// This requires a special test module that just computes and returns
	t.Skip("requires non-WASI test module")
}

// Ensure interface compliance at compile time.
var _ wasmbackend.WasmRunner = (*Client)(nil)
