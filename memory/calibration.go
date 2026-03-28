package memory

import (
	"database/sql"
	"fmt"
	"math"
)

// FeedbackRecord is a real-world outcome recorded by the user.
type FeedbackRecord struct {
	SimulationID     string  `json:"simulation_id"`
	PredictedFracture string `json:"predicted_fracture"`
	ActualOutcome    string  `json:"actual_outcome"`
	DeltaScore       float64 `json:"delta_score"` // -1.0 (completely wrong) to 1.0 (exactly right)
	Notes            string  `json:"notes"`
}

// ArchetypeCalibration holds the calibration state for a single archetype.
type ArchetypeCalibration struct {
	ArchetypeID     string  `json:"archetype_id"`
	AccuracyWeight  float64 `json:"accuracy_weight"`  // multiplier: 0.3 (less trusted) to 2.0 (highly trusted)
	FeedbackCount   int     `json:"feedback_count"`
	AverageAccuracy float64 `json:"average_accuracy"` // rolling average of delta scores
}

// Calibrator adjusts archetype weights based on real-world feedback.
type Calibrator struct {
	db *sql.DB
}

// NewCalibrator creates a Calibrator backed by the given SQLite DB.
func NewCalibrator(db *sql.DB) *Calibrator {
	return &Calibrator{db: db}
}

// RecordFeedback updates archetype calibration based on real-world outcome.
// Feedback persistence (INSERT into feedback table) is handled by the caller (handler).
// This method only updates accuracy_weight in archetype_calibration.
func (c *Calibrator) RecordFeedback(companyID string, feedback FeedbackRecord) error {
	return c.recalibrateForSimulation(feedback.SimulationID, feedback.DeltaScore)
}

// recalibrateForSimulation adjusts the accuracy_weight of archetypes that participated
// in the simulation, using an exponential moving average weighted by sample_count.
func (c *Calibrator) recalibrateForSimulation(simulationID string, deltaScore float64) error {
	// Get the domain (department) for this simulation from simulation_jobs
	var domain string
	if err := c.db.QueryRow(
		`SELECT department FROM simulation_jobs WHERE id = ?`, simulationID,
	).Scan(&domain); err != nil {
		// Fall back to a generic domain if no job row found
		domain = "market"
	}

	// Get distinct agents that participated in this simulation
	rows, err := c.db.Query(
		`SELECT DISTINCT agent_id FROM simulation_rounds WHERE simulation_id = ?`, simulationID,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	var agentIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			continue
		}
		agentIDs = append(agentIDs, id)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	// Normalize delta_score from [-1, 1] to [0, 1] for EMA computation
	// -1.0 → 0.0 (wrong), 0.0 → 0.5 (neutral), 1.0 → 1.0 (perfect)
	adjustment := (deltaScore + 1.0) / 2.0

	for _, agentID := range agentIDs {
		var currentWeight float64
		var sampleCount int

		err := c.db.QueryRow(`
			SELECT accuracy_weight, sample_count
			FROM archetype_calibration
			WHERE archetype_id = ? AND domain = ?
		`, agentID, domain).Scan(&currentWeight, &sampleCount)

		if err == sql.ErrNoRows {
			// First calibration for this agent+domain — start at neutral 1.0
			currentWeight = 1.0
			sampleCount = 0
		} else if err != nil {
			continue // skip on unexpected error
		}

		// Exponential moving average with decaying alpha:
		// alpha = 1 / (sample_count + 1) — stabilises as evidence accumulates
		alpha := 1.0 / (float64(sampleCount) + 1.0)
		newWeight := currentWeight*(1.0-alpha) + adjustment*alpha

		// Re-scale: neutral 0.5 → 1.0, perfect 1.0 → 2.0, worst 0.0 → 0.3
		calibrated := 0.3 + newWeight*1.7
		calibrated = math.Max(0.3, math.Min(2.0, calibrated))

		_, err = c.db.Exec(`
			INSERT INTO archetype_calibration (archetype_id, domain, accuracy_weight, sample_count, updated_at)
			VALUES (?, ?, ?, 1, unixepoch())
			ON CONFLICT(archetype_id, domain) DO UPDATE SET
				accuracy_weight = ?,
				sample_count    = sample_count + 1,
				updated_at      = unixepoch()
		`, agentID, domain, calibrated, calibrated)
		if err != nil {
			// Non-fatal: log is the caller's responsibility
			continue
		}
	}

	return nil
}

// GetCalibrationReport returns the calibration state for archetypes that have
// participated in simulations for the given company.
func (c *Calibrator) GetCalibrationReport(companyID string) ([]ArchetypeCalibration, error) {
	rows, err := c.db.Query(`
		SELECT
			ac.archetype_id,
			AVG(ac.accuracy_weight)  AS avg_weight,
			SUM(ac.sample_count)     AS total_samples,
			AVG(ac.accuracy_weight)  AS avg_accuracy
		FROM archetype_calibration ac
		INNER JOIN simulation_rounds sr ON sr.agent_id   = ac.archetype_id
		INNER JOIN simulation_jobs   sj ON sj.id         = sr.simulation_id
		WHERE sj.company = ?
		GROUP BY ac.archetype_id
	`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var calibrations []ArchetypeCalibration
	for rows.Next() {
		var cal ArchetypeCalibration
		if err := rows.Scan(
			&cal.ArchetypeID,
			&cal.AccuracyWeight,
			&cal.FeedbackCount,
			&cal.AverageAccuracy,
		); err != nil {
			continue
		}
		calibrations = append(calibrations, cal)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("get calibration report: %w", err)
	}
	return calibrations, nil
}

// ─── Causality graph ──────────────────────────────────────────────────────────

// CausalityGraph records a causal relationship between a decision and an outcome.
type CausalityGraph struct {
	db *sql.DB
}

// NewCausalityGraph creates a CausalityGraph.
func NewCausalityGraph(db *sql.DB) *CausalityGraph {
	return &CausalityGraph{db: db}
}

// RecordCausality records that a decision led to an outcome.
func (cg *CausalityGraph) RecordCausality(companyID, decisionDesc, outcomeDesc string) error {
	decisionID := hashString(companyID + "|decision|" + decisionDesc)
	cg.db.Exec(`
		INSERT OR IGNORE INTO causality_nodes (id, company_id, description, node_type, created_at)
		VALUES (?, ?, ?, 'decision', unixepoch())
	`, decisionID, companyID, decisionDesc)

	outcomeID := hashString(companyID + "|outcome|" + outcomeDesc)
	cg.db.Exec(`
		INSERT OR IGNORE INTO causality_nodes (id, company_id, description, node_type, created_at)
		VALUES (?, ?, ?, 'outcome', unixepoch())
	`, outcomeID, companyID, outcomeDesc)

	_, err := cg.db.Exec(`
		INSERT INTO causality_edges (from_node, to_node, strength, evidence)
		VALUES (?, ?, 0.5, 1)
		ON CONFLICT(from_node, to_node) DO UPDATE SET
			evidence = evidence + 1,
			strength = MIN(1.0, strength + 0.05)
	`, decisionID, outcomeID)
	return err
}

// RecordEdge inserts or increments a cause→effect edge, keyed by namespace.
// Used by DeepSearch causality ingestion; namespace is typically "sector::domain".
func (cg *CausalityGraph) RecordEdge(namespace, cause, effect string) error {
	causeID := hashString(namespace + "|cause|" + cause)
	cg.db.Exec(`
		INSERT OR IGNORE INTO causality_nodes (id, company_id, description, node_type, created_at)
		VALUES (?, ?, ?, 'cause', unixepoch())
	`, causeID, namespace, cause)

	effectID := hashString(namespace + "|effect|" + effect)
	cg.db.Exec(`
		INSERT OR IGNORE INTO causality_nodes (id, company_id, description, node_type, created_at)
		VALUES (?, ?, ?, 'effect', unixepoch())
	`, effectID, namespace, effect)

	_, err := cg.db.Exec(`
		INSERT INTO causality_edges (from_node, to_node, strength, evidence)
		VALUES (?, ?, 0.5, 1)
		ON CONFLICT(from_node, to_node) DO UPDATE SET
			evidence = evidence + 1,
			strength = MIN(1.0, strength + 0.05)
	`, causeID, effectID)
	return err
}

// CausalPath represents a learned decision → outcome relationship.
type CausalPath struct {
	Decision string  `json:"decision"`
	Outcome  string  `json:"outcome"`
	Strength float64 `json:"strength"` // 0.0-1.0
	Evidence int     `json:"evidence"` // number of times observed
}

// GetCausalChain returns the most evidenced causal paths from a given decision.
func (cg *CausalityGraph) GetCausalChain(companyID, decisionDesc string, depth int) ([]CausalPath, error) {
	decisionID := hashString(companyID + "|decision|" + decisionDesc)

	rows, err := cg.db.Query(`
		SELECT n2.description, e.strength, e.evidence
		FROM causality_edges e
		JOIN causality_nodes n2 ON e.to_node = n2.id
		WHERE e.from_node = ?
		ORDER BY e.evidence DESC, e.strength DESC
		LIMIT ?
	`, decisionID, depth)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var paths []CausalPath
	for rows.Next() {
		var p CausalPath
		if err := rows.Scan(&p.Outcome, &p.Strength, &p.Evidence); err != nil {
			continue
		}
		p.Decision = decisionDesc
		paths = append(paths, p)
	}
	return paths, nil
}

// ─── helpers ─────────────────────────────────────────────────────────────────

// hashString produces a deterministic short ID from a string.
func hashString(s string) string {
	h := 0
	for _, c := range s {
		h = h*31 + int(c)
	}
	return fmt.Sprintf("%d", h)
}
