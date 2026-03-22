package memory

import (
	"database/sql"
	"encoding/json"
	"math"
)

// FeedbackRecord is a real-world outcome recorded by the user.
type FeedbackRecord struct {
	SimulationID string  `json:"simulation_id"`
	Predicted    string  `json:"predicted"`
	Actual       string  `json:"actual"`
	DeltaScore   float64 `json:"delta_score"` // -1.0 (completely wrong) to 1.0 (exactly right)
	Notes        string  `json:"notes"`
}

// ArchetypeCalibration holds the calibration state for a single archetype.
type ArchetypeCalibration struct {
	ArchetypeID    string  `json:"archetype_id"`
	MemoryWeight   float64 `json:"memory_weight"`  // multiplier: 0.5 (less trusted) to 2.0 (highly trusted)
	FeedbackCount  int     `json:"feedback_count"`
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

// RecordFeedback stores a feedback record and updates archetype calibration.
func (c *Calibrator) RecordFeedback(companyID string, feedback FeedbackRecord) error {
	// Save feedback record
	_, err := c.db.Exec(`
		INSERT INTO feedback (id, simulation_id, company_id, predicted, actual, delta_score, notes, recorded_at)
		VALUES (lower(hex(randomblob(16))), ?, ?, ?, ?, ?, ?, unixepoch())
	`, feedback.SimulationID, companyID, feedback.Predicted, feedback.Actual, feedback.DeltaScore, feedback.Notes)
	if err != nil {
		return err
	}

	// Recalibrate archetypes that participated in this simulation
	return c.recalibrateForSimulation(feedback.SimulationID, feedback.DeltaScore)
}

// recalibrateForSimulation adjusts the memory_weight of archetypes that participated
// in the simulation based on the real-world feedback delta.
func (c *Calibrator) recalibrateForSimulation(simulationID string, deltaScore float64) error {
	// Get distinct agents that participated in this simulation
	rows, err := c.db.Query(`
		SELECT DISTINCT agent_id, agent_type FROM simulation_rounds WHERE simulation_id = ?
	`, simulationID)
	if err != nil {
		return err
	}
	defer rows.Close()

	var agents []struct {
		id        string
		agentType string
	}
	for rows.Next() {
		var id, agentType string
		if err := rows.Scan(&id, &agentType); err != nil {
			continue
		}
		agents = append(agents, struct {
			id        string
			agentType string
		}{id, agentType})
	}

	// For each agent, update their memory_weight in the archetypes table
	for _, agent := range agents {
		// Get current calibration
		var currentWeight float64
		var feedbackCount int
		err := c.db.QueryRow(`
			SELECT memory_weight FROM archetypes WHERE id = ?
		`, agent.id).Scan(&currentWeight)
		if err != nil {
			continue // archetype not in DB (built-in), skip
		}

		// Get feedback count for this archetype
		c.db.QueryRow(`
			SELECT COUNT(*) FROM feedback f
			JOIN simulation_rounds sr ON f.simulation_id = sr.simulation_id
			WHERE sr.agent_id = ?
		`, agent.id).Scan(&feedbackCount)

		// Exponential moving average: new_weight = old_weight * 0.9 + delta_adjustment * 0.1
		// delta_score: -1.0 = completely wrong, 1.0 = perfect
		// adjustment: positive delta → increase weight, negative → decrease
		adjustment := (deltaScore + 1.0) / 2.0 // normalize to 0.0-1.0
		newWeight := currentWeight*0.9 + adjustment*0.1

		// Clamp between 0.3 and 2.0
		newWeight = math.Max(0.3, math.Min(2.0, newWeight))

		c.db.Exec(`
			UPDATE archetypes SET memory_weight = ?, updated_at = unixepoch() WHERE id = ?
		`, newWeight, agent.id)
	}

	return nil
}

// GetCalibrationReport returns the calibration state for all archetypes of a company.
func (c *Calibrator) GetCalibrationReport(companyID string) ([]ArchetypeCalibration, error) {
	rows, err := c.db.Query(`
		SELECT a.id, a.memory_weight,
			COUNT(DISTINCT f.id) as feedback_count,
			COALESCE(AVG(f.delta_score), 0) as avg_accuracy
		FROM archetypes a
		LEFT JOIN simulation_rounds sr ON a.id = sr.agent_id
		LEFT JOIN feedback f ON sr.simulation_id = f.simulation_id
		WHERE a.company_id = ? OR a.company_id IS NULL
		GROUP BY a.id
	`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var calibrations []ArchetypeCalibration
	for rows.Next() {
		var cal ArchetypeCalibration
		if err := rows.Scan(&cal.ArchetypeID, &cal.MemoryWeight, &cal.FeedbackCount, &cal.AverageAccuracy); err != nil {
			continue
		}
		calibrations = append(calibrations, cal)
	}
	return calibrations, nil
}

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
	// Upsert decision node
	decisionID := hashString(companyID + "|decision|" + decisionDesc)
	cg.db.Exec(`
		INSERT OR IGNORE INTO causality_nodes (id, company_id, description, node_type, created_at)
		VALUES (?, ?, ?, 'decision', unixepoch())
	`, decisionID, companyID, decisionDesc)

	// Upsert outcome node
	outcomeID := hashString(companyID + "|outcome|" + outcomeDesc)
	cg.db.Exec(`
		INSERT OR IGNORE INTO causality_nodes (id, company_id, description, node_type, created_at)
		VALUES (?, ?, ?, 'outcome', unixepoch())
	`, outcomeID, companyID, outcomeDesc)

	// Upsert edge (increment evidence count)
	_, err := cg.db.Exec(`
		INSERT INTO causality_edges (from_node, to_node, strength, evidence)
		VALUES (?, ?, 0.5, 1)
		ON CONFLICT(from_node, to_node) DO UPDATE SET
			evidence = evidence + 1,
			strength = MIN(1.0, strength + 0.05)
	`, decisionID, outcomeID)
	return err
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

// CausalPath represents a learned decision → outcome relationship.
type CausalPath struct {
	Decision string  `json:"decision"`
	Outcome  string  `json:"outcome"`
	Strength float64 `json:"strength"` // 0.0-1.0
	Evidence int     `json:"evidence"` // number of times observed
}

// hashString produces a deterministic short ID from a string.
func hashString(s string) string {
	h := 0
	for _, c := range s {
		h = h*31 + int(c)
	}
	b, _ := json.Marshal(h)
	return string(b)
}
