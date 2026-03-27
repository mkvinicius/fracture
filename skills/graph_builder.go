package skills

import (
	"database/sql"
	"fmt"
	"log"
)

// Builder reads simulation data from SQLite and populates the knowledge graph.
type Builder struct {
	g  *Graph
	db *sql.DB
}

// NewBuilder returns a Builder that indexes simulation records into g.
func (g *Graph) NewBuilder() *Builder {
	return &Builder{g: g, db: g.db}
}

// FromSimulation reads a completed simulation from the DB and indexes its
// domain, agents, fracture proposals, and votes as graph nodes and edges.
// Errors from individual rows are logged but do not abort the indexing.
func (b *Builder) FromSimulation(simID string) error {
	// ── 1. Domain node (from simulation_jobs.department) ──────────────────
	var department string
	_ = b.db.QueryRow(
		`SELECT department FROM simulation_jobs WHERE id = ?`, simID,
	).Scan(&department)
	if department == "" {
		department = "market"
	}
	domainID := "domain:" + department
	if err := b.g.UpsertNode(Node{
		ID:        domainID,
		Type:      NodeDomain,
		Label:     department,
		Embedding: computeNodeVec(department),
	}); err != nil {
		log.Printf("[GraphRAG] upsert domain node %q: %v", department, err)
	}

	// ── 2. Rounds → agents + fracture nodes ───────────────────────────────
	rows, err := b.db.Query(`
		SELECT round_number, agent_id, agent_type, action_text,
		       tension_level, fracture_proposed, new_rule_json
		FROM simulation_rounds
		WHERE simulation_id = ?
		ORDER BY round_number
	`, simID)
	if err != nil {
		return fmt.Errorf("graph builder: query rounds for %s: %w", simID, err)
	}
	defer rows.Close()

	agentsSeen := map[string]bool{}

	for rows.Next() {
		var (
			round        int
			agentID      string
			agentType    string
			actionText   string
			tension      float64
			fractureProp int
			newRuleJSON  sql.NullString
		)
		if err := rows.Scan(
			&round, &agentID, &agentType, &actionText,
			&tension, &fractureProp, &newRuleJSON,
		); err != nil {
			continue
		}

		// Agent node (upsert once per simulation)
		agentNodeID := "agent:" + agentID
		if !agentsSeen[agentID] {
			agentsSeen[agentID] = true
			_ = b.g.UpsertNode(Node{
				ID:        agentNodeID,
				Type:      NodeAgent,
				Label:     agentID,
				Props:     map[string]string{"type": agentType},
				Embedding: computeNodeVec(agentID + " " + agentType),
			})
			// agent → domain
			_ = b.g.UpsertEdge(Edge{
				From: agentNodeID, To: domainID,
				Type: EdgeBelongsTo, Weight: 0.5,
			})
		}

		// Fracture node (only when the agent proposed a rule change)
		if fractureProp == 1 && newRuleJSON.Valid && newRuleJSON.String != "" {
			fractureID := fmt.Sprintf("fracture:%s:r%d:%s", simID, round, agentID)
			label := truncate(actionText, 80)
			_ = b.g.UpsertNode(Node{
				ID:   fractureID,
				Type: NodeFracture,
				Label: label,
				Props: map[string]string{
					"sim_id": simID,
					"round":  fmt.Sprintf("%d", round),
					"rule":   newRuleJSON.String,
				},
				Embedding: computeNodeVec(label),
			})
			// agent → fracture (agent influenced / proposed this fracture)
			_ = b.g.UpsertEdge(Edge{
				From: agentNodeID, To: fractureID,
				Type: EdgeInfluences, Weight: 0.8,
			})
			// fracture → domain
			_ = b.g.UpsertEdge(Edge{
				From: fractureID, To: domainID,
				Type: EdgeBelongsTo, Weight: 0.5,
			})
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("graph builder: scan rounds for %s: %w", simID, err)
	}

	// ── 3. Votes ──────────────────────────────────────────────────────────
	b.indexVotes(simID)

	return nil
}

// indexVotes reads fracture votes and records voted_for / voted_against edges.
// Silently skips if the fracture_votes table does not yet exist.
func (b *Builder) indexVotes(simID string) {
	rows, err := b.db.Query(`
		SELECT voter_id, proposal_id, vote, weight
		FROM fracture_votes
		WHERE simulation_id = ?
	`, simID)
	if err != nil {
		// Table may not exist in older schemas — not fatal.
		return
	}
	defer rows.Close()

	for rows.Next() {
		var voterID, proposalID, vote string
		var weight float64
		if err := rows.Scan(&voterID, &proposalID, &vote, &weight); err != nil {
			continue
		}
		agentNodeID := "agent:" + voterID
		fractureNodeID := "fracture-vote:" + proposalID

		et := EdgeVotedFor
		if vote == "against" || vote == "no" {
			et = EdgeVotedAgainst
		}
		_ = b.g.UpsertEdge(Edge{
			From: agentNodeID, To: fractureNodeID,
			Type: et, Weight: weight,
		})
	}
}

// truncate shortens s to at most n runes, appending "…" if truncated.
func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "…"
}
