package skills

import "github.com/fracture/fracture/engine"

// RetailSkill retorna a skill vertical de Retail & Consumer.
func RetailSkill() *Skill {
	return &Skill{
		ID:   "retail",
		Name: "Retail & Consumer",
		Description: "Simulação especializada para varejo, e-commerce, marketplace, " +
			"CPG e consumo.",
		Industries: []string{
			"varejo", "retail", "e-commerce", "ecommerce",
			"marketplace", "loja", "supermercado", "farmácia",
			"fashion", "moda", "consumer", "consumidor",
			"FMCG", "CPG", "alimentação", "food",
		},
		Rules: []*engine.Rule{
			{ID: "ret-001", Description: "Mercado Livre and Amazon dominate marketplace infrastructure in Brazil", Domain: engine.DomainMarket, Stability: 0.75, IsActive: true},
			{ID: "ret-002", Description: "Physical retail requires ANVISA compliance for food and pharma products", Domain: engine.DomainRegulation, Stability: 0.85, IsActive: true},
			{ID: "ret-003", Description: "PROCON and CDC protect consumers from abusive practices and price gouging", Domain: engine.DomainRegulation, Stability: 0.88, IsActive: true},
			{ID: "ret-004", Description: "Last-mile delivery is the primary competitive battleground in e-commerce", Domain: engine.DomainMarket, Stability: 0.55, IsActive: true},
			{ID: "ret-005", Description: "Installment payments (parcelamento) drive purchase decisions in Brazil", Domain: engine.DomainBehavior, Stability: 0.80, IsActive: true},
			{ID: "ret-006", Description: "Loyalty programs and cashback are table stakes for retail competition", Domain: engine.DomainMarket, Stability: 0.65, IsActive: true},
			{ID: "ret-007", Description: "Social commerce via Instagram and TikTok is growing faster than traditional e-commerce", Domain: engine.DomainCulture, Stability: 0.40, IsActive: true},
			{ID: "ret-008", Description: "Wholesale clubs (Atacadão, Sam's Club) are taking share from traditional retail", Domain: engine.DomainMarket, Stability: 0.60, IsActive: true},
			{ID: "ret-009", Description: "Product returns policy is a major competitive differentiator", Domain: engine.DomainMarket, Stability: 0.70, IsActive: true},
			{ID: "ret-010", Description: "Chinese cross-border e-commerce (Shein, Shopee, AliExpress) competes on price", Domain: engine.DomainGeopolitics, Stability: 0.45, IsActive: true},
			{ID: "ret-011", Description: "Omnichannel integration between physical and digital is competitive necessity", Domain: engine.DomainTechnology, Stability: 0.50, IsActive: true},
			{ID: "ret-012", Description: "Private label brands are taking share from national brands in all categories", Domain: engine.DomainMarket, Stability: 0.55, IsActive: true},
		},
		Agents: []SkillAgent{
			{
				Name:        "Marcos Galperin",
				Role:        "Marketplace & E-commerce Infrastructure Disruptor",
				Traits:      []string{"Mercado Livre founder", "marketplace model", "Mercado Pago", "logistics network", "fintech embedded in commerce"},
				Goals:       []string{"dominate Latin American commerce infrastructure", "embed payments in every transaction"},
				Biases:      []string{"vertical retail", "brand-only e-commerce", "logistics dependence"},
				Power:       0.95,
				IsDisruptor: true,
			},
			{
				Name:        "Luiza Trajano",
				Role:        "Physical Retail Champion & Omnichannel Pioneer",
				Traits:      []string{"Magazine Luiza", "Magalu ecosystem", "omnichannel", "social selling", "inclusion", "digital transformation of physical retail"},
				Goals:       []string{"prove physical retail can win in digital age", "serve Brazil's middle market"},
				Biases:      []string{"pure-play e-commerce", "marketplace commoditization", "premature physical store death"},
				Power:       0.90,
				IsDisruptor: false,
			},
			{
				Name:        "Maurício Skora",
				Role:        "D2C & Creator Commerce Disruptor",
				Traits:      []string{"direct to consumer", "social commerce", "creator economy", "TikTok shop", "brand without middleman"},
				Goals:       []string{"brands selling direct to fans", "eliminate distribution intermediary"},
				Biases:      []string{"marketplace dependency", "wholesale distribution"},
				Power:       0.75,
				IsDisruptor: true,
			},
			{
				Name:        "Abilio Diniz",
				Role:        "Supermarket & Food Retail Incumbent Guardian",
				Traits:      []string{"Pão de Açúcar", "food retail", "supermarket scale", "private label", "loyalty programs", "fresh food"},
				Goals:       []string{"defend supermarket format", "grow private label margin"},
				Biases:      []string{"delivery apps taking margin", "dark stores cannibalizing store traffic"},
				Power:       0.85,
				IsDisruptor: false,
			},
			{
				Name:        "Pedro Zemel",
				Role:        "Quick Commerce & Dark Store Disruptor",
				Traits:      []string{"10-minute delivery", "dark stores", "instant commerce", "Rappi", "iFood", "last-mile disruption"},
				Goals:       []string{"replace planned shopping with instant delivery", "own the last mile"},
				Biases:      []string{"planned shopping trips", "physical store dominance"},
				Power:       0.80,
				IsDisruptor: true,
			},
			{
				Name:        "Ralph Nader",
				Role:        "Consumer Protection Movement Founder",
				Traits:      []string{"Unsafe at Any Speed", "consumer rights movement", "corporate accountability", "product safety standards", "citizen advocacy", "regulatory capture prevention", "class action as consumer weapon"},
				Goals:       []string{"corporations accountable to consumers not just shareholders", "safety and transparency as non-negotiable"},
				Biases:      []string{"corporate self-regulation", "profit over consumer safety", "information asymmetry exploited against consumers"},
				Power:       0.88,
				IsDisruptor: true,
			},
		},
		Context: `RETAIL SECTOR CONTEXT FOR FRACTURE SIMULATION:
Brazil is the largest retail market in Latin America.
Key players: Mercado Livre (dominant marketplace), Magazine Luiza (omnichannel),
Americanas (restructuring), Via Varejo/Casas Bahia, Carrefour Brasil,
GPA/Pão de Açúcar, Atacadão/Grupo Carrefour, iFood (food delivery),
Shopee/Shein (cross-border), Riachuelo/Renner (fashion).
Critical dynamics: Chinese cross-border competition on price,
social commerce growth via TikTok/Instagram,
quick commerce (10-min delivery) disrupting traditional retail,
private label taking share from national brands,
installment payments driving purchases,
omnichannel as competitive necessity.`,
		Queries: []string{
			"varejo brasileiro e-commerce marketplace 2024 2025",
			"Mercado Livre Shopee Amazon Brasil competição",
			"social commerce TikTok Instagram Brasil vendas",
			"quick commerce dark store delivery Brasil iFood",
			"cross-border Shein Shopee regulação importação Brasil",
		},
	}
}
