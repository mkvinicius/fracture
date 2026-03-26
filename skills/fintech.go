package skills

import "github.com/fracture/fracture/engine"

// FintechSkill retorna a skill vertical de Fintech & Financial Services.
func FintechSkill() *Skill {
	return &Skill{
		ID:   "fintech",
		Name: "Fintech & Financial Services",
		Description: "Simulação especializada para fintechs, bancos, pagamentos, crédito, " +
			"seguros e open finance.",
		Industries: []string{
			"fintech", "banco", "bank", "pagamento", "payment",
			"crédito", "credit", "seguro", "insurance", "investimento",
			"investment", "open finance", "open banking", "crypto",
			"criptomoeda", "carteira digital", "wallet",
		},
		Rules: []*engine.Rule{
			{ID: "fin-001", Description: "BACEN (Central Bank) regulates all financial institutions and payment systems", Domain: engine.DomainRegulation, Stability: 0.95, IsActive: true},
			{ID: "fin-002", Description: "PIX instant payment system is free and mandatory for banks above threshold", Domain: engine.DomainTechnology, Stability: 0.90, IsActive: true},
			{ID: "fin-003", Description: "Open Finance Brazil requires banks to share customer data via API with consent", Domain: engine.DomainRegulation, Stability: 0.75, IsActive: true},
			{ID: "fin-004", Description: "Credit scoring (Serasa/SPC) controls access to credit for most Brazilians", Domain: engine.DomainMarket, Stability: 0.70, IsActive: true},
			{ID: "fin-005", Description: "Banking spread in Brazil is among the highest in the world", Domain: engine.DomainFinance, Stability: 0.55, IsActive: true},
			{ID: "fin-006", Description: "CVM regulates capital markets and investment products", Domain: engine.DomainRegulation, Stability: 0.88, IsActive: true},
			{ID: "fin-007", Description: "SUSEP regulates insurance market and InsurTech operations", Domain: engine.DomainRegulation, Stability: 0.82, IsActive: true},
			{ID: "fin-008", Description: "Crypto assets are regulated but not as legal tender in Brazil", Domain: engine.DomainRegulation, Stability: 0.60, IsActive: true},
			{ID: "fin-009", Description: "LGPD requires explicit consent for financial data processing", Domain: engine.DomainRegulation, Stability: 0.80, IsActive: true},
			{ID: "fin-010", Description: "Nubank model proved digital-only bank can scale to 90M+ customers", Domain: engine.DomainMarket, Stability: 0.40, IsActive: true},
			{ID: "fin-011", Description: "Traditional banks control corporate and SME credit relationships", Domain: engine.DomainMarket, Stability: 0.65, IsActive: true},
			{ID: "fin-012", Description: "DREX (digital real) is being developed by BACEN for programmable money", Domain: engine.DomainTechnology, Stability: 0.35, IsActive: true},
		},
		Agents: []SkillAgent{
			{
				Name:        "David Vélez",
				Role:        "Digital Bank Disruptor & Financial Inclusion Champion",
				Traits:      []string{"Nubank founder", "purple card", "no fees", "customer obsession", "financial inclusion", "90M customers"},
				Goals:       []string{"eliminate banking fees", "democratize financial access"},
				Biases:      []string{"incumbent bank complexity", "fee extraction", "bad UX as business model"},
				Power:       0.95,
				IsDisruptor: true,
			},
			{
				Name:        "Eduardo Muszkat",
				Role:        "Brazilian Open Finance & BACEN Regulation Expert",
				Traits:      []string{"open banking architect", "BACEN regulation", "API standardization", "data portability", "financial ecosystem"},
				Goals:       []string{"data portability", "competition through open infrastructure"},
				Biases:      []string{"incumbent data hoarding", "closed financial systems"},
				Power:       0.85,
				IsDisruptor: true,
			},
			{
				Name:        "André Esteves",
				Role:        "Investment Bank & Capital Markets Guardian",
				Traits:      []string{"BTG Pactual", "investment banking", "capital markets", "structured products", "trading", "wealth management"},
				Goals:       []string{"protect high-margin investment banking", "sophisticated product differentiation"},
				Biases:      []string{"commoditization of financial services", "race to zero fees"},
				Power:       0.90,
				IsDisruptor: false,
			},
			{
				Name:        "Sebastián Mejía",
				Role:        "Embedded Finance & BNPL Disruptor",
				Traits:      []string{"Rappi", "embedded finance", "super app", "BNPL", "financial services as feature", "latam fintech"},
				Goals:       []string{"embed financial services in everyday apps", "capture commerce and finance together"},
				Biases:      []string{"standalone financial products", "branch-based banking"},
				Power:       0.80,
				IsDisruptor: true,
			},
			{
				Name:        "Arminio Fraga",
				Role:        "Brazilian Monetary Policy Architect & Central Banking Pioneer",
				Traits:      []string{"Banco Central do Brasil president 1999-2002", "inflation targeting regime architect", "Real Plan legacy", "Gávea Investimentos", "emerging market monetary policy", "credibility as central bank asset", "exchange rate flexibility"},
				Goals:       []string{"price stability as foundation of economic development", "central bank credibility above all else"},
				Biases:      []string{"fiscal dominance over monetary policy", "political interference in central bank", "inflation tolerance"},
				Power:       0.95,
				IsDisruptor: false,
			},
			{
				Name:        "Hernando de Soto",
				Role:        "Financial Inclusion & Property Rights Economist",
				Traits:      []string{"The Mystery of Capital", "dead capital", "property rights for the poor", "informal economy", "legal empowerment", "blockchain as property registry", "financial inclusion through formalization"},
				Goals:       []string{"bring informal economy assets into formal financial system", "property rights as foundation of economic development"},
				Biases:      []string{"financial systems that exclude informal workers", "crypto speculation over real financial inclusion"},
				Power:       0.85,
				IsDisruptor: true,
			},
			{
				Name:        "Muhammad Yunus",
				Role:        "Microfinance Pioneer & Social Business Creator",
				Traits:      []string{"Grameen Bank", "Nobel Peace Prize 2006", "the poor are creditworthy", "social business", "microcredit for women", "poverty as system failure not personal failure", "bank for the poor"},
				Goals:       []string{"financial services as human right not privilege", "credit for the poor without collateral", "Grameen model — peer accountability replacing collateral requirement", "social business — companies that solve problems without profit motive"},
				Biases:      []string{"collateral requirements excluding the poor", "profit maximization over social impact in finance", "banking for the already wealthy"},
				Power:       0.88,
				IsDisruptor: true,
			},
		},
		Context: `FINTECH SECTOR CONTEXT FOR FRACTURE SIMULATION:
Brazil is the largest fintech market in Latin America.
Key players: Nubank (90M+ customers, largest neobank globally),
Itaú (largest private bank), Bradesco, Santander Brasil, BTG Pactual,
Mercado Pago, PicPay, C6 Bank, Inter.
Key regulators: BACEN (central bank, PIX, open finance), CVM (capital markets),
SUSEP (insurance), PGFN (tax).
Critical dynamics: PIX displaced card payments in P2P,
Open Finance Phase 4 enabling credit portability,
DREX (digital real) programmable money pilot,
banking spread among world's highest creating fintech opportunity,
Nubank model proving digital-only is viable at massive scale.`,
		Queries: []string{
			"fintech Brazil BACEN regulation open finance 2024 2025",
			"Nubank digital bank competition incumbents Brazil",
			"PIX open banking disruption traditional banks Brazil",
			"DREX digital real BACEN programmable money",
			"crypto regulation Brazil CVM BACEN 2025",
		},
	}
}
