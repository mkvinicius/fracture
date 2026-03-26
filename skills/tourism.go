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
				Name:        "Brian Chesky",
				Role:        "Hospitality Disruptor & Belonging Economy Champion",
				Traits:      []string{"Airbnb", "belong anywhere", "host economy", "experience marketplace", "home sharing", "local experiences"},
				Goals:       []string{"every home a potential accommodation", "experiences over hotels"},
				Biases:      []string{"standardized hotel experience", "OTA commoditization of hospitality"},
				Power:       0.90,
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
				Name:        "ANAC Director",
				Role:        "Aviation Safety & Market Regulator",
				Traits:      []string{"ANAC", "aviation safety", "route authorization", "slot allocation", "RBAC compliance"},
				Goals:       []string{"safety above all", "market competition within safety bounds"},
				Biases:      []string{"unsafe operators", "slot hoarding by dominant carriers"},
				Power:       0.90,
				IsDisruptor: false,
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
