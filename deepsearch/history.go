package deepsearch

import (
	"context"
	"fmt"
	"strings"

	"github.com/fracture/fracture/memory"
)

// EnrichWithHistory retrieves past simulation context for the given company and
// question, returning a string ready to be injected as Evidence in the World.
//
// Pipeline:
//  a) Search RAGStore for 3 most similar past simulation documents
//  b) Extract fracture points from those documents
//  c) Return a formatted context string listing the most frequent past ruptures
func (a *Agent) EnrichWithHistory(
	ctx context.Context,
	ragStore *memory.RAGStore,
	companyID, question string,
) string {
	if ragStore == nil || companyID == "" {
		return ""
	}

	docs, err := ragStore.Search(companyID, question, 3)
	if err != nil || len(docs) == 0 {
		return ""
	}

	// Collect unique fracture/rupture descriptions from retrieved documents
	seen := make(map[string]bool)
	var points []string

	for _, doc := range docs {
		if doc.Type != memory.DocFracturePoint && doc.Type != memory.DocCompanyContext {
			continue
		}
		// Extract the descriptive part after the colon (if any)
		content := strings.TrimSpace(doc.Content)
		if content == "" || seen[content] {
			continue
		}
		seen[content] = true
		points = append(points, content)
	}

	if len(points) == 0 {
		// Fall back to all retrieved docs
		for _, doc := range docs {
			content := strings.TrimSpace(doc.Content)
			if content != "" && !seen[content] {
				seen[content] = true
				points = append(points, content)
			}
		}
	}

	if len(points) == 0 {
		return ""
	}

	var sb strings.Builder
	fmt.Fprintf(&sb,
		"Em simulações anteriores para esta empresa, as rupturas mais frequentes foram:\n",
	)
	for i, p := range points {
		// Truncate long descriptions to keep the context concise
		if len(p) > 200 {
			p = p[:200] + "…"
		}
		fmt.Fprintf(&sb, "%d. %s\n", i+1, p)
	}
	return sb.String()
}
