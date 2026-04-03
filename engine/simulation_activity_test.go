package engine

import (
	"context"
	"testing"
	"time"
)

// ─── helpers ────────────────────────────────────────────────────────────────

// disruptorTestAgent is a minimal Agent with AgentDisruptor type for archetype tests.
type disruptorTestAgent struct {
	BaseAgent
}

func (d *disruptorTestAgent) React(_ context.Context, _ *World, _ AgentMemory, _ int, _ float64) (AgentAction, error) {
	return AgentAction{AgentID: d.ID()}, nil
}

func makeDisruptorTestAgent(id string, pw float64) Agent {
	p := Personality{Name: id, PowerWeight: pw}
	ba := NewBaseAgent(id, AgentDisruptor, DisruptorPermissions, p)
	return &disruptorTestAgent{BaseAgent: ba}
}

// drainActivityChannel collects all ActivityEvents available within timeout d.
func drainActivityChannel(ch chan ActivityEvent, d time.Duration) []ActivityEvent {
	var events []ActivityEvent
	deadline := time.After(d)
	for {
		select {
		case ev, ok := <-ch:
			if !ok {
				return events
			}
			events = append(events, ev)
		case <-deadline:
			return events
		}
	}
}

// runSimAndDrain runs a simulation to completion, draining the RoundResult
// channel, and then drains the activity bus channel with the given timeout.
func runSimAndDrain(t *testing.T, sim *Simulation, ch chan ActivityEvent, busTimeout time.Duration) []ActivityEvent {
	t.Helper()
	ctx := context.Background()
	rounds := sim.Run(ctx)
	for range rounds {
		// drain so the simulation can advance
	}
	return drainActivityChannel(ch, busTimeout)
}

// ─── 1. TestSimulationEmitsActivityEvents ────────────────────────────────────

func TestSimulationEmitsActivityEvents(t *testing.T) {
	const simID = "sim-act-001"
	const numAgents = 3
	const maxRounds = 2

	agents := []Agent{
		makeTestAgent("act-agent-1", 0.5),
		makeTestAgent("act-agent-2", 0.6),
		makeTestAgent("act-agent-3", 0.7),
	}

	world := DefaultWorldForDomain(DomainMarket, "test", "")
	cfg := SimulationConfig{
		ID:        simID,
		Question:  "activity events test",
		MaxRounds: maxRounds,
		Agents:    agents,
		World:     world,
		Memory:    nil,
		CouncilLLM: nil,
		VotingLLM:  nil,
		Mode:      DefaultConfigForMode(ModeStandard),
	}

	// Subscribe BEFORE running the simulation.
	ch := GlobalActivityBus.Subscribe(simID)
	t.Cleanup(func() {
		GlobalActivityBus.Unsubscribe(simID, ch)
	})

	sim := NewSimulation(cfg)
	events := runSimAndDrain(t, sim, ch, 2*time.Second)

	// At least numAgents * maxRounds events must be emitted (one per agent per round).
	minExpected := numAgents * maxRounds
	if len(events) < minExpected {
		t.Errorf("expected at least %d events, got %d", minExpected, len(events))
	}

	for i, ev := range events {
		if ev.SimulationID != simID {
			t.Errorf("event[%d]: SimulationID = %q; want %q", i, ev.SimulationID, simID)
		}
		if ev.Ts == 0 {
			t.Errorf("event[%d]: Ts == 0, expected timestamp to be set", i)
		}
	}
}

// ─── 2. TestSimulationActivityArchetypeTag ───────────────────────────────────

func TestSimulationActivityArchetypeTag(t *testing.T) {
	const simID = "sim-act-002"

	conformistAgent := makeTestAgent("act-conformist-1", 0.5)
	disruptorAgent := makeDisruptorTestAgent("act-disruptor-1", 0.5)

	agents := []Agent{conformistAgent, disruptorAgent}
	world := DefaultWorldForDomain(DomainMarket, "archetype tag test", "")
	cfg := SimulationConfig{
		ID:        simID,
		Question:  "archetype tag test",
		MaxRounds: 1,
		Agents:    agents,
		World:     world,
		Memory:    nil,
		CouncilLLM: nil,
		VotingLLM:  nil,
		Mode:      DefaultConfigForMode(ModeStandard),
	}

	ch := GlobalActivityBus.Subscribe(simID)
	t.Cleanup(func() {
		GlobalActivityBus.Unsubscribe(simID, ch)
	})

	sim := NewSimulation(cfg)
	events := runSimAndDrain(t, sim, ch, 2*time.Second)

	if len(events) == 0 {
		t.Fatal("expected at least 1 activity event, got none")
	}

	// Build a map of agentID -> archetype from events.
	archetypeByAgent := make(map[string]string)
	for _, ev := range events {
		if ev.ActionType == "react" || ev.ActionType == "propose" {
			archetypeByAgent[ev.AgentID] = ev.Archetype
		}
	}

	if arch, ok := archetypeByAgent[conformistAgent.ID()]; ok {
		if arch != "conformist" {
			t.Errorf("conformist agent archetype = %q; want \"conformist\"", arch)
		}
	} else {
		t.Errorf("no event found for conformist agent %q", conformistAgent.ID())
	}

	if arch, ok := archetypeByAgent[disruptorAgent.ID()]; ok {
		if arch != "disruptor" {
			t.Errorf("disruptor agent archetype = %q; want \"disruptor\"", arch)
		}
	} else {
		t.Errorf("no event found for disruptor agent %q", disruptorAgent.ID())
	}
}

// ─── 3. TestSimulationActivityTotalTokensMonotonic ───────────────────────────

func TestSimulationActivityTotalTokensMonotonic(t *testing.T) {
	const simID = "sim-act-003"

	agents := []Agent{
		makeTestAgent("act-tok-1", 0.5),
		makeTestAgent("act-tok-2", 0.6),
	}
	world := DefaultWorldForDomain(DomainMarket, "token monotonic test", "")
	cfg := SimulationConfig{
		ID:        simID,
		Question:  "token monotonic test",
		MaxRounds: 2,
		Agents:    agents,
		World:     world,
		Memory:    nil,
		CouncilLLM: nil,
		VotingLLM:  nil,
		Mode:      DefaultConfigForMode(ModeStandard),
	}

	ch := GlobalActivityBus.Subscribe(simID)
	t.Cleanup(func() {
		GlobalActivityBus.Unsubscribe(simID, ch)
	})

	// Drain the round channel first, then the activity channel.
	ctx := context.Background()
	sim := NewSimulation(cfg)
	rounds := sim.Run(ctx)
	for range rounds {
		// drain rounds so the simulation can complete
	}

	events := drainActivityChannel(ch, 2*time.Second)

	if len(events) < 2 {
		t.Fatalf("expected at least 2 events to check monotonicity, got %d", len(events))
	}

	// TotalTokens should be monotonically non-decreasing.
	for i := 1; i < len(events); i++ {
		if events[i].TotalTokens < events[i-1].TotalTokens {
			t.Errorf(
				"TotalTokens not monotonic at index %d: events[%d].TotalTokens=%d < events[%d].TotalTokens=%d",
				i, i, events[i].TotalTokens, i-1, events[i-1].TotalTokens,
			)
		}
	}
}

// ─── 4. TestActivitySnippetTruncation ────────────────────────────────────────
// Tests the helper directly: short text, truncation with ellipsis, and empty.

func TestActivitySnippetTruncation(t *testing.T) {
	tests := []struct {
		text string
		n    int
		want string
	}{
		{"hello", 10, "hello"},
		{"hello world", 5, "hello…"},
		{"", 10, ""},
	}
	for _, tc := range tests {
		got := activitySnippet(tc.text, tc.n)
		if got != tc.want {
			t.Errorf("activitySnippet(%q, %d) = %q; want %q", tc.text, tc.n, got, tc.want)
		}
	}
}

// ─── 5. TestFirstRuleKeyNilAndEmptyMap ───────────────────────────────────────
// Covers the nil-map and empty-map cases for firstRuleKey.

func TestFirstRuleKeyNilAndEmptyMap(t *testing.T) {
	if got := firstRuleKey(nil); got != "" {
		t.Errorf("firstRuleKey(nil) = %q; want \"\"", got)
	}
	if got := firstRuleKey(map[string]float64{}); got != "" {
		t.Errorf("firstRuleKey(empty map) = %q; want \"\"", got)
	}
}

// ─── 6. TestFirstRuleKeySingleEntry ──────────────────────────────────────────
// Covers a single-entry map; the only key must be returned.

func TestFirstRuleKeySingleEntry(t *testing.T) {
	m := map[string]float64{"rule-1": 0.3}
	got := firstRuleKey(m)
	if got != "rule-1" {
		t.Errorf("firstRuleKey(map{\"rule-1\": 0.3}) = %q; want \"rule-1\"", got)
	}
}
