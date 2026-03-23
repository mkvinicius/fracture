package db

import (
	"database/sql"
	"embed"
	"fmt"
	"os"
	"testing"
	"time"
)

// openTestDB opens a temporary SQLite database for testing.
func openTestDB(t *testing.T) *DB {
	t.Helper()

	f, err := os.CreateTemp(t.TempDir(), "fracture-test-*.db")
	if err != nil {
		t.Fatalf("create temp db file: %v", err)
	}
	f.Close()

	sqlDB, err := sql.Open("sqlite3", f.Name()+"?_journal_mode=WAL&_foreign_keys=on")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	// Apply schema using the embedded schemaFS from the package
	var schemaBytes []byte
	schemaBytes, err = schemaFS.ReadFile("schema.sql")
	if err != nil {
		t.Fatalf("read schema: %v", err)
	}
	if _, err := sqlDB.Exec(string(schemaBytes)); err != nil {
		t.Fatalf("apply schema: %v", err)
	}

	// Apply versioned migrations
	if err := Migrate(sqlDB); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	d := &DB{sqlDB}
	t.Cleanup(func() { d.Close() })
	return d
}

// Ensure schemaFS is accessible (it's package-level in db.go).
var _ embed.FS = schemaFS

// ─── Config tests ─────────────────────────────────────────────────────────────

func TestSetAndGetConfig(t *testing.T) {
	d := openTestDB(t)

	if err := d.SetConfig("test_key", "hello"); err != nil {
		t.Fatalf("SetConfig: %v", err)
	}
	val, err := d.GetConfig("test_key")
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if val != "hello" {
		t.Errorf("expected 'hello', got %q", val)
	}
}

func TestGetConfigMissing(t *testing.T) {
	d := openTestDB(t)
	val, err := d.GetConfig("nonexistent")
	if err != nil {
		t.Fatalf("GetConfig for missing key should not error: %v", err)
	}
	if val != "" {
		t.Errorf("expected empty string for missing key, got %q", val)
	}
}

func TestSetConfigUpsert(t *testing.T) {
	d := openTestDB(t)
	_ = d.SetConfig("k", "v1")
	_ = d.SetConfig("k", "v2")
	val, _ := d.GetConfig("k")
	if val != "v2" {
		t.Errorf("expected upsert to v2, got %q", val)
	}
}

// ─── Simulation Jobs tests ────────────────────────────────────────────────────

func TestUpsertAndGetJob(t *testing.T) {
	d := openTestDB(t)

	job := &JobRow{
		ID:         "test-id-1",
		Status:     "queued",
		Question:   "What disrupts the fitness market?",
		Department: "market",
		Rounds:     20,
		Company:    "AcmeFit",
		CreatedAt:  time.Now().Unix(),
	}

	if err := d.UpsertJob(job); err != nil {
		t.Fatalf("UpsertJob: %v", err)
	}

	got, err := d.GetJob("test-id-1")
	if err != nil {
		t.Fatalf("GetJob: %v", err)
	}
	if got.Status != "queued" {
		t.Errorf("expected status=queued, got %q", got.Status)
	}
	if got.Question != job.Question {
		t.Errorf("expected question=%q, got %q", job.Question, got.Question)
	}
	if got.Company != "AcmeFit" {
		t.Errorf("expected company=AcmeFit, got %q", got.Company)
	}
}

func TestUpsertJobStatusTransitions(t *testing.T) {
	d := openTestDB(t)

	job := &JobRow{
		ID:        "test-id-2",
		Status:    "queued",
		Question:  "test question",
		CreatedAt: time.Now().Unix(),
	}
	_ = d.UpsertJob(job)

	// Transition to running
	job.Status = "running"
	if err := d.UpsertJob(job); err != nil {
		t.Fatalf("UpsertJob (running): %v", err)
	}

	got, _ := d.GetJob("test-id-2")
	if got.Status != "running" {
		t.Errorf("expected running, got %q", got.Status)
	}

	// Transition to done with duration
	job.Status = "done"
	job.DurationMs = 45000
	_ = d.UpsertJob(job)

	got, _ = d.GetJob("test-id-2")
	if got.Status != "done" {
		t.Errorf("expected done, got %q", got.Status)
	}
	if got.DurationMs != 45000 {
		t.Errorf("expected duration=45000, got %d", got.DurationMs)
	}
}

func TestListJobs(t *testing.T) {
	d := openTestDB(t)

	for i := 0; i < 5; i++ {
		_ = d.UpsertJob(&JobRow{
			ID:        fmt.Sprintf("job-%d", i),
			Status:    "done",
			Question:  fmt.Sprintf("question %d", i),
			CreatedAt: time.Now().Unix() + int64(i),
		})
	}

	jobs, err := d.ListJobs()
	if err != nil {
		t.Fatalf("ListJobs: %v", err)
	}
	if len(jobs) != 5 {
		t.Errorf("expected 5 jobs, got %d", len(jobs))
	}
}

func TestDeleteJob(t *testing.T) {
	d := openTestDB(t)

	_ = d.UpsertJob(&JobRow{ID: "del-1", Status: "done", Question: "q", CreatedAt: time.Now().Unix()})
	_ = d.DeleteJob("del-1")

	_, err := d.GetJob("del-1")
	if err == nil {
		t.Error("expected error after delete, got nil")
	}
}

func TestMarkInterruptedJobsFailed(t *testing.T) {
	d := openTestDB(t)

	statuses := []string{"queued", "researching", "running", "done", "error"}
	for i, s := range statuses {
		_ = d.UpsertJob(&JobRow{
			ID:        fmt.Sprintf("job-%d", i),
			Status:    s,
			Question:  "q",
			CreatedAt: time.Now().Unix(),
		})
	}

	n, err := d.MarkInterruptedJobsFailed()
	if err != nil {
		t.Fatalf("MarkInterruptedJobsFailed: %v", err)
	}
	// queued(0), researching(1), running(2) = 3 interrupted
	if n != 3 {
		t.Errorf("expected 3 interrupted jobs marked, got %d", n)
	}

	// Verify terminal states are unchanged
	done, _ := d.GetJob("job-3")
	if done.Status != "done" {
		t.Errorf("done job should remain done, got %q", done.Status)
	}

	errJob, _ := d.GetJob("job-4")
	if errJob.Status != "error" {
		t.Errorf("error job should remain error, got %q", errJob.Status)
	}

	// Verify interrupted jobs are now error with a message
	for i := 0; i < 3; i++ {
		j, _ := d.GetJob(fmt.Sprintf("job-%d", i))
		if j.Status != "error" {
			t.Errorf("job-%d: expected error after mark, got %q", i, j.Status)
		}
		if j.Error == "" {
			t.Errorf("job-%d: expected non-empty error message", i)
		}
	}
}

func TestGetJobNotFound(t *testing.T) {
	d := openTestDB(t)
	_, err := d.GetJob("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent job, got nil")
	}
}

func TestJobResearchFields(t *testing.T) {
	d := openTestDB(t)

	job := &JobRow{
		ID:              "research-1",
		Status:          "done",
		Question:        "q",
		ResearchSources: 12,
		ResearchTokens:  4500,
		DurationMs:      30000,
		CreatedAt:       time.Now().Unix(),
	}
	_ = d.UpsertJob(job)

	got, _ := d.GetJob("research-1")
	if got.ResearchSources != 12 {
		t.Errorf("expected 12 sources, got %d", got.ResearchSources)
	}
	if got.ResearchTokens != 4500 {
		t.Errorf("expected 4500 tokens, got %d", got.ResearchTokens)
	}
}

// ─── Simulation persistence tests ────────────────────────────────────────────

func TestSaveAndGetSimulation(t *testing.T) {
	d := openTestDB(t)

	result := map[string]interface{}{
		"duration_ms": 12000,
		"summary":     "market will shift",
	}
	if err := d.SaveSimulation("sim-1", "test question", "market", 20, result); err != nil {
		t.Fatalf("SaveSimulation: %v", err)
	}

	row, err := d.GetSimulation("sim-1")
	if err != nil {
		t.Fatalf("GetSimulation: %v", err)
	}
	if row.Question != "test question" {
		t.Errorf("expected question, got %q", row.Question)
	}
	if row.ResultJSON == "" {
		t.Error("expected non-empty ResultJSON")
	}
}

func TestListSimulations(t *testing.T) {
	d := openTestDB(t)

	for i := 0; i < 3; i++ {
		_ = d.SaveSimulation(fmt.Sprintf("sim-%d", i), fmt.Sprintf("q%d", i), "market", 20, map[string]int{"n": i})
	}

	sims, err := d.ListSimulations()
	if err != nil {
		t.Fatalf("ListSimulations: %v", err)
	}
	if len(sims) != 3 {
		t.Errorf("expected 3 simulations, got %d", len(sims))
	}
}

func TestDeleteSimulation(t *testing.T) {
	d := openTestDB(t)

	_ = d.SaveSimulation("del-sim", "q", "market", 20, nil)
	_ = d.DeleteSimulation("del-sim")

	_, err := d.GetSimulation("del-sim")
	if err == nil {
		t.Error("expected error after delete, got nil")
	}
}
