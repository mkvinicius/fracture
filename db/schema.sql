-- FRACTURE Database Schema v1.1
-- SQLite, stored in platform data dir

PRAGMA journal_mode=WAL;
PRAGMA foreign_keys=ON;

-- ─── CONFIG ──────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS config (
    key        TEXT PRIMARY KEY,
    value      TEXT NOT NULL,
    updated_at INTEGER NOT NULL DEFAULT (unixepoch())
);

-- ─── SIMULATIONS ─────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS simulations (
    id          TEXT PRIMARY KEY,
    question    TEXT NOT NULL,
    department  TEXT NOT NULL DEFAULT 'market',
    rounds      INTEGER NOT NULL DEFAULT 20,
    result_json TEXT,
    created_at  INTEGER NOT NULL DEFAULT (unixepoch())
);

-- ─── FEEDBACK ────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS feedback (
    simulation_id TEXT PRIMARY KEY,
    outcome       TEXT NOT NULL,  -- accurate | inaccurate | partial
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

-- ─── INDEXES ─────────────────────────────────────────────────────────────────

CREATE INDEX IF NOT EXISTS idx_simulations_created ON simulations(created_at);
CREATE INDEX IF NOT EXISTS idx_audit_created ON audit_log(created_at);
