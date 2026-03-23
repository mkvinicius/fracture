package db

import (
	"database/sql"
	"encoding/json"
	"time"
)

// ─── Simulation Rounds ────────────────────────────────────────────────────────

// RoundRow mirrors a row in simulation_rounds.
type RoundRow struct {
	ID               string  `json:"id"`
	SimulationID     string  `json:"simulation_id"`
	RoundNumber      int     `json:"round_number"`
	AgentID          string  `json:"agent_id"`
	AgentType        string  `json:"agent_type"`
	ActionText       string  `json:"action_text"`
	TensionLevel     float64 `json:"tension_level"`
	FractureProposed bool    `json:"fracture_proposed"`
	FractureAccepted *bool   `json:"fracture_accepted,omitempty"`
	NewRuleJSON      string  `json:"new_rule_json,omitempty"`
	TokensUsed       int     `json:"tokens_used"`
	CreatedAt        int64   `json:"created_at"`
}

// SaveRound persists a single agent action from a simulation round.
func (d *DB) SaveRound(r *RoundRow) error {
	var newRuleJSON sql.NullString
	if r.NewRuleJSON != "" {
		newRuleJSON = sql.NullString{String: r.NewRuleJSON, Valid: true}
	}
	var fractureAccepted sql.NullInt64
	if r.FractureAccepted != nil {
		v := int64(0)
		if *r.FractureAccepted {
			v = 1
		}
		fractureAccepted = sql.NullInt64{Int64: v, Valid: true}
	}
	_, err := d.Exec(`
		INSERT INTO simulation_rounds
			(id, simulation_id, round_number, agent_id, agent_type, action_text,
			 tension_level, fracture_proposed, fracture_accepted, new_rule_json,
			 tokens_used, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			fracture_accepted = excluded.fracture_accepted
	`,
		r.ID, r.SimulationID, r.RoundNumber, r.AgentID, r.AgentType, r.ActionText,
		r.TensionLevel, boolToInt(r.FractureProposed), fractureAccepted, newRuleJSON,
		r.TokensUsed, r.CreatedAt,
	)
	return err
}

// ListRounds returns all rounds for a simulation ordered by round number.
func (d *DB) ListRounds(simulationID string) ([]RoundRow, error) {
	rows, err := d.Query(`
		SELECT id, simulation_id, round_number, agent_id, agent_type, action_text,
		       tension_level, fracture_proposed, fracture_accepted,
		       COALESCE(new_rule_json, ''), tokens_used, created_at
		FROM simulation_rounds
		WHERE simulation_id = ?
		ORDER BY round_number ASC, created_at ASC
	`, simulationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []RoundRow
	for rows.Next() {
		var r RoundRow
		var fractureProposed int
		var fractureAccepted sql.NullInt64
		if err := rows.Scan(
			&r.ID, &r.SimulationID, &r.RoundNumber, &r.AgentID, &r.AgentType, &r.ActionText,
			&r.TensionLevel, &fractureProposed, &fractureAccepted, &r.NewRuleJSON,
			&r.TokensUsed, &r.CreatedAt,
		); err != nil {
			continue
		}
		r.FractureProposed = fractureProposed == 1
		if fractureAccepted.Valid {
			v := fractureAccepted.Int64 == 1
			r.FractureAccepted = &v
		}
		result = append(result, r)
	}
	return result, nil
}

// ─── Fracture Votes ───────────────────────────────────────────────────────────

// VoteRow mirrors a row in fracture_votes.
type VoteRow struct {
	ID           string  `json:"id"`
	SimulationID string  `json:"simulation_id"`
	RoundNumber  int     `json:"round_number"`
	ProposalID   string  `json:"proposal_id"`
	VoterID      string  `json:"voter_id"`
	VoterType    string  `json:"voter_type"`
	Vote         bool    `json:"vote"`
	Weight       float64 `json:"weight"`
	Reasoning    string  `json:"reasoning,omitempty"`
	CreatedAt    int64   `json:"created_at"`
}

// SaveVote persists a single agent vote on a fracture proposal.
func (d *DB) SaveVote(v *VoteRow) error {
	_, err := d.Exec(`
		INSERT OR IGNORE INTO fracture_votes
			(id, simulation_id, round_number, proposal_id, voter_id, voter_type,
			 vote, weight, reasoning, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		v.ID, v.SimulationID, v.RoundNumber, v.ProposalID, v.VoterID, v.VoterType,
		boolToInt(v.Vote), v.Weight, v.Reasoning, v.CreatedAt,
	)
	return err
}

// ListVotes returns all votes for a simulation.
func (d *DB) ListVotes(simulationID string) ([]VoteRow, error) {
	rows, err := d.Query(`
		SELECT id, simulation_id, round_number, proposal_id, voter_id, voter_type,
		       vote, weight, COALESCE(reasoning, ''), created_at
		FROM fracture_votes WHERE simulation_id = ?
		ORDER BY round_number ASC, created_at ASC
	`, simulationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []VoteRow
	for rows.Next() {
		var v VoteRow
		var vote int
		if err := rows.Scan(
			&v.ID, &v.SimulationID, &v.RoundNumber, &v.ProposalID, &v.VoterID, &v.VoterType,
			&vote, &v.Weight, &v.Reasoning, &v.CreatedAt,
		); err != nil {
			continue
		}
		v.Vote = vote == 1
		result = append(result, v)
	}
	return result, nil
}

// ─── Report Generations ───────────────────────────────────────────────────────

// ReportGenRow mirrors a row in report_generations.
type ReportGenRow struct {
	ID           string `json:"id"`
	SimulationID string `json:"simulation_id"`
	ReportType   string `json:"report_type"`
	Status       string `json:"status"`
	TokensUsed   int    `json:"tokens_used"`
	DurationMs   int64  `json:"duration_ms"`
	ErrorMsg     string `json:"error_msg,omitempty"`
	CreatedAt    int64  `json:"created_at"`
	CompletedAt  *int64 `json:"completed_at,omitempty"`
}

// StartReportGen records the start of a report generation attempt.
func (d *DB) StartReportGen(id, simulationID, reportType string) error {
	_, err := d.Exec(`
		INSERT INTO report_generations (id, simulation_id, report_type, status, created_at)
		VALUES (?, ?, ?, 'started', ?)
	`, id, simulationID, reportType, time.Now().Unix())
	return err
}

// CompleteReportGen marks a report generation as done or error.
func (d *DB) CompleteReportGen(id, status, errorMsg string, tokensUsed int, durationMs int64) error {
	_, err := d.Exec(`
		UPDATE report_generations
		SET status = ?, error_msg = ?, tokens_used = ?, duration_ms = ?, completed_at = ?
		WHERE id = ?
	`, status, errorMsg, tokensUsed, durationMs, time.Now().Unix(), id)
	return err
}

// ListReportGens returns all report generation records for a simulation.
func (d *DB) ListReportGens(simulationID string) ([]ReportGenRow, error) {
	rows, err := d.Query(`
		SELECT id, simulation_id, report_type, status, tokens_used, duration_ms,
		       COALESCE(error_msg, ''), created_at, completed_at
		FROM report_generations WHERE simulation_id = ?
		ORDER BY created_at ASC
	`, simulationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []ReportGenRow
	for rows.Next() {
		var r ReportGenRow
		var completedAt sql.NullInt64
		if err := rows.Scan(
			&r.ID, &r.SimulationID, &r.ReportType, &r.Status, &r.TokensUsed, &r.DurationMs,
			&r.ErrorMsg, &r.CreatedAt, &completedAt,
		); err != nil {
			continue
		}
		if completedAt.Valid {
			r.CompletedAt = &completedAt.Int64
		}
		result = append(result, r)
	}
	return result, nil
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// MarshalJSON is a convenience wrapper for encoding structs to JSON strings.
func marshalJSON(v interface{}) string {
	if v == nil {
		return ""
	}
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}
