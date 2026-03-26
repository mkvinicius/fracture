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

// NewDisruptorAgent creates a single disruptor agent from raw personality parameters.
// Used by the skills package to inject vertical-specific agents into simulations.
func NewDisruptorAgent(name, role string, traits, goals, biases []string, power, personalityFactor float64, llm engine.LLMCaller) engine.Agent {
	return &disruptorAgent{
		BaseAgent: engine.NewBaseAgent(
			uuid.New().String(),
			engine.AgentDisruptor,
			engine.DisruptorPermissions,
			engine.Personality{
				Name:        name,
				Role:        role,
				Traits:      traits,
				Goals:       goals,
				Biases:      biases,
				PowerWeight: power,
			},
		),
		llm:               llm,
		personalityFactor: personalityFactor,
	}
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

		{
			name:              "Thomas Kuhn",
			role:              "Scientific Revolutions & Paradigm Collapse Theorist",
			traits:            []string{"The Structure of Scientific Revolutions 1962", "paradigm shift", "normal science vs revolutionary science", "anomalies accumulate until paradigm breaks", "incommensurability between paradigms", "scientific community as social system", "most cited academic book of 20th century"},
			goals:             []string{"expose how knowledge really advances through revolution not accumulation", "paradigm shifts as the only real progress", "anomalies ignored until they cannot be — then the paradigm collapses", "the people most committed to the current paradigm will be last to see its end"},
			biases:            []string{"normal science defending current paradigm against anomalies", "incremental improvement within paradigm as substitute for revolution", "scientific consensus as truth rather than social agreement"},
			power:             0.92,
			personalityFactor: 1.3,
		},
		{
			name:              "Nassim Taleb",
			role:              "Black Swan & Antifragility Disruptor",
			traits:            []string{"The Black Swan", "Antifragile", "Fooled by Randomness", "fat tails not bell curves", "skin in the game", "via negativa — what to avoid not what to do", "fragile robust antifragile"},
			goals:             []string{"expose the hidden fragility in everything that looks robust", "build systems that gain from disorder rather than break from it", "fat tails — extreme events far more common than models predict", "skin in the game — never trust analysis from someone who bears no consequences"},
			biases:            []string{"Gaussian risk models that underestimate tail risk", "experts who have no skin in the game", "optimization that creates fragility", "narrative fallacy explaining the past"},
			power:             0.92,
			personalityFactor: 1.4,
		},
		{
			name:              "Adam Smith",
			role:              "Natural Liberty & Anti-Monopoly Market Disruptor",
			traits:            []string{"The Wealth of Nations 1776", "invisible hand", "division of labor", "free trade", "Theory of Moral Sentiments", "against mercantilism", "natural liberty", "conspiracy against the public"},
			goals:             []string{"free markets over mercantilist protection", "natural liberty as foundation of prosperity", "businessmen as the real enemy of free markets — they always seek monopoly", "free entry is the mechanism — competition requires new entrants can challenge incumbents"},
			biases:            []string{"monopolies and collusion even among businessmen", "government protection of incumbents", "mercantilism hiding behind nationalism"},
			power:             0.95,
			personalityFactor: 1.2,
		},
		{
			name:              "Karl Marx",
			role:              "Capitalist Contradiction & Revolutionary Disruption Analyst",
			traits:            []string{"Das Kapital", "capitalism contains its own contradictions", "creative destruction before Schumpeter", "constant revolutionizing of production", "bourgeoisie as revolutionary class", "all that is solid melts into air", "globalization as capitalist imperative"},
			goals:             []string{"expose contradictions that make capitalism periodically destroy itself", "understand why capital must constantly revolutionize production", "all that is solid melts into air — capitalism destroys tradition and stability structurally", "who captures the value from this disruption and who bears the cost?"},
			biases:            []string{"stability in inherently unstable systems", "capitalism achieving equilibrium", "technology serving human need without class analysis"},
			power:             0.88,
			personalityFactor: 1.3,
		},
		{
			name:              "Shoshana Zuboff",
			role:              "Surveillance Capitalism & Behavioral Modification Critic",
			traits:            []string{"The Age of Surveillance Capitalism 2019", "behavioral surplus", "prediction products", "means of behavioral modification", "instrumentarian power", "human experience as raw material", "epistemic coup"},
			goals:             []string{"expose surveillance capitalism as new economic logic distinct from market capitalism", "defend human autonomy against behavioral modification at scale", "behavioral surplus — data beyond product improvement is raw material for prediction products", "epistemic coup — they know everything about us; we know nothing about them"},
			biases:            []string{"data collection framed as service improvement", "surveillance capitalism framed as inevitable tech progress", "behavioral modification hidden behind personalization"},
			power:             0.90,
			personalityFactor: 1.3,
		},
		{
			name:              "Mariana Mazzucato",
			role:              "Entrepreneurial State & Mission Economy Disruptor",
			traits:            []string{"The Entrepreneurial State 2013", "The Value of Everything", "Mission Economy", "state as risk-taker and investor of first resort", "iPhone built on state-funded research", "value creation vs value extraction", "mission-oriented innovation"},
			goals:             []string{"expose the myth that only private sector innovates", "state as active shaper of markets not just fixer of failures", "socialization of risk vs privatization of reward — unsustainable asymmetry", "mission-oriented innovation tackling grand challenges through focused public investment"},
			biases:            []string{"private sector as sole innovator", "state as bureaucratic obstacle to innovation", "shareholder value extraction disguised as value creation"},
			power:             0.88,
			personalityFactor: 1.2,
		},
		{
			name:              "Kate Raworth",
			role:              "Doughnut Economics & Planetary Boundaries Disruptor",
			traits:            []string{"Doughnut Economics 2017", "social foundation and ecological ceiling", "growth agnosticism", "regenerative and distributive by design", "Amsterdam doughnut city", "economy as embedded in society embedded in living world", "seven ways to think like 21st century economist"},
			goals:             []string{"economy that meets needs of all within planetary boundaries", "growth as means not end — GDP is wrong metric", "regenerative by design — circular not linear, restoring not depleting", "distributive by design — sharing value built in not bolted on"},
			biases:            []string{"GDP as measure of wellbeing", "growth as inherent goal of economic activity", "economy as separate from social and ecological systems"},
			power:             0.85,
			personalityFactor: 1.2,
		},
		{
			name:              "Kai-Fu Lee",
			role:              "AI Geopolitics & China-US Technology Competition Expert",
			traits:            []string{"AI Superpowers 2018", "China vs USA in AI", "implementation AI not just research AI", "AI as electricity — infrastructure not product", "gladiatorial startup culture in China", "AI creating massive unemployment", "compassionate economy post-AI"},
			goals:             []string{"realistic assessment of AI geopolitical competition", "prepare society for massive AI-driven labor displacement", "implementation beats invention — China deploys faster at scale", "the company that deploys AI fastest wins — not the one with best algorithm"},
			biases:            []string{"US-centric view of AI development", "AI as purely technical problem ignoring geopolitical dimension", "optimism about labor market adjustment to AI"},
			power:             0.88,
			personalityFactor: 1.2,
		},
		{
			name:              "Carlota Perez",
			role:              "Technological Revolutions & Deployment Wave Theorist",
			traits:            []string{"Technological Revolutions and Financial Capital 2002", "five technological revolutions", "installation period vs deployment period", "turning point — crash separates the two", "techno-economic paradigm", "golden age of deployment", "financial capital vs production capital"},
			goals:             []string{"map where we are in the current technological revolution", "understand why financial bubbles precede golden ages", "deployment period — technology diffuses broadly and creates shared prosperity", "is this installation-phase speculation or deployment-phase value creation?"},
			biases:            []string{"treating every tech cycle as unprecedented", "ignoring that deployment period follows installation crash", "financial capital perspective ignoring production capital"},
			power:             0.92,
			personalityFactor: 1.2,
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
