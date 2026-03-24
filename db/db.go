package db

import (
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var schemaFS embed.FS

// DB wraps *sql.DB with FRACTURE-specific helpers.
type DB struct {
	*sql.DB
}

// DataDir returns the platform-appropriate data directory for FRACTURE.
func DataDir() (string, error) {
	var base string
	switch runtime.GOOS {
	case "windows":
		base = os.Getenv("APPDATA")
		if base == "" {
			return "", fmt.Errorf("APPDATA not set")
		}
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, "Library", "Application Support")
	default: // linux and others
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, ".local", "share")
	}
	dir := filepath.Join(base, "FRACTURE")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	return dir, nil
}

// Open opens (or creates) the FRACTURE SQLite database and applies the schema.
func Open() (*DB, error) {
	dir, err := DataDir()
	if err != nil {
		return nil, fmt.Errorf("data dir: %w", err)
	}

	dbPath := filepath.Join(dir, "data.db")
	sqlDB, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	// Apply base schema (idempotent — uses CREATE TABLE IF NOT EXISTS)
	schema, err := schemaFS.ReadFile("schema.sql")
	if err != nil {
		return nil, fmt.Errorf("read schema: %w", err)
	}
	if _, err := sqlDB.Exec(string(schema)); err != nil {
		return nil, fmt.Errorf("apply schema: %w", err)
	}

	// Apply versioned migrations (idempotent — skips already-applied ones)
	if err := Migrate(sqlDB); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return &DB{sqlDB}, nil
}

// ─── Config ──────────────────────────────────────────────────────────────────

// GetConfig retrieves a config value by key.
func (d *DB) GetConfig(key string) (string, error) {
	var value string
	err := d.QueryRow(`SELECT value FROM config WHERE key = ?`, key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}

// SetConfig upserts a config key-value pair.
func (d *DB) SetConfig(key, value string) error {
	_, err := d.Exec(`
		INSERT INTO config (key, value, updated_at)
		VALUES (?, ?, unixepoch())
		ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = unixepoch()
	`, key, value)
	return err
}

// IsOnboarded returns true if the user has completed onboarding.
func (d *DB) IsOnboarded() (bool, error) {
	val, err := d.GetConfig("onboarding_complete")
	if err != nil {
		return false, err
	}
	return val == "true", nil
}

// ─── Simulations ─────────────────────────────────────────────────────────────

// SimulationRow is a lightweight DB row for listing simulations.
type SimulationRow struct {
	ID         string `json:"id"`
	Question   string `json:"question"`
	Department string `json:"department"`
	Rounds     int    `json:"rounds"`
	CreatedAt  int64  `json:"created_at"`
	DurationMs int64  `json:"duration_ms"`
	ResultJSON string `json:"-"`
}

// SaveSimulation persists a completed simulation result to the database.
func (d *DB) SaveSimulation(id, question, department string, rounds int, result interface{}) error {
	b, err := json.Marshal(result)
	if err != nil {
		return err
	}
	_, err = d.Exec(`
		INSERT INTO simulations (id, question, department, rounds, result_json, created_at)
		VALUES (?, ?, ?, ?, ?, unixepoch())
		ON CONFLICT(id) DO UPDATE SET result_json = excluded.result_json
	`, id, question, department, rounds, string(b))
	return err
}

// ListSimulations returns a summary list of all simulations from the DB.
func (d *DB) ListSimulations() ([]SimulationRow, error) {
	rows, err := d.Query(`
		SELECT id, question, department, rounds, created_at,
		       COALESCE(json_extract(result_json, '$.duration_ms'), 0)
		FROM simulations ORDER BY created_at DESC LIMIT 100
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []SimulationRow
	for rows.Next() {
		var r SimulationRow
		if err := rows.Scan(&r.ID, &r.Question, &r.Department, &r.Rounds, &r.CreatedAt, &r.DurationMs); err != nil {
			continue
		}
		result = append(result, r)
	}
	return result, nil
}

// GetSimulation returns the full result JSON for a simulation.
func (d *DB) GetSimulation(id string) (*SimulationRow, error) {
	var r SimulationRow
	err := d.QueryRow(`
		SELECT id, question, department, rounds, created_at,
		       COALESCE(json_extract(result_json, '$.duration_ms'), 0),
		       COALESCE(result_json, '{}')
		FROM simulations WHERE id = ?
	`, id).Scan(&r.ID, &r.Question, &r.Department, &r.Rounds, &r.CreatedAt, &r.DurationMs, &r.ResultJSON)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// DeleteSimulation removes a simulation from the DB.
func (d *DB) DeleteSimulation(id string) error {
	_, err := d.Exec(`DELETE FROM simulations WHERE id = ?`, id)
	return err
}

// ─── Simulation Jobs (persistent job state) ───────────────────────────────────────────────────

// JobRow mirrors the simulation_jobs table.
type JobRow struct {
	ID              string `json:"id"`
	Status          string `json:"status"`
	Question        string `json:"question"`
	Department      string `json:"department"`
	Rounds          int    `json:"rounds"`
	Company         string `json:"company,omitempty"`
	Error           string `json:"error,omitempty"`
	ResearchSources int    `json:"research_sources,omitempty"`
	ResearchTokens  int    `json:"research_tokens,omitempty"`
	DurationMs      int64  `json:"duration_ms,omitempty"`
	CreatedAt       int64  `json:"created_at"`
	UpdatedAt       int64  `json:"updated_at"`
	// Live progress fields (updated after each round via persistRound)
	CurrentRound    int     `json:"current_round,omitempty"`
	CurrentTension  float64 `json:"current_tension,omitempty"`
	FractureCount   int     `json:"fracture_count,omitempty"`
	LastAgentName   string  `json:"last_agent_name,omitempty"`
	LastAgentAction string  `json:"last_agent_action,omitempty"`
	TotalTokens     int     `json:"total_tokens,omitempty"`
}

// UpsertJob creates or updates a job row (called on every status transition and after each round).
func (d *DB) UpsertJob(j *JobRow) error {
	_, err := d.Exec(`
			INSERT INTO simulation_jobs
				(id, status, question, department, rounds, company, error_msg,
				 research_sources, research_tokens, duration_ms,
				 current_round, current_tension, fracture_count,
				 last_agent_name, last_agent_action, total_tokens,
				 created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, unixepoch())
			ON CONFLICT(id) DO UPDATE SET
				status           = excluded.status,
				error_msg        = excluded.error_msg,
				research_sources = excluded.research_sources,
				research_tokens  = excluded.research_tokens,
				duration_ms      = excluded.duration_ms,
				current_round    = excluded.current_round,
				current_tension  = excluded.current_tension,
				fracture_count   = excluded.fracture_count,
				last_agent_name  = excluded.last_agent_name,
				last_agent_action = excluded.last_agent_action,
				total_tokens     = excluded.total_tokens,
				updated_at       = unixepoch()
		`, j.ID, j.Status, j.Question, j.Department, j.Rounds, j.Company, j.Error,
		j.ResearchSources, j.ResearchTokens, j.DurationMs,
		j.CurrentRound, j.CurrentTension, j.FractureCount,
		j.LastAgentName, j.LastAgentAction, j.TotalTokens,
		j.CreatedAt)
	return err
}

// GetJob returns a single job row by ID.
func (d *DB) GetJob(id string) (*JobRow, error) {
	var j JobRow
	err := d.QueryRow(`
			SELECT id, status, question, department, rounds, company, error_msg,
			       research_sources, research_tokens, duration_ms, created_at, updated_at,
			       current_round, current_tension, fracture_count,
			       last_agent_name, last_agent_action, total_tokens
			FROM simulation_jobs WHERE id = ?
		`, id).Scan(&j.ID, &j.Status, &j.Question, &j.Department, &j.Rounds, &j.Company, &j.Error,
		&j.ResearchSources, &j.ResearchTokens, &j.DurationMs, &j.CreatedAt, &j.UpdatedAt,
		&j.CurrentRound, &j.CurrentTension, &j.FractureCount,
		&j.LastAgentName, &j.LastAgentAction, &j.TotalTokens)
	if err != nil {
		return nil, err
	}
	return &j, nil
}

// ListJobs returns all jobs ordered by creation time (newest first).
func (d *DB) ListJobs() ([]JobRow, error) {
	rows, err := d.Query(`
		SELECT id, status, question, department, rounds, company, error_msg,
		       research_sources, research_tokens, duration_ms, created_at, updated_at,
		       current_round, current_tension, fracture_count,
		       last_agent_name, last_agent_action, total_tokens
		FROM simulation_jobs ORDER BY created_at DESC LIMIT 200
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []JobRow
	for rows.Next() {
		var j JobRow
		if err := rows.Scan(&j.ID, &j.Status, &j.Question, &j.Department, &j.Rounds, &j.Company, &j.Error,
			&j.ResearchSources, &j.ResearchTokens, &j.DurationMs, &j.CreatedAt, &j.UpdatedAt,
			&j.CurrentRound, &j.CurrentTension, &j.FractureCount,
			&j.LastAgentName, &j.LastAgentAction, &j.TotalTokens); err != nil {
			continue
		}
		result = append(result, j)
	}
	return result, nil
}

// DeleteJob removes a job row by ID.
func (d *DB) DeleteJob(id string) error {
	_, err := d.Exec(`DELETE FROM simulation_jobs WHERE id = ?`, id)
	return err
}

// MarkInterruptedJobsFailed marks any jobs that were left in non-terminal states
// (queued/researching/running) as 'error' with a restart message.
// Call this once at startup to ensure a clean state after an unexpected shutdown.
func (d *DB) MarkInterruptedJobsFailed() (int, error) {
	res, err := d.Exec(`
		UPDATE simulation_jobs
		SET status = 'error',
		    error_msg = 'interrupted: process restarted before simulation completed',
		    updated_at = unixepoch()
		WHERE status IN ('queued', 'researching', 'running')
	`)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return int(n), nil
}

// ─── Domain Contexts ──────────────────────────────────────────────────────────

// DomainContextRow mirrors the domain_contexts table.
type DomainContextRow struct {
	SimulationID      string
	Domain            string
	Context           string
	Signals           string  // JSON array
	StabilityModifier float64
	Confidence        float64
	AffectedRules     string  // JSON array
	CachedAt          int64
}

// SaveDomainContext upserts a domain context row for a simulation.
func (d *DB) SaveDomainContext(simulationID, domain string, row DomainContextRow) error {
	_, err := d.Exec(`
		INSERT INTO domain_contexts
			(simulation_id, domain, context, signals, stability_modifier, confidence, affected_rules, cached_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, unixepoch())
		ON CONFLICT(simulation_id, domain) DO UPDATE SET
			context            = excluded.context,
			signals            = excluded.signals,
			stability_modifier = excluded.stability_modifier,
			confidence         = excluded.confidence,
			affected_rules     = excluded.affected_rules,
			cached_at          = excluded.cached_at
	`, simulationID, domain, row.Context, row.Signals, row.StabilityModifier, row.Confidence, row.AffectedRules)
	return err
}

// GetDomainContexts returns all domain context rows for a simulation.
func (d *DB) GetDomainContexts(simulationID string) ([]DomainContextRow, error) {
	rows, err := d.Query(`
		SELECT simulation_id, domain, context, signals, stability_modifier, confidence, affected_rules, cached_at
		FROM domain_contexts WHERE simulation_id = ?
	`, simulationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []DomainContextRow
	for rows.Next() {
		var r DomainContextRow
		if err := rows.Scan(&r.SimulationID, &r.Domain, &r.Context, &r.Signals,
			&r.StabilityModifier, &r.Confidence, &r.AffectedRules, &r.CachedAt); err != nil {
			continue
		}
		result = append(result, r)
	}
	return result, nil
}

// SaveFeedback stores user feedback for a simulation.
func (d *DB) SaveFeedback(simulationID, outcome, notes string) error {
	_, err := d.Exec(`
		INSERT INTO feedback (simulation_id, outcome, notes, created_at)
		VALUES (?, ?, ?, unixepoch())
		ON CONFLICT(simulation_id) DO UPDATE SET outcome = excluded.outcome, notes = excluded.notes
	`, simulationID, outcome, notes)
	return err
}

// ─── Audit Log ───────────────────────────────────────────────────────────────

// AuditRow is a single audit log entry.
type AuditRow struct {
	ID        int64  `json:"id"`
	Event     string `json:"event"`
	Actor     string `json:"actor"`
	Payload   string `json:"payload"`
	CreatedAt int64  `json:"created_at"`
}

// GetAuditLog returns the most recent N audit log entries.
func (d *DB) GetAuditLog(limit int) ([]AuditRow, error) {
	rows, err := d.Query(`
		SELECT id, event_type, entity_id, COALESCE(payload, '{}'), created_at
		FROM audit_log ORDER BY created_at DESC LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []AuditRow
	for rows.Next() {
		var r AuditRow
		if err := rows.Scan(&r.ID, &r.Event, &r.Actor, &r.Payload, &r.CreatedAt); err != nil {
			continue
		}
		result = append(result, r)
	}
	return result, nil
}
