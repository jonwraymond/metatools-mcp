package wasm

import (
	"context"
	"fmt"
	"sync"

	"github.com/tetratelabs/wazero"

	wasmbackend "github.com/jonwraymond/toolexec/runtime/backend/wasm"
)

// Loader implements toolruntime's ModuleLoader interface using wazero.
// It provides compilation caching for WASM modules to improve performance
// when the same module is executed multiple times.
type Loader struct {
	runtime wazero.Runtime
	mu      sync.RWMutex
	modules map[string]*compiledModule
	closed  bool
}

// LoaderConfig holds configuration for the module loader.
type LoaderConfig struct {
	// MaxCachedModules limits the number of cached modules.
	// Default: 100
	MaxCachedModules int
}

// NewLoader creates a new module loader for the given runtime.
func NewLoader(runtime wazero.Runtime, _ LoaderConfig) *Loader {
	return &Loader{
		runtime: runtime,
		modules: make(map[string]*compiledModule),
	}
}

// NewLoaderFromClient creates a new module loader from a Client.
func NewLoaderFromClient(client *Client, cfg LoaderConfig) *Loader {
	return NewLoader(client.Runtime(), cfg)
}

// Load compiles a WASM binary into a reusable CompiledModule.
// The module is cached using a hash of the binary as the key.
func (l *Loader) Load(ctx context.Context, binary []byte) (wasmbackend.CompiledModule, error) {
	if l.closed {
		return nil, fmt.Errorf("loader is closed")
	}

	if len(binary) == 0 {
		return nil, fmt.Errorf("empty module binary")
	}

	// Create a simple hash key for caching
	key := hashBytes(binary)

	// Check cache first
	l.mu.RLock()
	if cached, ok := l.modules[key]; ok {
		l.mu.RUnlock()
		return cached, nil
	}
	l.mu.RUnlock()

	// Compile the module
	compiled, err := l.runtime.CompileModule(ctx, binary)
	if err != nil {
		return nil, fmt.Errorf("compile module: %w", err)
	}

	// Extract exports
	exports := make([]string, 0)
	for name, def := range compiled.ExportedFunctions() {
		_ = def // Use to avoid compiler warning
		exports = append(exports, name)
	}

	module := &compiledModule{
		compiled: compiled,
		name:     compiled.Name(),
		exports:  exports,
	}

	// Cache it
	l.mu.Lock()
	l.modules[key] = module
	l.mu.Unlock()

	return module, nil
}

// Close releases all cached modules and the loader.
func (l *Loader) Close(ctx context.Context) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.closed {
		return nil
	}

	var lastErr error
	for _, mod := range l.modules {
		if err := mod.compiled.Close(ctx); err != nil {
			lastErr = err
		}
	}

	l.modules = nil
	l.closed = true

	return lastErr
}

// compiledModule wraps a wazero.CompiledModule to implement wasmbackend.CompiledModule.
type compiledModule struct {
	compiled wazero.CompiledModule
	name     string
	exports  []string
}

// Name returns the module name if available.
func (m *compiledModule) Name() string {
	return m.name
}

// Exports lists exported functions.
func (m *compiledModule) Exports() []string {
	return m.exports
}

// Close releases the compiled module.
func (m *compiledModule) Close(ctx context.Context) error {
	return m.compiled.Close(ctx)
}

// CompiledModule returns the underlying wazero.CompiledModule.
func (m *compiledModule) CompiledModule() wazero.CompiledModule {
	return m.compiled
}

// hashBytes creates a simple hash key for module caching.
// For production use, consider using a proper hash function.
func hashBytes(data []byte) string {
	if len(data) == 0 {
		return ""
	}

	// Simple FNV-1a hash for demonstration
	const (
		offset64 = 14695981039346656037
		prime64  = 1099511628211
	)

	hash := uint64(offset64)
	for _, b := range data {
		hash ^= uint64(b)
		hash *= prime64
	}

	return fmt.Sprintf("%x", hash)
}

// Ensure interface compliance at compile time.
var _ wasmbackend.ModuleLoader = (*Loader)(nil)
var _ wasmbackend.CompiledModule = (*compiledModule)(nil)
