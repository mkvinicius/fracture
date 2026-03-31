package memory

import (
	"database/sql"
	"math"
	"os"
	"testing"

	_ "modernc.org/sqlite"
)

// openCalibrationTestDB creates a temp SQLite DB with the tables needed by Calibrator.
func openCalibrationTestDB(t *testing.T) *sql.DB {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "calib-test-*.db")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	f.Close()

	db, err := sql.Open("sqlite", f.Name()+"?_foreign_keys=off")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS archetype_calibration (
			archetype_id    TEXT NOT NULL,
			domain          TEXT NOT NULL,
			accuracy_weight REAL NOT NULL DEFAULT 1.0,
			sample_count    INTEGER NOT NULL DEFAULT 0,
			updated_at      INTEGER NOT NULL DEFAULT (unixepoch()),
			PRIMARY KEY (archetype_id, domain)
		);
		CREATE TABLE IF NOT EXISTS simulation_rounds (
			id            TEXT PRIMARY KEY,
			simulation_id TEXT NOT NULL,
			round_number  INTEGER NOT NULL,
			agent_id      TEXT NOT NULL,
			agent_type    TEXT NOT NULL DEFAULT 'conformist',
			action_text   TEXT NOT NULL DEFAULT '',
			tension_level REAL NOT NULL DEFAULT 0.0,
			fracture_proposed INTEGER NOT NULL DEFAULT 0,
			tokens_used   INTEGER NOT NULL DEFAULT 0,
			created_at    INTEGER NOT NULL DEFAULT (unixepoch())
		);
		CREATE TABLE IF NOT EXISTS simulation_jobs (
			id         TEXT PRIMARY KEY,
			department TEXT NOT NULL DEFAULT 'market',
			company    TEXT NOT NULL DEFAULT '',
			status     TEXT NOT NULL DEFAULT 'done',
			question   TEXT NOT NULL DEFAULT '',
			rounds     INTEGER NOT NULL DEFAULT 20,
			error_msg  TEXT NOT NULL DEFAULT '',
			created_at INTEGER NOT NULL DEFAULT (unixepoch()),
			updated_at INTEGER NOT NULL DEFAULT (unixepoch())
		);
	`)
	if err != nil {
		t.Fatalf("create tables: %v", err)
	}

	t.Cleanup(func() { db.Close() })
	return db
}

func seedSimulation(t *testing.T, db *sql.DB, simID, agentID, domain string) {
	t.Helper()
	db.Exec(`INSERT OR IGNORE INTO simulation_jobs (id, department) VALUES (?, ?)`, simID, domain)
	db.Exec(`INSERT OR IGNORE INTO simulation_rounds (id, simulation_id, round_number, agent_id) VALUES (?, ?, 1, ?)`,
		simID+"-r1", simID, agentID)
}

func TestRecordFeedbackUpdatesAccuracyWeight(t *testing.T) {
	db := openCalibrationTestDB(t)
	c := NewCalibrator(db)

	const simID = "sim-calib-001"
	const agentID = "agent-alpha"
	seedSimulation(t, db, simID, agentID, "market")

	fb := FeedbackRecord{
		SimulationID: simID,
		DeltaScore:   1.0, // perfect prediction
	}
	if err := c.RecordFeedback("company-x", fb); err != nil {
		t.Fatalf("RecordFeedback: %v", err)
	}

	var w float64
	var n int
	err := db.QueryRow(`SELECT accuracy_weight, sample_count FROM archetype_calibration WHERE archetype_id = ? AND domain = ?`, agentID, "market").Scan(&w, &n)
	if err != nil {
		t.Fatalf("query calibration: %v", err)
	}
	if n != 1 {
		t.Errorf("expected sample_count=1, got %d", n)
	}
	// delta=1.0 → adjustment=1.0 → alpha=1.0 → new_weight=1.0 → calibrated = 0.3+1.0*1.7 = 2.0
	if math.Abs(w-2.0) > 0.01 {
		t.Errorf("expected accuracy_weight≈2.0 for perfect prediction, got %.4f", w)
	}
}

func TestRecordFeedbackNegativeDeltaDecreasesWeight(t *testing.T) {
	db := openCalibrationTestDB(t)
	c := NewCalibrator(db)

	const simID = "sim-calib-002"
	const agentID = "agent-beta"
	seedSimulation(t, db, simID, agentID, "technology")

	fb := FeedbackRecord{
		SimulationID: simID,
		DeltaScore:   -1.0, // completely wrong
	}
	if err := c.RecordFeedback("company-y", fb); err != nil {
		t.Fatalf("RecordFeedback: %v", err)
	}

	var w float64
	err := db.QueryRow(`SELECT accuracy_weight FROM archetype_calibration WHERE archetype_id = ? AND domain = ?`, agentID, "technology").Scan(&w)
	if err != nil {
		t.Fatalf("query calibration: %v", err)
	}
	// delta=-1.0 → adjustment=0.0 → calibrated = 0.3+0.0*1.7 = 0.3 (floor)
	if math.Abs(w-0.3) > 0.01 {
		t.Errorf("expected accuracy_weight≈0.3 for worst prediction, got %.4f", w)
	}
}

func TestRecordFeedbackAccumulatesSampleCount(t *testing.T) {
	db := openCalibrationTestDB(t)
	c := NewCalibrator(db)

	const agentID = "agent-gamma"

	for i := 1; i <= 3; i++ {
		simID := "sim-acc-" + string(rune('0'+i))
		seedSimulation(t, db, simID, agentID, "market")
		fb := FeedbackRecord{SimulationID: simID, DeltaScore: 0.5}
		if err := c.RecordFeedback("co", fb); err != nil {
			t.Fatalf("RecordFeedback %d: %v", i, err)
		}
	}

	var n int
	err := db.QueryRow(`SELECT sample_count FROM archetype_calibration WHERE archetype_id = ? AND domain = ?`, agentID, "market").Scan(&n)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if n != 3 {
		t.Errorf("expected sample_count=3 after 3 feedbacks, got %d", n)
	}
}

func TestGetCalibrationReportFiltersCompany(t *testing.T) {
	db := openCalibrationTestDB(t)
	c := NewCalibrator(db)

	// Seed a simulation for "company-a"
	const simID = "sim-report-001"
	const agentID = "agent-delta"
	db.Exec(`INSERT OR IGNORE INTO simulation_jobs (id, department, company) VALUES (?, 'market', 'company-a')`, simID)
	db.Exec(`INSERT OR IGNORE INTO simulation_rounds (id, simulation_id, round_number, agent_id) VALUES (?, ?, 1, ?)`, simID+"-r1", simID, agentID)

	// Insert calibration for this agent
	db.Exec(`INSERT OR IGNORE INTO archetype_calibration (archetype_id, domain, accuracy_weight, sample_count) VALUES (?, 'market', 1.5, 2)`, agentID)

	cals, err := c.GetCalibrationReport("company-a")
	if err != nil {
		t.Fatalf("GetCalibrationReport: %v", err)
	}
	if len(cals) == 0 {
		t.Fatal("expected at least one calibration, got 0")
	}
	found := false
	for _, cal := range cals {
		if cal.ArchetypeID == agentID {
			found = true
			if math.Abs(cal.AccuracyWeight-1.5) > 0.01 {
				t.Errorf("expected AccuracyWeight≈1.5, got %.4f", cal.AccuracyWeight)
			}
		}
	}
	if !found {
		t.Errorf("agentID %q not found in calibration report", agentID)
	}
}

func TestRecordFeedbackNoSimulationJobFallsBack(t *testing.T) {
	db := openCalibrationTestDB(t)
	c := NewCalibrator(db)

	// No simulation_jobs row — should fall back to domain="market" and not error
	const simID = "sim-no-job"
	db.Exec(`INSERT OR IGNORE INTO simulation_rounds (id, simulation_id, round_number, agent_id) VALUES (?, ?, 1, 'agent-orphan')`,
		simID+"-r1", simID)

	fb := FeedbackRecord{SimulationID: simID, DeltaScore: 0.0}
	if err := c.RecordFeedback("co", fb); err != nil {
		t.Errorf("RecordFeedback should not error on missing job: %v", err)
	}
}
