# PRD-192: Validation

**Phase:** 9 - Cleanup
**Priority:** Critical
**Effort:** 4 hours
**Dependencies:** PRD-191

---

## Objective

Perform comprehensive validation that the consolidation is complete and working.

---

## Tasks

### Task 1: Repository Validation

```bash
#!/bin/bash
set -e

echo "=== Repository Validation ==="

REPOS=(
  toolfoundation
  tooldiscovery
  toolexec
  toolcompose
  toolops
  toolprotocol
)

for repo in "${REPOS[@]}"; do
  echo ""
  echo "Checking $repo..."

  # Verify repo exists
  gh repo view "ApertureStack/$repo" > /dev/null

  # Clone and test
  cd /tmp
  rm -rf "$repo"
  git clone "git@github.com:ApertureStack/$repo.git"
  cd "$repo"

  # Build
  echo "  Building..."
  go build ./...

  # Test
  echo "  Testing..."
  go test ./...

  # Check CI status
  echo "  CI Status:"
  gh run list --limit 1

  echo "  ✓ $repo OK"
done

echo ""
echo "=== All repositories validated ==="
```

### Task 2: Integration Validation

```bash
#!/bin/bash
set -e

echo "=== Integration Validation ==="

cd /tmp
rm -rf metatools-mcp-test
git clone git@github.com:ApertureStack/metatools-mcp.git metatools-mcp-test
cd metatools-mcp-test

echo "Building metatools-mcp..."
go build ./cmd/metatools

echo "Running tests..."
go test ./...

echo "Starting server..."
./metatools serve &
PID=$!
sleep 3

echo "Testing tool list..."
# Test MCP tools/list
curl -s -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | jq .

echo "Stopping server..."
kill $PID

echo "✓ Integration validation passed"
```

### Task 3: Submodule Validation

```bash
#!/bin/bash
set -e

echo "=== Submodule Validation ==="

cd /tmp
rm -rf ApertureStack-test
git clone --recursive git@github.com:ApertureStack/ApertureStack.git ApertureStack-test
cd ApertureStack-test

echo "Checking submodules..."
git submodule status

echo "Building all submodules..."
for dir in toolfoundation tooldiscovery toolexec toolcompose toolops toolprotocol; do
  echo "  Building $dir..."
  cd "$dir"
  go build ./...
  cd ..
done

echo "✓ Submodule validation passed"
```

### Task 4: Documentation Validation

```bash
#!/bin/bash
set -e

echo "=== Documentation Validation ==="

cd /tmp
rm -rf ai-tools-stack-test
git clone git@github.com:ApertureStack/ai-tools-stack.git ai-tools-stack-test
cd ai-tools-stack-test

echo "Checking VERSIONS.md..."
grep -q "toolfoundation" VERSIONS.md && echo "  ✓ toolfoundation listed"
grep -q "tooldiscovery" VERSIONS.md && echo "  ✓ tooldiscovery listed"
grep -q "toolexec" VERSIONS.md && echo "  ✓ toolexec listed"
grep -q "toolcompose" VERSIONS.md && echo "  ✓ toolcompose listed"
grep -q "toolops" VERSIONS.md && echo "  ✓ toolops listed"
grep -q "toolprotocol" VERSIONS.md && echo "  ✓ toolprotocol listed"

echo "Building docs..."
mkdocs build

echo "✓ Documentation validation passed"
```

### Task 5: Archive Validation

```bash
#!/bin/bash
set -e

echo "=== Archive Validation ==="

ARCHIVED=(
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

for repo in "${ARCHIVED[@]}"; do
  ARCHIVED_STATUS=$(gh repo view "ApertureStack/$repo" --json isArchived -q '.isArchived')
  if [ "$ARCHIVED_STATUS" = "true" ]; then
    echo "  ✓ $repo is archived"
  else
    echo "  ✗ $repo is NOT archived"
    exit 1
  fi
done

echo "✓ Archive validation passed"
```

### Task 6: Final Checklist

Create validation report:

```markdown
# Consolidation Validation Report

Date: [YYYY-MM-DD]

## Repository Status

| Repository | Build | Tests | CI | Status |
|------------|-------|-------|-----|--------|
| toolfoundation | ✓ | ✓ | ✓ | OK |
| tooldiscovery | ✓ | ✓ | ✓ | OK |
| toolexec | ✓ | ✓ | ✓ | OK |
| toolcompose | ✓ | ✓ | ✓ | OK |
| toolops | ✓ | ✓ | ✓ | OK |
| toolprotocol | ✓ | ✓ | ✓ | OK |
| metatools-mcp | ✓ | ✓ | ✓ | OK |

## Integration Status

| Test | Status |
|------|--------|
| metatools-mcp build | ✓ |
| MCP server starts | ✓ |
| tools/list works | ✓ |
| Tool execution works | ✓ |

## Archive Status

All 13 standalone repos archived: ✓

## Submodule Status

All 6 consolidated repos as submodules: ✓

## Documentation Status

- VERSIONS.md updated: ✓
- README updated: ✓
- MkDocs builds: ✓

## Conclusion

Consolidation complete and validated.
```

---

## Verification Checklist

- [ ] All 6 consolidated repos build
- [ ] All 6 consolidated repos pass tests
- [ ] All 6 consolidated repos have passing CI
- [ ] metatools-mcp builds with new imports
- [ ] metatools-mcp tests pass
- [ ] MCP server runs correctly
- [ ] All 13 old repos archived
- [ ] Submodules work correctly
- [ ] Documentation updated
- [ ] Validation report created

---

## Acceptance Criteria

1. All smoke tests pass
2. No broken imports
3. CI green on all repos
4. Documentation accessible
5. Old repos archived and read-only

---

## Gate G7 Complete

Upon successful validation:
- ApertureStack consolidation is complete
- 15 repos → 8 repos (6 consolidated + metatools-mcp + ai-tools-stack)
- 29 packages organized into 6 consolidated repositories
- Full backward compatibility removed per requirements

---

## Post-Completion

- Monitor for any issues from external users
- Update any external documentation/links
- Consider GitHub redirects for archived repos
