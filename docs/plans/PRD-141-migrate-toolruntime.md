# PRD-141: Migrate toolruntime

**Phase:** 4 - Execution Layer
**Priority:** Critical
**Effort:** 4 hours
**Dependencies:** PRD-120
**Status:** Done (2026-01-31)

---

## Objective

Migrate the existing `toolruntime` repository into `toolexec/runtime/` as the second package in the consolidated execution layer.

---

## Source Analysis

**Current Location:** `github.com/jonwraymond/toolruntime`
**Target Location:** `github.com/jonwraymond/toolexec/runtime`

**Package Contents:**
- Runtime abstraction for tool execution
- 10 sandbox backends (unsafe, docker, containerd, kubernetes, firecracker, kata, gvisor, wasm, temporal, remote)
- 3 security profiles (dev, standard, hardened)
- Error handling with structured errors
- ~8,000 lines of code

---

## Deliverables

| Deliverable | Location | Description |
|-------------|----------|-------------|
| Runtime Package | `toolexec/runtime/` | Runtime abstraction |
| Backend Implementations | `toolexec/runtime/backend/` | Sandbox backends |
| Tests | `toolexec/runtime/*_test.go` | All existing tests |
| Documentation | `toolexec/runtime/doc.go` | Package documentation |

---

## Tasks

### Task 1: Clone and Analyze Source

```bash
cd /tmp/migration
git clone git@github.com:jonwraymond/toolruntime.git
cd toolruntime

ls -la
find . -name "*.go" | wc -l
go test ./...
```

### Task 2: Copy Source Files (Preserving Structure)

```bash
cd /tmp/migration/toolexec

# Create directories
mkdir -p runtime/backend

# Copy root package files
cp ../toolruntime/*.go runtime/

# Copy backend subdirectory
cp -r ../toolruntime/backend/* runtime/backend/

ls -la runtime/
ls -la runtime/backend/
```

### Task 3: Update Import Paths

```bash
cd /tmp/migration/toolexec

# Update all Go files recursively
find runtime -name "*.go" -exec sed -i '' 's|github.com/jonwraymond/toolruntime|github.com/jonwraymond/toolexec/runtime|g' {} \;

# Update toolmodel references
find runtime -name "*.go" -exec sed -i '' 's|github.com/jonwraymond/toolmodel|github.com/jonwraymond/toolfoundation/model|g' {} \;

# Verify
grep -r "jonwraymond/toolruntime\|jonwraymond/toolmodel" runtime/ || echo "✓ All imports updated"
```

### Task 4: Update Package Documentation

**File:** `toolexec/runtime/doc.go`

```go
// Package runtime provides sandboxed execution environments for tools.
//
// This package implements a runtime abstraction that enables tool execution
// in various isolated environments, from no isolation (development) to
// strict container-based isolation (production).
//
// # Backends
//
// The package supports 10 execution backends:
//
//   - unsafe: No isolation (development only)
//   - docker: Docker container isolation
//   - containerd: containerd-based containers
//   - kubernetes: Kubernetes pod execution
//   - firecracker: MicroVM isolation
//   - kata: Kata Containers
//   - gvisor: gVisor sandbox
//   - wasm: WebAssembly sandbox
//   - temporal: Temporal workflow execution
//   - remote: Remote execution via HTTP/gRPC
//
// # Security Profiles
//
// Three security profiles control isolation level:
//
//   - SecurityNone: No isolation (unsafe backend)
//   - SecurityBasic: Container isolation with defaults
//   - SecurityStrict: Hardened isolation with restricted capabilities
//
// # Usage
//
// Create a runtime with a specific backend:
//
//	rt, err := runtime.New(runtime.Config{
//	    Backend:  "docker",
//	    Security: runtime.SecurityBasic,
//	    Timeout:  30 * time.Second,
//	})
//
//	result, err := rt.Execute(ctx, runtime.Task{
//	    Tool:  tool,
//	    Input: input,
//	})
//
// # Resource Limits
//
// Configure resource limits for execution:
//
//	rt, _ := runtime.New(runtime.Config{
//	    Backend: "docker",
//	    Resources: runtime.Resources{
//	        Memory:    "256Mi",
//	        CPU:       "0.5",
//	        Timeout:   "30s",
//	        DiskSpace: "100Mi",
//	    },
//	})
//
// # Migration Note
//
// This package was migrated from github.com/jonwraymond/toolruntime as part of
// the ApertureStack consolidation.
package runtime
```

### Task 5: Backend Build Tags (Optional)

Backends are configured at runtime; no build tags are required. If tags are introduced
later, document them alongside the backend matrix.

### Task 6: Update go.mod Dependencies

```bash
cd /tmp/migration/toolexec

# The runtime package may have significant dependencies
cat go.mod

# Add required dependencies
go get github.com/docker/docker/client
go get k8s.io/client-go@latest

go mod tidy
```

### Task 7: Build and Test

```bash
cd /tmp/migration/toolexec

# Build all (without optional backends)
go build ./...

# Build with docker backend
go build -tags=docker ./...

# Test
go test -v -coverprofile=runtime_coverage.out ./runtime/...
go tool cover -func=runtime_coverage.out | grep total
```

### Task 8: Commit and Push

```bash
cd /tmp/migration/toolexec

git add -A
git commit -m "feat(runtime): migrate toolruntime package

Migrate the sandbox runtime from standalone toolruntime repository.

Package contents:
- Runtime interface for sandboxed execution
- 10 backend implementations
- 3 security profiles (none, basic, strict)
- Resource limit configuration
- Structured error handling

Backends (via build tags):
- unsafe: always included
- docker: -tags=docker
- containerd: -tags=containerd
- kubernetes: -tags=kubernetes
- firecracker: -tags=firecracker
- kata: -tags=kata
- gvisor: -tags=gvisor
- wasm: -tags=wasm
- temporal: -tags=temporal
- remote: always included

Dependencies:
- github.com/jonwraymond/toolfoundation/model
- Various backend-specific dependencies

This is part of the ApertureStack consolidation effort.

Migration: github.com/jonwraymond/toolruntime → toolexec/runtime

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"

git push origin main
```

---

## Key Interfaces

```go
package runtime

import (
    "context"
    "github.com/jonwraymond/toolfoundation/model"
)

// Runtime executes tools in isolated environments.
type Runtime interface {
    // Execute runs a tool in the sandbox.
    Execute(ctx context.Context, task Task) (*Result, error)

    // Close releases runtime resources.
    Close() error

    // Backend returns the backend name.
    Backend() string

    // Security returns the security profile.
    Security() SecurityProfile
}

// Task represents an execution task.
type Task struct {
    Tool      model.Tool
    Input     map[string]any
    Env       map[string]string
    Resources *Resources
}

// Result represents an execution result.
type Result struct {
    Output   any
    ExitCode int
    Stdout   string
    Stderr   string
    Duration time.Duration
}

// Config configures the runtime.
type Config struct {
    Backend   string
    Security  SecurityProfile
    Timeout   time.Duration
    Resources *Resources
}

// SecurityProfile defines isolation level.
type SecurityProfile int

const (
    SecurityNone SecurityProfile = iota
    SecurityBasic
    SecurityStrict
)

// Resources defines execution resource limits.
type Resources struct {
    Memory    string // e.g., "256Mi"
    CPU       string // e.g., "0.5"
    Timeout   string // e.g., "30s"
    DiskSpace string // e.g., "100Mi"
}
```

---

## Verification Checklist

- [ ] All source files copied (including backend/)
- [ ] Import paths updated
- [ ] `go build ./...` succeeds
- [ ] `go test ./...` passes
- [ ] Security profiles work
- [ ] Package documentation updated

---

## Acceptance Criteria

1. `toolexec/runtime` package builds successfully
2. All tests pass
3. Unsafe backend works without extra dependencies
4. Docker backend works when configured with a container runner
5. Security profiles enforce appropriate isolation

## Completion Notes

- Migration completed into `toolexec/runtime` with `runtime/backend/*`.
- Security profiles are `ProfileDev`, `ProfileStandard`, `ProfileHardened`.
- No build tags are required; backend selection is runtime-configured.

---

## Rollback Plan

```bash
cd /tmp/migration/toolexec
rm -rf runtime/
git checkout HEAD~1 -- .
git push origin main --force-with-lease
```

---

## Next Steps

- PRD-142: Migrate toolcode
- PRD-143: Extract toolbackend
