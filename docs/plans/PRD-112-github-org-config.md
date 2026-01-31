# PRD-112: GitHub Organization Configuration

**Phase:** 1 - Infrastructure Setup
**Priority:** High
**Effort:** 2 hours
**Dependencies:** PRD-110

---

## Objective

Configure GitHub organization settings, secrets, and branch protection rules for all ApertureStack repositories.

---

## Deliverables

| Deliverable | Description |
|-------------|-------------|
| Org Secrets | CODECOV_TOKEN, GH_TOKEN configured |
| Branch Protection | Main branch protection on all repos |
| Repository Settings | Consistent settings across repos |
| Teams | CODEOWNERS team configuration |

---

## Tasks

### Task 1: Configure Organization Secrets

**Commands:**
```bash
# Set organization-level secrets (requires admin access)
# These secrets are inherited by all repos in the org

# Codecov token for coverage uploads
gh secret set CODECOV_TOKEN --org ApertureStack --visibility all

# GitHub token for cross-repo access (PAT with repo scope)
gh secret set GH_TOKEN --org ApertureStack --visibility all

# Verify secrets are set
gh secret list --org ApertureStack
```

**Required Secrets:**

| Secret | Purpose | Scope |
|--------|---------|-------|
| `CODECOV_TOKEN` | Upload test coverage to Codecov | All repos |
| `GH_TOKEN` | Access private repos in go mod download | All repos |

### Task 2: Configure Branch Protection

**Script:** `scripts/configure-branch-protection.sh`

```bash
#!/bin/bash
set -euo pipefail

REPOS=(
  toolfoundation
  tooldiscovery
  toolexec
  toolcompose
  toolops
  toolprotocol
)

for repo in "${REPOS[@]}"; do
  echo "Configuring branch protection for ApertureStack/$repo..."

  gh api \
    --method PUT \
    -H "Accept: application/vnd.github+json" \
    "/repos/ApertureStack/$repo/branches/main/protection" \
    -f required_status_checks='{"strict":true,"contexts":["test","lint"]}' \
    -f enforce_admins=false \
    -f required_pull_request_reviews='{"dismiss_stale_reviews":true,"require_code_owner_reviews":true,"required_approving_review_count":1}' \
    -f restrictions=null \
    -f allow_force_pushes=false \
    -f allow_deletions=false \
    -f block_creations=false \
    -f required_conversation_resolution=true

  echo "✓ $repo configured"
done

echo "All repositories configured!"
```

**Execute:**
```bash
chmod +x scripts/configure-branch-protection.sh
./scripts/configure-branch-protection.sh
```

### Task 3: Configure Repository Settings

**Script:** `scripts/configure-repo-settings.sh`

```bash
#!/bin/bash
set -euo pipefail

REPOS=(
  toolfoundation
  tooldiscovery
  toolexec
  toolcompose
  toolops
  toolprotocol
)

for repo in "${REPOS[@]}"; do
  echo "Configuring settings for ApertureStack/$repo..."

  # Update repository settings
  gh api \
    --method PATCH \
    -H "Accept: application/vnd.github+json" \
    "/repos/ApertureStack/$repo" \
    -f has_issues=true \
    -f has_projects=false \
    -f has_wiki=false \
    -f has_discussions=false \
    -f allow_squash_merge=true \
    -f allow_merge_commit=false \
    -f allow_rebase_merge=true \
    -f allow_auto_merge=true \
    -f delete_branch_on_merge=true \
    -f allow_update_branch=true

  echo "✓ $repo settings updated"
done

echo "All repository settings configured!"
```

### Task 4: Configure Topics and Description

**Script:** `scripts/configure-repo-topics.sh`

```bash
#!/bin/bash
set -euo pipefail

declare -A DESCRIPTIONS=(
  ["toolfoundation"]="Core schemas, protocol adapters, and versioning for ApertureStack AI tool ecosystem"
  ["tooldiscovery"]="Tool registry, search, semantic indexing, and documentation"
  ["toolexec"]="Tool execution pipeline, runtime sandboxing, and code orchestration"
  ["toolcompose"]="Tool composition, toolsets, and agent skills"
  ["toolops"]="Observability, caching, resilience, health checks, and authentication"
  ["toolprotocol"]="Multi-protocol transport layer: MCP, A2A, ACP adapters"
)

COMMON_TOPICS="golang,ai-tools,mcp,llm-tools,agent-tools"

declare -A SPECIFIC_TOPICS=(
  ["toolfoundation"]="json-schema,protocol-adapters"
  ["tooldiscovery"]="search,bm25,semantic-search"
  ["toolexec"]="execution,sandbox,docker,wasm"
  ["toolcompose"]="composition,skills,toolsets"
  ["toolops"]="observability,caching,circuit-breaker,authentication"
  ["toolprotocol"]="mcp-protocol,a2a-protocol,grpc,websocket"
)

for repo in "${!DESCRIPTIONS[@]}"; do
  echo "Configuring ApertureStack/$repo..."

  # Update description
  gh repo edit "ApertureStack/$repo" --description "${DESCRIPTIONS[$repo]}"

  # Set topics
  TOPICS="$COMMON_TOPICS,${SPECIFIC_TOPICS[$repo]}"
  gh api \
    --method PUT \
    -H "Accept: application/vnd.github+json" \
    "/repos/ApertureStack/$repo/topics" \
    -f names="[$(echo $TOPICS | sed 's/,/","/g' | sed 's/^/"/' | sed 's/$/"/')"]"

  echo "✓ $repo configured"
done
```

### Task 5: Configure CODEOWNERS

Each repository already has `.github/CODEOWNERS` from PRD-110. Verify:

```bash
for repo in toolfoundation tooldiscovery toolexec toolcompose toolops toolprotocol; do
  echo "Checking CODEOWNERS in $repo..."
  gh api "/repos/ApertureStack/$repo/contents/.github/CODEOWNERS" -q '.content' | base64 -d
done
```

**Expected content:**
```
* @jonwraymond
```

### Task 6: Configure Security Settings

```bash
#!/bin/bash
set -euo pipefail

REPOS=(
  toolfoundation
  tooldiscovery
  toolexec
  toolcompose
  toolops
  toolprotocol
)

for repo in "${REPOS[@]}"; do
  echo "Enabling security features for ApertureStack/$repo..."

  # Enable Dependabot alerts
  gh api \
    --method PUT \
    -H "Accept: application/vnd.github+json" \
    "/repos/ApertureStack/$repo/vulnerability-alerts"

  # Enable Dependabot security updates
  gh api \
    --method PUT \
    -H "Accept: application/vnd.github+json" \
    "/repos/ApertureStack/$repo/automated-security-fixes"

  echo "✓ $repo security enabled"
done
```

### Task 7: Verification

```bash
# Verify all configurations
for repo in toolfoundation tooldiscovery toolexec toolcompose toolops toolprotocol; do
  echo "=== ApertureStack/$repo ==="

  # Check branch protection
  gh api "/repos/ApertureStack/$repo/branches/main/protection" -q '.required_status_checks.contexts[]' 2>/dev/null || echo "No branch protection"

  # Check secrets (just confirms they exist)
  gh secret list -R "ApertureStack/$repo" 2>/dev/null || echo "Using org secrets"

  # Check settings
  gh repo view "ApertureStack/$repo" --json allowSquashMerge,deleteBranchOnMerge

  echo ""
done
```

---

## Verification Checklist

- [ ] CODECOV_TOKEN org secret configured
- [ ] GH_TOKEN org secret configured
- [ ] Branch protection enabled on all repos
- [ ] Require status checks (test, lint)
- [ ] Require code review
- [ ] Squash merge enabled
- [ ] Delete branch on merge enabled
- [ ] Dependabot alerts enabled
- [ ] Repository topics set
- [ ] CODEOWNERS verified

---

## Acceptance Criteria

1. All org secrets are accessible to workflows
2. PRs require passing CI and code review
3. Force push to main is blocked
4. Consistent settings across all repos

---

## Rollback Plan

```bash
# Disable branch protection (allows force push to fix issues)
for repo in toolfoundation tooldiscovery toolexec toolcompose toolops toolprotocol; do
  gh api --method DELETE "/repos/ApertureStack/$repo/branches/main/protection"
done

# Remove org secrets
gh secret delete CODECOV_TOKEN --org ApertureStack
gh secret delete GH_TOKEN --org ApertureStack
```

---

## Next Steps

- PRD-113: Release Automation
- PRD-120: Migrate toolmodel
