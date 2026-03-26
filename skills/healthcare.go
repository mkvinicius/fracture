package skills

import "github.com/fracture/fracture/engine"

// HealthcareSkill retorna a skill vertical de Healthcare & Life Sciences.
func HealthcareSkill() *Skill {
	return &Skill{
		ID:   "healthcare",
		Name: "Healthcare & Life Sciences",
		Description: "Simulação especializada para hospitais, planos de saúde, farmacêuticas, " +
			"healthtechs e dispositivos médicos.",
		Industries: []string{
			"healthcare", "hospital", "saúde", "plano de saúde",
			"farmacêutica", "pharma", "healthtech", "medical",
			"clínica", "laboratório", "diagnóstico", "biotech",
		},
		Rules: []*engine.Rule{
			{ID: "hlt-001", Description: "ANVISA approval is mandatory before any product reaches the Brazilian market", Domain: engine.DomainRegulation, Stability: 0.92, IsActive: true},
			{ID: "hlt-002", Description: "ANS regulates health plan pricing, coverage, and network adequacy", Domain: engine.DomainRegulation, Stability: 0.88, IsActive: true},
			{ID: "hlt-003", Description: "CFM and CRM regulate medical practice and telemedicine boundaries", Domain: engine.DomainRegulation, Stability: 0.85, IsActive: true},
			{ID: "hlt-004", Description: "TISS standard controls data interoperability between payers and providers", Domain: engine.DomainTechnology, Stability: 0.70, IsActive: true},
			{ID: "hlt-005", Description: "Fee-for-service reimbursement model dominates Brazilian healthcare", Domain: engine.DomainFinance, Stability: 0.60, IsActive: true},
			{ID: "hlt-006", Description: "Healthcare data is protected under LGPD with sector-specific requirements", Domain: engine.DomainRegulation, Stability: 0.82, IsActive: true},
			{ID: "hlt-007", Description: "Physician autonomy in treatment decisions is legally protected", Domain: engine.DomainBehavior, Stability: 0.75, IsActive: true},
			{ID: "hlt-008", Description: "AI diagnostics must be validated by licensed physician before clinical use", Domain: engine.DomainRegulation, Stability: 0.55, IsActive: true},
			{ID: "hlt-009", Description: "Hospital beds and surgical suites require physical infrastructure investment", Domain: engine.DomainMarket, Stability: 0.80, IsActive: true},
			{ID: "hlt-010", Description: "Value-based care models are emerging but fee-for-service remains dominant", Domain: engine.DomainFinance, Stability: 0.45, IsActive: true},
			{ID: "hlt-011", Description: "Telemedicine is now permanently legal in Brazil post-COVID", Domain: engine.DomainRegulation, Stability: 0.78, IsActive: true},
			{ID: "hlt-012", Description: "Drug patent cliff creates generic opportunity every 10-20 years", Domain: engine.DomainMarket, Stability: 0.65, IsActive: true},
		},
		Agents: []SkillAgent{
			{
				Name:        "Paul Farmer",
				Role:        "Global Health Equity Champion",
				Traits:      []string{"health as human right", "accompaniment model", "community health workers", "structural violence", "Partners in Health"},
				Goals:       []string{"universal access to care", "equity over efficiency"},
				Biases:      []string{"profit-driven healthcare", "technology over human connection"},
				Power:       0.80,
				IsDisruptor: true,
			},
			{
				Name:        "Elisabeth Rosenthal",
				Role:        "Healthcare Cost Investigator & Patient Advocate",
				Traits:      []string{"An American Sickness", "medical billing opacity", "chargemaster", "price transparency", "healthcare industrial complex"},
				Goals:       []string{"price transparency", "patient protection from billing surprises"},
				Biases:      []string{"incumbent pricing opacity", "facility fees", "surprise billing"},
				Power:       0.70,
				IsDisruptor: true,
			},
			{
				Name:        "Atul Gawande",
				Role:        "Healthcare Systems & Quality Improvement Expert",
				Traits:      []string{"The Checklist Manifesto", "complications", "being mortal", "healthcare as craft", "systems thinking"},
				Goals:       []string{"reduce preventable harm", "improve care delivery systems"},
				Biases:      []string{"disruption that ignores clinical complexity", "technology without validation"},
				Power:       0.90,
				IsDisruptor: false,
			},
			{
				Name:        "Eric Topol",
				Role:        "Digital Medicine & AI Diagnostics Pioneer",
				Traits:      []string{"deep medicine", "AI in cardiology", "digital biomarkers", "smartphone diagnostics", "precision medicine"},
				Goals:       []string{"AI-augmented physician", "democratize diagnostics"},
				Biases:      []string{"slow regulatory approval for proven AI tools", "physician resistance to evidence-based tech"},
				Power:       0.85,
				IsDisruptor: true,
			},
			{
				Name:        "Ana Paula Etges",
				Role:        "Brazilian Health Economics & Value-Based Care Expert",
				Traits:      []string{"value-based healthcare", "VBHC Brazil", "outcomes measurement", "ICHOM", "bundled payments"},
				Goals:       []string{"transition from fee-for-service to value-based", "outcomes transparency"},
				Biases:      []string{"fee-for-service perpetuation", "volume over outcomes"},
				Power:       0.75,
				IsDisruptor: true,
			},
			{
				Name:        "Donald Berwick",
				Role:        "Healthcare Quality & Patient Safety Revolutionary",
				Traits:      []string{"Institute for Healthcare Improvement", "100,000 Lives Campaign", "Triple Aim framework", "patient safety systems", "zero preventable harm", "healthcare as system not individual acts", "improvement science"},
				Goals:       []string{"zero preventable deaths in healthcare", "improve experience of care population health and reduce cost simultaneously"},
				Biases:      []string{"individual blame culture instead of systems thinking", "profit over patient safety", "innovation without safety validation"},
				Power:       0.92,
				IsDisruptor: false,
			},
			{
				Name:        "Hans Rosling",
				Role:        "Global Health Data & Factfulness Pioneer",
				Traits:      []string{"Factfulness", "gap instinct", "negativity instinct", "data over drama", "world is better than you think", "Gapminder Foundation", "washing machine as most important invention"},
				Goals:       []string{"replace dramatic worldview with data-based worldview", "10 instincts that distort health data perception", "the world has improved dramatically — and almost nobody knows it", "Gapminder data revealing real health progress hidden by negativity instinct"},
				Biases:      []string{"pessimistic narratives not supported by data", "gap thinking dividing world into developed and developing", "negativity instinct amplified by media"},
				Power:       0.88,
				IsDisruptor: true,
			},
		},
		Context: `HEALTHCARE SECTOR CONTEXT FOR FRACTURE SIMULATION:
Brazil has the largest private healthcare market in Latin America.
Key players: Hapvida-NotreDame (largest health plan), Rede D'Or (largest hospital network),
Bradesco Saúde, SulAmérica, UNIMED cooperatives.
Key regulators: ANVISA (products/devices), ANS (health plans), CFM (medical practice).
Critical dynamics: fee-for-service vs value-based transition, AI diagnostics regulation,
telemedicine post-COVID legalization, LGPD healthcare data requirements,
chronic disease burden (diabetes, hypertension, cancer) driving demand.
Tech disruption: AI diagnostics, remote monitoring, precision medicine,
drug delivery innovation, hospital-at-home models.`,
		Queries: []string{
			"healthcare disruption Brazil ANS ANVISA regulation 2024 2025",
			"healthtech investment Brazil telemedicine AI diagnostics",
			"value-based care Brazil Hapvida NotreDame Rede D'Or",
			"digital health regulation LGPD patient data Brazil",
			"pharmaceutical market Brazil generic drugs patent cliff",
		},
	}
}
