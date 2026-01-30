package config

import (
	"errors"
	"fmt"
	"time"

	"github.com/jonwraymond/metatools-mcp/internal/middleware"
)

// AppConfig holds all metatools-mcp configuration loaded from files/env/flags.
type AppConfig struct {
	Server     ServerConfig      `koanf:"server"`
	Transport  TransportConfig   `koanf:"transport"`
	Search     AppSearchConfig   `koanf:"search"`
	Execution  ExecutionConfig   `koanf:"execution"`
	Providers  ProvidersConfig   `koanf:"providers"`
	Backends   BackendsConfig    `koanf:"backends"`
	Middleware middleware.Config `koanf:"middleware"`
}

// ServerConfig holds server identity settings.
type ServerConfig struct {
	Name    string `koanf:"name"`
	Version string `koanf:"version"`
}

// TransportConfig holds transport layer settings.
type TransportConfig struct {
	Type       string           `koanf:"type"`
	HTTP       HTTPConfig       `koanf:"http"`
	Streamable StreamableConfig `koanf:"streamable"`
}

// StreamableConfig holds Streamable HTTP transport settings.
type StreamableConfig struct {
	Stateless      bool          `koanf:"stateless"`
	JSONResponse   bool          `koanf:"json_response"`
	SessionTimeout time.Duration `koanf:"session_timeout"`
}

// HTTPConfig holds HTTP transport settings.
type HTTPConfig struct {
	Host string    `koanf:"host"`
	Port int       `koanf:"port"`
	TLS  TLSConfig `koanf:"tls"`
}

// TLSConfig holds TLS settings.
type TLSConfig struct {
	Enabled  bool   `koanf:"enabled"`
	CertFile string `koanf:"cert"`
	KeyFile  string `koanf:"key"`
}

// AppSearchConfig holds search strategy settings.
type AppSearchConfig struct {
	Strategy string     `koanf:"strategy"`
	BM25     BM25Config `koanf:"bm25"`
}

// BM25Config holds BM25 search settings.
type BM25Config struct {
	NameBoost      int `koanf:"name_boost"`
	NamespaceBoost int `koanf:"namespace_boost"`
	TagsBoost      int `koanf:"tags_boost"`
	MaxDocs        int `koanf:"max_docs"`
	MaxDocTextLen  int `koanf:"max_doctext_len"`
}

// ExecutionConfig holds tool execution settings.
type ExecutionConfig struct {
	Timeout       time.Duration `koanf:"timeout"`
	MaxToolCalls  int           `koanf:"max_tool_calls"`
	MaxChainSteps int           `koanf:"max_chain_steps"`
}

// ProvidersConfig holds tool provider settings.
type ProvidersConfig struct {
	SearchTools      ProviderEnabled   `koanf:"search_tools"`
	ListNamespaces   ProviderEnabled   `koanf:"list_namespaces"`
	DescribeTool     ProviderEnabled   `koanf:"describe_tool"`
	ListToolExamples ProviderEnabled   `koanf:"list_tool_examples"`
	RunTool          ProviderEnabled   `koanf:"run_tool"`
	RunChain         ProviderEnabled   `koanf:"run_chain"`
	ExecuteCode      ExecuteCodeConfig `koanf:"execute_code"`
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
	Local LocalBackendConfig `koanf:"local"`
}

// LocalBackendConfig holds local tool backend settings.
type LocalBackendConfig struct {
	Enabled bool     `koanf:"enabled"`
	Paths   []string `koanf:"paths"`
	Watch   bool     `koanf:"watch"`
}

var validAppTransports = map[string]bool{
	"stdio":      true,
	"sse":        true,
	"streamable": true,
}

var validAppSearchStrategies = map[string]bool{
	"bm25":     true,
	"lexical":  true,
	"semantic": true,
}

// DefaultAppConfig returns the default configuration.
func DefaultAppConfig() AppConfig {
	return AppConfig{
		Server: ServerConfig{
			Name:    "metatools-mcp",
			Version: "dev",
		},
		Transport: TransportConfig{
			Type: "stdio",
			HTTP: HTTPConfig{
				Host: "0.0.0.0",
				Port: 8080,
			},
			Streamable: StreamableConfig{
				Stateless:      false,
				JSONResponse:   false,
				SessionTimeout: 30 * time.Minute,
			},
		},
		Search: AppSearchConfig{
			Strategy: "lexical",
			BM25: BM25Config{
				NameBoost:      3,
				NamespaceBoost: 2,
				TagsBoost:      2,
				MaxDocs:        0,
				MaxDocTextLen:  0,
			},
		},
		Execution: ExecutionConfig{
			Timeout:       30 * time.Second,
			MaxToolCalls:  64,
			MaxChainSteps: 8,
		},
		Providers: ProvidersConfig{
			SearchTools:      ProviderEnabled{Enabled: true},
			ListNamespaces:   ProviderEnabled{Enabled: true},
			DescribeTool:     ProviderEnabled{Enabled: true},
			ListToolExamples: ProviderEnabled{Enabled: true},
			RunTool:          ProviderEnabled{Enabled: true},
			RunChain:         ProviderEnabled{Enabled: true},
			ExecuteCode:      ExecuteCodeConfig{Enabled: false, Sandbox: "dev"},
		},
		Backends: BackendsConfig{
			Local: LocalBackendConfig{
				Enabled: true,
				Paths:   []string{},
				Watch:   false,
			},
		},
		Middleware: middleware.Config{},
	}
}

// Validate checks the configuration for errors.
func (c *AppConfig) Validate() error {
	if !validAppTransports[c.Transport.Type] {
		return fmt.Errorf("invalid transport type %q, must be one of: stdio, sse, streamable", c.Transport.Type)
	}

	if c.Transport.Type != "stdio" {
		if c.Transport.HTTP.Port <= 0 || c.Transport.HTTP.Port > 65535 {
			return fmt.Errorf("invalid port %d, must be 1-65535", c.Transport.HTTP.Port)
		}
	}

	if !validAppSearchStrategies[c.Search.Strategy] {
		return fmt.Errorf("invalid search strategy %q, must be one of: bm25, lexical, semantic", c.Search.Strategy)
	}

	if c.Execution.Timeout < 0 {
		return errors.New("execution timeout cannot be negative")
	}

	return nil
}

// ToSearchConfig converts AppSearchConfig to the runtime SearchConfig used by bootstrap.
func (c AppSearchConfig) ToSearchConfig() SearchConfig {
	return SearchConfig{
		Strategy:           c.Strategy,
		BM25NameBoost:      c.BM25.NameBoost,
		BM25NamespaceBoost: c.BM25.NamespaceBoost,
		BM25TagsBoost:      c.BM25.TagsBoost,
		BM25MaxDocs:        c.BM25.MaxDocs,
		BM25MaxDocTextLen:  c.BM25.MaxDocTextLen,
	}
}
