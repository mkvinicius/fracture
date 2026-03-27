package skills

func init() {
	RegisterGraph(&SkillGraph{
		SkillID: "education",
		Relations: []MindRelation{
			{From: "Paulo Freire", To: "Salman Khan",
				Type:        "critiques",
				Description: "Freire would critique Khan Academy as banking model digitized — depositing knowledge into passive learners at scale.",
				Strength:    0.78},
			{From: "Howard Gardner", To: "Salman Khan",
				Type:        "influenced",
				Description: "Khan Academy's diverse content types reflect Gardner's multiple intelligences — different paths to same knowledge.",
				Strength:    0.72},
			{From: "Salman Khan", To: "Daphne Koller",
				Type:        "complements",
				Description: "Khan's free K-12 model and Koller's university-level Coursera are complementary layers of open education.",
				Strength:    0.85},
			{From: "Paulo Freire", To: "Howard Gardner",
				Type:        "contradicts",
				Description: "Freire's liberation pedagogy focuses on political consciousness; Gardner's multiple intelligences focuses on cognitive diversity. Different goals.",
				Strength:    0.65},
		},
	})
}
