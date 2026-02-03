CREATE TABLE IF NOT EXISTS runtime_limits (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    timeout_ms INTEGER NOT NULL DEFAULT 0,
    max_tool_calls INTEGER NOT NULL DEFAULT 0,
    max_chain_steps INTEGER NOT NULL DEFAULT 0,
    updated_at TEXT NOT NULL
);
