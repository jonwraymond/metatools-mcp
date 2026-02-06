package skills

import (
	"fmt"
	"sort"
	"strings"

	"github.com/jonwraymond/metatools-mcp/internal/toolset"
	"github.com/jonwraymond/toolcompose/skill"
)

const skillIDPrefix = "skill:"

// Skill represents a configured skill definition with guards.
type Skill struct {
	ID          string
	Name        string
	Description string
	ToolsetID   string
	Steps       []skill.Step
	Guards      []skill.Guard
}

// GuardSpec defines guard configuration.
type GuardSpec struct {
	MaxSteps int
	AllowIDs []string
}

// StepSpec defines a skill step.
type StepSpec struct {
	ID     string
	ToolID string
	Inputs map[string]any
}

// Spec defines a skill configuration.
type Spec struct {
	Name        string
	Description string
	ToolsetID   string
	Steps       []StepSpec
	Guards      GuardSpec
}

// Registry stores skills keyed by ID.
type Registry struct {
	skills map[string]*Skill
	order  []string
}

// NewRegistry constructs a registry from skills.
func NewRegistry(skills []*Skill) *Registry {
	reg := &Registry{
		skills: make(map[string]*Skill, len(skills)),
	}
	for _, s := range skills {
		if s == nil || s.ID == "" {
			continue
		}
		reg.skills[s.ID] = s
		reg.order = append(reg.order, s.ID)
	}
	sort.Strings(reg.order)
	return reg
}

// List returns skills in deterministic order.
func (r *Registry) List() []*Skill {
	if r == nil {
		return nil
	}
	out := make([]*Skill, 0, len(r.order))
	for _, id := range r.order {
		if s, ok := r.skills[id]; ok {
			out = append(out, s)
		}
	}
	return out
}

// Get returns a skill by ID.
func (r *Registry) Get(id string) (*Skill, bool) {
	if r == nil {
		return nil, false
	}
	s, ok := r.skills[id]
	return s, ok
}

// BuildRegistry builds a skill registry from specs.
func BuildRegistry(toolsets *toolset.Registry, specs []Spec) (*Registry, error) {
	if len(specs) == 0 {
		return NewRegistry(nil), nil
	}
	if toolsets == nil {
		toolsets = toolset.NewRegistry(nil)
	}

	out := make([]*Skill, 0, len(specs))
	seen := make(map[string]struct{}, len(specs))

	for _, cfg := range specs {
		if strings.TrimSpace(cfg.Name) == "" {
			return nil, fmt.Errorf("skill name is required")
		}
		id := skillID(cfg.Name)
		if _, exists := seen[id]; exists {
			return nil, fmt.Errorf("duplicate skill id %q", id)
		}
		seen[id] = struct{}{}

		var toolsetIDs []string
		if cfg.ToolsetID != "" {
			ts, ok := toolsets.Get(cfg.ToolsetID)
			if !ok {
				return nil, fmt.Errorf("skill %q references unknown toolset %q", cfg.Name, cfg.ToolsetID)
			}
			toolsetIDs = ts.ToolIDs()
		}

		steps := make([]skill.Step, len(cfg.Steps))
		for i, stepCfg := range cfg.Steps {
			step := skill.Step{
				ID:     stepCfg.ID,
				ToolID: stepCfg.ToolID,
				Inputs: stepCfg.Inputs,
			}
			if err := step.Validate(); err != nil {
				return nil, fmt.Errorf("skill %q step %d invalid: %w", cfg.Name, i, err)
			}
			steps[i] = step
		}

		allowedIDs := resolveAllowedIDs(cfg.Guards.AllowIDs, toolsetIDs)
		guards := make([]skill.Guard, 0, 2)
		if cfg.Guards.MaxSteps > 0 {
			guards = append(guards, skill.MaxStepsGuard(cfg.Guards.MaxSteps))
		}
		if len(allowedIDs) > 0 {
			guards = append(guards, skill.AllowedToolIDsGuard(allowedIDs))
		}

		out = append(out, &Skill{
			ID:          id,
			Name:        cfg.Name,
			Description: cfg.Description,
			ToolsetID:   cfg.ToolsetID,
			Steps:       steps,
			Guards:      guards,
		})
	}

	return NewRegistry(out), nil
}

func skillID(name string) string {
	return skillIDPrefix + slugify(name)
}

func resolveAllowedIDs(allowIDs, toolsetIDs []string) []string {
	if len(toolsetIDs) == 0 {
		return dedupeStrings(allowIDs)
	}
	if len(allowIDs) == 0 {
		return dedupeStrings(toolsetIDs)
	}
	allowSet := make(map[string]struct{}, len(allowIDs))
	for _, id := range allowIDs {
		allowSet[id] = struct{}{}
	}
	var out []string
	for _, id := range toolsetIDs {
		if _, ok := allowSet[id]; ok {
			out = append(out, id)
		}
	}
	return dedupeStrings(out)
}

func dedupeStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, v := range values {
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	sort.Strings(out)
	return out
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
		return "skill"
	}
	return out
}
