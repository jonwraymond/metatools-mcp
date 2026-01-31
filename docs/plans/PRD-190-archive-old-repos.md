# PRD-190: Archive Old Repos

**Phase:** 9 - Cleanup
**Priority:** High
**Effort:** 2 hours
**Dependencies:** PRD-180

---

## Objective

Archive all 13 standalone repositories that have been consolidated.

---

## Repositories to Archive

| Repository | Migrated To |
|------------|-------------|
| toolmodel | toolfoundation/model |
| tooladapter | toolfoundation/adapter |
| toolindex | tooldiscovery/index |
| toolsearch | tooldiscovery/search |
| toolsemantic | tooldiscovery/semantic |
| tooldocs | tooldiscovery/docs |
| toolrun | toolexec/run |
| toolruntime | toolexec/runtime |
| toolcode | toolexec/code |
| toolset | toolcompose/set |
| toolskill | toolcompose/skill |
| toolobserve | toolops/observe |
| toolcache | toolops/cache |

---

## Tasks

### Task 1: Add Deprecation Notices

For each repo, update README.md:

```markdown
# ⚠️ DEPRECATED

This repository has been archived. The code has been migrated to:

**[toolfoundation](https://github.com/ApertureStack/toolfoundation)** (model package)

## Migration

Update your imports:

```go
// Before
import "github.com/ApertureStack/toolmodel"

// After
import "github.com/ApertureStack/toolfoundation/model"
```

See [MIGRATION.md](./MIGRATION.md) for details.
```

### Task 2: Create Migration Guides

**File: MIGRATION.md** (per repo)

```markdown
# Migration Guide

## Import Changes

| Old | New |
|-----|-----|
| `github.com/ApertureStack/toolmodel` | `github.com/ApertureStack/toolfoundation/model` |

## Breaking Changes

- Package name changed from `toolmodel` to `model`
- All functionality preserved

## Steps

1. Update go.mod:
   ```bash
   go get github.com/ApertureStack/toolfoundation@latest
   ```

2. Update imports:
   ```bash
   sed -i 's|github.com/ApertureStack/toolmodel|github.com/ApertureStack/toolfoundation/model|g' *.go
   ```

3. Remove old dependency:
   ```bash
   go mod tidy
   ```
```

### Task 3: Archive Repositories

```bash
REPOS=(
  toolmodel
  tooladapter
  toolindex
  toolsearch
  toolsemantic
  tooldocs
  toolrun
  toolruntime
  toolcode
  toolset
  toolskill
  toolobserve
  toolcache
)

for repo in "${REPOS[@]}"; do
  echo "Archiving ApertureStack/$repo..."

  # Update README with deprecation notice
  # (done manually per repo with appropriate redirect)

  # Archive the repository
  gh repo archive "ApertureStack/$repo" --yes

  echo "✓ $repo archived"
done
```

### Task 4: Verify Archives

```bash
# List archived repos
gh repo list ApertureStack --archived --limit 20

# Verify each is read-only
for repo in toolmodel tooladapter toolindex toolsearch; do
  gh repo view "ApertureStack/$repo" --json isArchived -q '.isArchived'
done
```

---

## Verification Checklist

- [ ] All 13 repos have deprecation notices
- [ ] All 13 repos have MIGRATION.md
- [ ] All 13 repos are archived
- [ ] Archives are read-only
- [ ] Redirects work (README links to new location)

---

## Rollback Plan

```bash
# Unarchive a repo if needed
gh repo unarchive ApertureStack/toolmodel --yes
```

---

## Next Steps

- PRD-191: Update Submodules
- PRD-192: Validation
