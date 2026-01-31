# PRD-142: Migrate toolcode

**Phase:** 4 - Execution Layer
**Priority:** High
**Effort:** 4 hours
**Dependencies:** PRD-140

---

## Objective

Migrate the existing `toolcode` repository into `toolexec/code/` as the third package in the consolidated execution layer.

---

## Source Analysis

**Current Location:** `github.com/ApertureStack/toolcode`
**Target Location:** `github.com/ApertureStack/toolexec/code`

**Package Contents:**
- Code-based tool orchestration
- Dynamic tool generation from code
- TypeScript/JavaScript execution integration
- Tool chain composition via code
- ~2,000 lines of code

---

## Deliverables

| Deliverable | Location | Description |
|-------------|----------|-------------|
| Code Package | `toolexec/code/` | Code orchestration |
| Tests | `toolexec/code/*_test.go` | All existing tests |
| Documentation | `toolexec/code/doc.go` | Package documentation |

---

## Tasks

### Task 1: Clone and Analyze Source

```bash
cd /tmp/migration
git clone git@github.com:ApertureStack/toolcode.git
cd toolcode

ls -la
wc -l *.go
go test ./...
```

### Task 2: Copy Source Files

```bash
cd /tmp/migration/toolexec

mkdir -p code
cp ../toolcode/*.go code/

ls -la code/
```

### Task 3: Update Import Paths

```bash
cd /tmp/migration/toolexec/code

# Update self-reference
OLD_IMPORT="github.com/ApertureStack/toolcode"
NEW_IMPORT="github.com/ApertureStack/toolexec/code"

for file in *.go; do
  sed -i '' "s|$OLD_IMPORT|$NEW_IMPORT|g" "$file"
done

# Update toolmodel to toolfoundation/model
sed -i '' 's|github.com/ApertureStack/toolmodel|github.com/ApertureStack/toolfoundation/model|g' *.go

# Update toolrun to toolexec/run
sed -i '' 's|github.com/ApertureStack/toolrun|github.com/ApertureStack/toolexec/run|g' *.go

# Verify
grep -r "ApertureStack/toolcode\|ApertureStack/toolmodel\|ApertureStack/toolrun" . || echo "✓ All imports updated"
```

### Task 4: Update Package Documentation

**File:** `toolexec/code/doc.go`

```go
// Package code provides code-based tool orchestration for the ApertureStack ecosystem.
//
// This package enables developers to define tool workflows using code, supporting
// dynamic tool generation, conditional execution, and complex orchestration patterns.
//
// # Overview
//
// While toolexec/run handles individual tool execution and chains, the code package
// provides a higher-level abstraction for code-driven orchestration:
//
//   - Dynamic tool generation from runtime data
//   - Conditional branching based on execution results
//   - Parallel execution with result aggregation
//   - Error handling and retry logic
//
// # Usage
//
// Create a code orchestrator:
//
//	orchestrator := code.NewOrchestrator(code.Config{
//	    Runner: runner,
//	    Logger: logger,
//	})
//
// Define and execute a workflow:
//
//	workflow := code.Workflow{
//	    Name: "data-pipeline",
//	    Steps: []code.Step{
//	        {Tool: "fetch", Input: fetchInput},
//	        {Tool: "transform", DependsOn: []string{"fetch"}},
//	        {Tool: "store", DependsOn: []string{"transform"}},
//	    },
//	}
//
//	results, err := orchestrator.Execute(ctx, workflow)
//
// # TypeScript Integration
//
// For TypeScript-based orchestration, see the toolcodeengine companion:
//
//	engine := codeengine.New(codeengine.Config{
//	    Orchestrator: orchestrator,
//	    Runtime:      "deno",
//	})
//
//	result, err := engine.Execute(ctx, `
//	    const data = await tools.fetch({url: "..."});
//	    return tools.transform({data});
//	`)
//
// # Parallel Execution
//
// Execute multiple tools in parallel:
//
//	parallel := code.Parallel{
//	    Steps: []code.Step{
//	        {Tool: "api-a", Input: inputA},
//	        {Tool: "api-b", Input: inputB},
//	        {Tool: "api-c", Input: inputC},
//	    },
//	}
//
//	results, err := orchestrator.ExecuteParallel(ctx, parallel)
//
// # Migration Note
//
// This package was migrated from github.com/ApertureStack/toolcode as part of
// the ApertureStack consolidation.
package code
```

### Task 5: Verify Internal Dependencies

```bash
cd /tmp/migration/toolexec

# The code package depends on run package
grep -h "import" code/*.go | sort -u

# Should include:
# "github.com/ApertureStack/toolexec/run"
```

### Task 6: Build and Test

```bash
cd /tmp/migration/toolexec

go mod tidy
go build ./...
go test -v -coverprofile=code_coverage.out ./code/...

go tool cover -func=code_coverage.out | grep total
```

### Task 7: Commit and Push

```bash
cd /tmp/migration/toolexec

git add -A
git commit -m "feat(code): migrate toolcode package

Migrate code-based orchestration from standalone toolcode repository.

Package contents:
- Orchestrator for code-driven workflows
- Step-based workflow definition
- Parallel execution support
- Conditional branching
- Result aggregation

Features:
- Dynamic tool generation
- Dependency-based execution order
- Error handling and retries
- TypeScript/JavaScript integration (via codeengine)

Dependencies:
- github.com/ApertureStack/toolfoundation/model
- github.com/ApertureStack/toolexec/run

This is part of the ApertureStack consolidation effort.

Migration: github.com/ApertureStack/toolcode → toolexec/code

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"

git push origin main
```

---

## Key Interfaces

```go
package code

import (
    "context"
    "github.com/ApertureStack/toolexec/run"
)

// Orchestrator manages code-based tool workflows.
type Orchestrator interface {
    // Execute runs a workflow.
    Execute(ctx context.Context, workflow Workflow) ([]run.Result, error)

    // ExecuteParallel runs steps in parallel.
    ExecuteParallel(ctx context.Context, parallel Parallel) ([]run.Result, error)

    // ExecuteConditional runs steps conditionally.
    ExecuteConditional(ctx context.Context, cond Conditional) (*run.Result, error)
}

// Workflow represents a sequence of steps.
type Workflow struct {
    Name  string
    Steps []Step
}

// Step represents a workflow step.
type Step struct {
    ID        string
    Tool      string
    Input     map[string]any
    DependsOn []string
    Condition string // Expression for conditional execution
    Retry     *RetryConfig
}

// Parallel represents parallel execution.
type Parallel struct {
    Steps   []Step
    MaxWorkers int
}

// Conditional represents conditional execution.
type Conditional struct {
    Condition string
    IfTrue    Step
    IfFalse   *Step
}

// Config configures the orchestrator.
type Config struct {
    Runner run.Runner
    Logger Logger
    MaxParallel int
}
```

---

## Verification Checklist

- [ ] All source files copied
- [ ] Import paths updated
- [ ] Dependency on toolexec/run works
- [ ] `go build ./...` succeeds
- [ ] `go test ./...` passes
- [ ] Workflow execution works
- [ ] Parallel execution works
- [ ] Package documentation updated

---

## Acceptance Criteria

1. `toolexec/code` package builds successfully
2. All tests pass
3. Workflow execution produces correct results
4. Parallel execution respects MaxWorkers
5. Conditional execution works

---

## Rollback Plan

```bash
cd /tmp/migration/toolexec
rm -rf code/
git checkout HEAD~1 -- .
git push origin main --force-with-lease
```

---

## Next Steps

- PRD-143: Extract toolbackend
- Gate G3: Execution layer complete
