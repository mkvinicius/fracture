package archetypes

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"

	"github.com/fracture/fracture/engine"
	"github.com/google/uuid"
)

// disruptorAgent implements engine.Agent for Disruptor archetypes.
type disruptorAgent struct {
	engine.BaseAgent
	llm               engine.LLMCaller
	personalityFactor float64 // multiplier for fracture threshold (0.5-1.5)
}

func (d *disruptorAgent) React(
	ctx context.Context,
	world *engine.World,
	memory engine.AgentMemory,
	round int,
	tension float64,
) (engine.AgentAction, error) {
	p := d.Personality()

	activeRules := world.ActiveRules()
	var ruleLines []string
	for _, r := range activeRules {
		ruleLines = append(ruleLines, fmt.Sprintf(`{"id":"%s","description":"%s","domain":"%s","stability":%.2f}`,
			r.ID, r.Description, r.Domain, r.Stability))
	}

	// Most tense rules — prime candidates for disruption
	tenseRules := world.MostTenseRules(3)
	var tenseLines []string
	for _, r := range tenseRules {
		tenseLines = append(tenseLines, fmt.Sprintf("- %s (tension: %.2f)", r.Description, world.TensionMap[r.ID]))
	}

	// Determine if this agent should fire a FRACTURE POINT
	dissatisfaction := tension * d.personalityFactor
	threshold := engine.FractureThreshold(tension, dissatisfaction, d.personalityFactor)
	shouldFracture := engine.ShouldFireFracture(threshold)

	var fractureInstruction string
	if shouldFracture && d.Permissions().CanProposeRule {
		fractureInstruction = `
IMPORTANT: Based on the current tension level, you MUST propose a FRACTURE POINT.
Choose one of the most tense rules and propose how it could be fundamentally rewritten.
Your proposal should be bold, specific, and grounded in real-world precedent.
Set "fracture_proposal" to true and fill in the proposal fields.`
	} else {
		fractureInstruction = `Observe the world and describe what you see as the biggest opportunity for disruption.
Do NOT propose a formal rule change this round. Set "fracture_proposal" to false.`
	}

	systemPrompt := fmt.Sprintf(`You are %s in a strategic disruption simulation.
Role: %s
Traits: %s
Goals: %s
Biases: %s

You are a DISRUPTOR agent — you look for opportunities to rewrite the rules of the game.
You think like a founder, a regulator with a new mandate, or a movement leader.

Current world rules (JSON):
[%s]

Most tense rules (highest disruption potential):
%s

System tension: %.2f/1.0
%s

Respond in JSON format:
{
  "observation": "What you observe about the current state",
  "fracture_proposal": true/false,
  "proposed_rule_id": "id of rule to change (if fracture_proposal is true)",
  "new_description": "New version of the rule (if fracture_proposal is true)",
  "new_domain": "market|technology|regulation|behavior|culture",
  "new_stability": 0.0-1.0,
  "rationale": "Why this change would happen and who benefits",
  "tension_delta": {"rule_id": 0.05}
}`,
		p.Name, p.Role,
		strings.Join(p.Traits, ", "),
		strings.Join(p.Goals, ", "),
		strings.Join(p.Biases, ", "),
		strings.Join(ruleLines, ","),
		strings.Join(tenseLines, "\n"),
		tension,
		fractureInstruction,
	)

	userPrompt := fmt.Sprintf("Round %d: What do you observe, and what opportunity for disruption do you see?", round)

	raw, tokens, err := d.llm.Call(ctx, systemPrompt, userPrompt, d.Permissions().MaxTokensPerRound)
	if err != nil {
		return engine.AgentAction{}, err
	}

	var parsed struct {
		Observation      string             `json:"observation"`
		FractureProposal bool               `json:"fracture_proposal"`
		ProposedRuleID   string             `json:"proposed_rule_id"`
		NewDescription   string             `json:"new_description"`
		NewDomain        engine.RuleDomain  `json:"new_domain"`
		NewStability     float64            `json:"new_stability"`
		Rationale        string             `json:"rationale"`
		TensionDelta     map[string]float64 `json:"tension_delta"`
	}
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		parsed.Observation = raw
		parsed.FractureProposal = false
	}

	action := engine.AgentAction{
		AgentID:            d.ID(),
		AgentType:          engine.AgentDisruptor,
		Text:               parsed.Observation,
		IsFractureProposal: parsed.FractureProposal && parsed.ProposedRuleID != "",
		TensionDelta:       parsed.TensionDelta,
		TokensUsed:         tokens,
	}

	if action.IsFractureProposal {
		action.Proposal = &engine.RuleProposal{
			OriginalRuleID:  parsed.ProposedRuleID,
			NewDescription:  parsed.NewDescription,
			NewDomain:       parsed.NewDomain,
			NewStability:    parsed.NewStability,
			Rationale:       parsed.Rationale,
			ProposedByAgent: d.ID(),
		}
	}

	return action, nil
}

// BuiltinDisruptors returns 12 pre-defined Disruptor archetypes.
func BuiltinDisruptors(llm engine.LLMCaller) []engine.Agent {
	specs := []struct {
		name              string
		role              string
		traits            []string
		goals             []string
		biases            []string
		power             float64
		personalityFactor float64
	}{
		{
			name:              "Tech Innovator",
			role:              "Uses technology to break rules",
			traits:            []string{"visionary", "technical", "impatient", "first-principles"},
			goals:             []string{"automate everything", "eliminate intermediaries", "create platforms"},
			biases:            []string{"technology solutionism", "disruption fetish"},
			power:             0.7,
			personalityFactor: 1.3,
		},
		{
			name:              "Business Model Changer",
			role:              "Rewrites how value is created and captured",
			traits:            []string{"creative", "customer-obsessed", "margin-agnostic", "ecosystem-thinker"},
			goals:             []string{"find new revenue models", "own the platform", "make incumbents irrelevant"},
			biases:            []string{"platform bias", "network effect obsession"},
			power:             0.7,
			personalityFactor: 1.2,
		},
		{
			name:              "Progressive Regulator",
			role:              "Changes the rules of the game via regulation",
			traits:            []string{"idealistic", "politically-savvy", "long-term", "coalition-builder"},
			goals:             []string{"protect new entrants", "break monopolies", "enforce new standards"},
			biases:            []string{"regulatory capture risk", "unintended consequences"},
			power:             0.8,
			personalityFactor: 0.8, // fires less often but high impact
		},
		{
			name:              "Organized Consumer",
			role:              "Collective consumer movement that rewrites demand",
			traits:            []string{"values-driven", "social-media-native", "viral", "uncompromising"},
			goals:             []string{"force transparency", "create alternatives", "punish bad actors"},
			biases:            []string{"outrage bias", "purity spiral"},
			power:             0.6,
			personalityFactor: rand.Float64()*0.5 + 0.8,
		},
		// ── New 8 ─────────────────────────────────────────────────────────────
		{
			name:              "Venture Capital Fund",
			role:              "Patient capital hunting for category-defining bets",
			traits:            []string{"asymmetric-return-seeking", "narrative-driven", "portfolio-thinking", "contrarian"},
			goals:             []string{"fund the rule-breaker", "create new categories", "exit at 100x"},
			biases:            []string{"power law obsession", "founder worship", "FOMO on hot sectors"},
			power:             0.85,
			personalityFactor: 1.4,
		},
		{
			name:              "Big Tech Entrant",
			role:              "Platform giant entering adjacent market with unlimited resources",
			traits:            []string{"resource-abundant", "distribution-dominant", "ecosystem-colonizing", "patient"},
			goals:             []string{"own the new layer", "leverage existing user base", "commoditize the complement"},
			biases:            []string{"platform extension bias", "anti-competitive blind spot"},
			power:             0.95,
			personalityFactor: 1.1,
		},
		{
			name:              "Social Movement",
			role:              "Decentralized activist network rewriting cultural norms",
			traits:            []string{"decentralized", "values-absolute", "viral", "unpredictable"},
			goals:             []string{"force systemic change", "hold power accountable", "rewrite acceptable behavior"},
			biases:            []string{"moral purity spiral", "outrage amplification", "short attention span"},
			power:             0.65,
			personalityFactor: rand.Float64()*0.6 + 1.0,
		},
		{
			name:              "International Regulator",
			role:              "Supranational body imposing cross-border standards",
			traits:            []string{"jurisdiction-expanding", "precedent-setting", "slow-but-inevitable", "coalition-dependent"},
			goals:             []string{"harmonize global standards", "prevent regulatory arbitrage", "protect citizens at scale"},
			biases:            []string{"one-size-fits-all regulation", "enforcement lag"},
			power:             0.9,
			personalityFactor: 0.7,
		},
		{
			name:              "Open Source Community",
			role:              "Decentralized developer collective commoditizing proprietary value",
			traits:            []string{"collaborative", "anti-proprietary", "meritocratic", "globally-distributed"},
			goals:             []string{"make technology free", "prevent lock-in", "outcompete through community"},
			biases:            []string{"anti-commercial bias", "tragedy of the commons risk"},
			power:             0.7,
			personalityFactor: 1.2,
		},
		{
			name:              "Sovereign Wealth Fund",
			role:              "State-backed capital deploying geopolitical strategy through investment",
			traits:            []string{"patient-capital", "geopolitically-motivated", "opaque", "strategic-sector-focused"},
			goals:             []string{"secure strategic assets", "export national influence", "diversify state revenue"},
			biases:            []string{"national interest over returns", "opacity bias"},
			power:             0.9,
			personalityFactor: 0.9,
		},
		{
			name:              "Adjacent Startup",
			role:              "Well-funded startup attacking from an unexpected adjacent market",
			traits:            []string{"cross-domain", "fast-moving", "customer-obsessed", "rule-ignorant"},
			goals:             []string{"reframe the problem", "steal customers from incumbents", "force category redefinition"},
			biases:            []string{"outsider advantage overconfidence", "product-market fit tunnel vision"},
			power:             0.6,
			personalityFactor: 1.5,
		},
		{
			name:              "Whistleblower",
			role:              "Insider exposing systemic failures and forcing transparency",
			traits:            []string{"truth-driven", "risk-tolerant", "isolated", "high-impact-low-probability"},
			goals:             []string{"expose hidden rules", "trigger regulatory action", "shift public trust"},
			biases:            []string{"moral absolutism", "underestimation of personal cost"},
			power:             0.5,
			personalityFactor: rand.Float64()*1.0 + 0.5,
		},
	}

	agents := make([]engine.Agent, 0, len(specs))
	for _, s := range specs {
		agents = append(agents, &disruptorAgent{
			BaseAgent: engine.NewBaseAgent(
				uuid.New().String(),
				engine.AgentDisruptor,
				engine.DisruptorPermissions,
				engine.Personality{
					Name:        s.name,
					Role:        s.role,
					Traits:      s.traits,
					Goals:       s.goals,
					Biases:      s.biases,
					PowerWeight: s.power,
				},
			),
			llm:               llm,
			personalityFactor: s.personalityFactor,
		})
	}
	return agents
}
