# PRD-002: CLI Foundation with Cobra

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add Cobra CLI framework to metatools-mcp enabling subcommands (`serve`, `version`, `config`) and establishing the foundation for configuration-driven extensibility.

**Architecture:** Introduce Cobra as the CLI framework with a root command and initial subcommands. This replaces direct MCP server invocation with a structured CLI that supports flags, environment variables, and future config file integration. The existing stdio server behavior becomes `metatools serve --transport=stdio`.

**Tech Stack:** Cobra (github.com/spf13/cobra), Go 1.21+, existing metatools-mcp codebase

**Priority:** P0 - Foundation (Stream A, Phase 1 prerequisite)

**Scope:** CLI structure only - no config file loading yet (that's PRD-003)

---

## Context

The current metatools-mcp server starts directly via `main()` without CLI structure. The pluggable architecture proposal requires:
1. Subcommand support (`serve`, `version`, `config validate`)
2. Flag-based configuration (`--transport`, `--port`, `--config`)
3. Environment variable fallbacks
4. Help text and documentation

**Current State:**
```go
// cmd/metatools-mcp/main.go (current)
func main() {
    server, _ := server.New(cfg)
    _ = server.Run(context.Background(), &mcp.StdioTransport{})
}
```

**Target State:**
```bash
metatools serve              # Default: stdio transport
metatools serve --transport=sse --port=8080
metatools version
metatools config validate --config=metatools.yaml
```

---

## Tasks

### Task 1: Add Cobra Dependency

**Files:**
- Modify: `go.mod`
- Modify: `go.sum`

**Step 1: Add Cobra to go.mod**

Run:
```bash
cd /Users/jraymond/Documents/Projects/metatools-mcp && go get github.com/spf13/cobra@v1.8.0
```

**Step 2: Verify import works**

Run:
```bash
cd /Users/jraymond/Documents/Projects/metatools-mcp && go build ./...
```
Expected: Build succeeds

**Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "$(cat <<'EOF'
deps: add Cobra CLI framework

Foundation for subcommand-based CLI structure.

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 2: Create Root Command Structure

**Files:**
- Create: `cmd/metatools-mcp/cmd/root.go`
- Test: `cmd/metatools-mcp/cmd/root_test.go`

**Step 1: Write failing test for root command**

```go
// cmd/metatools-mcp/cmd/root_test.go
package cmd

import (
    "bytes"
    "testing"
)

func TestRootCmd_Help(t *testing.T) {
    cmd := NewRootCmd()
    buf := new(bytes.Buffer)
    cmd.SetOut(buf)
    cmd.SetArgs([]string{"--help"})

    err := cmd.Execute()
    if err != nil {
        t.Fatalf("Execute() error = %v", err)
    }

    output := buf.String()
    if !contains(output, "metatools-mcp") {
        t.Errorf("Help should contain 'metatools-mcp', got: %s", output)
    }
    if !contains(output, "serve") {
        t.Errorf("Help should list 'serve' subcommand, got: %s", output)
    }
    if !contains(output, "version") {
        t.Errorf("Help should list 'version' subcommand, got: %s", output)
    }
}

func contains(s, substr string) bool {
    return bytes.Contains([]byte(s), []byte(substr))
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/jraymond/Documents/Projects/metatools-mcp && go test ./cmd/metatools-mcp/cmd/... -run TestRootCmd_Help -v`
Expected: FAIL - NewRootCmd doesn't exist

**Step 3: Implement root command**

```go
// cmd/metatools-mcp/cmd/root.go
package cmd

import (
    "github.com/spf13/cobra"
)

var (
    // Version information (set at build time)
    Version   = "dev"
    GitCommit = "unknown"
    BuildDate = "unknown"
)

// NewRootCmd creates the root command for metatools-mcp.
func NewRootCmd() *cobra.Command {
    rootCmd := &cobra.Command{
        Use:   "metatools-mcp",
        Short: "MCP server for progressive tool discovery and execution",
        Long: `metatools-mcp is the MCP server that exposes the tool stack via a small,
progressive-disclosure tool surface. It composes toolmodel, toolindex, tooldocs,
toolrun, and optionally toolcode/toolruntime.

Use subcommands to start the server or manage configuration.`,
        SilenceUsage:  true,
        SilenceErrors: true,
    }

    // Add subcommands
    rootCmd.AddCommand(newServeCmd())
    rootCmd.AddCommand(newVersionCmd())

    return rootCmd
}

// Execute runs the root command.
func Execute() error {
    return NewRootCmd().Execute()
}
```

**Step 4: Run test to verify it passes**

Run: `cd /Users/jraymond/Documents/Projects/metatools-mcp && go test ./cmd/metatools-mcp/cmd/... -run TestRootCmd_Help -v`
Expected: PASS

**Step 5: Commit**

```bash
git add cmd/metatools-mcp/cmd/root.go cmd/metatools-mcp/cmd/root_test.go
git commit -m "$(cat <<'EOF'
feat(cli): add Cobra root command structure

- Create NewRootCmd() with program description
- Add Version, GitCommit, BuildDate variables for build-time injection
- Prepare for serve and version subcommands

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 3: Implement Version Command

**Files:**
- Create: `cmd/metatools-mcp/cmd/version.go`
- Test: `cmd/metatools-mcp/cmd/version_test.go`

**Step 1: Write failing test for version command**

```go
// cmd/metatools-mcp/cmd/version_test.go
package cmd

import (
    "bytes"
    "testing"
)

func TestVersionCmd(t *testing.T) {
    // Set test values
    Version = "1.2.3"
    GitCommit = "abc123"
    BuildDate = "2026-01-28"

    cmd := NewRootCmd()
    buf := new(bytes.Buffer)
    cmd.SetOut(buf)
    cmd.SetArgs([]string{"version"})

    err := cmd.Execute()
    if err != nil {
        t.Fatalf("Execute() error = %v", err)
    }

    output := buf.String()
    if !contains(output, "1.2.3") {
        t.Errorf("Version output should contain version, got: %s", output)
    }
    if !contains(output, "abc123") {
        t.Errorf("Version output should contain git commit, got: %s", output)
    }
}

func TestVersionCmd_JSON(t *testing.T) {
    Version = "1.2.3"
    GitCommit = "abc123"
    BuildDate = "2026-01-28"

    cmd := NewRootCmd()
    buf := new(bytes.Buffer)
    cmd.SetOut(buf)
    cmd.SetArgs([]string{"version", "--json"})

    err := cmd.Execute()
    if err != nil {
        t.Fatalf("Execute() error = %v", err)
    }

    output := buf.String()
    if !contains(output, `"version"`) {
        t.Errorf("JSON output should contain version field, got: %s", output)
    }
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/jraymond/Documents/Projects/metatools-mcp && go test ./cmd/metatools-mcp/cmd/... -run TestVersionCmd -v`
Expected: FAIL - newVersionCmd doesn't exist

**Step 3: Implement version command**

```go
// cmd/metatools-mcp/cmd/version.go
package cmd

import (
    "encoding/json"
    "fmt"
    "runtime"

    "github.com/spf13/cobra"
)

type versionInfo struct {
    Version   string `json:"version"`
    GitCommit string `json:"gitCommit"`
    BuildDate string `json:"buildDate"`
    GoVersion string `json:"goVersion"`
    Platform  string `json:"platform"`
}

func newVersionCmd() *cobra.Command {
    var jsonOutput bool

    cmd := &cobra.Command{
        Use:   "version",
        Short: "Print version information",
        Long:  "Print the version, git commit, build date, and Go version.",
        RunE: func(cmd *cobra.Command, args []string) error {
            info := versionInfo{
                Version:   Version,
                GitCommit: GitCommit,
                BuildDate: BuildDate,
                GoVersion: runtime.Version(),
                Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
            }

            if jsonOutput {
                enc := json.NewEncoder(cmd.OutOrStdout())
                enc.SetIndent("", "  ")
                return enc.Encode(info)
            }

            fmt.Fprintf(cmd.OutOrStdout(), "metatools-mcp %s\n", info.Version)
            fmt.Fprintf(cmd.OutOrStdout(), "  Git commit: %s\n", info.GitCommit)
            fmt.Fprintf(cmd.OutOrStdout(), "  Build date: %s\n", info.BuildDate)
            fmt.Fprintf(cmd.OutOrStdout(), "  Go version: %s\n", info.GoVersion)
            fmt.Fprintf(cmd.OutOrStdout(), "  Platform:   %s\n", info.Platform)
            return nil
        },
    }

    cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output version as JSON")

    return cmd
}
```

**Step 4: Run test to verify it passes**

Run: `cd /Users/jraymond/Documents/Projects/metatools-mcp && go test ./cmd/metatools-mcp/cmd/... -run TestVersionCmd -v`
Expected: PASS

**Step 5: Commit**

```bash
git add cmd/metatools-mcp/cmd/version.go cmd/metatools-mcp/cmd/version_test.go
git commit -m "$(cat <<'EOF'
feat(cli): add version subcommand with JSON output option

- Display version, git commit, build date, Go version, platform
- Support --json flag for machine-readable output

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 4: Implement Serve Command with Transport Flag

**Files:**
- Create: `cmd/metatools-mcp/cmd/serve.go`
- Test: `cmd/metatools-mcp/cmd/serve_test.go`

**Step 1: Write failing test for serve command flags**

```go
// cmd/metatools-mcp/cmd/serve_test.go
package cmd

import (
    "testing"
)

func TestServeCmd_Flags(t *testing.T) {
    cmd := newServeCmd()

    // Check transport flag exists
    transportFlag := cmd.Flags().Lookup("transport")
    if transportFlag == nil {
        t.Fatal("--transport flag not found")
    }
    if transportFlag.DefValue != "stdio" {
        t.Errorf("--transport default = %q, want %q", transportFlag.DefValue, "stdio")
    }

    // Check port flag exists
    portFlag := cmd.Flags().Lookup("port")
    if portFlag == nil {
        t.Fatal("--port flag not found")
    }
    if portFlag.DefValue != "8080" {
        t.Errorf("--port default = %q, want %q", portFlag.DefValue, "8080")
    }

    // Check config flag exists
    configFlag := cmd.Flags().Lookup("config")
    if configFlag == nil {
        t.Fatal("--config flag not found")
    }
}

func TestServeCmd_TransportValidation(t *testing.T) {
    tests := []struct {
        transport string
        wantErr   bool
    }{
        {"stdio", false},
        {"sse", false},
        {"http", false},
        {"invalid", true},
    }

    for _, tt := range tests {
        t.Run(tt.transport, func(t *testing.T) {
            err := validateTransport(tt.transport)
            if (err != nil) != tt.wantErr {
                t.Errorf("validateTransport(%q) error = %v, wantErr %v", tt.transport, err, tt.wantErr)
            }
        })
    }
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/jraymond/Documents/Projects/metatools-mcp && go test ./cmd/metatools-mcp/cmd/... -run TestServeCmd -v`
Expected: FAIL - newServeCmd doesn't exist

**Step 3: Implement serve command**

```go
// cmd/metatools-mcp/cmd/serve.go
package cmd

import (
    "context"
    "errors"
    "fmt"
    "os"
    "os/signal"
    "syscall"

    "github.com/spf13/cobra"
)

// ServeConfig holds serve command configuration.
type ServeConfig struct {
    Transport string
    Port      int
    Host      string
    Config    string
}

var validTransports = []string{"stdio", "sse", "http"}

func validateTransport(transport string) error {
    for _, valid := range validTransports {
        if transport == valid {
            return nil
        }
    }
    return fmt.Errorf("invalid transport %q, must be one of: %v", transport, validTransports)
}

func newServeCmd() *cobra.Command {
    cfg := &ServeConfig{}

    cmd := &cobra.Command{
        Use:   "serve",
        Short: "Start the MCP server",
        Long: `Start the metatools-mcp server with the specified transport.

Transports:
  stdio  - Standard input/output (default, for MCP clients like Claude Desktop)
  sse    - Server-Sent Events over HTTP (for web clients)
  http   - Simple HTTP request/response (for REST clients)

Examples:
  metatools-mcp serve                           # stdio mode (default)
  metatools-mcp serve --transport=sse --port=8080
  metatools-mcp serve --config=metatools.yaml`,
        PreRunE: func(cmd *cobra.Command, args []string) error {
            return validateTransport(cfg.Transport)
        },
        RunE: func(cmd *cobra.Command, args []string) error {
            return runServe(cmd.Context(), cfg)
        },
    }

    // Flags
    cmd.Flags().StringVarP(&cfg.Transport, "transport", "t", "stdio", "Transport type (stdio, sse, http)")
    cmd.Flags().IntVarP(&cfg.Port, "port", "p", 8080, "Port for HTTP transports")
    cmd.Flags().StringVar(&cfg.Host, "host", "0.0.0.0", "Host to bind for HTTP transports")
    cmd.Flags().StringVarP(&cfg.Config, "config", "c", "", "Path to config file")

    return cmd
}

func runServe(ctx context.Context, cfg *ServeConfig) error {
    // Setup signal handling
    ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
    defer cancel()

    // For now, just print what would happen
    // The actual server integration will be added in the next task
    fmt.Printf("Starting server with transport=%s\n", cfg.Transport)
    if cfg.Transport != "stdio" {
        fmt.Printf("Listening on %s:%d\n", cfg.Host, cfg.Port)
    }

    // TODO: Integrate with existing server code
    // This placeholder will be replaced when we integrate with the server package

    <-ctx.Done()
    fmt.Println("Shutting down...")
    return nil
}
```

**Step 4: Run test to verify it passes**

Run: `cd /Users/jraymond/Documents/Projects/metatools-mcp && go test ./cmd/metatools-mcp/cmd/... -run TestServeCmd -v`
Expected: PASS

**Step 5: Commit**

```bash
git add cmd/metatools-mcp/cmd/serve.go cmd/metatools-mcp/cmd/serve_test.go
git commit -m "$(cat <<'EOF'
feat(cli): add serve subcommand with transport, port, config flags

- Support --transport flag (stdio, sse, http)
- Support --port and --host flags for HTTP transports
- Support --config flag for config file path
- Add transport validation
- Placeholder for server integration

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 5: Update Main to Use Cobra

**Files:**
- Modify: `cmd/metatools-mcp/main.go`
- Test: Manual verification

**Step 1: Read current main.go**

Run: Read the current main.go to understand the existing structure.

**Step 2: Update main.go to use Cobra**

```go
// cmd/metatools-mcp/main.go
package main

import (
    "fmt"
    "os"

    "github.com/your-org/metatools-mcp/cmd/metatools-mcp/cmd"
)

func main() {
    if err := cmd.Execute(); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}
```

**Step 3: Verify it builds**

Run:
```bash
cd /Users/jraymond/Documents/Projects/metatools-mcp && go build ./cmd/metatools-mcp/...
```
Expected: Build succeeds

**Step 4: Verify help works**

Run:
```bash
cd /Users/jraymond/Documents/Projects/metatools-mcp && ./cmd/metatools-mcp/metatools-mcp --help
```
Expected: Shows help with serve and version subcommands

**Step 5: Commit**

```bash
git add cmd/metatools-mcp/main.go
git commit -m "$(cat <<'EOF'
feat(cli): update main.go to use Cobra CLI

- Replace direct server invocation with Cobra Execute()
- Maintain backward compatibility via 'serve' subcommand

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 6: Integrate Serve Command with Existing Server

**Files:**
- Modify: `cmd/metatools-mcp/cmd/serve.go`
- Test: `cmd/metatools-mcp/cmd/serve_test.go`

**Step 1: Write integration test**

```go
// Add to serve_test.go
func TestServeCmd_StdioIntegration(t *testing.T) {
    // This test verifies the serve command can create a server
    // It doesn't run the full server, just validates wiring

    cfg := &ServeConfig{
        Transport: "stdio",
    }

    // Verify we can create server config from CLI config
    serverCfg, err := buildServerConfig(cfg)
    if err != nil {
        t.Fatalf("buildServerConfig() error = %v", err)
    }

    if serverCfg == nil {
        t.Fatal("buildServerConfig() returned nil")
    }
}
```

**Step 2: Implement server integration**

Update `serve.go` to integrate with the existing server package:

```go
// Add to serve.go
import (
    "github.com/your-org/metatools-mcp/internal/adapters"
    "github.com/your-org/metatools-mcp/internal/server"
    "github.com/mark3labs/mcp-go/mcp"
    // ... existing imports
)

func buildServerConfig(cfg *ServeConfig) (*adapters.Config, error) {
    // Build server configuration from CLI flags
    // This bridges CLI config to the existing server config

    // For now, use existing defaults
    // Future PRDs will add Koanf config file loading here

    idx := toolindex.NewInMemoryIndex()
    docs := tooldocs.NewInMemoryStore(tooldocs.StoreOptions{Index: idx})
    runner := toolrun.NewRunner(toolrun.WithIndex(idx))

    return adapters.NewConfig(idx, docs, runner, nil), nil
}

func runServe(ctx context.Context, cfg *ServeConfig) error {
    ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
    defer cancel()

    serverCfg, err := buildServerConfig(cfg)
    if err != nil {
        return fmt.Errorf("build server config: %w", err)
    }

    srv, err := server.New(serverCfg)
    if err != nil {
        return fmt.Errorf("create server: %w", err)
    }

    // Select transport based on flag
    var transport mcp.Transport
    switch cfg.Transport {
    case "stdio":
        transport = &mcp.StdioTransport{}
    case "sse", "http":
        // SSE/HTTP transports will be implemented in PRD-004
        return fmt.Errorf("transport %q not yet implemented", cfg.Transport)
    default:
        return fmt.Errorf("unknown transport: %s", cfg.Transport)
    }

    fmt.Fprintf(os.Stderr, "Starting metatools-mcp server (transport=%s)\n", cfg.Transport)
    return srv.Run(ctx, transport)
}
```

**Step 3: Run test to verify it passes**

Run: `cd /Users/jraymond/Documents/Projects/metatools-mcp && go test ./cmd/metatools-mcp/cmd/... -v`
Expected: PASS

**Step 4: Commit**

```bash
git add cmd/metatools-mcp/cmd/serve.go cmd/metatools-mcp/cmd/serve_test.go
git commit -m "$(cat <<'EOF'
feat(cli): integrate serve command with existing server package

- Bridge CLI config to adapters.Config
- Create server with stdio transport
- Placeholder for SSE/HTTP transports (future PRD)

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 7: Add Build-Time Version Injection

**Files:**
- Modify: `Makefile` (or create if not exists)
- Modify: `.goreleaser.yaml` (if using goreleaser)

**Step 1: Create/update Makefile**

```makefile
# Makefile
VERSION ?= $(shell git describe --tags --always --dirty)
GIT_COMMIT ?= $(shell git rev-parse --short HEAD)
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS := -X github.com/your-org/metatools-mcp/cmd/metatools-mcp/cmd.Version=$(VERSION)
LDFLAGS += -X github.com/your-org/metatools-mcp/cmd/metatools-mcp/cmd.GitCommit=$(GIT_COMMIT)
LDFLAGS += -X github.com/your-org/metatools-mcp/cmd/metatools-mcp/cmd.BuildDate=$(BUILD_DATE)

.PHONY: build
build:
	go build -ldflags "$(LDFLAGS)" -o bin/metatools-mcp ./cmd/metatools-mcp

.PHONY: install
install:
	go install -ldflags "$(LDFLAGS)" ./cmd/metatools-mcp

.PHONY: test
test:
	go test ./...

.PHONY: clean
clean:
	rm -rf bin/
```

**Step 2: Verify build with version injection**

Run:
```bash
cd /Users/jraymond/Documents/Projects/metatools-mcp && make build && ./bin/metatools-mcp version
```
Expected: Shows actual git version and commit

**Step 3: Commit**

```bash
git add Makefile
git commit -m "$(cat <<'EOF'
build: add Makefile with version injection via ldflags

- Inject Version, GitCommit, BuildDate at build time
- Add build, install, test, clean targets

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 8: Add Environment Variable Support

**Files:**
- Modify: `cmd/metatools-mcp/cmd/serve.go`
- Test: `cmd/metatools-mcp/cmd/serve_test.go`

**Step 1: Write failing test for env var support**

```go
// Add to serve_test.go
func TestServeCmd_EnvVars(t *testing.T) {
    // Save and restore env
    oldTransport := os.Getenv("METATOOLS_TRANSPORT")
    oldPort := os.Getenv("METATOOLS_PORT")
    defer func() {
        os.Setenv("METATOOLS_TRANSPORT", oldTransport)
        os.Setenv("METATOOLS_PORT", oldPort)
    }()

    os.Setenv("METATOOLS_TRANSPORT", "sse")
    os.Setenv("METATOOLS_PORT", "9090")

    cmd := newServeCmd()

    // Parse empty args to trigger env var loading
    cmd.ParseFlags([]string{})

    // Flags should pick up env vars
    transport, _ := cmd.Flags().GetString("transport")
    port, _ := cmd.Flags().GetInt("port")

    if transport != "sse" {
        t.Errorf("transport = %q, want %q from env", transport, "sse")
    }
    if port != 9090 {
        t.Errorf("port = %d, want %d from env", port, 9090)
    }
}
```

**Step 2: Implement env var support**

Update `serve.go` to read env vars:

```go
// Add to serve.go in newServeCmd()
func newServeCmd() *cobra.Command {
    cfg := &ServeConfig{}

    // ... existing code ...

    // Bind environment variables
    // Cobra doesn't have built-in env support, so we do it manually
    cmd.PreRun = func(cmd *cobra.Command, args []string) {
        // Only use env if flag wasn't explicitly set
        if !cmd.Flags().Changed("transport") {
            if v := os.Getenv("METATOOLS_TRANSPORT"); v != "" {
                cfg.Transport = v
            }
        }
        if !cmd.Flags().Changed("port") {
            if v := os.Getenv("METATOOLS_PORT"); v != "" {
                if port, err := strconv.Atoi(v); err == nil {
                    cfg.Port = port
                }
            }
        }
        if !cmd.Flags().Changed("host") {
            if v := os.Getenv("METATOOLS_HOST"); v != "" {
                cfg.Host = v
            }
        }
        if !cmd.Flags().Changed("config") {
            if v := os.Getenv("METATOOLS_CONFIG"); v != "" {
                cfg.Config = v
            }
        }
    }

    return cmd
}
```

**Step 3: Run test to verify it passes**

Run: `cd /Users/jraymond/Documents/Projects/metatools-mcp && go test ./cmd/metatools-mcp/cmd/... -run TestServeCmd_EnvVars -v`
Expected: PASS

**Step 4: Commit**

```bash
git add cmd/metatools-mcp/cmd/serve.go cmd/metatools-mcp/cmd/serve_test.go
git commit -m "$(cat <<'EOF'
feat(cli): add environment variable support for serve flags

Environment variables (lower precedence than flags):
- METATOOLS_TRANSPORT
- METATOOLS_PORT
- METATOOLS_HOST
- METATOOLS_CONFIG

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Verification Checklist

- [ ] Cobra dependency added
- [ ] Root command with help text
- [ ] Version subcommand with --json flag
- [ ] Serve subcommand with --transport, --port, --host, --config flags
- [ ] Transport validation (stdio, sse, http)
- [ ] Environment variable fallbacks
- [ ] Build-time version injection
- [ ] Integration with existing server package
- [ ] All tests pass

## Definition of Done

1. All unit tests pass: `go test ./cmd/metatools-mcp/...`
2. `metatools-mcp --help` shows all subcommands
3. `metatools-mcp version` displays version info
4. `metatools-mcp serve` starts stdio server (backward compatible)
5. `METATOOLS_TRANSPORT=sse metatools-mcp serve` reads from env
6. Build injects version: `make build && ./bin/metatools-mcp version`

## Next PRD

PRD-003 will add Koanf configuration file support, enabling `metatools.yaml` config files.
