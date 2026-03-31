-- Migration 009: Unicorn features
-- Share tokens, prediction outcomes, scheduled simulations, API keys

ALTER TABLE simulations ADD COLUMN share_token TEXT NOT NULL DEFAULT '';
ALTER TABLE simulations ADD COLUMN company_size TEXT NOT NULL DEFAULT '';
ALTER TABLE simulations ADD COLUMN company_sector TEXT NOT NULL DEFAULT '';
ALTER TABLE simulations ADD COLUMN company_location TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_simulations_share ON simulations(share_token);

CREATE TABLE IF NOT EXISTS prediction_outcomes (
    id              TEXT PRIMARY KEY,
    simulation_id   TEXT NOT NULL,
    fracture_event_round INTEGER NOT NULL DEFAULT 0,
    rule_id         TEXT NOT NULL,
    prediction      TEXT NOT NULL,
    outcome         TEXT NOT NULL DEFAULT 'pending',
    notes           TEXT NOT NULL DEFAULT '',
    validated_at    INTEGER,
    created_at      INTEGER NOT NULL DEFAULT (unixepoch())
);

CREATE INDEX IF NOT EXISTS idx_outcomes_sim ON prediction_outcomes(simulation_id);
CREATE INDEX IF NOT EXISTS idx_outcomes_outcome ON prediction_outcomes(outcome);

CREATE TABLE IF NOT EXISTS scheduled_simulations (
    id          TEXT PRIMARY KEY,
    question    TEXT NOT NULL,
    department  TEXT NOT NULL DEFAULT 'market',
    rounds      INTEGER NOT NULL DEFAULT 20,
    context     TEXT NOT NULL DEFAULT '',
    interval_h  INTEGER NOT NULL DEFAULT 168,
    enabled     INTEGER NOT NULL DEFAULT 1,
    last_run_at INTEGER,
    next_run_at INTEGER NOT NULL DEFAULT (unixepoch()),
    created_at  INTEGER NOT NULL DEFAULT (unixepoch())
);

CREATE TABLE IF NOT EXISTS api_keys (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    key_hash    TEXT NOT NULL UNIQUE,
    key_prefix  TEXT NOT NULL,
    sims_used   INTEGER NOT NULL DEFAULT 0,
    sims_limit  INTEGER NOT NULL DEFAULT 0,
    enabled     INTEGER NOT NULL DEFAULT 1,
    created_at  INTEGER NOT NULL DEFAULT (unixepoch()),
    last_used_at INTEGER
);

CREATE INDEX IF NOT EXISTS idx_api_keys_hash ON api_keys(key_hash);
