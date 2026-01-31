# PRD-181: Update ai-tools-stack

**Phase:** 8 - Integration
**Priority:** High
**Effort:** 4 hours
**Dependencies:** PRD-180

---

## Objective

Update ai-tools-stack coordination repository with new consolidated structure.

---

## Tasks

### Task 1: Update VERSIONS.md

Replace standalone repos with consolidated repos:

```markdown
# ApertureStack Version Matrix

## Consolidated Repositories (v0.1.0)

| Repository | Version | Packages |
|------------|---------|----------|
| toolfoundation | v0.1.0 | model, adapter, version |
| tooldiscovery | v0.1.0 | index, search, semantic, docs |
| toolexec | v0.1.0 | run, runtime, code, backend |
| toolcompose | v0.1.0 | set, skill |
| toolops | v0.1.0 | observe, cache, resilience, health, auth |
| toolprotocol | v0.1.0 | transport, wire, discover, content, task, stream, session, elicit, resource, prompt |
| metatools-mcp | v0.2.0 | (uses all above) |

## Compatibility Matrix

| metatools-mcp | toolfoundation | tooldiscovery | toolexec | toolcompose | toolops | toolprotocol |
|---------------|----------------|---------------|----------|-------------|---------|--------------|
| v0.2.0 | v0.1.0 | v0.1.0 | v0.1.0 | v0.1.0 | v0.1.0 | v0.1.0 |

## Archived Repositories

The following repositories have been consolidated and archived:

- toolmodel → toolfoundation/model
- tooladapter → toolfoundation/adapter
- toolindex → tooldiscovery/index
- toolsearch → tooldiscovery/search
- toolsemantic → tooldiscovery/semantic
- tooldocs → tooldiscovery/docs
- toolrun → toolexec/run
- toolruntime → toolexec/runtime
- toolcode → toolexec/code
- toolset → toolcompose/set
- toolskill → toolcompose/skill
- toolobserve → toolops/observe
- toolcache → toolops/cache
```

### Task 2: Update go.mod

```go
module github.com/ApertureStack/ai-tools-stack

go 1.24

require (
    github.com/ApertureStack/toolfoundation v0.1.0
    github.com/ApertureStack/tooldiscovery v0.1.0
    github.com/ApertureStack/toolexec v0.1.0
    github.com/ApertureStack/toolcompose v0.1.0
    github.com/ApertureStack/toolops v0.1.0
    github.com/ApertureStack/toolprotocol v0.1.0
    github.com/ApertureStack/metatools-mcp v0.2.0
)
```

### Task 3: Update README.md

```markdown
# ApertureStack

AI Tool Ecosystem for building, discovering, and executing AI agent tools.

## Repositories

| Repository | Description |
|------------|-------------|
| [toolfoundation](https://github.com/ApertureStack/toolfoundation) | Core schemas, adapters, versioning |
| [tooldiscovery](https://github.com/ApertureStack/tooldiscovery) | Registry, search, semantic, docs |
| [toolexec](https://github.com/ApertureStack/toolexec) | Execution, runtime, code |
| [toolcompose](https://github.com/ApertureStack/toolcompose) | Toolsets, skills |
| [toolops](https://github.com/ApertureStack/toolops) | Observability, caching, auth |
| [toolprotocol](https://github.com/ApertureStack/toolprotocol) | MCP, A2A, ACP protocols |
| [metatools-mcp](https://github.com/ApertureStack/metatools-mcp) | MCP server |

## Quick Start

\`\`\`bash
go get github.com/ApertureStack/metatools-mcp@latest
\`\`\`

## Documentation

See [docs.aperturestack.dev](https://docs.aperturestack.dev)
```

### Task 4: Update Scripts

Update any automation scripts to use new repo names.

### Task 5: Commit

```bash
git add -A
git commit -m "feat: update for consolidated repositories

- Update VERSIONS.md with new repo structure
- Update go.mod dependencies
- Update README with new repos
- Archive old repo references

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"

git push origin main
```

---

## Next Steps

- PRD-182: Documentation Site
- PRD-190: Archive Old Repos
