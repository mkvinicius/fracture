CREATE TABLE IF NOT EXISTS confirmed_ruptures (
    id               TEXT PRIMARY KEY,
    simulation_id    TEXT NOT NULL,
    rule_id          TEXT NOT NULL DEFAULT '',
    rule_description TEXT NOT NULL DEFAULT '',
    notes            TEXT NOT NULL DEFAULT '',
    confirmed_at     INTEGER NOT NULL DEFAULT (unixepoch())
);

CREATE INDEX IF NOT EXISTS idx_confirmed_ruptures_sim ON confirmed_ruptures(simulation_id);
