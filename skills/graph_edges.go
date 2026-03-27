package skills

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// EdgeType classifies the semantic relationship between two nodes.
type EdgeType string

const (
	// EdgeInfluences: agent → fracture/rule — agent tensioned or proposed this.
	EdgeInfluences EdgeType = "influences"
	// EdgeCaused: fracture → outcome — a fracture event led to a real-world outcome.
	EdgeCaused EdgeType = "caused"
	// EdgeVotedFor: agent → fracture — agent voted in favour of the proposal.
	EdgeVotedFor EdgeType = "voted_for"
	// EdgeVotedAgainst: agent → fracture — agent voted against the proposal.
	EdgeVotedAgainst EdgeType = "voted_against"
	// EdgeCoOccurred: rule ↔ rule — both were tensioned in the same round.
	EdgeCoOccurred EdgeType = "co_occurred"
	// EdgeBelongsTo: rule/agent/fracture → domain — node lives in this domain.
	EdgeBelongsTo EdgeType = "belongs_to"
)

// Edge is a directed, weighted relationship between two graph nodes.
type Edge struct {
	ID        string
	From      string
	To        string
	Type      EdgeType
	Weight    float64           // [0, 1] — higher means stronger / more observed
	Props     map[string]string // arbitrary metadata
	CreatedAt time.Time
}

// UpsertEdge inserts a directed edge or, on conflict (same from/to/type),
// increments its weight by 0.1 (capped at 1.0).
func (g *Graph) UpsertEdge(e Edge) error {
	if e.ID == "" {
		e.ID = edgeID(e.From, e.To, string(e.Type))
	}
	if e.Weight == 0 {
		e.Weight = 0.5
	}
	props, _ := json.Marshal(e.Props)
	_, err := g.db.Exec(`
		INSERT INTO graph_edges (id, from_id, to_id, edge_type, weight, props, created_at)
		VALUES (?, ?, ?, ?, ?, ?, unixepoch())
		ON CONFLICT(from_id, to_id, edge_type) DO UPDATE SET
			weight = MIN(1.0, weight + 0.1),
			props  = excluded.props
	`, e.ID, e.From, e.To, string(e.Type), e.Weight, string(props))
	return err
}

// EdgesFrom returns all edges originating from nodeID.
// If types is non-empty, only edges whose type is in the list are returned.
// Results are ordered by weight descending.
func (g *Graph) EdgesFrom(nodeID string, types []EdgeType) ([]Edge, error) {
	if len(types) == 0 {
		rows, err := g.db.Query(`
			SELECT id, from_id, to_id, edge_type, weight, props, created_at
			FROM graph_edges
			WHERE from_id = ?
			ORDER BY weight DESC
		`, nodeID)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		return scanEdges(rows)
	}

	// Build parameterised IN clause dynamically.
	args := make([]interface{}, 0, len(types)+1)
	args = append(args, nodeID)
	ph := "("
	for i, t := range types {
		if i > 0 {
			ph += ","
		}
		ph += "?"
		args = append(args, string(t))
	}
	ph += ")"

	rows, err := g.db.Query(fmt.Sprintf(`
		SELECT id, from_id, to_id, edge_type, weight, props, created_at
		FROM graph_edges
		WHERE from_id = ? AND edge_type IN %s
		ORDER BY weight DESC
	`, ph), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanEdges(rows)
}

// EdgesTo returns all edges pointing to nodeID, ordered by weight descending.
func (g *Graph) EdgesTo(nodeID string) ([]Edge, error) {
	rows, err := g.db.Query(`
		SELECT id, from_id, to_id, edge_type, weight, props, created_at
		FROM graph_edges
		WHERE to_id = ?
		ORDER BY weight DESC
	`, nodeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanEdges(rows)
}

// scanEdges converts a *sql.Rows result into a []Edge slice.
func scanEdges(rows *sql.Rows) ([]Edge, error) {
	var edges []Edge
	for rows.Next() {
		var (
			e         Edge
			propsJSON string
			createdAt int64
		)
		if err := rows.Scan(
			&e.ID, &e.From, &e.To, (*string)(&e.Type),
			&e.Weight, &propsJSON, &createdAt,
		); err != nil {
			continue
		}
		e.CreatedAt = time.Unix(createdAt, 0)
		json.Unmarshal([]byte(propsJSON), &e.Props) //nolint:errcheck
		edges = append(edges, e)
	}
	return edges, rows.Err()
}

// edgeID builds a deterministic ID string for an edge triple.
func edgeID(from, to, edgeType string) string {
	return fmt.Sprintf("%d", fnvStr(from+"|"+to+"|"+edgeType))
}
