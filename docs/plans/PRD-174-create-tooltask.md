# PRD-174: Create tooltask

**Phase:** 7 - Protocol Layer
**Priority:** High
**Effort:** 10 hours
**Dependencies:** PRD-140, PRD-173
**Status:** Done (2026-02-01)

---

## Objective

Create `toolprotocol/task/` for task lifecycle management across protocols.

---

## Package Contents

- Task creation and tracking
- State machine (pending → running → complete/failed/cancelled)
- Progress updates
- Subscriptions + cancellation

## Implementation Summary

- Implemented `Manager` interface with in-memory store and strict transitions.
- Added subscription channels for progress updates.

---

## Key Implementation

```go
package task

import "context"

// State represents task state.
type State string

const (
    StatePending   State = "pending"
    StateRunning   State = "running"
    StateComplete  State = "complete"
    StateFailed    State = "failed"
    StateCancelled State = "cancelled"
)

// Task represents a long-running task.
type Task struct {
    ID          string
    State       State
    Progress    float64
    Message     string
    Result      any
    Error       error
    CreatedAt   time.Time
    UpdatedAt   time.Time
    CompletedAt *time.Time
}

// Manager manages task lifecycle.
type Manager struct {
    tasks map[string]*Task
    mu    sync.RWMutex
}

func (m *Manager) Create(ctx context.Context, id string) (*Task, error)
func (m *Manager) Get(ctx context.Context, id string) (*Task, error)
func (m *Manager) Update(ctx context.Context, id string, progress float64, message string) error
func (m *Manager) Complete(ctx context.Context, id string, result any) error
func (m *Manager) Fail(ctx context.Context, id string, err error) error
func (m *Manager) Cancel(ctx context.Context, id string) error
func (m *Manager) Subscribe(ctx context.Context, id string) (<-chan *Task, error)
```

---

## Commit Message

```
feat(task): add task lifecycle management

Create task package for long-running operations.

Features:
- Task state machine
- Progress tracking
- Task cancellation
- Event subscription

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
```

---

## Next Steps

- PRD-175: Create toolstream
