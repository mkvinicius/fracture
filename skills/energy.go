package skills

import "github.com/fracture/fracture/engine"

// EnergySkill retorna a skill vertical de Energia & Utilities.
func EnergySkill() *Skill {
	return &Skill{
		ID:          "energy",
		Name:        "Energia & Utilities",
		Description: "Simulação especializada para energia elétrica, solar, petróleo, gás, utilities e transição energética.",
		Industries: []string{
			"energia", "energy", "elétrica", "solar", "eólica",
			"petróleo", "oil", "gás", "gas", "utilities",
			"distribuidora", "geradora", "transmissora",
			"renovável", "renewable", "hidrelétrica", "biocombustível",
		},

		Rules: []*engine.Rule{
			{ID: "ene-001", Description: "ANEEL regulates electricity tariffs, concessions, and quality standards", Domain: engine.DomainRegulation, Stability: 0.88, IsActive: true},
			{ID: "ene-002", Description: "Brazil's energy matrix is 85% renewable — hydro dominance creates drought risk", Domain: engine.DomainGeopolitics, Stability: 0.72, IsActive: true},
			{ID: "ene-003", Description: "Distributed solar generation is growing exponentially — 20M+ units by 2026", Domain: engine.DomainTechnology, Stability: 0.40, IsActive: true},
			{ID: "ene-004", Description: "ANP regulates oil, gas, and biofuel production and distribution", Domain: engine.DomainRegulation, Stability: 0.85, IsActive: true},
			{ID: "ene-005", Description: "Free energy market (ACL) allows large consumers to choose suppliers directly", Domain: engine.DomainMarket, Stability: 0.65, IsActive: true},
			{ID: "ene-006", Description: "Energy transition requires R$200B+ in grid modernization by 2030", Domain: engine.DomainFinance, Stability: 0.55, IsActive: true},
			{ID: "ene-007", Description: "Petrobras pre-salt reserves make Brazil a major oil exporter despite energy transition", Domain: engine.DomainGeopolitics, Stability: 0.70, IsActive: true},
			{ID: "ene-008", Description: "Electric vehicle adoption will increase electricity demand 15-20% by 2030", Domain: engine.DomainTechnology, Stability: 0.45, IsActive: true},
			{ID: "ene-009", Description: "Hydrogen economy is emerging as new energy export opportunity for Brazil", Domain: engine.DomainTechnology, Stability: 0.30, IsActive: true},
			{ID: "ene-010", Description: "Energy poverty affects 3M+ Brazilian households disconnected from grid", Domain: engine.DomainGeopolitics, Stability: 0.68, IsActive: true},
			{ID: "ene-011", Description: "Carbon credits from renewable energy are creating new revenue streams", Domain: engine.DomainFinance, Stability: 0.38, IsActive: true},
			{ID: "ene-012", Description: "Battery storage is becoming economically viable — changing renewable dispatch economics", Domain: engine.DomainTechnology, Stability: 0.35, IsActive: true},
		},

		Agents: []SkillAgent{
			{
				Name:        "Jean-Paul Prates",
				Role:        "Petrobras & Oil Transition Strategist",
				Traits:      []string{"Petrobras", "pre-sal", "oil transition", "energy security", "just transition", "oil funding renewables"},
				Goals:       []string{"maximize Petrobras oil revenue while funding energy transition"},
				Biases:      []string{"premature oil exit", "stranded asset risk from over-investment"},
				Power:       0.90,
				IsDisruptor: false,
			},
			{
				Name:        "Roberto Wajsman",
				Role:        "Distributed Solar Disruptor & Energy Democratizer",
				Traits:      []string{"distributed generation", "solar rooftop", "energy cooperatives", "prosumer", "net metering", "democratize energy"},
				Goals:       []string{"solar on every rooftop", "eliminate dependence on distribution monopoly"},
				Biases:      []string{"distribution company monopoly", "ANEEL tariff penalizing self-generation"},
				Power:       0.80,
				IsDisruptor: true,
			},
			{
				Name:        "ANEEL Director",
				Role:        "Electricity Regulator & Tariff Guardian",
				Traits:      []string{"tariff regulation", "concession contracts", "quality standards", "cross-subsidy policy", "energy universalization"},
				Goals:       []string{"universal access", "tariff adequacy", "quality standards"},
				Biases:      []string{"cost shifts to captive consumers", "unregulated distributed generation growth"},
				Power:       0.92,
				IsDisruptor: false,
			},
			{
				Name:        "Rodrigo Limp",
				Role:        "Green Hydrogen & New Energy Economy Pioneer",
				Traits:      []string{"green hydrogen", "Petrobras renewables", "energy export", "electrolysis", "decarbonization"},
				Goals:       []string{"Brazil as green hydrogen exporter by 2030"},
				Biases:      []string{"fossil fuel dependency", "missing green hydrogen infrastructure investment"},
				Power:       0.75,
				IsDisruptor: true,
			},
		},

		Context: `ENERGY SECTOR CONTEXT FOR FRACTURE SIMULATION:
Brazil has the cleanest large-scale energy matrix in the world (85% renewable).
Key players: Petrobras (oil), Eletrobras (generation/transmission),
CPFL, Energisa, Equatorial (distribution), Eneva, CTEEP (transmission),
BYD, Growatt (solar equipment), AES Brasil (renewables).
Key regulators: ANEEL (electricity), ANP (oil/gas), MME (energy ministry),
IBAMA (environmental licensing for plants).
Critical dynamics: Distributed solar growing exponentially (20M units by 2026),
free energy market (ACL) opening to smaller consumers,
EV adoption increasing electricity demand,
green hydrogen as next export opportunity,
Petrobras pre-salt funding energy transition,
battery storage making renewables fully dispatchable.`,

		Queries: []string{
			"energia solar distribuída Brasil crescimento ANEEL 2024 2025",
			"transição energética Brasil petróleo renovável hidrogênio",
			"mercado livre energia ACL empresas Brasil",
			"Petrobras pré-sal transição energia verde",
			"armazenamento energia baterias Brasil grid modernização",
		},
	}
}
