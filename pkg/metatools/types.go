package metatools

import "errors"

// ToolSummary represents a minimal tool summary for search results
type ToolSummary struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	Namespace        string   `json:"namespace,omitempty"`
	ShortDescription string   `json:"shortDescription,omitempty"`
	Tags             []string `json:"tags,omitempty"`
}

// ErrorObject is the structured error returned in metatool responses
type ErrorObject struct {
	Code        string         `json:"code"`
	Message     string         `json:"message"`
	ToolID      string         `json:"tool_id,omitempty"`
	Op          *string        `json:"op,omitempty"`
	BackendKind *string        `json:"backend_kind,omitempty"`
	StepIndex   *int           `json:"step_index,omitempty"`
	Retryable   bool           `json:"retryable,omitempty"`
	Details     map[string]any `json:"details,omitempty"`
}

// BackendOverride specifies a specific backend to use
type BackendOverride struct {
	Kind       string `json:"kind"` // local, provider, mcp
	ServerName string `json:"serverName,omitempty"`
	ProviderID string `json:"providerId,omitempty"`
	ToolID     string `json:"toolId,omitempty"`
	Name       string `json:"name,omitempty"`
}

// SearchToolsInput is the input for search_tools
type SearchToolsInput struct {
	Query  string  `json:"query"`
	Limit  *int    `json:"limit,omitempty"`
	Cursor *string `json:"cursor,omitempty"`
}

// Validate checks that the input is valid
func (s *SearchToolsInput) Validate() error {
	return nil
}

// GetLimit returns the effective limit, applying defaults and caps
func (s *SearchToolsInput) GetLimit() int {
	if s.Limit == nil {
		return 20 // default
	}
	if *s.Limit > 100 {
		return 100 // max cap
	}
	if *s.Limit < 1 {
		return 1
	}
	return *s.Limit
}

// SearchToolsOutput is the output for search_tools
type SearchToolsOutput struct {
	Tools      []ToolSummary `json:"tools"`
	NextCursor *string       `json:"nextCursor,omitempty"`
}

// ListToolsInput is the input for list_tools
type ListToolsInput struct {
	Limit       *int    `json:"limit,omitempty"`
	Cursor      *string `json:"cursor,omitempty"`
	BackendKind *string `json:"backend_kind,omitempty"` // local, provider, mcp
	BackendName *string `json:"backend_name,omitempty"` // server name (mcp) or provider id
}

// Validate checks that the input is valid.
func (l *ListToolsInput) Validate() error {
	if l.BackendKind == nil {
		return nil
	}
	switch *l.BackendKind {
	case "local", "provider", "mcp":
		return nil
	default:
		return errors.New("backend_kind must be one of: local, provider, mcp")
	}
}

// GetLimit returns the effective limit, applying defaults and caps.
func (l *ListToolsInput) GetLimit() int {
	if l.Limit == nil {
		return 20
	}
	if *l.Limit > 100 {
		return 100
	}
	if *l.Limit < 1 {
		return 1
	}
	return *l.Limit
}

// ListToolsOutput is the output for list_tools.
type ListToolsOutput struct {
	Tools      []ToolSummary `json:"tools"`
	NextCursor *string       `json:"nextCursor,omitempty"`
}

// ListNamespacesInput is the input for list_namespaces
type ListNamespacesInput struct {
	Limit  *int    `json:"limit,omitempty"`
	Cursor *string `json:"cursor,omitempty"`
}

// Validate checks that the input is valid
func (l *ListNamespacesInput) Validate() error {
	return nil
}

// GetLimit returns the effective limit, applying defaults and caps
func (l *ListNamespacesInput) GetLimit() int {
	if l.Limit == nil {
		return 20
	}
	if *l.Limit > 100 {
		return 100
	}
	if *l.Limit < 1 {
		return 1
	}
	return *l.Limit
}

// ListNamespacesOutput is the output for list_namespaces
type ListNamespacesOutput struct {
	Namespaces []string `json:"namespaces"`
	NextCursor *string  `json:"nextCursor,omitempty"`
}

// DescribeToolInput is the input for describe_tool
type DescribeToolInput struct {
	ToolID      string `json:"tool_id"`
	DetailLevel string `json:"detail_level"` // summary, schema, full
	ExamplesMax *int   `json:"examples_max,omitempty"`
}

// Validate checks that the input is valid
func (d *DescribeToolInput) Validate() error {
	if d.ToolID == "" {
		return errors.New("tool_id is required")
	}
	switch d.DetailLevel {
	case "summary", "schema", "full":
		return nil
	default:
		return errors.New("detail_level must be one of: summary, schema, full")
	}
}

// DescribeToolOutput is the output for describe_tool
type DescribeToolOutput struct {
	Tool         any           `json:"tool,omitempty"`
	Summary      string        `json:"summary"`
	SchemaInfo   any           `json:"schemaInfo,omitempty"`
	Notes        *string       `json:"notes,omitempty"`
	Examples     []ToolExample `json:"examples,omitempty"`
	ExternalRefs []string      `json:"externalRefs,omitempty"`
}

// ToolExample represents a usage example for a tool
type ToolExample struct {
	ID          string         `json:"id,omitempty"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Args        map[string]any `json:"args"`
	ResultHint  string         `json:"resultHint,omitempty"`
}

// ListToolExamplesInput is the input for list_tool_examples
type ListToolExamplesInput struct {
	ToolID string `json:"tool_id"`
	Max    *int   `json:"max,omitempty"`
}

// Validate checks that the input is valid
func (l *ListToolExamplesInput) Validate() error {
	if l.ToolID == "" {
		return errors.New("tool_id is required")
	}
	if l.Max != nil && *l.Max < 1 {
		return errors.New("max must be >= 1")
	}
	return nil
}

// GetMax returns the effective max, applying defaults and caps
func (l *ListToolExamplesInput) GetMax() int {
	if l.Max == nil {
		return 5
	}
	if *l.Max > 5 {
		return 5
	}
	if *l.Max < 1 {
		return 1
	}
	return *l.Max
}

// ListToolExamplesOutput is the output for list_tool_examples
type ListToolExamplesOutput struct {
	Examples []ToolExample `json:"examples"`
}

// RunToolInput is the input for run_tool
type RunToolInput struct {
	ToolID           string           `json:"tool_id"`
	Args             map[string]any   `json:"args,omitempty"`
	Stream           bool             `json:"stream,omitempty"`
	IncludeTool      bool             `json:"include_tool,omitempty"`
	IncludeBackend   bool             `json:"include_backend,omitempty"`
	IncludeMCPResult bool             `json:"include_mcp_result,omitempty"`
	BackendOverride  *BackendOverride `json:"backend_override,omitempty"`
}

// Validate checks that the input is valid
func (r *RunToolInput) Validate() error {
	if r.ToolID == "" {
		return errors.New("tool_id is required")
	}
	return nil
}

// RunToolOutput is the output for run_tool
type RunToolOutput struct {
	Structured any          `json:"structured,omitempty"`
	Error      *ErrorObject `json:"error,omitempty"`
	Tool       any          `json:"tool,omitempty"`
	Backend    any          `json:"backend,omitempty"`
	MCPResult  any          `json:"mcpResult,omitempty"`
	DurationMs *int         `json:"durationMs,omitempty"`
}

// ToolsetSummary represents a minimal toolset summary.
type ToolsetSummary struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	ToolCount   int    `json:"toolCount"`
}

// ListToolsetsInput is the input for list_toolsets.
type ListToolsetsInput struct{}

// Validate checks that the input is valid.
func (l *ListToolsetsInput) Validate() error { return nil }

// ListToolsetsOutput is the output for list_toolsets.
type ListToolsetsOutput struct {
	Toolsets []ToolsetSummary `json:"toolsets"`
}

// DescribeToolsetInput is the input for describe_toolset.
type DescribeToolsetInput struct {
	ToolsetID string `json:"toolset_id"`
}

// Validate checks that the input is valid.
func (d *DescribeToolsetInput) Validate() error {
	if d.ToolsetID == "" {
		return errors.New("toolset_id is required")
	}
	return nil
}

// ToolsetDetail is the detailed view of a toolset.
type ToolsetDetail struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description,omitempty"`
	Tools       []ToolSummary `json:"tools,omitempty"`
}

// DescribeToolsetOutput is the output for describe_toolset.
type DescribeToolsetOutput struct {
	Toolset ToolsetDetail `json:"toolset"`
}

// SkillStep defines a skill step.
type SkillStep struct {
	ID     string         `json:"id"`
	ToolID string         `json:"tool_id"`
	Inputs map[string]any `json:"inputs,omitempty"`
}

// SkillDefinition describes a skill.
type SkillDefinition struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	Steps       []SkillStep `json:"steps"`
	ToolsetID   string      `json:"toolset_id,omitempty"`
}

// SkillSummary represents a minimal skill summary.
type SkillSummary struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	StepCount   int    `json:"stepCount"`
	ToolsetID   string `json:"toolset_id,omitempty"`
}

// ListSkillsInput is the input for list_skills.
type ListSkillsInput struct{}

// Validate checks that the input is valid.
func (l *ListSkillsInput) Validate() error { return nil }

// ListSkillsOutput is the output for list_skills.
type ListSkillsOutput struct {
	Skills []SkillSummary `json:"skills"`
}

// DescribeSkillInput is the input for describe_skill.
type DescribeSkillInput struct {
	SkillID string `json:"skill_id"`
}

// Validate checks that the input is valid.
func (d *DescribeSkillInput) Validate() error {
	if d.SkillID == "" {
		return errors.New("skill_id is required")
	}
	return nil
}

// DescribeSkillOutput is the output for describe_skill.
type DescribeSkillOutput struct {
	Skill SkillDefinition `json:"skill"`
}

// PlanSkillInput is the input for plan_skill.
type PlanSkillInput struct {
	SkillID string           `json:"skill_id,omitempty"`
	Skill   *SkillDefinition `json:"skill,omitempty"`
}

// Validate checks that the input is valid.
func (p *PlanSkillInput) Validate() error {
	if p.SkillID == "" && p.Skill == nil {
		return errors.New("skill_id or skill is required")
	}
	if p.SkillID != "" && p.Skill != nil {
		return errors.New("skill_id and skill cannot both be set")
	}
	if p.Skill != nil && p.Skill.Name == "" {
		return errors.New("skill.name is required")
	}
	return nil
}

// SkillPlan is the deterministic plan returned by plan_skill.
type SkillPlan struct {
	Name  string      `json:"name"`
	Steps []SkillStep `json:"steps"`
}

// PlanSkillOutput is the output for plan_skill.
type PlanSkillOutput struct {
	Plan SkillPlan `json:"plan"`
}

// RunSkillInput is the input for run_skill.
type RunSkillInput struct {
	SkillID      string           `json:"skill_id,omitempty"`
	Skill        *SkillDefinition `json:"skill,omitempty"`
	MaxSteps     *int             `json:"max_steps,omitempty"`
	MaxToolCalls *int             `json:"max_tool_calls,omitempty"`
	TimeoutMs    *int             `json:"timeout_ms,omitempty"`
}

// Validate checks that the input is valid.
func (r *RunSkillInput) Validate() error {
	if r.SkillID == "" && r.Skill == nil {
		return errors.New("skill_id or skill is required")
	}
	if r.SkillID != "" && r.Skill != nil {
		return errors.New("skill_id and skill cannot both be set")
	}
	if r.Skill != nil && r.Skill.Name == "" {
		return errors.New("skill.name is required")
	}
	return nil
}

// SkillStepResult is the output for a skill step.
type SkillStepResult struct {
	StepID string       `json:"step_id"`
	Value  any          `json:"value,omitempty"`
	Error  *ErrorObject `json:"error,omitempty"`
}

// RunSkillOutput is the output for run_skill.
type RunSkillOutput struct {
	Results    []SkillStepResult `json:"results,omitempty"`
	Error      *ErrorObject      `json:"error,omitempty"`
	DurationMs *int              `json:"durationMs,omitempty"`
}

// ChainStep represents a single step in a chain
type ChainStep struct {
	ToolID      string         `json:"tool_id"`
	Args        map[string]any `json:"args,omitempty"`
	UsePrevious bool           `json:"use_previous,omitempty"`
}

// RunChainInput is the input for run_chain
type RunChainInput struct {
	Steps           []ChainStep `json:"steps"`
	IncludeBackends *bool       `json:"include_backends,omitempty"`
	IncludeTools    *bool       `json:"include_tools,omitempty"`
}

// Validate checks that the input is valid
func (r *RunChainInput) Validate() error {
	if len(r.Steps) == 0 {
		return errors.New("steps must not be empty")
	}
	for i, step := range r.Steps {
		if step.ToolID == "" {
			return errors.New("step " + string(rune('0'+i)) + " missing tool_id")
		}
	}
	return nil
}

// GetIncludeBackends returns the effective include_backends value.
// Default: true.
func (r *RunChainInput) GetIncludeBackends() bool {
	if r.IncludeBackends == nil {
		return true
	}
	return *r.IncludeBackends
}

// GetIncludeTools returns the effective include_tools value.
// Default: false.
func (r *RunChainInput) GetIncludeTools() bool {
	if r.IncludeTools == nil {
		return false
	}
	return *r.IncludeTools
}

// ChainStepResult represents the result of a single chain step
type ChainStepResult struct {
	ToolID     string       `json:"tool_id"`
	Structured any          `json:"structured,omitempty"`
	Backend    any          `json:"backend,omitempty"`
	Tool       any          `json:"tool,omitempty"`
	Error      *ErrorObject `json:"error,omitempty"`
}

// RunChainOutput is the output for run_chain
type RunChainOutput struct {
	Results []ChainStepResult `json:"results"`
	Final   any               `json:"final,omitempty"`
	Error   *ErrorObject      `json:"error,omitempty"`
}

// ExecuteCodeInput is the input for execute_code
type ExecuteCodeInput struct {
	Language     string `json:"language"`
	Code         string `json:"code"`
	TimeoutMs    *int   `json:"timeout_ms,omitempty"`
	MaxToolCalls *int   `json:"max_tool_calls,omitempty"`
}

// ExecuteCodeOutput is the output for execute_code
type ExecuteCodeOutput struct {
	Value      any    `json:"value,omitempty"`
	Stdout     string `json:"stdout,omitempty"`
	Stderr     string `json:"stderr,omitempty"`
	DurationMs int    `json:"durationMs"`
}
