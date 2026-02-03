# PRD-178: Create toolresource

**Phase:** 7 - Protocol Layer
**Priority:** Medium
**Effort:** 10 hours
**Dependencies:** PRD-130
**Status:** Done (2026-02-01)

---

## Objective

Create `toolprotocol/resource/` for MCP Resources support.

---

## Package Contents

- Resource definition and storage
- Resource templates
- Subscription management
- Static provider helpers

## Implementation Summary

- Implemented `Provider` interface, registry, and subscription manager.
- Added static provider with template support and tests.

---

## Key Implementation

```go
package resource

import "context"

// Resource represents an MCP resource.
type Resource struct {
    URI         string
    Name        string
    Description string
    MIMEType    string
}

// Contents represents resource contents.
type Contents struct {
    URI      string
    MIMEType string
    Text     string
    Blob     []byte
}

// Template represents a resource template.
type Template struct {
    URITemplate string
    Name        string
    Description string
    MIMEType    string
}

// Provider serves resources.
type Provider interface {
    List(ctx context.Context) ([]Resource, error)
    Read(ctx context.Context, uri string) (*Contents, error)
    Templates(ctx context.Context) ([]Template, error)
}

// Subscriber receives resource updates.
type Subscriber interface {
    Subscribe(ctx context.Context, uri string) (<-chan *Contents, error)
    Unsubscribe(ctx context.Context, uri string) error
}

// Registry manages resource providers.
type Registry struct {
    providers map[string]Provider
    mu        sync.RWMutex
}

func (r *Registry) Register(scheme string, provider Provider) error
func (r *Registry) List(ctx context.Context) ([]Resource, error)
func (r *Registry) Read(ctx context.Context, uri string) (*Contents, error)
```

---

## Commit Message

```
feat(resource): add MCP resources support

Create resource package for content serving.

Features:
- Resource definition
- Template support
- Subscription management
- Multi-provider registry

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
```

---

## Next Steps

- PRD-179: Create toolprompt
