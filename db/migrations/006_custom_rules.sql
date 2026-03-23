-- Migration 006: Custom rules table
-- Allows companies to define their own world rules that are injected
-- into the simulation alongside the built-in domain rules.
CREATE TABLE IF NOT EXISTS custom_rules (
    id          TEXT PRIMARY KEY,
    company_id  TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL,
    domain      TEXT NOT NULL DEFAULT 'market',
    stability   REAL NOT NULL DEFAULT 0.5,   -- 0.0 (fragile) to 1.0 (immutable)
    is_active   INTEGER NOT NULL DEFAULT 1,
    created_at  INTEGER NOT NULL DEFAULT (unixepoch()),
    updated_at  INTEGER NOT NULL DEFAULT (unixepoch())
);
CREATE INDEX IF NOT EXISTS idx_custom_rules_company ON custom_rules(company_id);
CREATE INDEX IF NOT EXISTS idx_custom_rules_domain  ON custom_rules(domain);
