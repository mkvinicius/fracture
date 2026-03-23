-- Migration 002: Simulation jobs and archetype calibration
-- Adds persistent job state for resilience across restarts
-- and the archetype_calibration table for accuracy tracking

CREATE TABLE IF NOT EXISTS simulation_jobs (
    id               TEXT PRIMARY KEY,
    status           TEXT NOT NULL DEFAULT 'queued',
    question         TEXT NOT NULL,
    department       TEXT NOT NULL DEFAULT 'market',
    rounds           INTEGER NOT NULL DEFAULT 20,
    company          TEXT NOT NULL DEFAULT '',
    error_msg        TEXT NOT NULL DEFAULT '',
    research_sources INTEGER NOT NULL DEFAULT 0,
    research_tokens  INTEGER NOT NULL DEFAULT 0,
    duration_ms      INTEGER NOT NULL DEFAULT 0,
    created_at       INTEGER NOT NULL DEFAULT (unixepoch()),
    updated_at       INTEGER NOT NULL DEFAULT (unixepoch())
);
CREATE INDEX IF NOT EXISTS idx_sim_jobs_status  ON simulation_jobs(status);
CREATE INDEX IF NOT EXISTS idx_sim_jobs_created ON simulation_jobs(created_at);

CREATE TABLE IF NOT EXISTS archetype_calibration (
    archetype_id    TEXT NOT NULL,
    domain          TEXT NOT NULL,
    accuracy_weight REAL NOT NULL DEFAULT 1.0,
    sample_count    INTEGER NOT NULL DEFAULT 0,
    updated_at      INTEGER NOT NULL DEFAULT (unixepoch()),
    PRIMARY KEY (archetype_id, domain)
);
