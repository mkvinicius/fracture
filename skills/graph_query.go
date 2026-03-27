package skills

import (
	"fmt"
	"sort"
	"strings"
)

// ContextFragment is a piece of graph-retrieved evidence for council enrichment.
type ContextFragment struct {
	Text     string   // human-readable evidence string for LLM prompts
	Score    float64  // relevance to the query [0, 1]
	NodeType NodeType // origin node type
	Source   string   // origin node label
}

// queryRelated is the internal retrieval function. It ranks all stored nodes of
// council-relevant types by cosine similarity to the query vector, boosts
// fracture nodes that belong to the target domain, and returns up to limit results.
func (g *Graph) queryRelated(domain, question string, limit int) ([]ContextFragment, error) {
	if limit <= 0 {
		limit = 5
	}
	queryVec := computeNodeVec(question + " " + domain)
	domainID := "domain:" + domain

	var candidates []ContextFragment

	for _, ntype := range []NodeType{NodeFracture, NodeAgent, NodeRule, NodeOutcome} {
		nodes, err := g.ListNodes(ntype, limit*4)
		if err != nil {
			continue
		}
		for _, n := range nodes {
			emb := n.Embedding
			if len(emb) == 0 {
				emb = computeNodeVec(n.Label)
			}
			score := cosineSim(queryVec, emb)

			// Boost fracture nodes that belong to the queried domain.
			if ntype == NodeFracture {
				edges, _ := g.EdgesFrom(n.ID, []EdgeType{EdgeBelongsTo})
				for _, e := range edges {
					if e.To == domainID {
						score = score*0.6 + 0.4*e.Weight
						break
					}
				}
			}

			if score > 0.05 {
				candidates = append(candidates, ContextFragment{
					Text:     formatNodeContext(n, g),
					Score:    score,
					NodeType: ntype,
					Source:   n.Label,
				})
			}
		}
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Score > candidates[j].Score
	})
	if len(candidates) > limit {
		candidates = candidates[:limit]
	}
	return candidates, nil
}

// Neighbors returns the direct neighbours of nodeID reachable via edgeTypes.
// Pass nil edgeTypes to traverse all edge types. Results are ordered by edge weight.
func (g *Graph) Neighbors(nodeID string, edgeTypes []EdgeType, limit int) ([]Node, error) {
	edges, err := g.EdgesFrom(nodeID, edgeTypes)
	if err != nil {
		return nil, fmt.Errorf("neighbors of %q: %w", nodeID, err)
	}
	nodes := make([]Node, 0, minInt(limit, len(edges)))
	for i, e := range edges {
		if i >= limit {
			break
		}
		n, err := g.GetNode(e.To)
		if err != nil {
			continue
		}
		nodes = append(nodes, n)
	}
	return nodes, nil
}

// PathBetween finds a path from fromID to toID using breadth-first search up to
// maxDepth hops. Returns the sequence of nodes along the shortest path found,
// or nil if no path exists within the depth limit.
func (g *Graph) PathBetween(fromID, toID string, maxDepth int) ([]Node, error) {
	type step struct {
		id   string
		path []string
	}

	visited := map[string]bool{fromID: true}
	queue := []step{{id: fromID, path: []string{fromID}}}

	for depth := 0; depth < maxDepth && len(queue) > 0; depth++ {
		next := queue
		queue = nil
		for _, s := range next {
			edges, err := g.EdgesFrom(s.id, nil)
			if err != nil {
				continue
			}
			for _, e := range edges {
				if visited[e.To] {
					continue
				}
				newPath := append(append([]string(nil), s.path...), e.To)
				if e.To == toID {
					return resolveNodeIDs(g, newPath)
				}
				visited[e.To] = true
				queue = append(queue, step{id: e.To, path: newPath})
			}
		}
	}
	return nil, nil // no path found within maxDepth
}

// ─── helpers ─────────────────────────────────────────────────────────────────

// formatNodeContext renders a node as a human-readable string for LLM injection.
func formatNodeContext(n Node, g *Graph) string {
	switch n.Type {
	case NodeFracture:
		sim := n.Props["sim_id"]
		round := n.Props["round"]
		if sim != "" && round != "" {
			return fmt.Sprintf("Past fracture (sim %.8s, round %s): %s", sim, round, n.Label)
		}
		return "Past fracture: " + n.Label

	case NodeAgent:
		agentType := n.Props["type"]
		edges, _ := g.EdgesFrom(n.ID, []EdgeType{EdgeInfluences})
		influenced := make([]string, 0, len(edges))
		for _, e := range edges {
			if node, err := g.GetNode(e.To); err == nil {
				influenced = append(influenced, node.Label)
			}
		}
		base := fmt.Sprintf("Agent %s (%s)", n.Label, agentType)
		if len(influenced) > 0 {
			top := influenced
			if len(top) > 3 {
				top = top[:3]
			}
			base += " influenced: " + strings.Join(top, ", ")
		}
		return base

	case NodeRule:
		return "Rule: " + n.Label

	case NodeOutcome:
		return "Observed outcome: " + n.Label

	default:
		return n.Label
	}
}

// resolveNodeIDs converts a slice of node IDs to the corresponding Node structs.
// Nodes that cannot be found are skipped silently.
func resolveNodeIDs(g *Graph, ids []string) ([]Node, error) {
	nodes := make([]Node, 0, len(ids))
	for _, id := range ids {
		n, err := g.GetNode(id)
		if err != nil {
			continue
		}
		nodes = append(nodes, n)
	}
	return nodes, nil
}

// minInt returns the smaller of a and b.
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
