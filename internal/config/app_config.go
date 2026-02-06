// Package config defines application configuration models.
package config

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jonwraymond/metatools-mcp/internal/middleware"
)

// AppConfig holds all metatools-mcp configuration loaded from files/env/flags.
type AppConfig struct {
	Server        ServerConfig        `koanf:"server"`
	Transport     TransportConfig     `koanf:"transport"`
	Search        AppSearchConfig     `koanf:"search"`
	Execution     ExecutionConfig     `koanf:"execution"`
	Providers     ProvidersConfig     `koanf:"providers"`
	Backends      BackendsConfig      `koanf:"backends"`
	Secrets       SecretsConfig       `koanf:"secrets"`
	State         StateConfig         `koanf:"state"`
	Middleware    middleware.Config   `koanf:"middleware"`
	Toolsets      []ToolsetConfig     `koanf:"toolsets"`
	Skills        []SkillConfig       `koanf:"skills"`
	SkillDefaults SkillDefaultsConfig `koanf:"skill_defaults"`
	Health        HealthConfig        `koanf:"health"`
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
	Strategy string               `koanf:"strategy"`
	BM25     BM25Config           `koanf:"bm25"`
	Semantic SemanticSearchConfig `koanf:"semantic"`
}

// BM25Config holds BM25 search settings.
type BM25Config struct {
	NameBoost      int `koanf:"name_boost"`
	NamespaceBoost int `koanf:"namespace_boost"`
	TagsBoost      int `koanf:"tags_boost"`
	MaxDocs        int `koanf:"max_docs"`
	MaxDocTextLen  int `koanf:"max_doctext_len"`
}

// SemanticSearchConfig configures semantic or hybrid search.
type SemanticSearchConfig struct {
	Embedder string         `koanf:"embedder"`
	Config   map[string]any `koanf:"config"`
	Weight   float64        `koanf:"weight"`
}

// ExecutionConfig holds tool execution settings.
type ExecutionConfig struct {
	Timeout       time.Duration `koanf:"timeout"`
	MaxToolCalls  int           `koanf:"max_tool_calls"`
	MaxChainSteps int           `koanf:"max_chain_steps"`
}

// StateConfig holds persistent runtime configuration.
type StateConfig struct {
	RuntimeLimitsDB string `koanf:"runtime_limits_db"`
}

// SecretsConfig configures secret providers and resolution behavior.
type SecretsConfig struct {
	Strict    bool                          `koanf:"strict"`
	Providers map[string]SecretProviderConfig `koanf:"providers"`
}

// SecretProviderConfig configures a single secret provider instance.
type SecretProviderConfig struct {
	Enabled bool           `koanf:"enabled"`
	Config  map[string]any `koanf:"config"`
}

// ProvidersConfig holds tool provider settings.
type ProvidersConfig struct {
	SearchTools      ProviderEnabled   `koanf:"search_tools"`
	ListTools        ProviderEnabled   `koanf:"list_tools"`
	ListNamespaces   ProviderEnabled   `koanf:"list_namespaces"`
	DescribeTool     ProviderEnabled   `koanf:"describe_tool"`
	ListToolExamples ProviderEnabled   `koanf:"list_tool_examples"`
	RunTool          ProviderEnabled   `koanf:"run_tool"`
	RunChain         ProviderEnabled   `koanf:"run_chain"`
	ExecuteCode      ExecuteCodeConfig `koanf:"execute_code"`
	ListToolsets     ProviderEnabled   `koanf:"list_toolsets"`
	DescribeToolset  ProviderEnabled   `koanf:"describe_toolset"`
	ListSkills       ProviderEnabled   `koanf:"list_skills"`
	DescribeSkill    ProviderEnabled   `koanf:"describe_skill"`
	PlanSkill        ProviderEnabled   `koanf:"plan_skill"`
	RunSkill         ProviderEnabled   `koanf:"run_skill"`
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

// ToolsetConfig defines a configurable toolset.
type ToolsetConfig struct {
	Name             string   `koanf:"name"`
	Description      string   `koanf:"description"`
	NamespaceFilters []string `koanf:"namespace_filters"`
	TagFilters       []string `koanf:"tag_filters"`
	AllowIDs         []string `koanf:"allow_ids"`
	DenyIDs          []string `koanf:"deny_ids"`
	Policy           string   `koanf:"policy"`
	Exposure         string   `koanf:"exposure"`
}

// SkillStepConfig defines a skill step.
type SkillStepConfig struct {
	ID     string         `koanf:"id"`
	ToolID string         `koanf:"tool_id"`
	Inputs map[string]any `koanf:"inputs"`
}

// SkillGuardsConfig defines skill guard settings.
type SkillGuardsConfig struct {
	MaxSteps int      `koanf:"max_steps"`
	AllowIDs []string `koanf:"allow_ids"`
}

// SkillConfig defines a skill.
type SkillConfig struct {
	Name        string            `koanf:"name"`
	Description string            `koanf:"description"`
	ToolsetID   string            `koanf:"toolset_id"`
	Steps       []SkillStepConfig `koanf:"steps"`
	Guards      SkillGuardsConfig `koanf:"guards"`
}

// SkillDefaultsConfig defines default skill limits.
type SkillDefaultsConfig struct {
	MaxSteps     int           `koanf:"max_steps"`
	MaxToolCalls int           `koanf:"max_tool_calls"`
	Timeout      time.Duration `koanf:"timeout"`
}

// HealthConfig defines health endpoint settings.
type HealthConfig struct {
	Enabled bool   `koanf:"enabled"`
	Path    string `koanf:"http_path"`
}

// BackendsConfig holds backend source settings.
type BackendsConfig struct {
	Local LocalBackendConfig `koanf:"local"`
	MCP   []MCPBackendConfig `koanf:"mcp"`
	// MCPRefresh controls periodic refresh behavior for MCP backends.
	MCPRefresh MCPRefreshConfig `koanf:"mcp_refresh"`
}

// LocalBackendConfig holds local tool backend settings.
type LocalBackendConfig struct {
	Enabled bool     `koanf:"enabled"`
	Paths   []string `koanf:"paths"`
	Watch   bool     `koanf:"watch"`
}

// MCPBackendConfig holds remote MCP backend settings.
type MCPBackendConfig struct {
	Name       string            `koanf:"name"`
	URL        string            `koanf:"url"`
	Headers    map[string]string `koanf:"headers"`
	MaxRetries int               `koanf:"max_retries"`
}

// MCPRefreshConfig controls periodic refresh behavior for MCP backends.
type MCPRefreshConfig struct {
	Interval   time.Duration `koanf:"interval"`
	Jitter     time.Duration `koanf:"jitter"`
	StaleAfter time.Duration `koanf:"stale_after"`
	OnDemand   bool          `koanf:"on_demand"`
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
	"hybrid":   true,
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
			Semantic: SemanticSearchConfig{
				Embedder: "",
				Config:   map[string]any{},
				Weight:   0.5,
			},
		},
		Execution: ExecutionConfig{
			Timeout:       30 * time.Second,
			MaxToolCalls:  64,
			MaxChainSteps: 8,
		},
		Providers: ProvidersConfig{
			SearchTools:      ProviderEnabled{Enabled: true},
			ListTools:        ProviderEnabled{Enabled: true},
			ListNamespaces:   ProviderEnabled{Enabled: true},
			DescribeTool:     ProviderEnabled{Enabled: true},
			ListToolExamples: ProviderEnabled{Enabled: true},
			RunTool:          ProviderEnabled{Enabled: true},
			RunChain:         ProviderEnabled{Enabled: true},
			ExecuteCode:      ExecuteCodeConfig{Enabled: false, Sandbox: "dev"},
			ListToolsets:     ProviderEnabled{Enabled: true},
			DescribeToolset:  ProviderEnabled{Enabled: true},
			ListSkills:       ProviderEnabled{Enabled: true},
			DescribeSkill:    ProviderEnabled{Enabled: true},
			PlanSkill:        ProviderEnabled{Enabled: true},
			RunSkill:         ProviderEnabled{Enabled: true},
		},
		Backends: BackendsConfig{
			Local: LocalBackendConfig{
				Enabled: true,
				Paths:   []string{},
				Watch:   false,
			},
			MCP: nil,
			MCPRefresh: MCPRefreshConfig{
				Interval:   10 * time.Minute,
				Jitter:     30 * time.Second,
				StaleAfter: 15 * time.Minute,
				OnDemand:   true,
			},
		},
		Secrets: SecretsConfig{
			Strict:    true,
			Providers: map[string]SecretProviderConfig{},
		},
		State: StateConfig{
			RuntimeLimitsDB: "",
		},
		Middleware: middleware.Config{},
		SkillDefaults: SkillDefaultsConfig{
			MaxSteps:     16,
			MaxToolCalls: 64,
			Timeout:      30 * time.Second,
		},
		Health: HealthConfig{
			Enabled: false,
			Path:    "/healthz",
		},
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
		return fmt.Errorf("invalid search strategy %q, must be one of: bm25, lexical, semantic, hybrid", c.Search.Strategy)
	}

	if c.Execution.Timeout < 0 {
		return errors.New("execution timeout cannot be negative")
	}
	if c.Execution.MaxToolCalls < 0 {
		return errors.New("execution max tool calls cannot be negative")
	}
	if c.Execution.MaxChainSteps < 0 {
		return errors.New("execution max chain steps cannot be negative")
	}

	if c.SkillDefaults.MaxSteps < 0 {
		return errors.New("skill defaults max steps cannot be negative")
	}
	if c.SkillDefaults.MaxToolCalls < 0 {
		return errors.New("skill defaults max tool calls cannot be negative")
	}
	if c.SkillDefaults.Timeout < 0 {
		return errors.New("skill defaults timeout cannot be negative")
	}

	if c.Backends.MCPRefresh.Interval < 0 {
		return errors.New("mcp refresh interval cannot be negative")
	}
	if c.Backends.MCPRefresh.Jitter < 0 {
		return errors.New("mcp refresh jitter cannot be negative")
	}
	if c.Backends.MCPRefresh.StaleAfter < 0 {
		return errors.New("mcp refresh stale_after cannot be negative")
	}

	seenBackendNames := make(map[string]struct{}, len(c.Backends.MCP))
	for _, backend := range c.Backends.MCP {
		name := strings.TrimSpace(backend.Name)
		if name == "" {
			return errors.New("mcp backend name is required")
		}
		if strings.TrimSpace(backend.URL) == "" {
			return fmt.Errorf("mcp backend %q url is required", name)
		}
		if _, exists := seenBackendNames[name]; exists {
			return fmt.Errorf("duplicate mcp backend name %q", name)
		}
		seenBackendNames[name] = struct{}{}
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
		SemanticEmbedder:   c.Semantic.Embedder,
		SemanticConfig:     c.Semantic.Config,
		SemanticWeight:     c.Semantic.Weight,
	}
}
