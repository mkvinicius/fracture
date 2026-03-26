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
	if len(conformists) < 37 {
		t.Errorf("expected at least 37 conformist archetypes, got %d", len(conformists))
	}
}

func TestBuiltinDisruptorsCount(t *testing.T) {
	disruptors := BuiltinDisruptors(&mockLLM{})
	if len(disruptors) < 19 {
		t.Errorf("expected at least 19 disruptor archetypes, got %d", len(disruptors))
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
	all := append(BuiltinConformists(&mockLLM{}), BuiltinDisruptors(&mockLLM{})...)
	seen := make(map[string]bool)
	for _, a := range all {
		name := a.Personality().Name
		if seen[name] {
			t.Errorf("duplicate archetype name: %q", name)
		}
		seen[name] = true
	}
}

func TestTotalAgentCount(t *testing.T) {
	total := len(BuiltinConformists(&mockLLM{})) + len(BuiltinDisruptors(&mockLLM{}))
	if total < 56 {
		t.Errorf("expected at least 56 total agents, got %d", total)
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
