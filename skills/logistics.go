package skills

import "github.com/fracture/fracture/engine"

// LogisticsSkill retorna a skill vertical de Logística & Supply Chain.
func LogisticsSkill() *Skill {
	return &Skill{
		ID:          "logistics",
		Name:        "Logística & Supply Chain",
		Description: "Simulação especializada para logística, transportadoras, supply chain, last-mile e fretes.",
		Industries: []string{
			"logística", "logistics", "supply chain", "transporte",
			"frete", "armazém", "warehouse", "last-mile",
			"entrega", "delivery", "porto", "ferrovia",
			"caminhão", "shipping", "fulfillment", "3PL",
		},

		Rules: []*engine.Rule{
			{ID: "log-001", Description: "Road transport carries 65% of Brazilian freight — rail and waterway underutilized", Domain: engine.DomainMarket, Stability: 0.78, IsActive: true},
			{ID: "log-002", Description: "ANTT regulates road freight and sets minimum freight table (tabelamento)", Domain: engine.DomainRegulation, Stability: 0.80, IsActive: true},
			{ID: "log-003", Description: "Port congestion at Santos is the primary bottleneck for agricultural exports", Domain: engine.DomainGeopolitics, Stability: 0.65, IsActive: true},
			{ID: "log-004", Description: "Fuel cost represents 35-40% of road freight operating cost", Domain: engine.DomainFinance, Stability: 0.60, IsActive: true},
			{ID: "log-005", Description: "Last-mile delivery is the most expensive and competitive segment", Domain: engine.DomainMarket, Stability: 0.55, IsActive: true},
			{ID: "log-006", Description: "Gig economy drivers (motoboys) power urban delivery but face regulation pressure", Domain: engine.DomainBehavior, Stability: 0.50, IsActive: true},
			{ID: "log-007", Description: "Electric vehicles are reducing last-mile delivery cost in urban centers", Domain: engine.DomainTechnology, Stability: 0.38, IsActive: true},
			{ID: "log-008", Description: "Autonomous trucks are 5-10 years from commercial deployment in Brazil", Domain: engine.DomainTechnology, Stability: 0.30, IsActive: true},
			{ID: "log-009", Description: "Cold chain logistics is a critical bottleneck for pharma and food sectors", Domain: engine.DomainMarket, Stability: 0.68, IsActive: true},
			{ID: "log-010", Description: "Cross-docking and dark stores are replacing traditional distribution centers", Domain: engine.DomainTechnology, Stability: 0.42, IsActive: true},
			{ID: "log-011", Description: "Freight marketplace platforms are fragmenting the traditional carrier market", Domain: engine.DomainMarket, Stability: 0.45, IsActive: true},
			{ID: "log-012", Description: "Supply chain visibility and real-time tracking are table stakes for enterprise clients", Domain: engine.DomainTechnology, Stability: 0.55, IsActive: true},
		},

		Agents: []SkillAgent{
			{
				Name:        "Fábio Schvartsman",
				Role:        "Infrastructure & Port Logistics Champion",
				Traits:      []string{"port infrastructure", "multimodal logistics", "Santos port", "rail investment", "bulk commodity logistics"},
				Goals:       []string{"reduce logistics cost from farm to port", "infrastructure investment"},
				Biases:      []string{"road-only solutions", "regulatory barriers to rail expansion"},
				Power:       0.85,
				IsDisruptor: false,
			},
			{
				Name:        "Hau Lee",
				Role:        "Supply Chain Resilience & Triple-A Supply Chain Creator",
				Traits:      []string{"Triple-A Supply Chain", "agile adaptable aligned", "bullwhip effect", "supply chain resilience", "demand uncertainty", "Stanford supply chain forum", "HBR most influential supply chain article"},
				Goals:       []string{"supply chains that are agile adaptable and aligned", "resilience over efficiency in uncertain environments"},
				Biases:      []string{"pure efficiency over resilience", "just-in-time without buffer for disruption", "supply chain optimization ignoring demand uncertainty"},
				Power:       0.90,
				IsDisruptor: false,
			},
			{
				Name:        "Edward Glaeser",
				Role:        "Urban Economics & Infrastructure Investment Expert",
				Traits:      []string{"Triumph of the City", "urban agglomeration economics", "infrastructure investment ROI", "density as productivity", "transit vs road investment", "housing affordability", "city competitiveness"},
				Goals:       []string{"cities and infrastructure that maximize human productivity", "evidence-based infrastructure investment"},
				Biases:      []string{"infrastructure investment without demand analysis", "sprawl over density", "roads over transit in dense areas"},
				Power:       0.85,
				IsDisruptor: false,
			},
			{
				Name:        "Cristina Palmaka",
				Role:        "Last-Mile Innovation & Urban Delivery Disruptor",
				Traits:      []string{"last-mile optimization", "urban delivery", "electric bikes", "micro-fulfillment", "dark stores", "same-day delivery"},
				Goals:       []string{"reduce last-mile cost below R$5 per delivery", "carbon-neutral urban delivery"},
				Biases:      []string{"traditional distribution center model", "diesel delivery fleet"},
				Power:       0.75,
				IsDisruptor: true,
			},
			{
				Name:        "Martin Christopher",
				Role:        "Logistics & Supply Chain Management Discipline Founder",
				Traits:      []string{"Logistics and Supply Chain Management", "competing through superior supply chain", "time compression", "relationship-based supply chain", "supply chain vulnerability", "Cranfield School of Management", "supply chain as competitive weapon"},
				Goals:       []string{"supply chain as primary competitive differentiator", "time compression as market advantage"},
				Biases:      []string{"logistics as cost center not strategic asset", "adversarial supplier relationships"},
				Power:       0.85,
				IsDisruptor: false,
			},
		},

		Context: `LOGISTICS SECTOR CONTEXT FOR FRACTURE SIMULATION:
Brazil's logistics cost represents 12% of GDP — among the highest globally (Brazil Cost).
Key players: JSL, Localfrio (3PL), Correios (postal), Jadlog, Total Express (parcel),
iFood, Rappi (food delivery), CargoX, TruckPad (freight marketplace),
Porto de Santos (largest port in Latin America).
Key regulators: ANTT (road/rail), ANTAQ (waterway/port), ANAC (air cargo),
ANVISA (pharma/food cold chain).
Critical dynamics: ANTT minimum freight table limiting price competition,
gig economy motoboys powering urban delivery under regulatory pressure,
freight marketplaces fragmenting traditional carrier market,
last-mile as primary e-commerce competitive battleground,
port congestion adding 3-5 days to export cycles,
electric vehicle adoption reducing urban delivery cost.`,

		Queries: []string{
			"logística Brasil supply chain disruption tecnologia 2024 2025",
			"last-mile delivery Brasil iFood Rappi marketplace frete",
			"tabelamento frete ANTT digital marketplace logística",
			"porto Santos congestionamento infraestrutura Brasil",
			"veículo elétrico entrega urbana Brasil last-mile",
		},
	}
}
