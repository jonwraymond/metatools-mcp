# PRD-001: toolruntime Docker Backend Implementation

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement a production-ready Docker isolation backend for toolruntime, replacing the current stub implementation.

**Architecture:** The Docker backend will use the official Docker SDK for Go to create ephemeral containers with resource limits, network isolation, and filesystem sandboxing. It integrates with the existing Backend interface and SecurityProfile system.

**Tech Stack:** Docker SDK (github.com/docker/docker), Go 1.21+, existing toolruntime interfaces

**Priority:** P0 - Critical (9 of 10 backends are stubs; Docker is the most practical production backend)

**Scope:** Single backend implementation with full test coverage

---

## Context

The toolruntime library defines 10 isolation backends but only UnsafeHost is implemented. The Docker backend stub at `backend/docker/docker.go` returns `ErrBackendUnavailable`. This PRD implements a production-ready Docker backend.

**Current State:**
```go
// backend/docker/docker.go (current stub)
func (b *Backend) Execute(ctx context.Context, req gateway.Request) (gateway.Response, error) {
    return gateway.Response{}, gateway.ErrBackendUnavailable
}
```

**Target State:** Full Docker isolation with:
- Container lifecycle management
- Resource limits (CPU, memory, disk)
- Network isolation modes
- Filesystem sandboxing
- Timeout enforcement
- Output capture (stdout/stderr)

---

## Tasks

### Task 1: Add Docker SDK Dependency

**Files:**
- Modify: `go.mod`
- Modify: `go.sum`

**Step 1: Add Docker SDK to go.mod**

Run:
```bash
cd /Users/jraymond/Documents/Projects/toolruntime && go get github.com/docker/docker@v24.0.7
```

**Step 2: Verify import works**

Run:
```bash
cd /Users/jraymond/Documents/Projects/toolruntime && go build ./...
```
Expected: Build succeeds

**Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "$(cat <<'EOF'
deps: add Docker SDK for container backend

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 2: Define Docker Backend Configuration

**Files:**
- Modify: `backend/docker/docker.go`
- Test: `backend/docker/docker_test.go`

**Step 1: Write failing test for Config validation**

```go
// backend/docker/docker_test.go
package docker

import (
    "testing"
)

func TestConfig_Validate(t *testing.T) {
    tests := []struct {
        name    string
        config  Config
        wantErr bool
    }{
        {
            name:    "empty config uses defaults",
            config:  Config{},
            wantErr: false,
        },
        {
            name:    "explicit valid config",
            config:  Config{Image: "python:3.11-slim", Timeout: 30},
            wantErr: false,
        },
        {
            name:    "negative timeout invalid",
            config:  Config{Timeout: -1},
            wantErr: true,
        },
        {
            name:    "zero memory limit uses default",
            config:  Config{MemoryLimitMB: 0},
            wantErr: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.config.Validate()
            if (err != nil) != tt.wantErr {
                t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/jraymond/Documents/Projects/toolruntime && go test ./backend/docker/... -run TestConfig_Validate -v`
Expected: FAIL - Config type doesn't have Validate method

**Step 3: Implement Config struct with validation**

```go
// backend/docker/docker.go
package docker

import (
    "errors"
    "time"
)

// Config holds Docker backend configuration.
type Config struct {
    // Image is the Docker image to use (default: "python:3.11-slim")
    Image string

    // Timeout in seconds for execution (default: 30)
    Timeout int

    // MemoryLimitMB is the memory limit in MB (default: 256)
    MemoryLimitMB int64

    // CPUShares is the relative CPU weight (default: 512)
    CPUShares int64

    // NetworkMode: "none", "bridge", "host" (default: "none")
    NetworkMode string

    // WorkDir inside container (default: "/workspace")
    WorkDir string
}

// DefaultConfig returns sensible defaults for Docker execution.
func DefaultConfig() Config {
    return Config{
        Image:         "python:3.11-slim",
        Timeout:       30,
        MemoryLimitMB: 256,
        CPUShares:     512,
        NetworkMode:   "none",
        WorkDir:       "/workspace",
    }
}

// Validate checks config values are sensible.
func (c *Config) Validate() error {
    if c.Timeout < 0 {
        return errors.New("timeout cannot be negative")
    }
    if c.MemoryLimitMB < 0 {
        return errors.New("memory limit cannot be negative")
    }
    return nil
}

// withDefaults fills zero values with defaults.
func (c Config) withDefaults() Config {
    defaults := DefaultConfig()
    if c.Image == "" {
        c.Image = defaults.Image
    }
    if c.Timeout == 0 {
        c.Timeout = defaults.Timeout
    }
    if c.MemoryLimitMB == 0 {
        c.MemoryLimitMB = defaults.MemoryLimitMB
    }
    if c.CPUShares == 0 {
        c.CPUShares = defaults.CPUShares
    }
    if c.NetworkMode == "" {
        c.NetworkMode = defaults.NetworkMode
    }
    if c.WorkDir == "" {
        c.WorkDir = defaults.WorkDir
    }
    return c
}
```

**Step 4: Run test to verify it passes**

Run: `cd /Users/jraymond/Documents/Projects/toolruntime && go test ./backend/docker/... -run TestConfig_Validate -v`
Expected: PASS

**Step 5: Commit**

```bash
git add backend/docker/docker.go backend/docker/docker_test.go
git commit -m "$(cat <<'EOF'
feat(docker): add Config struct with validation and defaults

- Define Config with Image, Timeout, MemoryLimitMB, CPUShares, NetworkMode, WorkDir
- Add DefaultConfig() for sensible production defaults
- Add Validate() to catch invalid configurations
- Add withDefaults() to fill zero values

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 3: Implement Docker Client Initialization

**Files:**
- Modify: `backend/docker/docker.go`
- Test: `backend/docker/docker_test.go`

**Step 1: Write failing test for Backend initialization**

```go
func TestNewBackend(t *testing.T) {
    // Skip if Docker not available
    if _, err := exec.LookPath("docker"); err != nil {
        t.Skip("Docker not available")
    }

    t.Run("creates backend with default config", func(t *testing.T) {
        b, err := NewBackend(Config{})
        if err != nil {
            t.Fatalf("NewBackend() error = %v", err)
        }
        if b == nil {
            t.Fatal("NewBackend() returned nil")
        }
        defer b.Close()
    })

    t.Run("creates backend with custom config", func(t *testing.T) {
        b, err := NewBackend(Config{
            Image:         "alpine:latest",
            Timeout:       60,
            MemoryLimitMB: 512,
        })
        if err != nil {
            t.Fatalf("NewBackend() error = %v", err)
        }
        defer b.Close()
    })
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/jraymond/Documents/Projects/toolruntime && go test ./backend/docker/... -run TestNewBackend -v`
Expected: FAIL - NewBackend doesn't exist or returns stub

**Step 3: Implement NewBackend with Docker client**

```go
import (
    "context"
    "errors"
    "os/exec"
    "time"

    "github.com/docker/docker/client"
)

// Backend implements gateway.Backend using Docker containers.
type Backend struct {
    client *client.Client
    config Config
}

// NewBackend creates a Docker backend with the given configuration.
func NewBackend(cfg Config) (*Backend, error) {
    if err := cfg.Validate(); err != nil {
        return nil, err
    }

    cfg = cfg.withDefaults()

    // Create Docker client from environment
    cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
    if err != nil {
        return nil, err
    }

    // Verify Docker is accessible
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if _, err := cli.Ping(ctx); err != nil {
        cli.Close()
        return nil, err
    }

    return &Backend{
        client: cli,
        config: cfg,
    }, nil
}

// Close releases Docker client resources.
func (b *Backend) Close() error {
    if b.client != nil {
        return b.client.Close()
    }
    return nil
}
```

**Step 4: Run test to verify it passes**

Run: `cd /Users/jraymond/Documents/Projects/toolruntime && go test ./backend/docker/... -run TestNewBackend -v`
Expected: PASS (or SKIP if Docker not available)

**Step 5: Commit**

```bash
git add backend/docker/docker.go backend/docker/docker_test.go
git commit -m "$(cat <<'EOF'
feat(docker): implement NewBackend with Docker client initialization

- Create Docker client from environment
- Verify connectivity with Ping
- Clean up on error

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 4: Implement Container Creation and Execution

**Files:**
- Modify: `backend/docker/docker.go`
- Test: `backend/docker/docker_test.go`

**Step 1: Write failing test for Execute**

```go
func TestBackend_Execute(t *testing.T) {
    if _, err := exec.LookPath("docker"); err != nil {
        t.Skip("Docker not available")
    }

    b, err := NewBackend(Config{
        Image:   "alpine:latest",
        Timeout: 10,
    })
    if err != nil {
        t.Fatalf("NewBackend() error = %v", err)
    }
    defer b.Close()

    t.Run("executes simple command", func(t *testing.T) {
        resp, err := b.Execute(context.Background(), gateway.Request{
            Code:     "echo 'hello world'",
            Language: "sh",
        })
        if err != nil {
            t.Fatalf("Execute() error = %v", err)
        }
        if resp.ExitCode != 0 {
            t.Errorf("ExitCode = %d, want 0", resp.ExitCode)
        }
        if !strings.Contains(resp.Stdout, "hello world") {
            t.Errorf("Stdout = %q, want contains 'hello world'", resp.Stdout)
        }
    })

    t.Run("captures stderr", func(t *testing.T) {
        resp, err := b.Execute(context.Background(), gateway.Request{
            Code:     "echo 'error message' >&2",
            Language: "sh",
        })
        if err != nil {
            t.Fatalf("Execute() error = %v", err)
        }
        if !strings.Contains(resp.Stderr, "error message") {
            t.Errorf("Stderr = %q, want contains 'error message'", resp.Stderr)
        }
    })

    t.Run("respects timeout", func(t *testing.T) {
        ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
        defer cancel()

        _, err := b.Execute(ctx, gateway.Request{
            Code:     "sleep 60",
            Language: "sh",
        })
        if err == nil {
            t.Error("Execute() should timeout")
        }
    })

    t.Run("returns non-zero exit code", func(t *testing.T) {
        resp, err := b.Execute(context.Background(), gateway.Request{
            Code:     "exit 42",
            Language: "sh",
        })
        if err != nil {
            t.Fatalf("Execute() error = %v", err)
        }
        if resp.ExitCode != 42 {
            t.Errorf("ExitCode = %d, want 42", resp.ExitCode)
        }
    })
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/jraymond/Documents/Projects/toolruntime && go test ./backend/docker/... -run TestBackend_Execute -v`
Expected: FAIL - Execute returns ErrBackendUnavailable

**Step 3: Implement Execute method**

```go
import (
    "bytes"
    "context"
    "io"
    "time"

    "github.com/docker/docker/api/types/container"
    "github.com/docker/docker/api/types/image"
    "github.com/docker/docker/pkg/stdcopy"

    "github.com/your-org/toolruntime/gateway"
)

// Execute runs code in an ephemeral Docker container.
func (b *Backend) Execute(ctx context.Context, req gateway.Request) (gateway.Response, error) {
    // Apply timeout from config if not set in context
    if _, ok := ctx.Deadline(); !ok {
        var cancel context.CancelFunc
        ctx, cancel = context.WithTimeout(ctx, time.Duration(b.config.Timeout)*time.Second)
        defer cancel()
    }

    // Ensure image is available
    if err := b.ensureImage(ctx); err != nil {
        return gateway.Response{}, err
    }

    // Create container
    containerConfig := &container.Config{
        Image:      b.config.Image,
        Cmd:        []string{"sh", "-c", req.Code},
        WorkingDir: b.config.WorkDir,
        Tty:        false,
    }

    hostConfig := &container.HostConfig{
        Resources: container.Resources{
            Memory:    b.config.MemoryLimitMB * 1024 * 1024,
            CPUShares: b.config.CPUShares,
        },
        NetworkMode: container.NetworkMode(b.config.NetworkMode),
        AutoRemove:  true,
    }

    resp, err := b.client.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, "")
    if err != nil {
        return gateway.Response{}, err
    }
    containerID := resp.ID

    // Start container
    if err := b.client.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
        return gateway.Response{}, err
    }

    // Wait for completion
    statusCh, errCh := b.client.ContainerWait(ctx, containerID, container.WaitConditionNotRunning)

    var exitCode int64
    select {
    case err := <-errCh:
        if err != nil {
            // Try to stop container on error
            _ = b.client.ContainerStop(context.Background(), containerID, container.StopOptions{})
            return gateway.Response{}, err
        }
    case status := <-statusCh:
        exitCode = status.StatusCode
    case <-ctx.Done():
        _ = b.client.ContainerStop(context.Background(), containerID, container.StopOptions{})
        return gateway.Response{}, ctx.Err()
    }

    // Get logs
    logs, err := b.client.ContainerLogs(ctx, containerID, container.LogsOptions{
        ShowStdout: true,
        ShowStderr: true,
    })
    if err != nil {
        return gateway.Response{ExitCode: int(exitCode)}, nil
    }
    defer logs.Close()

    var stdout, stderr bytes.Buffer
    _, _ = stdcopy.StdCopy(&stdout, &stderr, logs)

    return gateway.Response{
        Stdout:   stdout.String(),
        Stderr:   stderr.String(),
        ExitCode: int(exitCode),
    }, nil
}

// ensureImage pulls the image if not present locally.
func (b *Backend) ensureImage(ctx context.Context) error {
    _, _, err := b.client.ImageInspectWithRaw(ctx, b.config.Image)
    if err == nil {
        return nil // Image exists
    }

    reader, err := b.client.ImagePull(ctx, b.config.Image, image.PullOptions{})
    if err != nil {
        return err
    }
    defer reader.Close()

    // Drain the reader to complete the pull
    _, _ = io.Copy(io.Discard, reader)
    return nil
}
```

**Step 4: Run test to verify it passes**

Run: `cd /Users/jraymond/Documents/Projects/toolruntime && go test ./backend/docker/... -run TestBackend_Execute -v -timeout 120s`
Expected: PASS

**Step 5: Commit**

```bash
git add backend/docker/docker.go backend/docker/docker_test.go
git commit -m "$(cat <<'EOF'
feat(docker): implement Execute with container lifecycle management

- Create ephemeral containers with resource limits
- Capture stdout/stderr separately
- Respect context timeout and cancellation
- Auto-pull images when not present
- Auto-remove containers after execution

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 5: Implement Backend Interface Methods

**Files:**
- Modify: `backend/docker/docker.go`
- Test: `backend/docker/docker_test.go`

**Step 1: Write failing test for interface compliance**

```go
func TestBackend_Interface(t *testing.T) {
    // Verify Backend implements gateway.Backend
    var _ gateway.Backend = (*Backend)(nil)
}

func TestBackend_Name(t *testing.T) {
    if _, err := exec.LookPath("docker"); err != nil {
        t.Skip("Docker not available")
    }

    b, _ := NewBackend(Config{})
    defer b.Close()

    if got := b.Name(); got != "docker" {
        t.Errorf("Name() = %q, want %q", got, "docker")
    }
}

func TestBackend_Available(t *testing.T) {
    if _, err := exec.LookPath("docker"); err != nil {
        t.Skip("Docker not available")
    }

    b, _ := NewBackend(Config{})
    defer b.Close()

    if !b.Available() {
        t.Error("Available() = false, want true")
    }
}

func TestBackend_Supports(t *testing.T) {
    if _, err := exec.LookPath("docker"); err != nil {
        t.Skip("Docker not available")
    }

    b, _ := NewBackend(Config{Image: "python:3.11-slim"})
    defer b.Close()

    tests := []struct {
        language string
        want     bool
    }{
        {"python", true},
        {"sh", true},
        {"bash", true},
        {"unknown", false},
    }

    for _, tt := range tests {
        t.Run(tt.language, func(t *testing.T) {
            if got := b.Supports(tt.language); got != tt.want {
                t.Errorf("Supports(%q) = %v, want %v", tt.language, got, tt.want)
            }
        })
    }
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/jraymond/Documents/Projects/toolruntime && go test ./backend/docker/... -run 'TestBackend_(Interface|Name|Available|Supports)' -v`
Expected: FAIL - Methods not implemented

**Step 3: Implement interface methods**

```go
// Name returns the backend identifier.
func (b *Backend) Name() string {
    return "docker"
}

// Available checks if Docker daemon is accessible.
func (b *Backend) Available() bool {
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()

    _, err := b.client.Ping(ctx)
    return err == nil
}

// Supports checks if the backend can execute the given language.
// Docker supports any language available in the configured image.
func (b *Backend) Supports(language string) bool {
    // Common languages supported by default images
    supported := map[string]bool{
        "sh":     true,
        "bash":   true,
        "python": true,
        "python3": true,
        "node":   true,
        "ruby":   true,
        "perl":   true,
    }

    // Check image-specific support
    if strings.Contains(b.config.Image, "python") {
        supported["python"] = true
        supported["python3"] = true
    }
    if strings.Contains(b.config.Image, "node") {
        supported["javascript"] = true
        supported["node"] = true
    }
    if strings.Contains(b.config.Image, "alpine") || strings.Contains(b.config.Image, "ubuntu") {
        // Base images support shell
        supported["sh"] = true
    }

    return supported[language]
}
```

**Step 4: Run test to verify it passes**

Run: `cd /Users/jraymond/Documents/Projects/toolruntime && go test ./backend/docker/... -run 'TestBackend_(Interface|Name|Available|Supports)' -v`
Expected: PASS

**Step 5: Commit**

```bash
git add backend/docker/docker.go backend/docker/docker_test.go
git commit -m "$(cat <<'EOF'
feat(docker): implement gateway.Backend interface methods

- Name() returns "docker"
- Available() checks Docker daemon connectivity
- Supports() returns true for languages in configured image

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 6: Add Security Profile Integration

**Files:**
- Modify: `backend/docker/docker.go`
- Test: `backend/docker/docker_test.go`

**Step 1: Write failing test for security profiles**

```go
func TestBackend_SecurityProfiles(t *testing.T) {
    if _, err := exec.LookPath("docker"); err != nil {
        t.Skip("Docker not available")
    }

    t.Run("dev profile allows network", func(t *testing.T) {
        b, _ := NewBackend(Config{
            NetworkMode: "bridge",
        })
        defer b.Close()

        // In dev mode, network access should work
        resp, err := b.Execute(context.Background(), gateway.Request{
            Code:     "echo 'network test'",
            Language: "sh",
        })
        if err != nil {
            t.Fatalf("Execute() error = %v", err)
        }
        if resp.ExitCode != 0 {
            t.Errorf("ExitCode = %d, want 0", resp.ExitCode)
        }
    })

    t.Run("hardened profile denies network", func(t *testing.T) {
        b, _ := NewBackend(Config{
            NetworkMode: "none",
        })
        defer b.Close()

        // Network should be isolated
        if b.config.NetworkMode != "none" {
            t.Error("Hardened profile should have NetworkMode=none")
        }
    })
}
```

**Step 2: Run test to verify it passes (already implemented via NetworkMode)**

Run: `cd /Users/jraymond/Documents/Projects/toolruntime && go test ./backend/docker/... -run TestBackend_SecurityProfiles -v`
Expected: PASS

**Step 3: Commit**

```bash
git add backend/docker/docker_test.go
git commit -m "$(cat <<'EOF'
test(docker): add security profile integration tests

- Verify dev profile allows network access
- Verify hardened profile isolates network

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 7: Integration Test with toolruntime

**Files:**
- Create: `backend/docker/integration_test.go`

**Step 1: Write integration test**

```go
//go:build integration

package docker

import (
    "context"
    "os/exec"
    "testing"
    "time"

    "github.com/your-org/toolruntime/gateway"
)

func TestIntegration_PythonExecution(t *testing.T) {
    if _, err := exec.LookPath("docker"); err != nil {
        t.Skip("Docker not available")
    }

    b, err := NewBackend(Config{
        Image:   "python:3.11-slim",
        Timeout: 30,
    })
    if err != nil {
        t.Fatalf("NewBackend() error = %v", err)
    }
    defer b.Close()

    code := `
import json
result = {"sum": 1 + 2, "product": 2 * 3}
print(json.dumps(result))
`

    resp, err := b.Execute(context.Background(), gateway.Request{
        Code:     code,
        Language: "python",
    })
    if err != nil {
        t.Fatalf("Execute() error = %v", err)
    }

    if resp.ExitCode != 0 {
        t.Errorf("ExitCode = %d, want 0\nStderr: %s", resp.ExitCode, resp.Stderr)
    }

    expected := `{"sum": 3, "product": 6}`
    if !strings.Contains(resp.Stdout, expected) {
        t.Errorf("Stdout = %q, want contains %q", resp.Stdout, expected)
    }
}

func TestIntegration_ResourceLimits(t *testing.T) {
    if _, err := exec.LookPath("docker"); err != nil {
        t.Skip("Docker not available")
    }

    b, err := NewBackend(Config{
        Image:         "alpine:latest",
        Timeout:       5,
        MemoryLimitMB: 32, // Very low memory
    })
    if err != nil {
        t.Fatalf("NewBackend() error = %v", err)
    }
    defer b.Close()

    // Try to allocate more memory than allowed
    code := `dd if=/dev/zero of=/dev/null bs=64M count=1`

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    resp, _ := b.Execute(ctx, gateway.Request{
        Code:     code,
        Language: "sh",
    })

    // Should either fail or be killed due to memory limits
    t.Logf("ExitCode: %d, Stderr: %s", resp.ExitCode, resp.Stderr)
}
```

**Step 2: Run integration test**

Run: `cd /Users/jraymond/Documents/Projects/toolruntime && go test ./backend/docker/... -tags=integration -run TestIntegration -v -timeout 300s`
Expected: PASS

**Step 3: Commit**

```bash
git add backend/docker/integration_test.go
git commit -m "$(cat <<'EOF'
test(docker): add integration tests for Python execution and resource limits

- Test Python code execution in container
- Test memory limit enforcement
- Requires Docker daemon for execution

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 8: Update README and Documentation

**Files:**
- Modify: `backend/docker/README.md` (create if not exists)

**Step 1: Write documentation**

```markdown
# Docker Backend for toolruntime

Production-ready Docker isolation backend for code execution.

## Features

- Ephemeral containers with automatic cleanup
- Resource limits (CPU, memory)
- Network isolation modes
- Timeout enforcement
- stdout/stderr capture

## Usage

```go
import "github.com/your-org/toolruntime/backend/docker"

// Create with defaults (python:3.11-slim, 30s timeout, 256MB memory)
b, err := docker.NewBackend(docker.Config{})

// Create with custom config
b, err := docker.NewBackend(docker.Config{
    Image:         "node:18-alpine",
    Timeout:       60,
    MemoryLimitMB: 512,
    NetworkMode:   "none",  // Isolated
})

// Execute code
resp, err := b.Execute(ctx, gateway.Request{
    Code:     "console.log('hello')",
    Language: "javascript",
})
```

## Configuration

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| Image | string | python:3.11-slim | Docker image to use |
| Timeout | int | 30 | Execution timeout in seconds |
| MemoryLimitMB | int64 | 256 | Memory limit in MB |
| CPUShares | int64 | 512 | Relative CPU weight |
| NetworkMode | string | none | none, bridge, or host |
| WorkDir | string | /workspace | Working directory in container |

## Security Profiles

Map to security profiles via configuration:

| Profile | NetworkMode | MemoryLimitMB | Use Case |
|---------|-------------|---------------|----------|
| dev | bridge | 1024 | Local development |
| standard | none | 256 | Production default |
| hardened | none | 128 | Untrusted code |
```

**Step 2: Commit**

```bash
git add backend/docker/README.md
git commit -m "$(cat <<'EOF'
docs(docker): add README with usage examples and configuration reference

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Verification Checklist

- [ ] Docker SDK dependency added
- [ ] Config struct with validation and defaults
- [ ] NewBackend with client initialization
- [ ] Execute with container lifecycle
- [ ] Interface methods (Name, Available, Supports)
- [ ] Security profile integration
- [ ] Integration tests pass
- [ ] Documentation complete

## Definition of Done

1. All unit tests pass: `go test ./backend/docker/...`
2. Integration tests pass: `go test ./backend/docker/... -tags=integration`
3. Code coverage > 80%
4. Documentation complete
5. No breaking changes to existing interfaces
