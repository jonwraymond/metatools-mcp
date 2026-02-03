# Tool Library Categorization Analysis

**Date:** 2026-01-30
**Purpose:** Consolidate existing libraries into logical groupings for easier maintenance and discovery.

---

## Executive Summary

The metatools ecosystem has grown to **13 production libraries** plus several proposed additions. This document analyzes each library and proposes a consolidation strategy organized by **area of concern**.

### Current State (Verified)

| Category | Count | Libraries |
|----------|-------|-----------|
| **Existing (Complete)** | 11 | toolmodel, tooladapter, toolindex, toolsearch, tooldocs, toolrun, toolruntime, toolcode, toolset, toolobserve, toolcache |
| **Existing (Partial)** | 2 | toolsemantic (~800 LOC), toolskill (~400 LOC) |
| **Extracted from metatools-mcp** | 2 | toolops/auth (~6,400 LOC), toolbackend (~600 LOC) |
| **Proposed (Cross-Cutting)** | 2 | toolresilience, toolhealth |
| **Proposed (Protocol Layer)** | 10 | tooltransport, toolwire, tooldiscover, toolcontent, tooltask, toolstream, toolsession, toolelicit, toolresource, toolprompt |
| **Internal (metatools-mcp)** | 2 | toolgateway, multi-tenancy |
| **Total** | 29 | All libraries accounted for |

### Actual Line Counts (Verified)

| Library | Estimated | Actual |
|---------|-----------|--------|
| toolmodel | ~2.5k | 2,221 |
| tooladapter | ~3k | 1,930 |
| toolindex | ~2k | 3,010 |
| toolrun | ~3k | 4,905 |
| toolruntime | ~4k | 1,958 |
| toolops/auth (extracted) | ~2.5k | 6,389 |
| toolbackend (internal) | ~300 | 634 |

---

## Complete Library Inventory

### Tier 1: Foundation (Data Model + Protocol)

These libraries define the canonical data structures and protocol translation layers.

| Library | Purpose | Status | Dependencies | Lines |
|---------|---------|--------|--------------|-------|
| **toolmodel** | Canonical MCP tool schema, validation, dual serialization (JSON/MsgPack) | Complete | None | ~2.5k |
| **tooladapter** | Protocol adapters (MCP, OpenAI, Anthropic, LangChain), canonical tool abstraction | Complete | toolmodel | ~3k |
| **toolversion** | Semantic versioning, protocol negotiation, deprecation tracking | Not Started | toolmodel | ~1k est |

**Consolidation Recommendation:** Merge into single **toolfoundation** repo with subpackages:
- `toolfoundation/model` (current toolmodel)
- `toolfoundation/adapter` (current tooladapter)
- `toolfoundation/version` (proposed toolversion)

**Rationale:** All three deal with fundamental tool representation - model defines schema, adapter translates protocols, version manages compatibility. Alternatively, keep toolmodel standalone (zero deps) and merge adapter+version.

---

### Tier 2: Discovery (Search + Indexing)

These libraries handle tool discovery, indexing, and search.

| Library | Purpose | Status | Dependencies | Lines |
|---------|---------|--------|--------------|-------|
| **toolindex** | In-memory registry, progressive disclosure (Summary/Schema/Full), Searcher interface | Complete | toolmodel | ~2k |
| **toolsearch** | BM25 search implementation, Bleve integration, fingerprint caching | Complete | toolindex, Bleve | ~1.5k |
| **toolsemantic** | Hybrid search (BM25 + vector), embeddings, reranking | Partial | toolindex, toolsearch | ~500 |

**Consolidation Recommendation:** Merge into single **tooldiscovery** repo with subpackages:
- `tooldiscovery/index` (current toolindex)
- `tooldiscovery/search` (current toolsearch)
- `tooldiscovery/semantic` (current toolsemantic)

**Rationale:** These are tightly coupled - search implementations plug into index via Searcher interface. Single repo simplifies versioning and cross-package testing.

---

### Tier 3: Documentation (Content + Examples)

| Library | Purpose | Status | Dependencies | Lines |
|---------|---------|--------|--------------|-------|
| **tooldocs** | Progressive documentation (Summary/Schema/Full), example storage | Complete | toolmodel | ~1.5k |

**Consolidation Recommendation:** Keep standalone or merge into tooldiscovery as `tooldiscovery/docs`.

---

### Tier 4: Execution (Core Runtime)

These libraries handle tool execution, chaining, and sandbox isolation.

| Library | Purpose | Status | Dependencies | Lines |
|---------|---------|--------|--------------|-------|
| **toolrun** | Execution pipeline, chain execution, backend dispatch, streaming | Complete | toolmodel | ~3k |
| **toolruntime** | 10 sandbox backends (docker, wasm, firecracker, etc.), security profiles | Complete | None (interfaces) | ~4k |
| **toolcode** | Code-based orchestration, toolcodeengine integration | Complete | toolrun | ~2k |

**Consolidation Recommendation:** Merge into single **toolexec** repo with subpackages:
- `toolexec/run` (current toolrun)
- `toolexec/runtime` (current toolruntime)
- `toolexec/code` (current toolcode)

**Rationale:** These form the execution core - run dispatches to runtime backends, code builds on run for orchestration. Single repo enables atomic updates when execution semantics change.

---

### Tier 5: Composition (Higher-Level Abstractions)

| Library | Purpose | Status | Dependencies | Lines |
|---------|---------|--------|--------------|-------|
| **toolset** | Composable tool collections, filtering, access control | Complete | tooladapter, toolindex | ~1k |
| **toolskill** | SKILL.md-compatible agent skills, multi-step workflows | Partial | toolset, toolrun | ~500 |

**Consolidation Recommendation:** Merge into single **toolcompose** repo:
- `toolcompose/set` (current toolset)
- `toolcompose/skill` (current toolskill)

**Rationale:** Both deal with composing tools into higher-level constructs. Skills are built on toolsets.

---

### Tier 6: Cross-Cutting (Observability + Performance)

| Library | Purpose | Status | Dependencies | Lines |
|---------|---------|--------|--------------|-------|
| **toolobserve** | OpenTelemetry tracing, metrics, structured logging | Complete | OTel SDK | ~2k |
| **toolcache** | Response caching (Memory, Redis), TTL management | Complete | None | ~1.5k |

**Proposed Libraries (from ROADMAP):**

| Library | Purpose | Status | Dependencies |
|---------|---------|--------|--------------|
| **toolresilience** | Circuit breaker, retry, bulkhead, timeout | Not Started | None |
| **toolhealth** | Health checks, readiness probes | Not Started | None |

**Consolidation Recommendation:** Merge into single **toolops** repo:
- `toolops/observe` (current toolobserve)
- `toolops/cache` (current toolcache)
- `toolops/resilience` (proposed toolresilience)
- `toolops/health` (proposed toolhealth)

**Rationale:** All cross-cutting operational concerns. These are typically configured together in production deployments.

---

### Tier 7: Protocol Layer (NEW) - `toolprotocol`

These libraries implement cross-protocol abstractions for agent interoperability.

#### Cross-Protocol Feature Mapping

Research shows significant overlap across MCP, A2A, and ACP:

| Concept | MCP | A2A | ACP | Generic Interface |
|---------|-----|-----|-----|-------------------|
| **Discovery** | tools/list, resources/list | Agent Cards (`/.well-known/agent.json`) | Agent metadata (offline) | `Discoverable` |
| **Content** | TextContent, ImageContent, EmbeddedResource | TextPart, FilePart, DataPart | MessagePart + metadata | `ContentPart` |
| **Execution** | tools/call (sync) | Task lifecycle (stateful) | Sync POST / Async taskId | `Task` |
| **Streaming** | SSE | SSE + Push notifications | Async polling/subscribe | `UpdateChannel` |
| **Sessions** | Session + roots | Task context | Distributed sessions (URI) | `Session` |
| **User Input** | elicitation/create | input-required state | Message threading | `Elicitation` |

#### Library Components

| Library | Purpose | Status | Cross-Protocol |
|---------|---------|--------|----------------|
| **tooltransport** | Wire layer (HTTP, gRPC, WS, Stdio) | Not Started | All protocols |
| **toolwire** | Protocol wire adapters (MCP JSON-RPC, A2A, ACP REST, ANP) | Not Started | All protocols |
| **tooldiscover** | Capability discovery (Agent Cards, tools/list) | Not Started | MCP, A2A, ACP |
| **toolcontent** | Content/Part abstraction (text, file, data) | Not Started | MCP, A2A, ACP |
| **tooltask** | Task lifecycle (stateful execution) | Not Started | A2A, ACP (extends toolrun) |
| **toolstream** | Streaming/updates (SSE, webhooks, polling) | Not Started | MCP, A2A, ACP |
| **toolsession** | Session/context management | Not Started | MCP, A2A, ACP |
| **toolelicit** | User input elicitation | Not Started | MCP, A2A |
| **toolresource** | Resource providers (files, DBs, APIs) | Not Started | MCP |
| **toolprompt** | Prompt templates with arguments | Not Started | MCP |

**Note:** `toolwire` is different from existing `tooladapter`:
- `tooladapter` (Tier 1) = Tool **schema** translation (MCP ↔ OpenAI ↔ Anthropic tool formats)
- `toolwire` (Tier 7) = Protocol **wire** translation (JSON-RPC ↔ REST ↔ JSON-LD)

**Consolidation Recommendation:** Create **toolprotocol** repo with subpackages:

```
toolprotocol/
├── transport/          # Wire layer (HTTP, gRPC, WebSocket, Stdio)
├── wire/               # Protocol wire adapters
│   ├── mcp/            # MCP JSON-RPC 2.0
│   ├── a2a/            # A2A + Agent Cards
│   ├── acp/            # ACP REST
│   └── anp/            # ANP JSON-LD + DIDs
├── discover/           # Cross-protocol capability discovery
├── content/            # Content/Part abstraction
├── task/               # Task lifecycle (extends toolrun)
├── stream/             # Streaming/update channels
├── session/            # Session/context management
├── elicit/             # User input elicitation
├── resource/           # MCP Resources
└── prompt/             # MCP Prompts
```

**Design Principles:**
1. **Generic First**: Core interfaces are protocol-agnostic
2. **Adapter Pattern**: Protocol-specific encoding in adapters
3. **Composable**: Features can be used independently
4. **Backward Compatible**: MCP remains primary, others are optional

See **MULTI-PROTOCOL-TRANSPORT.md** for detailed architecture.

**Rationale:** Cross-protocol abstractions enable interoperability. Same tools accessible via MCP, A2A, or REST without code changes.

---

### Tier 8: Protocol + Transport (MCP Server)

These exist within metatools-mcp rather than as separate libraries:

| Component | Purpose | Status | Location |
|-----------|---------|--------|----------|
| **Transport** | Stdio, SSE, Streamable HTTP | Complete | internal/transport |
| **Provider Registry** | Tool provider registration | Complete | internal/provider |
| **Backend Registry** | Backend registration, aggregator | Complete | toolexec/backend |
| **Middleware Chain** | Logging, metrics, auth, rate limiting | Complete | internal/middleware |
| **Auth** | JWT, API Key, RBAC, OAuth2, JWKS | Complete | toolops/auth |

**Proposed Libraries (from ROADMAP):**

| Library | Purpose | Status | Dependencies |
|---------|---------|--------|--------------|
| **toolgateway** | Auth proxy, analytics, multi-upstream routing | Not Started | All above |
| **toolresource** | MCP Resources support | Not Started | toolindex |
| **toolversion** | Semantic versioning, protocol negotiation | Not Started | toolmodel |

**Consolidation Recommendation:** Keep metatools-mcp as the MCP server integration layer. Gateway functionality could be a separate repo or part of metatools-mcp.

---

## Proposed Consolidated Structure

```
ApertureStack/
├── toolfoundation/         # Foundation Consolidated
│   ├── model/              # Canonical schema, validation (zero deps)
│   ├── adapter/            # Protocol adapters (MCP, OpenAI, Anthropic)
│   └── version/            # Semantic versioning, negotiation
│
├── tooldiscovery/          # Discovery Consolidated
│   ├── index/              # Registry, progressive disclosure
│   ├── search/             # BM25 search
│   ├── semantic/           # Hybrid search, vectors
│   └── docs/               # Documentation storage
│
├── toolexec/               # Execution Consolidated
│   ├── run/                # Execution pipeline, dispatch
│   ├── runtime/            # Sandbox backends
│   ├── code/               # Code orchestration
│   └── backend/            # Backend abstraction (extracted)
│
├── toolcompose/            # Composition Consolidated
│   ├── set/                # Toolsets, filtering
│   └── skill/              # Agent skills
│
├── toolops/                # Operations Consolidated
│   ├── observe/            # OpenTelemetry
│   ├── cache/              # Response caching
│   ├── resilience/         # Circuit breaker, retry
│   ├── health/             # Health checks
│   └── auth/               # Authentication (extracted)
│
├── toolprotocol/            # Protocol Layer (NEW) - Cross-Protocol
│   ├── transport/          # Wire layer (HTTP, gRPC, WS, Stdio)
│   ├── wire/               # Protocol wire adapters (MCP, A2A, ACP, ANP)
│   ├── discover/           # Capability discovery
│   ├── content/            # Content/Part abstraction
│   ├── task/               # Task lifecycle
│   ├── stream/             # Streaming/updates
│   ├── session/            # Session/context
│   ├── elicit/             # User input elicitation
│   ├── resource/           # MCP Resources
│   └── prompt/             # MCP Prompts
│
└── metatools-mcp/          # MCP Server (thin composition layer)
    ├── internal/transport/ # MCP transports (Stdio, SSE, HTTP)
    ├── internal/provider/  # MCP ToolProvider
    ├── internal/middleware/# MCP middleware chain
    ├── internal/handlers/  # Metatool handlers
    └── internal/server/    # Server wiring + toolgateway
```

---

## Dependency Graph (Consolidated)

```
                    ┌─────────────────────────────────────────────────────────┐
                    │                    metatools-mcp                         │
                    │   (Transport, Provider, Backend, Middleware, Auth)      │
                    │   + toolgateway, toolresource, multi-tenancy            │
                    └─────────────────────────────────────────────────────────┘
                                              │
                    ┌─────────────────────────┼─────────────────────────┐
                    │                         │                         │
                    ▼                         ▼                         ▼
           ┌───────────────┐         ┌───────────────┐         ┌───────────────┐
           │  toolcompose  │         │   toolexec    │         │    toolops    │
           │  (set, skill) │         │(run,runtime,  │         │(observe,cache,│
           └───────────────┘         │    code)      │         │resilience,    │
                    │                └───────────────┘         │health)        │
                    │                         │                └───────────────┘
                    │                         │
                    └────────────┬────────────┘
                                 │
                                 ▼
                        ┌───────────────┐
                        │ tooldiscovery │
                        │(index, search,│
                        │semantic, docs)│
                        └───────────────┘
                                 │
                                 ▼
                        ┌───────────────┐
                        │toolfoundation │
                        │(model,adapter,│
                        │   version)    │
                        └───────────────┘
```

---

## Migration Strategy

### Phase 1: No Breaking Changes (Immediate)

1. Create consolidated repos with subpackages
2. Re-export from original package paths for backward compatibility
3. Deprecate original repos with redirects

### Phase 2: Version Bump (v2)

1. Update all imports to consolidated paths
2. Remove deprecated re-exports
3. Major version bump for consolidated repos

### Phase 3: Cleanup

1. Archive original repos
2. Update documentation
3. Update ci-tools-stack version matrix

---

## Considerations

### Pros of Consolidation

| Benefit | Description |
|---------|-------------|
| **Simpler Versioning** | Fewer repos to coordinate releases |
| **Atomic Updates** | Cross-package changes in single PR |
| **Easier Discovery** | 6 repos instead of 13+ |
| **Reduced go.mod Complexity** | Fewer dependencies to manage |
| **Unified Testing** | Integration tests in same repo |

### Cons of Consolidation

| Concern | Mitigation |
|---------|------------|
| **Larger Repos** | Subpackages allow granular imports |
| **Build Times** | Go's incremental builds minimize impact |
| **Release Coupling** | SemVer subpackages if needed |
| **Historical Git** | Use `git subtree` to preserve history |

---

## Remaining ROADMAP Items

Items kept in metatools-mcp (not extracted to standalone libraries):

| Item | Recommendation |
|------|----------------|
| **Multi-Tenancy** | Keep in metatools-mcp as internal/tenant |
| **toolgateway** | Keep in metatools-mcp as internal/gateway |

**Note:** `toolresource` and `toolprompt` moved to `toolprotocol` as they're protocol-level features that could apply to multiple protocols.

---

## Summary

| Before | After | Reduction |
|--------|-------|-----------|
| 29 libraries (13 existing + 16 new/extract) | 6 consolidated repos + metatools-mcp | 79% reduction |

### Complete Library Mapping

| Consolidated Repo | Contains | Status |
|-------------------|----------|--------|
| **toolfoundation** | toolmodel + tooladapter + toolversion | 2 Complete, 1 Not Started |
| **tooldiscovery** | toolindex + toolsearch + toolsemantic + tooldocs | 3 Complete, 1 Partial |
| **toolexec** | toolrun + toolruntime + toolcode + **toolbackend** | 3 Complete, 1 Extract |
| **toolcompose** | toolset + toolskill | 1 Complete, 1 Partial |
| **toolops** | toolobserve + toolcache + toolresilience + toolhealth + **toolops/auth** | 2 Complete, 2 Not Started, 1 Extracted |
| **toolprotocol** | tooltransport + toolwire + tooldiscover + toolcontent + tooltask + toolstream + toolsession + toolelicit + toolresource + toolprompt | All Not Started |
| **metatools-mcp** | Transport, Provider, Middleware, Handlers + toolgateway, multi-tenancy | Core Complete (thin layer) |

### Extraction from metatools-mcp

| Component | Lines | Destination | Benefit |
|-----------|-------|-------------|---------|
| **toolops/auth** | 6,389 | toolops/auth | Reusable auth for any server (gRPC, REST, etc.) |
| **toolbackend** | 634 | toolexec/backend | Clean execution abstraction |
| Remaining | ~14,000 | metatools-mcp | MCP server composition layer |

### Protocol Layer (NEW) - Cross-Protocol Abstractions

| Library | Protocols | Purpose | Priority |
|---------|-----------|---------|----------|
| **tooltransport** | All | Wire layer (HTTP, gRPC, WS, Stdio) | P0 |
| **toolwire** | All | Protocol wire adapters (JSON-RPC, REST, JSON-LD) | P0 |
| **tooldiscover** | MCP, A2A, ACP | Capability discovery (Agent Cards, tools/list) | P1 |
| **toolcontent** | MCP, A2A, ACP | Content/Part abstraction (text, file, data) | P1 |
| **tooltask** | A2A, ACP | Task lifecycle (stateful execution) | P2 |
| **toolstream** | MCP, A2A, ACP | Streaming/updates (SSE, webhooks, polling) | P2 |
| **toolsession** | MCP, A2A, ACP | Session/context management | P3 |
| **toolelicit** | MCP, A2A | User input elicitation | P3 |
| **toolresource** | MCP | Resource providers (files, DBs, APIs) | P1 |
| **toolprompt** | MCP | Prompt templates with arguments | P2 |

**Note on naming:**
- `tooladapter` (Tier 1/toolfoundation) = Tool **schema** adapters (MCP ↔ OpenAI ↔ Anthropic formats)
- `toolwire` (Tier 7/toolprotocol) = Protocol **wire** adapters (JSON-RPC ↔ REST ↔ JSON-LD)

### Full Library Inventory (All 29 Libraries)

#### Tier 1: Foundation (toolfoundation)

| # | Library | Status | Lines | Notes |
|---|---------|--------|-------|-------|
| 1 | toolmodel | Complete | 2,221 | Zero deps, canonical schema |
| 2 | tooladapter | Complete | 1,930 | Tool **schema** adapters (MCP ↔ OpenAI ↔ Anthropic) |
| 3 | toolversion | Not Started | ~1k est | Semantic versioning |

#### Tier 2-3: Discovery (tooldiscovery)

| # | Library | Status | Lines | Notes |
|---|---------|--------|-------|-------|
| 4 | toolindex | Complete | 3,010 | Registry + search interface |
| 5 | toolsearch | Complete | 2,103 | BM25 implementation |
| 6 | toolsemantic | Partial | 790 | Hybrid search (interfaces only) |
| 7 | tooldocs | Complete | 2,225 | Progressive documentation |

#### Tier 4: Execution (toolexec)

| # | Library | Status | Lines | Notes |
|---|---------|--------|-------|-------|
| 8 | toolrun | Complete | 4,905 | Execution pipeline |
| 9 | toolruntime | Complete | 1,958 | 10 sandbox backends |
| 10 | toolcode | Complete | 3,389 | Code orchestration |
| 11 | **toolbackend** | **Extract** | 634 | Backend interface (from metatools-mcp) |

#### Tier 5: Composition (toolcompose)

| # | Library | Status | Lines | Notes |
|---|---------|--------|-------|-------|
| 12 | toolset | Complete | 2,542 | Toolset composition |
| 13 | toolskill | Partial | 401 | Agent skills (types only) |

#### Tier 6: Operations (toolops)

| # | Library | Status | Lines | Notes |
|---|---------|--------|-------|-------|
| 14 | toolobserve | Complete | 2,225 | OpenTelemetry |
| 15 | toolcache | Complete | 2,184 | Response caching |
| 16 | toolresilience | Not Started | ~1k est | Circuit breaker, retry |
| 17 | toolhealth | Not Started | ~500 est | Health checks |
| 18 | **toolops/auth** | **Extracted** | 6,389 | Auth/Authz (from metatools-mcp) |

#### Tier 7: Protocol Layer (toolprotocol)

| # | Library | Status | Lines | Notes |
|---|---------|--------|-------|-------|
| 19 | tooltransport | Not Started | ~1k est | Wire layer (HTTP, gRPC, WS, Stdio) |
| 20 | **toolwire** | Not Started | ~2k est | Protocol **wire** adapters (JSON-RPC, REST, JSON-LD) |
| 21 | tooldiscover | Not Started | ~1k est | Capability discovery |
| 22 | toolcontent | Not Started | ~1k est | Content/Part abstraction |
| 23 | tooltask | Not Started | ~1.5k est | Task lifecycle |
| 24 | toolstream | Not Started | ~1k est | Streaming/updates |
| 25 | toolsession | Not Started | ~800 est | Session/context |
| 26 | toolelicit | Not Started | ~500 est | User input elicitation |
| 27 | toolresource | Not Started | ~1.5k est | MCP Resources (files, DBs, APIs) |
| 28 | toolprompt | Not Started | ~1k est | MCP Prompts (templates) |

#### Tier 8: Server (metatools-mcp internal)

| # | Library | Status | Lines | Notes |
|---|---------|--------|-------|-------|
| 29 | toolgateway | Not Started | ~2k est | Auth proxy, analytics |
| — | multi-tenancy | Not Started | ~1.5k est | Internal module |

---

**Naming Clarification:**
- `tooladapter` (#2, toolfoundation) = Tool **schema** adapters (how tools are defined across LLM providers)
- `toolwire` (#20, toolprotocol) = Protocol **wire** adapters (how agents communicate over the wire)

**Extraction from metatools-mcp:**
- **toolops/auth** (6,389 lines): JWT, API Key, OAuth2, JWKS, RBAC - completely MCP-independent
- **toolbackend** (634 lines): Backend interface + registry - uses toolmodel, not MCP

This consolidation reduces cognitive overhead while maintaining clean separation of concerns. Each consolidated repo represents a distinct **area of concern** with clear boundaries.
