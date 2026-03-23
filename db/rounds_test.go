package db

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

// ─── Rounds tests ─────────────────────────────────────────────────────────────

func TestSaveAndListRounds(t *testing.T) {
	d := openTestDB(t)

	// First insert a parent simulation (FK constraint)
	simID := uuid.New().String()
	if err := d.SaveSimulation(simID, "test question", "market", 20, map[string]string{"ok": "true"}); err != nil {
		t.Fatalf("SaveSimulation: %v", err)
	}

	rows := []*RoundRow{
		{
			ID:               uuid.New().String(),
			SimulationID:     simID,
			RoundNumber:      1,
			AgentID:          "agent-1",
			AgentType:        "conformist",
			ActionText:       "Agent observes the market",
			TensionLevel:     0.3,
			FractureProposed: false,
			TokensUsed:       120,
			CreatedAt:        time.Now().Unix(),
		},
		{
			ID:               uuid.New().String(),
			SimulationID:     simID,
			RoundNumber:      1,
			AgentID:          "agent-2",
			AgentType:        "disruptor",
			ActionText:       "Agent proposes a new rule",
			TensionLevel:     0.3,
			FractureProposed: true,
			NewRuleJSON:      `{"rule_id":"r1","description":"new rule"}`,
			TokensUsed:       200,
			CreatedAt:        time.Now().Unix(),
		},
	}

	for _, r := range rows {
		if err := d.SaveRound(r); err != nil {
			t.Fatalf("SaveRound: %v", err)
		}
	}

	result, err := d.ListRounds(simID)
	if err != nil {
		t.Fatalf("ListRounds: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 rounds, got %d", len(result))
	}

	// Verify fields of second round
	r2 := result[1]
	if r2.AgentID != "agent-2" {
		t.Errorf("expected agent-2, got %q", r2.AgentID)
	}
	if !r2.FractureProposed {
		t.Error("expected FractureProposed=true")
	}
	if r2.NewRuleJSON == "" {
		t.Error("expected non-empty NewRuleJSON")
	}
}

func TestSaveRoundUpsertAccepted(t *testing.T) {
	d := openTestDB(t)

	simID := uuid.New().String()
	_ = d.SaveSimulation(simID, "q", "market", 10, nil)

	roundID := uuid.New().String()
	row := &RoundRow{
		ID:               roundID,
		SimulationID:     simID,
		RoundNumber:      2,
		AgentID:          "agent-x",
		AgentType:        "disruptor",
		ActionText:       "fracture proposal",
		TensionLevel:     0.8,
		FractureProposed: true,
		CreatedAt:        time.Now().Unix(),
	}
	if err := d.SaveRound(row); err != nil {
		t.Fatalf("SaveRound initial: %v", err)
	}

	// Now update with accepted=true
	accepted := true
	row.FractureAccepted = &accepted
	if err := d.SaveRound(row); err != nil {
		t.Fatalf("SaveRound upsert: %v", err)
	}

	result, _ := d.ListRounds(simID)
	if len(result) != 1 {
		t.Fatalf("expected 1 round, got %d", len(result))
	}
	if result[0].FractureAccepted == nil || !*result[0].FractureAccepted {
		t.Error("expected FractureAccepted=true after upsert")
	}
}

func TestListRoundsEmpty(t *testing.T) {
	d := openTestDB(t)
	simID := uuid.New().String()
	_ = d.SaveSimulation(simID, "q", "market", 5, nil)

	result, err := d.ListRounds(simID)
	if err != nil {
		t.Fatalf("ListRounds empty: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected 0 rounds, got %d", len(result))
	}
}

// ─── Votes tests ──────────────────────────────────────────────────────────────

func TestSaveAndListVotes(t *testing.T) {
	d := openTestDB(t)

	simID := uuid.New().String()
	_ = d.SaveSimulation(simID, "q", "market", 20, nil)

	votes := []*VoteRow{
		{
			ID:           uuid.New().String(),
			SimulationID: simID,
			RoundNumber:  3,
			ProposalID:   "prop-1",
			VoterID:      "voter-a",
			VoterType:    "Conformist CEO",
			Vote:         false,
			Weight:       1.0,
			Reasoning:    "Too risky",
			CreatedAt:    time.Now().Unix(),
		},
		{
			ID:           uuid.New().String(),
			SimulationID: simID,
			RoundNumber:  3,
			ProposalID:   "prop-1",
			VoterID:      "voter-b",
			VoterType:    "Disruptor Startup",
			Vote:         true,
			Weight:       0.8,
			Reasoning:    "Great opportunity",
			CreatedAt:    time.Now().Unix(),
		},
	}

	for _, v := range votes {
		if err := d.SaveVote(v); err != nil {
			t.Fatalf("SaveVote: %v", err)
		}
	}

	result, err := d.ListVotes(simID)
	if err != nil {
		t.Fatalf("ListVotes: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 votes, got %d", len(result))
	}

	// Verify vote values
	if result[0].Vote != false {
		t.Error("expected first vote=false")
	}
	if result[1].Vote != true {
		t.Error("expected second vote=true")
	}
	if result[1].Reasoning != "Great opportunity" {
		t.Errorf("unexpected reasoning: %q", result[1].Reasoning)
	}
}

func TestSaveVoteIdempotent(t *testing.T) {
	d := openTestDB(t)

	simID := uuid.New().String()
	_ = d.SaveSimulation(simID, "q", "market", 20, nil)

	voteID := uuid.New().String()
	v := &VoteRow{
		ID:           voteID,
		SimulationID: simID,
		RoundNumber:  1,
		ProposalID:   "prop-x",
		VoterID:      "voter-c",
		VoterType:    "Analyst",
		Vote:         true,
		Weight:       1.0,
		CreatedAt:    time.Now().Unix(),
	}
	_ = d.SaveVote(v)
	// Second insert with same ID should be ignored (INSERT OR IGNORE)
	if err := d.SaveVote(v); err != nil {
		t.Fatalf("SaveVote idempotent: %v", err)
	}

	result, _ := d.ListVotes(simID)
	if len(result) != 1 {
		t.Errorf("expected 1 vote after idempotent insert, got %d", len(result))
	}
}

// ─── Report generations tests ─────────────────────────────────────────────────

func TestStartAndCompleteReportGen(t *testing.T) {
	d := openTestDB(t)

	simID := uuid.New().String()
	_ = d.SaveSimulation(simID, "q", "market", 20, nil)

	genID := uuid.New().String()
	if err := d.StartReportGen(genID, simID, "probable_future"); err != nil {
		t.Fatalf("StartReportGen: %v", err)
	}

	if err := d.CompleteReportGen(genID, "done", "", 350, 4200); err != nil {
		t.Fatalf("CompleteReportGen: %v", err)
	}

	result, err := d.ListReportGens(simID)
	if err != nil {
		t.Fatalf("ListReportGens: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 report gen, got %d", len(result))
	}

	rg := result[0]
	if rg.Status != "done" {
		t.Errorf("expected status=done, got %q", rg.Status)
	}
	if rg.TokensUsed != 350 {
		t.Errorf("expected 350 tokens, got %d", rg.TokensUsed)
	}
	if rg.DurationMs != 4200 {
		t.Errorf("expected 4200ms, got %d", rg.DurationMs)
	}
	if rg.CompletedAt == nil {
		t.Error("expected non-nil CompletedAt")
	}
}

func TestCompleteReportGenError(t *testing.T) {
	d := openTestDB(t)

	simID := uuid.New().String()
	_ = d.SaveSimulation(simID, "q", "market", 20, nil)

	genID := uuid.New().String()
	_ = d.StartReportGen(genID, simID, "rupture_scenarios")
	_ = d.CompleteReportGen(genID, "error", "LLM timeout", 0, 0)

	result, _ := d.ListReportGens(simID)
	if len(result) != 1 {
		t.Fatalf("expected 1 report gen, got %d", len(result))
	}
	if result[0].ErrorMsg != "LLM timeout" {
		t.Errorf("expected error msg 'LLM timeout', got %q", result[0].ErrorMsg)
	}
}
