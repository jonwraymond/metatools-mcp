package limits

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

func TestSQLiteStore_SaveLoad(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	store, err := NewSQLiteStore(db)
	if err != nil {
		t.Fatalf("NewSQLiteStore: %v", err)
	}

	ctx := context.Background()
	if _, ok, err := store.Load(ctx); err != nil || ok {
		t.Fatalf("Load() before save: ok=%v err=%v", ok, err)
	}

	limits := RuntimeLimits{
		Timeout:       15 * time.Second,
		MaxToolCalls:  20,
		MaxChainSteps: 4,
	}

	if err := store.Save(ctx, limits); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	loaded, ok, err := store.Load(ctx)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if !ok {
		t.Fatalf("Load() ok=false after save")
	}
	if loaded.Timeout != limits.Timeout {
		t.Fatalf("Timeout = %v, want %v", loaded.Timeout, limits.Timeout)
	}
	if loaded.MaxToolCalls != limits.MaxToolCalls {
		t.Fatalf("MaxToolCalls = %d, want %d", loaded.MaxToolCalls, limits.MaxToolCalls)
	}
	if loaded.MaxChainSteps != limits.MaxChainSteps {
		t.Fatalf("MaxChainSteps = %d, want %d", loaded.MaxChainSteps, limits.MaxChainSteps)
	}
}
