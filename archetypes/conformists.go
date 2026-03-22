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

// newConformist is a factory helper.
func newConformist(name, role string, traits, goals, biases []string, power float64, llm engine.LLMCaller) *conformistAgent {
	return &conformistAgent{
		BaseAgent: engine.BaseAgent{},
		llm:       llm,
	}
}

// BuiltinConformists returns the 8 pre-defined Conformist archetypes.
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
