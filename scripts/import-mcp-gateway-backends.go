package main

import (
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type gatewayConfig struct {
	Backends []gatewayBackend `yaml:"backends"`
}

type gatewayBackend struct {
	Name      string            `yaml:"name"`
	Transport string            `yaml:"transport"`
	URL       string            `yaml:"url"`
	Headers   map[string]string `yaml:"headers"`
	Command   string            `yaml:"command"`
	Args      []string          `yaml:"args"`
	Timeout   string            `yaml:"timeout"`
}

type outputConfig struct {
	Secrets  outputSecrets  `yaml:"secrets"`
	Backends outputBackends `yaml:"backends"`
}

type outputSecrets struct {
	Strict    bool                     `yaml:"strict"`
	Providers map[string]outputProvider `yaml:"providers"`
}

type outputProvider struct {
	Enabled bool           `yaml:"enabled"`
	Config  map[string]any `yaml:"config"`
}

type outputBackends struct {
	MCP []outputMCPBackend `yaml:"mcp"`
}

type outputMCPBackend struct {
	Name       string            `yaml:"name"`
	URL        string            `yaml:"url"`
	Headers    map[string]string `yaml:"headers,omitempty"`
	MaxRetries int               `yaml:"max_retries"`
}

func main() {
	in := flag.String("in", "/Users/jraymond/Documents/Projects/mcp-gateway/config.multi-proxy.yaml", "Path to mcp-gateway multi-proxy config")
	out := flag.String("out", "examples/backends.mcp-gateway.local.yaml", "Output path for metatools-mcp backend config example (local-only)")
	flag.Parse()

	b, err := os.ReadFile(*in)
	if err != nil {
		fatalf("read input: %v", err)
	}

	rendered, err := importBackendsYAML(b)
	if err != nil {
		fatalf("import: %v", err)
	}

	if err := os.WriteFile(*out, rendered, 0o644); err != nil {
		fatalf("write output: %v", err)
	}
}

func fatalf(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

func importBackendsYAML(input []byte) ([]byte, error) {
	var cfg gatewayConfig
	if err := yaml.Unmarshal(input, &cfg); err != nil {
		return nil, fmt.Errorf("parse yaml: %w", err)
	}

	remote := make([]outputMCPBackend, 0, len(cfg.Backends))
	for _, b := range cfg.Backends {
		if !isRemoteHTTPOrSSEBackend(b) {
			continue
		}
		remote = append(remote, outputMCPBackend{
			Name:       b.Name,
			URL:        b.URL,
			Headers:    cloneHeaders(b.Headers),
			MaxRetries: 5,
		})
	}

	sort.Slice(remote, func(i, j int) bool { return remote[i].Name < remote[j].Name })

	out := outputConfig{
		Secrets: outputSecrets{
			Strict: true,
			Providers: map[string]outputProvider{
				"bws": {
					Enabled: true,
					Config: map[string]any{
						"access_token":    "${BWS_ACCESS_TOKEN}",
						"organization_id": "${BWS_ORG_ID}",
						"cache_ttl":       "10m",
					},
				},
			},
		},
		Backends: outputBackends{MCP: remote},
	}

	node, err := stableYAML(out)
	if err != nil {
		return nil, err
	}
	return renderYAML(node)
}

func isRemoteHTTPOrSSEBackend(b gatewayBackend) bool {
	if strings.TrimSpace(b.URL) == "" {
		return false
	}
	if strings.TrimSpace(b.Command) != "" {
		return false
	}

	// Prefer explicit transport when present.
	if strings.TrimSpace(b.Transport) == "streamable-http" {
		return true
	}

	parsed, err := url.Parse(b.URL)
	if err != nil {
		return false
	}
	switch parsed.Scheme {
	case "http", "https", "sse":
		return true
	default:
		return false
	}
}

func cloneHeaders(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		if strings.TrimSpace(k) == "" {
			continue
		}
		out[k] = v
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func stableYAML(out outputConfig) (*yaml.Node, error) {
	// Encode deterministically using yaml.Node so map key order doesn't churn diffs.
	var root yaml.Node
	if err := root.Encode(out); err != nil {
		return nil, fmt.Errorf("encode yaml: %w", err)
	}
	if root.Kind == 0 {
		return nil, errors.New("unexpected yaml encoding")
	}
	sortMappingNodeKeys(&root)
	return &root, nil
}

func sortMappingNodeKeys(n *yaml.Node) {
	if n == nil {
		return
	}
	switch n.Kind {
	case yaml.MappingNode:
		// Content is [k1, v1, k2, v2, ...]. Sort by key value.
		type kv struct{ k, v *yaml.Node }
		pairs := make([]kv, 0, len(n.Content)/2)
		for i := 0; i+1 < len(n.Content); i += 2 {
			pairs = append(pairs, kv{k: n.Content[i], v: n.Content[i+1]})
		}
		sort.SliceStable(pairs, func(i, j int) bool { return pairs[i].k.Value < pairs[j].k.Value })
		n.Content = n.Content[:0]
		for _, p := range pairs {
			n.Content = append(n.Content, p.k, p.v)
		}
		for _, p := range pairs {
			sortMappingNodeKeys(p.v)
		}
	case yaml.SequenceNode, yaml.DocumentNode:
		for _, c := range n.Content {
			sortMappingNodeKeys(c)
		}
	default:
		// Scalars: nothing.
	}
}

func renderYAML(doc *yaml.Node) ([]byte, error) {
	if doc == nil {
		return nil, errors.New("yaml doc is nil")
	}
	b, err := yaml.Marshal(doc)
	if err != nil {
		return nil, fmt.Errorf("marshal yaml: %w", err)
	}
	return b, nil
}
