# PRD-180: Update metatools-mcp

**Phase:** 8 - Integration
**Priority:** Critical
**Effort:** 12 hours
**Dependencies:** All Phase 2-7 PRDs
**Status:** Done (2026-02-01)

---

## Objective

Update metatools-mcp to use all consolidated repositories instead of standalone repos.

---

## Tasks

### Task 1: Update go.mod

Replace all standalone imports with consolidated repos:

```go
// Before
require (
    github.com/jonwraymond/toolmodel v0.x.x
    github.com/jonwraymond/tooladapter v0.x.x
    github.com/jonwraymond/toolindex v0.x.x
    github.com/jonwraymond/toolsearch v0.x.x
    github.com/jonwraymond/toolrun v0.x.x
    github.com/jonwraymond/toolruntime v0.x.x
    github.com/jonwraymond/toolcode v0.x.x
    github.com/jonwraymond/toolset v0.x.x
    github.com/jonwraymond/toolobserve v0.x.x
    github.com/jonwraymond/toolcache v0.x.x
)

// After
require (
    github.com/jonwraymond/toolfoundation v0.1.0
    github.com/jonwraymond/tooldiscovery v0.1.0
    github.com/jonwraymond/toolexec v0.1.0
    github.com/jonwraymond/toolcompose v0.1.0
    github.com/jonwraymond/toolops v0.1.0
    github.com/jonwraymond/toolprotocol v0.1.0
)
```

### Task 2: Update Import Statements

```bash
# Find all files with old imports
grep -r "github.com/jonwraymond/tool" --include="*.go" | grep -v "toolfoundation\|tooldiscovery\|toolexec\|toolcompose\|toolops\|toolprotocol"

# Update imports using sed
find . -name "*.go" -exec sed -i '' \
  -e 's|github.com/jonwraymond/toolmodel|github.com/jonwraymond/toolfoundation/model|g' \
  -e 's|github.com/jonwraymond/tooladapter|github.com/jonwraymond/toolfoundation/adapter|g' \
  -e 's|github.com/jonwraymond/toolindex|github.com/jonwraymond/tooldiscovery/index|g' \
  -e 's|github.com/jonwraymond/toolsearch|github.com/jonwraymond/tooldiscovery/search|g' \
  -e 's|github.com/jonwraymond/toolrun|github.com/jonwraymond/toolexec/run|g' \
  -e 's|github.com/jonwraymond/toolruntime|github.com/jonwraymond/toolexec/runtime|g' \
  -e 's|github.com/jonwraymond/toolcode|github.com/jonwraymond/toolexec/code|g' \
  -e 's|github.com/jonwraymond/toolset|github.com/jonwraymond/toolcompose/set|g' \
  -e 's|github.com/jonwraymond/toolobserve|github.com/jonwraymond/toolops/observe|g' \
  -e 's|github.com/jonwraymond/toolcache|github.com/jonwraymond/toolops/cache|g' \
  {} \;
```

### Task 3: Remove Internal Packages

Extract code that was internalized into consolidated repos:

```bash
# Remove internal packages now in consolidated repos
rm -rf internal/backend/  # Now in toolexec/backend
rm -rf internal/auth/     # Now in toolops/auth
```

### Task 4: Update Internal References

Update remaining internal code to use new packages:

```go
// Before
import "metatools-mcp/internal/auth"

// After
import "github.com/jonwraymond/toolops/auth"
```

### Task 5: Build and Test

```bash
go mod tidy
go build ./...
go test ./...
```

### Task 6: Commit

```bash
git add -A
git commit -m "feat: migrate to consolidated repositories

Update all imports to use consolidated ApertureStack repos:

Foundation:
- toolmodel → toolfoundation/model
- tooladapter → toolfoundation/adapter

Discovery:
- toolindex → tooldiscovery/index
- toolsearch → tooldiscovery/search

Execution:
- toolrun → toolexec/run
- toolruntime → toolexec/runtime
- toolcode → toolexec/code

Composition:
- toolset → toolcompose/set

Operations:
- toolobserve → toolops/observe
- toolcache → toolops/cache

## Implementation Summary

- All code imports updated to consolidated packages under `github.com/jonwraymond`.
- Docs/diagrams updated to reference consolidated layers and build tags.
- go.mod cleaned to depend on tooldiscovery/toolexec/toolfoundation.
- internal/auth → toolops/auth

BREAKING CHANGE: All import paths have changed.

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"

git push origin main
```

---

## Verification Checklist

- [x] All old imports replaced
- [x] go mod tidy succeeds
- [x] go build succeeds
- [x] All tests pass
- [ ] MCP server runs correctly
- [ ] Tool execution works

---

## Next Steps

- PRD-181: Update ai-tools-stack
- PRD-182: Documentation Site
