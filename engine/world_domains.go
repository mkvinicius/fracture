package engine

import (
	"context"
	"fmt"
	"math"
)

// stabilityModifier returns the per-domain factor used to reduce rule stability
// when research confidence indicates that domain rules are under pressure.
func stabilityModifier(domain RuleDomain) float64 {
	switch domain {
	case DomainTechnology:
		return 0.30
	case DomainCulture:
		return 0.35
	case DomainRegulation:
		return 0.10
	case DomainMarket:
		return 0.20
	case DomainBehavior:
		return 0.25
	case DomainGeopolitics:
		return 0.20
	case DomainFinance:
		return 0.15
	default:
		return 0.20
	}
}

// DefaultWorldForDomainWithContext builds a domain world enriched with
// DeepSearch findings. It respects context cancellation, stores research
// context as Evidence (not as a voting Rule), and applies a confidence-weighted
// stability reduction to rules identified as affected by the research.
func DefaultWorldForDomainWithContext(
	ctx context.Context,
	domain RuleDomain,
	question string,
	extraContext string,
	affectedRules []string,
	confidence float64,
) (*World, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	world := DefaultWorldForDomain(domain, question, "")

	// Remove context-seed rule — context is stored in Evidence instead so it
	// does not participate in tension voting.
	delete(world.Rules, "context-seed")
	delete(world.TensionMap, "context-seed")

	// Store domain context as Evidence (read-only field, not a Rule).
	world.Evidence = extraContext

	// Apply confidence-weighted stability pressure to the affected rules.
	mod := stabilityModifier(domain)
	for _, ruleID := range affectedRules {
		if r, ok := world.Rules[ruleID]; ok {
			adjusted := r.Stability * (1.0 - confidence*mod)
			r.Stability = math.Max(0.05, math.Min(0.95, adjusted))
		}
	}

	return world, nil
}

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
	case DomainGeopolitics:
		rules = geopoliticsRules()
	case DomainFinance:
		rules = financeRules()
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
		{ID: "mkt-009", Description: "Pricing power belongs to whoever controls the bottleneck in the value chain", Domain: DomainMarket, Stability: 0.68, IsActive: true},
		{ID: "mkt-010", Description: "Market share is a lagging indicator of competitive position", Domain: DomainMarket, Stability: 0.62, IsActive: true},
		{ID: "mkt-011", Description: "Category leaders set the pricing anchor for the entire market", Domain: DomainMarket, Stability: 0.70, IsActive: true},
		{ID: "mkt-012", Description: "Customer acquisition cost rises as a market matures", Domain: DomainMarket, Stability: 0.65, IsActive: true},
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
		{ID: "tech-007", Description: "Proprietary data moats create sustainable competitive advantage", Domain: DomainTechnology, Stability: 0.55, IsActive: true},
		{ID: "tech-008", Description: "Hardware and software are developed and sold separately", Domain: DomainTechnology, Stability: 0.50, IsActive: true},
		{ID: "tech-009", Description: "Open source commoditizes yesterday's proprietary technology", Domain: DomainTechnology, Stability: 0.42, IsActive: true},
		{ID: "tech-010", Description: "Edge computing reduces latency but increases operational complexity", Domain: DomainTechnology, Stability: 0.58, IsActive: true},
	}
}

func regulationRules() []*Rule {
	return []*Rule{
		{ID: "reg-001", Description: "Regulatory approval is required before market entry in regulated sectors", Domain: DomainRegulation, Stability: 0.85, IsActive: true},
		{ID: "reg-002", Description: "Compliance costs are borne by the regulated entity", Domain: DomainRegulation, Stability: 0.80, IsActive: true},
		{ID: "reg-003", Description: "Data must be stored within national borders (data sovereignty)", Domain: DomainRegulation, Stability: 0.60, IsActive: true},
		{ID: "reg-004", Description: "Consumer protection laws limit pricing and contract terms", Domain: DomainRegulation, Stability: 0.75, IsActive: true},
		{ID: "reg-005", Description: "Antitrust enforcement prevents monopolistic consolidation", Domain: DomainRegulation, Stability: 0.55, IsActive: true},
		{ID: "reg-006", Description: "AI systems must be auditable and explainable to regulators", Domain: DomainRegulation, Stability: 0.45, IsActive: true},
		{ID: "reg-007", Description: "Environmental disclosure is voluntary for private companies", Domain: DomainRegulation, Stability: 0.40, IsActive: true},
		{ID: "reg-008", Description: "Regulatory sandboxes allow limited market testing without full compliance", Domain: DomainRegulation, Stability: 0.50, IsActive: true},
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
		{ID: "beh-007", Description: "Remote work is a benefit, not the default mode of operation", Domain: DomainBehavior, Stability: 0.38, IsActive: true},
		{ID: "beh-008", Description: "Knowledge workers are expected to specialize in a single domain", Domain: DomainBehavior, Stability: 0.52, IsActive: true},
		{ID: "beh-009", Description: "Leadership is earned through tenure and internal promotion", Domain: DomainBehavior, Stability: 0.58, IsActive: true},
	}
}

func cultureRules() []*Rule {
	return []*Rule{
		{ID: "cul-001", Description: "Consumer preferences evolve gradually over years", Domain: DomainCulture, Stability: 0.60, IsActive: true},
		{ID: "cul-002", Description: "Social media shapes brand perception in real-time", Domain: DomainCulture, Stability: 0.35, IsActive: true},
		{ID: "cul-003", Description: "Generational differences drive distinct consumption patterns", Domain: DomainCulture, Stability: 0.70, IsActive: true},
		{ID: "cul-004", Description: "Environmental and social values increasingly influence purchase decisions", Domain: DomainCulture, Stability: 0.45, IsActive: true},
		{ID: "cul-005", Description: "Trust in institutions and established brands is declining", Domain: DomainCulture, Stability: 0.40, IsActive: true},
		{ID: "cul-006", Description: "Creator economy enables individuals to build audiences that rival media companies", Domain: DomainCulture, Stability: 0.38, IsActive: true},
		{ID: "cul-007", Description: "Authenticity and transparency are rewarded more than polish and perfection", Domain: DomainCulture, Stability: 0.42, IsActive: true},
		{ID: "cul-008", Description: "Community belonging drives purchasing decisions as much as product features", Domain: DomainCulture, Stability: 0.48, IsActive: true},
	}
}

func geopoliticsRules() []*Rule {
	return []*Rule{
		{ID: "geo-001", Description: "Free trade agreements reduce tariffs and enable global supply chains", Domain: DomainGeopolitics, Stability: 0.50, IsActive: true},
		{ID: "geo-002", Description: "Technology exports are controlled by national security considerations", Domain: DomainGeopolitics, Stability: 0.60, IsActive: true},
		{ID: "geo-003", Description: "Geopolitical alliances shape which markets companies can access", Domain: DomainGeopolitics, Stability: 0.65, IsActive: true},
		{ID: "geo-004", Description: "Supply chain resilience requires geographic diversification", Domain: DomainGeopolitics, Stability: 0.45, IsActive: true},
		{ID: "geo-005", Description: "Economic sanctions can cut off access to critical markets overnight", Domain: DomainGeopolitics, Stability: 0.55, IsActive: true},
		{ID: "geo-006", Description: "National champions receive state support to compete globally", Domain: DomainGeopolitics, Stability: 0.58, IsActive: true},
		{ID: "geo-007", Description: "Digital sovereignty laws require local data infrastructure", Domain: DomainGeopolitics, Stability: 0.48, IsActive: true},
		{ID: "geo-008", Description: "Currency exchange risk is a cost of doing business internationally", Domain: DomainGeopolitics, Stability: 0.70, IsActive: true},
	}
}

func financeRules() []*Rule {
	return []*Rule{
		{ID: "fin-001", Description: "Capital allocation follows risk-adjusted return expectations", Domain: DomainFinance, Stability: 0.75, IsActive: true},
		{ID: "fin-002", Description: "Equity dilution is the price of venture capital growth", Domain: DomainFinance, Stability: 0.70, IsActive: true},
		{ID: "fin-003", Description: "Profitability is required for long-term business survival", Domain: DomainFinance, Stability: 0.72, IsActive: true},
		{ID: "fin-004", Description: "Interest rates set by central banks determine the cost of capital", Domain: DomainFinance, Stability: 0.80, IsActive: true},
		{ID: "fin-005", Description: "IPOs and M&A are the primary exit paths for private investors", Domain: DomainFinance, Stability: 0.65, IsActive: true},
		{ID: "fin-006", Description: "ESG criteria increasingly influence institutional capital allocation", Domain: DomainFinance, Stability: 0.45, IsActive: true},
		{ID: "fin-007", Description: "Tokenization enables fractional ownership of previously illiquid assets", Domain: DomainFinance, Stability: 0.35, IsActive: true},
		{ID: "fin-008", Description: "Revenue multiples compress as growth slows and rates rise", Domain: DomainFinance, Stability: 0.60, IsActive: true},
	}
}

