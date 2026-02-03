# PRD-171: Create toolwire

**Phase:** 7 - Protocol Layer
**Priority:** Critical
**Effort:** 12 hours
**Dependencies:** PRD-170
**Status:** Done (2026-02-01)

---

## Objective

Create `toolprotocol/wire/` for protocol wire adapters supporting MCP, A2A (Google), and ACP (IBM).

---

## Package Design

**Location:** `github.com/jonwraymond/toolprotocol/wire`

**Purpose:**
- Wire protocol adapters
- MCP JSON-RPC 2.0 encoding
- A2A protocol encoding
- ACP protocol encoding
- Protocol negotiation

---

## Deliverables

| Deliverable | Location | Description |
|-------------|----------|-------------|
| Wire Package | `toolprotocol/wire/` | Protocol adapters |
| MCP | `wire/mcp.go` | MCP wire adapter |
| A2A | `wire/a2a.go` | Google A2A adapter |
| ACP | `wire/acp.go` | IBM ACP adapter |
| Registry | `wire/registry.go` | Wire registry + factory |
| Tests | `wire/*_test.go` | Comprehensive tests |

## Implementation Summary

- Implemented `Wire` interface with deterministic encode/decode.
- Added MCP, A2A, and ACP wire formats plus registry and error types.

---

## Tasks

### Task 1: Define Wire Interface

**File:** `toolprotocol/wire/wire.go`

```go
package wire

import (
    "context"
    "github.com/ApertureStack/toolfoundation/model"
)

// Wire encodes/decodes protocol messages.
type Wire interface {
    // Name returns the protocol name.
    Name() string

    // Version returns the protocol version.
    Version() string

    // EncodeRequest encodes a tool request.
    EncodeRequest(ctx context.Context, req *ToolRequest) ([]byte, error)

    // DecodeRequest decodes a tool request.
    DecodeRequest(ctx context.Context, data []byte) (*ToolRequest, error)

    // EncodeResponse encodes a tool response.
    EncodeResponse(ctx context.Context, resp *ToolResponse) ([]byte, error)

    // DecodeResponse decodes a tool response.
    DecodeResponse(ctx context.Context, data []byte) (*ToolResponse, error)

    // EncodeToolList encodes a tool list.
    EncodeToolList(ctx context.Context, tools []model.Tool) ([]byte, error)

    // DecodeToolList decodes a tool list.
    DecodeToolList(ctx context.Context, data []byte) ([]model.Tool, error)
}

// ToolRequest represents a canonical tool call request.
type ToolRequest struct {
    ID        string
    ToolID    string
    Arguments map[string]any
    Context   map[string]any
}

// ToolResponse represents a canonical tool call response.
type ToolResponse struct {
    ID      string
    Content []Content
    IsError bool
    Error   *ToolError
}

// Content represents response content.
type Content struct {
    Type string // "text", "image", "resource"
    Text string
    Data []byte
    URI  string
}

// ToolError represents a tool execution error.
type ToolError struct {
    Code    int
    Message string
    Data    any
}
```

### Task 2: Implement MCP Wire

**File:** `toolprotocol/wire/mcp.go`

```go
package wire

import (
    "context"
    "encoding/json"

    "github.com/ApertureStack/toolfoundation/model"
)

// MCPWire implements Wire for MCP protocol.
type MCPWire struct {
    version string
}

// NewMCPWire creates a new MCP wire adapter.
func NewMCPWire() *MCPWire {
    return &MCPWire{version: "2024-11-05"}
}

func (w *MCPWire) Name() string    { return "mcp" }
func (w *MCPWire) Version() string { return w.version }

// MCP JSON-RPC structures
type mcpRequest struct {
    JSONRPC string         `json:"jsonrpc"`
    ID      any            `json:"id"`
    Method  string         `json:"method"`
    Params  map[string]any `json:"params,omitempty"`
}

type mcpResponse struct {
    JSONRPC string    `json:"jsonrpc"`
    ID      any       `json:"id"`
    Result  any       `json:"result,omitempty"`
    Error   *mcpError `json:"error,omitempty"`
}

type mcpError struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Data    any    `json:"data,omitempty"`
}

func (w *MCPWire) EncodeRequest(ctx context.Context, req *ToolRequest) ([]byte, error) {
    mcpReq := mcpRequest{
        JSONRPC: "2.0",
        ID:      req.ID,
        Method:  "tools/call",
        Params: map[string]any{
            "name":      req.ToolID,
            "arguments": req.Arguments,
        },
    }
    return json.Marshal(mcpReq)
}

func (w *MCPWire) DecodeRequest(ctx context.Context, data []byte) (*ToolRequest, error) {
    var mcpReq mcpRequest
    if err := json.Unmarshal(data, &mcpReq); err != nil {
        return nil, err
    }

    req := &ToolRequest{
        ID:        fmt.Sprintf("%v", mcpReq.ID),
        Arguments: make(map[string]any),
    }

    if params, ok := mcpReq.Params["name"].(string); ok {
        req.ToolID = params
    }
    if args, ok := mcpReq.Params["arguments"].(map[string]any); ok {
        req.Arguments = args
    }

    return req, nil
}

func (w *MCPWire) EncodeResponse(ctx context.Context, resp *ToolResponse) ([]byte, error) {
    mcpResp := mcpResponse{
        JSONRPC: "2.0",
        ID:      resp.ID,
    }

    if resp.IsError {
        mcpResp.Error = &mcpError{
            Code:    resp.Error.Code,
            Message: resp.Error.Message,
            Data:    resp.Error.Data,
        }
    } else {
        // Convert content to MCP format
        content := make([]map[string]any, len(resp.Content))
        for i, c := range resp.Content {
            content[i] = map[string]any{
                "type": c.Type,
                "text": c.Text,
            }
        }
        mcpResp.Result = map[string]any{
            "content": content,
            "isError": false,
        }
    }

    return json.Marshal(mcpResp)
}

func (w *MCPWire) DecodeResponse(ctx context.Context, data []byte) (*ToolResponse, error) {
    var mcpResp mcpResponse
    if err := json.Unmarshal(data, &mcpResp); err != nil {
        return nil, err
    }

    resp := &ToolResponse{
        ID: fmt.Sprintf("%v", mcpResp.ID),
    }

    if mcpResp.Error != nil {
        resp.IsError = true
        resp.Error = &ToolError{
            Code:    mcpResp.Error.Code,
            Message: mcpResp.Error.Message,
            Data:    mcpResp.Error.Data,
        }
    } else if result, ok := mcpResp.Result.(map[string]any); ok {
        if content, ok := result["content"].([]any); ok {
            for _, c := range content {
                if cm, ok := c.(map[string]any); ok {
                    resp.Content = append(resp.Content, Content{
                        Type: cm["type"].(string),
                        Text: cm["text"].(string),
                    })
                }
            }
        }
    }

    return resp, nil
}

func (w *MCPWire) EncodeToolList(ctx context.Context, tools []model.Tool) ([]byte, error) {
    mcpTools := make([]map[string]any, len(tools))
    for i, tool := range tools {
        mcpTools[i] = map[string]any{
            "name":        tool.ID,
            "description": tool.Description,
            "inputSchema": tool.InputSchema,
        }
    }

    return json.Marshal(mcpResponse{
        JSONRPC: "2.0",
        Result: map[string]any{
            "tools": mcpTools,
        },
    })
}

func (w *MCPWire) DecodeToolList(ctx context.Context, data []byte) ([]model.Tool, error) {
    var mcpResp mcpResponse
    if err := json.Unmarshal(data, &mcpResp); err != nil {
        return nil, err
    }

    result, ok := mcpResp.Result.(map[string]any)
    if !ok {
        return nil, nil
    }

    toolList, ok := result["tools"].([]any)
    if !ok {
        return nil, nil
    }

    tools := make([]model.Tool, len(toolList))
    for i, t := range toolList {
        tm := t.(map[string]any)
        tools[i] = model.Tool{
            ID:          tm["name"].(string),
            Name:        tm["name"].(string),
            Description: tm["description"].(string),
        }
    }

    return tools, nil
}
```

### Task 3: Implement A2A Wire

**File:** `toolprotocol/wire/a2a.go`

```go
package wire

import (
    "context"
    "encoding/json"

    "github.com/ApertureStack/toolfoundation/model"
)

// A2AWire implements Wire for Google A2A protocol.
type A2AWire struct {
    version string
}

// NewA2AWire creates a new A2A wire adapter.
func NewA2AWire() *A2AWire {
    return &A2AWire{version: "1.0"}
}

func (w *A2AWire) Name() string    { return "a2a" }
func (w *A2AWire) Version() string { return w.version }

// A2A uses different message structures focused on agent-to-agent communication
type a2aMessage struct {
    ID        string         `json:"id"`
    Type      string         `json:"type"` // "task", "result", "error"
    AgentID   string         `json:"agent_id"`
    TaskID    string         `json:"task_id,omitempty"`
    Action    string         `json:"action,omitempty"`
    Input     map[string]any `json:"input,omitempty"`
    Output    any            `json:"output,omitempty"`
    Error     *a2aError      `json:"error,omitempty"`
}

type a2aError struct {
    Type    string `json:"type"`
    Message string `json:"message"`
}

func (w *A2AWire) EncodeRequest(ctx context.Context, req *ToolRequest) ([]byte, error) {
    msg := a2aMessage{
        ID:     req.ID,
        Type:   "task",
        Action: req.ToolID,
        Input:  req.Arguments,
    }
    return json.Marshal(msg)
}

func (w *A2AWire) DecodeRequest(ctx context.Context, data []byte) (*ToolRequest, error) {
    var msg a2aMessage
    if err := json.Unmarshal(data, &msg); err != nil {
        return nil, err
    }

    return &ToolRequest{
        ID:        msg.ID,
        ToolID:    msg.Action,
        Arguments: msg.Input,
    }, nil
}

func (w *A2AWire) EncodeResponse(ctx context.Context, resp *ToolResponse) ([]byte, error) {
    msg := a2aMessage{
        ID:   resp.ID,
        Type: "result",
    }

    if resp.IsError {
        msg.Type = "error"
        msg.Error = &a2aError{
            Type:    "execution_error",
            Message: resp.Error.Message,
        }
    } else {
        // Combine content into output
        var text string
        for _, c := range resp.Content {
            text += c.Text
        }
        msg.Output = text
    }

    return json.Marshal(msg)
}

func (w *A2AWire) DecodeResponse(ctx context.Context, data []byte) (*ToolResponse, error) {
    var msg a2aMessage
    if err := json.Unmarshal(data, &msg); err != nil {
        return nil, err
    }

    resp := &ToolResponse{ID: msg.ID}

    if msg.Error != nil {
        resp.IsError = true
        resp.Error = &ToolError{
            Code:    -1,
            Message: msg.Error.Message,
        }
    } else if text, ok := msg.Output.(string); ok {
        resp.Content = []Content{{Type: "text", Text: text}}
    }

    return resp, nil
}

func (w *A2AWire) EncodeToolList(ctx context.Context, tools []model.Tool) ([]byte, error) {
    a2aTools := make([]map[string]any, len(tools))
    for i, tool := range tools {
        a2aTools[i] = map[string]any{
            "action":      tool.ID,
            "description": tool.Description,
            "parameters":  tool.InputSchema,
        }
    }
    return json.Marshal(map[string]any{"actions": a2aTools})
}

func (w *A2AWire) DecodeToolList(ctx context.Context, data []byte) ([]model.Tool, error) {
    var result map[string]any
    if err := json.Unmarshal(data, &result); err != nil {
        return nil, err
    }

    actions, ok := result["actions"].([]any)
    if !ok {
        return nil, nil
    }

    tools := make([]model.Tool, len(actions))
    for i, a := range actions {
        am := a.(map[string]any)
        tools[i] = model.Tool{
            ID:          am["action"].(string),
            Name:        am["action"].(string),
            Description: am["description"].(string),
        }
    }
    return tools, nil
}
```

### Task 4: Create Package Documentation and Commit

```bash
cd /tmp/migration/toolprotocol

git add -A
git commit -m "feat(wire): add protocol wire adapters

Create wire package for protocol encoding/decoding.

Package contents:
- Wire interface for protocol adapters
- MCPWire for MCP JSON-RPC 2.0
- A2AWire for Google A2A protocol
- ACPWire for IBM ACP protocol
- Canonical request/response types

Features:
- Protocol-agnostic tool requests
- Bidirectional conversion
- Tool list encoding
- Error propagation

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"

git push origin main
```

---

## Next Steps

- PRD-172: Create tooldiscover
- PRD-173: Create toolcontent
