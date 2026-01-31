# PRD-191: Update Submodules

**Phase:** 9 - Cleanup
**Priority:** High
**Effort:** 2 hours
**Dependencies:** PRD-190

---

## Objective

Update ApertureStack root to use new consolidated submodules.

---

## Tasks

### Task 1: Remove Old Submodules

```bash
cd /Users/jraymond/Documents/Projects/ApertureStack

# Remove old submodules
OLD_REPOS=(
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

for repo in "${OLD_REPOS[@]}"; do
  git submodule deinit -f "$repo" 2>/dev/null || true
  git rm -f "$repo" 2>/dev/null || true
  rm -rf ".git/modules/$repo" 2>/dev/null || true
done
```

### Task 2: Add New Submodules

```bash
cd /Users/jraymond/Documents/Projects/ApertureStack

# Add consolidated submodules
git submodule add git@github.com:ApertureStack/toolfoundation.git toolfoundation
git submodule add git@github.com:ApertureStack/tooldiscovery.git tooldiscovery
git submodule add git@github.com:ApertureStack/toolexec.git toolexec
git submodule add git@github.com:ApertureStack/toolcompose.git toolcompose
git submodule add git@github.com:ApertureStack/toolops.git toolops
git submodule add git@github.com:ApertureStack/toolprotocol.git toolprotocol
```

### Task 3: Update .gitmodules

Verify `.gitmodules` looks like:

```ini
[submodule "toolfoundation"]
    path = toolfoundation
    url = git@github.com:ApertureStack/toolfoundation.git

[submodule "tooldiscovery"]
    path = tooldiscovery
    url = git@github.com:ApertureStack/tooldiscovery.git

[submodule "toolexec"]
    path = toolexec
    url = git@github.com:ApertureStack/toolexec.git

[submodule "toolcompose"]
    path = toolcompose
    url = git@github.com:ApertureStack/toolcompose.git

[submodule "toolops"]
    path = toolops
    url = git@github.com:ApertureStack/toolops.git

[submodule "toolprotocol"]
    path = toolprotocol
    url = git@github.com:ApertureStack/toolprotocol.git

[submodule "metatools-mcp"]
    path = metatools-mcp
    url = git@github.com:ApertureStack/metatools-mcp.git

[submodule "ai-tools-stack"]
    path = ai-tools-stack
    url = git@github.com:ApertureStack/ai-tools-stack.git
```

### Task 4: Commit Changes

```bash
git add .gitmodules
git add toolfoundation tooldiscovery toolexec toolcompose toolops toolprotocol

git commit -m "feat: update submodules to consolidated repos

Remove old standalone submodules:
- toolmodel, tooladapter
- toolindex, toolsearch, toolsemantic, tooldocs
- toolrun, toolruntime, toolcode
- toolset, toolskill
- toolobserve, toolcache

Add consolidated submodules:
- toolfoundation (model, adapter, version)
- tooldiscovery (index, search, semantic, docs)
- toolexec (run, runtime, code, backend)
- toolcompose (set, skill)
- toolops (observe, cache, resilience, health, auth)
- toolprotocol (transport, wire, discover, content, task, stream, session, elicit, resource, prompt)

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"

git push origin main
```

### Task 5: Initialize Submodules

```bash
# On fresh clone
git submodule update --init --recursive

# Verify
git submodule status
```

---

## Expected Structure

```
ApertureStack/
├── toolfoundation/      # NEW
├── tooldiscovery/       # NEW
├── toolexec/            # NEW
├── toolcompose/         # NEW
├── toolops/             # NEW
├── toolprotocol/        # NEW
├── metatools-mcp/       # Existing (updated)
├── ai-tools-stack/      # Existing (updated)
└── .gitmodules
```

---

## Verification Checklist

- [ ] Old submodules removed
- [ ] New submodules added
- [ ] .gitmodules updated
- [ ] Committed and pushed
- [ ] Fresh clone works
- [ ] `git submodule update --init` works

---

## Next Steps

- PRD-192: Validation
- Gate G7: Full validation complete
