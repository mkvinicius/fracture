package engine

import (
	"context"
	"fmt"
	"strings"
)

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

// agentVote determines how an agent votes on a fracture proposal.
// Uses goal-alignment scoring: agents vote YES when their goals match the proposal,
// NO when their personality traits (risk-aversion, conservatism) oppose it,
// and apply domain-specific resistance logic for regulatory proposals.
func (v *Voter) agentVote(agent Agent, proposal RuleProposal) (bool, string) {
	p := agent.Personality()

	// Disruptors always vote YES — they exist to break rules
	if agent.Type() == AgentDisruptor {
		return true, "Disruptors support rule changes by nature"
	}

	// Detect risk-averse / conservative personality
	isRiskAverse := false
	for _, trait := range p.Traits {
		t := strings.ToLower(trait)
		if t == "conservative" || t == "risk-averse" || t == "status quo bias" ||
			t == "process-driven" || t == "compliance-obsessed" || t == "precautionary principle" {
			isRiskAverse = true
			break
		}
	}

	// Extremely fragile proposals (NewStability < 0.25) are seen as dangerously experimental
	if isRiskAverse && proposal.NewStability < 0.25 {
		return false, fmt.Sprintf("%s opposes a dangerously unstable proposal (stability %.2f)", p.Name, proposal.NewStability)
	}

	// Regulatory proposals: risk-averse agents resist disruption of rules they depend on
	if proposal.NewDomain == DomainRegulation && isRiskAverse {
		return false, p.Name + " opposes regulatory disruption — too much uncertainty"
	}

	// Goal-alignment scoring: count how many goals are echoed in the proposal text
	proposalText := strings.ToLower(proposal.NewDescription + " " + proposal.Rationale)
	alignScore := 0
	for _, goal := range p.Goals {
		// Extract meaningful keywords from each goal (skip short filler words)
		for _, word := range strings.Fields(strings.ToLower(goal)) {
			if len(word) > 4 && strings.Contains(proposalText, word) {
				alignScore++
				break // count at most once per goal
			}
		}
	}

	if alignScore >= 2 {
		return true, fmt.Sprintf("%s sees this proposal advancing %d of their strategic goals", p.Name, alignScore)
	}
	if alignScore == 1 {
		// Weak alignment: fragile rules are allowed, stable ones less so
		if proposal.NewStability < 0.50 {
			return false, p.Name + " is cautious — proposal is still too disruptive"
		}
		return true, p.Name + " cautiously supports this — one goal aligns"
	}

	// No goal alignment: conformists default to resistance
	return false, p.Name + " sees no benefit — no strategic goals advanced"
}
