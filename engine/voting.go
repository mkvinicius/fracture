package engine

import "context"

// VoteRecord captures one agent's vote on a FRACTURE POINT proposal.
type VoteRecord struct {
	AgentID     string  `json:"agent_id"`
	AgentName   string  `json:"agent_name"`
	Vote        bool    `json:"vote"`   // true = accept, false = reject
	Weight      float64 `json:"weight"` // power weight of this agent
	Rationale   string  `json:"rationale"`
}

// Voter manages the weighted voting mechanism for FRACTURE POINT proposals.
type Voter struct {
	agents []Agent
}

// NewVoter creates a Voter from the simulation's agent pool.
func NewVoter(agents []Agent) *Voter {
	return &Voter{agents: agents}
}

// Vote runs the weighted vote on a RuleProposal.
// Returns (accepted bool, breakdown []VoteRecord).
// A proposal is accepted when weighted YES votes exceed 50% of total weight.
func (v *Voter) Vote(ctx context.Context, proposal RuleProposal, actions []AgentAction) (bool, []VoteRecord) {
	// Build a quick lookup: agentID -> action text (for context)
	actionByAgent := make(map[string]AgentAction, len(actions))
	for _, a := range actions {
		actionByAgent[a.AgentID] = a
	}

	var breakdown []VoteRecord
	var totalWeight, yesWeight float64

	for _, agent := range v.agents {
		p := agent.Personality()
		weight := p.PowerWeight
		if weight == 0 {
			weight = 0.5 // default equal weight
		}
		totalWeight += weight

		// Determine vote based on agent type and personality alignment
		vote, rationale := v.agentVote(agent, proposal)

		breakdown = append(breakdown, VoteRecord{
			AgentID:   agent.ID(),
			AgentName: p.Name,
			Vote:      vote,
			Weight:    weight,
			Rationale: rationale,
		})

		if vote {
			yesWeight += weight
		}
	}

	accepted := totalWeight > 0 && (yesWeight/totalWeight) > 0.5
	return accepted, breakdown
}

// agentVote produces a simple heuristic vote for an agent.
// In production this would call the LLM; here we use a fast heuristic
// to avoid N extra LLM calls per FRACTURE POINT.
func (v *Voter) agentVote(agent Agent, proposal RuleProposal) (bool, string) {
	p := agent.Personality()

	// Disruptors always vote YES on fracture proposals
	if agent.Type() == AgentDisruptor {
		return true, "Disruptors support rule changes by nature"
	}

	// Conformists vote based on how much the new rule threatens their stability
	// Low stability rules (fragile) are more likely to be accepted
	if proposal.NewStability < 0.4 {
		return true, p.Name + " sees this change as low-risk"
	}

	// Domain-based heuristic: regulators resist market changes
	if proposal.NewDomain == DomainRegulation {
		for _, trait := range p.Traits {
			if trait == "conservative" || trait == "risk-averse" {
				return false, p.Name + " opposes regulatory disruption"
			}
		}
	}

	// Default: conformists resist change
	return false, p.Name + " prefers the current rules"
}
