-- Migration 001: Initial schema
-- Tables: config, simulations, feedback, audit_log, agent_memory

CREATE TABLE IF NOT EXISTS config (
    key        TEXT PRIMARY KEY,
    value      TEXT NOT NULL DEFAULT '',
    updated_at INTEGER NOT NULL DEFAULT (unixepoch())
);

CREATE TABLE IF NOT EXISTS simulations (
    id          TEXT PRIMARY KEY,
    question    TEXT NOT NULL,
    department  TEXT NOT NULL DEFAULT 'market',
    rounds      INTEGER NOT NULL DEFAULT 20,
    result_json TEXT,
    created_at  INTEGER NOT NULL DEFAULT (unixepoch())
);
CREATE INDEX IF NOT EXISTS idx_simulations_created ON simulations(created_at);

CREATE TABLE IF NOT EXISTS feedback (
    id            TEXT PRIMARY KEY,
    simulation_id TEXT NOT NULL,
    company_id    TEXT NOT NULL DEFAULT '',
    predicted     TEXT NOT NULL DEFAULT '',
    actual        TEXT NOT NULL DEFAULT '',
    delta_score   REAL NOT NULL DEFAULT 0.0,
    outcome       TEXT NOT NULL DEFAULT '',
    notes         TEXT,
    recorded_at   INTEGER NOT NULL DEFAULT (unixepoch()),
    created_at    INTEGER NOT NULL DEFAULT (unixepoch())
);
CREATE INDEX IF NOT EXISTS idx_feedback_sim ON feedback(simulation_id);

CREATE TABLE IF NOT EXISTS audit_log (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    event_type TEXT NOT NULL,
    entity_id  TEXT,
    payload    TEXT,
    hmac_sig   TEXT NOT NULL DEFAULT '',
    created_at INTEGER NOT NULL DEFAULT (unixepoch())
);
CREATE INDEX IF NOT EXISTS idx_audit_created ON audit_log(created_at);

CREATE TABLE IF NOT EXISTS agent_memory (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    agent_id   TEXT NOT NULL,
    content    TEXT NOT NULL,
    embedding  BLOB,
    created_at INTEGER NOT NULL DEFAULT (unixepoch())
);
CREATE INDEX IF NOT EXISTS idx_memory_agent ON agent_memory(agent_id);
