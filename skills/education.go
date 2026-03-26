package skills

import "github.com/fracture/fracture/engine"

// EducationSkill retorna a skill vertical de Education & EdTech.
func EducationSkill() *Skill {
	return &Skill{
		ID:   "education",
		Name: "Education & EdTech",
		Description: "Simulação especializada para instituições de ensino, edtechs, " +
			"treinamento corporativo e aprendizagem.",
		Industries: []string{
			"educação", "education", "edtech", "ensino",
			"escola", "universidade", "faculdade", "treinamento",
			"training", "learning", "e-learning", "EAD",
			"corporativo", "capacitação",
		},
		Rules: []*engine.Rule{
			{ID: "edu-001", Description: "MEC regulates and authorizes higher education institutions and courses", Domain: engine.DomainRegulation, Stability: 0.88, IsActive: true},
			{ID: "edu-002", Description: "ENADE and SINAES evaluate institutional quality for accreditation", Domain: engine.DomainRegulation, Stability: 0.82, IsActive: true},
			{ID: "edu-003", Description: "FIES and ProUni provide student financing for private institutions", Domain: engine.DomainFinance, Stability: 0.70, IsActive: true},
			{ID: "edu-004", Description: "EAD (distance learning) has no student limit per course since 2019", Domain: engine.DomainRegulation, Stability: 0.75, IsActive: true},
			{ID: "edu-005", Description: "Kroton/Cogna, Yduqs, Ser Educacional dominate the for-profit higher ed market", Domain: engine.DomainMarket, Stability: 0.65, IsActive: true},
			{ID: "edu-006", Description: "Corporate L&D budgets are shifting from classroom to digital platforms", Domain: engine.DomainBehavior, Stability: 0.50, IsActive: true},
			{ID: "edu-007", Description: "AI tutors and adaptive learning are replacing standardized content delivery", Domain: engine.DomainTechnology, Stability: 0.35, IsActive: true},
			{ID: "edu-008", Description: "Credentials and diplomas are still the primary hiring signal for most employers", Domain: engine.DomainBehavior, Stability: 0.72, IsActive: true},
			{ID: "edu-009", Description: "Skills-based hiring is growing but credential bias remains dominant", Domain: engine.DomainBehavior, Stability: 0.45, IsActive: true},
			{ID: "edu-010", Description: "International certifications (AWS, Google, Microsoft) compete with university degrees in tech", Domain: engine.DomainMarket, Stability: 0.48, IsActive: true},
		},
		Agents: []SkillAgent{
			{
				Name:        "Salman Khan",
				Role:        "Free Education & Mastery Learning Disruptor",
				Traits:      []string{"Khan Academy", "mastery learning", "free education", "AI tutor", "flipped classroom", "access over credential"},
				Goals:       []string{"free world-class education for everyone", "replace passive lecture with active learning"},
				Biases:      []string{"paid content behind paywalls", "lecture-based passive learning", "credential over competency"},
				Power:       0.85,
				IsDisruptor: true,
			},
			{
				Name:        "Howard Gardner",
				Role:        "Multiple Intelligences Theory Creator",
				Traits:      []string{"Frames of Mind", "multiple intelligences", "linguistic mathematical spatial musical bodily interpersonal intrapersonal naturalist", "intelligence is plural not singular", "education must serve diverse minds", "IQ as incomplete measure", "Harvard Project Zero"},
				Goals:       []string{"education that honors all types of intelligence", "assessment beyond standardized testing"},
				Biases:      []string{"single IQ measure as intelligence", "standardized curriculum ignoring individual strengths", "factory model education"},
				Power:       0.88,
				IsDisruptor: true,
			},
			{
				Name:        "Daphne Koller",
				Role:        "Online Learning Platform Pioneer",
				Traits:      []string{"Coursera", "MOOCs", "university partnerships", "professional certificates", "skills-based learning"},
				Goals:       []string{"democratize university-quality education", "skills credentials over degrees"},
				Biases:      []string{"traditional university credential monopoly", "geographic barriers to quality education"},
				Power:       0.80,
				IsDisruptor: true,
			},
			{
				Name:        "Paulo Freire",
				Role:        "Critical Pedagogy Pioneer & Education as Liberation",
				Traits:      []string{"Pedagogy of the Oppressed", "banking model of education critique", "conscientização", "dialogue as pedagogy", "education as political act", "praxis reflection and action", "literacy as empowerment"},
				Goals:       []string{"education liberates rather than domesticates", "learner as subject not object of education"},
				Biases:      []string{"banking model passive information transfer", "education that reproduces inequality", "standardized testing as measurement of human worth"},
				Power:       0.95,
				IsDisruptor: true,
			},
		},
		Context: `EDUCATION SECTOR CONTEXT FOR FRACTURE SIMULATION:
Brazil has the largest private higher education market in the world by enrollment.
Key players: Cogna/Kroton (largest), Yduqs, Ser Educacional, Anima,
FUVEST/USP (public), FGV, Insper (premium private).
Key regulators: MEC (authorization and accreditation), CAPES (graduate programs),
INEP (ENADE quality assessment).
Critical dynamics: EAD (distance learning) now represents 60%+ of enrollments,
AI tutors beginning to replace standardized content,
skills-based hiring threatening traditional degree value,
international platforms (Coursera, Udemy, LinkedIn Learning) competing,
corporate L&D shifting to microlearning and platforms,
FIES student financing creating/destroying enrollment cycles.`,
		Queries: []string{
			"edtech Brazil investment education disruption 2024 2025",
			"EAD ensino distância regulação MEC crescimento Brasil",
			"inteligência artificial educação tutor AI Brasil impacto",
			"skills-based hiring credencial diploma disruption",
			"Cogna Kroton Yduqs mercado educação superior Brasil",
		},
	}
}
