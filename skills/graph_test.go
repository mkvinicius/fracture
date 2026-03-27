package skills_test

import (
	"testing"

	"github.com/fracture/fracture/skills"
)

func TestGraphRegistry(t *testing.T) {
	// Verifica que os grafos foram registrados
	for _, id := range []string{
		"manufacturing", "logistics", "healthcare",
		"fintech", "education", "agro",
	} {
		if _, ok := skills.Graphs[id]; !ok {
			t.Errorf("graph not registered: %s", id)
		}
	}
}

func TestGetRelations(t *testing.T) {
	rels := skills.GetRelations("manufacturing", "Taiichi Ohno")
	if len(rels) == 0 {
		t.Error("expected relations for Taiichi Ohno")
	}
}

func TestFormatRelationsContext(t *testing.T) {
	ctx := skills.FormatRelationsContext(
		"manufacturing",
		[]string{"Taiichi Ohno", "W. Edwards Deming"},
	)
	if ctx == "" {
		t.Error("expected non-empty context")
	}
	if len(ctx) < 50 {
		t.Error("context too short")
	}
}

func TestFormatRelationsContextUnknownSkill(t *testing.T) {
	ctx := skills.FormatRelationsContext("unknown", []string{"anyone"})
	if ctx != "" {
		t.Error("expected empty context for unknown skill")
	}
}
