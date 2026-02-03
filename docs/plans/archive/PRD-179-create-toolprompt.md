# PRD-179: Create toolprompt

**Phase:** 7 - Protocol Layer
**Priority:** Medium
**Effort:** 8 hours
**Dependencies:** PRD-173
**Status:** Done (2026-02-01)

---

## Objective

Create `toolprotocol/prompt/` for MCP Prompts support.

---

## Package Contents

- Prompt definition and storage
- Prompt templates with arguments
- Message generation
- Prompt registry

## Implementation Summary

- Implemented `Provider` + registry with argument validation.
- Added helpers for message/content creation.

---

## Key Implementation

```go
package prompt

import "context"

// Prompt represents an MCP prompt.
type Prompt struct {
    Name        string
    Description string
    Arguments   []Argument
}

// Argument describes a prompt argument.
type Argument struct {
    Name        string
    Description string
    Required    bool
}

// Message represents a generated message.
type Message struct {
    Role    string // "user" or "assistant"
    Content Content
}

// Content represents message content.
type Content struct {
    Type     string // "text", "image", "resource"
    Text     string
    MIMEType string
    Data     []byte
    Resource *ResourceRef
}

// ResourceRef references a resource.
type ResourceRef struct {
    URI string
}

// Provider serves prompts.
type Provider interface {
    List(ctx context.Context) ([]Prompt, error)
    Get(ctx context.Context, name string, args map[string]string) ([]Message, error)
}

// Registry manages prompt providers.
type Registry struct {
    prompts map[string]*Prompt
    handlers map[string]PromptHandler
    mu      sync.RWMutex
}

// PromptHandler generates messages from arguments.
type PromptHandler func(ctx context.Context, args map[string]string) ([]Message, error)

func (r *Registry) Register(prompt Prompt, handler PromptHandler) error
func (r *Registry) List(ctx context.Context) ([]Prompt, error)
func (r *Registry) Get(ctx context.Context, name string, args map[string]string) ([]Message, error)
```

---

## Commit Message

```
feat(prompt): add MCP prompts support

Create prompt package for prompt templates.

Features:
- Prompt definition
- Argument handling
- Message generation
- Provider registry

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
```

---

## Next Steps

- Gate G5: Protocol layer complete (all 10 packages)
- PRD-180: Update metatools-mcp
