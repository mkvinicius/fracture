package skills

import "github.com/fracture/fracture/engine"

// TourismSkill retorna a skill vertical de Turismo & Hospitalidade.
func TourismSkill() *Skill {
	return &Skill{
		ID:          "tourism",
		Name:        "Turismo & Hospitalidade",
		Description: "Simulação especializada para hotelaria, turismo, agências de viagem, companhias aéreas e experiências.",
		Industries: []string{
			"turismo", "tourism", "hotel", "hospitalidade", "hospitality",
			"viagem", "travel", "companhia aérea", "airline",
			"agência", "cruzeiro", "resort", "pousada",
			"OTA", "airbnb", "experiência", "destination",
		},

		Rules: []*engine.Rule{
			{ID: "tur-001", Description: "ANAC regulates aviation safety and market entry for airlines", Domain: engine.DomainRegulation, Stability: 0.88, IsActive: true},
			{ID: "tur-002", Description: "Embratur and state tourism boards promote Brazil internationally", Domain: engine.DomainRegulation, Stability: 0.70, IsActive: true},
			{ID: "tur-003", Description: "OTAs (Booking.com, Expedia, Decolar) control 60%+ of online hotel bookings", Domain: engine.DomainMarket, Stability: 0.68, IsActive: true},
			{ID: "tur-004", Description: "Airbnb has 5M+ listings in Brazil — disrupting hotel occupancy in leisure markets", Domain: engine.DomainMarket, Stability: 0.55, IsActive: true},
			{ID: "tur-005", Description: "LATAM and Gol dominate Brazilian aviation — duopoly pricing concerns", Domain: engine.DomainMarket, Stability: 0.65, IsActive: true},
			{ID: "tur-006", Description: "Visa requirements significantly impact international tourism flows to Brazil", Domain: engine.DomainGeopolitics, Stability: 0.60, IsActive: true},
			{ID: "tur-007", Description: "Sustainability and ecotourism are growing premium segments", Domain: engine.DomainCulture, Stability: 0.48, IsActive: true},
			{ID: "tur-008", Description: "AI travel assistants are replacing traditional travel agents for leisure travel", Domain: engine.DomainTechnology, Stability: 0.38, IsActive: true},
			{ID: "tur-009", Description: "Currency devaluation makes Brazil affordable for international visitors", Domain: engine.DomainFinance, Stability: 0.55, IsActive: true},
			{ID: "tur-010", Description: "Experience economy — travelers pay premium for unique local experiences over standard accommodation", Domain: engine.DomainCulture, Stability: 0.50, IsActive: true},
		},

		Agents: []SkillAgent{
			{
				Name:        "Herb Kelleher",
				Role:        "Low-Cost Aviation Pioneer & Southwest Airlines Founder",
				Traits:      []string{"Southwest Airlines", "low-cost carrier model", "point-to-point routes", "employees first customers second", "fun as competitive strategy", "no assigned seats", "quick turnaround"},
				Goals:       []string{"democratize air travel for everyone", "prove low-cost can coexist with great culture"},
				Biases:      []string{"hub-and-spoke complexity", "legacy carrier cost structure", "unions as adversaries rather than partners"},
				Power:       0.88,
				IsDisruptor: true,
			},
			{
				Name:        "Rui Chammas",
				Role:        "Brazilian Aviation & LATAM Network Strategist",
				Traits:      []string{"LATAM Airlines", "network strategy", "hub-and-spoke", "loyalty program", "fleet management", "yield management"},
				Goals:       []string{"maintain route network profitability", "loyalty program stickiness"},
				Biases:      []string{"low-cost carrier price wars", "OTA disintermediation of direct bookings"},
				Power:       0.85,
				IsDisruptor: false,
			},
			{
				Name:        "Juan Trippe",
				Role:        "Mass Aviation Pioneer & Pan Am Founder",
				Traits:      []string{"Pan American Airways", "democratize air travel", "Boeing 747 the jumbo jet", "international route pioneer", "aviation as instrument of peace", "mass travel creating global understanding"},
				Goals:       []string{"air travel for everyone not just the wealthy", "aviation connecting cultures and creating peace"},
				Biases:      []string{"aviation as luxury for the few", "protectionist route restrictions"},
				Power:       0.85,
				IsDisruptor: true,
			},
			{
				Name:        "Guilherme Paulus",
				Role:        "Brazilian Tourism Champion & CVC Founder",
				Traits:      []string{"CVC", "packaged tourism", "Brazilian traveler", "domestic tourism", "MICE market"},
				Goals:       []string{"package travel for Brazilian middle class", "domestic destination development"},
				Biases:      []string{"OTA disintermediation", "AI replacing travel agents"},
				Power:       0.80,
				IsDisruptor: false,
			},
			{
				Name:        "Chip Conley",
				Role:        "Boutique Hospitality & Experience Economy Pioneer",
				Traits:      []string{"boutique hotels", "experience economy", "emotional economy", "Maslow meets hospitality", "wisdom at work"},
				Goals:       []string{"premium experiential travel over commodity accommodation"},
				Biases:      []string{"standardization killing hospitality soul", "OTA commoditization"},
				Power:       0.75,
				IsDisruptor: true,
			},
		},

		Context: `TOURISM & HOSPITALITY CONTEXT FOR FRACTURE SIMULATION:
Brazil received 6.6M international tourists in 2023 — below potential.
Key players: LATAM, Gol (airlines), CVC, Decolar (OTAs/agencies),
Booking.com, Airbnb (accommodation platforms),
Rede Hotéis Deville, Pestana (hotel chains),
Embratur (promotion), ANAC (aviation regulation).
Key dynamics: OTAs controlling 60%+ of hotel bookings creating margin pressure,
Airbnb disrupting leisure hotel occupancy in beach/ecotourism markets,
visa-free policy driving international arrivals growth,
AI travel planners threatening traditional agency model,
experience economy driving premium segments,
sustainability and ecotourism growing as premium differentiator.`,

		Queries: []string{
			"turismo Brasil crescimento pós-COVID viagem 2024 2025",
			"Airbnb hotéis Brasil impacto ocupação disrupção",
			"LATAM Gol aviação Brasil competição low-cost",
			"ecoturismo sustentável Brasil premium segmento crescimento",
			"agência viagem OTA digital Brasil transformação",
		},
	}
}
