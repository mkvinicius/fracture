package skills

import "strings"

type MindRelation struct {
	From        string
	To          string
	Type        string  // "influenced", "contradicts", "complements", "critiques"
	Description string
	Strength    float64
}

type SkillGraph struct {
	SkillID   string
	Relations []MindRelation
}

var Graphs = map[string]*SkillGraph{}

func RegisterGraph(g *SkillGraph) {
	Graphs[g.SkillID] = g
}

func GetRelations(skillID, agentName string) []MindRelation {
	g, ok := Graphs[skillID]
	if !ok {
		return nil
	}
	var result []MindRelation
	for _, r := range g.Relations {
		if r.From == agentName || r.To == agentName {
			result = append(result, r)
		}
	}
	return result
}

// AllMindNames returns all unique mind names registered in a skill's graph.
func AllMindNames(skillID string) []string {
	g, ok := Graphs[skillID]
	if !ok {
		return nil
	}
	seen := map[string]bool{}
	var names []string
	for _, r := range g.Relations {
		if !seen[r.From] {
			seen[r.From] = true
			names = append(names, r.From)
		}
		if !seen[r.To] {
			seen[r.To] = true
			names = append(names, r.To)
		}
	}
	return names
}

func FormatRelationsContext(skillID string, agentNames []string) string {
	g, ok := Graphs[skillID]
	if !ok {
		return ""
	}
	nameSet := map[string]bool{}
	for _, n := range agentNames {
		nameSet[n] = true
	}
	var lines []string
	seen := map[string]bool{}
	for _, r := range g.Relations {
		if !nameSet[r.From] && !nameSet[r.To] {
			continue
		}
		key := r.From + "|" + r.To
		if seen[key] {
			continue
		}
		seen[key] = true
		typeLabel := map[string]string{
			"influenced":  "INFLUENCIOU",
			"contradicts": "CONTRADIZ",
			"complements": "COMPLEMENTA",
			"critiques":   "CRITICA",
		}[r.Type]
		if typeLabel == "" {
			typeLabel = strings.ToUpper(r.Type)
		}
		lines = append(lines,
			r.From+" "+typeLabel+" "+r.To+": "+r.Description)
	}
	if len(lines) == 0 {
		return ""
	}
	return "Relações intelectuais entre os participantes:\n" +
		strings.Join(lines, "\n")
}
