package skills

import "github.com/fracture/fracture/engine"

// ConstructionSkill retorna a skill vertical de Construção Civil & PropTech.
func ConstructionSkill() *Skill {
	return &Skill{
		ID:          "construction",
		Name:        "Construção Civil & PropTech",
		Description: "Simulação especializada para incorporadoras, construtoras, proptechs e mercado imobiliário.",
		Industries: []string{
			"construção", "construction", "imobiliário", "real estate",
			"proptech", "incorporadora", "construtora", "imóvel",
			"apartamento", "loteamento", "retrofit", "BIM",
			"habitação", "aluguel", "fundo imobiliário", "FII",
		},

		Rules: []*engine.Rule{
			{ID: "con-001", Description: "CAIXA and SBPE control 70%+ of mortgage financing through FGTS rules", Domain: engine.DomainFinance, Stability: 0.82, IsActive: true},
			{ID: "con-002", Description: "Minha Casa Minha Vida determines demand for low-income housing", Domain: engine.DomainRegulation, Stability: 0.75, IsActive: true},
			{ID: "con-003", Description: "ABNT NBR norms and CREA/CAU regulate construction standards and professionals", Domain: engine.DomainRegulation, Stability: 0.88, IsActive: true},
			{ID: "con-004", Description: "Interest rates (SELIC) directly determine mortgage affordability and demand", Domain: engine.DomainFinance, Stability: 0.70, IsActive: true},
			{ID: "con-005", Description: "Land cost is the primary constraint in urban real estate development", Domain: engine.DomainMarket, Stability: 0.78, IsActive: true},
			{ID: "con-006", Description: "BIM (Building Information Modeling) is mandatory for public works above threshold", Domain: engine.DomainTechnology, Stability: 0.60, IsActive: true},
			{ID: "con-007", Description: "PropTech platforms are reducing time-to-sale but not disrupting construction itself", Domain: engine.DomainTechnology, Stability: 0.50, IsActive: true},
			{ID: "con-008", Description: "Modular and industrialized construction is reducing cost and timeline 30-40%", Domain: engine.DomainTechnology, Stability: 0.40, IsActive: true},
			{ID: "con-009", Description: "Environmental licensing delays are the primary bottleneck for large developments", Domain: engine.DomainRegulation, Stability: 0.65, IsActive: true},
			{ID: "con-010", Description: "FIIs (Real Estate Investment Trusts) democratized real estate investment in Brazil", Domain: engine.DomainFinance, Stability: 0.68, IsActive: true},
			{ID: "con-011", Description: "Remote work permanently changed office demand — vacancy rates rising in CBDs", Domain: engine.DomainBehavior, Stability: 0.55, IsActive: true},
			{ID: "con-012", Description: "Short-term rental (Airbnb) is restructuring residential and hospitality markets", Domain: engine.DomainMarket, Stability: 0.48, IsActive: true},
		},

		Agents: []SkillAgent{
			{
				Name:        "Elie Horn",
				Role:        "Mass Market Housing Champion & Cyrela Founder",
				Traits:      []string{"Cyrela", "mass market housing", "MCMV", "land bank strategy", "launch velocity"},
				Goals:       []string{"maximize launches in growing cities", "capture MCMV demand"},
				Biases:      []string{"interest rate exposure", "modular construction threatening margins"},
				Power:       0.90,
				IsDisruptor: false,
			},
			{
				Name:        "Eduardo Fischer",
				Role:        "PropTech Disruptor & Digital Real Estate Pioneer",
				Traits:      []string{"proptech", "digital mortgage", "iBuyer model", "data-driven pricing", "transaction speed"},
				Goals:       []string{"reduce real estate transaction friction", "data-driven property pricing"},
				Biases:      []string{"broker cartel", "manual appraisal process", "analog transaction flow"},
				Power:       0.80,
				IsDisruptor: true,
			},
			{
				Name:        "CREA Director",
				Role:        "Engineering & Architecture Professional Regulator",
				Traits:      []string{"CREA regulation", "professional responsibility", "ABNT norms", "ART technical responsibility note", "professional standards"},
				Goals:       []string{"ensure construction safety", "maintain professional standards"},
				Biases:      []string{"unregistered professionals", "BIM replacing traditional processes prematurely"},
				Power:       0.85,
				IsDisruptor: false,
			},
			{
				Name:        "Patrícia Pereira",
				Role:        "Modular Construction & Industrialization Pioneer",
				Traits:      []string{"industrialized construction", "modular systems", "offsite fabrication", "cost reduction", "speed to delivery"},
				Goals:       []string{"reduce construction time 40%", "industrialize what has been artisanal"},
				Biases:      []string{"traditional construction workforce resistance", "ABNT norm rigidity"},
				Power:       0.75,
				IsDisruptor: true,
			},
			{
				Name:        "Sergio Cano",
				Role:        "Affordable Housing & MCMV Policy Expert",
				Traits:      []string{"Minha Casa Minha Vida", "social housing", "CAIXA financing", "FGTS policy", "housing deficit"},
				Goals:       []string{"house Brazil 8M unit deficit", "maintain MCMV funding"},
				Biases:      []string{"market-rate housing crowding out affordable", "interest rates killing demand"},
				Power:       0.80,
				IsDisruptor: false,
			},
		},

		Context: `CONSTRUCTION SECTOR CONTEXT FOR FRACTURE SIMULATION:
Brazil's construction sector represents 6.5% of GDP.
Key players: MRV, Cyrela, EZTec, Tenda (residential),
Tegma, Andrade Gutierrez (infrastructure),
QuintoAndar, Loft (proptech), Zap Imóveis, VivaReal (portals).
Key regulators: CAIXA (FGTS/SBPE financing), CREA/CAU (professionals),
municipal prefeituras (zoning/licensing), IBAMA (environmental).
Critical dynamics: SELIC rate directly controls demand,
MCMV program is largest demand driver for lower income,
modular/industrialized construction growing fast,
office market disrupted by remote work,
short-term rental restructuring residential markets,
PropTechs reducing friction but not yet disrupting construction itself.`,

		Queries: []string{
			"mercado imobiliário Brasil proptech disruption 2024 2025",
			"construção modular industrializada Brasil crescimento",
			"MCMV Minha Casa Minha Vida financiamento CAIXA",
			"QuintoAndar Loft proptech Brazil real estate",
			"escritórios vacância remote work Brasil mercado corporativo",
		},
	}
}
