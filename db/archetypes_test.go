package db

import (
	"testing"
)

// ─── Archetype CRUD tests ─────────────────────────────────────────────────────

func TestCreateAndGetArchetype(t *testing.T) {
	d := openTestDB(t)

	a := &ArchetypeRow{
		ID:           "test-arch-1",
		CompanyID:    "acme",
		Name:         "The Contrarian",
		AgentType:    "disruptor",
		Description:  "Challenges every assumption",
		MemoryWeight: 1.2,
		IsActive:     true,
	}
	if err := d.CreateArchetype(a); err != nil {
		t.Fatalf("CreateArchetype: %v", err)
	}

	got, err := d.GetArchetype("test-arch-1")
	if err != nil {
		t.Fatalf("GetArchetype: %v", err)
	}
	if got.Name != a.Name {
		t.Errorf("Name: want %q, got %q", a.Name, got.Name)
	}
	if got.AgentType != a.AgentType {
		t.Errorf("AgentType: want %q, got %q", a.AgentType, got.AgentType)
	}
	if got.MemoryWeight != a.MemoryWeight {
		t.Errorf("MemoryWeight: want %v, got %v", a.MemoryWeight, got.MemoryWeight)
	}
	if !got.IsActive {
		t.Error("IsActive should be true")
	}
}

func TestUpdateArchetype(t *testing.T) {
	d := openTestDB(t)

	a := &ArchetypeRow{
		ID:           "upd-arch-1",
		CompanyID:    "acme",
		Name:         "Old Name",
		AgentType:    "conformist",
		Description:  "Old description",
		MemoryWeight: 0.8,
		IsActive:     true,
	}
	if err := d.CreateArchetype(a); err != nil {
		t.Fatalf("CreateArchetype: %v", err)
	}

	if err := d.UpdateArchetype("upd-arch-1", "New Name", "New description", 1.5, false); err != nil {
		t.Fatalf("UpdateArchetype: %v", err)
	}

	got, err := d.GetArchetype("upd-arch-1")
	if err != nil {
		t.Fatalf("GetArchetype after update: %v", err)
	}
	if got.Name != "New Name" {
		t.Errorf("Name after update: want %q, got %q", "New Name", got.Name)
	}
	if got.MemoryWeight != 1.5 {
		t.Errorf("MemoryWeight after update: want 1.5, got %v", got.MemoryWeight)
	}
	if got.IsActive {
		t.Error("IsActive should be false after update")
	}
}

func TestListArchetypes(t *testing.T) {
	d := openTestDB(t)

	for i, a := range []ArchetypeRow{
		{ID: "la-1", CompanyID: "acme", Name: "A1", AgentType: "conformist", Description: "d1", MemoryWeight: 0.7, IsActive: true},
		{ID: "la-2", CompanyID: "acme", Name: "A2", AgentType: "disruptor", Description: "d2", MemoryWeight: 0.9, IsActive: true},
		{ID: "la-3", CompanyID: "other", Name: "A3", AgentType: "conformist", Description: "d3", MemoryWeight: 0.6, IsActive: true},
	} {
		if err := d.CreateArchetype(&a); err != nil {
			t.Fatalf("CreateArchetype[%d]: %v", i, err)
		}
	}

	// List for acme: should return 2
	rows, err := d.ListArchetypes("acme")
	if err != nil {
		t.Fatalf("ListArchetypes(acme): %v", err)
	}
	if len(rows) != 2 {
		t.Errorf("ListArchetypes(acme): want 2, got %d", len(rows))
	}

	// List all (empty string): should return 3
	all, err := d.ListArchetypes("")
	if err != nil {
		t.Fatalf("ListArchetypes(all): %v", err)
	}
	if len(all) != 3 {
		t.Errorf("ListArchetypes(all): want 3, got %d", len(all))
	}
}

func TestDeleteArchetype(t *testing.T) {
	d := openTestDB(t)

	a := &ArchetypeRow{
		ID: "del-arch-1", CompanyID: "acme", Name: "ToDelete",
		AgentType: "conformist", Description: "d", MemoryWeight: 0.5, IsActive: true,
	}
	if err := d.CreateArchetype(a); err != nil {
		t.Fatalf("CreateArchetype: %v", err)
	}
	if err := d.DeleteArchetype("del-arch-1"); err != nil {
		t.Fatalf("DeleteArchetype: %v", err)
	}
	_, err := d.GetArchetype("del-arch-1")
	if err == nil {
		t.Error("expected error after delete, got nil")
	}
}

func TestBuiltinArchetypeCannotBeDeleted(t *testing.T) {
	d := openTestDB(t)

	// Built-in archetypes have company_id = ''; inserting one to simulate
	builtin := &ArchetypeRow{
		ID: "builtin-1", CompanyID: "", Name: "Built-in",
		AgentType: "conformist", Description: "d", MemoryWeight: 0.7, IsActive: true,
	}
	if err := d.CreateArchetype(builtin); err != nil {
		t.Fatalf("CreateArchetype builtin: %v", err)
	}
	// DeleteArchetype should silently succeed but not delete (WHERE company_id != '')
	if err := d.DeleteArchetype("builtin-1"); err != nil {
		t.Fatalf("DeleteArchetype builtin: %v", err)
	}
	// Should still exist
	got, err := d.GetArchetype("builtin-1")
	if err != nil {
		t.Fatalf("GetArchetype after delete attempt: %v", err)
	}
	if got.ID != "builtin-1" {
		t.Error("built-in archetype should not have been deleted")
	}
}

// ─── Custom Rule CRUD tests ───────────────────────────────────────────────────

func TestCreateAndGetCustomRule(t *testing.T) {
	d := openTestDB(t)

	r := &CustomRuleRow{
		ID:          "rule-1",
		CompanyID:   "acme",
		Description: "No aggressive pricing",
		Domain:      "market",
		Stability:   0.8,
		IsActive:    true,
	}
	if err := d.CreateCustomRule(r); err != nil {
		t.Fatalf("CreateCustomRule: %v", err)
	}

	got, err := d.GetCustomRule("rule-1")
	if err != nil {
		t.Fatalf("GetCustomRule: %v", err)
	}
	if got.Description != r.Description {
		t.Errorf("Description: want %q, got %q", r.Description, got.Description)
	}
	if got.Domain != r.Domain {
		t.Errorf("Domain: want %q, got %q", r.Domain, got.Domain)
	}
	if got.Stability != r.Stability {
		t.Errorf("Stability: want %v, got %v", r.Stability, got.Stability)
	}
	if !got.IsActive {
		t.Error("IsActive should be true")
	}
}

func TestUpdateCustomRule(t *testing.T) {
	d := openTestDB(t)

	r := &CustomRuleRow{
		ID: "rule-upd-1", CompanyID: "acme",
		Description: "Old rule", Domain: "technology", Stability: 0.5, IsActive: true,
	}
	if err := d.CreateCustomRule(r); err != nil {
		t.Fatalf("CreateCustomRule: %v", err)
	}
	if err := d.UpdateCustomRule("rule-upd-1", "Updated rule", "behavior", 0.9, false); err != nil {
		t.Fatalf("UpdateCustomRule: %v", err)
	}
	got, err := d.GetCustomRule("rule-upd-1")
	if err != nil {
		t.Fatalf("GetCustomRule after update: %v", err)
	}
	if got.Description != "Updated rule" {
		t.Errorf("Description: want %q, got %q", "Updated rule", got.Description)
	}
	if got.Domain != "behavior" {
		t.Errorf("Domain: want %q, got %q", "behavior", got.Domain)
	}
	if got.Stability != 0.9 {
		t.Errorf("Stability: want 0.9, got %v", got.Stability)
	}
	if got.IsActive {
		t.Error("IsActive should be false after update")
	}
}

func TestListCustomRules(t *testing.T) {
	d := openTestDB(t)

	for i, r := range []CustomRuleRow{
		{ID: "lr-1", CompanyID: "acme", Description: "R1", Domain: "market", Stability: 0.5, IsActive: true},
		{ID: "lr-2", CompanyID: "acme", Description: "R2", Domain: "technology", Stability: 0.7, IsActive: true},
		{ID: "lr-3", CompanyID: "beta", Description: "R3", Domain: "market", Stability: 0.6, IsActive: true},
	} {
		if err := d.CreateCustomRule(&r); err != nil {
			t.Fatalf("CreateCustomRule[%d]: %v", i, err)
		}
	}

	rules, err := d.ListCustomRules("acme")
	if err != nil {
		t.Fatalf("ListCustomRules: %v", err)
	}
	if len(rules) != 2 {
		t.Errorf("ListCustomRules(acme): want 2, got %d", len(rules))
	}

	betaRules, err := d.ListCustomRules("beta")
	if err != nil {
		t.Fatalf("ListCustomRules(beta): %v", err)
	}
	if len(betaRules) != 1 {
		t.Errorf("ListCustomRules(beta): want 1, got %d", len(betaRules))
	}
}

func TestDeleteCustomRule(t *testing.T) {
	d := openTestDB(t)

	r := &CustomRuleRow{
		ID: "del-rule-1", CompanyID: "acme",
		Description: "ToDelete", Domain: "market", Stability: 0.5, IsActive: true,
	}
	if err := d.CreateCustomRule(r); err != nil {
		t.Fatalf("CreateCustomRule: %v", err)
	}
	if err := d.DeleteCustomRule("del-rule-1"); err != nil {
		t.Fatalf("DeleteCustomRule: %v", err)
	}
	_, err := d.GetCustomRule("del-rule-1")
	if err == nil {
		t.Error("expected error after delete, got nil")
	}
}

// ─── Progress columns tests ───────────────────────────────────────────────────

func TestJobProgressColumns(t *testing.T) {
	d := openTestDB(t)

	// Create a job
	job := &JobRow{
		ID:         "prog-job-1",
		Status:     "running",
		Question:   "Will AI replace analysts?",
		Department: "technology",
		Rounds:     40,
		Company:    "TechCorp",
		CreatedAt:  1000,
	}
	if err := d.UpsertJob(job); err != nil {
		t.Fatalf("UpsertJob: %v", err)
	}

	// Update with progress fields
	job.CurrentRound = 15
	job.CurrentTension = 0.72
	job.FractureCount = 3
	job.LastAgentName = "The Visionary"
	job.LastAgentAction = "Proposed radical market exit"
	job.TotalTokens = 12450
	if err := d.UpsertJob(job); err != nil {
		t.Fatalf("UpsertJob with progress: %v", err)
	}

	// Retrieve and verify
	got, err := d.GetJob("prog-job-1")
	if err != nil {
		t.Fatalf("GetJob: %v", err)
	}
	if got.CurrentRound != 15 {
		t.Errorf("CurrentRound: want 15, got %d", got.CurrentRound)
	}
	if got.CurrentTension != 0.72 {
		t.Errorf("CurrentTension: want 0.72, got %v", got.CurrentTension)
	}
	if got.FractureCount != 3 {
		t.Errorf("FractureCount: want 3, got %d", got.FractureCount)
	}
	if got.LastAgentName != "The Visionary" {
		t.Errorf("LastAgentName: want %q, got %q", "The Visionary", got.LastAgentName)
	}
	if got.LastAgentAction != "Proposed radical market exit" {
		t.Errorf("LastAgentAction: want %q, got %q", "Proposed radical market exit", got.LastAgentAction)
	}
	if got.TotalTokens != 12450 {
		t.Errorf("TotalTokens: want 12450, got %d", got.TotalTokens)
	}
}

func TestJobProgressSurvivesListJobs(t *testing.T) {
	d := openTestDB(t)

	job := &JobRow{
		ID: "prog-list-1", Status: "running",
		Question: "Q", Department: "market", Rounds: 20, Company: "Co",
		CurrentRound: 10, CurrentTension: 0.55, FractureCount: 1,
		LastAgentName: "The Rebel", LastAgentAction: "Disrupted supply chain",
		TotalTokens: 5000, CreatedAt: 2000,
	}
	if err := d.UpsertJob(job); err != nil {
		t.Fatalf("UpsertJob: %v", err)
	}

	jobs, err := d.ListJobs()
	if err != nil {
		t.Fatalf("ListJobs: %v", err)
	}
	if len(jobs) != 1 {
		t.Fatalf("ListJobs: want 1, got %d", len(jobs))
	}
	j := jobs[0]
	if j.CurrentRound != 10 {
		t.Errorf("ListJobs CurrentRound: want 10, got %d", j.CurrentRound)
	}
	if j.FractureCount != 1 {
		t.Errorf("ListJobs FractureCount: want 1, got %d", j.FractureCount)
	}
	if j.TotalTokens != 5000 {
		t.Errorf("ListJobs TotalTokens: want 5000, got %d", j.TotalTokens)
	}
}

// ─── Report generation wiring tests ──────────────────────────────────────────

func TestReportGenerationLifecycle(t *testing.T) {
	d := openTestDB(t)

	// Requires a simulation row first (FK)
	if err := d.SaveSimulation("sim-rg-1", "Q", "market", 20, map[string]string{"status": "done"}); err != nil {
		t.Fatalf("SaveSimulation: %v", err)
	}

	genID := "gen-rg-1"
	if err := d.StartReportGen(genID, "sim-rg-1", "full"); err != nil {
		t.Fatalf("StartReportGen: %v", err)
	}

	// Complete successfully
	if err := d.CompleteReportGen(genID, "done", "", 12300, 8500); err != nil {
		t.Fatalf("CompleteReportGen: %v", err)
	}

	gens, err := d.ListReportGens("sim-rg-1")
	if err != nil {
		t.Fatalf("ListReportGens: %v", err)
	}
	if len(gens) != 1 {
		t.Fatalf("ListReportGens: want 1, got %d", len(gens))
	}
	g := gens[0]
	if g.Status != "done" {
		t.Errorf("Status: want %q, got %q", "done", g.Status)
	}
	if g.DurationMs != 8500 {
		t.Errorf("DurationMs: want 8500, got %d", g.DurationMs)
	}
	if g.TokensUsed != 12300 {
		t.Errorf("TokensUsed: want 12300, got %d", g.TokensUsed)
	}
}

func TestReportGenerationError(t *testing.T) {
	d := openTestDB(t)

	if err := d.SaveSimulation("sim-rg-err", "Q", "market", 20, map[string]string{}); err != nil {
		t.Fatalf("SaveSimulation: %v", err)
	}

	genID := "gen-rg-err"
	if err := d.StartReportGen(genID, "sim-rg-err", "full"); err != nil {
		t.Fatalf("StartReportGen: %v", err)
	}

	// Complete with error
	if err := d.CompleteReportGen(genID, "error", "LLM timeout", 0, 500); err != nil {
		t.Fatalf("CompleteReportGen with error: %v", err)
	}

	gens, err := d.ListReportGens("sim-rg-err")
	if err != nil {
		t.Fatalf("ListReportGens: %v", err)
	}
	if len(gens) != 1 {
		t.Fatalf("ListReportGens: want 1, got %d", len(gens))
	}
	if gens[0].Status != "error" {
		t.Errorf("Status: want %q, got %q", "error", gens[0].Status)
	}
	if gens[0].ErrorMsg != "LLM timeout" {
		t.Errorf("ErrorMsg: want %q, got %q", "LLM timeout", gens[0].ErrorMsg)
	}
}
