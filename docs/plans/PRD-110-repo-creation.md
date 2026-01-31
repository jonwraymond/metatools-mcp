# PRD-110: Repository Creation

**Phase:** 1 - Infrastructure Setup
**Priority:** Critical
**Effort:** 8 hours
**Dependencies:** PRD-100

---

## Objective

Create 6 new consolidated repositories with standardized structure, ready for code migration.

---

## Repositories to Create

| Repository | GitHub URL | Description |
|------------|------------|-------------|
| toolfoundation | github.com/ApertureStack/toolfoundation | Canonical schema, protocol adapters, versioning |
| tooldiscovery | github.com/ApertureStack/tooldiscovery | Registry, search, semantic, documentation |
| toolexec | github.com/ApertureStack/toolexec | Execution pipeline, runtime, code orchestration |
| toolcompose | github.com/ApertureStack/toolcompose | Toolsets, agent skills |
| toolops | github.com/ApertureStack/toolops | Observability, caching, resilience, health, auth |
| toolprotocol | github.com/ApertureStack/toolprotocol | Transport, wire adapters, protocol features |

---

## Standard Repository Structure

Each repository follows this exact structure:

```
{repo-name}/
├── .github/
│   ├── CODEOWNERS
│   ├── PULL_REQUEST_TEMPLATE.md
│   ├── ISSUE_TEMPLATE/
│   │   ├── bug_report.md
│   │   └── feature_request.md
│   └── workflows/
│       ├── ci.yml
│       ├── lint-security.yml
│       ├── commitlint.yml
│       └── release-please.yml
├── docs/
│   ├── index.md
│   ├── design-notes.md
│   └── user-journey.md
├── examples/
│   └── .gitkeep
├── .gitignore
├── .golangci.yml
├── commitlint.config.cjs
├── go.mod
├── go.sum
├── LICENSE
├── README.md
├── CHANGELOG.md
├── release-please-config.json
└── .release-please-manifest.json
```

---

## Tasks

### Task 1: Create GitHub Repositories

**Commands:**
```bash
# Using GitHub CLI
for repo in toolfoundation tooldiscovery toolexec toolcompose toolops toolprotocol; do
  gh repo create ApertureStack/$repo \
    --public \
    --description "ApertureStack $repo - AI tool ecosystem" \
    --license MIT \
    --clone
done
```

**Verification:**
```bash
gh repo list ApertureStack --limit 20
```

### Task 2: Initialize Repository Structure

For each repository, create the standard structure:

**Script: `scripts/init-repo.sh`**
```bash
#!/bin/bash
set -euo pipefail

REPO_NAME=$1
MODULE_PATH="github.com/ApertureStack/$REPO_NAME"

cd "$REPO_NAME"

# Create directories
mkdir -p .github/workflows .github/ISSUE_TEMPLATE docs examples

# Create .gitignore
cat > .gitignore << 'EOF'
# Binaries
*.exe
*.exe~
*.dll
*.so
*.dylib
*.test
*.out

# Go
vendor/
go.work
go.work.sum

# IDE
.idea/
.vscode/
*.swp
*.swo

# OS
.DS_Store
Thumbs.db

# Coverage
coverage.out
coverage.html

# Build
dist/
bin/
EOF

# Create go.mod
cat > go.mod << EOF
module $MODULE_PATH

go 1.24
EOF

# Create LICENSE (MIT)
cat > LICENSE << 'EOF'
MIT License

Copyright (c) 2026 ApertureStack

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
EOF

# Create README.md
cat > README.md << EOF
# $REPO_NAME

Part of the [ApertureStack](https://github.com/ApertureStack) AI tool ecosystem.

## Installation

\`\`\`bash
go get $MODULE_PATH
\`\`\`

## Packages

| Package | Description | Documentation |
|---------|-------------|---------------|
| TBD | TBD | [docs](./docs/) |

## License

MIT License - see [LICENSE](./LICENSE)
EOF

# Create CHANGELOG.md
cat > CHANGELOG.md << 'EOF'
# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial repository structure
EOF

# Create .golangci.yml
cat > .golangci.yml << 'EOF'
run:
  timeout: 5m
  modules-download-mode: readonly

linters:
  enable:
    - errcheck
    - gosimple
    - govet
    - ineffassign
    - staticcheck
    - unused
    - gofmt
    - goimports
    - misspell
    - unconvert
    - unparam

linters-settings:
  goimports:
    local-prefixes: github.com/ApertureStack

issues:
  exclude-use-default: false
  max-issues-per-linter: 0
  max-same-issues: 0
EOF

# Create commitlint.config.cjs
cat > commitlint.config.cjs << 'EOF'
module.exports = { extends: ['@commitlint/config-conventional'] };
EOF

# Create release-please-config.json
cat > release-please-config.json << 'EOF'
{
  "$schema": "https://raw.githubusercontent.com/googleapis/release-please/main/schemas/config.json",
  "release-type": "go",
  "packages": {
    ".": {}
  },
  "changelog-sections": [
    {"type": "feat", "section": "Features"},
    {"type": "fix", "section": "Bug Fixes"},
    {"type": "perf", "section": "Performance Improvements"},
    {"type": "refactor", "section": "Code Refactoring"},
    {"type": "docs", "section": "Documentation"},
    {"type": "chore", "section": "Miscellaneous"}
  ]
}
EOF

# Create .release-please-manifest.json
echo '{".":" 0.0.0"}' > .release-please-manifest.json

# Create GitHub templates
cat > .github/CODEOWNERS << 'EOF'
* @jonwraymond
EOF

cat > .github/PULL_REQUEST_TEMPLATE.md << 'EOF'
## Description

<!-- Describe your changes -->

## Type of Change

- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Checklist

- [ ] Tests pass (`go test ./...`)
- [ ] Linter passes (`golangci-lint run`)
- [ ] Documentation updated
- [ ] CHANGELOG.md updated (if applicable)
EOF

cat > .github/ISSUE_TEMPLATE/bug_report.md << 'EOF'
---
name: Bug Report
about: Report a bug
title: '[BUG] '
labels: bug
---

## Description

## Steps to Reproduce

1.
2.
3.

## Expected Behavior

## Actual Behavior

## Environment

- Go version:
- OS:
EOF

cat > .github/ISSUE_TEMPLATE/feature_request.md << 'EOF'
---
name: Feature Request
about: Suggest a feature
title: '[FEATURE] '
labels: enhancement
---

## Description

## Use Case

## Proposed Solution

## Alternatives Considered
EOF

# Create docs
cat > docs/index.md << EOF
# $REPO_NAME

Overview documentation for $REPO_NAME.

## Packages

TBD

## Getting Started

TBD
EOF

cat > docs/design-notes.md << 'EOF'
# Design Notes

## Architecture Decisions

TBD

## Trade-offs

TBD
EOF

cat > docs/user-journey.md << 'EOF'
# User Journey

## Installation

TBD

## Basic Usage

TBD

## Advanced Usage

TBD
EOF

# Create examples placeholder
touch examples/.gitkeep

echo "✅ Initialized $REPO_NAME"
```

**Execute:**
```bash
chmod +x scripts/init-repo.sh
for repo in toolfoundation tooldiscovery toolexec toolcompose toolops toolprotocol; do
  ./scripts/init-repo.sh $repo
done
```

### Task 3: Create CI Workflows

**File: `.github/workflows/ci.yml`**
```yaml
name: CI

on:
  push:
    branches: ["main"]
    paths-ignore:
      - "**.md"
      - "docs/**"
  pull_request:
    paths-ignore:
      - "**.md"
      - "docs/**"

permissions:
  contents: read

jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        go: ["1.24"]
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ matrix.go }}
          cache: true

      - name: Download dependencies
        run: go mod download

      - name: Verify dependencies
        run: go mod verify

      - name: Build
        run: go build ./...

      - name: Test
        run: go test -race -coverprofile=coverage.out -covermode=atomic ./...

      - name: Upload coverage
        uses: codecov/codecov-action@v4
        with:
          token: ${{ secrets.CODECOV_TOKEN }}
          files: coverage.out
          fail_ci_if_error: false
```

**File: `.github/workflows/lint-security.yml`**
```yaml
name: Lint & Security

on:
  push:
    branches: ["main"]
  pull_request:

permissions:
  contents: read
  security-events: write

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
          cache: true

      - name: golangci-lint
        uses: golangci/golangci-lint-action@v6
        with:
          version: latest
          args: --timeout=5m

  security:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version-file: go.mod

      - name: Run gosec
        uses: securego/gosec@master
        with:
          args: -fmt sarif -out gosec.sarif ./...

      - name: Upload SARIF
        uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: gosec.sarif
```

**File: `.github/workflows/commitlint.yml`**
```yaml
name: Commitlint

on:
  push:
    branches: ["main"]
  pull_request:

jobs:
  commitlint:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Setup Node
        uses: actions/setup-node@v4
        with:
          node-version: "20"

      - name: Install commitlint
        run: npm install -g @commitlint/cli @commitlint/config-conventional

      - name: Lint commits
        run: npx commitlint --from HEAD~1 --to HEAD --verbose
```

**File: `.github/workflows/release-please.yml`**
```yaml
name: Release Please

on:
  push:
    branches: ["main"]

permissions:
  contents: write
  pull-requests: write

jobs:
  release-please:
    runs-on: ubuntu-latest
    steps:
      - name: Release Please
        uses: google-github-actions/release-please-action@v4
        with:
          token: ${{ secrets.GITHUB_TOKEN }}
          config-file: release-please-config.json
          manifest-file: .release-please-manifest.json
```

### Task 4: Push Initial Structure

```bash
for repo in toolfoundation tooldiscovery toolexec toolcompose toolops toolprotocol; do
  cd $repo
  git add -A
  git commit -m "feat: initial repository structure

- Add standard directory layout
- Add CI/CD workflows
- Add documentation templates
- Add release automation

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
  git push origin main
  cd ..
done
```

### Task 5: Add as Submodules to ApertureStack

```bash
cd /Users/jraymond/Documents/Projects/ApertureStack

for repo in toolfoundation tooldiscovery toolexec toolcompose toolops toolprotocol; do
  git submodule add git@github.com:ApertureStack/$repo.git $repo
done

git add .gitmodules
git commit -m "feat: add consolidated repository submodules

- toolfoundation (model, adapter, version)
- tooldiscovery (index, search, semantic, docs)
- toolexec (run, runtime, code, backend)
- toolcompose (set, skill)
- toolops (observe, cache, resilience, health, auth)
- toolprotocol (transport, wire, discover, content, task, stream, session, elicit, resource, prompt)

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

---

## Verification Checklist

- [ ] All 6 repositories created on GitHub
- [ ] Each repo has correct structure
- [ ] Each repo has all 4 workflow files
- [ ] All workflows pass initial run
- [ ] Repos added as submodules to ApertureStack
- [ ] Git push succeeds for all repos

---

## Acceptance Criteria

1. `gh repo view ApertureStack/{toolfoundation,tooldiscovery,toolexec,toolcompose,toolops,toolprotocol}` succeeds
2. Each repo has passing CI badge
3. Release-please PR created on first push
4. Submodules appear in ApertureStack root

---

## Rollback Plan

```bash
# Delete repos if needed
for repo in toolfoundation tooldiscovery toolexec toolcompose toolops toolprotocol; do
  gh repo delete ApertureStack/$repo --yes
done

# Remove submodules
for repo in toolfoundation tooldiscovery toolexec toolcompose toolops toolprotocol; do
  git submodule deinit -f $repo
  git rm -f $repo
  rm -rf .git/modules/$repo
done
```

---

## Next Steps

- PRD-111: CI/CD Templates (reusable workflows)
- PRD-112: GitHub Org Config (secrets)
- PRD-113: Release Automation (multi-package)
