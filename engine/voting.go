package engine

import (
	"context"
	"fmt"
	"sort"
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

// llmVotingThreshold is the minimum PowerWeight for LLM-driven voting.
// Agents at or above this threshold reason via LLM; others use the fast heuristic.
const llmVotingThreshold = 0.85

// maxLLMVoters caps the number of LLM calls per fracture vote to control cost.
const maxLLMVoters = 5

// Voter manages the weighted voting mechanism for FRACTURE POINT proposals.
type Voter struct {
	agents []Agent
	llm    LLMCaller // optional; if set, high-power agents vote via LLM reasoning
}

// NewVoter creates a Voter. If llm is non-nil, the top-N agents by power weight
// will reason through their vote using the LLM instead of the fast heuristic.
func NewVoter(agents []Agent, llm LLMCaller) *Voter {
	return &Voter{agents: agents, llm: llm}
}

// Vote runs the weighted vote on a RuleProposal.
// Returns (accepted bool, breakdown []VoteRecord).
// A proposal is accepted when weighted YES votes exceed 50% of total weight.
// If v.llm is set, the top maxLLMVoters agents by power weight vote via LLM reasoning.
func (v *Voter) Vote(ctx context.Context, proposal RuleProposal, actions []AgentAction) (bool, []VoteRecord) {
	// Identify which agents qualify for LLM voting (sorted by power, top N)
	llmVoterIDs := v.selectLLMVoters()

	var breakdown []VoteRecord
	var totalWeight, yesWeight float64

	for _, agent := range v.agents {
		p := agent.Personality()
		weight := p.PowerWeight
		if weight == 0 {
			weight = 0.5
		}
		totalWeight += weight

		var vote bool
		var rationale string

		if v.llm != nil && llmVoterIDs[agent.ID()] {
			vote, rationale = v.agentVoteLLM(ctx, agent, proposal)
		} else {
			vote, rationale = v.agentVote(agent, proposal)
		}

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

// selectLLMVoters returns a set of agent IDs that should vote via LLM.
// Picks the top maxLLMVoters agents by power weight above llmVotingThreshold.
func (v *Voter) selectLLMVoters() map[string]bool {
	if v.llm == nil {
		return nil
	}
	type pair struct {
		id     string
		weight float64
	}
	var eligible []pair
	for _, a := range v.agents {
		pw := a.Personality().PowerWeight
		if pw >= llmVotingThreshold {
			eligible = append(eligible, pair{a.ID(), pw})
		}
	}
	sort.Slice(eligible, func(i, j int) bool { return eligible[i].weight > eligible[j].weight })
	result := make(map[string]bool, maxLLMVoters)
	for i, p := range eligible {
		if i >= maxLLMVoters {
			break
		}
		result[p.id] = true
	}
	return result
}

// agentVoteLLM calls the LLM to reason through a vote for a high-power agent.
// Falls back to heuristic on error.
func (v *Voter) agentVoteLLM(ctx context.Context, agent Agent, proposal RuleProposal) (bool, string) {
	p := agent.Personality()

	system := fmt.Sprintf(
		`You are %s — %s.
Your traits: %s
Your goals: %s
Your biases: %s

You are voting on a proposed rule change in a strategic simulation.
Answer with exactly: ACCEPT or REJECT, followed by one sentence of reasoning.
Be decisive and stay in character. Respond in Brazilian Portuguese (PT-BR).`,
		p.Name, p.Role,
		strings.Join(p.Traits, ", "),
		strings.Join(p.Goals, ", "),
		strings.Join(p.Biases, ", "),
	)

	user := fmt.Sprintf(
		"Proposed change: \"%s\"\nDomain: %s | New stability: %.2f\nRationale: %s\n\nDo you ACCEPT or REJECT?",
		proposal.NewDescription, proposal.NewDomain, proposal.NewStability, proposal.Rationale,
	)

	raw, _, err := v.llm.Call(ctx, system, user, 120)
	if err != nil {
		return v.agentVote(agent, proposal) // fallback to heuristic
	}

	upper := strings.ToUpper(strings.TrimSpace(raw))
	vote := strings.HasPrefix(upper, "ACCEPT")

	// Extract the reasoning sentence after ACCEPT/REJECT
	rationale := raw
	if idx := strings.Index(raw, " "); idx > 0 && idx < 10 {
		rationale = strings.TrimSpace(raw[idx:])
	}
	return vote, fmt.Sprintf("[LLM] %s", rationale)
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
