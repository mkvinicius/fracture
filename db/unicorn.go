package db

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"
)

// ─── Share Tokens ─────────────────────────────────────────────────────────────

// GenerateShareToken creates a unique share token for a simulation and persists it.
func (d *DB) GenerateShareToken(simulationID string) (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := hex.EncodeToString(b)
	_, err := d.Exec(`UPDATE simulations SET share_token = ? WHERE id = ?`, token, simulationID)
	if err != nil {
		return "", err
	}
	return token, nil
}

// GetSimulationByShareToken returns the result JSON for a publicly shared simulation.
func (d *DB) GetSimulationByShareToken(token string) (*SimulationRow, error) {
	if token == "" {
		return nil, fmt.Errorf("empty token")
	}
	var r SimulationRow
	err := d.QueryRow(`
		SELECT id, question, department, rounds, created_at,
		       COALESCE(json_extract(result_json, '$.duration_ms'), 0),
		       COALESCE(result_json, '{}')
		FROM simulations WHERE share_token = ?
	`, token).Scan(&r.ID, &r.Question, &r.Department, &r.Rounds, &r.CreatedAt, &r.DurationMs, &r.ResultJSON)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// ─── Prediction Outcomes (Accuracy Tracking) ──────────────────────────────────

// PredictionOutcome tracks whether a simulation prediction came true.
type PredictionOutcome struct {
	ID                  string  `json:"id"`
	SimulationID        string  `json:"simulation_id"`
	FractureEventRound  int     `json:"fracture_event_round"`
	RuleID              string  `json:"rule_id"`
	Prediction          string  `json:"prediction"`
	Outcome             string  `json:"outcome"` // pending | confirmed | refuted | partial
	Notes               string  `json:"notes"`
	ValidatedAt         *int64  `json:"validated_at,omitempty"`
	CreatedAt           int64   `json:"created_at"`
}

// SavePredictionOutcome upserts a prediction outcome record.
func (d *DB) SavePredictionOutcome(p PredictionOutcome) error {
	_, err := d.Exec(`
		INSERT INTO prediction_outcomes
			(id, simulation_id, fracture_event_round, rule_id, prediction, outcome, notes, validated_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, unixepoch())
		ON CONFLICT(id) DO UPDATE SET
			outcome      = excluded.outcome,
			notes        = excluded.notes,
			validated_at = CASE WHEN excluded.outcome != 'pending' THEN unixepoch() ELSE validated_at END
	`, p.ID, p.SimulationID, p.FractureEventRound, p.RuleID, p.Prediction, p.Outcome, p.Notes, p.ValidatedAt)
	return err
}

// GetPredictionOutcomes returns all prediction outcomes for a simulation.
func (d *DB) GetPredictionOutcomes(simulationID string) ([]PredictionOutcome, error) {
	rows, err := d.Query(`
		SELECT id, simulation_id, fracture_event_round, rule_id, prediction,
		       outcome, notes, validated_at, created_at
		FROM prediction_outcomes WHERE simulation_id = ? ORDER BY created_at DESC
	`, simulationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []PredictionOutcome
	for rows.Next() {
		var p PredictionOutcome
		if err := rows.Scan(&p.ID, &p.SimulationID, &p.FractureEventRound, &p.RuleID,
			&p.Prediction, &p.Outcome, &p.Notes, &p.ValidatedAt, &p.CreatedAt); err != nil {
			continue
		}
		result = append(result, p)
	}
	return result, rows.Err()
}

// AccuracyStats returns accuracy statistics across all validated predictions.
type AccuracyStats struct {
	Total     int     `json:"total"`
	Confirmed int     `json:"confirmed"`
	Refuted   int     `json:"refuted"`
	Partial   int     `json:"partial"`
	Pending   int     `json:"pending"`
	Score     float64 `json:"score"` // (confirmed + 0.5*partial) / (confirmed+refuted+partial)
}

// GetAccuracyStats returns global accuracy statistics.
func (d *DB) GetAccuracyStats() (*AccuracyStats, error) {
	rows, err := d.Query(`
		SELECT outcome, COUNT(*) FROM prediction_outcomes GROUP BY outcome
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	s := &AccuracyStats{}
	for rows.Next() {
		var outcome string
		var count int
		if err := rows.Scan(&outcome, &count); err != nil {
			continue
		}
		s.Total += count
		switch outcome {
		case "confirmed":
			s.Confirmed = count
		case "refuted":
			s.Refuted = count
		case "partial":
			s.Partial = count
		case "pending":
			s.Pending = count
		}
	}
	validated := s.Confirmed + s.Refuted + s.Partial
	if validated > 0 {
		s.Score = (float64(s.Confirmed) + 0.5*float64(s.Partial)) / float64(validated)
	}
	return s, nil
}

// ─── Scheduled Simulations ────────────────────────────────────────────────────

// ScheduledSim represents a recurring simulation.
type ScheduledSim struct {
	ID         string `json:"id"`
	Question   string `json:"question"`
	Department string `json:"department"`
	Rounds     int    `json:"rounds"`
	Context    string `json:"context,omitempty"`
	IntervalH  int    `json:"interval_h"` // hours between runs (e.g. 168 = weekly)
	Enabled    bool   `json:"enabled"`
	LastRunAt  *int64 `json:"last_run_at,omitempty"`
	NextRunAt  int64  `json:"next_run_at"`
	CreatedAt  int64  `json:"created_at"`
}

// CreateScheduledSim inserts a new scheduled simulation.
func (d *DB) CreateScheduledSim(s ScheduledSim) error {
	enabled := 1
	if !s.Enabled {
		enabled = 0
	}
	_, err := d.Exec(`
		INSERT INTO scheduled_simulations
			(id, question, department, rounds, context, interval_h, enabled, next_run_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, unixepoch())
	`, s.ID, s.Question, s.Department, s.Rounds, s.Context, s.IntervalH, enabled, s.NextRunAt)
	return err
}

// ListScheduledSims returns all scheduled simulations.
func (d *DB) ListScheduledSims() ([]ScheduledSim, error) {
	rows, err := d.Query(`
		SELECT id, question, department, rounds, context, interval_h,
		       enabled, last_run_at, next_run_at, created_at
		FROM scheduled_simulations ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []ScheduledSim
	for rows.Next() {
		var s ScheduledSim
		var enabled int
		if err := rows.Scan(&s.ID, &s.Question, &s.Department, &s.Rounds, &s.Context,
			&s.IntervalH, &enabled, &s.LastRunAt, &s.NextRunAt, &s.CreatedAt); err != nil {
			continue
		}
		s.Enabled = enabled == 1
		result = append(result, s)
	}
	return result, rows.Err()
}

// GetDueScheduledSims returns all enabled scheduled simulations whose next_run_at is past.
func (d *DB) GetDueScheduledSims() ([]ScheduledSim, error) {
	rows, err := d.Query(`
		SELECT id, question, department, rounds, context, interval_h,
		       enabled, last_run_at, next_run_at, created_at
		FROM scheduled_simulations
		WHERE enabled = 1 AND next_run_at <= ?
	`, time.Now().Unix())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []ScheduledSim
	for rows.Next() {
		var s ScheduledSim
		var enabled int
		if err := rows.Scan(&s.ID, &s.Question, &s.Department, &s.Rounds, &s.Context,
			&s.IntervalH, &enabled, &s.LastRunAt, &s.NextRunAt, &s.CreatedAt); err != nil {
			continue
		}
		s.Enabled = enabled == 1
		result = append(result, s)
	}
	return result, rows.Err()
}

// MarkScheduledSimRan updates last_run_at and schedules the next run.
func (d *DB) MarkScheduledSimRan(id string, intervalH int) error {
	next := time.Now().Add(time.Duration(intervalH) * time.Hour).Unix()
	_, err := d.Exec(`
		UPDATE scheduled_simulations SET last_run_at = unixepoch(), next_run_at = ?, updated_at = unixepoch()
		WHERE id = ?
	`, next, id)
	return err
}

// DeleteScheduledSim removes a scheduled simulation.
func (d *DB) DeleteScheduledSim(id string) error {
	_, err := d.Exec(`DELETE FROM scheduled_simulations WHERE id = ?`, id)
	return err
}

// UpdateScheduledSim enables or disables a scheduled simulation.
func (d *DB) UpdateScheduledSim(id string, enabled bool) error {
	v := 0
	if enabled {
		v = 1
	}
	_, err := d.Exec(`UPDATE scheduled_simulations SET enabled = ? WHERE id = ?`, v, id)
	return err
}

// ─── API Keys ─────────────────────────────────────────────────────────────────

// APIKey represents a public API key for external integrations.
type APIKey struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	KeyPrefix  string  `json:"key_prefix"`  // e.g. "frc_abc12..."
	SimsUsed   int     `json:"sims_used"`
	SimsLimit  int     `json:"sims_limit"`  // 0 = unlimited
	Enabled    bool    `json:"enabled"`
	CreatedAt  int64   `json:"created_at"`
	LastUsedAt *int64  `json:"last_used_at,omitempty"`
	// RawKey is only set on creation, never stored
	RawKey     string  `json:"raw_key,omitempty"`
}

// CreateAPIKey generates a new API key, stores the hash, and returns the raw key once.
func (d *DB) CreateAPIKey(id, name string, simsLimit int) (*APIKey, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	rawKey := "frc_" + hex.EncodeToString(b)
	hash := sha256.Sum256([]byte(rawKey))
	keyHash := hex.EncodeToString(hash[:])
	prefix := rawKey[:12] + "..."

	_, err := d.Exec(`
		INSERT INTO api_keys (id, name, key_hash, key_prefix, sims_limit, enabled, created_at)
		VALUES (?, ?, ?, ?, ?, 1, unixepoch())
	`, id, name, keyHash, prefix, simsLimit)
	if err != nil {
		return nil, err
	}
	return &APIKey{
		ID:        id,
		Name:      name,
		KeyPrefix: prefix,
		SimsLimit: simsLimit,
		Enabled:   true,
		RawKey:    rawKey,
	}, nil
}

// ValidateAPIKey checks if a raw key is valid and not over limit. Returns the key record.
func (d *DB) ValidateAPIKey(rawKey string) (*APIKey, error) {
	hash := sha256.Sum256([]byte(rawKey))
	keyHash := hex.EncodeToString(hash[:])

	var k APIKey
	var enabled int
	err := d.QueryRow(`
		SELECT id, name, key_prefix, sims_used, sims_limit, enabled, created_at, last_used_at
		FROM api_keys WHERE key_hash = ?
	`, keyHash).Scan(&k.ID, &k.Name, &k.KeyPrefix, &k.SimsUsed, &k.SimsLimit, &enabled, &k.CreatedAt, &k.LastUsedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("invalid API key")
	}
	if err != nil {
		return nil, err
	}
	k.Enabled = enabled == 1
	if !k.Enabled {
		return nil, fmt.Errorf("API key disabled")
	}
	if k.SimsLimit > 0 && k.SimsUsed >= k.SimsLimit {
		return nil, fmt.Errorf("API key limit reached (%d/%d)", k.SimsUsed, k.SimsLimit)
	}
	return &k, nil
}

// IncrementAPIKeyUsage increments sims_used and updates last_used_at.
func (d *DB) IncrementAPIKeyUsage(keyID string) error {
	_, err := d.Exec(`
		UPDATE api_keys SET sims_used = sims_used + 1, last_used_at = unixepoch() WHERE id = ?
	`, keyID)
	return err
}

// ListAPIKeys returns all API keys (without hashes).
func (d *DB) ListAPIKeys() ([]APIKey, error) {
	rows, err := d.Query(`
		SELECT id, name, key_prefix, sims_used, sims_limit, enabled, created_at, last_used_at
		FROM api_keys ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []APIKey
	for rows.Next() {
		var k APIKey
		var enabled int
		if err := rows.Scan(&k.ID, &k.Name, &k.KeyPrefix, &k.SimsUsed, &k.SimsLimit,
			&enabled, &k.CreatedAt, &k.LastUsedAt); err != nil {
			continue
		}
		k.Enabled = enabled == 1
		result = append(result, k)
	}
	return result, rows.Err()
}

// DeleteAPIKey removes an API key.
func (d *DB) DeleteAPIKey(id string) error {
	_, err := d.Exec(`DELETE FROM api_keys WHERE id = ?`, id)
	return err
}
