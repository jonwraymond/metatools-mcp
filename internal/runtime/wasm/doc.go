// Package wasm provides wazero SDK implementations of toolruntime's WASM interfaces.
//
// This package implements:
//   - WasmRunner: executes code in WASM sandboxes with full lifecycle management
//   - ModuleLoader: compiles and caches WASM modules for reuse
//   - HealthChecker: verifies wazero runtime availability
//   - StreamRunner: provides real-time stdout/stderr streaming
//
// # Architecture
//
// The implementations in this package satisfy interfaces defined in toolruntime/backend/wasm.
// This separation allows toolruntime to remain agnostic to the specific WASM runtime
// while metatools-mcp provides the concrete wazero SDK integration.
//
// # Usage
//
// Create a client and wire it into toolruntime's WASM backend:
//
//	client, err := wasm.NewClient(wasm.ClientConfig{})
//	if err != nil {
//	    return err
//	}
//	defer client.Close(context.Background())
//
//	backend := wasmbackend.New(wasmbackend.Config{
//	    Runtime:       "wazero",
//	    EnableWASI:    true,
//	    Client:        client,
//	    HealthChecker: wasm.NewHealthCheck(client),
//	})
//
// # Testing
//
// Tests use embedded WASM binaries for deterministic testing.
// Use -short to skip integration tests:
//
//	go test ./internal/runtime/wasm/... -short
//
// # Security
//
// The client enforces security settings from WasmSpec:
//   - Memory limits (via MemoryPages)
//   - WASI sandboxing (filesystem, environment isolation)
//   - No network access by default
//   - Fuel-based CPU limiting (optional)
//
// Module instances are always closed after execution, even on errors.
package wasm
