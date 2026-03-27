package skills

func init() {
	RegisterGraph(&SkillGraph{
		SkillID: "fintech",
		Relations: []MindRelation{
			{From: "Hernando de Soto", To: "Muhammad Yunus",
				Type:        "influenced",
				Description: "De Soto's property rights for the poor directly informed Yunus's microcredit model — both attack same exclusion from different angles.",
				Strength:    0.82},
			{From: "David Vélez", To: "Arminio Fraga",
				Type:        "complements",
				Description: "Vélez's digital disruption operates within the monetary stability framework Fraga built as BACEN president.",
				Strength:    0.78},
			{From: "Muhammad Yunus", To: "André Esteves",
				Type:        "critiques",
				Description: "Yunus's social business model directly critiques investment banking's profit maximization without social purpose.",
				Strength:    0.80},
			{From: "Arminio Fraga", To: "David Vélez",
				Type:        "complements",
				Description: "Fraga's credibility-first monetary policy created the stable environment where Nubank's disruption became possible.",
				Strength:    0.75},
		},
	})
}
