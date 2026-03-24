package memory

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/fracture/fracture/engine"
	"github.com/google/uuid"
)

// Store implements engine.AgentMemory backed by SQLite.
type Store struct {
	db *sql.DB
}

// NewStore creates a memory Store using the given SQLite DB connection.
func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

// RecentActions returns the N most recent actions for a given agent.
func (s *Store) RecentActions(agentID string, n int) []engine.AgentAction {
	rows, err := s.db.Query(`
		SELECT action_text, agent_type, fracture_proposed, tokens_used
		FROM simulation_rounds
		WHERE agent_id = ?
		ORDER BY created_at DESC
		LIMIT ?
	`, agentID, n)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var actions []engine.AgentAction
	for rows.Next() {
		var (
			text             string
			agentType        string
			fractureProposed int
			tokens           int
		)
		if err := rows.Scan(&text, &agentType, &fractureProposed, &tokens); err != nil {
			continue
		}
		actions = append(actions, engine.AgentAction{
			AgentID:            agentID,
			AgentType:          engine.AgentType(agentType),
			Text:               text,
			IsFractureProposal: fractureProposed == 1,
			TokensUsed:         tokens,
		})
	}
	if rows.Err() != nil {
		return nil
	}
	return actions
}

// SimilarContexts returns N stored action texts most semantically similar to query.
// In this implementation we use a simple keyword overlap (no vector DB required).
// For production, replace with FAISS or similar.
func (s *Store) SimilarContexts(query string, n int) []string {
	rows, err := s.db.Query(`
		SELECT action_text FROM simulation_rounds
		ORDER BY created_at DESC
		LIMIT ?
	`, n*5) // over-fetch then filter
	if err != nil {
		return nil
	}
	defer rows.Close()

	var candidates []string
	for rows.Next() {
		var text string
		if err := rows.Scan(&text); err != nil {
			continue
		}
		candidates = append(candidates, text)
	}
	if rows.Err() != nil {
		return nil
	}

	// Simple keyword overlap scoring
	queryWords := tokenize(query)
	type scored struct {
		text  string
		score int
	}
	var scored_ []scored
	for _, c := range candidates {
		cWords := tokenize(c)
		score := overlap(queryWords, cWords)
		if score > 0 {
			scored_ = append(scored_, scored{c, score})
		}
	}

	// Sort by score descending
	for i := 0; i < len(scored_); i++ {
		for j := i + 1; j < len(scored_); j++ {
			if scored_[j].score > scored_[i].score {
				scored_[i], scored_[j] = scored_[j], scored_[i]
			}
		}
	}

	result := make([]string, 0, n)
	for i, s := range scored_ {
		if i >= n {
			break
		}
		result = append(result, s.text)
	}
	return result
}

// SaveRound persists a simulation round to the database.
func (s *Store) SaveRound(simulationID string, round int, action engine.AgentAction, tension float64) error {
	var newRuleJSON *string
	if action.Proposal != nil {
		b, _ := json.Marshal(action.Proposal)
		str := string(b)
		newRuleJSON = &str
	}

	_, err := s.db.Exec(`
		INSERT INTO simulation_rounds
			(id, simulation_id, round_number, agent_id, agent_type, action_text,
			 tension_level, fracture_proposed, new_rule_json, tokens_used, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		uuid.New().String(),
		simulationID,
		round,
		action.AgentID,
		string(action.AgentType),
		action.Text,
		tension,
		boolToInt(action.IsFractureProposal),
		newRuleJSON,
		action.TokensUsed,
		time.Now().Unix(),
	)
	return err
}

// SaveFeedback records a real-world outcome for calibration.
func (s *Store) SaveFeedback(simulationID, companyID, predicted, actual string, delta float64, notes string) error {
	_, err := s.db.Exec(`
		INSERT INTO feedback (id, simulation_id, company_id, predicted, actual, delta_score, notes, recorded_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`,
		uuid.New().String(),
		simulationID,
		companyID,
		predicted,
		actual,
		delta,
		notes,
		time.Now().Unix(),
	)
	return err
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func tokenize(s string) map[string]struct{} {
	words := make(map[string]struct{})
	word := ""
	for _, r := range s {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' {
			word += string(r)
		} else if word != "" {
			words[word] = struct{}{}
			word = ""
		}
	}
	if word != "" {
		words[word] = struct{}{}
	}
	return words
}

func overlap(a, b map[string]struct{}) int {
	count := 0
	for w := range a {
		if _, ok := b[w]; ok {
			count++
		}
	}
	return count
}
