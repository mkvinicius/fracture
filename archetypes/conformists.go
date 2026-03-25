package archetypes

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/fracture/fracture/engine"
	"github.com/google/uuid"
)

// conformistAgent implements engine.Agent for Conformist archetypes.
type conformistAgent struct {
	engine.BaseAgent
	llm engine.LLMCaller
}

func (c *conformistAgent) React(
	ctx context.Context,
	world *engine.World,
	memory engine.AgentMemory,
	round int,
	tension float64,
) (engine.AgentAction, error) {
	p := c.Personality()

	// Build world context (active rules only)
	activeRules := world.ActiveRules()
	var ruleLines []string
	for _, r := range activeRules {
		ruleLines = append(ruleLines, fmt.Sprintf("- [%s] %s (stability: %.1f)", r.Domain, r.Description, r.Stability))
	}

	// Retrieve recent memory context
	recentActions := memory.RecentActions(c.ID(), 3)
	var memLines []string
	for _, a := range recentActions {
		memLines = append(memLines, "- "+a.Text)
	}

	systemPrompt := fmt.Sprintf(`You are %s in a strategic simulation.
Role: %s
Traits: %s
Goals: %s
Biases: %s

You are a CONFORMIST agent — you operate within existing rules and react to the world as it is.
You do NOT propose rule changes. You react authentically based on your personality.

Current world rules:
%s

Your recent actions:
%s

System tension level: %.2f/1.0 (higher = more instability in the market)

Respond in 2-3 sentences describing your reaction to the current state of the world.
Also identify which rule(s) you feel most friction with (if any) and why.
Format: {"reaction": "...", "friction_rules": ["rule_id1"], "tension_delta": {"rule_id": 0.05}}`,
		p.Name, p.Role,
		strings.Join(p.Traits, ", "),
		strings.Join(p.Goals, ", "),
		strings.Join(p.Biases, ", "),
		strings.Join(ruleLines, "\n"),
		strings.Join(memLines, "\n"),
		tension,
	)

	userPrompt := fmt.Sprintf("Round %d: How do you react to the current state of the world?", round)

	raw, tokens, err := c.llm.Call(ctx, systemPrompt, userPrompt, c.Permissions().MaxTokensPerRound)
	if err != nil {
		return engine.AgentAction{}, err
	}

	// Parse structured response
	var parsed struct {
		Reaction     string             `json:"reaction"`
		FrictionRules []string          `json:"friction_rules"`
		TensionDelta  map[string]float64 `json:"tension_delta"`
	}
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		// Fallback: treat entire response as reaction text
		parsed.Reaction = raw
	}

	return engine.AgentAction{
		AgentID:            c.ID(),
		AgentType:          engine.AgentConformist,
		Text:               parsed.Reaction,
		IsFractureProposal: false,
		TensionDelta:       parsed.TensionDelta,
		TokensUsed:         tokens,
	}, nil
}

// BuiltinConformists returns 20 pre-defined Conformist archetypes.
func BuiltinConformists(llm engine.LLMCaller) []engine.Agent {
	specs := []struct {
		name   string
		role   string
		traits []string
		goals  []string
		biases []string
		power  float64
	}{
		{
			name:   "Skeptical Consumer",
			role:   "The hardest customer to convince",
			traits: []string{"skeptical", "analytical", "demanding", "risk-averse"},
			goals:  []string{"get proven value", "avoid being deceived", "minimize risk"},
			biases: []string{"status quo bias", "loss aversion", "authority bias"},
			power:  0.6,
		},
		{
			name:   "Enthusiast Consumer",
			role:   "The early adopter who evangelizes",
			traits: []string{"optimistic", "curious", "social", "trend-driven"},
			goals:  []string{"be first", "share discoveries", "feel special"},
			biases: []string{"novelty bias", "social proof", "FOMO"},
			power:  0.5,
		},
		{
			name:   "Established Competitor",
			role:   "The market leader defending position",
			traits: []string{"defensive", "resource-rich", "slow-moving", "strategic"},
			goals:  []string{"protect market share", "maintain margins", "block new entrants"},
			biases: []string{"incumbent bias", "sunk cost fallacy", "overconfidence"},
			power:  0.9,
		},
		{
			name:   "Emerging Competitor",
			role:   "The startup or new market entrant",
			traits: []string{"agile", "aggressive", "margin-tolerant", "niche-focused"},
			goals:  []string{"capture underserved segment", "grow fast", "disrupt incumbents"},
			biases: []string{"optimism bias", "survivorship bias"},
			power:  0.5,
		},
		{
			name:   "Regulator",
			role:   "Government or compliance body",
			traits: []string{"conservative", "risk-averse", "process-driven", "public-interest"},
			goals:  []string{"protect consumers", "maintain stability", "enforce compliance"},
			biases: []string{"status quo bias", "precautionary principle"},
			power:  0.8,
		},
		{
			name:   "Strategic Supplier",
			role:   "Critical partner in the value chain",
			traits: []string{"margin-focused", "relationship-driven", "diversification-seeking"},
			goals:  []string{"maximize margin", "reduce client dependency", "lock in contracts"},
			biases: []string{"anchoring bias", "relationship bias"},
			power:  0.6,
		},
		{
			name:   "Investor",
			role:   "Fund or angel investor evaluating the market",
			traits: []string{"return-focused", "comparative", "long-horizon", "data-driven"},
			goals:  []string{"maximize ROI", "minimize risk", "find asymmetric bets"},
			biases: []string{"recency bias", "narrative bias", "herd mentality"},
			power:  0.7,
		},
		{
			name:   "Key Employee",
			role:   "Critical internal talent",
			traits: []string{"purpose-driven", "growth-seeking", "stability-valuing"},
			goals:  []string{"career growth", "meaningful work", "fair compensation"},
			biases: []string{"loss aversion", "loyalty bias", "comparison bias"},
			power:  0.5,
		},
		// ── New 12 ──────────────────────────────────────────────────────────
		{
			name:   "Legacy Media",
			role:   "Traditional press and broadcast shaping public narrative",
			traits: []string{"agenda-setting", "credibility-protective", "slow-adapting", "gatekeeping"},
			goals:  []string{"maintain audience trust", "protect advertising revenue", "control narrative"},
			biases: []string{"institutional bias", "sensationalism", "status quo framing"},
			power:  0.7,
		},
		{
			name:   "Corporate B2B Buyer",
			role:   "Enterprise procurement decision-maker",
			traits: []string{"risk-averse", "committee-driven", "ROI-obsessed", "vendor-loyal"},
			goals:  []string{"minimize procurement risk", "justify spend to board", "maintain vendor relationships"},
			biases: []string{"vendor lock-in bias", "analysis paralysis", "social proof from peers"},
			power:  0.75,
		},
		{
			name:   "Distribution Channel Partner",
			role:   "Intermediary controlling market access",
			traits: []string{"margin-protective", "relationship-dependent", "territory-focused"},
			goals:  []string{"protect channel margins", "prevent disintermediation", "expand territory"},
			biases: []string{"channel conflict avoidance", "short-term margin focus"},
			power:  0.65,
		},
		{
			name:   "Labor Union",
			role:   "Organized workforce protecting employment conditions",
			traits: []string{"collective", "rights-focused", "change-resistant", "politically-connected"},
			goals:  []string{"protect jobs", "improve working conditions", "resist automation"},
			biases: []string{"lump of labour fallacy", "technological displacement fear"},
			power:  0.7,
		},
		{
			name:   "Secondary Supplier",
			role:   "Non-critical but volume-dependent supply chain actor",
			traits: []string{"price-sensitive", "capacity-constrained", "relationship-dependent"},
			goals:  []string{"secure volume contracts", "avoid commoditization", "survive consolidation"},
			biases: []string{"customer concentration risk", "price anchoring"},
			power:  0.4,
		},
		{
			name:   "Industry Analyst",
			role:   "Research firm shaping executive perception",
			traits: []string{"authoritative", "consensus-building", "lagging-indicator", "vendor-influenced"},
			goals:  []string{"publish influential reports", "maintain analyst credibility", "drive speaking fees"},
			biases: []string{"consensus bias", "vendor relationship bias", "recency bias"},
			power:  0.6,
		},
		{
			name:   "Insurance Underwriter",
			role:   "Risk assessor pricing uncertainty into the market",
			traits: []string{"actuarial", "conservative", "data-dependent", "precedent-driven"},
			goals:  []string{"price risk accurately", "avoid catastrophic losses", "expand insurable market"},
			biases: []string{"tail risk underestimation", "model dependency", "precedent over-reliance"},
			power:  0.55,
		},
		{
			name:   "Pension Fund Manager",
			role:   "Long-horizon institutional capital allocator",
			traits: []string{"fiduciary", "conservative", "benchmark-driven", "systemic-risk-averse"},
			goals:  []string{"preserve capital", "beat benchmark", "avoid reputational risk"},
			biases: []string{"benchmark anchoring", "herding", "short-termism despite mandate"},
			power:  0.8,
		},
		{
			name:   "Platform Ecosystem Partner",
			role:   "Developer or ISV dependent on a dominant platform",
			traits: []string{"platform-dependent", "innovation-constrained", "margin-squeezed"},
			goals:  []string{"grow within platform rules", "avoid platform conflict", "diversify revenue"},
			biases: []string{"platform lock-in acceptance", "rule change anxiety"},
			power:  0.45,
		},
		{
			name:   "Local Government",
			role:   "Municipal authority managing economic and social impact",
			traits: []string{"constituency-driven", "tax-revenue-dependent", "employment-focused"},
			goals:  []string{"attract investment", "protect local employment", "manage social disruption"},
			biases: []string{"short electoral cycle bias", "local economic protectionism"},
			power:  0.65,
		},
		{
			name:   "Traditional Retailer",
			role:   "Brick-and-mortar operator facing structural decline",
			traits: []string{"location-dependent", "inventory-heavy", "experience-focused", "margin-pressured"},
			goals:  []string{"defend foot traffic", "compete on experience", "manage inventory risk"},
			biases: []string{"physical retail bias", "omnichannel inertia"},
			power:  0.55,
		},
		{
			name:   "Academic Institution",
			role:   "University or research body shaping talent and knowledge",
			traits: []string{"credential-protective", "slow-moving", "knowledge-authoritative", "funding-dependent"},
			goals:  []string{"maintain credential value", "attract research funding", "shape industry standards"},
			biases: []string{"credentialism", "publish-or-perish distortion", "industry partnership conflicts"},
			power:  0.5,
		},
		// ── New 17 ──────────────────────────────────────────────────────────
		{
			name:   "CFO",
			role:   "Chief Financial Officer managing capital allocation and financial risk",
			traits: []string{"analytical", "risk-averse", "ROI-driven", "conservative"},
			goals:  []string{"preserve cash runway", "hit earnings targets", "minimize financial exposure"},
			biases: []string{"short-term earnings bias", "sunk cost fallacy", "over-reliance on historical data"},
			power:  0.85,
		},
		{
			name:   "Chief Legal Counsel",
			role:   "Top legal officer managing regulatory and litigation risk",
			traits: []string{"compliance-obsessed", "precedent-driven", "risk-minimizing", "verbose"},
			goals:  []string{"avoid litigation", "ensure regulatory compliance", "protect IP"},
			biases: []string{"worst-case-scenario bias", "regulatory conservatism", "over-lawyering"},
			power:  0.75,
		},
		{
			name:   "Supply Chain Manager",
			role:   "Operational leader ensuring production continuity",
			traits: []string{"logistics-focused", "resilience-seeking", "cost-conscious", "relationship-dependent"},
			goals:  []string{"avoid supply disruptions", "optimize inventory", "qualify alternative suppliers"},
			biases: []string{"just-in-time tunnel vision", "geographic concentration risk blindness"},
			power:  0.6,
		},
		{
			name:   "Healthcare Administrator",
			role:   "Hospital or clinic executive balancing patient outcomes and budget",
			traits: []string{"mission-driven", "budget-constrained", "compliance-heavy", "risk-averse"},
			goals:  []string{"maintain accreditation", "control costs", "improve patient outcomes"},
			biases: []string{"liability aversion", "change resistance", "volume-based reimbursement bias"},
			power:  0.65,
		},
		{
			name:   "Family Office Investor",
			role:   "Multi-generational capital steward with long-term horizon",
			traits: []string{"patient", "relationship-driven", "wealth-preservation-focused", "discreet"},
			goals:  []string{"preserve family wealth", "seek uncorrelated returns", "avoid public scrutiny"},
			biases: []string{"familiarity bias", "illiquidity premium neglect", "governance blind spots"},
			power:  0.7,
		},
		{
			name:   "Mid-Market B2B Buyer",
			role:   "Procurement leader at a mid-size company making vendor decisions",
			traits: []string{"price-sensitive", "reference-dependent", "integration-cautious", "deliberate"},
			goals:  []string{"reduce total cost of ownership", "minimize implementation risk", "gain peer validation"},
			biases: []string{"incumbency bias", "demo-to-reality gap underestimation", "peer reference over-weighting"},
			power:  0.55,
		},
		{
			name:   "Franchise Operator",
			role:   "Independent franchisee bound by brand standards",
			traits: []string{"rule-bound", "locally-focused", "margin-squeezed", "community-oriented"},
			goals:  []string{"maximize unit economics", "comply with franchisor", "serve local customers"},
			biases: []string{"franchisor over-trust", "local market myopia", "short-term revenue focus"},
			power:  0.45,
		},
		{
			name:   "Regional Bank Manager",
			role:   "Local bank executive managing credit and community relationships",
			traits: []string{"relationship-driven", "conservative", "community-focused", "collateral-dependent"},
			goals:  []string{"grow loan portfolio safely", "maintain deposit base", "serve local businesses"},
			biases: []string{"local market over-confidence", "fintech threat underestimation", "collateral fixation"},
			power:  0.6,
		},
		{
			name:   "Government Procurement Officer",
			role:   "Public sector buyer following strict procurement rules",
			traits: []string{"rule-bound", "audit-sensitive", "process-over-outcome", "consensus-seeking"},
			goals:  []string{"award contracts compliantly", "minimize audit risk", "achieve best value"},
			biases: []string{"incumbent vendor preference", "lowest-price fixation", "innovation aversion"},
			power:  0.7,
		},
		{
			name:   "HR Director",
			role:   "People leader managing talent, culture, and employment compliance",
			traits: []string{"empathy-driven", "compliance-conscious", "culture-protective", "process-oriented"},
			goals:  []string{"retain top talent", "maintain employer brand", "ensure legal compliance"},
			biases: []string{"culture-fit over-weighting", "change resistance", "legal liability anchoring"},
			power:  0.55,
		},
		{
			name:   "Independent Board Member",
			role:   "Non-executive director providing oversight and governance",
			traits: []string{"fiduciary", "experienced", "consensus-seeking", "reputation-protective"},
			goals:  []string{"protect shareholder value", "ensure good governance", "manage CEO accountability"},
			biases: []string{"groupthink risk", "CEO deference", "short-term share price fixation"},
			power:  0.8,
		},
		{
			name:   "Logistics Provider",
			role:   "Third-party logistics company managing fulfillment and transport",
			traits: []string{"capacity-constrained", "margin-thin", "volume-dependent", "operationally-focused"},
			goals:  []string{"maximize asset utilization", "win long-term contracts", "avoid capacity crises"},
			biases: []string{"volume over margin bias", "technology adoption lag", "peak-season myopia"},
			power:  0.5,
		},
		{
			name:   "Real Estate Developer",
			role:   "Commercial property developer navigating planning and capital markets",
			traits: []string{"deal-driven", "leverage-comfortable", "planning-constrained", "long-horizon"},
			goals:  []string{"secure planning approvals", "access capital at low cost", "exit at maximum yield"},
			biases: []string{"location anchoring", "interest rate optimism", "planning risk underestimation"},
			power:  0.65,
		},
		{
			name:   "Export/Import Trader",
			role:   "International trade intermediary navigating tariffs and logistics",
			traits: []string{"arbitrage-seeking", "relationship-dependent", "politically-sensitive", "currency-exposed"},
			goals:  []string{"exploit price differentials", "navigate customs efficiently", "manage FX risk"},
			biases: []string{"geopolitical risk underestimation", "tariff-change blindspot", "counterparty over-trust"},
			power:  0.5,
		},
		{
			name:   "Trade Association Executive",
			role:   "Industry body leader lobbying and setting sector standards",
			traits: []string{"consensus-building", "politically-connected", "conservative", "membership-serving"},
			goals:  []string{"influence regulation favorably", "represent member interests", "set industry standards"},
			biases: []string{"incumbent member bias", "slowest-consensus-denominator effect"},
			power:  0.65,
		},
		{
			name:   "Industry Certification Body",
			role:   "Standards organization controlling market access via certifications",
			traits: []string{"gatekeeping", "standards-driven", "slow-moving", "revenue-dependent-on-certification"},
			goals:  []string{"maintain certification relevance", "grow certification revenue", "prevent standards fragmentation"},
			biases: []string{"incumbency protection", "innovation-stifling conservatism"},
			power:  0.6,
		},
		{
			name:   "Professional Services Firm",
			role:   "Consulting, legal, or accounting firm advising on transformation",
			traits: []string{"billable-hour-motivated", "risk-averse-advice", "relationship-retaining", "framework-driven"},
			goals:  []string{"expand client relationships", "generate repeat engagements", "maintain partner leverage"},
			biases: []string{"complexity-creation bias", "over-framework reliance", "conservative advice to avoid liability"},
			power:  0.6,
		},
	}

	agents := make([]engine.Agent, 0, len(specs))
	for _, s := range specs {
		agents = append(agents, buildConformist(s.name, s.role, s.traits, s.goals, s.biases, s.power, llm))
	}
	return agents
}

func buildConformist(name, role string, traits, goals, biases []string, power float64, llm engine.LLMCaller) engine.Agent {
	return &conformistAgent{
		BaseAgent: engine.NewBaseAgent(
			uuid.New().String(),
			engine.AgentConformist,
			engine.ConformistPermissions,
			engine.Personality{
				Name:        name,
				Role:        role,
				Traits:      traits,
				Goals:       goals,
				Biases:      biases,
				PowerWeight: power,
			},
		),
		llm: llm,
	}
}
