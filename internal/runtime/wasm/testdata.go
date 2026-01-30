package wasm

import (
	_ "embed"
)

// Test WASM binaries embedded at compile time.
// These are minimal WASI programs compiled from WAT source.

//go:embed testdata/hello.wasm
var helloWasmBytes []byte

//go:embed testdata/exit42.wasm
var exit42WasmBytes []byte

//go:embed testdata/stderr.wasm
var stderrWasmBytes []byte

//go:embed testdata/infinite_loop.wasm
var infiniteLoopWasmBytes []byte
