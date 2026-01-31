# PRD-177: Create toolelicit

**Phase:** 7 - Protocol Layer
**Priority:** Medium
**Effort:** 6 hours
**Dependencies:** PRD-173

---

## Objective

Create `toolprotocol/elicit/` for user input elicitation (MCP feature).

---

## Package Contents

- Elicitation requests
- Input types (text, confirmation, choice)
- Response handling
- Timeout management

---

## Key Implementation

```go
package elicit

import "context"

// Request represents an elicitation request.
type Request struct {
    ID          string
    Type        RequestType
    Message     string
    Schema      any        // JSON Schema for structured input
    Choices     []Choice   // For choice type
    Default     any
    Timeout     time.Duration
}

// RequestType defines elicitation types.
type RequestType string

const (
    TypeText         RequestType = "text"
    TypeConfirmation RequestType = "confirmation"
    TypeChoice       RequestType = "choice"
    TypeForm         RequestType = "form"
)

// Choice represents a selection option.
type Choice struct {
    ID          string
    Label       string
    Description string
}

// Response represents user response.
type Response struct {
    RequestID string
    Value     any
    Cancelled bool
    Timeout   bool
}

// Elicitor handles user input requests.
type Elicitor interface {
    Elicit(ctx context.Context, req *Request) (*Response, error)
}

// Handler processes elicitation on client side.
type Handler interface {
    Handle(ctx context.Context, req *Request) (*Response, error)
}
```

---

## Commit Message

```
feat(elicit): add input elicitation

Create elicit package for user input requests.

Features:
- Multiple input types
- JSON Schema validation
- Timeout handling
- Choice selection

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
```

---

## Next Steps

- PRD-178: Create toolresource
