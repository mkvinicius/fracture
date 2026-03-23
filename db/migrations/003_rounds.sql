-- Migration 003: Simulation execution trail
-- Adds simulation_rounds, fracture_votes, and report_generations
-- for replay, debug, comparison between runs, and future fine-tuning

-- ─── SIMULATION ROUNDS ────────────────────────────────────────────────────────
-- One row per agent action per round. Enables full replay and per-agent analytics.
CREATE TABLE IF NOT EXISTS simulation_rounds (
    id                TEXT PRIMARY KEY,
    simulation_id     TEXT NOT NULL REFERENCES simulations(id) ON DELETE CASCADE,
    round_number      INTEGER NOT NULL,
    agent_id          TEXT NOT NULL,
    agent_type        TEXT NOT NULL,
    action_text       TEXT NOT NULL,
    tension_level     REAL NOT NULL DEFAULT 0.0,
    fracture_proposed INTEGER NOT NULL DEFAULT 0,  -- 0 | 1
    fracture_accepted INTEGER,                      -- NULL until voted, then 0 | 1
    new_rule_json     TEXT,                         -- JSON of FractureProposal if proposed
    tokens_used       INTEGER NOT NULL DEFAULT 0,
    created_at        INTEGER NOT NULL DEFAULT (unixepoch())
);
CREATE INDEX IF NOT EXISTS idx_rounds_sim     ON simulation_rounds(simulation_id);
CREATE INDEX IF NOT EXISTS idx_rounds_agent   ON simulation_rounds(agent_id);
CREATE INDEX IF NOT EXISTS idx_rounds_round   ON simulation_rounds(simulation_id, round_number);

-- ─── FRACTURE VOTES ───────────────────────────────────────────────────────────
-- One row per agent vote on each fracture proposal.
CREATE TABLE IF NOT EXISTS fracture_votes (
    id            TEXT PRIMARY KEY,
    simulation_id TEXT NOT NULL REFERENCES simulations(id) ON DELETE CASCADE,
    round_number  INTEGER NOT NULL,
    proposal_id   TEXT NOT NULL,   -- references simulation_rounds.id where fracture_proposed=1
    voter_id      TEXT NOT NULL,
    voter_type    TEXT NOT NULL,
    vote          INTEGER NOT NULL,  -- 1 = yes, 0 = no
    weight        REAL NOT NULL DEFAULT 1.0,
    reasoning     TEXT,
    created_at    INTEGER NOT NULL DEFAULT (unixepoch())
);
CREATE INDEX IF NOT EXISTS idx_votes_sim      ON fracture_votes(simulation_id);
CREATE INDEX IF NOT EXISTS idx_votes_proposal ON fracture_votes(proposal_id);

-- ─── REPORT GENERATIONS ──────────────────────────────────────────────────────
-- Tracks each report generation attempt with timing and token usage.
CREATE TABLE IF NOT EXISTS report_generations (
    id            TEXT PRIMARY KEY,
    simulation_id TEXT NOT NULL REFERENCES simulations(id) ON DELETE CASCADE,
    report_type   TEXT NOT NULL,   -- probable_future | rupture_scenarios | coalitions | timeline | playbook | full
    status        TEXT NOT NULL DEFAULT 'started',  -- started | done | error
    tokens_used   INTEGER NOT NULL DEFAULT 0,
    duration_ms   INTEGER NOT NULL DEFAULT 0,
    error_msg     TEXT,
    created_at    INTEGER NOT NULL DEFAULT (unixepoch()),
    completed_at  INTEGER
);
CREATE INDEX IF NOT EXISTS idx_report_gen_sim ON report_generations(simulation_id);
