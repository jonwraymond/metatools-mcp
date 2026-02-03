package limits

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// SQLiteStore persists runtime limits in SQLite.
type SQLiteStore struct {
	db *sql.DB
}

// OpenSQLite opens a SQLite database and applies migrations.
func OpenSQLite(path string) (*SQLiteStore, func() error, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, nil, fmt.Errorf("open sqlite: %w", err)
	}
	store, err := NewSQLiteStore(db)
	if err != nil {
		_ = db.Close()
		return nil, nil, err
	}
	return store, db.Close, nil
}

// NewSQLiteStore creates a store and applies migrations.
func NewSQLiteStore(db *sql.DB) (*SQLiteStore, error) {
	if db == nil {
		return nil, fmt.Errorf("sqlite db is nil")
	}
	if err := applyMigrations(db); err != nil {
		return nil, err
	}
	return &SQLiteStore{db: db}, nil
}

// Load returns the persisted limits. ok=false if no row exists.
func (s *SQLiteStore) Load(ctx context.Context) (RuntimeLimits, bool, error) {
	if s == nil || s.db == nil {
		return RuntimeLimits{}, false, fmt.Errorf("sqlite store not configured")
	}

	row := s.db.QueryRowContext(ctx, `
		SELECT timeout_ms, max_tool_calls, max_chain_steps
		FROM runtime_limits
		WHERE id = 1`)

	var timeoutMs int64
	var maxCalls int
	var maxSteps int
	if err := row.Scan(&timeoutMs, &maxCalls, &maxSteps); err != nil {
		if err == sql.ErrNoRows {
			return RuntimeLimits{}, false, nil
		}
		return RuntimeLimits{}, false, fmt.Errorf("load runtime limits: %w", err)
	}

	return RuntimeLimits{
		Timeout:       time.Duration(timeoutMs) * time.Millisecond,
		MaxToolCalls:  maxCalls,
		MaxChainSteps: maxSteps,
	}, true, nil
}

// Save upserts the persisted limits.
func (s *SQLiteStore) Save(ctx context.Context, limits RuntimeLimits) error {
	if s == nil || s.db == nil {
		return fmt.Errorf("sqlite store not configured")
	}

	timeoutMs := limits.Timeout.Milliseconds()
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO runtime_limits (id, timeout_ms, max_tool_calls, max_chain_steps, updated_at)
		VALUES (1, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			timeout_ms = excluded.timeout_ms,
			max_tool_calls = excluded.max_tool_calls,
			max_chain_steps = excluded.max_chain_steps,
			updated_at = excluded.updated_at
	`, timeoutMs, limits.MaxToolCalls, limits.MaxChainSteps, time.Now().UTC().Format(time.RFC3339Nano))
	if err != nil {
		return fmt.Errorf("save runtime limits: %w", err)
	}
	return nil
}

func applyMigrations(db *sql.DB) error {
	entries, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		content, err := migrationsFS.ReadFile("migrations/" + name)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}
		statements := strings.Split(string(content), ";")
		for _, stmt := range statements {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}
			if _, err := db.Exec(stmt); err != nil {
				return fmt.Errorf("apply migration %s: %w", name, err)
			}
		}
	}
	return nil
}
