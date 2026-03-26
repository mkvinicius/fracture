package archetypes

import (
	"context"
	"testing"

	"github.com/fracture/fracture/engine"
)


// mockLLM is a minimal LLMCaller for testing (no real API calls).
type mockLLM struct{}

func (m *mockLLM) Call(ctx context.Context, systemPrompt, userPrompt string, maxTokens int) (string, int, error) {
	return `{"action": "observe", "reasoning": "test", "tension_delta": {}}`, 10, nil
}

func TestBuiltinConformistsCount(t *testing.T) {
	conformists := BuiltinConformists(&mockLLM{})
	if len(conformists) < 59 {
		t.Errorf("expected at least 59 conformist archetypes, got %d", len(conformists))
	}
}

func TestBuiltinDisruptorsCount(t *testing.T) {
	disruptors := BuiltinDisruptors(&mockLLM{})
	if len(disruptors) < 28 {
		t.Errorf("expected at least 28 disruptor archetypes, got %d", len(disruptors))
	}
}

func TestArchetypePersonalityFields(t *testing.T) {
	all := append(BuiltinConformists(&mockLLM{}), BuiltinDisruptors(&mockLLM{})...)
	for _, a := range all {
		p := a.Personality()
		if p.Name == "" {
			t.Errorf("archetype has empty Name: %+v", p)
		}
		if p.Role == "" {
			t.Errorf("archetype %q has empty Role", p.Name)
		}
		if p.PowerWeight < 0 || p.PowerWeight > 1 {
			t.Errorf("archetype %q has invalid PowerWeight %f (must be 0.0-1.0)", p.Name, p.PowerWeight)
		}
		if len(p.Traits) == 0 {
			t.Errorf("archetype %q has no Traits", p.Name)
		}
	}
}

func TestNoDuplicateArchetypeNames(t *testing.T) {
	// Check no duplicates within each pool independently.
	// The same historical figure may intentionally appear in both conformist and
	// disruptor pools with different roles (e.g. Adam Smith, Karl Marx).
	pools := [][]engine.Agent{
		BuiltinConformists(&mockLLM{}),
		BuiltinDisruptors(&mockLLM{}),
	}
	for _, pool := range pools {
		seen := make(map[string]bool)
		for _, a := range pool {
			name := a.Personality().Name
			if seen[name] {
				t.Errorf("duplicate archetype name within pool: %q", name)
			}
			seen[name] = true
		}
	}
}

func TestTotalAgentCount(t *testing.T) {
	total := len(BuiltinConformists(&mockLLM{})) + len(BuiltinDisruptors(&mockLLM{}))
	if total < 87 {
		t.Errorf("expected at least 87 total agents, got %d", total)
	}
}

func TestAgentTypes(t *testing.T) {
	for _, a := range BuiltinConformists(&mockLLM{}) {
		if a.Type() != engine.AgentConformist {
			t.Errorf("conformist %q has wrong type: %q", a.Personality().Name, a.Type())
		}
	}
	for _, a := range BuiltinDisruptors(&mockLLM{}) {
		if a.Type() != engine.AgentDisruptor {
			t.Errorf("disruptor %q has wrong type: %q", a.Personality().Name, a.Type())
		}
	}
}
