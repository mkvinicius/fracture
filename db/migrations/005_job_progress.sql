-- Migration 005: Live progress fields on simulation_jobs
-- NOTE: New databases created after this migration was introduced already have
-- these columns in schema.sql (CREATE TABLE IF NOT EXISTS). This migration
-- exists only to add the columns to existing databases that were created before
-- the schema.sql was updated. SQLite does not support IF NOT EXISTS on ALTER TABLE,
-- so we use a workaround: attempt the ALTER and ignore "duplicate column" errors
-- at the application level (see db/migrate.go which uses ignoreDuplicateColumn).
ALTER TABLE simulation_jobs ADD COLUMN current_round    INTEGER NOT NULL DEFAULT 0;
ALTER TABLE simulation_jobs ADD COLUMN current_tension  REAL    NOT NULL DEFAULT 0.0;
ALTER TABLE simulation_jobs ADD COLUMN fracture_count   INTEGER NOT NULL DEFAULT 0;
ALTER TABLE simulation_jobs ADD COLUMN last_agent_name  TEXT    NOT NULL DEFAULT '';
ALTER TABLE simulation_jobs ADD COLUMN last_agent_action TEXT   NOT NULL DEFAULT '';
ALTER TABLE simulation_jobs ADD COLUMN total_tokens     INTEGER NOT NULL DEFAULT 0;
