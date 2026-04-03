package engine

import (
	"math"
	"testing"
)

const cascadeEpsilon = 1e-9

// helpers

func makeWorld(rules []*Rule) *World {
	w := NewWorld(rules)
	return w
}

func setTension(w *World, ruleID string, val float64) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.TensionMap[ruleID] = val
}

func getTension(w *World, ruleID string) float64 {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.TensionMap[ruleID]
}

// TestCascadeNoCascade_NoDependents: rule with no dependents — only its own tension increases.
func TestCascadeNoCascade_NoDependents(t *testing.T) {
	rules := []*Rule{
		{ID: "A", Description: "Rule A", Domain: DomainMarket, Stability: 0.5, IsActive: true, DependsOn: nil},
		{ID: "B", Description: "Rule B", Domain: DomainMarket, Stability: 0.5, IsActive: true, DependsOn: nil},
	}
	w := makeWorld(rules)

	w.IncreaseTension("A", 0.3)

	if got := getTension(w, "A"); math.Abs(got-0.3) > cascadeEpsilon {
		t.Errorf("expected A tension=0.3, got %f", got)
	}
	if got := getTension(w, "B"); math.Abs(got-0.0) > cascadeEpsilon {
		t.Errorf("expected B tension=0.0 (no cascade), got %f", got)
	}
}

// TestCascadePropagates30Percent: rule B depends on A; IncreaseTension(A, 0.4) → B gets +0.12.
func TestCascadePropagates30Percent(t *testing.T) {
	rules := []*Rule{
		{ID: "A", Description: "Rule A", Domain: DomainMarket, Stability: 0.5, IsActive: true, DependsOn: nil},
		{ID: "B", Description: "Rule B", Domain: DomainMarket, Stability: 0.5, IsActive: true, DependsOn: []string{"A"}},
	}
	w := makeWorld(rules)

	w.IncreaseTension("A", 0.4)

	wantA := 0.4
	wantB := 0.4 * 0.30 // 0.12

	if got := getTension(w, "A"); math.Abs(got-wantA) > cascadeEpsilon {
		t.Errorf("expected A tension=%.4f, got %.10f", wantA, got)
	}
	if got := getTension(w, "B"); math.Abs(got-wantB) > cascadeEpsilon {
		t.Errorf("expected B tension=%.4f (30%% of 0.4), got %.10f", wantB, got)
	}
}

// TestCascadeMultipleDependents: rules B and C both depend on A; both get 30% cascade.
func TestCascadeMultipleDependents(t *testing.T) {
	rules := []*Rule{
		{ID: "A", Description: "Rule A", Domain: DomainMarket, Stability: 0.5, IsActive: true, DependsOn: nil},
		{ID: "B", Description: "Rule B", Domain: DomainMarket, Stability: 0.5, IsActive: true, DependsOn: []string{"A"}},
		{ID: "C", Description: "Rule C", Domain: DomainMarket, Stability: 0.5, IsActive: true, DependsOn: []string{"A"}},
	}
	w := makeWorld(rules)

	delta := 0.5
	w.IncreaseTension("A", delta)

	wantCascade := delta * 0.30

	if got := getTension(w, "B"); math.Abs(got-wantCascade) > cascadeEpsilon {
		t.Errorf("expected B tension=%.4f, got %.10f", wantCascade, got)
	}
	if got := getTension(w, "C"); math.Abs(got-wantCascade) > cascadeEpsilon {
		t.Errorf("expected C tension=%.4f, got %.10f", wantCascade, got)
	}
}

// TestCascadeClampedAt1: A at 0.9, B (depends on A) at 0.95; IncreaseTension(A, 0.5) → B clamped to 1.0.
func TestCascadeClampedAt1(t *testing.T) {
	rules := []*Rule{
		{ID: "A", Description: "Rule A", Domain: DomainMarket, Stability: 0.5, IsActive: true, DependsOn: nil},
		{ID: "B", Description: "Rule B", Domain: DomainMarket, Stability: 0.5, IsActive: true, DependsOn: []string{"A"}},
	}
	w := makeWorld(rules)
	setTension(w, "A", 0.9)
	setTension(w, "B", 0.95)

	w.IncreaseTension("A", 0.5)

	// A: min(1.0, 0.9+0.5) = 1.0
	if got := getTension(w, "A"); math.Abs(got-1.0) > cascadeEpsilon {
		t.Errorf("expected A tension=1.0 (clamped), got %f", got)
	}
	// B: min(1.0, 0.95 + 0.5*0.3) = min(1.0, 1.1) = 1.0
	if got := getTension(w, "B"); math.Abs(got-1.0) > cascadeEpsilon {
		t.Errorf("expected B tension=1.0 (clamped), got %f", got)
	}
}

// TestCascadeInactiveRuleSkipped: rule B is inactive and depends on A; B should NOT be cascaded.
func TestCascadeInactiveRuleSkipped(t *testing.T) {
	rules := []*Rule{
		{ID: "A", Description: "Rule A", Domain: DomainMarket, Stability: 0.5, IsActive: true, DependsOn: nil},
		{ID: "B", Description: "Rule B", Domain: DomainMarket, Stability: 0.5, IsActive: false, DependsOn: []string{"A"}},
	}
	w := makeWorld(rules)

	w.IncreaseTension("A", 0.4)

	// B is inactive — should remain at 0.0
	if got := getTension(w, "B"); math.Abs(got-0.0) > cascadeEpsilon {
		t.Errorf("expected B tension=0.0 (inactive, skipped), got %f", got)
	}
}

// TestCascadeSelfNotApplied: IncreaseTension(A) does not double-count A's own tension.
func TestCascadeSelfNotApplied(t *testing.T) {
	rules := []*Rule{
		// A lists itself in DependsOn — should still not double-apply
		{ID: "A", Description: "Rule A", Domain: DomainMarket, Stability: 0.5, IsActive: true, DependsOn: []string{"A"}},
	}
	w := makeWorld(rules)

	w.IncreaseTension("A", 0.4)

	// The cascade loop skips id == ruleID, so A should only get +0.4
	if got := getTension(w, "A"); math.Abs(got-0.4) > cascadeEpsilon {
		t.Errorf("expected A tension=0.4 (no self-cascade), got %f", got)
	}
}

// TestCascadeNotTransitive: A→B→C; IncreaseTension(A) cascades to B but NOT to C.
func TestCascadeNotTransitive(t *testing.T) {
	rules := []*Rule{
		{ID: "A", Description: "Rule A", Domain: DomainMarket, Stability: 0.5, IsActive: true, DependsOn: nil},
		{ID: "B", Description: "Rule B", Domain: DomainMarket, Stability: 0.5, IsActive: true, DependsOn: []string{"A"}},
		{ID: "C", Description: "Rule C", Domain: DomainMarket, Stability: 0.5, IsActive: true, DependsOn: []string{"B"}},
	}
	w := makeWorld(rules)

	delta := 0.4
	w.IncreaseTension("A", delta)

	wantB := delta * 0.30 // 0.12

	if got := getTension(w, "B"); math.Abs(got-wantB) > cascadeEpsilon {
		t.Errorf("expected B tension=%.4f, got %.10f", wantB, got)
	}
	// C depends on B (not A), so it should NOT receive cascade from IncreaseTension(A)
	if got := getTension(w, "C"); math.Abs(got-0.0) > cascadeEpsilon {
		t.Errorf("expected C tension=0.0 (cascade is one-hop only), got %f", got)
	}
}

// TestIncreaseTensionUnknownRuleNoOp: calling IncreaseTension for an unknown ruleID does nothing.
func TestIncreaseTensionUnknownRuleNoOp(t *testing.T) {
	rules := []*Rule{
		{ID: "A", Description: "Rule A", Domain: DomainMarket, Stability: 0.5, IsActive: true, DependsOn: nil},
	}
	w := makeWorld(rules)

	// Should not panic
	w.IncreaseTension("does-not-exist", 0.5)

	// A remains untouched
	if got := getTension(w, "A"); math.Abs(got-0.0) > cascadeEpsilon {
		t.Errorf("expected A tension=0.0 (unrelated rule), got %f", got)
	}
}
