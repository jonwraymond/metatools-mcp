# PRD-014: toolskill Library Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement agent skills support with SKILL.md parsing, skill registry, composition DSL, and runtime execution.

**Architecture:** The toolskill library provides higher-level agent behaviors that compose tools into reusable workflows. Skills sit above tools in the abstraction hierarchy - while tools are atomic operations, skills orchestrate multiple tools with workflow logic (research-and-summarize, debug-and-fix, deploy-with-validation).

**Tech Stack:** Go 1.21+, yaml.v3 (frontmatter), goldmark (markdown), semver

**Dependencies:** toolset (tool access), toolrun (execution), toolobserve (optional tracing)

---

## Task 1: Core Types and Skill Interface

**Files:**
- Create: `toolskill/skill.go`
- Create: `toolskill/manifest.go`
- Test: `toolskill/skill_test.go`

**Step 1: Write the failing test**

```go
// toolskill/skill_test.go
package toolskill

import (
    "context"
    "testing"

    "github.com/Masterminds/semver/v3"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestSkillManifest_Validate(t *testing.T) {
    tests := []struct {
        name    string
        manifest SkillManifest
        wantErr bool
    }{
        {
            name: "valid manifest",
            manifest: SkillManifest{
                ID:          "research-tool",
                Name:        "research-tool",
                Version:     "1.0.0",
                Description: "Use when researching tools in the codebase",
                Tags:        []string{"research", "discovery"},
            },
            wantErr: false,
        },
        {
            name: "missing ID",
            manifest: SkillManifest{
                Name:        "research-tool",
                Version:     "1.0.0",
                Description: "Use when researching",
            },
            wantErr: true,
        },
        {
            name: "missing name",
            manifest: SkillManifest{
                ID:          "research-tool",
                Version:     "1.0.0",
                Description: "Use when researching",
            },
            wantErr: true,
        },
        {
            name: "invalid version",
            manifest: SkillManifest{
                ID:          "research-tool",
                Name:        "research-tool",
                Version:     "not-semver",
                Description: "Use when researching",
            },
            wantErr: true,
        },
        {
            name: "description must start with Use when",
            manifest: SkillManifest{
                ID:          "research-tool",
                Name:        "research-tool",
                Version:     "1.0.0",
                Description: "Researches tools", // Missing "Use when"
            },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.manifest.Validate()
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}

func TestSkillInput_Get(t *testing.T) {
    input := SkillInput{
        "topic":  "kubernetes",
        "depth":  3,
        "tags":   []string{"container", "orchestration"},
    }

    // String retrieval
    topic, ok := input.GetString("topic")
    assert.True(t, ok)
    assert.Equal(t, "kubernetes", topic)

    // Int retrieval
    depth, ok := input.GetInt("depth")
    assert.True(t, ok)
    assert.Equal(t, 3, depth)

    // Missing key
    _, ok = input.GetString("missing")
    assert.False(t, ok)
}

func TestSkillOutput_Success(t *testing.T) {
    output := NewSkillOutput(map[string]any{
        "summary": "Found 5 tools",
        "tools":   []string{"tool1", "tool2"},
    })

    assert.True(t, output.Success)
    assert.Nil(t, output.Error)
    assert.Equal(t, "Found 5 tools", output.Result["summary"])
}

func TestSkillOutput_Failure(t *testing.T) {
    output := NewSkillOutputError(ErrSkillExecutionFailed, "step 3 timed out")

    assert.False(t, output.Success)
    assert.NotNil(t, output.Error)
    assert.Contains(t, output.ErrorMessage, "step 3 timed out")
}

// Mock skill for testing
type mockSkill struct {
    id       string
    name     string
    version  *semver.Version
    manifest *SkillManifest
    tools    []string
    steps    []StepDefinition
}

func (m *mockSkill) ID() string                    { return m.id }
func (m *mockSkill) Name() string                  { return m.name }
func (m *mockSkill) Version() *semver.Version      { return m.version }
func (m *mockSkill) Manifest() *SkillManifest      { return m.manifest }
func (m *mockSkill) RequiredTools() []string       { return m.tools }
func (m *mockSkill) Steps() []StepDefinition       { return m.steps }
func (m *mockSkill) Execute(ctx context.Context, input SkillInput) (*SkillOutput, error) {
    return NewSkillOutput(map[string]any{"executed": true}), nil
}

func TestSkillInterface(t *testing.T) {
    ver, _ := semver.NewVersion("1.0.0")
    skill := &mockSkill{
        id:      "test-skill",
        name:    "test-skill",
        version: ver,
        manifest: &SkillManifest{
            ID:          "test-skill",
            Name:        "test-skill",
            Version:     "1.0.0",
            Description: "Use when testing skill interface",
        },
        tools: []string{"search_tools", "describe_tool"},
        steps: []StepDefinition{
            {ID: "step1", Name: "Search", Tool: "search_tools"},
        },
    }

    // Verify interface compliance
    var _ Skill = skill

    assert.Equal(t, "test-skill", skill.ID())
    assert.Equal(t, "test-skill", skill.Name())
    assert.Equal(t, "1.0.0", skill.Version().String())
    assert.Len(t, skill.RequiredTools(), 2)
    assert.Len(t, skill.Steps(), 1)

    // Execute
    output, err := skill.Execute(context.Background(), SkillInput{"query": "test"})
    require.NoError(t, err)
    assert.True(t, output.Success)
}
```

**Step 2: Run test to verify it fails**

Run: `cd toolskill && go test -v -run TestSkill`
Expected: FAIL with "package toolskill is not in std"

**Step 3: Write minimal implementation**

```go
// toolskill/skill.go
package toolskill

import (
    "context"
    "errors"
    "fmt"
    "strings"

    "github.com/Masterminds/semver/v3"
)

// Errors
var (
    ErrSkillNotFound        = errors.New("skill not found")
    ErrSkillExecutionFailed = errors.New("skill execution failed")
    ErrInvalidManifest      = errors.New("invalid skill manifest")
    ErrMissingTool          = errors.New("required tool not available")
    ErrStepFailed           = errors.New("step execution failed")
)

// Skill defines the interface for reusable agent behaviors.
// Skills orchestrate multiple tools into workflows.
type Skill interface {
    // Identity
    ID() string
    Name() string
    Version() *semver.Version

    // Manifest for discovery/advertisement
    Manifest() *SkillManifest

    // Execution
    Execute(ctx context.Context, input SkillInput) (*SkillOutput, error)

    // Introspection
    RequiredTools() []string
    Steps() []StepDefinition
}

// SkillInput represents input parameters for skill execution.
type SkillInput map[string]any

// GetString retrieves a string value from input.
func (i SkillInput) GetString(key string) (string, bool) {
    v, ok := i[key]
    if !ok {
        return "", false
    }
    s, ok := v.(string)
    return s, ok
}

// GetInt retrieves an int value from input.
func (i SkillInput) GetInt(key string) (int, bool) {
    v, ok := i[key]
    if !ok {
        return 0, false
    }
    switch n := v.(type) {
    case int:
        return n, true
    case int64:
        return int(n), true
    case float64:
        return int(n), true
    default:
        return 0, false
    }
}

// GetBool retrieves a bool value from input.
func (i SkillInput) GetBool(key string) (bool, bool) {
    v, ok := i[key]
    if !ok {
        return false, false
    }
    b, ok := v.(bool)
    return b, ok
}

// SkillOutput represents the result of skill execution.
type SkillOutput struct {
    Success      bool           `json:"success"`
    Result       map[string]any `json:"result,omitempty"`
    Error        error          `json:"-"`
    ErrorMessage string         `json:"error,omitempty"`

    // Execution metadata
    StepsCompleted []string `json:"stepsCompleted,omitempty"`
    Duration       int64    `json:"durationMs,omitempty"`
}

// NewSkillOutput creates a successful output.
func NewSkillOutput(result map[string]any) *SkillOutput {
    return &SkillOutput{
        Success: true,
        Result:  result,
    }
}

// NewSkillOutputError creates a failed output.
func NewSkillOutputError(err error, message string) *SkillOutput {
    return &SkillOutput{
        Success:      false,
        Error:        err,
        ErrorMessage: message,
    }
}

// StepDefinition defines a single step within a skill.
type StepDefinition struct {
    ID           string                           `json:"id"`
    Name         string                           `json:"name"`
    Tool         string                           `json:"tool"`
    InputMapper  func(SkillContext) any           `json:"-"`
    OutputMapper func(any) any                    `json:"-"`
    Condition    func(SkillContext) bool          `json:"-"`
    OnError      ErrorHandler                     `json:"-"`
    Timeout      int64                            `json:"timeoutMs,omitempty"`
    Optional     bool                             `json:"optional,omitempty"`
}

// ErrorHandler defines how to handle step errors.
type ErrorHandler func(ctx SkillContext, err error) error

// SkillContext provides execution context to steps.
type SkillContext struct {
    Input    SkillInput     // Original skill input
    Results  map[string]any // Results from previous steps
    Metadata map[string]any // Execution metadata
    StepID   string         // Current step ID
}

// Get retrieves a value from results.
func (c *SkillContext) Get(key string) (any, bool) {
    v, ok := c.Results[key]
    return v, ok
}

// Set stores a value in results.
func (c *SkillContext) Set(key string, value any) {
    if c.Results == nil {
        c.Results = make(map[string]any)
    }
    c.Results[key] = value
}
```

```go
// toolskill/manifest.go
package toolskill

import (
    "encoding/json"
    "fmt"
    "strings"

    "github.com/Masterminds/semver/v3"
)

// SkillManifest describes skill capabilities for discovery.
// Aligned with A2A (Agent-to-Agent) protocol.
type SkillManifest struct {
    ID          string         `json:"id" yaml:"id"`
    Name        string         `json:"name" yaml:"name"`
    Version     string         `json:"version" yaml:"version"`
    Description string         `json:"description" yaml:"description"`
    InputSchema map[string]any `json:"inputSchema,omitempty" yaml:"inputSchema,omitempty"`
    OutputSchema map[string]any `json:"outputSchema,omitempty" yaml:"outputSchema,omitempty"`
    Tags        []string       `json:"tags,omitempty" yaml:"tags,omitempty"`

    // Dependencies
    RequiredTools  []string `json:"requiredTools,omitempty" yaml:"requiredTools,omitempty"`
    RequiredSkills []string `json:"requiredSkills,omitempty" yaml:"requiredSkills,omitempty"`

    // Execution hints
    EstimatedSteps int  `json:"estimatedSteps,omitempty" yaml:"estimatedSteps,omitempty"`
    Idempotent     bool `json:"idempotent,omitempty" yaml:"idempotent,omitempty"`
    SupportsPause  bool `json:"supportsPause,omitempty" yaml:"supportsPause,omitempty"`

    // Metadata
    Author  string         `json:"author,omitempty" yaml:"author,omitempty"`
    License string         `json:"license,omitempty" yaml:"license,omitempty"`
    Extra   map[string]any `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// Validate checks if the manifest is valid.
func (m *SkillManifest) Validate() error {
    if m.ID == "" {
        return fmt.Errorf("%w: ID is required", ErrInvalidManifest)
    }
    if m.Name == "" {
        return fmt.Errorf("%w: Name is required", ErrInvalidManifest)
    }
    if m.Version == "" {
        return fmt.Errorf("%w: Version is required", ErrInvalidManifest)
    }
    if _, err := semver.NewVersion(m.Version); err != nil {
        return fmt.Errorf("%w: invalid version %q: %v", ErrInvalidManifest, m.Version, err)
    }
    if m.Description == "" {
        return fmt.Errorf("%w: Description is required", ErrInvalidManifest)
    }
    // Per SKILL.md standard, description should start with "Use when"
    if !strings.HasPrefix(strings.ToLower(m.Description), "use when") {
        return fmt.Errorf("%w: Description must start with 'Use when' per SKILL.md standard", ErrInvalidManifest)
    }
    return nil
}

// SemVer returns the parsed semantic version.
func (m *SkillManifest) SemVer() (*semver.Version, error) {
    return semver.NewVersion(m.Version)
}

// ToJSON serializes the manifest to JSON.
func (m *SkillManifest) ToJSON() ([]byte, error) {
    return json.MarshalIndent(m, "", "  ")
}

// ManifestFromJSON deserializes a manifest from JSON.
func ManifestFromJSON(data []byte) (*SkillManifest, error) {
    var m SkillManifest
    if err := json.Unmarshal(data, &m); err != nil {
        return nil, fmt.Errorf("failed to parse manifest: %w", err)
    }
    return &m, nil
}
```

**Step 4: Run test to verify it passes**

Run: `cd toolskill && go test -v -run TestSkill`
Expected: PASS

**Step 5: Commit**

```bash
git add toolskill/
git commit -m "$(cat <<'EOF'
feat(toolskill): add core types and Skill interface

- Add Skill interface with ID, Name, Version, Manifest, Execute
- Add SkillManifest type with A2A-aligned fields
- Add SkillInput/SkillOutput types with helpers
- Add StepDefinition for workflow steps
- Add SkillContext for execution context
- Validate manifest per SKILL.md standard (description starts with "Use when")

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Task 2: SKILL.md Parser

**Files:**
- Create: `toolskill/skillmd/parser.go`
- Test: `toolskill/skillmd/parser_test.go`

**Step 1: Write the failing test**

```go
// toolskill/skillmd/parser_test.go
package skillmd

import (
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestParser_Parse(t *testing.T) {
    content := `---
name: research-tool
description: Use when researching tools in a codebase
metadata:
  author: metatools
  version: "1.0.0"
---

# Research Tool

## Overview

Discovers and analyzes tools in a codebase.

## When to Use

- When exploring unfamiliar codebases
- When looking for specific tool capabilities

## How It Works

1. Search for tools matching query
2. Get detailed documentation
3. Summarize findings
`

    parser := NewParser()
    skill, err := parser.ParseBytes([]byte(content))
    require.NoError(t, err)

    assert.Equal(t, "research-tool", skill.Name)
    assert.Equal(t, "Use when researching tools in a codebase", skill.Description)
    assert.Equal(t, "metatools", skill.Metadata["author"])
    assert.Equal(t, "1.0.0", skill.Metadata["version"])
    assert.Contains(t, skill.Overview, "Discovers and analyzes")
    assert.Contains(t, skill.WhenToUse, "exploring unfamiliar")
    assert.Contains(t, skill.HowItWorks, "Search for tools")
}

func TestParser_ParseMinimal(t *testing.T) {
    content := `---
name: simple-skill
description: Use when doing simple things
---

# Simple Skill

Basic skill content.
`

    parser := NewParser()
    skill, err := parser.ParseBytes([]byte(content))
    require.NoError(t, err)

    assert.Equal(t, "simple-skill", skill.Name)
    assert.Equal(t, "Use when doing simple things", skill.Description)
    assert.Empty(t, skill.Metadata)
}

func TestParser_ParseMissingFrontmatter(t *testing.T) {
    content := `# No Frontmatter

Just markdown content.
`

    parser := NewParser()
    _, err := parser.ParseBytes([]byte(content))
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "frontmatter")
}

func TestParser_ParseMissingName(t *testing.T) {
    content := `---
description: Use when testing
---

# Test
`

    parser := NewParser()
    _, err := parser.ParseBytes([]byte(content))
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "name")
}

func TestParser_ParseMissingDescription(t *testing.T) {
    content := `---
name: test-skill
---

# Test
`

    parser := NewParser()
    _, err := parser.ParseBytes([]byte(content))
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "description")
}

func TestParser_Validate(t *testing.T) {
    parser := NewParser()

    tests := []struct {
        name    string
        skill   *SkillMD
        wantErr bool
    }{
        {
            name: "valid skill",
            skill: &SkillMD{
                Name:        "valid-skill",
                Description: "Use when testing validation",
            },
            wantErr: false,
        },
        {
            name: "invalid name with spaces",
            skill: &SkillMD{
                Name:        "invalid skill name",
                Description: "Use when testing",
            },
            wantErr: true,
        },
        {
            name: "description not starting with Use when",
            skill: &SkillMD{
                Name:        "skill",
                Description: "Does something cool",
            },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            errs := parser.Validate(tt.skill)
            if tt.wantErr {
                assert.NotEmpty(t, errs)
            } else {
                assert.Empty(t, errs)
            }
        })
    }
}

func TestSkillMD_ToManifest(t *testing.T) {
    skill := &SkillMD{
        Name:        "test-skill",
        Description: "Use when testing manifest conversion",
        Metadata: map[string]any{
            "version": "2.0.0",
            "author":  "test",
        },
    }

    manifest := skill.ToManifest()
    assert.Equal(t, "test-skill", manifest.ID)
    assert.Equal(t, "test-skill", manifest.Name)
    assert.Equal(t, "2.0.0", manifest.Version)
    assert.Equal(t, "Use when testing manifest conversion", manifest.Description)
    assert.Equal(t, "test", manifest.Author)
}
```

**Step 2: Run test to verify it fails**

Run: `cd toolskill && go test -v ./skillmd/...`
Expected: FAIL with "package toolskill/skillmd is not in std"

**Step 3: Write minimal implementation**

```go
// toolskill/skillmd/parser.go
package skillmd

import (
    "bytes"
    "errors"
    "fmt"
    "os"
    "regexp"
    "strings"

    "gopkg.in/yaml.v3"

    "github.com/yourusername/toolskill"
)

// SkillMD represents a parsed SKILL.md file.
type SkillMD struct {
    // YAML Frontmatter (required)
    Name        string         `yaml:"name"`
    Description string         `yaml:"description"`
    License     string         `yaml:"license,omitempty"`
    Metadata    map[string]any `yaml:"metadata,omitempty"`

    // Parsed Content
    Content    string            // Raw markdown content (after frontmatter)
    Overview   string            // Extracted ## Overview section
    WhenToUse  string            // Extracted ## When to Use section
    HowItWorks string            // Extracted ## How It Works section
    Sections   map[string]string // All other sections
}

// ToManifest converts SkillMD to a SkillManifest.
func (s *SkillMD) ToManifest() *toolskill.SkillManifest {
    version := "1.0.0"
    author := ""

    if s.Metadata != nil {
        if v, ok := s.Metadata["version"].(string); ok {
            version = v
        }
        if a, ok := s.Metadata["author"].(string); ok {
            author = a
        }
    }

    return &toolskill.SkillManifest{
        ID:          s.Name,
        Name:        s.Name,
        Version:     version,
        Description: s.Description,
        Author:      author,
        License:     s.License,
        Extra:       s.Metadata,
    }
}

// ValidationError represents a validation issue.
type ValidationError struct {
    Field   string
    Message string
}

func (e ValidationError) Error() string {
    return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// Parser reads and parses SKILL.md files.
type Parser struct {
    // strictMode enforces all validations
    strictMode bool
}

// NewParser creates a new SKILL.md parser.
func NewParser() *Parser {
    return &Parser{strictMode: true}
}

// NewLenientParser creates a parser that allows some validation failures.
func NewLenientParser() *Parser {
    return &Parser{strictMode: false}
}

// Parse reads and parses a SKILL.md file from path.
func (p *Parser) Parse(path string) (*SkillMD, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("failed to read skill file: %w", err)
    }
    return p.ParseBytes(data)
}

// ParseBytes parses SKILL.md content from bytes.
func (p *Parser) ParseBytes(data []byte) (*SkillMD, error) {
    // Split frontmatter from content
    frontmatter, content, err := splitFrontmatter(data)
    if err != nil {
        return nil, err
    }

    // Parse YAML frontmatter
    var skill SkillMD
    if err := yaml.Unmarshal(frontmatter, &skill); err != nil {
        return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
    }

    // Validate required fields
    if skill.Name == "" {
        return nil, errors.New("skill name is required in frontmatter")
    }
    if skill.Description == "" {
        return nil, errors.New("skill description is required in frontmatter")
    }

    // Store raw content
    skill.Content = string(content)

    // Extract standard sections
    skill.Sections = extractSections(content)
    skill.Overview = skill.Sections["overview"]
    skill.WhenToUse = skill.Sections["when to use"]
    skill.HowItWorks = skill.Sections["how it works"]

    return &skill, nil
}

// Validate checks a SkillMD for issues.
func (p *Parser) Validate(skill *SkillMD) []ValidationError {
    var errs []ValidationError

    // Name must be lowercase with hyphens only
    namePattern := regexp.MustCompile(`^[a-z][a-z0-9-]*[a-z0-9]$|^[a-z]$`)
    if !namePattern.MatchString(skill.Name) {
        errs = append(errs, ValidationError{
            Field:   "name",
            Message: "must be lowercase with hyphens (e.g., 'my-skill')",
        })
    }

    // Description must start with "Use when"
    if !strings.HasPrefix(strings.ToLower(skill.Description), "use when") {
        errs = append(errs, ValidationError{
            Field:   "description",
            Message: "must start with 'Use when' per SKILL.md standard",
        })
    }

    return errs
}

// splitFrontmatter separates YAML frontmatter from markdown content.
func splitFrontmatter(data []byte) ([]byte, []byte, error) {
    const delimiter = "---"

    // Must start with ---
    if !bytes.HasPrefix(data, []byte(delimiter)) {
        return nil, nil, errors.New("SKILL.md must start with YAML frontmatter (---)")
    }

    // Find closing ---
    rest := data[len(delimiter):]
    idx := bytes.Index(rest, []byte("\n"+delimiter))
    if idx == -1 {
        return nil, nil, errors.New("SKILL.md frontmatter not closed (missing ---)")
    }

    frontmatter := rest[:idx]
    content := rest[idx+len(delimiter)+1:] // Skip newline + ---

    // Skip leading newlines from content
    content = bytes.TrimLeft(content, "\n")

    return frontmatter, content, nil
}

// extractSections parses markdown content into named sections.
func extractSections(content []byte) map[string]string {
    sections := make(map[string]string)
    lines := strings.Split(string(content), "\n")

    var currentSection string
    var currentContent strings.Builder

    for _, line := range lines {
        // Check for ## heading
        if strings.HasPrefix(line, "## ") {
            // Save previous section
            if currentSection != "" {
                sections[strings.ToLower(currentSection)] = strings.TrimSpace(currentContent.String())
            }

            // Start new section
            currentSection = strings.TrimPrefix(line, "## ")
            currentContent.Reset()
        } else if currentSection != "" {
            currentContent.WriteString(line)
            currentContent.WriteString("\n")
        }
    }

    // Save last section
    if currentSection != "" {
        sections[strings.ToLower(currentSection)] = strings.TrimSpace(currentContent.String())
    }

    return sections
}
```

**Step 4: Run test to verify it passes**

Run: `cd toolskill && go test -v ./skillmd/...`
Expected: PASS

**Step 5: Commit**

```bash
git add toolskill/skillmd/
git commit -m "$(cat <<'EOF'
feat(toolskill): add SKILL.md parser

- Add SkillMD type representing parsed SKILL.md files
- Add Parser with Parse and ParseBytes methods
- Extract YAML frontmatter (name, description, metadata)
- Extract standard sections (Overview, When to Use, How It Works)
- Add validation for SKILL.md standard compliance
- Add ToManifest conversion helper

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Task 3: Skill Registry

**Files:**
- Create: `toolskill/registry.go`
- Test: `toolskill/registry_test.go`

**Step 1: Write the failing test**

```go
// toolskill/registry_test.go
package toolskill

import (
    "testing"

    "github.com/Masterminds/semver/v3"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func newTestSkill(id, name, version string, tags []string) *mockSkill {
    ver, _ := semver.NewVersion(version)
    return &mockSkill{
        id:      id,
        name:    name,
        version: ver,
        manifest: &SkillManifest{
            ID:          id,
            Name:        name,
            Version:     version,
            Description: "Use when testing " + name,
            Tags:        tags,
        },
    }
}

func TestRegistry_Register(t *testing.T) {
    reg := NewRegistry()

    skill := newTestSkill("research-v1", "research", "1.0.0", []string{"discovery"})

    err := reg.Register(skill)
    require.NoError(t, err)

    // Verify registration
    got, err := reg.Get("research-v1")
    require.NoError(t, err)
    assert.Equal(t, "research-v1", got.ID())
}

func TestRegistry_RegisterDuplicate(t *testing.T) {
    reg := NewRegistry()

    skill := newTestSkill("research-v1", "research", "1.0.0", nil)

    err := reg.Register(skill)
    require.NoError(t, err)

    // Duplicate registration should fail
    err = reg.Register(skill)
    assert.Error(t, err)
    assert.ErrorIs(t, err, ErrSkillAlreadyRegistered)
}

func TestRegistry_Get(t *testing.T) {
    reg := NewRegistry()
    skill := newTestSkill("my-skill", "my-skill", "1.0.0", nil)
    _ = reg.Register(skill)

    got, err := reg.Get("my-skill")
    require.NoError(t, err)
    assert.Equal(t, "my-skill", got.ID())
}

func TestRegistry_GetNotFound(t *testing.T) {
    reg := NewRegistry()

    _, err := reg.Get("nonexistent")
    assert.Error(t, err)
    assert.ErrorIs(t, err, ErrSkillNotFound)
}

func TestRegistry_GetByName(t *testing.T) {
    reg := NewRegistry()

    // Register multiple versions
    v1 := newTestSkill("research-v1", "research", "1.0.0", nil)
    v2 := newTestSkill("research-v2", "research", "2.0.0", nil)
    v3 := newTestSkill("research-v3", "research", "3.0.0", nil)

    _ = reg.Register(v1)
    _ = reg.Register(v2)
    _ = reg.Register(v3)

    // Get latest (no constraint)
    got, err := reg.GetByName("research", "")
    require.NoError(t, err)
    assert.Equal(t, "3.0.0", got.Version().String())

    // Get with constraint
    got, err = reg.GetByName("research", ">=1.0.0 <2.0.0")
    require.NoError(t, err)
    assert.Equal(t, "1.0.0", got.Version().String())

    // Get with ^2 constraint
    got, err = reg.GetByName("research", "^2.0.0")
    require.NoError(t, err)
    assert.Equal(t, "2.0.0", got.Version().String())
}

func TestRegistry_GetByNameNotFound(t *testing.T) {
    reg := NewRegistry()

    _, err := reg.GetByName("nonexistent", "")
    assert.Error(t, err)
    assert.ErrorIs(t, err, ErrSkillNotFound)
}

func TestRegistry_List(t *testing.T) {
    reg := NewRegistry()

    s1 := newTestSkill("skill-a", "skill-a", "1.0.0", nil)
    s2 := newTestSkill("skill-b", "skill-b", "1.0.0", nil)

    _ = reg.Register(s1)
    _ = reg.Register(s2)

    manifests := reg.List()
    assert.Len(t, manifests, 2)

    ids := make([]string, len(manifests))
    for i, m := range manifests {
        ids[i] = m.ID
    }
    assert.Contains(t, ids, "skill-a")
    assert.Contains(t, ids, "skill-b")
}

func TestRegistry_ListByTag(t *testing.T) {
    reg := NewRegistry()

    s1 := newTestSkill("search-skill", "search-skill", "1.0.0", []string{"discovery", "search"})
    s2 := newTestSkill("code-skill", "code-skill", "1.0.0", []string{"code", "execution"})
    s3 := newTestSkill("explore-skill", "explore-skill", "1.0.0", []string{"discovery", "exploration"})

    _ = reg.Register(s1)
    _ = reg.Register(s2)
    _ = reg.Register(s3)

    // Filter by tag
    results := reg.ListByTag("discovery")
    assert.Len(t, results, 2)

    results = reg.ListByTag("code")
    assert.Len(t, results, 1)
    assert.Equal(t, "code-skill", results[0].ID)

    results = reg.ListByTag("nonexistent")
    assert.Len(t, results, 0)
}

func TestRegistry_Unregister(t *testing.T) {
    reg := NewRegistry()

    skill := newTestSkill("to-remove", "to-remove", "1.0.0", nil)
    _ = reg.Register(skill)

    err := reg.Unregister("to-remove")
    require.NoError(t, err)

    _, err = reg.Get("to-remove")
    assert.ErrorIs(t, err, ErrSkillNotFound)
}

func TestRegistry_UnregisterNotFound(t *testing.T) {
    reg := NewRegistry()

    err := reg.Unregister("nonexistent")
    assert.ErrorIs(t, err, ErrSkillNotFound)
}

func TestRegistry_Advertise(t *testing.T) {
    reg := NewRegistry()

    s1 := newTestSkill("public-skill", "public-skill", "1.0.0", nil)
    _ = reg.Register(s1)

    manifests := reg.Advertise()
    assert.Len(t, manifests, 1)
    assert.Equal(t, "public-skill", manifests[0].ID)
}
```

**Step 2: Run test to verify it fails**

Run: `cd toolskill && go test -v -run TestRegistry`
Expected: FAIL with "undefined: NewRegistry"

**Step 3: Write minimal implementation**

```go
// toolskill/registry.go
package toolskill

import (
    "errors"
    "fmt"
    "sort"
    "sync"

    "github.com/Masterminds/semver/v3"
)

var (
    ErrSkillAlreadyRegistered = errors.New("skill already registered")
)

// SkillRegistry manages skill discovery and lookup.
type SkillRegistry interface {
    Register(skill Skill) error
    Get(id string) (Skill, error)
    GetByName(name string, versionConstraint string) (Skill, error)
    List() []SkillManifest
    ListByTag(tag string) []SkillManifest
    Unregister(id string) error

    // Discovery for A2A
    Advertise() []SkillManifest
}

// Registry is the default implementation of SkillRegistry.
type Registry struct {
    mu     sync.RWMutex
    skills map[string]Skill // id -> skill
    byName map[string][]Skill // name -> skills (multiple versions)
}

// NewRegistry creates a new skill registry.
func NewRegistry() *Registry {
    return &Registry{
        skills: make(map[string]Skill),
        byName: make(map[string][]Skill),
    }
}

// Register adds a skill to the registry.
func (r *Registry) Register(skill Skill) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    id := skill.ID()
    if _, exists := r.skills[id]; exists {
        return fmt.Errorf("%w: %s", ErrSkillAlreadyRegistered, id)
    }

    r.skills[id] = skill
    r.byName[skill.Name()] = append(r.byName[skill.Name()], skill)

    return nil
}

// Get retrieves a skill by ID.
func (r *Registry) Get(id string) (Skill, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()

    skill, ok := r.skills[id]
    if !ok {
        return nil, fmt.Errorf("%w: %s", ErrSkillNotFound, id)
    }
    return skill, nil
}

// GetByName retrieves a skill by name with optional version constraint.
// If constraint is empty, returns the latest version.
func (r *Registry) GetByName(name string, versionConstraint string) (Skill, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()

    skills, ok := r.byName[name]
    if !ok || len(skills) == 0 {
        return nil, fmt.Errorf("%w: no skill named %q", ErrSkillNotFound, name)
    }

    // No constraint - return latest
    if versionConstraint == "" {
        return r.latestVersion(skills), nil
    }

    // Parse constraint
    constraint, err := semver.NewConstraint(versionConstraint)
    if err != nil {
        return nil, fmt.Errorf("invalid version constraint: %w", err)
    }

    // Find matching versions
    var matching []Skill
    for _, s := range skills {
        if constraint.Check(s.Version()) {
            matching = append(matching, s)
        }
    }

    if len(matching) == 0 {
        return nil, fmt.Errorf("%w: no skill %q matching constraint %q", ErrSkillNotFound, name, versionConstraint)
    }

    return r.latestVersion(matching), nil
}

// latestVersion returns the skill with highest version.
func (r *Registry) latestVersion(skills []Skill) Skill {
    if len(skills) == 1 {
        return skills[0]
    }

    sorted := make([]Skill, len(skills))
    copy(sorted, skills)

    sort.Slice(sorted, func(i, j int) bool {
        return sorted[i].Version().GreaterThan(sorted[j].Version())
    })

    return sorted[0]
}

// List returns all registered skill manifests.
func (r *Registry) List() []SkillManifest {
    r.mu.RLock()
    defer r.mu.RUnlock()

    manifests := make([]SkillManifest, 0, len(r.skills))
    for _, skill := range r.skills {
        manifests = append(manifests, *skill.Manifest())
    }

    // Sort by ID for deterministic order
    sort.Slice(manifests, func(i, j int) bool {
        return manifests[i].ID < manifests[j].ID
    })

    return manifests
}

// ListByTag returns skills that have the specified tag.
func (r *Registry) ListByTag(tag string) []SkillManifest {
    r.mu.RLock()
    defer r.mu.RUnlock()

    var manifests []SkillManifest
    for _, skill := range r.skills {
        m := skill.Manifest()
        for _, t := range m.Tags {
            if t == tag {
                manifests = append(manifests, *m)
                break
            }
        }
    }

    sort.Slice(manifests, func(i, j int) bool {
        return manifests[i].ID < manifests[j].ID
    })

    return manifests
}

// Unregister removes a skill from the registry.
func (r *Registry) Unregister(id string) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    skill, ok := r.skills[id]
    if !ok {
        return fmt.Errorf("%w: %s", ErrSkillNotFound, id)
    }

    delete(r.skills, id)

    // Remove from byName
    name := skill.Name()
    skills := r.byName[name]
    for i, s := range skills {
        if s.ID() == id {
            r.byName[name] = append(skills[:i], skills[i+1:]...)
            break
        }
    }
    if len(r.byName[name]) == 0 {
        delete(r.byName, name)
    }

    return nil
}

// Advertise returns manifests for A2A discovery.
func (r *Registry) Advertise() []SkillManifest {
    return r.List()
}
```

**Step 4: Run test to verify it passes**

Run: `cd toolskill && go test -v -run TestRegistry`
Expected: PASS

**Step 5: Commit**

```bash
git add toolskill/registry.go toolskill/registry_test.go
git commit -m "$(cat <<'EOF'
feat(toolskill): add skill registry with version resolution

- Add SkillRegistry interface for skill management
- Add Registry implementation with thread-safe operations
- Support version constraints via semver
- Add GetByName with constraint resolution (^2.0.0, >=1.0.0 <2.0.0)
- Add ListByTag for tag-based filtering
- Add Advertise for A2A protocol discovery

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Task 4: Skill Builder and Composition

**Files:**
- Create: `toolskill/builder.go`
- Test: `toolskill/builder_test.go`

**Step 1: Write the failing test**

```go
// toolskill/builder_test.go
package toolskill

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestBuilder_Basic(t *testing.T) {
    skill, err := NewBuilder("research-skill").
        WithVersion("1.0.0").
        WithDescription("Use when researching topics").
        WithTags("research", "discovery").
        Build()

    require.NoError(t, err)
    assert.Equal(t, "research-skill", skill.ID())
    assert.Equal(t, "research-skill", skill.Name())
    assert.Equal(t, "1.0.0", skill.Version().String())
    assert.Equal(t, "Use when researching topics", skill.Manifest().Description)
    assert.Contains(t, skill.Manifest().Tags, "research")
}

func TestBuilder_WithSteps(t *testing.T) {
    skill, err := NewBuilder("multi-step").
        WithVersion("1.0.0").
        WithDescription("Use when doing multi-step work").
        Step("search", "search_tools").
            WithInputMapper(func(ctx SkillContext) any {
                query, _ := ctx.Input.GetString("query")
                return map[string]any{"query": query}
            }).
            Done().
        Step("describe", "describe_tool").
            WithInputMapper(func(ctx SkillContext) any {
                // Get first result from search
                return map[string]any{"tool": "result"}
            }).
            Done().
        Build()

    require.NoError(t, err)
    assert.Len(t, skill.Steps(), 2)
    assert.Equal(t, "search", skill.Steps()[0].ID)
    assert.Equal(t, "describe", skill.Steps()[1].ID)
}

func TestBuilder_RequiredTools(t *testing.T) {
    skill, err := NewBuilder("tool-user").
        WithVersion("1.0.0").
        WithDescription("Use when using tools").
        RequireTools("search_tools", "describe_tool", "run_tool").
        Build()

    require.NoError(t, err)
    tools := skill.RequiredTools()
    assert.Len(t, tools, 3)
    assert.Contains(t, tools, "search_tools")
}

func TestBuilder_Conditional(t *testing.T) {
    skill, err := NewBuilder("conditional-skill").
        WithVersion("1.0.0").
        WithDescription("Use when testing conditions").
        Step("maybe-run", "some_tool").
            WithCondition(func(ctx SkillContext) bool {
                enabled, _ := ctx.Input.GetBool("enabled")
                return enabled
            }).
            Done().
        Build()

    require.NoError(t, err)
    assert.Len(t, skill.Steps(), 1)
    assert.NotNil(t, skill.Steps()[0].Condition)
}

func TestBuilder_ValidationFailure(t *testing.T) {
    _, err := NewBuilder("").
        WithVersion("1.0.0").
        Build()

    assert.Error(t, err)
    assert.Contains(t, err.Error(), "name")
}

func TestBuilder_MissingVersion(t *testing.T) {
    _, err := NewBuilder("test-skill").
        WithDescription("Use when testing").
        Build()

    assert.Error(t, err)
    assert.Contains(t, err.Error(), "version")
}

func TestBuilder_MissingDescription(t *testing.T) {
    _, err := NewBuilder("test-skill").
        WithVersion("1.0.0").
        Build()

    assert.Error(t, err)
    assert.Contains(t, err.Error(), "description")
}

func TestBuilder_InputOutputSchema(t *testing.T) {
    inputSchema := map[string]any{
        "type": "object",
        "properties": map[string]any{
            "query": map[string]any{"type": "string"},
        },
        "required": []string{"query"},
    }

    skill, err := NewBuilder("schema-skill").
        WithVersion("1.0.0").
        WithDescription("Use when testing schemas").
        WithInputSchema(inputSchema).
        Build()

    require.NoError(t, err)
    assert.Equal(t, inputSchema, skill.Manifest().InputSchema)
}

func TestCompositeSkill_Execute(t *testing.T) {
    // Create a simple skill with two steps
    skill, err := NewBuilder("test-execution").
        WithVersion("1.0.0").
        WithDescription("Use when testing execution").
        Step("step1", "tool1").
            WithInputMapper(func(ctx SkillContext) any {
                return map[string]any{"from": "step1"}
            }).
            Done().
        Build()

    require.NoError(t, err)

    // Execute - will fail because no runtime, but structure should be valid
    output, err := skill.Execute(context.Background(), SkillInput{"test": true})

    // Without a configured runtime, execution returns error
    assert.Error(t, err)
    assert.Nil(t, output)
}
```

**Step 2: Run test to verify it fails**

Run: `cd toolskill && go test -v -run TestBuilder`
Expected: FAIL with "undefined: NewBuilder"

**Step 3: Write minimal implementation**

```go
// toolskill/builder.go
package toolskill

import (
    "context"
    "fmt"

    "github.com/Masterminds/semver/v3"
)

// Builder provides a fluent API for building skills.
type Builder struct {
    name          string
    version       string
    description   string
    tags          []string
    requiredTools []string
    inputSchema   map[string]any
    outputSchema  map[string]any
    steps         []StepDefinition
    idempotent    bool
    supportsPause bool
    author        string
}

// NewBuilder creates a new skill builder.
func NewBuilder(name string) *Builder {
    return &Builder{
        name:          name,
        tags:          []string{},
        requiredTools: []string{},
        steps:         []StepDefinition{},
    }
}

// WithVersion sets the skill version.
func (b *Builder) WithVersion(version string) *Builder {
    b.version = version
    return b
}

// WithDescription sets the skill description.
func (b *Builder) WithDescription(desc string) *Builder {
    b.description = desc
    return b
}

// WithTags adds tags to the skill.
func (b *Builder) WithTags(tags ...string) *Builder {
    b.tags = append(b.tags, tags...)
    return b
}

// RequireTools specifies tools this skill depends on.
func (b *Builder) RequireTools(tools ...string) *Builder {
    b.requiredTools = append(b.requiredTools, tools...)
    return b
}

// WithInputSchema sets the input JSON schema.
func (b *Builder) WithInputSchema(schema map[string]any) *Builder {
    b.inputSchema = schema
    return b
}

// WithOutputSchema sets the output JSON schema.
func (b *Builder) WithOutputSchema(schema map[string]any) *Builder {
    b.outputSchema = schema
    return b
}

// WithAuthor sets the skill author.
func (b *Builder) WithAuthor(author string) *Builder {
    b.author = author
    return b
}

// Idempotent marks the skill as idempotent.
func (b *Builder) Idempotent() *Builder {
    b.idempotent = true
    return b
}

// SupportsPause marks the skill as supporting pause/resume.
func (b *Builder) SupportsPause() *Builder {
    b.supportsPause = true
    return b
}

// Step begins defining a new step.
func (b *Builder) Step(id, tool string) *StepBuilder {
    return &StepBuilder{
        parent: b,
        step: StepDefinition{
            ID:   id,
            Name: id,
            Tool: tool,
        },
    }
}

// Build creates the skill from the builder configuration.
func (b *Builder) Build() (Skill, error) {
    // Validation
    if b.name == "" {
        return nil, fmt.Errorf("skill name is required")
    }
    if b.version == "" {
        return nil, fmt.Errorf("skill version is required")
    }
    if b.description == "" {
        return nil, fmt.Errorf("skill description is required")
    }

    ver, err := semver.NewVersion(b.version)
    if err != nil {
        return nil, fmt.Errorf("invalid version: %w", err)
    }

    // Collect tool names from steps
    toolSet := make(map[string]bool)
    for _, tool := range b.requiredTools {
        toolSet[tool] = true
    }
    for _, step := range b.steps {
        if step.Tool != "" {
            toolSet[step.Tool] = true
        }
    }
    tools := make([]string, 0, len(toolSet))
    for tool := range toolSet {
        tools = append(tools, tool)
    }

    manifest := &SkillManifest{
        ID:             b.name,
        Name:           b.name,
        Version:        b.version,
        Description:    b.description,
        Tags:           b.tags,
        RequiredTools:  tools,
        InputSchema:    b.inputSchema,
        OutputSchema:   b.outputSchema,
        EstimatedSteps: len(b.steps),
        Idempotent:     b.idempotent,
        SupportsPause:  b.supportsPause,
        Author:         b.author,
    }

    return &CompositeSkill{
        id:       b.name,
        name:     b.name,
        version:  ver,
        manifest: manifest,
        steps:    b.steps,
        tools:    tools,
    }, nil
}

// StepBuilder builds individual steps.
type StepBuilder struct {
    parent *Builder
    step   StepDefinition
}

// WithName sets the step display name.
func (sb *StepBuilder) WithName(name string) *StepBuilder {
    sb.step.Name = name
    return sb
}

// WithInputMapper sets how to transform skill context to tool input.
func (sb *StepBuilder) WithInputMapper(mapper func(SkillContext) any) *StepBuilder {
    sb.step.InputMapper = mapper
    return sb
}

// WithOutputMapper sets how to transform tool output.
func (sb *StepBuilder) WithOutputMapper(mapper func(any) any) *StepBuilder {
    sb.step.OutputMapper = mapper
    return sb
}

// WithCondition sets a condition for step execution.
func (sb *StepBuilder) WithCondition(cond func(SkillContext) bool) *StepBuilder {
    sb.step.Condition = cond
    return sb
}

// WithTimeout sets step timeout in milliseconds.
func (sb *StepBuilder) WithTimeout(ms int64) *StepBuilder {
    sb.step.Timeout = ms
    return sb
}

// Optional marks the step as optional (continue on failure).
func (sb *StepBuilder) Optional() *StepBuilder {
    sb.step.Optional = true
    return sb
}

// OnError sets the error handler.
func (sb *StepBuilder) OnError(handler ErrorHandler) *StepBuilder {
    sb.step.OnError = handler
    return sb
}

// Done finishes step configuration and returns to the builder.
func (sb *StepBuilder) Done() *Builder {
    sb.parent.steps = append(sb.parent.steps, sb.step)
    return sb.parent
}

// CompositeSkill is a skill built from steps.
type CompositeSkill struct {
    id       string
    name     string
    version  *semver.Version
    manifest *SkillManifest
    steps    []StepDefinition
    tools    []string
    runtime  SkillRuntime
}

func (s *CompositeSkill) ID() string               { return s.id }
func (s *CompositeSkill) Name() string             { return s.name }
func (s *CompositeSkill) Version() *semver.Version { return s.version }
func (s *CompositeSkill) Manifest() *SkillManifest { return s.manifest }
func (s *CompositeSkill) RequiredTools() []string  { return s.tools }
func (s *CompositeSkill) Steps() []StepDefinition  { return s.steps }

// SetRuntime configures the runtime for execution.
func (s *CompositeSkill) SetRuntime(rt SkillRuntime) {
    s.runtime = rt
}

// Execute runs the skill with the given input.
func (s *CompositeSkill) Execute(ctx context.Context, input SkillInput) (*SkillOutput, error) {
    if s.runtime == nil {
        return nil, fmt.Errorf("skill runtime not configured")
    }
    return s.runtime.Execute(ctx, s, input)
}
```

**Step 4: Run test to verify it passes**

Run: `cd toolskill && go test -v -run TestBuilder`
Expected: PASS

**Step 5: Commit**

```bash
git add toolskill/builder.go toolskill/builder_test.go
git commit -m "$(cat <<'EOF'
feat(toolskill): add fluent builder for skill composition

- Add Builder with fluent API for skill creation
- Add StepBuilder for defining workflow steps
- Support input/output mappers for data transformation
- Support conditional step execution
- Support step timeouts and optional steps
- Add CompositeSkill implementation
- Extract required tools from steps automatically

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Task 5: Skill Runtime

**Files:**
- Create: `toolskill/runtime.go`
- Test: `toolskill/runtime_test.go`

**Step 1: Write the failing test**

```go
// toolskill/runtime_test.go
package toolskill

import (
    "context"
    "errors"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

// mockToolExecutor simulates tool execution
type mockToolExecutor struct {
    results map[string]any
    errors  map[string]error
    calls   []string
}

func (m *mockToolExecutor) Execute(ctx context.Context, tool string, input any) (any, error) {
    m.calls = append(m.calls, tool)
    if err, ok := m.errors[tool]; ok {
        return nil, err
    }
    if result, ok := m.results[tool]; ok {
        return result, nil
    }
    return map[string]any{"tool": tool, "executed": true}, nil
}

func TestRuntime_Execute(t *testing.T) {
    executor := &mockToolExecutor{
        results: map[string]any{
            "search_tools":  map[string]any{"tools": []string{"tool1", "tool2"}},
            "describe_tool": map[string]any{"description": "A test tool"},
        },
    }

    runtime := NewRuntime(RuntimeConfig{
        ToolExecutor: executor,
    })

    skill, _ := NewBuilder("test-skill").
        WithVersion("1.0.0").
        WithDescription("Use when testing runtime").
        Step("search", "search_tools").
            WithInputMapper(func(ctx SkillContext) any {
                return map[string]any{"query": ctx.Input["query"]}
            }).
            Done().
        Step("describe", "describe_tool").
            WithInputMapper(func(ctx SkillContext) any {
                return map[string]any{"tool": "tool1"}
            }).
            Done().
        Build()

    output, err := runtime.Execute(context.Background(), skill, SkillInput{"query": "test"})
    require.NoError(t, err)
    assert.True(t, output.Success)
    assert.Len(t, output.StepsCompleted, 2)
    assert.Contains(t, output.StepsCompleted, "search")
    assert.Contains(t, output.StepsCompleted, "describe")

    // Verify tools were called in order
    assert.Equal(t, []string{"search_tools", "describe_tool"}, executor.calls)
}

func TestRuntime_ConditionalStep(t *testing.T) {
    executor := &mockToolExecutor{}

    runtime := NewRuntime(RuntimeConfig{
        ToolExecutor: executor,
    })

    skill, _ := NewBuilder("conditional").
        WithVersion("1.0.0").
        WithDescription("Use when testing conditions").
        Step("always", "tool1").Done().
        Step("maybe", "tool2").
            WithCondition(func(ctx SkillContext) bool {
                return ctx.Input["enabled"] == true
            }).
            Done().
        Step("final", "tool3").Done().
        Build()

    // With condition false
    output, err := runtime.Execute(context.Background(), skill, SkillInput{"enabled": false})
    require.NoError(t, err)
    assert.True(t, output.Success)
    assert.Equal(t, []string{"tool1", "tool3"}, executor.calls) // tool2 skipped

    // With condition true
    executor.calls = nil
    output, err = runtime.Execute(context.Background(), skill, SkillInput{"enabled": true})
    require.NoError(t, err)
    assert.Equal(t, []string{"tool1", "tool2", "tool3"}, executor.calls)
}

func TestRuntime_StepFailure(t *testing.T) {
    executor := &mockToolExecutor{
        errors: map[string]error{
            "failing_tool": errors.New("tool failed"),
        },
    }

    runtime := NewRuntime(RuntimeConfig{
        ToolExecutor: executor,
    })

    skill, _ := NewBuilder("failing-skill").
        WithVersion("1.0.0").
        WithDescription("Use when testing failures").
        Step("step1", "good_tool").Done().
        Step("step2", "failing_tool").Done().
        Step("step3", "another_tool").Done().
        Build()

    output, err := runtime.Execute(context.Background(), skill, SkillInput{})
    assert.Error(t, err)
    assert.False(t, output.Success)
    assert.Len(t, output.StepsCompleted, 1) // Only step1 completed
    assert.Contains(t, output.ErrorMessage, "failing_tool")
}

func TestRuntime_OptionalStep(t *testing.T) {
    executor := &mockToolExecutor{
        errors: map[string]error{
            "optional_tool": errors.New("optional failed"),
        },
    }

    runtime := NewRuntime(RuntimeConfig{
        ToolExecutor: executor,
    })

    skill, _ := NewBuilder("optional-skill").
        WithVersion("1.0.0").
        WithDescription("Use when testing optional steps").
        Step("step1", "tool1").Done().
        Step("step2", "optional_tool").Optional().Done().
        Step("step3", "tool3").Done().
        Build()

    output, err := runtime.Execute(context.Background(), skill, SkillInput{})
    require.NoError(t, err) // Should succeed despite optional step failure
    assert.True(t, output.Success)
    assert.Contains(t, output.StepsCompleted, "step1")
    assert.Contains(t, output.StepsCompleted, "step3")
}

func TestRuntime_Timeout(t *testing.T) {
    executor := &mockToolExecutor{}

    runtime := NewRuntime(RuntimeConfig{
        ToolExecutor:   executor,
        DefaultTimeout: 50 * time.Millisecond,
    })

    // Create context that will timeout
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
    defer cancel()

    skill, _ := NewBuilder("timeout-skill").
        WithVersion("1.0.0").
        WithDescription("Use when testing timeouts").
        Step("step1", "tool1").Done().
        Build()

    // Simulate slow execution by waiting
    time.Sleep(20 * time.Millisecond)

    _, err := runtime.Execute(ctx, skill, SkillInput{})
    assert.Error(t, err)
    assert.True(t, errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled))
}

func TestRuntime_GetStatus(t *testing.T) {
    executor := &mockToolExecutor{}
    runtime := NewRuntime(RuntimeConfig{ToolExecutor: executor})

    skill, _ := NewBuilder("status-skill").
        WithVersion("1.0.0").
        WithDescription("Use when testing status").
        Step("step1", "tool1").Done().
        Build()

    // Start async execution
    execID, err := runtime.ExecuteAsync(context.Background(), skill, SkillInput{})
    require.NoError(t, err)

    // Check status (may be running or completed)
    status, err := runtime.GetStatus(execID)
    require.NoError(t, err)
    assert.Equal(t, execID, status.ID)
    assert.Equal(t, "status-skill", status.SkillID)
}

func TestRuntime_Cancel(t *testing.T) {
    executor := &mockToolExecutor{}
    runtime := NewRuntime(RuntimeConfig{ToolExecutor: executor})

    skill, _ := NewBuilder("cancel-skill").
        WithVersion("1.0.0").
        WithDescription("Use when testing cancellation").
        Step("step1", "tool1").Done().
        Build()

    execID, _ := runtime.ExecuteAsync(context.Background(), skill, SkillInput{})

    err := runtime.Cancel(execID)
    assert.NoError(t, err)
}
```

**Step 2: Run test to verify it fails**

Run: `cd toolskill && go test -v -run TestRuntime`
Expected: FAIL with "undefined: NewRuntime"

**Step 3: Write minimal implementation**

```go
// toolskill/runtime.go
package toolskill

import (
    "context"
    "fmt"
    "sync"
    "sync/atomic"
    "time"
)

// ExecutionID uniquely identifies a skill execution.
type ExecutionID string

// ExecutionState represents the state of skill execution.
type ExecutionState string

const (
    StatePending   ExecutionState = "pending"
    StateRunning   ExecutionState = "running"
    StatePaused    ExecutionState = "paused"
    StateCompleted ExecutionState = "completed"
    StateFailed    ExecutionState = "failed"
    StateCancelled ExecutionState = "cancelled"
)

// ToolExecutor executes individual tools.
type ToolExecutor interface {
    Execute(ctx context.Context, tool string, input any) (any, error)
}

// SkillRuntime executes skills.
type SkillRuntime interface {
    Execute(ctx context.Context, skill Skill, input SkillInput) (*SkillOutput, error)
    ExecuteAsync(ctx context.Context, skill Skill, input SkillInput) (ExecutionID, error)
    GetStatus(execID ExecutionID) (*ExecutionStatus, error)
    Pause(execID ExecutionID) error
    Resume(execID ExecutionID) error
    Cancel(execID ExecutionID) error
}

// ExecutionStatus tracks skill execution progress.
type ExecutionStatus struct {
    ID             ExecutionID
    SkillID        string
    State          ExecutionState
    CurrentStep    string
    CompletedSteps []string
    Progress       float64
    StartedAt      time.Time
    CompletedAt    *time.Time
    Error          error
}

// RuntimeConfig configures the skill runtime.
type RuntimeConfig struct {
    ToolExecutor   ToolExecutor
    DefaultTimeout time.Duration
    MaxConcurrent  int
}

// Runtime is the default SkillRuntime implementation.
type Runtime struct {
    executor       ToolExecutor
    defaultTimeout time.Duration
    maxConcurrent  int

    mu         sync.RWMutex
    executions map[ExecutionID]*execution
    nextID     atomic.Uint64
}

type execution struct {
    id        ExecutionID
    skill     Skill
    input     SkillInput
    status    *ExecutionStatus
    cancel    context.CancelFunc
    done      chan struct{}
    output    *SkillOutput
    err       error
}

// NewRuntime creates a new skill runtime.
func NewRuntime(cfg RuntimeConfig) *Runtime {
    if cfg.DefaultTimeout == 0 {
        cfg.DefaultTimeout = 30 * time.Second
    }
    if cfg.MaxConcurrent == 0 {
        cfg.MaxConcurrent = 10
    }

    return &Runtime{
        executor:       cfg.ToolExecutor,
        defaultTimeout: cfg.DefaultTimeout,
        maxConcurrent:  cfg.MaxConcurrent,
        executions:     make(map[ExecutionID]*execution),
    }
}

// Execute runs a skill synchronously.
func (r *Runtime) Execute(ctx context.Context, skill Skill, input SkillInput) (*SkillOutput, error) {
    // Check context before starting
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }

    execCtx := &SkillContext{
        Input:    input,
        Results:  make(map[string]any),
        Metadata: make(map[string]any),
    }

    steps := skill.Steps()
    completed := make([]string, 0, len(steps))
    startTime := time.Now()

    for _, step := range steps {
        // Check context
        select {
        case <-ctx.Done():
            return &SkillOutput{
                Success:        false,
                Error:          ctx.Err(),
                ErrorMessage:   ctx.Err().Error(),
                StepsCompleted: completed,
                Duration:       time.Since(startTime).Milliseconds(),
            }, ctx.Err()
        default:
        }

        execCtx.StepID = step.ID

        // Check condition
        if step.Condition != nil && !step.Condition(*execCtx) {
            continue // Skip this step
        }

        // Prepare input
        var toolInput any
        if step.InputMapper != nil {
            toolInput = step.InputMapper(*execCtx)
        } else {
            toolInput = input
        }

        // Execute tool
        result, err := r.executor.Execute(ctx, step.Tool, toolInput)
        if err != nil {
            if step.Optional {
                // Continue on optional step failure
                continue
            }

            // Handle error
            if step.OnError != nil {
                if handlerErr := step.OnError(*execCtx, err); handlerErr != nil {
                    return &SkillOutput{
                        Success:        false,
                        Error:          handlerErr,
                        ErrorMessage:   fmt.Sprintf("step %s failed: %v", step.ID, err),
                        StepsCompleted: completed,
                        Duration:       time.Since(startTime).Milliseconds(),
                    }, handlerErr
                }
            } else {
                return &SkillOutput{
                    Success:        false,
                    Error:          err,
                    ErrorMessage:   fmt.Sprintf("step %s (%s) failed: %v", step.ID, step.Tool, err),
                    StepsCompleted: completed,
                    Duration:       time.Since(startTime).Milliseconds(),
                }, fmt.Errorf("%w: %v", ErrStepFailed, err)
            }
        }

        // Transform output if mapper provided
        if step.OutputMapper != nil {
            result = step.OutputMapper(result)
        }

        // Store result
        execCtx.Results[step.ID] = result
        completed = append(completed, step.ID)
    }

    return &SkillOutput{
        Success:        true,
        Result:         execCtx.Results,
        StepsCompleted: completed,
        Duration:       time.Since(startTime).Milliseconds(),
    }, nil
}

// ExecuteAsync runs a skill asynchronously.
func (r *Runtime) ExecuteAsync(ctx context.Context, skill Skill, input SkillInput) (ExecutionID, error) {
    id := ExecutionID(fmt.Sprintf("exec-%d", r.nextID.Add(1)))

    execCtx, cancel := context.WithCancel(ctx)

    exec := &execution{
        id:     id,
        skill:  skill,
        input:  input,
        cancel: cancel,
        done:   make(chan struct{}),
        status: &ExecutionStatus{
            ID:        id,
            SkillID:   skill.ID(),
            State:     StatePending,
            StartedAt: time.Now(),
        },
    }

    r.mu.Lock()
    r.executions[id] = exec
    r.mu.Unlock()

    go func() {
        defer close(exec.done)

        exec.status.State = StateRunning
        output, err := r.Execute(execCtx, skill, input)

        exec.output = output
        exec.err = err

        now := time.Now()
        exec.status.CompletedAt = &now

        if err != nil {
            exec.status.State = StateFailed
            exec.status.Error = err
        } else {
            exec.status.State = StateCompleted
            exec.status.CompletedSteps = output.StepsCompleted
            exec.status.Progress = 1.0
        }
    }()

    return id, nil
}

// GetStatus retrieves execution status.
func (r *Runtime) GetStatus(execID ExecutionID) (*ExecutionStatus, error) {
    r.mu.RLock()
    exec, ok := r.executions[execID]
    r.mu.RUnlock()

    if !ok {
        return nil, fmt.Errorf("execution not found: %s", execID)
    }

    return exec.status, nil
}

// Pause pauses execution (not yet implemented).
func (r *Runtime) Pause(execID ExecutionID) error {
    return fmt.Errorf("pause not yet implemented")
}

// Resume resumes execution (not yet implemented).
func (r *Runtime) Resume(execID ExecutionID) error {
    return fmt.Errorf("resume not yet implemented")
}

// Cancel cancels an execution.
func (r *Runtime) Cancel(execID ExecutionID) error {
    r.mu.RLock()
    exec, ok := r.executions[execID]
    r.mu.RUnlock()

    if !ok {
        return fmt.Errorf("execution not found: %s", execID)
    }

    exec.cancel()
    exec.status.State = StateCancelled
    return nil
}
```

**Step 4: Run test to verify it passes**

Run: `cd toolskill && go test -v -run TestRuntime`
Expected: PASS

**Step 5: Commit**

```bash
git add toolskill/runtime.go toolskill/runtime_test.go
git commit -m "$(cat <<'EOF'
feat(toolskill): add skill runtime with step execution

- Add SkillRuntime interface for skill execution
- Add Runtime implementation with sync/async execution
- Execute steps sequentially with context propagation
- Support conditional steps via Condition function
- Support optional steps that continue on failure
- Add input/output mappers for data transformation
- Track execution status with progress updates
- Support cancellation via context

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Task 6: Discovery and Integration

**Files:**
- Create: `toolskill/discovery.go`
- Create: `toolskill/provider.go`
- Test: `toolskill/discovery_test.go`

**Step 1: Write the failing test**

```go
// toolskill/discovery_test.go
package toolskill

import (
    "os"
    "path/filepath"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestDiscovery_ScanDirectory(t *testing.T) {
    // Create temp directory with skill files
    tmpDir, err := os.MkdirTemp("", "skills-test")
    require.NoError(t, err)
    defer os.RemoveAll(tmpDir)

    // Create skill directory structure
    skill1Dir := filepath.Join(tmpDir, "research-skill")
    require.NoError(t, os.MkdirAll(skill1Dir, 0755))

    skill1Content := `---
name: research-skill
description: Use when researching topics
metadata:
  version: "1.0.0"
---

# Research Skill

## Overview
Helps research topics.
`
    require.NoError(t, os.WriteFile(filepath.Join(skill1Dir, "SKILL.md"), []byte(skill1Content), 0644))

    skill2Dir := filepath.Join(tmpDir, "debug-skill")
    require.NoError(t, os.MkdirAll(skill2Dir, 0755))

    skill2Content := `---
name: debug-skill
description: Use when debugging issues
metadata:
  version: "2.0.0"
---

# Debug Skill

## Overview
Helps debug code.
`
    require.NoError(t, os.WriteFile(filepath.Join(skill2Dir, "SKILL.md"), []byte(skill2Content), 0644))

    // Scan directory
    discovery := NewDiscovery()
    skills, err := discovery.ScanDirectory(tmpDir)
    require.NoError(t, err)

    assert.Len(t, skills, 2)

    // Verify skills were found
    names := make([]string, len(skills))
    for i, s := range skills {
        names[i] = s.Name
    }
    assert.Contains(t, names, "research-skill")
    assert.Contains(t, names, "debug-skill")
}

func TestDiscovery_ScanDirectoryNested(t *testing.T) {
    tmpDir, err := os.MkdirTemp("", "skills-nested")
    require.NoError(t, err)
    defer os.RemoveAll(tmpDir)

    // Create nested structure
    nestedDir := filepath.Join(tmpDir, "category", "subcategory", "my-skill")
    require.NoError(t, os.MkdirAll(nestedDir, 0755))

    content := `---
name: nested-skill
description: Use when testing nested discovery
---

# Nested Skill
`
    require.NoError(t, os.WriteFile(filepath.Join(nestedDir, "SKILL.md"), []byte(content), 0644))

    discovery := NewDiscovery()
    skills, err := discovery.ScanDirectory(tmpDir)
    require.NoError(t, err)

    assert.Len(t, skills, 1)
    assert.Equal(t, "nested-skill", skills[0].Name)
}

func TestDiscovery_ScanEmptyDirectory(t *testing.T) {
    tmpDir, err := os.MkdirTemp("", "skills-empty")
    require.NoError(t, err)
    defer os.RemoveAll(tmpDir)

    discovery := NewDiscovery()
    skills, err := discovery.ScanDirectory(tmpDir)
    require.NoError(t, err)
    assert.Len(t, skills, 0)
}

func TestDiscovery_ScanNonexistentDirectory(t *testing.T) {
    discovery := NewDiscovery()
    _, err := discovery.ScanDirectory("/nonexistent/path")
    assert.Error(t, err)
}

func TestSkillToolProvider_Tools(t *testing.T) {
    registry := NewRegistry()

    skill1, _ := NewBuilder("research").
        WithVersion("1.0.0").
        WithDescription("Use when researching").
        WithInputSchema(map[string]any{
            "type": "object",
            "properties": map[string]any{
                "query": map[string]any{"type": "string"},
            },
        }).
        Build()

    skill2, _ := NewBuilder("debug").
        WithVersion("1.0.0").
        WithDescription("Use when debugging").
        Build()

    _ = registry.Register(skill1)
    _ = registry.Register(skill2)

    provider := NewSkillToolProvider(registry, nil)
    tools := provider.Tools()

    assert.Len(t, tools, 2)

    // Verify tool naming convention
    names := make([]string, len(tools))
    for i, tool := range tools {
        names[i] = tool.Name
    }
    assert.Contains(t, names, "skill:research")
    assert.Contains(t, names, "skill:debug")

    // Verify description prefix
    for _, tool := range tools {
        assert.Contains(t, tool.Description, "[SKILL]")
    }
}

func TestSkillToolProvider_Handle(t *testing.T) {
    registry := NewRegistry()
    executor := &mockToolExecutor{}
    runtime := NewRuntime(RuntimeConfig{ToolExecutor: executor})

    skill, _ := NewBuilder("test-handle").
        WithVersion("1.0.0").
        WithDescription("Use when testing handle").
        Step("step1", "tool1").Done().
        Build()
    skill.(*CompositeSkill).SetRuntime(runtime)

    _ = registry.Register(skill)

    provider := NewSkillToolProvider(registry, runtime)

    // Handle skill execution
    result, err := provider.Handle("skill:test-handle", map[string]any{"input": "value"})
    require.NoError(t, err)

    output, ok := result.(*SkillOutput)
    require.True(t, ok)
    assert.True(t, output.Success)
}
```

**Step 2: Run test to verify it fails**

Run: `cd toolskill && go test -v -run "TestDiscovery|TestSkillToolProvider"`
Expected: FAIL with "undefined: NewDiscovery"

**Step 3: Write minimal implementation**

```go
// toolskill/discovery.go
package toolskill

import (
    "os"
    "path/filepath"

    "github.com/yourusername/toolskill/skillmd"
)

// Discovery finds skills in standard locations.
type Discovery struct {
    parser *skillmd.Parser
}

// NewDiscovery creates a new skill discovery instance.
func NewDiscovery() *Discovery {
    return &Discovery{
        parser: skillmd.NewParser(),
    }
}

// ScanDirectory finds all SKILL.md files in a directory tree.
func (d *Discovery) ScanDirectory(root string) ([]*skillmd.SkillMD, error) {
    var skills []*skillmd.SkillMD

    err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
        if err != nil {
            return err
        }

        if entry.IsDir() {
            return nil
        }

        if entry.Name() == "SKILL.md" {
            skill, err := d.parser.Parse(path)
            if err != nil {
                // Log warning but continue scanning
                return nil
            }
            skills = append(skills, skill)
        }

        return nil
    })

    if err != nil {
        return nil, err
    }

    return skills, nil
}

// ScanUserSkills scans ~/.claude/skills/ for user-defined skills.
func (d *Discovery) ScanUserSkills() ([]*skillmd.SkillMD, error) {
    homeDir, err := os.UserHomeDir()
    if err != nil {
        return nil, err
    }

    // Check both possible locations
    paths := []string{
        filepath.Join(homeDir, ".claude", "skills"),
        filepath.Join(homeDir, ".config", "claude", "skills"),
    }

    var allSkills []*skillmd.SkillMD
    for _, path := range paths {
        if _, err := os.Stat(path); os.IsNotExist(err) {
            continue
        }
        skills, err := d.ScanDirectory(path)
        if err != nil {
            continue
        }
        allSkills = append(allSkills, skills...)
    }

    return allSkills, nil
}

// ScanProjectSkills scans .claude/skills/ in a project root.
func (d *Discovery) ScanProjectSkills(projectRoot string) ([]*skillmd.SkillMD, error) {
    skillsDir := filepath.Join(projectRoot, ".claude", "skills")
    if _, err := os.Stat(skillsDir); os.IsNotExist(err) {
        return nil, nil
    }
    return d.ScanDirectory(skillsDir)
}

// ScanPluginSkills scans a plugin directory for skills.
func (d *Discovery) ScanPluginSkills(pluginDir string) ([]*skillmd.SkillMD, error) {
    skillsDir := filepath.Join(pluginDir, "skills")
    if _, err := os.Stat(skillsDir); os.IsNotExist(err) {
        return nil, nil
    }
    return d.ScanDirectory(skillsDir)
}
```

```go
// toolskill/provider.go
package toolskill

import (
    "context"
    "encoding/json"
    "fmt"
)

// Tool represents an MCP tool definition.
type Tool struct {
    Name        string         `json:"name"`
    Description string         `json:"description"`
    InputSchema map[string]any `json:"inputSchema,omitempty"`
}

// SkillToolProvider exposes skills as MCP tools.
type SkillToolProvider struct {
    registry SkillRegistry
    runtime  SkillRuntime
}

// NewSkillToolProvider creates a provider that exposes skills as tools.
func NewSkillToolProvider(registry SkillRegistry, runtime SkillRuntime) *SkillToolProvider {
    return &SkillToolProvider{
        registry: registry,
        runtime:  runtime,
    }
}

// Tools returns all skills as MCP tool definitions.
func (p *SkillToolProvider) Tools() []*Tool {
    manifests := p.registry.List()
    tools := make([]*Tool, 0, len(manifests))

    for _, m := range manifests {
        tool := &Tool{
            Name:        "skill:" + m.Name,
            Description: fmt.Sprintf("[SKILL] %s", m.Description),
            InputSchema: m.InputSchema,
        }
        tools = append(tools, tool)
    }

    return tools
}

// Handle executes a skill via its tool name.
func (p *SkillToolProvider) Handle(toolName string, input any) (any, error) {
    // Extract skill name from tool name (remove "skill:" prefix)
    if len(toolName) <= 6 || toolName[:6] != "skill:" {
        return nil, fmt.Errorf("invalid skill tool name: %s", toolName)
    }
    skillName := toolName[6:]

    // Get skill from registry
    skill, err := p.registry.GetByName(skillName, "")
    if err != nil {
        return nil, fmt.Errorf("skill not found: %s", skillName)
    }

    // Convert input to SkillInput
    var skillInput SkillInput
    switch v := input.(type) {
    case SkillInput:
        skillInput = v
    case map[string]any:
        skillInput = SkillInput(v)
    case json.RawMessage:
        if err := json.Unmarshal(v, &skillInput); err != nil {
            return nil, fmt.Errorf("invalid input: %w", err)
        }
    default:
        return nil, fmt.Errorf("unsupported input type: %T", input)
    }

    // Execute skill
    if p.runtime == nil {
        return nil, fmt.Errorf("runtime not configured")
    }

    return p.runtime.Execute(context.Background(), skill, skillInput)
}

// Name returns the provider name.
func (p *SkillToolProvider) Name() string {
    return "skills"
}
```

**Step 4: Run test to verify it passes**

Run: `cd toolskill && go test -v -run "TestDiscovery|TestSkillToolProvider"`
Expected: PASS

**Step 5: Commit**

```bash
git add toolskill/discovery.go toolskill/provider.go toolskill/discovery_test.go
git commit -m "$(cat <<'EOF'
feat(toolskill): add discovery and MCP tool provider

- Add Discovery for finding SKILL.md files in directories
- Support user skills (~/.claude/skills/)
- Support project skills (.claude/skills/)
- Support plugin skills (plugins/*/skills/)
- Add SkillToolProvider exposing skills as MCP tools
- Use skill: prefix for tool naming convention
- Add [SKILL] prefix to descriptions for visibility

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Verification Checklist

Before marking this PRD complete, verify:

- [ ] All tests pass: `go test ./toolskill/...`
- [ ] Code coverage > 80%: `go test -cover ./toolskill/...`
- [ ] No lint errors: `golangci-lint run ./toolskill/...`
- [ ] Examples compile: `go build ./toolskill/examples/...`
- [ ] Integration with toolset works
- [ ] SKILL.md parsing handles edge cases
- [ ] Registry thread-safety verified
- [ ] Runtime handles cancellation properly

## Definition of Done

1. **Core Types** - Skill interface, SkillManifest, SkillInput/Output defined
2. **SKILL.md Parser** - Parses standard format with frontmatter extraction
3. **Registry** - Thread-safe registration with version constraint resolution
4. **Builder** - Fluent API for skill composition with steps
5. **Runtime** - Executes skills with conditional steps and error handling
6. **Discovery** - Scans standard locations for SKILL.md files
7. **Provider** - Exposes skills as MCP tools with skill: prefix
8. **Documentation** - README.md with examples

## Architecture Notes

### Abstraction Hierarchy

```
Agents → Skills → Toolsets → Tools
```

Skills orchestrate multiple tools into higher-level behaviors. The `skill:` prefix distinguishes skill tools from regular tools in MCP listings.

### SKILL.md Standard

The toolskill library implements the Agent Skills Open Standard:
- YAML frontmatter with name and description (required)
- Description must start with "Use when" for discovery optimization
- Markdown body with Overview, When to Use, How It Works sections

### Integration Points

- **toolset**: Skills use toolsets for filtered tool access
- **toolrun**: Runtime delegates to toolrun for execution
- **toolobserve**: Optional tracing integration via middleware
- **toolversion**: Manifest includes version for compatibility
