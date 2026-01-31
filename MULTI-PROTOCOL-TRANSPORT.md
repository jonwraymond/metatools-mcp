# Multi-Protocol Transport Architecture

**Date:** 2026-01-30
**Purpose:** Design a generic transport layer supporting MCP, A2A, and other agent protocols.

---

## Executive Summary

The AI agent protocol landscape has evolved beyond MCP to include **four major standards**:

| Protocol | Developer | Focus | Transport |
|----------|-----------|-------|-----------|
| **MCP** | Anthropic | AI ↔ Tools/Resources | JSON-RPC 2.0 over HTTP/Stdio/SSE |
| **A2A** | Google | Agent ↔ Agent | HTTP/SSE with Agent Cards |
| **ACP** | IBM | Agent Communication | RESTful HTTP, multipart MIME |
| **ANP** | Community | Decentralized Agents | HTTP/JSON-LD with DIDs |

These protocols are **complementary, not competing**:
- **MCP** = Tool/Resource integration layer
- **A2A** = Enterprise agent collaboration
- **ACP** = Brokered agent orchestration
- **ANP** = Open internet agent discovery

**Recommendation:** Add `tooltransport` to `toolprotocol` with a **multi-protocol transport abstraction** that can adapt to any JSON-RPC or REST-based agent protocol.

---

## Protocol Technical Analysis

### Transport Layer Comparison

| Protocol | Wire Format | Streaming | Discovery | Security |
|----------|-------------|-----------|-----------|----------|
| **MCP** | JSON-RPC 2.0 | SSE | Static/Manual | TLS, Bearer tokens |
| **A2A** | JSON-RPC 2.0 | SSE + Push | Agent Cards (HTTP GET) | TLS, JWS signatures |
| **ACP** | REST/JSON | Multipart streams | Registry-based | mTLS, tokens |
| **ANP** | JSON-LD | Negotiated | DID + Search engines | DIDs, DNS validation |

### Message Format Patterns

All four protocols converge on similar patterns:

```
┌─────────────────────────────────────────────────────────────────────┐
│                      Common Message Pattern                          │
├─────────────────────────────────────────────────────────────────────┤
│  • Request/Response envelope (JSON-based)                            │
│  • Method/Action identifier                                          │
│  • Typed parameters/arguments                                        │
│  • Result or Error response                                          │
│  • Optional streaming for long-running operations                    │
│  • Capability/Schema discovery endpoint                              │
└─────────────────────────────────────────────────────────────────────┘
```

### Protocol Layering Strategy

Research confirms these protocols can be **layered**:

```
┌─────────────────────────────────────────────────────────────────────┐
│                      Agent Application Layer                         │
│         (toolskill, toolset, toolrun - Protocol Agnostic)           │
├─────────────────────────────────────────────────────────────────────┤
│                     Protocol Adapter Layer                           │
│    ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐              │
│    │   MCP   │  │   A2A   │  │   ACP   │  │   ANP   │              │
│    │ Adapter │  │ Adapter │  │ Adapter │  │ Adapter │              │
│    └────┬────┘  └────┬────┘  └────┬────┘  └────┬────┘              │
├─────────┼───────────┼───────────┼───────────┼──────────────────────┤
│         │           │           │           │                       │
│         └───────────┴─────┬─────┴───────────┘                       │
│                           │                                          │
│                 Generic Transport Layer                              │
│    ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐              │
│    │  HTTP   │  │  gRPC   │  │Websocket│  │  Stdio  │              │
│    └─────────┘  └─────────┘  └─────────┘  └─────────┘              │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Proposed tooltransport Design

### Core Abstraction

The key insight is separating **wire transport** from **protocol semantics**:

```go
// tooltransport/transport.go

// Transport handles raw message delivery (wire layer)
type Transport interface {
    // Metadata
    Name() string
    Info() TransportInfo

    // Lifecycle
    Start(ctx context.Context) error
    Close() error

    // Message handling
    Send(ctx context.Context, msg []byte) error
    Receive(ctx context.Context) ([]byte, error)

    // Streaming (optional)
    Stream(ctx context.Context) (MessageStream, error)
}

// TransportInfo describes transport configuration
type TransportInfo struct {
    Name     string            // "http", "grpc", "websocket", "stdio"
    Address  string            // Endpoint address
    Path     string            // Request path (HTTP)
    Secure   bool              // TLS enabled
    Metadata map[string]string // Protocol-specific metadata
}

// MessageStream for bidirectional streaming
type MessageStream interface {
    Send(msg []byte) error
    Recv() ([]byte, error)
    Close() error
}
```

### Protocol Adapter Interface

Adapters translate between protocol-specific formats and canonical tool operations:

```go
// tooltransport/adapter.go

// ProtocolAdapter translates protocol messages to canonical operations
type ProtocolAdapter interface {
    // Protocol identification
    Name() string              // "mcp", "a2a", "acp", "anp"
    Version() string           // Protocol version

    // Message translation
    DecodeRequest(raw []byte) (*CanonicalRequest, error)
    EncodeResponse(resp *CanonicalResponse) ([]byte, error)

    // Discovery
    DiscoveryEndpoint() string      // e.g., "/.well-known/agent.json"
    ParseCapabilities(raw []byte) (*Capabilities, error)

    // Error handling
    EncodeError(err error) ([]byte, error)
}

// CanonicalRequest represents a protocol-agnostic request
type CanonicalRequest struct {
    ID        string                 // Request ID
    Method    string                 // Operation name
    Params    map[string]any         // Arguments
    Metadata  map[string]string      // Headers, auth, etc.
    Streaming bool                   // Expects streaming response
}

// CanonicalResponse represents a protocol-agnostic response
type CanonicalResponse struct {
    ID       string                 // Correlates to request
    Result   any                    // Success result
    Error    *CanonicalError        // Error if failed
    Metadata map[string]string      // Response headers
}
```

### Server Abstraction

A unified server that can serve multiple protocols:

```go
// tooltransport/server.go

// Server serves tool capabilities over multiple transports
type Server interface {
    // Registration
    RegisterTransport(t Transport) error
    RegisterAdapter(a ProtocolAdapter) error
    RegisterHandler(h RequestHandler) error

    // Lifecycle
    Start(ctx context.Context) error
    Shutdown(ctx context.Context) error

    // Introspection
    Transports() []Transport
    Adapters() []ProtocolAdapter
}

// RequestHandler processes canonical requests
type RequestHandler interface {
    Handle(ctx context.Context, req *CanonicalRequest) (*CanonicalResponse, error)
    Capabilities() *Capabilities
}

// Capabilities describes what the server can do
type Capabilities struct {
    Tools     []toolmodel.Tool       // Available tools
    Resources []Resource             // MCP resources
    Prompts   []Prompt               // MCP prompts
    Skills    []Skill                // A2A skills/capabilities
}
```

---

## Transport Implementations

### HTTP Transport

Supports MCP, A2A, ACP via HTTP:

```go
// tooltransport/http/transport.go

type HTTPTransport struct {
    config     HTTPConfig
    server     *http.Server
    handler    http.Handler
    messageCh  chan []byte
}

type HTTPConfig struct {
    Address     string
    Path        string          // Default: "/api/v1"
    TLS         *tls.Config
    CORS        CORSConfig
    Timeout     time.Duration
    // Protocol-specific
    SSEEnabled  bool            // For MCP streaming
    WebhookURL  string          // For A2A callbacks
}

func (t *HTTPTransport) Name() string { return "http" }

func (t *HTTPTransport) Start(ctx context.Context) error {
    // Start HTTP server with SSE support
}
```

### gRPC Transport

For high-performance binary communication:

```go
// tooltransport/grpc/transport.go

type GRPCTransport struct {
    config   GRPCConfig
    server   *grpc.Server
    listener net.Listener
}

type GRPCConfig struct {
    Address     string
    TLS         *tls.Config
    MaxMsgSize  int
    Reflection  bool            // Enable gRPC reflection
}

func (t *GRPCTransport) Name() string { return "grpc" }
```

### WebSocket Transport

For persistent bidirectional connections:

```go
// tooltransport/websocket/transport.go

type WebSocketTransport struct {
    config   WSConfig
    upgrader websocket.Upgrader
    conns    map[string]*websocket.Conn
    mu       sync.RWMutex
}

type WSConfig struct {
    Address     string
    Path        string          // Default: "/ws"
    TLS         *tls.Config
    PingInterval time.Duration
    MaxConns    int
}

func (t *WebSocketTransport) Name() string { return "websocket" }
```

### Stdio Transport

For local tool servers (MCP primary use case):

```go
// tooltransport/stdio/transport.go

type StdioTransport struct {
    config   StdioConfig
    reader   *bufio.Reader
    writer   *bufio.Writer
}

type StdioConfig struct {
    Input   io.Reader       // Default: os.Stdin
    Output  io.Writer       // Default: os.Stdout
    Framing FramingType     // Line-delimited or length-prefixed
}

func (t *StdioTransport) Name() string { return "stdio" }
```

---

## Protocol Adapter Implementations

### MCP Adapter

```go
// tooltransport/adapter/mcp/adapter.go

type MCPAdapter struct {
    version string  // "2025-11-25"
}

func (a *MCPAdapter) Name() string { return "mcp" }
func (a *MCPAdapter) Version() string { return a.version }
func (a *MCPAdapter) DiscoveryEndpoint() string { return "" } // MCP uses static config

func (a *MCPAdapter) DecodeRequest(raw []byte) (*CanonicalRequest, error) {
    var rpc jsonrpc.Request
    if err := json.Unmarshal(raw, &rpc); err != nil {
        return nil, err
    }
    return &CanonicalRequest{
        ID:     rpc.ID,
        Method: rpc.Method, // "tools/call", "resources/read", etc.
        Params: rpc.Params.(map[string]any),
    }, nil
}
```

### A2A Adapter

```go
// tooltransport/adapter/a2a/adapter.go

type A2AAdapter struct {
    version string  // "1.0"
}

func (a *A2AAdapter) Name() string { return "a2a" }
func (a *A2AAdapter) Version() string { return a.version }
func (a *A2AAdapter) DiscoveryEndpoint() string { return "/.well-known/agent.json" }

func (a *A2AAdapter) DecodeRequest(raw []byte) (*CanonicalRequest, error) {
    var task A2ATask
    if err := json.Unmarshal(raw, &task); err != nil {
        return nil, err
    }
    return &CanonicalRequest{
        ID:        task.ID,
        Method:    "task/execute", // A2A is task-centric
        Params:    map[string]any{"skill": task.Skill, "input": task.Input},
        Streaming: task.StreamingHint,
    }, nil
}

// A2A Agent Card for discovery
type AgentCard struct {
    Name         string   `json:"name"`
    Description  string   `json:"description"`
    URL          string   `json:"url"`
    Version      string   `json:"version"`
    Skills       []Skill  `json:"skills"`
    AuthRequired bool     `json:"authRequired"`
    Protocols    []string `json:"protocols"` // ["a2a", "mcp"]
}
```

### ACP Adapter

```go
// tooltransport/adapter/acp/adapter.go

type ACPAdapter struct {
    version  string
    registry string  // Registry URL
}

func (a *ACPAdapter) Name() string { return "acp" }
func (a *ACPAdapter) DiscoveryEndpoint() string { return "/agent-detail" }

func (a *ACPAdapter) DecodeRequest(raw []byte) (*CanonicalRequest, error) {
    var msg ACPMessage
    if err := json.Unmarshal(raw, &msg); err != nil {
        return nil, err
    }
    return &CanonicalRequest{
        ID:     msg.ConversationID,
        Method: msg.Intent,
        Params: msg.Content,
    }, nil
}
```

### ANP Adapter

```go
// tooltransport/adapter/anp/adapter.go

type ANPAdapter struct {
    didResolver DIDResolver
}

func (a *ANPAdapter) Name() string { return "anp" }
func (a *ANPAdapter) DiscoveryEndpoint() string { return "/.well-known/did.json" }

func (a *ANPAdapter) DecodeRequest(raw []byte) (*CanonicalRequest, error) {
    var ld map[string]any
    if err := json.Unmarshal(raw, &ld); err != nil {
        return nil, err
    }
    // Parse JSON-LD and extract action
    action := ld["@type"].(string)
    return &CanonicalRequest{
        ID:     ld["@id"].(string),
        Method: action,
        Params: ld,
    }, nil
}
```

---

## Cross-Protocol Interfaces

Research reveals significant overlap across MCP, A2A, and ACP. These can be abstracted into **protocol-agnostic interfaces**:

### Feature Mapping

| Concept | MCP | A2A | ACP | Generic Interface |
|---------|-----|-----|-----|-------------------|
| **Discovery** | tools/list, resources/list | Agent Cards | Agent Detail | `Discoverable` |
| **Content** | TextContent, ImageContent | TextPart, FilePart, DataPart | MessagePart | `ContentPart` |
| **Execution** | tools/call (sync) | Task lifecycle (stateful) | Sync/Async | `Task` |
| **Streaming** | SSE | SSE + Push | Polling/Subscribe | `UpdateChannel` |
| **Sessions** | Session + roots | Task context | Distributed sessions | `Session` |
| **User Input** | elicitation/create | input-required state | Message threading | `Elicitation` |

### tooldiscover - Capability Discovery

```go
// toolprotocol/discover/discover.go

// Discoverable represents any agent/server that can advertise capabilities
type Discoverable interface {
    // Identity
    Name() string
    Description() string
    Version() string

    // Capabilities
    Capabilities() *Capabilities

    // Protocol-specific discovery
    DiscoveryEndpoint() string  // "/.well-known/agent.json", "", etc.
}

// Capabilities is the unified capability model
type Capabilities struct {
    Tools     []Tool     // Executable functions
    Resources []Resource // Readable data (MCP)
    Prompts   []Prompt   // Templates (MCP)
    Skills    []Skill    // Agent skills (A2A)
    Protocols []string   // Supported protocols ["mcp", "a2a"]
}

// AgentCard generates A2A-compatible Agent Card
func (c *Capabilities) ToAgentCard() *a2a.AgentCard

// ToolsList generates MCP-compatible tools/list response
func (c *Capabilities) ToMCPToolsList() *mcp.ListToolsResult
```

### toolcontent - Content/Part Abstraction

```go
// toolprotocol/content/content.go

// ContentPart is the unified content model
type ContentPart interface {
    Type() ContentType
    Data() []byte
    Metadata() map[string]any
}

type ContentType string

const (
    ContentText   ContentType = "text"
    ContentFile   ContentType = "file"
    ContentImage  ContentType = "image"
    ContentAudio  ContentType = "audio"
    ContentData   ContentType = "data"   // Structured JSON
)

// TextPart for text content
type TextPart struct {
    Text     string
    MimeType string  // "text/plain", "text/markdown"
}

// FilePart for file content
type FilePart struct {
    Name     string
    Data     []byte
    MimeType string
    URI      string  // Optional: file://, https://
}

// DataPart for structured data
type DataPart struct {
    Schema string         // JSON Schema reference
    Data   map[string]any
}

// Protocol conversions
func (p *TextPart) ToMCP() *mcp.TextContent
func (p *TextPart) ToA2A() *a2a.TextPart
func (p *FilePart) ToMCP() *mcp.EmbeddedResource
func (p *FilePart) ToA2A() *a2a.FilePart
```

### tooltask - Task Lifecycle

```go
// toolprotocol/task/task.go

// Task represents a unit of work (extends toolrun concepts)
type Task struct {
    ID          string
    Status      TaskStatus
    Input       []ContentPart
    Output      []ContentPart  // Artifacts
    Error       *TaskError
    CreatedAt   time.Time
    UpdatedAt   time.Time
    Metadata    map[string]any
}

type TaskStatus string

const (
    TaskSubmitted     TaskStatus = "submitted"
    TaskWorking       TaskStatus = "working"
    TaskInputRequired TaskStatus = "input-required"  // Needs user input
    TaskCompleted     TaskStatus = "completed"
    TaskFailed        TaskStatus = "failed"
    TaskCanceled      TaskStatus = "canceled"
)

// TaskManager handles task lifecycle
type TaskManager interface {
    Submit(ctx context.Context, input []ContentPart) (*Task, error)
    Get(ctx context.Context, id string) (*Task, error)
    Cancel(ctx context.Context, id string) error
    Subscribe(ctx context.Context, id string) (<-chan TaskUpdate, error)
}

// Integration with toolrun
func (t *Task) ToToolRunRequest() *toolrun.Request
func TaskFromToolRunResult(result *toolrun.Result) *Task
```

### toolstream - Streaming/Updates

```go
// toolprotocol/stream/stream.go

// UpdateChannel abstracts different streaming mechanisms
type UpdateChannel interface {
    Send(update Update) error
    Close() error
}

type Update struct {
    Type    UpdateType
    TaskID  string
    Payload any
    Time    time.Time
}

type UpdateType string

const (
    UpdateProgress UpdateType = "progress"
    UpdatePartial  UpdateType = "partial"   // Streaming content
    UpdateComplete UpdateType = "complete"
    UpdateError    UpdateType = "error"
)

// SSE implementation (MCP, A2A)
type SSEChannel struct { ... }

// Webhook implementation (A2A push notifications)
type WebhookChannel struct { ... }

// Polling implementation (ACP async)
type PollingChannel struct { ... }
```

### toolsession - Session/Context

```go
// toolprotocol/session/session.go

// Session maintains conversation/task context
type Session struct {
    ID        string
    ClientID  string
    CreatedAt time.Time
    ExpiresAt time.Time
    Context   *SessionContext
}

type SessionContext struct {
    // MCP roots
    Roots     []Root

    // Conversation history (ACP)
    Messages  []Message

    // Shared resources (ACP distributed sessions)
    Resources map[string]string  // URI -> content hash

    // Custom metadata
    Metadata  map[string]any
}

// SessionManager handles session lifecycle
type SessionManager interface {
    Create(ctx context.Context) (*Session, error)
    Get(ctx context.Context, id string) (*Session, error)
    Update(ctx context.Context, session *Session) error
    Delete(ctx context.Context, id string) error
}
```

### toolelicit - User Input Elicitation

```go
// toolprotocol/elicit/elicit.go

// Elicitor requests user input during execution
type Elicitor interface {
    Elicit(ctx context.Context, req *ElicitRequest) (*ElicitResponse, error)
}

type ElicitRequest struct {
    Message string
    Schema  *jsonschema.Schema  // Expected response format
    Options []ElicitOption      // Predefined choices
    Timeout time.Duration
}

type ElicitOption struct {
    Label string
    Value any
}

type ElicitResponse struct {
    Action  ElicitAction  // accept, decline, cancel
    Content any           // User's response
}

type ElicitAction string

const (
    ElicitAccept  ElicitAction = "accept"
    ElicitDecline ElicitAction = "decline"
    ElicitCancel  ElicitAction = "cancel"
)

// Protocol mapping
// MCP: elicitation/create
// A2A: Task with status "input-required"
// ACP: Message with input request type
```

---

## Integration with metatools-mcp

### Current Architecture

```
metatools-mcp/internal/transport/
├── transport.go       # Transport interface (MCP-specific)
├── stdio.go           # StdioTransport
├── sse.go             # SSETransport
└── http.go            # StreamableHTTPTransport
```

### Proposed Architecture

```
toolprotocol/
├── transport/                    # Wire layer
│   ├── transport.go              # Transport interface
│   ├── http/                     # HTTP transport
│   ├── grpc/                     # gRPC transport
│   ├── websocket/                # WebSocket transport
│   └── stdio/                    # Stdio transport
├── wire/                         # Protocol wire adapters
│   ├── mcp/                      # MCP JSON-RPC
│   ├── a2a/                      # A2A + Agent Cards
│   ├── acp/                      # ACP REST
│   └── anp/                      # ANP JSON-LD
├── discover/                     # Capability discovery
├── content/                      # Content/Part abstraction
├── task/                         # Task lifecycle
├── stream/                       # Streaming/updates
├── session/                      # Session management
└── elicit/                       # User input elicitation

metatools-mcp/
├── internal/transport/           # MCP-specific wiring (thin)
│   └── mcp.go                    # Wraps tooltransport for MCP server
└── ...
```

---

## Multi-Protocol Server Example

```go
package main

import (
    "github.com/aperturestack/toolprotocol/transport"
    "github.com/aperturestack/toolprotocol/transport/http"
    "github.com/aperturestack/toolprotocol/transport/grpc"
    "github.com/aperturestack/toolprotocol/adapter/mcp"
    "github.com/aperturestack/toolprotocol/adapter/a2a"
)

func main() {
    // Create multi-protocol server
    server := transport.NewServer()

    // Register transports
    server.RegisterTransport(http.New(http.Config{
        Address:    ":8080",
        SSEEnabled: true,
    }))
    server.RegisterTransport(grpc.New(grpc.Config{
        Address: ":9090",
    }))

    // Register protocol adapters
    server.RegisterAdapter(mcp.New("2025-11-25"))
    server.RegisterAdapter(a2a.New("1.0"))

    // Register tool handler (from toolrun)
    server.RegisterHandler(toolHandler)

    // Start serving
    ctx := context.Background()
    if err := server.Start(ctx); err != nil {
        log.Fatal(err)
    }

    // Server now accepts:
    // - MCP requests on HTTP :8080 (JSON-RPC)
    // - A2A requests on HTTP :8080 (task lifecycle)
    // - Both on gRPC :9090
}
```

---

## Feature Comparison: Current vs. Proposed

| Feature | Current (metatools-mcp) | Proposed (toolprotocol) |
|---------|------------------------|--------------------------|
| **Protocols** | | |
| MCP Support | ✅ Complete | ✅ Via adapter |
| A2A Support | ❌ None | ✅ Via adapter |
| ACP Support | ❌ None | ✅ Via adapter |
| ANP Support | ❌ None | ⚠️ Future |
| **Transports** | | |
| HTTP Transport | ✅ SSE + Streamable | ✅ Generic HTTP |
| gRPC Transport | ❌ None | ✅ Native |
| WebSocket | ❌ None | ✅ Native |
| Stdio | ✅ MCP-specific | ✅ Generic |
| **Cross-Protocol Features** | | |
| Discovery | ✅ tools/list only | ✅ Agent Cards + tools/list |
| Content abstraction | ❌ MCP types only | ✅ Unified ContentPart |
| Task lifecycle | ❌ Sync only | ✅ Stateful tasks (A2A/ACP) |
| Streaming | ✅ SSE only | ✅ SSE, webhooks, polling |
| Sessions | ❌ None | ✅ Distributed sessions |
| User input | ❌ None | ✅ Elicitation |
| **Architecture** | | |
| Multi-protocol | ❌ MCP only | ✅ Pluggable adapters |
| Protocol negotiation | ❌ None | ✅ Via discovery |

---

## Implementation Priority

### Phase 1: Foundation (Weeks 1-2)

| Component | Effort | Value |
|-----------|--------|-------|
| tooltransport core | 3 days | Wire layer foundation |
| HTTP transport (SSE) | 3 days | MCP compatibility |
| MCP adapter | 2 days | MCP compatibility |
| toolcontent | 2 days | Content abstraction |

### Phase 2: A2A Support (Weeks 3-4)

| Component | Effort | Value |
|-----------|--------|-------|
| A2A adapter | 3 days | Cross-agent collaboration |
| tooldiscover | 3 days | Agent Cards + tools/list |
| tooltask | 3 days | Task lifecycle |
| toolstream (SSE) | 1 day | Streaming foundation |

### Phase 3: Extended Transports (Weeks 5-6)

| Component | Effort | Value |
|-----------|--------|-------|
| gRPC transport | 4 days | High-performance |
| WebSocket transport | 2 days | Real-time bidirectional |
| toolsession | 2 days | Session management |
| toolstream (webhooks) | 2 days | A2A push notifications |

### Phase 4: ACP + Elicitation (Weeks 7-8)

| Component | Effort | Value |
|-----------|--------|-------|
| ACP adapter | 3 days | IBM ecosystem |
| toolelicit | 3 days | User input requests |
| toolstream (polling) | 2 days | ACP async pattern |
| Integration tests | 2 days | Cross-protocol validation |

### Phase 5: Future (Post v1.0)

| Component | Effort | Value |
|-----------|--------|-------|
| ANP adapter | 2 weeks | Decentralized agents |
| DID support | 1 week | Decentralized identity |

**Total: 8 weeks for MCP + A2A + ACP support**

---

## Benefits

### For metatools-mcp

1. **Protocol Independence**: Backend (toolrun, toolruntime) stays agnostic
2. **Future-Proof**: New protocols added via adapters, not rewrites
3. **Interoperability**: Same tools accessible via MCP, A2A, REST
4. **Performance**: gRPC option for high-throughput scenarios

### For the Ecosystem

1. **A2A Integration**: Tools can participate in Google's agent collaboration
2. **Enterprise Ready**: Support for ACP's brokered orchestration
3. **Open Standards**: ANP support for decentralized discovery
4. **Unified Development**: One codebase, multiple protocols

---

## Connection to Existing Architecture

The multi-protocol transport fits into the established architecture:

```
                              ┌────────────────────────────┐
                              │     tooltransport          │
                              │  (HTTP, gRPC, WS, Stdio)   │
                              └────────────┬───────────────┘
                                           │
              ┌────────────────────────────┼────────────────────────────┐
              │                            │                            │
              ▼                            ▼                            ▼
     ┌────────────────┐          ┌────────────────┐          ┌────────────────┐
     │  MCP Adapter   │          │  A2A Adapter   │          │  ACP Adapter   │
     │  (JSON-RPC)    │          │  (Agent Cards) │          │  (Registry)    │
     └────────┬───────┘          └────────┬───────┘          └────────┬───────┘
              │                            │                            │
              └────────────────────────────┼────────────────────────────┘
                                           │
                                           ▼
                              ┌────────────────────────────┐
                              │   CanonicalRequest/Response │
                              │   (Protocol-agnostic)       │
                              └────────────┬───────────────┘
                                           │
                                           ▼
                              ┌────────────────────────────┐
                              │        toolrun              │
                              │   (Execution Pipeline)      │
                              └────────────┬───────────────┘
                                           │
                                           ▼
                              ┌────────────────────────────┐
                              │       toolruntime           │
                              │   (Sandbox Backends)        │
                              └────────────────────────────┘
```

---

## Sources

- [Survey of Agent Interoperability Protocols](https://arxiv.org/html/2505.02279v1) - arXiv 2025
- [AI Agent Protocols Comparison](https://dev.to/dr_hernani_costa/ai-agent-protocols-mcp-vs-a2a-vs-anp-vs-acp-3870) - DEV Community
- [MCP vs A2A Guide](https://onereach.ai/blog/guide-choosing-mcp-vs-a2a-protocols/) - OneReach.ai
- [Top AI Agent Protocols 2026](https://getstream.io/blog/ai-agent-protocols/) - GetStream
- [MCP Specification 2025-11-25](https://modelcontextprotocol.io/specification/2025-11-25)
- [A2A Protocol Specification](https://google.github.io/a2a)

---

## Next Steps

1. **Update LIBRARY-CATEGORIZATION.md** - Add tooltransport to toolprotocol
2. **Create PRD** - Detailed implementation plan for tooltransport
3. **Prototype** - Start with HTTP transport + MCP adapter
4. **Validate** - Ensure backward compatibility with existing MCP clients
