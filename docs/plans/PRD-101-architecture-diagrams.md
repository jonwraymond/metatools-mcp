# PRD-101: Architecture Diagrams

**Phase:** 0 - Planning & Documentation
**Priority:** High
**Effort:** 4 hours
**Dependencies:** PRD-100

---

## Objective

Create comprehensive architecture diagrams using Mermaid syntax to visualize the consolidated ecosystem structure, layer dependencies, and data flows.

---

## Deliverables

| Deliverable | Location | Description |
|-------------|----------|-------------|
| Layer Architecture | `docs/diagrams/layer-architecture.md` | 8-tier layer diagram |
| Repository Map | `docs/diagrams/repository-map.md` | 6 repos with packages |
| Dependency Graph | `docs/diagrams/dependency-graph.md` | Inter-repo dependencies |
| Data Flow | `docs/diagrams/data-flow.md` | Request/response flows |
| Protocol Adapters | `docs/diagrams/protocol-adapters.md` | Multi-protocol support |

---

## Tasks

### Task 1: Create Layer Architecture Diagram

**File:** `docs/diagrams/layer-architecture.md`

```markdown
# Layer Architecture

## Overview

The ApertureStack ecosystem is organized into 8 distinct tiers, each with clear responsibilities and dependencies.

## Diagram

\`\`\`mermaid
graph TB
    subgraph "Tier 8: Application"
        metatools-mcp["metatools-mcp<br/>(MCP Server)"]
    end

    subgraph "Tier 7: Protocol"
        toolprotocol["toolprotocol"]
        subgraph "toolprotocol packages"
            transport["transport"]
            wire["wire"]
            discover["discover"]
            content["content"]
            task["task"]
            stream["stream"]
            session["session"]
            elicit["elicit"]
            resource["resource"]
            prompt["prompt"]
        end
    end

    subgraph "Tier 6: Operations"
        toolops["toolops"]
        subgraph "toolops packages"
            observe["observe"]
            cache["cache"]
            resilience["resilience"]
            health["health"]
            auth["auth"]
        end
    end

    subgraph "Tier 5: Composition"
        toolcompose["toolcompose"]
        subgraph "toolcompose packages"
            set["set"]
            skill["skill"]
        end
    end

    subgraph "Tier 4: Execution"
        toolexec["toolexec"]
        subgraph "toolexec packages"
            run["run"]
            runtime["runtime"]
            code["code"]
            backend["backend"]
        end
    end

    subgraph "Tier 3: Discovery"
        tooldiscovery["tooldiscovery"]
        subgraph "tooldiscovery packages"
            index["index"]
            search["search"]
            semantic["semantic"]
            docs["docs"]
        end
    end

    subgraph "Tier 2: Foundation"
        toolfoundation["toolfoundation"]
        subgraph "toolfoundation packages"
            model["model"]
            adapter["adapter"]
            version["version"]
        end
    end

    subgraph "Tier 1: External"
        mcp-sdk["MCP Go SDK"]
    end

    metatools-mcp --> toolprotocol
    metatools-mcp --> toolops
    metatools-mcp --> toolcompose
    metatools-mcp --> toolexec
    metatools-mcp --> tooldiscovery

    toolprotocol --> toolfoundation
    toolops --> toolfoundation
    toolcompose --> toolexec
    toolcompose --> tooldiscovery
    toolexec --> toolfoundation
    tooldiscovery --> toolfoundation
    toolfoundation --> mcp-sdk
\`\`\`

## Layer Descriptions

| Tier | Repository | Purpose |
|------|------------|---------|
| 8 | metatools-mcp | MCP server exposing metatools |
| 7 | toolprotocol | Multi-protocol transport and wire adapters |
| 6 | toolops | Cross-cutting operational concerns |
| 5 | toolcompose | Tool composition and agent skills |
| 4 | toolexec | Tool execution and runtime |
| 3 | tooldiscovery | Tool registry and search |
| 2 | toolfoundation | Core schemas and adapters |
| 1 | (external) | MCP Go SDK |
\`\`\`

### Task 2: Create Repository Map

**File:** `docs/diagrams/repository-map.md`

```markdown
# Repository Map

## Consolidated Structure

\`\`\`mermaid
graph LR
    subgraph "toolfoundation"
        tf-model["model/"]
        tf-adapter["adapter/"]
        tf-version["version/"]
    end

    subgraph "tooldiscovery"
        td-index["index/"]
        td-search["search/"]
        td-semantic["semantic/"]
        td-docs["docs/"]
    end

    subgraph "toolexec"
        te-run["run/"]
        te-runtime["runtime/"]
        te-code["code/"]
        te-backend["backend/"]
    end

    subgraph "toolcompose"
        tc-set["set/"]
        tc-skill["skill/"]
    end

    subgraph "toolops"
        to-observe["observe/"]
        to-cache["cache/"]
        to-resilience["resilience/"]
        to-health["health/"]
        to-auth["auth/"]
    end

    subgraph "toolprotocol"
        tp-transport["transport/"]
        tp-wire["wire/"]
        tp-discover["discover/"]
        tp-content["content/"]
        tp-task["task/"]
        tp-stream["stream/"]
        tp-session["session/"]
        tp-elicit["elicit/"]
        tp-resource["resource/"]
        tp-prompt["prompt/"]
    end
\`\`\`

## Package Counts

| Repository | Packages | Lines (est.) |
|------------|----------|--------------|
| toolfoundation | 3 | ~8,000 |
| tooldiscovery | 4 | ~10,000 |
| toolexec | 4 | ~15,000 |
| toolcompose | 2 | ~5,000 |
| toolops | 5 | ~12,000 |
| toolprotocol | 10 | ~20,000 |
| **Total** | **28** | **~70,000** |
```

### Task 3: Create Dependency Graph

**File:** `docs/diagrams/dependency-graph.md`

```markdown
# Dependency Graph

## Inter-Repository Dependencies

\`\`\`mermaid
graph TD
    metatools-mcp --> toolprotocol
    metatools-mcp --> toolops
    metatools-mcp --> toolcompose
    metatools-mcp --> toolexec
    metatools-mcp --> tooldiscovery
    metatools-mcp --> toolfoundation

    toolprotocol --> toolfoundation
    toolprotocol --> toolops

    toolops --> toolfoundation

    toolcompose --> toolexec
    toolcompose --> tooldiscovery
    toolcompose --> toolfoundation

    toolexec --> toolfoundation

    tooldiscovery --> toolfoundation

    toolfoundation --> mcp-sdk["MCP Go SDK"]
\`\`\`

## Dependency Matrix

|  | foundation | discovery | exec | compose | ops | protocol |
|--|------------|-----------|------|---------|-----|----------|
| **foundation** | - | | | | | |
| **discovery** | ✓ | - | | | | |
| **exec** | ✓ | | - | | | |
| **compose** | ✓ | ✓ | ✓ | - | | |
| **ops** | ✓ | | | | - | |
| **protocol** | ✓ | | | | ✓ | - |
| **metatools-mcp** | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |

## Build Order

Based on dependencies, build order is:
1. toolfoundation (no internal deps)
2. tooldiscovery, toolexec, toolops (parallel, depend on foundation)
3. toolcompose (depends on discovery + exec)
4. toolprotocol (depends on foundation + ops)
5. metatools-mcp (depends on all)
```

### Task 4: Create Data Flow Diagram

**File:** `docs/diagrams/data-flow.md`

```markdown
# Data Flow

## Tool Execution Flow

\`\`\`mermaid
sequenceDiagram
    participant Client
    participant Transport as toolprotocol/transport
    participant Wire as toolprotocol/wire
    participant Auth as toolops/auth
    participant Cache as toolops/cache
    participant Exec as toolexec/run
    participant Runtime as toolexec/runtime

    Client->>Transport: HTTP/gRPC/WebSocket Request
    Transport->>Wire: Decode Protocol (MCP/A2A/ACP)
    Wire->>Auth: Authenticate Request
    Auth->>Cache: Check Cache

    alt Cache Hit
        Cache-->>Wire: Cached Result
    else Cache Miss
        Cache->>Exec: Execute Tool
        Exec->>Runtime: Sandbox Execution
        Runtime-->>Exec: Result
        Exec->>Cache: Store Result
        Cache-->>Wire: Fresh Result
    end

    Wire->>Transport: Encode Response
    Transport-->>Client: Response
\`\`\`

## Tool Discovery Flow

\`\`\`mermaid
sequenceDiagram
    participant Client
    participant Index as tooldiscovery/index
    participant Search as tooldiscovery/search
    participant Semantic as tooldiscovery/semantic
    participant Docs as tooldiscovery/docs

    Client->>Index: List/Search Tools
    Index->>Search: BM25 Search

    opt Semantic Search Enabled
        Search->>Semantic: Vector Search
        Semantic-->>Search: Semantic Results
    end

    Search-->>Index: Ranked Results
    Index->>Docs: Get Documentation
    Docs-->>Index: Tool Docs
    Index-->>Client: Tools with Docs
\`\`\`
```

### Task 5: Create Protocol Adapters Diagram

**File:** `docs/diagrams/protocol-adapters.md`

```markdown
# Protocol Adapters

## Multi-Protocol Architecture

\`\`\`mermaid
graph TB
    subgraph "Clients"
        mcp-client["MCP Client"]
        a2a-client["A2A Agent"]
        acp-client["ACP Client"]
        http-client["HTTP Client"]
    end

    subgraph "Transport Layer"
        stdio["Stdio"]
        sse["SSE"]
        ws["WebSocket"]
        grpc["gRPC"]
        http["HTTP"]
    end

    subgraph "Wire Adapters"
        mcp-wire["MCP Wire"]
        a2a-wire["A2A Wire"]
        acp-wire["ACP Wire"]
    end

    subgraph "Canonical Layer"
        canonical["Canonical Tool Interface"]
    end

    mcp-client --> stdio
    mcp-client --> sse
    a2a-client --> grpc
    acp-client --> http
    http-client --> http

    stdio --> mcp-wire
    sse --> mcp-wire
    grpc --> a2a-wire
    http --> acp-wire

    mcp-wire --> canonical
    a2a-wire --> canonical
    acp-wire --> canonical
\`\`\`

## Protocol Feature Matrix

| Feature | MCP | A2A | ACP |
|---------|-----|-----|-----|
| Tool Discovery | ✓ | ✓ | ✓ |
| Tool Execution | ✓ | ✓ | ✓ |
| Streaming | ✓ | ✓ | ✓ |
| Resources | ✓ | - | - |
| Prompts | ✓ | - | - |
| Sessions | - | ✓ | ✓ |
| Tasks | - | ✓ | ✓ |
| Elicitation | ✓ | - | - |
```

---

## Verification Checklist

- [ ] All 5 diagram files created
- [ ] Mermaid syntax validates (no errors)
- [ ] Diagrams render in GitHub markdown preview
- [ ] All 6 repositories represented
- [ ] All 28 packages shown
- [ ] Dependency arrows are accurate
- [ ] Protocol adapters show all 3 protocols

---

## Acceptance Criteria

1. Diagrams provide clear visual understanding of architecture
2. All relationships between components are accurate
3. Diagrams can be embedded in documentation
4. Mermaid syntax is valid and renders correctly

---

## Rollback Plan

```bash
# Remove diagram files
rm -rf docs/diagrams/
```

---

## Next Steps

- PRD-102: Schema Definitions
- PRD-110: Repository Creation
