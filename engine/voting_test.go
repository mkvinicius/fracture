package engine

import (
	"context"
	"strings"
	"testing"
)

// ─── mockLLM ─────────────────────────────────────────────────────────────────

type mockLLM struct {
	response string
	err      error
	calls    int
}

func (m *mockLLM) Call(_ context.Context, _, _ string, _ int) (string, int, error) {
	m.calls++
	return m.response, len(m.response), m.err
}

// ─── helpers ─────────────────────────────────────────────────────────────────

// makeDisruptorAgent returns an agent of type AgentDisruptor with the given id/power.
func makeDisruptorAgent(id string, pw float64) Agent {
	p := Personality{Name: id, PowerWeight: pw}
	ba := NewBaseAgent(id, AgentDisruptor, DisruptorPermissions, p)
	return &testAgent{BaseAgent: ba}
}

// makeConformistWithTraits builds a conformist agent with specific traits and goals.
func makeConformistWithTraits(id string, pw float64, traits, goals []string) Agent {
	p := Personality{Name: id, PowerWeight: pw, Traits: traits, Goals: goals}
	ba := NewBaseAgent(id, AgentConformist, ConformistPermissions, p)
	return &testAgent{BaseAgent: ba}
}

// simpleProposal creates a RuleProposal for testing.
func simpleProposal(desc string, domain RuleDomain, stability float64) RuleProposal {
	return RuleProposal{
		OriginalRuleID: "rule-1",
		NewDescription: desc,
		NewDomain:      domain,
		NewStability:   stability,
		Rationale:      "test rationale",
		ProposedByAgent: "disruptor-1",
	}
}

// ─── selectLLMVoters ─────────────────────────────────────────────────────────

// TestSelectLLMVoters_NilLLMReturnsNil: NewVoter with nil llm → selectLLMVoters returns nil.
func TestSelectLLMVoters_NilLLMReturnsNil(t *testing.T) {
	agents := []Agent{makeTestAgent("a1", 0.9)}
	v := NewVoter(agents, nil)
	result := v.selectLLMVoters()
	if result != nil {
		t.Errorf("expected nil map when llm is nil, got %v", result)
	}
}

// TestSelectLLMVoters_BelowThresholdExcluded: agents with PowerWeight < 0.85 not selected.
func TestSelectLLMVoters_BelowThresholdExcluded(t *testing.T) {
	agents := []Agent{
		makeTestAgent("low1", 0.50),
		makeTestAgent("low2", 0.84),
		makeTestAgent("below", 0.84999),
	}
	llm := &mockLLM{response: "ACCEPT reason"}
	v := NewVoter(agents, llm)
	result := v.selectLLMVoters()
	for _, a := range agents {
		if result[a.ID()] {
			t.Errorf("agent %q (pw=%.5f) should be excluded (below threshold 0.85)", a.ID(), a.Personality().PowerWeight)
		}
	}
}

// TestSelectLLMVoters_AboveThresholdIncluded: agents with PowerWeight ≥ 0.85 are selected.
func TestSelectLLMVoters_AboveThresholdIncluded(t *testing.T) {
	agents := []Agent{
		makeTestAgent("exact", 0.85),
		makeTestAgent("above", 0.90),
		makeTestAgent("low", 0.80),
	}
	llm := &mockLLM{response: "ACCEPT reason"}
	v := NewVoter(agents, llm)
	result := v.selectLLMVoters()

	if !result["exact"] {
		t.Error("agent 'exact' (pw=0.85) should be included (at threshold)")
	}
	if !result["above"] {
		t.Error("agent 'above' (pw=0.90) should be included (above threshold)")
	}
	if result["low"] {
		t.Error("agent 'low' (pw=0.80) should be excluded (below threshold)")
	}
}

// TestSelectLLMVoters_CappedAtMaxLLMVoters: if 10 agents qualify, only top 5 are selected.
func TestSelectLLMVoters_CappedAtMaxLLMVoters(t *testing.T) {
	agents := make([]Agent, 10)
	for i := 0; i < 10; i++ {
		pw := 0.85 + float64(i)*0.01 // 0.85 to 0.94 — all qualify
		agents[i] = makeTestAgent("a"+string(rune('0'+i)), pw)
	}
	llm := &mockLLM{response: "ACCEPT"}
	v := NewVoter(agents, llm)
	result := v.selectLLMVoters()

	if len(result) != maxLLMVoters {
		t.Errorf("expected exactly %d LLM voters, got %d", maxLLMVoters, len(result))
	}
}

// TestSelectLLMVoters_SortedByPower: top-5 by power are selected, not arbitrary 5.
func TestSelectLLMVoters_SortedByPower(t *testing.T) {
	// 7 qualifying agents; top 5 by power should be selected.
	agents := []Agent{
		makeTestAgent("pw-0.86", 0.86),
		makeTestAgent("pw-0.99", 0.99),
		makeTestAgent("pw-0.87", 0.87),
		makeTestAgent("pw-0.95", 0.95),
		makeTestAgent("pw-0.92", 0.92),
		makeTestAgent("pw-0.88", 0.88),
		makeTestAgent("pw-0.91", 0.91),
	}
	llm := &mockLLM{response: "ACCEPT"}
	v := NewVoter(agents, llm)
	result := v.selectLLMVoters()

	// Top 5 by power: 0.99, 0.95, 0.92, 0.91, 0.88
	expectedIn := []string{"pw-0.99", "pw-0.95", "pw-0.92", "pw-0.91", "pw-0.88"}
	expectedOut := []string{"pw-0.86", "pw-0.87"}

	for _, id := range expectedIn {
		if !result[id] {
			t.Errorf("agent %q should be in top-%d LLM voters", id, maxLLMVoters)
		}
	}
	for _, id := range expectedOut {
		if result[id] {
			t.Errorf("agent %q should NOT be in top-%d LLM voters (power too low)", id, maxLLMVoters)
		}
	}
}

// ─── agentVote (heuristic) ────────────────────────────────────────────────────

// TestAgentVote_DisruptorAlwaysAccepts: disruptor agent always returns vote=true.
func TestAgentVote_DisruptorAlwaysAccepts(t *testing.T) {
	v := NewVoter(nil, nil)
	disruptor := makeDisruptorAgent("d1", 0.7)
	proposal := simpleProposal("anything", DomainMarket, 0.1)

	vote, _ := v.agentVote(disruptor, proposal)
	if !vote {
		t.Error("disruptor should always vote YES")
	}
}

// TestAgentVote_RiskAverseRejectsUnstable: conformist with "risk-averse" trait and
// proposal.NewStability=0.1 → vote=false.
func TestAgentVote_RiskAverseRejectsUnstable(t *testing.T) {
	v := NewVoter(nil, nil)
	agent := makeConformistWithTraits("cautious", 0.5, []string{"risk-averse"}, nil)
	proposal := simpleProposal("some change", DomainMarket, 0.1)

	vote, _ := v.agentVote(agent, proposal)
	if vote {
		t.Error("risk-averse conformist should reject unstable proposals (stability=0.1)")
	}
}

// TestAgentVote_RiskAverseRejectsRegulation: conformist with "conservative" trait,
// proposal domain=DomainRegulation → vote=false.
func TestAgentVote_RiskAverseRejectsRegulation(t *testing.T) {
	v := NewVoter(nil, nil)
	agent := makeConformistWithTraits("conservative-agent", 0.5, []string{"conservative"}, nil)
	proposal := simpleProposal("regulatory overhaul", DomainRegulation, 0.6)

	vote, _ := v.agentVote(agent, proposal)
	if vote {
		t.Error("conservative conformist should reject regulatory disruption")
	}
}

// TestAgentVote_GoalAlignment2OrMore: conformist with goals matching proposal text → vote=true.
func TestAgentVote_GoalAlignment2OrMore(t *testing.T) {
	v := NewVoter(nil, nil)
	// Goals contain words that also appear in the proposal description/rationale
	goals := []string{
		"increase market efficiency",   // "market" and "efficiency" appear
		"reduce compliance overhead",   // "compliance" and "reduce" appear
	}
	proposal := RuleProposal{
		OriginalRuleID:  "rule-1",
		NewDescription:  "reduce market friction and improve efficiency",
		NewDomain:       DomainMarket,
		NewStability:    0.6,
		Rationale:       "lower compliance burden",
		ProposedByAgent: "d1",
	}
	agent := makeConformistWithTraits("goal-aligned", 0.5, nil, goals)

	vote, _ := v.agentVote(agent, proposal)
	if !vote {
		t.Error("conformist with 2+ goal alignments should vote YES")
	}
}

// TestAgentVote_NoGoalAlignment: conformist with goals unrelated to proposal → vote=false.
func TestAgentVote_NoGoalAlignment(t *testing.T) {
	v := NewVoter(nil, nil)
	goals := []string{
		"preserve ancient traditions",
		"protect endangered species",
	}
	proposal := simpleProposal("accelerate digital transformation via cloud infrastructure", DomainTechnology, 0.7)
	agent := makeConformistWithTraits("unaligned", 0.5, nil, goals)

	vote, _ := v.agentVote(agent, proposal)
	if vote {
		t.Error("conformist with no goal alignment should vote NO")
	}
}

// ─── Vote (full weighted vote) ────────────────────────────────────────────────

// TestVote_MajorityAccepts: all agents vote YES → accepted=true.
func TestVote_MajorityAccepts(t *testing.T) {
	agents := []Agent{
		makeDisruptorAgent("d1", 1.0),
		makeDisruptorAgent("d2", 1.0),
		makeDisruptorAgent("d3", 1.0),
	}
	v := NewVoter(agents, nil)
	proposal := simpleProposal("expand markets", DomainMarket, 0.5)

	accepted, records := v.Vote(context.Background(), proposal, nil)
	if !accepted {
		t.Error("all disruptors vote YES — should be accepted")
	}
	if len(records) != len(agents) {
		t.Errorf("expected %d vote records, got %d", len(agents), len(records))
	}
}

// TestVote_MajorityRejects: all risk-averse conformists see no goal alignment → rejected.
func TestVote_MajorityRejects(t *testing.T) {
	agents := []Agent{
		makeConformistWithTraits("c1", 1.0, []string{"risk-averse"}, []string{"preserve ancient traditions"}),
		makeConformistWithTraits("c2", 1.0, []string{"conservative"}, []string{"protect local economy"}),
		makeConformistWithTraits("c3", 1.0, []string{"process-driven"}, []string{"maintain bureaucratic order"}),
	}
	v := NewVoter(agents, nil)
	// Proposal in regulation domain — conservative/risk-averse conformists will reject
	proposal := simpleProposal("overhaul regulatory framework", DomainRegulation, 0.5)

	accepted, _ := v.Vote(context.Background(), proposal, nil)
	if accepted {
		t.Error("all risk-averse conformists facing regulatory change should reject — should not be accepted")
	}
}

// TestVote_WeightedVoteCorrect: one high-weight agent (pw=10) votes YES, many low-weight (pw=0.1) vote NO → accepted=true.
func TestVote_WeightedVoteCorrect(t *testing.T) {
	// High-power disruptor: votes YES
	highPower := makeDisruptorAgent("d-high", 10.0)

	// 5 low-power conformists that will reject: no goal alignment, non-regulatory, non-fragile
	// so they fall through to "no goal alignment → NO"
	lowPowerAgents := make([]Agent, 5)
	for i := 0; i < 5; i++ {
		lowPowerAgents[i] = makeConformistWithTraits(
			"c-low-"+string(rune('0'+i)),
			0.1,
			nil,
			[]string{"preserve ancient traditions"},
		)
	}

	agents := append([]Agent{highPower}, lowPowerAgents...)
	v := NewVoter(agents, nil)

	proposal := simpleProposal("accelerate digital cloud transformation", DomainTechnology, 0.7)
	accepted, records := v.Vote(context.Background(), proposal, nil)

	// Verify: high-power agent voted YES
	var highRecord VoteRecord
	for _, r := range records {
		if r.AgentID == "d-high" {
			highRecord = r
			break
		}
	}
	if !highRecord.Vote {
		t.Error("high-power disruptor should vote YES")
	}

	// yesWeight = 10.0, totalWeight = 10.0 + 5*0.1 = 10.5 → 10/10.5 ≈ 0.952 > 0.5
	if !accepted {
		t.Error("high-weight YES should outweigh many low-weight NO votes")
	}
}

// TestVote_LLMVoteLLMPathCalled: with mockLLM returning "ACCEPT reason",
// agent with pw≥0.85 gets [LLM] rationale in vote record.
func TestVote_LLMVoteLLMPathCalled(t *testing.T) {
	llm := &mockLLM{response: "ACCEPT because it improves efficiency"}
	highPower := makeTestAgent("llm-agent", 0.9) // pw=0.9 ≥ llmVotingThreshold

	agents := []Agent{highPower}
	v := NewVoter(agents, llm)

	proposal := simpleProposal("improve market efficiency", DomainMarket, 0.6)
	_, records := v.Vote(context.Background(), proposal, nil)

	if len(records) == 0 {
		t.Fatal("expected at least one vote record")
	}

	rec := records[0]
	if !strings.HasPrefix(rec.Rationale, "[LLM]") {
		t.Errorf("expected LLM rationale to start with '[LLM]', got: %q", rec.Rationale)
	}
	if llm.calls == 0 {
		t.Error("expected LLM to be called at least once for high-power agent")
	}
}
