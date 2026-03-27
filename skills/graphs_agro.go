package skills

func init() {
	RegisterGraph(&SkillGraph{
		SkillID: "agro",
		Relations: []MindRelation{
			{From: "Norman Borlaug", To: "Vandana Shiva",
				Type:        "contradicts",
				Description: "Borlaug's yield maximization through science directly contradicts Shiva's seed sovereignty and biodiversity preservation.",
				Strength:    0.90},
			{From: "Gordon Conway", To: "Norman Borlaug",
				Type:        "complements",
				Description: "Conway's doubly green revolution synthesizes Borlaug's productivity imperative with ecological sustainability.",
				Strength:    0.85},
			{From: "Raj Patel", To: "Norman Borlaug",
				Type:        "critiques",
				Description: "Patel documents the externalized social and ecological costs of Borlaug's Green Revolution that yield metrics ignore.",
				Strength:    0.82},
			{From: "Vandana Shiva", To: "Gordon Conway",
				Type:        "critiques",
				Description: "Shiva critiques Conway's synthesis as still industrial — doubly green is still dependent on external inputs.",
				Strength:    0.70},
		},
	})
}
