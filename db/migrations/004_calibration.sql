-- Migration 004: Memory, calibration and causality graph
-- Adds archetypes, causality_nodes, and causality_edges
-- to close the gap between memory/calibration.go and the actual schema

-- ─── ARCHETYPES ───────────────────────────────────────────────────────────────
-- Stores per-company archetype overrides and calibration weights.
-- Built-in archetypes (company_id IS NULL) are never modified.
CREATE TABLE IF NOT EXISTS archetypes (
    id            TEXT PRIMARY KEY,
    company_id    TEXT,                              -- NULL = built-in, non-NULL = company override
    name          TEXT NOT NULL,
    agent_type    TEXT NOT NULL,
    description   TEXT NOT NULL DEFAULT '',
    memory_weight REAL NOT NULL DEFAULT 1.0,         -- calibration multiplier (0.3–2.0)
    is_active     INTEGER NOT NULL DEFAULT 1,
    created_at    INTEGER NOT NULL DEFAULT (unixepoch()),
    updated_at    INTEGER NOT NULL DEFAULT (unixepoch())
);
CREATE INDEX IF NOT EXISTS idx_archetypes_company ON archetypes(company_id);
CREATE INDEX IF NOT EXISTS idx_archetypes_type    ON archetypes(agent_type);

-- ─── CAUSALITY NODES ─────────────────────────────────────────────────────────
-- Nodes in the causal graph: decisions and outcomes observed across simulations.
CREATE TABLE IF NOT EXISTS causality_nodes (
    id          TEXT PRIMARY KEY,
    company_id  TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL,
    node_type   TEXT NOT NULL DEFAULT 'decision',   -- decision | outcome
    created_at  INTEGER NOT NULL DEFAULT (unixepoch())
);
CREATE INDEX IF NOT EXISTS idx_causality_nodes_company ON causality_nodes(company_id);
CREATE INDEX IF NOT EXISTS idx_causality_nodes_type    ON causality_nodes(node_type);

-- ─── CAUSALITY EDGES ─────────────────────────────────────────────────────────
-- Directed edges between nodes with strength and evidence count.
-- ON CONFLICT used by RecordCausality to increment evidence.
CREATE TABLE IF NOT EXISTS causality_edges (
    from_node  TEXT NOT NULL REFERENCES causality_nodes(id) ON DELETE CASCADE,
    to_node    TEXT NOT NULL REFERENCES causality_nodes(id) ON DELETE CASCADE,
    strength   REAL NOT NULL DEFAULT 0.5,   -- 0.0–1.0
    evidence   INTEGER NOT NULL DEFAULT 1,  -- times this path was observed
    PRIMARY KEY (from_node, to_node)
);
CREATE INDEX IF NOT EXISTS idx_causality_edges_from ON causality_edges(from_node);
CREATE INDEX IF NOT EXISTS idx_causality_edges_to   ON causality_edges(to_node);
