package memory

import (
	"database/sql"
	"os"
	"testing"

	"github.com/fracture/fracture/engine"
	_ "github.com/glebarez/go-sqlite3"
)

// openTestDB creates a temp SQLite DB with the minimal schema needed by Store.
func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "memory-test-*.db")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	f.Close()

	db, err := sql.Open("sqlite3", f.Name()+"?_foreign_keys=off")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	// Minimal schema — only what Store reads/writes.
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS simulation_rounds (
			id                TEXT PRIMARY KEY,
			simulation_id     TEXT NOT NULL,
			round_number      INTEGER NOT NULL,
			agent_id          TEXT NOT NULL,
			agent_type        TEXT NOT NULL,
			action_text       TEXT NOT NULL DEFAULT '',
			tension_level     REAL NOT NULL DEFAULT 0.0,
			fracture_proposed INTEGER NOT NULL DEFAULT 0,
			fracture_accepted INTEGER,
			new_rule_json     TEXT,
			tokens_used       INTEGER NOT NULL DEFAULT 0,
			created_at        INTEGER NOT NULL DEFAULT (unixepoch())
		)
	`)
	if err != nil {
		t.Fatalf("create table: %v", err)
	}

	t.Cleanup(func() { db.Close() })
	return db
}

func TestSaveAndLoadMemory(t *testing.T) {
	db := openTestDB(t)
	store := NewStore(db)

	const simID = "sim-001"
	const agentID = "agent-abc"

	action := engine.AgentAction{
		AgentID:   agentID,
		AgentType: engine.AgentConformist,
		Text:      "The market looks stable with mild competitive pressure.",
		TokensUsed: 42,
	}

	// Save the round
	if err := store.SaveRound(simID, 1, action, 0.35); err != nil {
		t.Fatalf("SaveRound: %v", err)
	}

	// Retrieve via RecentActions
	actions := store.RecentActions(agentID, 5)
	if len(actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(actions))
	}
	got := actions[0]
	if got.Text != action.Text {
		t.Errorf("Text mismatch: got %q, want %q", got.Text, action.Text)
	}
	if got.AgentType != action.AgentType {
		t.Errorf("AgentType mismatch: got %q, want %q", got.AgentType, action.AgentType)
	}
	if got.TokensUsed != action.TokensUsed {
		t.Errorf("TokensUsed mismatch: got %d, want %d", got.TokensUsed, action.TokensUsed)
	}
}

func TestRecentActionsLimit(t *testing.T) {
	db := openTestDB(t)
	store := NewStore(db)

	const simID = "sim-002"
	const agentID = "agent-xyz"

	// Save 5 rounds
	for i := 1; i <= 5; i++ {
		action := engine.AgentAction{
			AgentID:   agentID,
			AgentType: engine.AgentDisruptor,
			Text:      "disruption observation",
		}
		if err := store.SaveRound(simID, i, action, 0.5); err != nil {
			t.Fatalf("SaveRound round %d: %v", i, err)
		}
	}

	// Ask for only 3
	actions := store.RecentActions(agentID, 3)
	if len(actions) != 3 {
		t.Errorf("expected 3 actions (limit), got %d", len(actions))
	}
}

func TestRecentActionsUnknownAgent(t *testing.T) {
	db := openTestDB(t)
	store := NewStore(db)

	actions := store.RecentActions("nonexistent-agent", 10)
	if actions != nil && len(actions) != 0 {
		t.Errorf("expected empty result for unknown agent, got %d actions", len(actions))
	}
}
