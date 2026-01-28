# PRD-003: Configuration Layer with Koanf

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add Koanf configuration file support enabling YAML-based configuration for all extension points, with environment variable overrides and validation.

**Architecture:** Introduce Koanf as the configuration library with a structured Config type that maps to all 13 extension points. Configuration is loaded in order: defaults → config file → environment variables → CLI flags. The config exposes search strategy, backend selection, runtime settings, and middleware options.

**Tech Stack:** Koanf (github.com/knadh/koanf), Go 1.21+, existing Cobra CLI from PRD-002

**Priority:** P0 - Foundation (Stream A, Phase 1 - enables all pluggability)

**Scope:** Config loading and validation - actual feature flags/middleware deferred to later PRDs

**Dependencies:** PRD-002 (CLI Foundation)

---

## Context

With Cobra CLI in place (PRD-002), we need configuration file support. The pluggable architecture proposal defines a comprehensive config schema covering:
- Server settings (name, version)
- Transport settings (type, port, TLS)
- Search strategy (bm25, semantic)
- Execution settings (timeout, limits)
- Backend configuration (local, MCP, API)
- Middleware chain (logging, auth, rate limit)

**Current State:** CLI flags only, no config file support

**Target State:**
```yaml
# metatools.yaml
server:
  name: metatools-mcp
  version: "0.2.0"

transport:
  type: stdio  # or sse, http

search:
  strategy: bm25
  bm25:
    name_boost: 3.0

execution:
  timeout: 30s
  max_tool_calls: 64
```

---

## Tasks

### Task 1: Add Koanf Dependency

**Files:**
- Modify: `go.mod`
- Modify: `go.sum`

**Step 1: Add Koanf with YAML and env providers**

Run:
```bash
cd /Users/jraymond/Documents/Projects/metatools-mcp && \
go get github.com/knadh/koanf/v2@v2.1.0 && \
go get github.com/knadh/koanf/parsers/yaml@v0.1.0 && \
go get github.com/knadh/koanf/providers/file@v0.1.0 && \
go get github.com/knadh/koanf/providers/env@v0.1.0 && \
go get github.com/knadh/koanf/providers/structs@v0.1.0
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
deps: add Koanf configuration library with YAML and env providers

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 2: Define Configuration Struct

**Files:**
- Create: `internal/config/config.go`
- Test: `internal/config/config_test.go`

**Step 1: Write failing test for Config struct**

```go
// internal/config/config_test.go
package config

import (
    "testing"
    "time"
)

func TestDefaultConfig(t *testing.T) {
    cfg := DefaultConfig()

    // Server defaults
    if cfg.Server.Name != "metatools-mcp" {
        t.Errorf("Server.Name = %q, want %q", cfg.Server.Name, "metatools-mcp")
    }

    // Transport defaults
    if cfg.Transport.Type != "stdio" {
        t.Errorf("Transport.Type = %q, want %q", cfg.Transport.Type, "stdio")
    }
    if cfg.Transport.HTTP.Port != 8080 {
        t.Errorf("Transport.HTTP.Port = %d, want %d", cfg.Transport.HTTP.Port, 8080)
    }

    // Search defaults
    if cfg.Search.Strategy != "bm25" {
        t.Errorf("Search.Strategy = %q, want %q", cfg.Search.Strategy, "bm25")
    }

    // Execution defaults
    if cfg.Execution.Timeout != 30*time.Second {
        t.Errorf("Execution.Timeout = %v, want %v", cfg.Execution.Timeout, 30*time.Second)
    }
    if cfg.Execution.MaxToolCalls != 64 {
        t.Errorf("Execution.MaxToolCalls = %d, want %d", cfg.Execution.MaxToolCalls, 64)
    }
}

func TestConfig_Validate(t *testing.T) {
    tests := []struct {
        name    string
        modify  func(*Config)
        wantErr bool
    }{
        {
            name:    "default config is valid",
            modify:  func(c *Config) {},
            wantErr: false,
        },
        {
            name:    "invalid transport type",
            modify:  func(c *Config) { c.Transport.Type = "invalid" },
            wantErr: true,
        },
        {
            name:    "invalid search strategy",
            modify:  func(c *Config) { c.Search.Strategy = "invalid" },
            wantErr: true,
        },
        {
            name:    "negative timeout",
            modify:  func(c *Config) { c.Execution.Timeout = -1 * time.Second },
            wantErr: true,
        },
        {
            name:    "zero max tool calls uses default",
            modify:  func(c *Config) { c.Execution.MaxToolCalls = 0 },
            wantErr: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            cfg := DefaultConfig()
            tt.modify(&cfg)
            err := cfg.Validate()
            if (err != nil) != tt.wantErr {
                t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/jraymond/Documents/Projects/metatools-mcp && go test ./internal/config/... -v`
Expected: FAIL - Config type doesn't exist

**Step 3: Implement Config struct**

```go
// internal/config/config.go
package config

import (
    "errors"
    "fmt"
    "time"
)

// Config holds all metatools-mcp configuration.
type Config struct {
    Server    ServerConfig    `koanf:"server"`
    Transport TransportConfig `koanf:"transport"`
    Search    SearchConfig    `koanf:"search"`
    Execution ExecutionConfig `koanf:"execution"`
    Providers ProvidersConfig `koanf:"providers"`
    Backends  BackendsConfig  `koanf:"backends"`
}

// ServerConfig holds server identity settings.
type ServerConfig struct {
    Name    string `koanf:"name"`
    Version string `koanf:"version"`
}

// TransportConfig holds transport layer settings.
type TransportConfig struct {
    Type string           `koanf:"type"`
    HTTP HTTPConfig       `koanf:"http"`
}

// HTTPConfig holds HTTP transport settings.
type HTTPConfig struct {
    Host     string `koanf:"host"`
    Port     int    `koanf:"port"`
    TLS      TLSConfig `koanf:"tls"`
}

// TLSConfig holds TLS settings.
type TLSConfig struct {
    Enabled  bool   `koanf:"enabled"`
    CertFile string `koanf:"cert"`
    KeyFile  string `koanf:"key"`
}

// SearchConfig holds search strategy settings.
type SearchConfig struct {
    Strategy string      `koanf:"strategy"`
    BM25     BM25Config  `koanf:"bm25"`
}

// BM25Config holds BM25 search settings.
type BM25Config struct {
    NameBoost      float64 `koanf:"name_boost"`
    NamespaceBoost float64 `koanf:"namespace_boost"`
    TagsBoost      float64 `koanf:"tags_boost"`
    MaxDocs        int     `koanf:"max_docs"`
    MaxDocTextLen  int     `koanf:"max_doctext_len"`
}

// ExecutionConfig holds tool execution settings.
type ExecutionConfig struct {
    Timeout       time.Duration `koanf:"timeout"`
    MaxToolCalls  int           `koanf:"max_tool_calls"`
    MaxChainSteps int           `koanf:"max_chain_steps"`
}

// ProvidersConfig holds tool provider settings.
type ProvidersConfig struct {
    SearchTools  ProviderEnabled `koanf:"search_tools"`
    DescribeTool ProviderEnabled `koanf:"describe_tool"`
    RunTool      ProviderEnabled `koanf:"run_tool"`
    RunChain     ProviderEnabled `koanf:"run_chain"`
    ExecuteCode  ExecuteCodeConfig `koanf:"execute_code"`
}

// ProviderEnabled is a simple on/off provider config.
type ProviderEnabled struct {
    Enabled bool `koanf:"enabled"`
}

// ExecuteCodeConfig holds code execution provider settings.
type ExecuteCodeConfig struct {
    Enabled bool   `koanf:"enabled"`
    Sandbox string `koanf:"sandbox"`
}

// BackendsConfig holds backend source settings.
type BackendsConfig struct {
    Local   LocalBackendConfig `koanf:"local"`
    // Future: OpenAI, MCP, HTTP backends
}

// LocalBackendConfig holds local tool backend settings.
type LocalBackendConfig struct {
    Enabled bool     `koanf:"enabled"`
    Paths   []string `koanf:"paths"`
    Watch   bool     `koanf:"watch"`
}

// Valid transport types.
var validTransports = map[string]bool{
    "stdio": true,
    "sse":   true,
    "http":  true,
}

// Valid search strategies.
var validSearchStrategies = map[string]bool{
    "bm25":     true,
    "lexical":  true,
    "semantic": true,
}

// DefaultConfig returns the default configuration.
func DefaultConfig() Config {
    return Config{
        Server: ServerConfig{
            Name:    "metatools-mcp",
            Version: "0.2.0",
        },
        Transport: TransportConfig{
            Type: "stdio",
            HTTP: HTTPConfig{
                Host: "0.0.0.0",
                Port: 8080,
            },
        },
        Search: SearchConfig{
            Strategy: "bm25",
            BM25: BM25Config{
                NameBoost:      3.0,
                NamespaceBoost: 2.0,
                TagsBoost:      2.0,
                MaxDocs:        0,        // unlimited
                MaxDocTextLen:  0,        // unlimited
            },
        },
        Execution: ExecutionConfig{
            Timeout:       30 * time.Second,
            MaxToolCalls:  64,
            MaxChainSteps: 8,
        },
        Providers: ProvidersConfig{
            SearchTools:  ProviderEnabled{Enabled: true},
            DescribeTool: ProviderEnabled{Enabled: true},
            RunTool:      ProviderEnabled{Enabled: true},
            RunChain:     ProviderEnabled{Enabled: true},
            ExecuteCode:  ExecuteCodeConfig{Enabled: false, Sandbox: "dev"},
        },
        Backends: BackendsConfig{
            Local: LocalBackendConfig{
                Enabled: true,
                Paths:   []string{},
                Watch:   false,
            },
        },
    }
}

// Validate checks the configuration for errors.
func (c *Config) Validate() error {
    // Transport validation
    if !validTransports[c.Transport.Type] {
        return fmt.Errorf("invalid transport type %q, must be one of: stdio, sse, http", c.Transport.Type)
    }

    // HTTP validation
    if c.Transport.Type != "stdio" {
        if c.Transport.HTTP.Port <= 0 || c.Transport.HTTP.Port > 65535 {
            return fmt.Errorf("invalid port %d, must be 1-65535", c.Transport.HTTP.Port)
        }
    }

    // Search validation
    if !validSearchStrategies[c.Search.Strategy] {
        return fmt.Errorf("invalid search strategy %q, must be one of: bm25, lexical, semantic", c.Search.Strategy)
    }

    // Execution validation
    if c.Execution.Timeout < 0 {
        return errors.New("execution timeout cannot be negative")
    }

    return nil
}
```

**Step 4: Run test to verify it passes**

Run: `cd /Users/jraymond/Documents/Projects/metatools-mcp && go test ./internal/config/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/config/config.go internal/config/config_test.go
git commit -m "$(cat <<'EOF'
feat(config): add Config struct with defaults and validation

Define configuration for:
- Server (name, version)
- Transport (type, HTTP settings)
- Search (strategy, BM25 params)
- Execution (timeout, limits)
- Providers (enabled flags)
- Backends (local paths)

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 3: Implement Koanf Loader

**Files:**
- Create: `internal/config/loader.go`
- Test: `internal/config/loader_test.go`

**Step 1: Write failing test for config loading**

```go
// internal/config/loader_test.go
package config

import (
    "os"
    "path/filepath"
    "testing"
    "time"
)

func TestLoad_DefaultsWhenNoFile(t *testing.T) {
    cfg, err := Load("")
    if err != nil {
        t.Fatalf("Load() error = %v", err)
    }

    // Should have defaults
    if cfg.Transport.Type != "stdio" {
        t.Errorf("Transport.Type = %q, want %q", cfg.Transport.Type, "stdio")
    }
}

func TestLoad_FromYAMLFile(t *testing.T) {
    // Create temp config file
    dir := t.TempDir()
    configPath := filepath.Join(dir, "metatools.yaml")

    yaml := `
server:
  name: test-server
transport:
  type: sse
  http:
    port: 9090
search:
  strategy: bm25
  bm25:
    name_boost: 5.0
execution:
  timeout: 60s
`
    if err := os.WriteFile(configPath, []byte(yaml), 0644); err != nil {
        t.Fatalf("WriteFile() error = %v", err)
    }

    cfg, err := Load(configPath)
    if err != nil {
        t.Fatalf("Load() error = %v", err)
    }

    if cfg.Server.Name != "test-server" {
        t.Errorf("Server.Name = %q, want %q", cfg.Server.Name, "test-server")
    }
    if cfg.Transport.Type != "sse" {
        t.Errorf("Transport.Type = %q, want %q", cfg.Transport.Type, "sse")
    }
    if cfg.Transport.HTTP.Port != 9090 {
        t.Errorf("Transport.HTTP.Port = %d, want %d", cfg.Transport.HTTP.Port, 9090)
    }
    if cfg.Search.BM25.NameBoost != 5.0 {
        t.Errorf("Search.BM25.NameBoost = %f, want %f", cfg.Search.BM25.NameBoost, 5.0)
    }
    if cfg.Execution.Timeout != 60*time.Second {
        t.Errorf("Execution.Timeout = %v, want %v", cfg.Execution.Timeout, 60*time.Second)
    }
}

func TestLoad_EnvOverrides(t *testing.T) {
    // Set env vars
    os.Setenv("METATOOLS_TRANSPORT_TYPE", "http")
    os.Setenv("METATOOLS_TRANSPORT_HTTP_PORT", "3000")
    defer func() {
        os.Unsetenv("METATOOLS_TRANSPORT_TYPE")
        os.Unsetenv("METATOOLS_TRANSPORT_HTTP_PORT")
    }()

    cfg, err := Load("")
    if err != nil {
        t.Fatalf("Load() error = %v", err)
    }

    if cfg.Transport.Type != "http" {
        t.Errorf("Transport.Type = %q, want %q from env", cfg.Transport.Type, "http")
    }
    if cfg.Transport.HTTP.Port != 3000 {
        t.Errorf("Transport.HTTP.Port = %d, want %d from env", cfg.Transport.HTTP.Port, 3000)
    }
}

func TestLoad_InvalidYAML(t *testing.T) {
    dir := t.TempDir()
    configPath := filepath.Join(dir, "bad.yaml")

    if err := os.WriteFile(configPath, []byte("invalid: yaml: ["), 0644); err != nil {
        t.Fatalf("WriteFile() error = %v", err)
    }

    _, err := Load(configPath)
    if err == nil {
        t.Error("Load() should fail with invalid YAML")
    }
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/jraymond/Documents/Projects/metatools-mcp && go test ./internal/config/... -run TestLoad -v`
Expected: FAIL - Load function doesn't exist

**Step 3: Implement Koanf loader**

```go
// internal/config/loader.go
package config

import (
    "fmt"
    "os"
    "strings"

    "github.com/knadh/koanf/parsers/yaml"
    "github.com/knadh/koanf/providers/env"
    "github.com/knadh/koanf/providers/file"
    "github.com/knadh/koanf/providers/structs"
    "github.com/knadh/koanf/v2"
)

// Load loads configuration from file and environment.
// Order of precedence (lowest to highest):
// 1. Default values
// 2. Config file (if path provided)
// 3. Environment variables (METATOOLS_*)
func Load(configPath string) (*Config, error) {
    k := koanf.New(".")

    // 1. Load defaults
    defaults := DefaultConfig()
    if err := k.Load(structs.Provider(defaults, "koanf"), nil); err != nil {
        return nil, fmt.Errorf("load defaults: %w", err)
    }

    // 2. Load config file if provided
    if configPath != "" {
        if _, err := os.Stat(configPath); err == nil {
            if err := k.Load(file.Provider(configPath), yaml.Parser()); err != nil {
                return nil, fmt.Errorf("load config file: %w", err)
            }
        } else if !os.IsNotExist(err) {
            return nil, fmt.Errorf("stat config file: %w", err)
        }
    }

    // 3. Load environment variables
    // METATOOLS_TRANSPORT_TYPE -> transport.type
    envProvider := env.Provider("METATOOLS_", ".", func(s string) string {
        return strings.Replace(
            strings.ToLower(strings.TrimPrefix(s, "METATOOLS_")),
            "_",
            ".",
            -1,
        )
    })
    if err := k.Load(envProvider, nil); err != nil {
        return nil, fmt.Errorf("load env vars: %w", err)
    }

    // Unmarshal to Config struct
    var cfg Config
    if err := k.Unmarshal("", &cfg); err != nil {
        return nil, fmt.Errorf("unmarshal config: %w", err)
    }

    // Validate
    if err := cfg.Validate(); err != nil {
        return nil, fmt.Errorf("validate config: %w", err)
    }

    return &cfg, nil
}

// MustLoad loads configuration or panics.
func MustLoad(configPath string) *Config {
    cfg, err := Load(configPath)
    if err != nil {
        panic(err)
    }
    return cfg
}

// LoadWithOverrides loads config with CLI flag overrides.
func LoadWithOverrides(configPath string, overrides map[string]interface{}) (*Config, error) {
    k := koanf.New(".")

    // 1. Load defaults
    defaults := DefaultConfig()
    if err := k.Load(structs.Provider(defaults, "koanf"), nil); err != nil {
        return nil, fmt.Errorf("load defaults: %w", err)
    }

    // 2. Load config file
    if configPath != "" {
        if _, err := os.Stat(configPath); err == nil {
            if err := k.Load(file.Provider(configPath), yaml.Parser()); err != nil {
                return nil, fmt.Errorf("load config file: %w", err)
            }
        }
    }

    // 3. Load environment variables
    envProvider := env.Provider("METATOOLS_", ".", func(s string) string {
        return strings.Replace(
            strings.ToLower(strings.TrimPrefix(s, "METATOOLS_")),
            "_",
            ".",
            -1,
        )
    })
    if err := k.Load(envProvider, nil); err != nil {
        return nil, fmt.Errorf("load env vars: %w", err)
    }

    // 4. Apply CLI overrides (highest precedence)
    for key, value := range overrides {
        k.Set(key, value)
    }

    // Unmarshal
    var cfg Config
    if err := k.Unmarshal("", &cfg); err != nil {
        return nil, fmt.Errorf("unmarshal config: %w", err)
    }

    // Validate
    if err := cfg.Validate(); err != nil {
        return nil, fmt.Errorf("validate config: %w", err)
    }

    return &cfg, nil
}
```

**Step 4: Run test to verify it passes**

Run: `cd /Users/jraymond/Documents/Projects/metatools-mcp && go test ./internal/config/... -run TestLoad -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/config/loader.go internal/config/loader_test.go
git commit -m "$(cat <<'EOF'
feat(config): implement Koanf config loader with env overrides

- Load defaults, config file, then env vars (in order of precedence)
- Support METATOOLS_* environment variables
- Add LoadWithOverrides for CLI flag integration
- Validate config after loading

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 4: Integrate Config with Serve Command

**Files:**
- Modify: `cmd/metatools-mcp/cmd/serve.go`
- Test: `cmd/metatools-mcp/cmd/serve_test.go`

**Step 1: Write integration test**

```go
// Add to serve_test.go
func TestServeCmd_ConfigFile(t *testing.T) {
    // Create temp config file
    dir := t.TempDir()
    configPath := filepath.Join(dir, "metatools.yaml")

    yaml := `
transport:
  type: sse
  http:
    port: 9999
`
    if err := os.WriteFile(configPath, []byte(yaml), 0644); err != nil {
        t.Fatalf("WriteFile() error = %v", err)
    }

    cfg, err := loadServeConfig(configPath, &ServeConfig{})
    if err != nil {
        t.Fatalf("loadServeConfig() error = %v", err)
    }

    if cfg.Transport.Type != "sse" {
        t.Errorf("Transport.Type = %q, want %q", cfg.Transport.Type, "sse")
    }
    if cfg.Transport.HTTP.Port != 9999 {
        t.Errorf("Transport.HTTP.Port = %d, want %d", cfg.Transport.HTTP.Port, 9999)
    }
}

func TestServeCmd_CLIOverridesConfig(t *testing.T) {
    // Create temp config file with port 9999
    dir := t.TempDir()
    configPath := filepath.Join(dir, "metatools.yaml")

    yaml := `
transport:
  type: sse
  http:
    port: 9999
`
    if err := os.WriteFile(configPath, []byte(yaml), 0644); err != nil {
        t.Fatalf("WriteFile() error = %v", err)
    }

    // CLI flag should override config file
    cliCfg := &ServeConfig{
        Transport: "http",  // Override from CLI
        Port:      3000,    // Override from CLI
    }

    cfg, err := loadServeConfig(configPath, cliCfg)
    if err != nil {
        t.Fatalf("loadServeConfig() error = %v", err)
    }

    // CLI should win over config file
    if cfg.Transport.Type != "http" {
        t.Errorf("Transport.Type = %q, want %q from CLI", cfg.Transport.Type, "http")
    }
    if cfg.Transport.HTTP.Port != 3000 {
        t.Errorf("Transport.HTTP.Port = %d, want %d from CLI", cfg.Transport.HTTP.Port, 3000)
    }
}
```

**Step 2: Update serve.go to use config**

```go
// Update serve.go
import (
    "github.com/your-org/metatools-mcp/internal/config"
    // ... existing imports
)

// loadServeConfig loads config with CLI overrides.
func loadServeConfig(configPath string, cli *ServeConfig) (*config.Config, error) {
    // Build overrides from CLI flags
    overrides := make(map[string]interface{})

    // Only override if CLI flag was explicitly set (non-default)
    if cli.Transport != "" && cli.Transport != "stdio" {
        overrides["transport.type"] = cli.Transport
    }
    if cli.Port != 0 && cli.Port != 8080 {
        overrides["transport.http.port"] = cli.Port
    }
    if cli.Host != "" && cli.Host != "0.0.0.0" {
        overrides["transport.http.host"] = cli.Host
    }

    return config.LoadWithOverrides(configPath, overrides)
}

// Update runServe to use config
func runServe(ctx context.Context, cliCfg *ServeConfig) error {
    ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
    defer cancel()

    // Load configuration
    cfg, err := loadServeConfig(cliCfg.Config, cliCfg)
    if err != nil {
        return fmt.Errorf("load config: %w", err)
    }

    // Build server from config
    serverCfg, err := buildServerConfigFromConfig(cfg)
    if err != nil {
        return fmt.Errorf("build server config: %w", err)
    }

    srv, err := server.New(serverCfg)
    if err != nil {
        return fmt.Errorf("create server: %w", err)
    }

    // Select transport
    var transport mcp.Transport
    switch cfg.Transport.Type {
    case "stdio":
        transport = &mcp.StdioTransport{}
    case "sse", "http":
        return fmt.Errorf("transport %q not yet implemented", cfg.Transport.Type)
    default:
        return fmt.Errorf("unknown transport: %s", cfg.Transport.Type)
    }

    fmt.Fprintf(os.Stderr, "Starting %s (transport=%s)\n", cfg.Server.Name, cfg.Transport.Type)
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
feat(cli): integrate Koanf config with serve command

- Load config file via --config flag
- CLI flags override config file values
- Precedence: defaults < config < env < CLI

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 5: Add Config Validate Command

**Files:**
- Create: `cmd/metatools-mcp/cmd/config.go`
- Test: `cmd/metatools-mcp/cmd/config_test.go`

**Step 1: Write failing test for config validate**

```go
// cmd/metatools-mcp/cmd/config_test.go
package cmd

import (
    "bytes"
    "os"
    "path/filepath"
    "testing"
)

func TestConfigValidateCmd(t *testing.T) {
    t.Run("valid config", func(t *testing.T) {
        dir := t.TempDir()
        configPath := filepath.Join(dir, "metatools.yaml")

        yaml := `
server:
  name: test
transport:
  type: stdio
`
        os.WriteFile(configPath, []byte(yaml), 0644)

        cmd := NewRootCmd()
        buf := new(bytes.Buffer)
        cmd.SetOut(buf)
        cmd.SetArgs([]string{"config", "validate", "--config", configPath})

        err := cmd.Execute()
        if err != nil {
            t.Fatalf("Execute() error = %v", err)
        }

        if !contains(buf.String(), "valid") {
            t.Errorf("Output should indicate config is valid, got: %s", buf.String())
        }
    })

    t.Run("invalid config", func(t *testing.T) {
        dir := t.TempDir()
        configPath := filepath.Join(dir, "metatools.yaml")

        yaml := `
transport:
  type: invalid_transport
`
        os.WriteFile(configPath, []byte(yaml), 0644)

        cmd := NewRootCmd()
        cmd.SetArgs([]string{"config", "validate", "--config", configPath})

        err := cmd.Execute()
        if err == nil {
            t.Error("Execute() should fail with invalid config")
        }
    })
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/jraymond/Documents/Projects/metatools-mcp && go test ./cmd/metatools-mcp/cmd/... -run TestConfigValidateCmd -v`
Expected: FAIL - config command doesn't exist

**Step 3: Implement config command**

```go
// cmd/metatools-mcp/cmd/config.go
package cmd

import (
    "fmt"

    "github.com/spf13/cobra"
    "github.com/your-org/metatools-mcp/internal/config"
)

func newConfigCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "config",
        Short: "Configuration management commands",
        Long:  "Commands for validating and inspecting configuration.",
    }

    cmd.AddCommand(newConfigValidateCmd())
    cmd.AddCommand(newConfigShowCmd())

    return cmd
}

func newConfigValidateCmd() *cobra.Command {
    var configPath string

    cmd := &cobra.Command{
        Use:   "validate",
        Short: "Validate configuration file",
        Long:  "Load and validate the configuration file, reporting any errors.",
        RunE: func(cmd *cobra.Command, args []string) error {
            cfg, err := config.Load(configPath)
            if err != nil {
                return fmt.Errorf("configuration invalid: %w", err)
            }

            fmt.Fprintln(cmd.OutOrStdout(), "Configuration is valid")
            fmt.Fprintf(cmd.OutOrStdout(), "  Server: %s\n", cfg.Server.Name)
            fmt.Fprintf(cmd.OutOrStdout(), "  Transport: %s\n", cfg.Transport.Type)
            fmt.Fprintf(cmd.OutOrStdout(), "  Search: %s\n", cfg.Search.Strategy)

            return nil
        },
    }

    cmd.Flags().StringVarP(&configPath, "config", "c", "", "Path to config file (required)")
    cmd.MarkFlagRequired("config")

    return cmd
}

func newConfigShowCmd() *cobra.Command {
    var configPath string

    cmd := &cobra.Command{
        Use:   "show",
        Short: "Show effective configuration",
        Long:  "Load configuration and display the effective values after all merging.",
        RunE: func(cmd *cobra.Command, args []string) error {
            cfg, err := config.Load(configPath)
            if err != nil {
                return err
            }

            // Pretty print config
            fmt.Fprintf(cmd.OutOrStdout(), "server:\n")
            fmt.Fprintf(cmd.OutOrStdout(), "  name: %s\n", cfg.Server.Name)
            fmt.Fprintf(cmd.OutOrStdout(), "  version: %s\n", cfg.Server.Version)
            fmt.Fprintf(cmd.OutOrStdout(), "\ntransport:\n")
            fmt.Fprintf(cmd.OutOrStdout(), "  type: %s\n", cfg.Transport.Type)
            fmt.Fprintf(cmd.OutOrStdout(), "  http:\n")
            fmt.Fprintf(cmd.OutOrStdout(), "    host: %s\n", cfg.Transport.HTTP.Host)
            fmt.Fprintf(cmd.OutOrStdout(), "    port: %d\n", cfg.Transport.HTTP.Port)
            fmt.Fprintf(cmd.OutOrStdout(), "\nsearch:\n")
            fmt.Fprintf(cmd.OutOrStdout(), "  strategy: %s\n", cfg.Search.Strategy)
            fmt.Fprintf(cmd.OutOrStdout(), "\nexecution:\n")
            fmt.Fprintf(cmd.OutOrStdout(), "  timeout: %s\n", cfg.Execution.Timeout)
            fmt.Fprintf(cmd.OutOrStdout(), "  max_tool_calls: %d\n", cfg.Execution.MaxToolCalls)

            return nil
        },
    }

    cmd.Flags().StringVarP(&configPath, "config", "c", "", "Path to config file")

    return cmd
}
```

**Step 4: Add config command to root**

Update `root.go`:
```go
func NewRootCmd() *cobra.Command {
    // ... existing code ...

    rootCmd.AddCommand(newServeCmd())
    rootCmd.AddCommand(newVersionCmd())
    rootCmd.AddCommand(newConfigCmd())  // Add this line

    return rootCmd
}
```

**Step 5: Run test to verify it passes**

Run: `cd /Users/jraymond/Documents/Projects/metatools-mcp && go test ./cmd/metatools-mcp/cmd/... -run TestConfigValidateCmd -v`
Expected: PASS

**Step 6: Commit**

```bash
git add cmd/metatools-mcp/cmd/config.go cmd/metatools-mcp/cmd/config_test.go cmd/metatools-mcp/cmd/root.go
git commit -m "$(cat <<'EOF'
feat(cli): add config validate and show subcommands

- 'config validate' loads and validates config file
- 'config show' displays effective configuration after merging

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 6: Create Example Config File

**Files:**
- Create: `metatools.example.yaml`

**Step 1: Create example config**

```yaml
# metatools.example.yaml
# Example configuration for metatools-mcp
# Copy to metatools.yaml and customize

server:
  name: metatools-mcp
  version: "0.2.0"

# Transport configuration
transport:
  type: stdio  # stdio | sse | http

  # HTTP settings (used when type is sse or http)
  http:
    host: "0.0.0.0"
    port: 8080
    tls:
      enabled: false
      cert: /etc/ssl/cert.pem
      key: /etc/ssl/key.pem

# Search strategy configuration
search:
  strategy: bm25  # bm25 | lexical | semantic

  # BM25-specific settings
  bm25:
    name_boost: 3.0
    namespace_boost: 2.0
    tags_boost: 2.0
    max_docs: 0        # 0 = unlimited
    max_doctext_len: 0  # 0 = unlimited

# Execution settings
execution:
  timeout: 30s
  max_tool_calls: 64
  max_chain_steps: 8

# Tool provider settings
providers:
  search_tools:
    enabled: true
  describe_tool:
    enabled: true
  run_tool:
    enabled: true
  run_chain:
    enabled: true
  execute_code:
    enabled: false  # Requires toolruntime build tag
    sandbox: dev    # dev | standard | hardened

# Backend settings
backends:
  local:
    enabled: true
    paths:
      - ~/.config/metatools/tools
      - /usr/share/metatools/tools
    watch: false  # Hot reload on changes

# Environment variable overrides:
# METATOOLS_TRANSPORT_TYPE=sse
# METATOOLS_TRANSPORT_HTTP_PORT=9090
# METATOOLS_SEARCH_STRATEGY=semantic
# METATOOLS_EXECUTION_TIMEOUT=60s
```

**Step 2: Commit**

```bash
git add metatools.example.yaml
git commit -m "$(cat <<'EOF'
docs: add example configuration file

Complete example showing all configuration options with documentation.

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Verification Checklist

- [ ] Koanf dependency added
- [ ] Config struct with all fields
- [ ] DefaultConfig() with sensible defaults
- [ ] Validate() with error messages
- [ ] Load() from file + env
- [ ] LoadWithOverrides() for CLI integration
- [ ] Serve command uses config
- [ ] Config validate command works
- [ ] Config show command works
- [ ] Example config file

## Definition of Done

1. All tests pass: `go test ./internal/config/... && go test ./cmd/metatools-mcp/cmd/...`
2. `metatools-mcp config validate --config=metatools.yaml` validates config
3. `metatools-mcp config show` displays effective config
4. `metatools-mcp serve --config=metatools.yaml` uses config file
5. Environment variables override config file
6. CLI flags override environment variables

## Next PRD

PRD-004 will implement the SSE Transport Layer, enabling HTTP/SSE mode for web clients.
