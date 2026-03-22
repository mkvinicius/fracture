-- FRACTURE Database Schema
-- Simplified schema aligned with db.go helpers

-- ─── CONFIG ──────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS config (
    key        TEXT PRIMARY KEY,
    value      TEXT NOT NULL DEFAULT '',
    updated_at INTEGER NOT NULL DEFAULT (unixepoch())
);

-- ─── SIMULATIONS ─────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS simulations (
    id          TEXT PRIMARY KEY,
    question    TEXT NOT NULL,
    department  TEXT NOT NULL DEFAULT 'market',
    rounds      INTEGER NOT NULL DEFAULT 20,
    result_json TEXT,           -- Full JSON result (report + simulation data)
    created_at  INTEGER NOT NULL DEFAULT (unixepoch())
);

CREATE INDEX IF NOT EXISTS idx_simulations_created ON simulations(created_at);

-- ─── FEEDBACK ────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS feedback (
    simulation_id TEXT PRIMARY KEY,
    outcome       TEXT NOT NULL,   -- accurate | inaccurate | partial
    notes         TEXT,
    created_at    INTEGER NOT NULL DEFAULT (unixepoch())
);

-- ─── AUDIT LOG (append-only) ─────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS audit_log (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    event_type TEXT NOT NULL,
    entity_id  TEXT,
    payload    TEXT,
    hmac_sig   TEXT NOT NULL DEFAULT '',
    created_at INTEGER NOT NULL DEFAULT (unixepoch())
);

CREATE INDEX IF NOT EXISTS idx_audit_created ON audit_log(created_at);

-- ─── MEMORY (agent memory store) ─────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS agent_memory (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    agent_id   TEXT NOT NULL,
    content    TEXT NOT NULL,
    embedding  BLOB,
    created_at INTEGER NOT NULL DEFAULT (unixepoch())
);

CREATE INDEX IF NOT EXISTS idx_memory_agent ON agent_memory(agent_id);

-- ─── ARCHETYPE CALIBRATION ───────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS archetype_calibration (
    archetype_id    TEXT NOT NULL,
    domain          TEXT NOT NULL,
    accuracy_weight REAL NOT NULL DEFAULT 1.0,
    sample_count    INTEGER NOT NULL DEFAULT 0,
    updated_at      INTEGER NOT NULL DEFAULT (unixepoch()),
    PRIMARY KEY (archetype_id, domain)
);
