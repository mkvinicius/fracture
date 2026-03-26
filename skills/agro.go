package skills

import "github.com/fracture/fracture/engine"

// AgroSkill retorna a skill vertical de Agro & AgriTech.
func AgroSkill() *Skill {
	return &Skill{
		ID:          "agro",
		Name:        "Agro & AgriTech",
		Description: "Simulação especializada para agronegócio, commodities, agritechs e cadeias do agro.",
		Industries: []string{
			"agro", "agronegócio", "agricultura", "agritech",
			"fazenda", "soja", "milho", "cana", "pecuária",
			"commodity", "commodities", "rural", "cooperativa",
			"fertilizante", "defensivo", "irrigação", "colheita",
		},

		Rules: []*engine.Rule{
			{ID: "agr-001", Description: "Brazil controls 30%+ of global soy and coffee exports — weather risk is systemic", Domain: engine.DomainGeopolitics, Stability: 0.75, IsActive: true},
			{ID: "agr-002", Description: "MAPA (Ministério da Agricultura) regulates pesticides, seeds, and animal health", Domain: engine.DomainRegulation, Stability: 0.85, IsActive: true},
			{ID: "agr-003", Description: "Commodity prices are set globally — Brazilian producers are price takers not makers", Domain: engine.DomainFinance, Stability: 0.80, IsActive: true},
			{ID: "agr-004", Description: "Rural credit (PRONAF, PRONAMP) is subsidized and politically sensitive", Domain: engine.DomainFinance, Stability: 0.72, IsActive: true},
			{ID: "agr-005", Description: "ESG and deforestation pressure threatens Brazilian agro market access in EU", Domain: engine.DomainGeopolitics, Stability: 0.55, IsActive: true},
			{ID: "agr-006", Description: "Precision agriculture (drones, sensors, AI) is reducing input costs 20-40%", Domain: engine.DomainTechnology, Stability: 0.45, IsActive: true},
			{ID: "agr-007", Description: "Cooperatives (COAMO, Aurora, Cocamar) control significant processing and distribution", Domain: engine.DomainMarket, Stability: 0.70, IsActive: true},
			{ID: "agr-008", Description: "Bioinputs are replacing traditional pesticides in premium and export markets", Domain: engine.DomainTechnology, Stability: 0.40, IsActive: true},
			{ID: "agr-009", Description: "Land concentration creates structural inequality and political tension", Domain: engine.DomainGeopolitics, Stability: 0.65, IsActive: true},
			{ID: "agr-010", Description: "Traceability and carbon credit markets are emerging as new revenue streams", Domain: engine.DomainFinance, Stability: 0.35, IsActive: true},
			{ID: "agr-011", Description: "China is Brazil's largest agricultural customer — geopolitical risk is concentrated", Domain: engine.DomainGeopolitics, Stability: 0.60, IsActive: true},
			{ID: "agr-012", Description: "Vertical farming and controlled environment agriculture challenge open-field economics", Domain: engine.DomainTechnology, Stability: 0.30, IsActive: true},
		},

		Agents: []SkillAgent{
			{
				Name:        "Marcos Jank",
				Role:        "Brazilian Agribusiness Strategist & Global Trade Expert",
				Traits:      []string{"agronegócio global", "market access", "EU deforestation regulation", "commodity geopolitics", "Brazil as food supplier"},
				Goals:       []string{"protect market access for Brazilian agro", "navigate ESG pressure"},
				Biases:      []string{"EU protectionism disguised as ESG", "deforestation narrative weaponized against Brazil"},
				Power:       0.90,
				IsDisruptor: false,
			},
			{
				Name:        "Vandana Shiva",
				Role:        "Agricultural Biodiversity & Food Sovereignty Champion",
				Traits:      []string{"Staying Alive", "Monocultures of the Mind", "seed sovereignty", "biodiversity over monoculture", "traditional knowledge protection", "Navdanya movement", "food as commons not commodity"},
				Goals:       []string{"seed sovereignty for farmers", "biodiversity as agricultural resilience"},
				Biases:      []string{"GMO monocultures", "corporate seed patents", "Green Revolution as ecological disaster"},
				Power:       0.85,
				IsDisruptor: true,
			},
			{
				Name:        "Blairo Maggi",
				Role:        "Large-Scale Soy Producer & Commodity Champion",
				Traits:      []string{"soy king", "Mato Grosso", "large-scale farming", "commodity exports", "infrastructure logistics"},
				Goals:       []string{"maximize commodity export volume", "reduce logistics cost to port"},
				Biases:      []string{"ESG restrictions on farming practices", "EU market access restrictions"},
				Power:       0.85,
				IsDisruptor: false,
			},
			{
				Name:        "Raj Patel",
				Role:        "Food Systems Economist & The Value of Nothing Author",
				Traits:      []string{"Stuffed and Starved", "The Value of Nothing", "food systems as political economy", "true cost accounting", "food sovereignty", "corporate food system critique", "agroecology"},
				Goals:       []string{"true cost of food including environmental and social externalities", "food systems that nourish people and planet"},
				Biases:      []string{"commodity food systems externalizing true costs", "corporate control of food supply chains"},
				Power:       0.80,
				IsDisruptor: true,
			},
			{
				Name:        "Norman Borlaug",
				Role:        "Green Revolution Father & Nobel Peace Prize Agricultural Scientist",
				Traits:      []string{"Nobel Peace Prize 1970", "saved 1 billion lives", "high-yield wheat varieties", "Green Revolution", "science over ideology in agriculture", "food security as peace", "yield per hectare obsession"},
				Goals:       []string{"feed the world through science", "eliminate famine through agricultural productivity"},
				Biases:      []string{"ideology blocking agricultural science", "romanticizing low-yield traditional farming", "anti-GMO sentiment without scientific basis"},
				Power:       0.95,
				IsDisruptor: false,
			},
			{
				Name:        "Gordon Conway",
				Role:        "Doubly Green Revolution & Sustainable Agriculture Synthesizer",
				Traits:      []string{"One Billion Hungry", "doubly green revolution", "productivity AND sustainability", "agroecological intensification", "climate-smart agriculture", "rejecting false choice between yield and environment"},
				Goals:       []string{"feed the world without destroying the planet", "doubly green — more productive AND more sustainable", "agroecological intensification — use ecology to drive productivity", "reject the false choice: high yield and environmental sustainability are compatible"},
				Biases:      []string{"false choice between productivity and sustainability", "purely industrial agriculture ignoring ecological limits", "purely organic ignoring yield requirements"},
				Power:       0.85,
				IsDisruptor: true,
			},
		},

		Context: `AGRO SECTOR CONTEXT FOR FRACTURE SIMULATION:
Brazil is the world's largest exporter of soy, coffee, beef, sugar, and orange juice.
Agro represents 25%+ of Brazilian GDP and 50%+ of exports.
Key players: JBS, Marfrig, BRF (protein), COFCO, Cargill, ADM, Bunge (trading),
Embrapa (research), COAMO, Aurora (cooperatives).
Key regulators: MAPA (agriculture ministry), IBAMA (environmental),
CONAB (supply monitoring), ANA (water resources).
Critical dynamics: EU Deforestation Regulation threatening market access,
China concentration risk (60%+ of soy exports),
precision agriculture reducing input costs 20-40%,
carbon credit markets emerging,
bioinputs replacing traditional pesticides,
ESG pressure from global supply chains.`,

		Queries: []string{
			"agronegócio Brasil tecnologia agritech disruption 2024 2025",
			"EU deforestation regulation Brazil soy market access",
			"precision agriculture drones AI Brazil crop management",
			"carbon credits agro Brazil traceability ESG",
			"China Brazil soy geopolitical risk commodity exports",
		},
	}
}
