package skills

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// NodeType classifies what a node in the knowledge graph represents.
type NodeType string

const (
	NodeRule     NodeType = "rule"     // a simulation rule (built-in or custom)
	NodeAgent    NodeType = "agent"    // an archetype / agent ID
	NodeFracture NodeType = "fracture" // a proposed or accepted fracture event
	NodeDomain   NodeType = "domain"   // a RuleDomain (market, technology, …)
	NodeOutcome  NodeType = "outcome"  // a real-world outcome recorded via feedback
)

// Node is a vertex in the GraphRAG knowledge graph.
type Node struct {
	ID        string
	Type      NodeType
	Label     string            // human-readable name / description
	Props     map[string]string // arbitrary key-value metadata
	Embedding []float64         // TF-IDF vector (embedDim dimensions)
	CreatedAt time.Time
}

// UpsertNode inserts a node or updates its label, embedding, and props on conflict.
func (g *Graph) UpsertNode(n Node) error {
	props, _ := json.Marshal(n.Props)
	emb := serializeVec(n.Embedding)
	_, err := g.db.Exec(`
		INSERT INTO graph_nodes (id, node_type, label, embedding, props, created_at)
		VALUES (?, ?, ?, ?, ?, unixepoch())
		ON CONFLICT(id) DO UPDATE SET
			label     = excluded.label,
			embedding = excluded.embedding,
			props     = excluded.props
	`, n.ID, string(n.Type), n.Label, emb, string(props))
	return err
}

// GetNode retrieves a single node by ID.
func (g *Graph) GetNode(id string) (Node, error) {
	var (
		n         Node
		propsJSON string
		embBlob   []byte
		createdAt int64
	)
	err := g.db.QueryRow(`
		SELECT id, node_type, label, embedding, props, created_at
		FROM graph_nodes WHERE id = ?
	`, id).Scan(&n.ID, (*string)(&n.Type), &n.Label, &embBlob, &propsJSON, &createdAt)
	if err == sql.ErrNoRows {
		return Node{}, fmt.Errorf("graph node %q not found", id)
	}
	if err != nil {
		return Node{}, err
	}
	n.Embedding = deserializeVec(embBlob)
	n.CreatedAt = time.Unix(createdAt, 0)
	json.Unmarshal([]byte(propsJSON), &n.Props) //nolint:errcheck
	return n, nil
}

// ListNodes returns up to limit nodes of the given type, newest first.
func (g *Graph) ListNodes(ntype NodeType, limit int) ([]Node, error) {
	rows, err := g.db.Query(`
		SELECT id, node_type, label, embedding, props, created_at
		FROM graph_nodes
		WHERE node_type = ?
		ORDER BY created_at DESC
		LIMIT ?
	`, string(ntype), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanNodes(rows)
}

// scanNodes converts a *sql.Rows result into a []Node slice.
func scanNodes(rows *sql.Rows) ([]Node, error) {
	var nodes []Node
	for rows.Next() {
		var (
			n         Node
			propsJSON string
			embBlob   []byte
			createdAt int64
		)
		if err := rows.Scan(
			&n.ID, (*string)(&n.Type), &n.Label, &embBlob, &propsJSON, &createdAt,
		); err != nil {
			continue
		}
		n.Embedding = deserializeVec(embBlob)
		n.CreatedAt = time.Unix(createdAt, 0)
		json.Unmarshal([]byte(propsJSON), &n.Props) //nolint:errcheck
		nodes = append(nodes, n)
	}
	return nodes, rows.Err()
}
