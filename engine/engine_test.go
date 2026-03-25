package engine

import (
	"context"
	"testing"
)

func TestSimulationConfigDefaults(t *testing.T) {
	cfg := SimulationConfig{}
	if cfg.MaxRounds == 0 {
		cfg.MaxRounds = 40
	}
	if cfg.MaxRounds != 40 {
		t.Errorf("default MaxRounds should be 40, got %d", cfg.MaxRounds)
	}
}

func TestDefaultWorldForDomainNotNil(t *testing.T) {
	domains := []RuleDomain{
		DomainMarket, DomainTechnology, DomainRegulation,
		DomainBehavior, DomainCulture, DomainGeopolitics, DomainFinance,
	}
	for _, domain := range domains {
		world := DefaultWorldForDomain(domain, "test question", "")
		if world == nil {
			t.Errorf("DefaultWorldForDomain(%q) returned nil", domain)
			continue
		}
		if len(world.Rules) == 0 {
			t.Errorf("DefaultWorldForDomain(%q) returned world with no rules", domain)
		}
	}
}

func TestWorldRulesMinimum(t *testing.T) {
	world := DefaultWorldForDomain(DomainTechnology, "test", "")
	if len(world.Rules) < 5 {
		t.Errorf("Technology domain should have at least 5 rules, got %d", len(world.Rules))
	}
}

func TestWorldRuleFields(t *testing.T) {
	world := DefaultWorldForDomain(DomainMarket, "test question", "")
	for id, r := range world.Rules {
		if r.ID == "" {
			t.Errorf("rule %q has empty ID", id)
		}
		if r.Description == "" {
			t.Errorf("rule %q has empty Description", id)
		}
		if r.Stability < 0 || r.Stability > 1 {
			t.Errorf("rule %q has invalid Stability %f (must be 0.0-1.0)", id, r.Stability)
		}
	}
}

func TestWorldWithExtraContext(t *testing.T) {
	extraCtx := "This company operates in the Brazilian fintech market with 50 employees."
	world := DefaultWorldForDomain(DomainFinance, "test question", extraCtx)
	if world == nil {
		t.Fatal("DefaultWorldForDomain returned nil with extra context")
	}
	if len(world.Rules) == 0 {
		t.Error("World should have at least one rule when extra context is provided")
	}
}

func TestNewWorldCreation(t *testing.T) {
	rules := []*Rule{
		{ID: "r1", Description: "Test rule 1", Domain: DomainMarket, Stability: 0.7, IsActive: true},
		{ID: "r2", Description: "Test rule 2", Domain: DomainTechnology, Stability: 0.5, IsActive: true},
	}
	world := NewWorld(rules)
	if world == nil {
		t.Fatal("NewWorld returned nil")
	}
	if len(world.Rules) != 2 {
		t.Errorf("expected 2 rules, got %d", len(world.Rules))
	}
	for _, r := range rules {
		if _, ok := world.Rules[r.ID]; !ok {
			t.Errorf("rule %q not found in world", r.ID)
		}
		if world.TensionMap[r.ID] != 0.0 {
			t.Errorf("initial tension for rule %q should be 0.0", r.ID)
		}
	}
}

// ─── ApplyCalibration tests ───────────────────────────────────────────────────

type testAgent struct {
	BaseAgent
}

func (t *testAgent) React(_ context.Context, _ *World, _ AgentMemory, _ int, _ float64) (AgentAction, error) {
	return AgentAction{AgentID: t.ID()}, nil
}

func makeTestAgent(id string, pw float64) Agent {
	p := Personality{Name: id, PowerWeight: pw}
	ba := NewBaseAgent(id, AgentConformist, ConformistPermissions, p)
	return &testAgent{BaseAgent: ba}
}

func TestApplyCalibrationScalesPowerWeight(t *testing.T) {
	agents := []Agent{
		makeTestAgent("agent-1", 1.0),
		makeTestAgent("agent-2", 1.0),
	}
	cals := []AgentCalibration{
		{AgentID: "agent-1", AccuracyWeight: 2.0},
	}
	result := ApplyCalibration(agents, cals)
	if len(result) != 2 {
		t.Fatalf("expected 2 agents, got %d", len(result))
	}
	// agent-1 should have PowerWeight = 1.0 * 2.0 = 2.0
	if pw := result[0].Personality().PowerWeight; pw != 2.0 {
		t.Errorf("agent-1 PowerWeight: expected 2.0, got %.2f", pw)
	}
	// agent-2 unchanged
	if pw := result[1].Personality().PowerWeight; pw != 1.0 {
		t.Errorf("agent-2 PowerWeight: expected 1.0 (unchanged), got %.2f", pw)
	}
}

func TestApplyCalibrationClampsHigh(t *testing.T) {
	agents := []Agent{makeTestAgent("a", 5.0)}
	cals := []AgentCalibration{{AgentID: "a", AccuracyWeight: 10.0}}
	result := ApplyCalibration(agents, cals)
	pw := result[0].Personality().PowerWeight
	if pw > 10.0 {
		t.Errorf("PowerWeight %f should be clamped to ≤10.0", pw)
	}
}

func TestApplyCalibrationClampsLow(t *testing.T) {
	agents := []Agent{makeTestAgent("a", 1.0)}
	cals := []AgentCalibration{{AgentID: "a", AccuracyWeight: 0.0}}
	result := ApplyCalibration(agents, cals)
	pw := result[0].Personality().PowerWeight
	if pw < 0.1 {
		t.Errorf("PowerWeight %f should be clamped to ≥0.1", pw)
	}
}

func TestApplyCalibrationEmptyNoChange(t *testing.T) {
	agents := []Agent{makeTestAgent("x", 1.5)}
	result := ApplyCalibration(agents, nil)
	if result[0].Personality().PowerWeight != 1.5 {
		t.Errorf("empty calibrations should leave PowerWeight unchanged")
	}
}

func TestApplyCalibrationNeutralNoWrap(t *testing.T) {
	agents := []Agent{makeTestAgent("y", 0.7)}
	cals := []AgentCalibration{{AgentID: "y", AccuracyWeight: 1.0}}
	result := ApplyCalibration(agents, cals)
	// weight=1.0 means no-op — original agent returned
	if result[0].Personality().PowerWeight != 0.7 {
		t.Errorf("AccuracyWeight=1.0 should return original agent unchanged, got %.2f", result[0].Personality().PowerWeight)
	}
}
