# Usage

## Build and run (stdio)

```bash
go run ./cmd/metatools serve
```

## CLI overview

```bash
metatools serve --transport=stdio                        # Local/Claude Desktop (default)
metatools serve --transport=streamable --port=8080       # HTTP clients (recommended)
metatools serve --transport=sse --port=8080              # Legacy HTTP clients (deprecated)
metatools version
metatools config validate --config examples/metatools.yaml
```

## Transport selection

| Transport | Use Case | Protocol |
|-----------|----------|----------|
| `stdio` | Claude Desktop, local CLI clients | stdin/stdout JSON-RPC |
| `streamable` | Web apps, REST APIs, remote clients | HTTP POST/GET/DELETE (MCP 2025-11-25) |
| `sse` | Legacy web clients | HTTP + Server-Sent Events (deprecated) |

### Streamable HTTP (recommended for HTTP)

Streamable HTTP is the MCP spec (2025-11-25) transport replacing SSE:

```bash
# Basic HTTP server
metatools serve --transport=streamable --port=8080

# With TLS
metatools serve --transport=streamable --port=443 \
  --tls --tls-cert=cert.pem --tls-key=key.pem

# Stateless mode (no session tracking)
metatools serve --transport=streamable --port=8080 --stateless
```

**Protocol flow:**
1. Client POSTs JSON-RPC to `/mcp` with `initialize` request
2. Server responds with `Mcp-Session-Id` header
3. Client includes session ID in subsequent requests
4. Client may GET `/mcp` for server notification stream
5. Client DELETEs `/mcp` to terminate session

**YAML configuration:**
```yaml
transport:
  type: streamable
  http:
    host: 0.0.0.0
    port: 8080
    tls:
      enabled: true
      cert: /path/to/cert.pem
      key: /path/to/key.pem
  streamable:
    stateless: false        # Enable session management
    json_response: false    # Use SSE streaming (default)
    session_timeout: 30m    # Clean up idle sessions
```

## Configuration files (Koanf)

Config precedence:
1. Defaults
2. Config file (`--config`)
3. Environment variables (`METATOOLS_` prefix)
4. CLI flags

Example file: `examples/metatools.yaml`

Environment variables in config files:
- `${VAR}` and `$VAR` are expanded from the process environment.
- Missing `${VAR}` values fail fast at startup (recommended for secrets).
- Use `$$` to write a literal `$` without triggering expansion.
- `secretref:<provider>:<ref>` values are resolved via the `secrets` config block
  before backend clients are constructed (fail fast in strict mode).

## MCP backends (remote MCP servers)

You can aggregate tools from remote MCP servers and execute them through the
runner by configuring `backends.mcp`.

```yaml
backends:
  mcp:
    - name: "deepwiki"
      url: "https://mcp.deepwiki.com/mcp"
      headers: {}
      max_retries: 5
  mcp_refresh:
    interval: 10m     # periodic refresh cadence (0 to disable)
    jitter: 30s       # randomized jitter to avoid stampedes
    stale_after: 15m  # on-demand refresh if older than this
    on_demand: true   # refresh on search/list when stale
```

If a backend requires authentication headers, you can inject them via env vars:

```yaml
backends:
  mcp:
    - name: "private-mcp"
      url: "https://mcp.example.com/mcp"
      headers:
        Authorization: "Bearer ${MCP_BACKEND_TOKEN}"
```

### Secret refs (optional)

If you don't want secrets in the environment, metatools-mcp can resolve
`secretref:<provider>:<ref>` at startup using `toolops/secret` providers.

Example: Bitwarden Secrets Manager provider (`bws`) via `toolops-integrations`:

```yaml
secrets:
  strict: true
  providers:
    bws:
      enabled: true
      config:
        access_token: ${BWS_ACCESS_TOKEN}
        organization_id: ${BWS_ORG_ID}
        cache_ttl: 10m

backends:
  mcp:
    - name: "supabase"
      url: "https://mcp.supabase.com/mcp"
      headers:
        Authorization: "Bearer secretref:bws:project/dotenv/key/SUPABASE_ACCESS_TOKEN"
```

Notes:
- Resolution is strict by default: missing `${ENV}` or unresolved `secretref:*`
  fails the server startup.
- Errors identify the backend and config field, but never include secret values.

This registers tools from the remote server into the index and enables the
runner to dispatch to them via MCP backend calls.

## Provider toggles

Built-in metatools can be enabled/disabled via `providers.*.enabled` in the
config file (see the `providers` block in `examples/metatools.yaml`). This
controls which MCP tools are registered at startup.

New providers for toolsets and skills:
- `list_tools` (paged tool inventory)
- `list_toolsets`, `describe_toolset`
- `list_skills`, `describe_skill`, `plan_skill`, `run_skill`

## Toolsets and skills

Toolsets are deterministic collections of tools built from index filters and
policies. Skills are pre-registered workflows that can be planned and executed
by the server with guardrails.

```yaml
toolsets:
  - name: "core"
    description: "Core safe tools"
    namespace_filters: ["local", "core"]
    tag_filters: ["safe"]
    allow_ids: []
    deny_ids: []
    policy: "allow_all"

skills:
  - name: "ping_check"
    description: "Run a simple ping tool"
    toolset_id: "toolset:core"
    steps:
      - id: "ping"
        tool_id: "local.ping"
        inputs: {}
    guards:
      max_steps: 4
      allow_ids: []

skill_defaults:
  max_steps: 8
  max_tool_calls: 32
  timeout: 30s
```

## Middleware chain

Configure optional middleware in `middleware.chain` (ordered) with per-middleware
settings under `middleware.configs`. Built-in middleware: `auth`, `logging`,
`metrics`, `ratelimit`, `audit`.

Example (JWT auth + RBAC):

```yaml
middleware:
  chain: ["auth", "logging", "metrics"]
  configs:
    auth:
      config:
        allow_anonymous: false
        authenticators:
          - name: "jwt"
            config:
              issuer: "https://auth.example.com"
              audience: "metatools"
              jwks_url: "https://auth.example.com/.well-known/jwks.json"
        authorizer:
          name: "simple_rbac"
          config:
            roles:
              admin:
                allow:
                  - tool: "*"
                    actions: ["call", "list"]
```

### Toolops integration (observe/cache/resilience)

Toolops wrappers are configured outside the middleware chain and apply to
execution paths (`run_tool`, `run_chain`, `execute_code`, `run_skill`).

```yaml
middleware:
  observe:
    enabled: true
    config:
      service: "metatools-mcp"
  cache:
    enabled: false
    policy:
      ttl: 5m
  resilience:
    enabled: true
    retry:
      enabled: true
      config:
        max_attempts: 3
    circuit:
      enabled: true
      config:
        failure_threshold: 5
        success_threshold: 2
    timeout: 30s
```

## Enable BM25 search (build tag + env)

```bash
go build -tags toolsearch ./cmd/metatools
METATOOLS_SEARCH_STRATEGY=bm25 ./metatools
```

## Enable semantic or hybrid search (BYO embedder)

Semantic/hybrid search requires the `toolsemantic` build tag and an embedder
adapter registered at runtime.

```bash
go build -tags toolsemantic ./cmd/metatools
METATOOLS_SEARCH_STRATEGY=semantic METATOOLS_SEARCH_SEMANTIC_EMBEDDER=my-embedder ./metatools
```

If `semantic` is requested without an embedder, the server falls back to lexical
search. If `hybrid` is requested without an embedder, it falls back to BM25 when
available, otherwise lexical.

```yaml
search:
  strategy: semantic
  semantic:
    embedder: "my-embedder"
    config:
      model: "text-embedding-3-small"
    weight: 0.5
```

## Environment variables

### CLI defaults (serve command)

These map directly to `metatools serve` flags when the flags are not set:

| Variable | Default | Description |
|----------|---------|-------------|
| `METATOOLS_TRANSPORT` | `stdio` | Transport type: `stdio`, `streamable`, `sse` |
| `METATOOLS_PORT` | `8080` | Port for HTTP transports |
| `METATOOLS_HOST` | `0.0.0.0` | Host/interface for HTTP transports |
| `METATOOLS_CONFIG` | "" | Path to config file |

### Transport configuration (Koanf config)

These map to the config schema loaded by `config.Load`:

| Variable | Default | Description |
|----------|---------|-------------|
| `METATOOLS_TRANSPORT_TYPE` | `stdio` | Transport type: `stdio`, `streamable`, `sse` |
| `METATOOLS_TRANSPORT_HTTP_HOST` | `0.0.0.0` | Host/interface for HTTP transports |
| `METATOOLS_TRANSPORT_HTTP_PORT` | `8080` | Port for HTTP transports |
| `METATOOLS_TRANSPORT_HTTP_TLS_ENABLED` | `false` | Enable TLS for HTTP transports |
| `METATOOLS_TRANSPORT_HTTP_TLS_CERT` | "" | TLS certificate path |
| `METATOOLS_TRANSPORT_HTTP_TLS_KEY` | "" | TLS key path |
| `METATOOLS_TRANSPORT_STREAMABLE_STATELESS` | `false` | Disable session management |
| `METATOOLS_TRANSPORT_STREAMABLE_JSON_RESPONSE` | `false` | Prefer JSON over SSE streaming |
| `METATOOLS_TRANSPORT_STREAMABLE_SESSION_TIMEOUT` | `30m` | Idle session cleanup duration |
| `METATOOLS_STATE_RUNTIME_LIMITS_DB` | "" | SQLite file for persisted execution limits |

### Runtime configuration (toolruntime build tag)

| Variable | Default | Description |
|----------|---------|-------------|
| `METATOOLS_RUNTIME_PROFILE` | `dev` | `dev` (unsafe) or `standard` (Docker) |
| `METATOOLS_DOCKER_IMAGE` | `toolruntime-sandbox:latest` | Docker image for standard profile |
| `METATOOLS_WASM_ENABLED` | `false` | Enable WASM backend (wazero) |
| `METATOOLS_RUNTIME_BACKEND` | `docker` | Preferred standard backend: `docker` or `wasm` |

### Search configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `METATOOLS_SEARCH_STRATEGY` | `lexical` | `lexical`, `bm25`, `semantic`, or `hybrid` |
| `METATOOLS_SEARCH_BM25_NAME_BOOST` | `3` | BM25 name field boost |
| `METATOOLS_SEARCH_BM25_NAMESPACE_BOOST` | `2` | BM25 namespace field boost |
| `METATOOLS_SEARCH_BM25_TAGS_BOOST` | `2` | BM25 tags field boost |
| `METATOOLS_SEARCH_BM25_MAX_DOCS` | `0` | Max docs to index (0=unlimited) |
| `METATOOLS_SEARCH_BM25_MAX_DOCTEXT_LEN` | `0` | Max doc text length (0=unlimited) |
| `METATOOLS_SEARCH_SEMANTIC_EMBEDDER` | "" | Embedder registry key (semantic/hybrid) |
| `METATOOLS_SEARCH_SEMANTIC_WEIGHT` | `0.5` | Hybrid semantic weight |
| `METATOOLS_NOTIFY_TOOL_LIST_CHANGED` | `true` | Emit `notifications/tools/list_changed` on index updates |
| `METATOOLS_NOTIFY_TOOL_LIST_CHANGED_DEBOUNCE_MS` | `150` | Debounce window for list change notifications |

## Pagination and cursors

- `search_tools`, `list_tools`, and `list_namespaces` accept `limit` (default 20, max 100) and `cursor`.
- Responses include `nextCursor` when more results are available.
- Cursor tokens are opaque and invalid cursors return JSON-RPC invalid params.

## Tool list change notifications

- `notifications/tools/list_changed` is emitted when the underlying tooldiscovery/index changes.
- Notifications are debounced to avoid client spam and can be disabled with `METATOOLS_NOTIFY_TOOL_LIST_CHANGED=false`.

## Progress notifications

When callers supply a progress token, `run_tool`, `run_chain`, and `execute_code`
emit progress notifications. If the runner exposes progress callbacks, step-level
updates are forwarded; otherwise a coarse start/end signal is emitted.

`run_skill` forwards progress from the underlying runner where available.

## Health endpoint (HTTP transports)

Streamable HTTP and SSE transports can expose a lightweight health endpoint.

```yaml
health:
  enabled: true
  http_path: /healthz
```

When enabled, `GET /healthz` returns a 200 with a simple JSON body.

## Optional toolruntime support

```bash
go run -tags toolruntime ./cmd/metatools
```

This enables `execute_code` backed by a `toolexec/runtime` engine.
By default it uses the `dev` (unsafe) profile; set
`METATOOLS_RUNTIME_PROFILE=standard` to enable the Docker backend when
available. To use WASM instead, set `METATOOLS_WASM_ENABLED=true` and
`METATOOLS_RUNTIME_BACKEND=wasm`.
