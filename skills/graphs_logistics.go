package skills

func init() {
	RegisterGraph(&SkillGraph{
		SkillID: "logistics",
		Relations: []MindRelation{
			{From: "Hau Lee", To: "Martin Christopher",
				Type:        "complements",
				Description: "Lee's Triple-A supply chain (agile, adaptable, aligned) complements Christopher's time compression and relationship focus.",
				Strength:    0.85},
			{From: "Martin Christopher", To: "Yossi Sheffi",
				Type:        "influenced",
				Description: "Christopher's supply chain vulnerability work directly influenced Sheffi's resilient enterprise framework.",
				Strength:    0.80},
			{From: "Hau Lee", To: "Yossi Sheffi",
				Type:        "complements",
				Description: "Lee's agility focus and Sheffi's resilience focus are complementary — agility enables resilience.",
				Strength:    0.88},
			{From: "Edward Glaeser", To: "Hau Lee",
				Type:        "complements",
				Description: "Glaeser's urban agglomeration economics explains why logistics hubs form where Lee's supply chains concentrate.",
				Strength:    0.70},
		},
	})
}
