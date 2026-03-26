package skills

import "github.com/fracture/fracture/engine"

// LegalSkill retorna a skill vertical de Legal & LegalTech.
func LegalSkill() *Skill {
	return &Skill{
		ID:   "legal",
		Name: "Legal & LegalTech",
		Description: "Simulação especializada para escritórios de advocacia, legaltech, " +
			"compliance e serviços jurídicos.",
		Industries: []string{
			"jurídico", "legal", "advocacia", "legaltech",
			"compliance", "escritório", "law firm", "tribunal",
			"contencioso", "regulatório", "LGPD", "trabalhista",
		},
		Rules: []*engine.Rule{
			{ID: "leg-001", Description: "OAB regulates legal practice and prohibits non-lawyer ownership of law firms", Domain: engine.DomainRegulation, Stability: 0.90, IsActive: true},
			{ID: "leg-002", Description: "Attorney-client privilege is constitutionally protected", Domain: engine.DomainRegulation, Stability: 0.95, IsActive: true},
			{ID: "leg-003", Description: "CNJ mandates digital processes through PJe system in federal courts", Domain: engine.DomainTechnology, Stability: 0.80, IsActive: true},
			{ID: "leg-004", Description: "Legal fees are partially regulated by OAB tables but negotiable for large clients", Domain: engine.DomainFinance, Stability: 0.65, IsActive: true},
			{ID: "leg-005", Description: "LGPD compliance requires legal counsel for data governance programs", Domain: engine.DomainRegulation, Stability: 0.82, IsActive: true},
			{ID: "leg-006", Description: "Brazil has 1.3M lawyers — highest per capita ratio in the world", Domain: engine.DomainMarket, Stability: 0.75, IsActive: true},
			{ID: "leg-007", Description: "AI-generated legal documents require attorney review and signature", Domain: engine.DomainRegulation, Stability: 0.70, IsActive: true},
			{ID: "leg-008", Description: "Arbitration is growing as alternative to slow judicial system", Domain: engine.DomainMarket, Stability: 0.55, IsActive: true},
			{ID: "leg-009", Description: "Legal market is fragmented — top 10 firms represent less than 5% of revenue", Domain: engine.DomainMarket, Stability: 0.70, IsActive: true},
			{ID: "leg-010", Description: "ESG compliance and sustainability reporting are creating new legal demand", Domain: engine.DomainRegulation, Stability: 0.45, IsActive: true},
		},
		Agents: []SkillAgent{
			{
				Name:        "Daniel Kessler",
				Role:        "LegalTech Disruptor & Access to Justice Champion",
				Traits:      []string{"legaltech", "document automation", "AI contracts", "access to justice", "democratize legal services"},
				Goals:       []string{"make legal services affordable", "automate routine legal work"},
				Biases:      []string{"OAB protectionism", "billable hour model", "legal complexity as business model"},
				Power:       0.80,
				IsDisruptor: true,
			},
			{
				Name:        "Modesto Carvalhosa",
				Role:        "Corporate Law & Traditional Practice Guardian",
				Traits:      []string{"corporate law", "M&A", "governance", "traditional practice", "relationship-based legal services", "OAB ethics"},
				Goals:       []string{"maintain quality of legal practice", "protect attorney-client relationship"},
				Biases:      []string{"commoditization of legal work", "AI replacing judgment", "non-lawyer ownership"},
				Power:       0.85,
				IsDisruptor: false,
			},
			{
				Name:        "Sobral Pinto",
				Role:        "Brazilian Legal Ethics & Human Rights Champion",
				Traits:      []string{"defender of political prisoners", "legal ethics above client interest", "OAB independence", "rule of law over political power", "courage to defend the unpopular", "law as instrument of justice not power", "professional independence"},
				Goals:       []string{"law must serve justice not power", "attorney independence from political pressure"},
				Biases:      []string{"law as instrument of oppression", "political interference in judiciary", "commercial interests over legal ethics"},
				Power:       0.95,
				IsDisruptor: false,
			},
			{
				Name:        "Sérgio Bermudes",
				Role:        "Litigation & Arbitration Specialist",
				Traits:      []string{"litigation strategy", "arbitration", "supreme court practice", "procedural mastery", "relationship with judiciary"},
				Goals:       []string{"win cases for major clients", "maintain judiciary relationships"},
				Biases:      []string{"alternative dispute resolution eating litigation fees", "digital court commoditization"},
				Power:       0.90,
				IsDisruptor: false,
			},
		},
		Context: `LEGAL SECTOR CONTEXT FOR FRACTURE SIMULATION:
Brazil has 1.3M lawyers — the highest per capita globally.
Key players: Pinheiro Neto, Machado Meyer, TozziniFreire, Demarest,
Trench Rossi Watanabe, Lefosse (major corporate firms).
Key regulators: OAB (professional regulation), CNJ (court administration),
STJ/STF (superior courts setting precedent).
Critical dynamics: AI contract review tools threatening associate work,
PJe digital court system creating data for legal AI,
LGPD creating compliance demand,
arbitration growing as alternative to 10-year litigation,
legaltech funding increasing in Brazil,
OAB resistance to non-lawyer ownership blocking international law firm entry.`,
		Queries: []string{
			"legaltech Brazil OAB regulation AI contracts 2024 2025",
			"inteligência artificial direito advocacia Brasil impacto",
			"LGPD compliance legal market Brazil demand",
			"arbitragem crescimento Brasil CARF CAM-CCBC",
			"escritório advocacia transformação digital Brasil",
		},
	}
}
