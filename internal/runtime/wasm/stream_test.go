package wasm

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	wasmbackend "github.com/jonwraymond/toolruntime/backend/wasm"
)

func TestStreamClient_RunStream_Success(t *testing.T) {
	client, err := NewStreamClient(ClientConfig{})
	require.NoError(t, err)
	defer func() { _ = client.Close(context.Background()) }()

	events, err := client.RunStream(context.Background(), wasmbackend.Spec{
		Module: helloWasm,
		Security: wasmbackend.SecuritySpec{
			EnableWASI: true,
		},
	})
	require.NoError(t, err)

	var stdout strings.Builder
	var exitCode int
	var gotExit bool

	for event := range events {
		switch event.Type {
		case wasmbackend.StreamEventStdout:
			stdout.Write(event.Data)
		case wasmbackend.StreamEventExit:
			exitCode = event.ExitCode
			gotExit = true
		case wasmbackend.StreamEventError:
			t.Fatalf("unexpected error: %v", event.Error)
		}
	}

	assert.True(t, gotExit, "should receive exit event")
	assert.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "hello")
}

func TestStreamClient_RunStream_Stderr(t *testing.T) {
	client, err := NewStreamClient(ClientConfig{})
	require.NoError(t, err)
	defer func() { _ = client.Close(context.Background()) }()

	events, err := client.RunStream(context.Background(), wasmbackend.Spec{
		Module: stderrWasm,
		Security: wasmbackend.SecuritySpec{
			EnableWASI: true,
		},
	})
	require.NoError(t, err)

	var stderr strings.Builder
	var gotStderr bool

	for event := range events {
		switch event.Type {
		case wasmbackend.StreamEventStderr:
			stderr.Write(event.Data)
			gotStderr = true
		case wasmbackend.StreamEventExit:
			// Expected
		case wasmbackend.StreamEventError:
			t.Fatalf("unexpected error: %v", event.Error)
		}
	}

	assert.True(t, gotStderr, "should receive stderr events")
	assert.Contains(t, stderr.String(), "error")
}

func TestStreamClient_RunStream_ExitEvent(t *testing.T) {
	client, err := NewStreamClient(ClientConfig{})
	require.NoError(t, err)
	defer func() { _ = client.Close(context.Background()) }()

	events, err := client.RunStream(context.Background(), wasmbackend.Spec{
		Module: exit42WasmBytes,
		Security: wasmbackend.SecuritySpec{
			EnableWASI: true,
		},
	})
	require.NoError(t, err)

	var exitEvent *wasmbackend.StreamEvent
	for event := range events {
		if event.Type == wasmbackend.StreamEventExit {
			e := event // Copy to avoid pointer issues
			exitEvent = &e
		}
	}

	require.NotNil(t, exitEvent, "should receive exit event")
	assert.Equal(t, 42, exitEvent.ExitCode)
}

func TestStreamClient_RunStream_ContextCancel(t *testing.T) {
	client, err := NewStreamClient(ClientConfig{})
	require.NoError(t, err)
	defer func() { _ = client.Close(context.Background()) }()

	ctx, cancel := context.WithCancel(context.Background())

	events, err := client.RunStream(ctx, wasmbackend.Spec{
		Module: infiniteLoopWasm,
		Security: wasmbackend.SecuritySpec{
			EnableWASI: true,
		},
	})
	require.NoError(t, err)

	// Cancel after a short delay
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	var gotError bool
	for event := range events {
		if event.Type == wasmbackend.StreamEventError {
			gotError = true
			assert.ErrorIs(t, event.Error, context.Canceled)
		}
	}

	assert.True(t, gotError, "should receive error event on cancellation")
}

func TestStreamClient_RunStream_InvalidModule(t *testing.T) {
	client, err := NewStreamClient(ClientConfig{})
	require.NoError(t, err)
	defer func() { _ = client.Close(context.Background()) }()

	events, err := client.RunStream(context.Background(), wasmbackend.Spec{
		Module: []byte("invalid"),
		Security: wasmbackend.SecuritySpec{
			EnableWASI: true,
		},
	})
	require.NoError(t, err) // RunStream itself doesn't fail, errors come through channel

	var gotError bool
	for event := range events {
		if event.Type == wasmbackend.StreamEventError {
			gotError = true
		}
	}

	assert.True(t, gotError, "should receive error event for invalid module")
}

func TestStreamClient_RunStream_EmptyModule(t *testing.T) {
	client, err := NewStreamClient(ClientConfig{})
	require.NoError(t, err)
	defer func() { _ = client.Close(context.Background()) }()

	events, err := client.RunStream(context.Background(), wasmbackend.Spec{
		Module: nil,
		Security: wasmbackend.SecuritySpec{
			EnableWASI: true,
		},
	})
	require.NoError(t, err)

	var gotError bool
	for event := range events {
		if event.Type == wasmbackend.StreamEventError {
			gotError = true
			assert.ErrorIs(t, event.Error, wasmbackend.ErrInvalidModule)
		}
	}

	assert.True(t, gotError, "should receive error event for empty module")
}

func TestStreamClient_RunStream_ClosedClient(t *testing.T) {
	client, err := NewStreamClient(ClientConfig{})
	require.NoError(t, err)

	err = client.Close(context.Background())
	require.NoError(t, err)

	_, err = client.RunStream(context.Background(), wasmbackend.Spec{
		Module: helloWasm,
	})
	require.Error(t, err)
}

// Ensure interface compliance at compile time.
var _ wasmbackend.StreamRunner = (*StreamClient)(nil)
