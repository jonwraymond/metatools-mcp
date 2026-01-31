# PRD-140: Migrate toolrun

**Phase:** 4 - Execution Layer
**Priority:** Critical
**Effort:** 4 hours
**Dependencies:** PRD-120

---

## Objective

Migrate the existing `toolrun` repository into `toolexec/run/` as the first package in the consolidated execution layer.

---

## Source Analysis

**Current Location:** `github.com/ApertureStack/toolrun`
**Target Location:** `github.com/ApertureStack/toolexec/run`

**Package Contents:**
- Tool execution pipeline (6-step execution)
- Backend dispatch (local, docker, wasm, remote)
- Chain execution for tool sequences
- Streaming support
- ~5,000 lines of code

---

## Deliverables

| Deliverable | Location | Description |
|-------------|----------|-------------|
| Run Package | `toolexec/run/` | Execution pipeline |
| Tests | `toolexec/run/*_test.go` | All existing tests |
| Documentation | `toolexec/run/doc.go` | Package documentation |

---

## Tasks

### Task 1: Prepare Target Repository

```bash
cd /tmp/migration
git clone git@github.com:ApertureStack/toolexec.git
cd toolexec

mkdir -p run
```

### Task 2: Clone and Analyze Source

```bash
cd /tmp/migration
git clone git@github.com:ApertureStack/toolrun.git
cd toolrun

ls -la
wc -l *.go
go test ./...
```

### Task 3: Copy Source Files

```bash
cd /tmp/migration

cp toolrun/*.go toolexec/run/
ls -la toolexec/run/
```

### Task 4: Update Import Paths

```bash
cd /tmp/migration/toolexec/run

# Update self-reference
OLD_IMPORT="github.com/ApertureStack/toolrun"
NEW_IMPORT="github.com/ApertureStack/toolexec/run"

for file in *.go; do
  sed -i '' "s|$OLD_IMPORT|$NEW_IMPORT|g" "$file"
done

# Update toolmodel to toolfoundation/model
OLD_MODEL="github.com/ApertureStack/toolmodel"
NEW_MODEL="github.com/ApertureStack/toolfoundation/model"

for file in *.go; do
  sed -i '' "s|$OLD_MODEL|$NEW_MODEL|g" "$file"
done

# Update toolruntime if referenced
OLD_RUNTIME="github.com/ApertureStack/toolruntime"
NEW_RUNTIME="github.com/ApertureStack/toolexec/runtime"

for file in *.go; do
  sed -i '' "s|$OLD_RUNTIME|$NEW_RUNTIME|g" "$file"
done

# Verify
grep -r "ApertureStack/toolrun\|ApertureStack/toolmodel\|ApertureStack/toolruntime" . || echo "✓ All imports updated"
```

### Task 5: Update Package Documentation

**File:** `toolexec/run/doc.go`

```go
// Package run provides the tool execution pipeline for the ApertureStack ecosystem.
//
// This package implements a 6-step execution pipeline that handles tool invocation
// from request to response, with support for multiple execution backends and
// chain execution.
//
// # Execution Pipeline
//
// The pipeline consists of six steps:
//
//  1. Validate: Check input against tool schema
//  2. Authorize: Verify execution permissions
//  3. Prepare: Set up execution context
//  4. Execute: Run tool via selected backend
//  5. Transform: Post-process output
//  6. Respond: Format and return result
//
// # Usage
//
// Create a runner and execute tools:
//
//	runner := run.NewRunner(run.Config{
//	    DefaultBackend: "local",
//	    Timeout:        30 * time.Second,
//	})
//
//	result, err := runner.Run(ctx, run.Request{
//	    ToolID: "calculator",
//	    Input:  map[string]any{"operation": "add", "a": 1, "b": 2},
//	})
//
// # Backends
//
// The runner supports multiple execution backends:
//
//   - local: Direct in-process execution
//   - docker: Container-based isolation
//   - wasm: WebAssembly sandbox
//   - remote: HTTP/gRPC remote execution
//
// # Chain Execution
//
// Execute a sequence of tools where output flows to input:
//
//	chain := run.Chain{
//	    Steps: []run.ChainStep{
//	        {ToolID: "fetch-data", Input: map[string]any{"url": "..."}},
//	        {ToolID: "transform", InputMapping: map[string]string{"data": "$.output"}},
//	        {ToolID: "store", InputMapping: map[string]string{"content": "$.output"}},
//	    },
//	}
//	results, err := runner.RunChain(ctx, chain)
//
// # Streaming
//
// For long-running tools, use streaming execution:
//
//	stream, err := runner.RunStream(ctx, request)
//	for event := range stream.Events() {
//	    fmt.Printf("Progress: %s\n", event.Message)
//	}
//	result := stream.Result()
//
// # Migration Note
//
// This package was migrated from github.com/ApertureStack/toolrun as part of
// the ApertureStack consolidation.
package run
```

### Task 6: Build and Test

```bash
cd /tmp/migration/toolexec

go mod tidy
go build ./...
go test -v -coverprofile=run_coverage.out ./run/...

go tool cover -func=run_coverage.out | grep total
```

### Task 7: Commit and Push

```bash
cd /tmp/migration/toolexec

git add -A
git commit -m "feat(run): migrate toolrun package

Migrate the execution pipeline from standalone toolrun repository.

Package contents:
- 6-step execution pipeline
- Multi-backend dispatch (local, docker, wasm, remote)
- Chain execution for tool sequences
- Streaming execution support
- Configurable timeouts and retries

Features:
- Input validation against tool schema
- Authorization hooks
- Context preparation
- Backend selection
- Output transformation
- Response formatting

Dependencies:
- github.com/ApertureStack/toolfoundation/model

This is part of the ApertureStack consolidation effort.

Migration: github.com/ApertureStack/toolrun → toolexec/run

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"

git push origin main
```

---

## Key Interfaces

```go
package run

import (
    "context"
    "github.com/ApertureStack/toolfoundation/model"
)

// Runner executes tools.
type Runner interface {
    // Run executes a single tool.
    Run(ctx context.Context, req Request) (*Result, error)

    // RunChain executes a sequence of tools.
    RunChain(ctx context.Context, chain Chain) ([]Result, error)

    // RunStream executes a tool with streaming output.
    RunStream(ctx context.Context, req Request) (Stream, error)
}

// Request represents an execution request.
type Request struct {
    ToolID   string
    Tool     *model.Tool // Optional: provide tool directly
    Input    map[string]any
    Options  Options
    Metadata map[string]any
}

// Result represents an execution result.
type Result struct {
    ToolID   string
    Output   any
    Error    *Error
    Metrics  Metrics
    Metadata map[string]any
}

// Chain represents a sequence of tool executions.
type Chain struct {
    ID    string
    Steps []ChainStep
}

// ChainStep represents a step in a chain.
type ChainStep struct {
    ToolID       string
    Input        map[string]any
    InputMapping map[string]string // JSONPath mappings
}

// Backend executes tools in a specific environment.
type Backend interface {
    Execute(ctx context.Context, tool model.Tool, input map[string]any) (any, error)
    Name() string
}
```

---

## Verification Checklist

- [ ] All source files copied
- [ ] Import paths updated
- [ ] `go build ./...` succeeds
- [ ] `go test ./...` passes
- [ ] 6-step pipeline works
- [ ] Chain execution works
- [ ] Streaming works
- [ ] Package documentation updated

---

## Acceptance Criteria

1. `toolexec/run` package builds successfully
2. All tests pass
3. Single tool execution works
4. Chain execution produces correct results
5. Multiple backends supported

---

## Rollback Plan

```bash
cd /tmp/migration/toolexec
rm -rf run/
git checkout HEAD~1 -- .
git push origin main --force-with-lease
```

---

## Next Steps

- PRD-141: Migrate toolruntime
- PRD-142: Migrate toolcode
