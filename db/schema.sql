-- FRACTURE Database Schema
-- SQLite, stored at ~/.fracture/data.db

PRAGMA journal_mode=WAL;
PRAGMA foreign_keys=ON;

-- ─────────────────────────────────────────
-- CONFIGURATION
-- ─────────────────────────────────────────

CREATE TABLE IF NOT EXISTS config (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at INTEGER NOT NULL DEFAULT (unixepoch())
);

-- ─────────────────────────────────────────
-- COMPANY PROFILE
-- ─────────────────────────────────────────

CREATE TABLE IF NOT EXISTS company (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    sector      TEXT NOT NULL,
    size        TEXT NOT NULL,  -- startup | small | medium | large | enterprise
    description TEXT,
    competitors TEXT,           -- JSON array of competitor names
    product     TEXT,
    created_at  INTEGER NOT NULL DEFAULT (unixepoch()),
    updated_at  INTEGER NOT NULL DEFAULT (unixepoch())
);

-- ─────────────────────────────────────────
-- RULES (World Graph)
-- ─────────────────────────────────────────

CREATE TABLE IF NOT EXISTS rules (
    id          TEXT PRIMARY KEY,
    company_id  TEXT REFERENCES company(id),  -- NULL = global/sector rule
    description TEXT NOT NULL,
    domain      TEXT NOT NULL,  -- market | technology | regulation | behavior | culture
    stability   REAL NOT NULL DEFAULT 0.7,    -- 0.0 (fragile) to 1.0 (immutable)
    sector      TEXT,
    is_active   INTEGER NOT NULL DEFAULT 1,
    created_at  INTEGER NOT NULL DEFAULT (unixepoch()),
    updated_at  INTEGER NOT NULL DEFAULT (unixepoch())
);

CREATE TABLE IF NOT EXISTS rule_relations (
    rule_id     TEXT NOT NULL REFERENCES rules(id),
    depends_on  TEXT NOT NULL REFERENCES rules(id),
    strength    REAL NOT NULL DEFAULT 0.5,
    PRIMARY KEY (rule_id, depends_on)
);

-- ─────────────────────────────────────────
-- ARCHETYPES
-- ─────────────────────────────────────────

CREATE TABLE IF NOT EXISTS archetypes (
    id              TEXT PRIMARY KEY,
    company_id      TEXT REFERENCES company(id),  -- NULL = global archetype
    name            TEXT NOT NULL,
    type            TEXT NOT NULL,  -- conformist | disruptor
    domain          TEXT,           -- marketing | hr | finance | sales | product | strategy
    personality     TEXT NOT NULL,  -- JSON: traits, goals, biases, power_weight
    memory_weight   REAL NOT NULL DEFAULT 1.0,  -- calibration multiplier
    is_active       INTEGER NOT NULL DEFAULT 1,
    created_at      INTEGER NOT NULL DEFAULT (unixepoch()),
    updated_at      INTEGER NOT NULL DEFAULT (unixepoch())
);

-- ─────────────────────────────────────────
-- SIMULATIONS
-- ─────────────────────────────────────────

CREATE TABLE IF NOT EXISTS simulations (
    id              TEXT PRIMARY KEY,
    company_id      TEXT NOT NULL REFERENCES company(id),
    title           TEXT NOT NULL,
    question        TEXT NOT NULL,
    department      TEXT,
    template_id     TEXT,
    rules_snapshot  TEXT NOT NULL,  -- JSON snapshot of rules at simulation time
    agents_snapshot TEXT NOT NULL,  -- JSON snapshot of archetypes used
    rounds_target   INTEGER NOT NULL DEFAULT 20,
    rounds_done     INTEGER NOT NULL DEFAULT 0,
    status          TEXT NOT NULL DEFAULT 'pending',  -- pending | running | done | error
    model_config    TEXT NOT NULL,  -- JSON: which model for each role
    started_at      INTEGER,
    completed_at    INTEGER,
    created_at      INTEGER NOT NULL DEFAULT (unixepoch())
);

-- ─────────────────────────────────────────
-- SIMULATION ROUNDS
-- ─────────────────────────────────────────

CREATE TABLE IF NOT EXISTS simulation_rounds (
    id                  TEXT PRIMARY KEY,
    simulation_id       TEXT NOT NULL REFERENCES simulations(id),
    round_number        INTEGER NOT NULL,
    agent_id            TEXT NOT NULL,
    agent_type          TEXT NOT NULL,  -- conformist | disruptor
    action_text         TEXT NOT NULL,
    tension_level       REAL NOT NULL DEFAULT 0.0,
    fracture_proposed   INTEGER NOT NULL DEFAULT 0,
    fracture_rule_id    TEXT,
    fracture_accepted   INTEGER,
    new_rule_json       TEXT,           -- JSON of proposed new rule if fracture
    tokens_used         INTEGER NOT NULL DEFAULT 0,
    created_at          INTEGER NOT NULL DEFAULT (unixepoch())
);

-- ─────────────────────────────────────────
-- SIMULATION RESULTS
-- ─────────────────────────────────────────

CREATE TABLE IF NOT EXISTS simulation_results (
    id              TEXT PRIMARY KEY,
    simulation_id   TEXT NOT NULL REFERENCES simulations(id),
    result_type     TEXT NOT NULL,  -- probable_future | tension_map | rupture_scenarios
    content         TEXT NOT NULL,  -- JSON structured result
    confidence      REAL,
    created_at      INTEGER NOT NULL DEFAULT (unixepoch())
);

-- ─────────────────────────────────────────
-- FEEDBACK (Real-world outcomes)
-- ─────────────────────────────────────────

CREATE TABLE IF NOT EXISTS feedback (
    id              TEXT PRIMARY KEY,
    simulation_id   TEXT NOT NULL REFERENCES simulations(id),
    company_id      TEXT NOT NULL REFERENCES company(id),
    predicted       TEXT NOT NULL,  -- what FRACTURE predicted
    actual          TEXT NOT NULL,  -- what actually happened
    delta_score     REAL,           -- accuracy delta (-1.0 to 1.0)
    notes           TEXT,
    recorded_at     INTEGER NOT NULL DEFAULT (unixepoch())
);

-- ─────────────────────────────────────────
-- CAUSALITY GRAPH (learned over time)
-- ─────────────────────────────────────────

CREATE TABLE IF NOT EXISTS causality_nodes (
    id          TEXT PRIMARY KEY,
    company_id  TEXT NOT NULL REFERENCES company(id),
    description TEXT NOT NULL,
    node_type   TEXT NOT NULL,  -- decision | outcome | context
    created_at  INTEGER NOT NULL DEFAULT (unixepoch())
);

CREATE TABLE IF NOT EXISTS causality_edges (
    from_node   TEXT NOT NULL REFERENCES causality_nodes(id),
    to_node     TEXT NOT NULL REFERENCES causality_nodes(id),
    strength    REAL NOT NULL DEFAULT 0.5,
    evidence    INTEGER NOT NULL DEFAULT 1,  -- number of times this edge was observed
    PRIMARY KEY (from_node, to_node)
);

-- ─────────────────────────────────────────
-- AUDIT LOG (append-only, never updated)
-- ─────────────────────────────────────────

CREATE TABLE IF NOT EXISTS audit_log (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    event_type  TEXT NOT NULL,
    entity_id   TEXT,
    payload     TEXT,           -- JSON
    hmac_sig    TEXT NOT NULL,  -- HMAC-SHA256 of (event_type + entity_id + payload + prev_sig)
    created_at  INTEGER NOT NULL DEFAULT (unixepoch())
);

-- ─────────────────────────────────────────
-- TEMPLATES
-- ─────────────────────────────────────────

CREATE TABLE IF NOT EXISTS templates (
    id              TEXT PRIMARY KEY,
    name            TEXT NOT NULL,
    department      TEXT NOT NULL,
    question        TEXT NOT NULL,
    description     TEXT,
    rules_preset    TEXT,   -- JSON array of rule IDs or sector name
    agents_preset   TEXT,   -- JSON array of archetype IDs
    is_builtin      INTEGER NOT NULL DEFAULT 1,
    created_at      INTEGER NOT NULL DEFAULT (unixepoch())
);

-- ─────────────────────────────────────────
-- INDEXES
-- ─────────────────────────────────────────

CREATE INDEX IF NOT EXISTS idx_simulations_company ON simulations(company_id);
CREATE INDEX IF NOT EXISTS idx_simulations_status ON simulations(status);
CREATE INDEX IF NOT EXISTS idx_rounds_simulation ON simulation_rounds(simulation_id);
CREATE INDEX IF NOT EXISTS idx_feedback_company ON feedback(company_id);
CREATE INDEX IF NOT EXISTS idx_audit_created ON audit_log(created_at);
CREATE INDEX IF NOT EXISTS idx_rules_company ON rules(company_id);
CREATE INDEX IF NOT EXISTS idx_archetypes_company ON archetypes(company_id);
