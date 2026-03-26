package skills

import (
	"strings"

	"github.com/fracture/fracture/engine"
)

// Skill representa um pacote de conhecimento vertical
// que enriquece a simulação para um setor específico.
type Skill struct {
	ID          string         // ex: "healthcare"
	Name        string         // ex: "Healthcare & Life Sciences"
	Description string         // para o usuário
	Industries  []string       // ex: ["hospital", "pharma", "healthtech"]
	Rules       []*engine.Rule // regras específicas do setor
	Agents      []SkillAgent   // arquétipos extras do setor
	Context     string         // contexto pré-carregado para DeepSearch
	Queries     []string       // queries especializadas para DeepSearch
}

// SkillAgent descreve um arquétipo extra injetado pela skill.
type SkillAgent struct {
	Name        string
	Role        string
	Traits      []string
	Goals       []string
	Biases      []string
	Power       float64
	IsDisruptor bool
}

// Registry mapeia industry slug → Skill.
var Registry = map[string]*Skill{}

// Register indexa a skill por ID e por cada industry slug.
func Register(s *Skill) {
	Registry[s.ID] = s
	for _, industry := range s.Industries {
		Registry[industry] = s
	}
}

// Detect tenta inferir a skill a partir de palavras-chave na pergunta ou department.
func Detect(question, department string) *Skill {
	q := strings.ToLower(question + " " + department)
	// Use a set to avoid double-checking same skill multiple times
	checked := map[string]bool{}
	for _, skill := range Registry {
		if checked[skill.ID] {
			continue
		}
		checked[skill.ID] = true
		for _, industry := range skill.Industries {
			if strings.Contains(q, strings.ToLower(industry)) {
				return skill
			}
		}
		if strings.Contains(q, skill.ID) {
			return skill
		}
	}
	return nil
}
