package wasm

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	wasmbackend "github.com/jonwraymond/toolruntime/backend/wasm"
)

func TestHealthCheck_Ping_Success(t *testing.T) {
	client, err := NewClient(ClientConfig{})
	require.NoError(t, err)
	defer func() { _ = client.Close(context.Background()) }()

	health := NewHealthCheck(client)

	err = health.Ping(context.Background())
	require.NoError(t, err)
}

func TestHealthCheck_Ping_NilClient(t *testing.T) {
	health := NewHealthCheck(nil)

	err := health.Ping(context.Background())
	require.Error(t, err)
	assert.ErrorIs(t, err, wasmbackend.ErrWASMRuntimeNotAvailable)
}

func TestHealthCheck_Ping_ClosedClient(t *testing.T) {
	client, err := NewClient(ClientConfig{})
	require.NoError(t, err)

	// Close the client
	err = client.Close(context.Background())
	require.NoError(t, err)

	health := NewHealthCheck(client)

	err = health.Ping(context.Background())
	require.Error(t, err)
	assert.ErrorIs(t, err, wasmbackend.ErrWASMRuntimeNotAvailable)
}

func TestHealthCheck_Ping_CancelledContext(t *testing.T) {
	client, err := NewClient(ClientConfig{})
	require.NoError(t, err)
	defer func() { _ = client.Close(context.Background()) }()

	health := NewHealthCheck(client)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = health.Ping(ctx)
	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestHealthCheck_Info_Success(t *testing.T) {
	client, err := NewClient(ClientConfig{})
	require.NoError(t, err)
	defer func() { _ = client.Close(context.Background()) }()

	health := NewHealthCheck(client)

	info, err := health.Info(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "wazero", info.Name)
	assert.NotEmpty(t, info.Version)
	assert.Contains(t, info.Features, "wasi_snapshot_preview1")
}

func TestHealthCheck_Info_NilClient(t *testing.T) {
	health := NewHealthCheck(nil)

	_, err := health.Info(context.Background())
	require.Error(t, err)
	assert.ErrorIs(t, err, wasmbackend.ErrWASMRuntimeNotAvailable)
}

func TestRuntimeHealthCheck_Ping_Success(t *testing.T) {
	client, err := NewClient(ClientConfig{})
	require.NoError(t, err)
	defer func() { _ = client.Close(context.Background()) }()

	health := NewHealthCheckFromRuntime(client.Runtime())

	err = health.Ping(context.Background())
	require.NoError(t, err)
}

func TestRuntimeHealthCheck_Ping_NilRuntime(t *testing.T) {
	health := NewHealthCheckFromRuntime(nil)

	err := health.Ping(context.Background())
	require.Error(t, err)
	assert.ErrorIs(t, err, wasmbackend.ErrWASMRuntimeNotAvailable)
}

func TestRuntimeHealthCheck_Info_Success(t *testing.T) {
	client, err := NewClient(ClientConfig{})
	require.NoError(t, err)
	defer func() { _ = client.Close(context.Background()) }()

	health := NewHealthCheckFromRuntime(client.Runtime())

	info, err := health.Info(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "wazero", info.Name)
}

// Ensure interface compliance at compile time.
var _ wasmbackend.HealthChecker = (*HealthCheck)(nil)
var _ wasmbackend.HealthChecker = (*RuntimeHealthCheck)(nil)
