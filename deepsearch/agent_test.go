package deepsearch

import (
	"context"
	"testing"
)

// mockLLM is a no-op LLMCaller for tests that don't need real responses.
type mockLLM struct{}

func (m *mockLLM) Call(_ context.Context, _, _ string, _ int) (string, int, error) {
	return "", 0, nil
}

// TestHashQuestion verifies that fallbackQueries is deterministic:
// same inputs always produce identical outputs (no randomness).
func TestHashQuestion(t *testing.T) {
	a := New(&mockLLM{}, DefaultConfig())
	const (
		question = "Will AI disrupt the SaaS market?"
		company  = "Acme Corp"
		sector   = "software"
	)
	first := a.fallbackQueries(question, company, sector)
	for i := 0; i < 5; i++ {
		got := a.fallbackQueries(question, company, sector)
		if len(got) != len(first) {
			t.Fatalf("run %d: got %d queries, want %d", i+1, len(got), len(first))
		}
		for j, q := range got {
			if q != first[j] {
				t.Errorf("run %d query[%d]: got %q, want %q", i+1, j, q, first[j])
			}
		}
	}
}

// TestFallbackQueries verifies that fallbackQueries returns exactly 4 non-empty queries.
func TestFallbackQueries(t *testing.T) {
	a := New(&mockLLM{}, DefaultConfig())
	queries := a.fallbackQueries("How will regulation affect fintech?", "PayCo", "fintech")

	if len(queries) != 4 {
		t.Errorf("expected 4 queries, got %d", len(queries))
	}
	for i, q := range queries {
		if q == "" {
			t.Errorf("query[%d] is empty", i)
		}
	}
}
