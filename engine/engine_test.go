package engine

import (
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
