package skills

import "github.com/fracture/fracture/engine"

// ManufacturingSkill retorna a skill vertical de Indústria & Manufatura.
func ManufacturingSkill() *Skill {
	return &Skill{
		ID:          "manufacturing",
		Name:        "Indústria & Manufatura",
		Description: "Simulação especializada para indústria, manufatura, Indústria 4.0, automação e cadeia produtiva.",
		Industries: []string{
			"indústria", "manufacturing", "manufatura", "fábrica",
			"industrial", "automação", "automation", "robótica",
			"IoT industrial", "indústria 4.0", "supply chain industrial",
			"metalmecânica", "química", "têxtil", "calçados",
		},

		Rules: []*engine.Rule{
			{ID: "mfg-001", Description: "ABDI and CNI represent industry interests in regulatory discussions", Domain: engine.DomainRegulation, Stability: 0.78, IsActive: true},
			{ID: "mfg-002", Description: "Custo Brasil (taxes, infrastructure, credit) adds 30%+ to production cost", Domain: engine.DomainFinance, Stability: 0.72, IsActive: true},
			{ID: "mfg-003", Description: "Reshoring from China is creating new manufacturing opportunities globally", Domain: engine.DomainGeopolitics, Stability: 0.50, IsActive: true},
			{ID: "mfg-004", Description: "Industry 4.0 (IoT, AI, robotics) is reducing labor content 40-60% in advanced manufacturing", Domain: engine.DomainTechnology, Stability: 0.42, IsActive: true},
			{ID: "mfg-005", Description: "Brazilian industry competes with Chinese imports on price in most categories", Domain: engine.DomainGeopolitics, Stability: 0.68, IsActive: true},
			{ID: "mfg-006", Description: "BNDES provides long-term industrial investment financing at subsidized rates", Domain: engine.DomainFinance, Stability: 0.75, IsActive: true},
			{ID: "mfg-007", Description: "Environmental compliance (CONAMA) adds cost but enables premium market access", Domain: engine.DomainRegulation, Stability: 0.70, IsActive: true},
			{ID: "mfg-008", Description: "Additive manufacturing (3D printing) is disrupting tooling and spare parts logistics", Domain: engine.DomainTechnology, Stability: 0.38, IsActive: true},
			{ID: "mfg-009", Description: "Energy cost is 8-15% of production cost for energy-intensive industries", Domain: engine.DomainFinance, Stability: 0.65, IsActive: true},
			{ID: "mfg-010", Description: "Labor productivity in Brazilian industry is 30% below OECD average", Domain: engine.DomainBehavior, Stability: 0.72, IsActive: true},
			{ID: "mfg-011", Description: "Predictive maintenance through IoT is reducing unplanned downtime 50%+", Domain: engine.DomainTechnology, Stability: 0.45, IsActive: true},
			{ID: "mfg-012", Description: "Circular economy regulations are emerging — waste as input is becoming mandatory", Domain: engine.DomainRegulation, Stability: 0.40, IsActive: true},
		},

		Agents: []SkillAgent{
			{
				Name:        "Paulo Skaf",
				Role:        "Brazilian Industry Champion & Custo Brasil Fighter",
				Traits:      []string{"FIESP", "custo Brasil", "industrial policy", "tax reform", "infrastructure investment", "industry competitiveness"},
				Goals:       []string{"reduce custo Brasil", "protect domestic industry from Chinese competition"},
				Biases:      []string{"import liberalization", "tax burden on industry", "infrastructure bottlenecks"},
				Power:       0.88,
				IsDisruptor: false,
			},
			{
				Name:        "Klaus Schwab",
				Role:        "Industry 4.0 & Fourth Industrial Revolution Architect",
				Traits:      []string{"World Economic Forum", "fourth industrial revolution", "stakeholder capitalism", "technology and society", "future of work"},
				Goals:       []string{"inclusive technological transformation", "stakeholder capitalism over shareholder primacy"},
				Biases:      []string{"technology without social impact consideration", "winner-take-all automation"},
				Power:       0.85,
				IsDisruptor: true,
			},
			{
				Name:        "Sergio Rial",
				Role:        "Industrial Restructuring & Capital Efficiency Expert",
				Traits:      []string{"industrial restructuring", "operational efficiency", "EBITDA optimization", "capex discipline", "return on invested capital"},
				Goals:       []string{"maximize industrial ROIC", "eliminate inefficiency"},
				Biases:      []string{"investment without return discipline", "capacity expansion at wrong cycle"},
				Power:       0.82,
				IsDisruptor: false,
			},
			{
				Name:        "Daniela Chiaretti",
				Role:        "Circular Economy & Industrial Sustainability Disruptor",
				Traits:      []string{"circular economy", "industrial ecology", "waste as input", "extended producer responsibility", "biomimicry"},
				Goals:       []string{"zero industrial waste", "circular material flows"},
				Biases:      []string{"linear take-make-waste model", "externalized environmental costs"},
				Power:       0.75,
				IsDisruptor: true,
			},
		},

		Context: `MANUFACTURING SECTOR CONTEXT FOR FRACTURE SIMULATION:
Brazil's industry represents 20% of GDP with 8M+ formal workers.
Key players: Embraer (aerospace), Gerdau, Usiminas (steel),
Braskem (petrochemicals), WEG (motors/energy), Randon (trailers),
Marcopolo (buses), Embraco/Nidec (compressors).
Key regulators: ABDI (industrial development), CNI (industry confederation),
BNDES (development bank), CONAMA (environmental), INMETRO (quality).
Critical dynamics: Custo Brasil making local production uncompetitive vs imports,
reshoring creating new opportunities in electronics and critical minerals,
Industry 4.0 reducing labor content in advanced manufacturing,
Chinese competition intensifying across all segments,
energy transition requiring industrial decarbonization,
additive manufacturing disrupting tooling and spare parts.`,

		Queries: []string{
			"indústria 4.0 Brasil automação robótica manufatura 2024 2025",
			"reshoring nearshoring Brasil oportunidade industrial China",
			"custo Brasil competitividade industrial reforma tributária",
			"WEG Embraer Gerdau competição mercado global",
			"economia circular manufatura sustentabilidade Brasil",
		},
	}
}
