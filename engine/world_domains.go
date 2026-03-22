package engine

import "fmt"

// DefaultWorldForDomain creates a World pre-populated with domain-specific rules.
// The question and context are used to inject additional tension seeds.
func DefaultWorldForDomain(domain RuleDomain, question, context string) *World {
	var rules []*Rule

	switch domain {
	case DomainTechnology:
		rules = technologyRules()
	case DomainRegulation:
		rules = regulationRules()
	case DomainBehavior:
		rules = behaviorRules()
	case DomainCulture:
		rules = cultureRules()
	default: // market + anything else
		rules = marketRules()
	}

	// Add context rule if provided
	if question != "" {
		rules = append(rules, &Rule{
			ID:          "context-seed",
			Description: fmt.Sprintf("Current situation under analysis: %s", question),
			Domain:      domain,
			Stability:   0.3, // low stability — this is the thing being questioned
			IsActive:    true,
		})
	}

	return NewWorld(rules)
}

func marketRules() []*Rule {
	return []*Rule{
		{ID: "mkt-001", Description: "Customers pay before receiving value (subscription or purchase)", Domain: DomainMarket, Stability: 0.75, IsActive: true},
		{ID: "mkt-002", Description: "Companies compete primarily on price and product quality", Domain: DomainMarket, Stability: 0.70, IsActive: true},
		{ID: "mkt-003", Description: "Brand trust is built over years through consistent delivery", Domain: DomainMarket, Stability: 0.80, IsActive: true},
		{ID: "mkt-004", Description: "Distribution channels are controlled by established intermediaries", Domain: DomainMarket, Stability: 0.55, IsActive: true},
		{ID: "mkt-005", Description: "Customer data belongs to the company that collects it", Domain: DomainMarket, Stability: 0.45, IsActive: true},
		{ID: "mkt-006", Description: "Marketing spend scales linearly with customer acquisition", Domain: DomainMarket, Stability: 0.60, IsActive: true},
		{ID: "mkt-007", Description: "Switching costs protect incumbent market positions", Domain: DomainMarket, Stability: 0.65, IsActive: true},
		{ID: "mkt-008", Description: "Network effects favor the largest player in a category", Domain: DomainMarket, Stability: 0.72, IsActive: true},
	}
}

func technologyRules() []*Rule {
	return []*Rule{
		{ID: "tech-001", Description: "Software requires human developers to build and maintain", Domain: DomainTechnology, Stability: 0.40, IsActive: true},
		{ID: "tech-002", Description: "Data processing happens in centralized cloud infrastructure", Domain: DomainTechnology, Stability: 0.55, IsActive: true},
		{ID: "tech-003", Description: "AI models require large datasets and significant compute to train", Domain: DomainTechnology, Stability: 0.50, IsActive: true},
		{ID: "tech-004", Description: "Security requires ongoing human oversight and patching", Domain: DomainTechnology, Stability: 0.65, IsActive: true},
		{ID: "tech-005", Description: "Integration between systems requires custom development work", Domain: DomainTechnology, Stability: 0.45, IsActive: true},
		{ID: "tech-006", Description: "Technology adoption follows a predictable S-curve", Domain: DomainTechnology, Stability: 0.60, IsActive: true},
	}
}

func regulationRules() []*Rule {
	return []*Rule{
		{ID: "reg-001", Description: "Regulatory approval is required before market entry in regulated sectors", Domain: DomainRegulation, Stability: 0.85, IsActive: true},
		{ID: "reg-002", Description: "Compliance costs are borne by the regulated entity", Domain: DomainRegulation, Stability: 0.80, IsActive: true},
		{ID: "reg-003", Description: "Data must be stored within national borders (data sovereignty)", Domain: DomainRegulation, Stability: 0.60, IsActive: true},
		{ID: "reg-004", Description: "Consumer protection laws limit pricing and contract terms", Domain: DomainRegulation, Stability: 0.75, IsActive: true},
		{ID: "reg-005", Description: "Antitrust enforcement prevents monopolistic consolidation", Domain: DomainRegulation, Stability: 0.55, IsActive: true},
	}
}

func behaviorRules() []*Rule {
	return []*Rule{
		{ID: "beh-001", Description: "Employees expect stable employment and career progression", Domain: DomainBehavior, Stability: 0.55, IsActive: true},
		{ID: "beh-002", Description: "Talent is acquired through traditional hiring processes", Domain: DomainBehavior, Stability: 0.50, IsActive: true},
		{ID: "beh-003", Description: "Productivity is measured by hours worked and output volume", Domain: DomainBehavior, Stability: 0.45, IsActive: true},
		{ID: "beh-004", Description: "Management hierarchies determine decision-making authority", Domain: DomainBehavior, Stability: 0.60, IsActive: true},
		{ID: "beh-005", Description: "Company culture is shaped by in-person interactions", Domain: DomainBehavior, Stability: 0.40, IsActive: true},
		{ID: "beh-006", Description: "Compensation is primarily salary-based with annual reviews", Domain: DomainBehavior, Stability: 0.65, IsActive: true},
	}
}

func cultureRules() []*Rule {
	return []*Rule{
		{ID: "cul-001", Description: "Consumer preferences evolve gradually over years", Domain: DomainCulture, Stability: 0.60, IsActive: true},
		{ID: "cul-002", Description: "Social media shapes brand perception in real-time", Domain: DomainCulture, Stability: 0.35, IsActive: true},
		{ID: "cul-003", Description: "Generational differences drive distinct consumption patterns", Domain: DomainCulture, Stability: 0.70, IsActive: true},
		{ID: "cul-004", Description: "Environmental and social values increasingly influence purchase decisions", Domain: DomainCulture, Stability: 0.45, IsActive: true},
		{ID: "cul-005", Description: "Trust in institutions and established brands is declining", Domain: DomainCulture, Stability: 0.40, IsActive: true},
	}
}
