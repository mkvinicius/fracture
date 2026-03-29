package memory

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
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
// Tries embedding-based search first; falls back to keyword overlap.
func (s *Store) SimilarContexts(query string, n int) []string {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// tenta busca semântica com embeddings
	queryEmb, err := embedText(ctx, query)
	if err == nil && queryEmb != nil {
		results := s.similarByEmbedding(queryEmb, n)
		if len(results) > 0 {
			return results
		}
	}
	// fallback: keyword overlap original
	return s.similarByKeyword(query, n)
}

func (s *Store) similarByEmbedding(queryEmb []float32, n int) []string {
	rows, err := s.db.Query(`
		SELECT content, embedding FROM agent_memory
		WHERE embedding IS NOT NULL
		ORDER BY created_at DESC
		LIMIT 500
	`)
	if err != nil {
		return nil
	}
	defer rows.Close()

	type scored struct {
		text  string
		score float64
	}
	var results []scored

	for rows.Next() {
		var content string
		var blob []byte
		if err := rows.Scan(&content, &blob); err != nil {
			continue
		}
		emb := blobToEmbedding(blob)
		score := cosineSimF32(queryEmb, emb)
		if score > 0.5 { // só retorna contextos realmente similares
			results = append(results, scored{content, score})
		}
	}

	// insertion sort por score descendente
	for i := 1; i < len(results); i++ {
		for j := i; j > 0 && results[j].score > results[j-1].score; j-- {
			results[j], results[j-1] = results[j-1], results[j]
		}
	}

	out := make([]string, 0, n)
	for i := 0; i < n && i < len(results); i++ {
		out = append(out, results[i].text)
	}
	return out
}

func (s *Store) similarByKeyword(query string, n int) []string {
	rows, err := s.db.Query(`
		SELECT action_text FROM simulation_rounds
		ORDER BY created_at DESC
		LIMIT ?
	`, n*5)
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

	queryWords := tokenize(query)
	type scored struct {
		text  string
		score int
	}
	var results []scored
	for _, c := range candidates {
		cWords := tokenize(c)
		sc := overlap(queryWords, cWords)
		if sc > 0 {
			results = append(results, scored{c, sc})
		}
	}

	// insertion sort por score descendente
	for i := 1; i < len(results); i++ {
		for j := i; j > 0 && results[j].score > results[j-1].score; j-- {
			results[j], results[j-1] = results[j-1], results[j]
		}
	}

	out := make([]string, 0, n)
	for i := 0; i < n && i < len(results); i++ {
		out = append(out, results[i].text)
	}
	return out
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

	// grava em agent_memory com embedding (fallback silencioso se falhar)
	db := s.db
	agentID, content := action.AgentID, action.Text
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		emb, _ := embedText(ctx, content)
		blob := embeddingToBlob(emb)
		db.ExecContext(ctx,
			`INSERT INTO agent_memory (agent_id, content, embedding) VALUES (?, ?, ?)`,
			agentID, content, blob,
		)
	}()

	return err
}

// ─── Simulation history ───────────────────────────────────────────────────────

// SimulationSummary is a lightweight summary of a past simulation for a company.
type SimulationSummary struct {
	SimulationID   string
	Question       string
	FracturePoints []string
	Confidence     float64
	ConvergedAt    int
	CreatedAt      time.Time
}

// GetSimulationHistory returns the N most recent simulation summaries for a company.
// FracturePoints are extracted from the stored result_json.
func (s *Store) GetSimulationHistory(companyID string, n int) ([]SimulationSummary, error) {
	rows, err := s.db.Query(`
		SELECT sim.id, sim.question, sim.rounds, COALESCE(sim.result_json, '{}'), sim.created_at
		FROM simulations sim
		INNER JOIN simulation_jobs sj ON sj.id = sim.id
		WHERE sj.company = ?
		ORDER BY sim.created_at DESC
		LIMIT ?
	`, companyID, n)
	if err != nil {
		return nil, fmt.Errorf("get simulation history: %w", err)
	}
	defer rows.Close()

	var summaries []SimulationSummary
	for rows.Next() {
		var (
			id         string
			question   string
			rounds     int
			resultJSON string
			createdAt  int64
		)
		if err := rows.Scan(&id, &question, &rounds, &resultJSON, &createdAt); err != nil {
			continue
		}

		fps, conf := extractFracturePoints(resultJSON)
		summaries = append(summaries, SimulationSummary{
			SimulationID:   id,
			Question:       question,
			FracturePoints: fps,
			Confidence:     conf,
			ConvergedAt:    rounds,
			CreatedAt:      time.Unix(createdAt, 0),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("get simulation history scan: %w", err)
	}
	return summaries, nil
}

// extractFracturePoints parses result_json to collect fracture descriptions and
// returns a (fracture_points []string, avg_confidence float64) pair.
func extractFracturePoints(resultJSON string) ([]string, float64) {
	// Minimal struct — only fields we need
	var result struct {
		FractureEvents []struct {
			ProposedBy string `json:"proposed_by"`
			Accepted   bool   `json:"accepted"`
			Confidence float64 `json:"confidence"`
			Proposal   struct {
				NewDescription string `json:"new_description"`
				OriginalRuleID string `json:"original_rule_id"`
			} `json:"proposal"`
		} `json:"fracture_events"`
		RuptureScenarios []struct {
			RuleDescription string  `json:"rule_description"`
			Probability     float64 `json:"probability"`
		} `json:"rupture_scenarios"`
	}

	if err := json.Unmarshal([]byte(resultJSON), &result); err != nil {
		return nil, 0
	}

	seen := make(map[string]bool)
	var fps []string
	var totalConf float64
	var count int

	for _, fe := range result.FractureEvents {
		desc := fe.Proposal.NewDescription
		if desc == "" {
			desc = fe.Proposal.OriginalRuleID
		}
		if desc != "" && !seen[desc] {
			seen[desc] = true
			fps = append(fps, desc)
			totalConf += fe.Confidence
			count++
		}
	}

	// Fallback: use rupture scenarios if no fracture events found
	if len(fps) == 0 {
		for _, rs := range result.RuptureScenarios {
			if rs.RuleDescription != "" && !seen[rs.RuleDescription] {
				seen[rs.RuleDescription] = true
				fps = append(fps, rs.RuleDescription)
				totalConf += rs.Probability
				count++
			}
		}
	}

	var avgConf float64
	if count > 0 {
		avgConf = totalConf / float64(count)
	}
	return fps, avgConf
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
