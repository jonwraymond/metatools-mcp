package wasm

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	wasmbackend "github.com/jonwraymond/toolexec/runtime/backend/wasm"
)

func TestLoader_Load_Success(t *testing.T) {
	client, err := NewClient(ClientConfig{})
	require.NoError(t, err)
	defer func() { _ = client.Close(context.Background()) }()

	loader := NewLoaderFromClient(client, LoaderConfig{})
	defer func() { _ = loader.Close(context.Background()) }()

	module, err := loader.Load(context.Background(), helloWasm)
	require.NoError(t, err)
	assert.NotNil(t, module)

	// Should have exports
	exports := module.Exports()
	assert.Contains(t, exports, "_start")
}

func TestLoader_Load_Invalid(t *testing.T) {
	client, err := NewClient(ClientConfig{})
	require.NoError(t, err)
	defer func() { _ = client.Close(context.Background()) }()

	loader := NewLoaderFromClient(client, LoaderConfig{})
	defer func() { _ = loader.Close(context.Background()) }()

	_, err = loader.Load(context.Background(), []byte("invalid wasm"))
	require.Error(t, err)
}

func TestLoader_Load_Empty(t *testing.T) {
	client, err := NewClient(ClientConfig{})
	require.NoError(t, err)
	defer func() { _ = client.Close(context.Background()) }()

	loader := NewLoaderFromClient(client, LoaderConfig{})
	defer func() { _ = loader.Close(context.Background()) }()

	_, err = loader.Load(context.Background(), nil)
	require.Error(t, err)
}

func TestLoader_Load_Caching(t *testing.T) {
	client, err := NewClient(ClientConfig{})
	require.NoError(t, err)
	defer func() { _ = client.Close(context.Background()) }()

	loader := NewLoaderFromClient(client, LoaderConfig{})
	defer func() { _ = loader.Close(context.Background()) }()

	// Load the same module twice
	module1, err := loader.Load(context.Background(), helloWasm)
	require.NoError(t, err)

	module2, err := loader.Load(context.Background(), helloWasm)
	require.NoError(t, err)

	// Should return the same cached module
	assert.Equal(t, module1, module2)
}

func TestLoader_Load_DifferentModules(t *testing.T) {
	client, err := NewClient(ClientConfig{})
	require.NoError(t, err)
	defer func() { _ = client.Close(context.Background()) }()

	loader := NewLoaderFromClient(client, LoaderConfig{})
	defer func() { _ = loader.Close(context.Background()) }()

	// Load different modules
	module1, err := loader.Load(context.Background(), helloWasm)
	require.NoError(t, err)

	module2, err := loader.Load(context.Background(), exit42WasmBytes)
	require.NoError(t, err)

	// Should be different
	assert.NotEqual(t, module1, module2)
}

func TestLoader_Close(t *testing.T) {
	client, err := NewClient(ClientConfig{})
	require.NoError(t, err)
	defer func() { _ = client.Close(context.Background()) }()

	loader := NewLoaderFromClient(client, LoaderConfig{})

	// Load a module
	_, err = loader.Load(context.Background(), helloWasm)
	require.NoError(t, err)

	// Close the loader
	err = loader.Close(context.Background())
	require.NoError(t, err)

	// After close, loading should fail
	_, err = loader.Load(context.Background(), helloWasm)
	require.Error(t, err)
}

func TestLoader_Close_Idempotent(t *testing.T) {
	client, err := NewClient(ClientConfig{})
	require.NoError(t, err)
	defer func() { _ = client.Close(context.Background()) }()

	loader := NewLoaderFromClient(client, LoaderConfig{})

	// Close multiple times should not error
	err = loader.Close(context.Background())
	require.NoError(t, err)

	err = loader.Close(context.Background())
	require.NoError(t, err)
}

func TestCompiledModule_Close(t *testing.T) {
	client, err := NewClient(ClientConfig{})
	require.NoError(t, err)
	defer func() { _ = client.Close(context.Background()) }()

	loader := NewLoaderFromClient(client, LoaderConfig{})
	defer func() { _ = loader.Close(context.Background()) }()

	module, err := loader.Load(context.Background(), helloWasm)
	require.NoError(t, err)

	// Close individual module
	err = module.Close(context.Background())
	require.NoError(t, err)
}

// Ensure interface compliance at compile time.
var _ wasmbackend.ModuleLoader = (*Loader)(nil)
var _ wasmbackend.CompiledModule = (*compiledModule)(nil)
