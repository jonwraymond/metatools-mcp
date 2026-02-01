# Architecture

`metatools-mcp` composes the core libraries and exposes a small MCP tool surface.

## Transport layer

The transport layer abstracts how clients connect to the MCP server. All transports
implement the `Transport` interface, enabling protocol flexibility without changing
server logic.

```mermaid
%%{init: {'theme': 'base', 'themeVariables': {'primaryColor': '#805ad5'}}}%%
classDiagram
    class Transport {
        <<interface>>
        +Name() string
        +Info() Info
        +Serve(ctx, server) error
        +Close() error
    }

    class StdioTransport {
        +Name() "stdio"
        +Serve() stdin/stdout
    }

    class StreamableHTTPTransport {
        +Config StreamableHTTPConfig
        +Name() "streamable"
        +Serve() HTTP POST/GET/DELETE
    }

    class SSETransport {
        +Config SSEConfig
        +Name() "sse"
        +Serve() HTTP + SSE
    }

    Transport <|.. StdioTransport
    Transport <|.. StreamableHTTPTransport
    Transport <|.. SSETransport

    note for StreamableHTTPTransport "MCP spec 2025-03-26\nRecommended for HTTP"
    note for SSETransport "Deprecated"
```

| Transport | Protocol | Session | Best For |
|-----------|----------|---------|----------|
| `stdio` | stdin/stdout JSON-RPC | Implicit | Claude Desktop, local CLI |
| `streamable` | HTTP POST/GET/DELETE | Mcp-Session-Id header | Web apps, remote clients |
| `sse` | HTTP + Server-Sent Events | Cookie-based | Legacy (deprecated) |

## Component wiring


![Diagram](assets/diagrams/component-wiring.svg)

## Runtime layer (execute_code)

When built with the `toolruntime` tag, `execute_code` is backed by a runtime
that selects between an unsafe dev backend and a Docker-backed standard
backend. Docker is opt-in via `METATOOLS_RUNTIME_PROFILE=standard`.

Standard isolation can also be provided by the WASM backend when enabled
(`METATOOLS_WASM_ENABLED=true`, `METATOOLS_RUNTIME_BACKEND=wasm`). If Docker
is unavailable and WASM is enabled, the server falls back to WASM for the
standard profile.

```mermaid
%%{init: {'theme': 'base', 'themeVariables': {'primaryColor': '#2b6cb0'}}}%%
flowchart LR
    Agent["AI Agent"] --> MCP["metatools-mcp"]
    MCP --> Toolcode["toolexec/code Executor"]
    Toolcode --> Runtime["toolexec/runtime"]

    Runtime --> Dev["dev profile<br/>unsafe subprocess"]
    Runtime --> Standard["standard profile<br/>Docker sandbox"]

    style MCP fill:#2b6cb0,stroke:#2c5282
    style Toolcode fill:#6b46c1,stroke:#553c9a
    style Runtime fill:#4a5568,stroke:#2d3748
    style Dev fill:#dd6b20,stroke:#c05621
    style Standard fill:#2f855a,stroke:#276749
```

## Progressive disclosure flow


![Diagram](assets/diagrams/progressive-disclosure.svg)


## MCP tool mapping


![Diagram](assets/diagrams/mcp-tool-mapping.svg)
