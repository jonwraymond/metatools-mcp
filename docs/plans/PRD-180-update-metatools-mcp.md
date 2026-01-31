# PRD-180: Update metatools-mcp

**Phase:** 8 - Integration
**Priority:** Critical
**Effort:** 12 hours
**Dependencies:** All Phase 2-7 PRDs

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
    github.com/ApertureStack/toolmodel v0.x.x
    github.com/ApertureStack/tooladapter v0.x.x
    github.com/ApertureStack/toolindex v0.x.x
    github.com/ApertureStack/toolsearch v0.x.x
    github.com/ApertureStack/toolrun v0.x.x
    github.com/ApertureStack/toolruntime v0.x.x
    github.com/ApertureStack/toolcode v0.x.x
    github.com/ApertureStack/toolset v0.x.x
    github.com/ApertureStack/toolobserve v0.x.x
    github.com/ApertureStack/toolcache v0.x.x
)

// After
require (
    github.com/ApertureStack/toolfoundation v0.1.0
    github.com/ApertureStack/tooldiscovery v0.1.0
    github.com/ApertureStack/toolexec v0.1.0
    github.com/ApertureStack/toolcompose v0.1.0
    github.com/ApertureStack/toolops v0.1.0
    github.com/ApertureStack/toolprotocol v0.1.0
)
```

### Task 2: Update Import Statements

```bash
# Find all files with old imports
grep -r "github.com/ApertureStack/tool" --include="*.go" | grep -v "toolfoundation\|tooldiscovery\|toolexec\|toolcompose\|toolops\|toolprotocol"

# Update imports using sed
find . -name "*.go" -exec sed -i '' \
  -e 's|github.com/ApertureStack/toolmodel|github.com/ApertureStack/toolfoundation/model|g' \
  -e 's|github.com/ApertureStack/tooladapter|github.com/ApertureStack/toolfoundation/adapter|g' \
  -e 's|github.com/ApertureStack/toolindex|github.com/ApertureStack/tooldiscovery/index|g' \
  -e 's|github.com/ApertureStack/toolsearch|github.com/ApertureStack/tooldiscovery/search|g' \
  -e 's|github.com/ApertureStack/toolrun|github.com/ApertureStack/toolexec/run|g' \
  -e 's|github.com/ApertureStack/toolruntime|github.com/ApertureStack/toolexec/runtime|g' \
  -e 's|github.com/ApertureStack/toolcode|github.com/ApertureStack/toolexec/code|g' \
  -e 's|github.com/ApertureStack/toolset|github.com/ApertureStack/toolcompose/set|g' \
  -e 's|github.com/ApertureStack/toolobserve|github.com/ApertureStack/toolops/observe|g' \
  -e 's|github.com/ApertureStack/toolcache|github.com/ApertureStack/toolops/cache|g' \
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
import "github.com/ApertureStack/toolops/auth"
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
- internal/auth → toolops/auth

BREAKING CHANGE: All import paths have changed.

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"

git push origin main
```

---

## Verification Checklist

- [ ] All old imports replaced
- [ ] go mod tidy succeeds
- [ ] go build succeeds
- [ ] All tests pass
- [ ] MCP server runs correctly
- [ ] Tool execution works

---

## Next Steps

- PRD-181: Update ai-tools-stack
- PRD-182: Documentation Site
