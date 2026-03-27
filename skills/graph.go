// Package skills provides the GraphRAG knowledge graph for FRACTURE.
//
// It stores simulation artefacts (agents, rules, fracture events, outcomes)
// as nodes and their semantic relationships as weighted directed edges.
// Councils query the graph to inject relevant past-simulation context into
// their LLM prompts, improving debate quality across repeated simulations.
//
// Schema (two tables, created on first use):
//
//	graph_nodes  – vertices: id, node_type, label, embedding (BLOB), props (JSON)
//	graph_edges  – directed edges: from_id, to_id, edge_type, weight
//
// The package does NOT import engine or any other FRACTURE package, so it can
// be safely imported by engine without creating a circular dependency.
package skills

import (
	"database/sql"
	"fmt"
	"strings"
)

// Graph is the in-process GraphRAG skill backed by SQLite.
// Create one per process with NewGraph and pass it to engine.BuildCouncilsWithGraph.
type Graph struct {
	db *sql.DB
}

// NewGraph initialises the Graph, creating its tables if they do not exist.
func NewGraph(db *sql.DB) (*Graph, error) {
	g := &Graph{db: db}
	if err := g.initSchema(); err != nil {
		return nil, fmt.Errorf("graphrag: init schema: %w", err)
	}
	return g, nil
}

// graphSchema creates the two graph tables and their covering indices.
// All statements use IF NOT EXISTS so they are safe to run on every startup.
const graphSchema = `
CREATE TABLE IF NOT EXISTS graph_nodes (
	id         TEXT    PRIMARY KEY,
	node_type  TEXT    NOT NULL,
	label      TEXT    NOT NULL,
	embedding  BLOB,
	props      TEXT    NOT NULL DEFAULT '{}',
	created_at INTEGER NOT NULL DEFAULT (unixepoch())
);
CREATE INDEX IF NOT EXISTS idx_graph_nodes_type ON graph_nodes(node_type);

CREATE TABLE IF NOT EXISTS graph_edges (
	id         TEXT    PRIMARY KEY,
	from_id    TEXT    NOT NULL,
	to_id      TEXT    NOT NULL,
	edge_type  TEXT    NOT NULL,
	weight     REAL    NOT NULL DEFAULT 0.5,
	props      TEXT    NOT NULL DEFAULT '{}',
	created_at INTEGER NOT NULL DEFAULT (unixepoch()),
	UNIQUE(from_id, to_id, edge_type)
);
CREATE INDEX IF NOT EXISTS idx_graph_edges_from ON graph_edges(from_id);
CREATE INDEX IF NOT EXISTS idx_graph_edges_to   ON graph_edges(to_id);
CREATE INDEX IF NOT EXISTS idx_graph_edges_type ON graph_edges(edge_type);
`

func (g *Graph) initSchema() error {
	_, err := g.db.Exec(graphSchema)
	return err
}

// ─── engine.SkillGraph implementation ────────────────────────────────────────

// RelatedContext retrieves up to limit relevant past-simulation context strings
// for the given domain and question. It satisfies the engine.SkillGraph interface,
// returning plain strings so engine does not need to import this package.
func (g *Graph) RelatedContext(domain, question string, limit int) ([]string, error) {
	frags, err := g.queryRelated(domain, question, limit)
	if err != nil {
		return nil, err
	}
	texts := make([]string, 0, len(frags))
	for _, f := range frags {
		texts = append(texts, f.Text)
	}
	return texts, nil
}

// ─── Live recording helpers ───────────────────────────────────────────────────

// RecordFracture adds a fracture node and its relationship to the proposing agent.
// Call this from the simulation runner after each fracture event to keep the
// graph up to date without running a full FromSimulation pass.
func (g *Graph) RecordFracture(simID, ruleID, ruleDesc, agentID string, round int, accepted bool) error {
	fractureID := fmt.Sprintf("fracture:%s:r%d:%s", simID, round, agentID)
	_ = g.UpsertNode(Node{
		ID:    fractureID,
		Type:  NodeFracture,
		Label: truncate(ruleDesc, 80),
		Props: map[string]string{
			"sim_id":   simID,
			"rule_id":  ruleID,
			"round":    fmt.Sprintf("%d", round),
			"accepted": fmt.Sprintf("%v", accepted),
		},
		Embedding: computeNodeVec(ruleDesc),
	})
	agentNodeID := "agent:" + agentID
	_ = g.UpsertNode(Node{ID: agentNodeID, Type: NodeAgent, Label: agentID})
	return g.UpsertEdge(Edge{
		From: agentNodeID, To: fractureID,
		Type: EdgeInfluences, Weight: 0.8,
	})
}

// RecordVote records a vote edge from an agent to a fracture proposal.
// vote should be "for" / "yes" (→ EdgeVotedFor) or "against" / "no" (→ EdgeVotedAgainst).
func (g *Graph) RecordVote(simID, agentID, proposalID, vote string, weight float64) error {
	agentNodeID := "agent:" + agentID
	_ = g.UpsertNode(Node{ID: agentNodeID, Type: NodeAgent, Label: agentID})
	fractureNodeID := "fracture-vote:" + proposalID
	et := EdgeVotedFor
	if strings.EqualFold(vote, "against") || strings.EqualFold(vote, "no") {
		et = EdgeVotedAgainst
	}
	return g.UpsertEdge(Edge{
		From: agentNodeID, To: fractureNodeID,
		Type: et, Weight: weight,
	})
}
