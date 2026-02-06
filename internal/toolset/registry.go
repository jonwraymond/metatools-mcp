package toolset

import (
	"fmt"
	"sort"
	"strings"

	toolset "github.com/jonwraymond/toolcompose/set"
	"github.com/jonwraymond/tooldiscovery/index"
	"github.com/jonwraymond/toolfoundation/adapter"
)

const (
	toolsetIDPrefix = "toolset:"
	pageSize        = 200
)

// Toolset wraps a composed toolset with metadata.
type Toolset struct {
	ID          string
	Name        string
	Description string
	Tools       []*adapter.CanonicalTool
}

// Spec defines a toolset configuration.
type Spec struct {
	Name             string
	Description      string
	NamespaceFilters []string
	TagFilters       []string
	AllowIDs         []string
	DenyIDs          []string
	Policy           string
}

// ToolIDs returns tool IDs sorted lexicographically.
func (t *Toolset) ToolIDs() []string {
	ids := make([]string, 0, len(t.Tools))
	for _, tool := range t.Tools {
		if tool == nil {
			continue
		}
		ids = append(ids, tool.ID())
	}
	sort.Strings(ids)
	return ids
}

// Registry stores toolsets keyed by ID.
type Registry struct {
	sets  map[string]*Toolset
	order []string
}

// NewRegistry constructs a registry from toolsets.
func NewRegistry(toolsets []*Toolset) *Registry {
	reg := &Registry{
		sets: make(map[string]*Toolset, len(toolsets)),
	}
	for _, ts := range toolsets {
		if ts == nil || ts.ID == "" {
			continue
		}
		reg.sets[ts.ID] = ts
		reg.order = append(reg.order, ts.ID)
	}
	sort.Strings(reg.order)
	return reg
}

// List returns toolsets in deterministic order.
func (r *Registry) List() []*Toolset {
	if r == nil {
		return nil
	}
	out := make([]*Toolset, 0, len(r.order))
	for _, id := range r.order {
		if ts, ok := r.sets[id]; ok {
			out = append(out, ts)
		}
	}
	return out
}

// Get returns a toolset by ID.
func (r *Registry) Get(id string) (*Toolset, bool) {
	if r == nil {
		return nil, false
	}
	ts, ok := r.sets[id]
	return ts, ok
}

// BuildRegistry builds a toolset registry from specs.
func BuildRegistry(idx index.Index, specs []Spec) (*Registry, error) {
	if len(specs) == 0 {
		return NewRegistry(nil), nil
	}
	if idx == nil {
		return nil, fmt.Errorf("toolset registry requires an index")
	}

	canonicalTools, err := listCanonicalTools(idx)
	if err != nil {
		return nil, err
	}

	toolsets := make([]*Toolset, 0, len(specs))
	seen := make(map[string]struct{}, len(specs))

	for _, cfg := range specs {
		if strings.TrimSpace(cfg.Name) == "" {
			return nil, fmt.Errorf("toolset name is required")
		}
		id := toolsetID(cfg.Name)
		if _, exists := seen[id]; exists {
			return nil, fmt.Errorf("duplicate toolset id %q", id)
		}
		seen[id] = struct{}{}

		builder := toolset.NewBuilder(cfg.Name).FromTools(canonicalTools)
		if len(cfg.NamespaceFilters) > 0 {
			builder.WithNamespaces(cfg.NamespaceFilters)
		}
		if len(cfg.TagFilters) > 0 {
			builder.WithTags(cfg.TagFilters)
		}
		if len(cfg.AllowIDs) > 0 {
			builder.WithTools(cfg.AllowIDs)
		}
		if len(cfg.DenyIDs) > 0 {
			builder.ExcludeTools(cfg.DenyIDs)
		}
		if strings.TrimSpace(cfg.Policy) != "" {
			policy, err := policyFromConfig(cfg.Policy)
			if err != nil {
				return nil, err
			}
			builder.WithPolicy(policy)
		}

		ts, err := builder.Build()
		if err != nil {
			return nil, err
		}

		toolsets = append(toolsets, &Toolset{
			ID:          id,
			Name:        cfg.Name,
			Description: cfg.Description,
			Tools:       ts.Tools(),
		})
	}

	return NewRegistry(toolsets), nil
}

func listCanonicalTools(idx index.Index) ([]*adapter.CanonicalTool, error) {
	ids, err := listToolIDs(idx)
	if err != nil {
		return nil, err
	}
	mcpAdapter := adapter.NewMCPAdapter()
	tools := make([]*adapter.CanonicalTool, 0, len(ids))
	for _, id := range ids {
		tool, _, err := idx.GetTool(id)
		if err != nil {
			return nil, err
		}
		ct, err := mcpAdapter.ToCanonical(tool)
		if err != nil {
			return nil, err
		}
		tools = append(tools, ct)
	}
	return tools, nil
}

func listToolIDs(idx index.Index) ([]string, error) {
	cursor := ""
	ids := make([]string, 0)
	seen := make(map[string]struct{})

	for {
		results, next, err := idx.SearchPage("", pageSize, cursor)
		if err != nil {
			return nil, err
		}
		for _, summary := range results {
			if summary.ID == "" {
				continue
			}
			if _, ok := seen[summary.ID]; ok {
				continue
			}
			seen[summary.ID] = struct{}{}
			ids = append(ids, summary.ID)
		}
		if next == "" {
			break
		}
		cursor = next
	}

	sort.Strings(ids)
	return ids, nil
}

func toolsetID(name string) string {
	return toolsetIDPrefix + slugify(name)
}

func policyFromConfig(raw string) (toolset.Policy, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "allow_all":
		return toolset.AllowAll(), nil
	case "deny_all":
		return toolset.DenyAll(), nil
	case "":
		return nil, nil
	default:
		return nil, fmt.Errorf("unknown toolset policy %q", raw)
	}
}

func slugify(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	var b strings.Builder
	prevDash := false
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			prevDash = false
			continue
		}
		if !prevDash {
			b.WriteByte('-')
			prevDash = true
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "toolset"
	}
	return out
}
