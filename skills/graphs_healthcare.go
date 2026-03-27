package skills

func init() {
	RegisterGraph(&SkillGraph{
		SkillID: "healthcare",
		Relations: []MindRelation{
			{From: "Donald Berwick", To: "Atul Gawande",
				Type:        "influenced",
				Description: "Gawande's checklist manifesto builds directly on Berwick's IHI systems thinking and patient safety work.",
				Strength:    0.88},
			{From: "Paul Farmer", To: "Elisabeth Rosenthal",
				Type:        "contradicts",
				Description: "Farmer focuses on access as primary barrier; Rosenthal focuses on cost as primary barrier. Both right, different entry points.",
				Strength:    0.75},
			{From: "Atul Gawande", To: "Donald Berwick",
				Type:        "complements",
				Description: "Gawande provides clinical narrative evidence for Berwick's systemic quality improvement arguments.",
				Strength:    0.90},
			{From: "Hans Rosling", To: "Paul Farmer",
				Type:        "critiques",
				Description: "Rosling's data-driven factfulness critiques narrative-based advocacy — even well-intentioned like Farmer's.",
				Strength:    0.70},
			{From: "Eric Topol", To: "Donald Berwick",
				Type:        "complements",
				Description: "Topol's digital medicine adds the technological layer to Berwick's systems-based quality improvement.",
				Strength:    0.82},
		},
	})
}
