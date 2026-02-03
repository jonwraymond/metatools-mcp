# PRD-111: CI/CD Templates

**Phase:** 1 - Infrastructure Setup
**Priority:** High
**Effort:** 4 hours
**Dependencies:** PRD-110

---

## Objective

Create reusable GitHub Actions workflow templates that will be used across all 6 consolidated repositories, ensuring consistent CI/CD practices.

---

## Deliverables

| Deliverable | Location | Description |
|-------------|----------|-------------|
| CI Template | `.github/workflow-templates/ci.yml` | Test, build, coverage |
| Lint/Security | `.github/workflow-templates/lint-security.yml` | Linting + security scan |
| Commitlint | `.github/workflow-templates/commitlint.yml` | Conventional commits |
| Release Please | `.github/workflow-templates/release-please.yml` | Automated releases |
| Dependency Review | `.github/workflow-templates/dependency-review.yml` | Dependency checks |

---

## Tasks

### Task 1: Create Workflow Templates Directory

```bash
mkdir -p .github/workflow-templates
```

### Task 2: CI Workflow Template

**File:** `.github/workflow-templates/ci.yml`

```yaml
# Template for ci.yml in all consolidated repos
# Copy to .github/workflows/ci.yml

name: CI

on:
  push:
    branches: ["main"]
    paths-ignore:
      - "**.md"
      - "docs/**"
      - "examples/**"
  pull_request:
    paths-ignore:
      - "**.md"
      - "docs/**"
      - "examples/**"

permissions:
  contents: read

concurrency:
  group: ${{ github.workflow }}-${{ github.ref }}
  cancel-in-progress: true

env:
  GO_VERSION: "1.24"
  GOPRIVATE: "github.com/ApertureStack/*"

jobs:
  test:
    name: Test
    runs-on: ubuntu-latest
    strategy:
      fail-fast: false
      matrix:
        go: ["1.23", "1.24"]
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ matrix.go }}
          cache: true

      - name: Configure Git for private repos
        run: |
          git config --global url."https://${{ secrets.GH_TOKEN }}@github.com/".insteadOf "https://github.com/"

      - name: Download dependencies
        run: go mod download

      - name: Verify dependencies
        run: go mod verify

      - name: Build
        run: go build -v ./...

      - name: Test with coverage
        run: go test -race -coverprofile=coverage.out -covermode=atomic -v ./...

      - name: Upload coverage
        if: matrix.go == '1.24'
        uses: codecov/codecov-action@v4
        with:
          token: ${{ secrets.CODECOV_TOKEN }}
          files: coverage.out
          fail_ci_if_error: false
          verbose: true

  integration:
    name: Integration Tests
    runs-on: ubuntu-latest
    needs: test
    if: github.event_name == 'push' && github.ref == 'refs/heads/main'
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
          cache: true

      - name: Run integration tests
        run: go test -tags=integration -v ./...
        env:
          INTEGRATION_TEST: "true"
```

### Task 3: Lint & Security Workflow Template

**File:** `.github/workflow-templates/lint-security.yml`

```yaml
# Template for lint-security.yml in all consolidated repos
# Copy to .github/workflows/lint-security.yml

name: Lint & Security

on:
  push:
    branches: ["main"]
  pull_request:

permissions:
  contents: read
  security-events: write

concurrency:
  group: ${{ github.workflow }}-${{ github.ref }}
  cancel-in-progress: true

jobs:
  lint:
    name: Lint
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
          args: --timeout=5m --config=.golangci.yml

  security:
    name: Security Scan
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
          args: -fmt sarif -out gosec.sarif -exclude-dir=examples ./...

      - name: Upload SARIF
        uses: github/codeql-action/upload-sarif@v3
        if: always()
        with:
          sarif_file: gosec.sarif

  govulncheck:
    name: Vulnerability Check
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version-file: go.mod

      - name: Install govulncheck
        run: go install golang.org/x/vuln/cmd/govulncheck@latest

      - name: Run govulncheck
        run: govulncheck ./...
```

### Task 4: Commitlint Workflow Template

**File:** `.github/workflow-templates/commitlint.yml`

```yaml
# Template for commitlint.yml in all consolidated repos
# Copy to .github/workflows/commitlint.yml

name: Commitlint

on:
  push:
    branches: ["main"]
  pull_request:

permissions:
  contents: read

jobs:
  commitlint:
    name: Lint Commits
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

      - name: Validate current commit (push)
        if: github.event_name == 'push'
        run: npx commitlint --from HEAD~1 --to HEAD --verbose

      - name: Validate PR commits
        if: github.event_name == 'pull_request'
        run: npx commitlint --from ${{ github.event.pull_request.base.sha }} --to ${{ github.event.pull_request.head.sha }} --verbose
```

### Task 5: Release Please Workflow Template

**File:** `.github/workflow-templates/release-please.yml`

```yaml
# Template for release-please.yml in all consolidated repos
# Copy to .github/workflows/release-please.yml

name: Release Please

on:
  push:
    branches: ["main"]

permissions:
  contents: write
  pull-requests: write

jobs:
  release-please:
    name: Release
    runs-on: ubuntu-latest
    outputs:
      release_created: ${{ steps.release.outputs.release_created }}
      tag_name: ${{ steps.release.outputs.tag_name }}
      version: ${{ steps.release.outputs.version }}
    steps:
      - name: Release Please
        id: release
        uses: google-github-actions/release-please-action@v4
        with:
          token: ${{ secrets.GITHUB_TOKEN }}
          config-file: release-please-config.json
          manifest-file: .release-please-manifest.json

      - name: Checkout (on release)
        if: steps.release.outputs.release_created
        uses: actions/checkout@v4

      - name: Set up Go (on release)
        if: steps.release.outputs.release_created
        uses: actions/setup-go@v5
        with:
          go-version-file: go.mod

      - name: Verify module (on release)
        if: steps.release.outputs.release_created
        run: |
          go mod verify
          go build ./...
          go test ./...
```

### Task 6: Dependency Review Workflow Template

**File:** `.github/workflow-templates/dependency-review.yml`

```yaml
# Template for dependency-review.yml in all consolidated repos
# Copy to .github/workflows/dependency-review.yml

name: Dependency Review

on:
  pull_request:
    paths:
      - "go.mod"
      - "go.sum"

permissions:
  contents: read
  pull-requests: write

jobs:
  dependency-review:
    name: Review Dependencies
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Dependency Review
        uses: actions/dependency-review-action@v4
        with:
          fail-on-severity: high
          deny-licenses: GPL-3.0, AGPL-3.0
          allow-licenses: MIT, Apache-2.0, BSD-2-Clause, BSD-3-Clause, ISC
```

### Task 7: Create Template README

**File:** `.github/workflow-templates/README.md`

```markdown
# Workflow Templates

Reusable GitHub Actions workflows for ApertureStack consolidated repositories.

## Usage

Copy the needed workflow files to `.github/workflows/` in each repository:

```bash
# From repository root
cp /path/to/workflow-templates/*.yml .github/workflows/
```

## Workflows

| Workflow | Triggers | Purpose |
|----------|----------|---------|
| `ci.yml` | push, PR | Build, test, coverage |
| `lint-security.yml` | push, PR | golangci-lint, gosec, govulncheck |
| `commitlint.yml` | push, PR | Conventional commits |
| `release-please.yml` | push to main | Automated releases |
| `dependency-review.yml` | PR (go.mod) | License + vulnerability check |

## Required Secrets

| Secret | Scope | Purpose |
|--------|-------|---------|
| `CODECOV_TOKEN` | Org | Coverage upload |
| `GH_TOKEN` | Org | Private repo access |

## Customization

Each workflow includes comments for customization points:
- Go version matrix
- Test tags
- Lint configuration
- Security exclusions
```

---

## Verification Checklist

- [ ] All 5 workflow templates created
- [ ] YAML syntax valid
- [ ] Consistent naming conventions
- [ ] Required secrets documented
- [ ] README explains usage
- [ ] Templates tested locally with `act` (optional)

**Validation:**
```bash
# Validate YAML syntax
for f in .github/workflow-templates/*.yml; do
  python3 -c "import yaml; yaml.safe_load(open('$f'))" && echo "✓ $f" || echo "✗ $f"
done
```

---

## Acceptance Criteria

1. All templates are valid GitHub Actions syntax
2. Workflows can be copied directly to repos
3. Documentation is complete
4. Secrets requirements are documented

---

## Rollback Plan

```bash
rm -rf .github/workflow-templates/
```

---

## Next Steps

- PRD-112: GitHub Org Config (configure secrets)
- PRD-113: Release Automation (release-please config)
