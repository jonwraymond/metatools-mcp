# PRD-172: Create tooldiscover

**Phase:** 7 - Protocol Layer
**Priority:** High
**Effort:** 8 hours
**Dependencies:** PRD-171

---

## Objective

Create `toolprotocol/discover/` for capability discovery across protocols.

---

## Package Contents

- Capability advertisement
- Service discovery
- Protocol negotiation
- Agent card support (A2A)

---

## Key Implementation

```go
package discover

// Discoverable represents a discoverable service.
type Discoverable interface {
    Name() string
    Description() string
    Version() string
    Capabilities() *Capabilities
    DiscoveryEndpoint() string
}

// Capabilities describes service capabilities.
type Capabilities struct {
    Tools       bool
    Resources   bool
    Prompts     bool
    Streaming   bool
    Sampling    bool
    Extensions  []string
}

// Discovery performs service discovery.
type Discovery struct {
    registry map[string]Discoverable
}

func (d *Discovery) Register(svc Discoverable) error
func (d *Discovery) Discover(ctx context.Context, filter DiscoveryFilter) ([]Discoverable, error)
func (d *Discovery) Negotiate(ctx context.Context, client, server *Capabilities) (*Capabilities, error)
```

---

## Commit Message

```
feat(discover): add capability discovery

Create discover package for service discovery.

Features:
- Capability advertisement
- Service registration
- Protocol negotiation
- Agent card support

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
```

---

## Next Steps

- PRD-173: Create toolcontent
