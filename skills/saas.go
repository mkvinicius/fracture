package skills

import "github.com/fracture/fracture/engine"

// SaaSSkill retorna a skill vertical de SaaS & Tech B2B.
func SaaSSkill() *Skill {
	return &Skill{
		ID:          "saas",
		Name:        "SaaS & Tech B2B",
		Description: "Simulação especializada para SaaS, software empresarial, plataformas B2B e startups de tecnologia.",
		Industries: []string{
			"saas", "software", "tecnologia", "startup",
			"B2B", "plataforma", "platform", "API", "cloud",
			"nuvem", "ERP", "CRM", "marketplace B2B",
			"developer tools", "devtools", "vertical software",
		},

		Rules: []*engine.Rule{
			{ID: "sas-001", Description: "Annual recurring revenue (ARR) is the primary valuation metric for SaaS", Domain: engine.DomainFinance, Stability: 0.80, IsActive: true},
			{ID: "sas-002", Description: "Net Revenue Retention above 120% is the gold standard for enterprise SaaS", Domain: engine.DomainMarket, Stability: 0.75, IsActive: true},
			{ID: "sas-003", Description: "AI is commoditizing software features that previously required months to build", Domain: engine.DomainTechnology, Stability: 0.35, IsActive: true},
			{ID: "sas-004", Description: "Enterprise sales cycles in Brazil average 6-18 months for tickets above R$100k", Domain: engine.DomainBehavior, Stability: 0.70, IsActive: true},
			{ID: "sas-005", Description: "LGPD compliance is mandatory for all B2B software handling personal data", Domain: engine.DomainRegulation, Stability: 0.82, IsActive: true},
			{ID: "sas-006", Description: "Vertical SaaS commands 3-5x higher multiples than horizontal SaaS", Domain: engine.DomainFinance, Stability: 0.68, IsActive: true},
			{ID: "sas-007", Description: "Open source AI models are threatening proprietary ML/AI product moats", Domain: engine.DomainTechnology, Stability: 0.38, IsActive: true},
			{ID: "sas-008", Description: "Product-led growth (PLG) is replacing top-down enterprise sales for SMB", Domain: engine.DomainMarket, Stability: 0.52, IsActive: true},
			{ID: "sas-009", Description: "Customer success determines churn — poor CS destroys NRR regardless of product quality", Domain: engine.DomainBehavior, Stability: 0.72, IsActive: true},
			{ID: "sas-010", Description: "Brazilian SaaS companies face US competitor pressure in enterprise segment", Domain: engine.DomainGeopolitics, Stability: 0.60, IsActive: true},
			{ID: "sas-011", Description: "AI coding tools (GitHub Copilot, Claude Code) are reducing development cost 30-60%", Domain: engine.DomainTechnology, Stability: 0.40, IsActive: true},
			{ID: "sas-012", Description: "Embedded payments and fintech features increase SaaS LTV 3-5x", Domain: engine.DomainFinance, Stability: 0.50, IsActive: true},
		},

		Agents: []SkillAgent{
			{
				Name:        "Ben Horowitz",
				Role:        "Hard Things About Hard Things & Wartime CEO Expert",
				Traits:      []string{"The Hard Thing About Hard Things", "wartime vs peacetime CEO", "people management in crisis", "nobody cares just bring solutions", "the struggle", "a16z founding partner", "product-market fit is not enough"},
				Goals:       []string{"build companies that survive the hard things", "management as craft not formula"},
				Biases:      []string{"management frameworks that ignore psychological reality", "peacetime thinking in wartime situations"},
				Power:       0.88,
				IsDisruptor: false,
			},
			{
				Name:        "Jason Lemkin",
				Role:        "SaaS Metrics & Enterprise Go-to-Market Expert",
				Traits:      []string{"SaaStr", "SaaS metrics", "ARR growth", "enterprise sales motion", "SDR/AE model", "CAC payback"},
				Goals:       []string{"predictable revenue growth", "efficient GTM"},
				Biases:      []string{"product-led only without enterprise motion", "ignoring unit economics"},
				Power:       0.88,
				IsDisruptor: false,
			},
			{
				Name:        "Henrique Dubugras",
				Role:        "Infrastructure Disruptor & Embedded Finance Pioneer",
				Traits:      []string{"Brex", "embedded finance", "software eating financial services", "developer-first", "API-first fintech"},
				Goals:       []string{"embed financial services in every B2B software"},
				Biases:      []string{"standalone financial products", "software without financial layer"},
				Power:       0.85,
				IsDisruptor: true,
			},
			{
				Name:        "Aaron Levie",
				Role:        "Enterprise Cloud & AI Transformation Champion",
				Traits:      []string{"Box", "enterprise AI", "cloud transformation", "content management", "AI workflow automation"},
				Goals:       []string{"AI transforms every enterprise workflow"},
				Biases:      []string{"on-premise legacy systems", "AI hype without ROI"},
				Power:       0.80,
				IsDisruptor: true,
			},
			{
				Name:        "David Sacks",
				Role:        "SaaS Metrics Purist & Capital Efficiency Guardian",
				Traits:      []string{"Craft Ventures", "magic number", "CAC payback", "Rule of 40", "capital efficiency", "burn multiple"},
				Goals:       []string{"efficient growth over growth at any cost"},
				Biases:      []string{"blitzscaling without unit economics", "vanity ARR growth"},
				Power:       0.82,
				IsDisruptor: false,
			},
			{
				Name:        "Geoffrey Moore",
				Role:        "Crossing the Chasm & Technology Adoption Expert",
				Traits:      []string{"Crossing the Chasm 1991", "the chasm between early adopters and early majority", "whole product solution", "bowling alley strategy", "Zone to Win", "Escape Velocity", "technology adoption life cycle"},
				Goals:       []string{"cross the chasm from early adopters to mainstream market", "whole product that early majority actually needs", "bowling alley — dominate one niche before expanding", "early adopter success is not mainstream traction — the chasm is real"},
				Biases:      []string{"early adopter success masquerading as mainstream traction", "feature product vs whole product", "ignoring the chasm until it is too late"},
				Power:       0.90,
				IsDisruptor: false,
			},
		},

		Context: `SAAS B2B SECTOR CONTEXT FOR FRACTURE SIMULATION:
Brazil has the largest B2B SaaS market in Latin America.
Key players: TOTVS (ERP leader), Linx (retail tech), Pipefy (workflow),
RD Station (marketing), Zendesk BR, Salesforce Brazil,
Conta Azul, Omie (SMB accounting), QuintoAndar (proptech).
Key dynamics: AI coding tools reducing development cost 30-60%,
vertical SaaS commanding premium multiples,
open source AI threatening proprietary ML features,
US competitors (Salesforce, SAP, Oracle) competing for enterprise,
embedded fintech features becoming competitive necessity,
PLG replacing top-down sales for SMB segment,
LGPD compliance mandatory for all data-handling products.`,

		Queries: []string{
			"SaaS B2B Brasil startups tecnologia disruption 2024 2025",
			"AI coding tools impact software development cost Brazil",
			"vertical SaaS Brazil NRR metrics enterprise growth",
			"TOTVS ERP Brasil competição cloud migration",
			"embedded finance SaaS Brazil product expansion",
		},
	}
}
