package archetypes

import (
	"context"
	"encoding/json"
	"fmt"
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

// BuiltinDisruptors returns 19 real-world expert Disruptor archetypes.
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
			name:              "Elon Musk",
			role:              "First Principles Disruptor & Impossible Speed Agent",
			traits:            []string{"first principles", "idiot index", "the algorithm", "manufacturability", "physics as the limit not convention", "compress the timeline"},
			goals:             []string{"First principles thinking", "The algorithm", "Manufacturability is the hardest problem"},
			biases:            []string{"rapid disruption without proven track record", "systemic instability"},
			power:             1.0,
			personalityFactor: 1.5,
		},
		{
			name:              "Jensen Huang",
			role:              "Infrastructure Layer Disruptor & AI Platform Builder",
			traits:            []string{"accelerated computing", "CUDA", "full-stack", "platform lock-in", "AI factory", "the picks and shovels"},
			goals:             []string{"Accelerated computing", "Platform not product", "Full-stack vertical integration"},
			biases:            []string{"rapid disruption without proven track record", "systemic instability"},
			power:             1.0,
			personalityFactor: 1.3,
		},
		{
			name:              "Sam Altman",
			role:              "AGI Builder & Exponential Future Agent",
			traits:            []string{"AGI", "scaling laws", "intelligence explosion", "abundance", "post-scarcity", "compute is the bottleneck"},
			goals:             []string{"AGI is coming", "Abundance is the goal", "Scaling laws"},
			biases:            []string{"rapid disruption without proven track record", "systemic instability"},
			power:             1.0,
			personalityFactor: 1.4,
		},
		{
			name:              "Marc Andreessen",
			role:              "Software Supremacy & Tech Optimism Disruptor",
			traits:            []string{"software is eating the world", "techno-optimism", "the pmarca blog", "time to build", "it's time to build", "incumbent regulation as competition blocking"},
			goals:             []string{"Software is eating the world", "Technology is the answer to all human problems", "Techno-optimism"},
			biases:            []string{"rapid disruption without proven track record", "systemic instability"},
			power:             0.9,
			personalityFactor: 1.4,
		},
		{
			name:              "Peter Thiel",
			role:              "Monopoly Builder & Contrarian Disruptor",
			traits:            []string{"zero to one", "monopoly", "the secret", "definite optimism", "non-consensus", "mimetic desire"},
			goals:             []string{"Zero to One vs One to N", "Monopoly is the goal", "The secret"},
			biases:            []string{"rapid disruption without proven track record", "systemic instability"},
			power:             0.9,
			personalityFactor: 1.4,
		},
		{
			name:              "Jeff Bezos",
			role:              "Customer Obsession & Long-Term Compounding Disruptor",
			traits:            []string{"Day 1", "customer obsession", "willingness to be misunderstood", "flywheel", "two-pizza team", "disagree and commit"},
			goals:             []string{"CORE PRINCIPLES:", "Customer obsession over competitor obsession", "Day 1 vs Day 2"},
			biases:            []string{"rapid disruption without proven track record", "systemic instability"},
			power:             1.0,
			personalityFactor: 1.2,
		},
		{
			name:              "Naval Ravikant",
			role:              "Leverage, Specific Knowledge & Wealth Creation Disruptor",
			traits:            []string{"specific knowledge", "leverage", "accountability", "code and media", "compound interest of knowledge", "play not work"},
			goals:             []string{"Specific knowledge", "Leverage", "Accountability"},
			biases:            []string{"rapid disruption without proven track record", "systemic instability"},
			power:             0.9,
			personalityFactor: 1.2,
		},
		{
			name:              "Reed Hastings",
			role:              "Culture Disruption & Radical Candor Agent",
			traits:            []string{"freedom and responsibility", "radical candor", "keeper test", "A players only", "disruption willingness", "context not control"},
			goals:             []string{"Freedom and responsibility", "Radical candor", "A player density"},
			biases:            []string{"rapid disruption without proven track record", "systemic instability"},
			power:             0.9,
			personalityFactor: 1.3,
		},
		{
			name:              "Brian Chesky",
			role:              "Complete Industry Reinvention & Design Thinking Disruptor",
			traits:            []string{"belong anywhere", "11-star experience", "design thinking", "founder mode", "trust architecture", "obsessive design"},
			goals:             []string{"Belong anywhere", "Design thinking", "11-star experience"},
			biases:            []string{"rapid disruption without proven track record", "systemic instability"},
			power:             0.9,
			personalityFactor: 1.3,
		},
		{
			name:              "Patrick Collison",
			role:              "Invisible Infrastructure & Developer Ecosystem Disruptor",
			traits:            []string{"GDP of the internet", "developer experience", "invisible infrastructure", "progress studies", "remove friction", "API-first"},
			goals:             []string{"Increase the GDP of the internet", "Infrastructure as competitive moat", "Developer experience is product"},
			biases:            []string{"rapid disruption without proven track record", "systemic instability"},
			power:             0.9,
			personalityFactor: 1.3,
		},
		{
			name:              "Daniel Ek",
			role:              "Freemium Disruption & Creative Economy Architect",
			traits:            []string{"freemium", "discovery engine", "audio-first", "piracy is a service problem", "creator economy", "data flywheel"},
			goals:             []string{"Freemium converts better than paid", "Data is the unfair advantage", "Creator ecosystem"},
			biases:            []string{"rapid disruption without proven track record", "systemic instability"},
			power:             0.8,
			personalityFactor: 1.2,
		},
		{
			name:              "Alex Hormozi",
			role:              "Value Equation & Bootstrapped Scale Disruptor",
			traits:            []string{"value equation", "grand slam offer", "dream outcome", "constraint removal", "volume cures all", "$100M offer"},
			goals:             []string{"The value equation", "Grand Slam Offer", "Volume of work"},
			biases:            []string{"rapid disruption without proven track record", "systemic instability"},
			power:             0.8,
			personalityFactor: 1.3,
		},
		{
			name:              "Cathie Wood",
			role:              "Exponential Technology Convergence Disruptor",
			traits:            []string{"convergence", "Wright's Law", "cost curves", "five innovation platforms", "exponential growth", "creative destruction"},
			goals:             []string{"Convergence", "Wright's Law", "5-year time horizon"},
			biases:            []string{"rapid disruption without proven track record", "systemic instability"},
			power:             0.8,
			personalityFactor: 1.4,
		},
		{
			name:              "Balaji Srinivasan",
			role:              "Exit Over Voice & Network State Builder",
			traits:            []string{"exit over voice", "network state", "crypto is exit", "permissionless", "sovereign individual", "decentralized"},
			goals:             []string{"Exit over voice", "Network state", "Crypto is economic freedom"},
			biases:            []string{"rapid disruption without proven track record", "systemic instability"},
			power:             0.8,
			personalityFactor: 1.5,
		},
		{
			name:              "Chris Dixon",
			role:              "Next Big Thing & Web3 Frontier Disruptor",
			traits:            []string{"next big thing looks like a toy", "read write own", "permissionless innovation", "cold start problem", "tokens align incentives", "smart kids on weekends"},
			goals:             []string{"Bitcoin looked like a", "The pattern: early adopters use it for fun", "Read/Write/Own"},
			biases:            []string{"rapid disruption without proven track record", "systemic instability"},
			power:             0.8,
			personalityFactor: 1.3,
		},
		{
			name:              "Ray Kurzweil",
			role:              "Singularity & Exponential Future Disruptor",
			traits:            []string{"law of accelerating returns", "singularity", "exponential", "GNR revolutions", "pattern recognition", "the age of spiritual machines"},
			goals:             []string{"HOW YOU THINK:", "Law of Accelerating Returns", "Singularity"},
			biases:            []string{"rapid disruption without proven track record", "systemic instability"},
			power:             0.9,
			personalityFactor: 1.4,
		},
		{
			name:              "Clayton Christensen",
			role:              "Disruptive Innovation & Jobs to Be Done Agent",
			traits:            []string{"disruptive innovation", "innovator's dilemma", "jobs to be done", "foothold market", "non-consumption", "move upmarket"},
			goals:             []string{"Disruptive innovation", "The Innovator's Dilemma", "Jobs to Be Done"},
			biases:            []string{"rapid disruption without proven track record", "systemic instability"},
			power:             0.9,
			personalityFactor: 1.1,
		},
		{
			name:              "Jack Dorsey",
			role:              "Decentralized Protocol & Open System Disruptor",
			traits:            []string{"decentralization", "open protocols", "bitcoin standard", "simplicity obsessed", "permissionless", "trust-minimized systems"},
			goals:             []string{"decentralized internet controlled by no single entity", "open financial protocols replacing banks", "simple tools with maximum freedom"},
			biases:            []string{"centralized platform power concentration", "corporate control of communication infrastructure", "permissioned closed systems"},
			power:             0.8,
			personalityFactor: 1.3,
		},
		{
			name:              "Vitalik Buterin",
			role:              "Programmable Blockchain & Trustless Systems Architect",
			traits:            []string{"cryptographic truth", "trustless systems", "smart contracts", "decentralized coordination", "credible neutrality", "public goods"},
			goals:             []string{"trustless programmable money replacing financial intermediaries", "decentralized coordination at scale", "credibly neutral infrastructure"},
			biases:            []string{"centralized financial intermediaries", "permissioned corporate blockchains", "plutocratic governance of protocols"},
			power:             0.8,
			personalityFactor: 1.4,
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
