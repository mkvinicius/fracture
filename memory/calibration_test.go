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

// openCalibrationTestDBWithCausality creates a temp SQLite DB with all tables needed
// by Calibrator, including the causality graph tables.
func openCalibrationTestDBWithCausality(t *testing.T) *sql.DB {
	t.Helper()
	db := openCalibrationTestDB(t)

	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS causality_nodes (
			id          TEXT PRIMARY KEY,
			company_id  TEXT NOT NULL DEFAULT '',
			description TEXT NOT NULL DEFAULT '',
			node_type   TEXT NOT NULL DEFAULT 'decision',
			created_at  INTEGER NOT NULL DEFAULT (unixepoch())
		);
		CREATE TABLE IF NOT EXISTS causality_edges (
			id        TEXT PRIMARY KEY DEFAULT (hex(randomblob(8))),
			from_node TEXT NOT NULL,
			to_node   TEXT NOT NULL,
			strength  REAL NOT NULL DEFAULT 0.5,
			evidence  INTEGER NOT NULL DEFAULT 1,
			company_id TEXT NOT NULL DEFAULT '',
			created_at INTEGER NOT NULL DEFAULT (unixepoch()),
			UNIQUE(from_node, to_node)
		);
	`)
	if err != nil {
		t.Fatalf("create causality tables: %v", err)
	}
	return db
}

// ── GetCalibrationReportByDomain ─────────────────────────────────────────────

func TestGetCalibrationReportByDomain_ReturnsMatchingDomain(t *testing.T) {
	db := openCalibrationTestDB(t)
	c := NewCalibrator(db)

	const simID = "sim-domain-001"
	const agentID = "agent-fin-1"
	seedSimulation(t, db, simID, agentID, "finance")

	// Insert calibration row directly for "agent-fin-1" in domain "finance"
	db.Exec(`INSERT OR IGNORE INTO archetype_calibration (archetype_id, domain, accuracy_weight, sample_count) VALUES (?, 'finance', 1.8, 3)`, agentID)

	cals, err := c.GetCalibrationReportByDomain("finance")
	if err != nil {
		t.Fatalf("GetCalibrationReportByDomain: %v", err)
	}
	if len(cals) == 0 {
		t.Fatal("expected at least one calibration record, got 0")
	}

	found := false
	for _, cal := range cals {
		if cal.ArchetypeID == agentID {
			found = true
			if math.Abs(cal.AccuracyWeight-1.8) > 0.01 {
				t.Errorf("expected AccuracyWeight≈1.8 for %q, got %.4f", agentID, cal.AccuracyWeight)
			}
		}
	}
	if !found {
		t.Errorf("agent %q not found in GetCalibrationReportByDomain(\"finance\") results", agentID)
	}
}

func TestGetCalibrationReportByDomain_OtherDomainNotReturned(t *testing.T) {
	db := openCalibrationTestDB(t)
	c := NewCalibrator(db)

	// Seed a simulation for domain "market"
	const simID = "sim-domain-002"
	const agentID = "agent-market-1"
	seedSimulation(t, db, simID, agentID, "market")
	db.Exec(`INSERT OR IGNORE INTO archetype_calibration (archetype_id, domain, accuracy_weight, sample_count) VALUES (?, 'market', 1.5, 2)`, agentID)

	// Query for "technology" — the market agent should NOT appear
	cals, err := c.GetCalibrationReportByDomain("technology")
	if err != nil {
		t.Fatalf("GetCalibrationReportByDomain: %v", err)
	}
	for _, cal := range cals {
		if cal.ArchetypeID == agentID {
			t.Errorf("agent %q (domain=market) should not appear in GetCalibrationReportByDomain(\"technology\")", agentID)
		}
	}
}

// ── RecordProposalAccuracy ────────────────────────────────────────────────────

func TestRecordProposalAccuracy_UpdatesEMA(t *testing.T) {
	db := openCalibrationTestDB(t)
	c := NewCalibrator(db)

	// Perfect acceptance rate (1.0) → deltaScore = 1.0 → weight should rise above 1.0
	err := c.RecordProposalAccuracy("market", []ProposalAccuracyRecord{
		{AgentID: "agent-prop-1", Proposed: 5, Accepted: 5, Rate: 1.0},
	})
	if err != nil {
		t.Fatalf("RecordProposalAccuracy: %v", err)
	}

	var w float64
	if err := db.QueryRow(`SELECT accuracy_weight FROM archetype_calibration WHERE archetype_id = ? AND domain = ?`,
		"agent-prop-1", "market").Scan(&w); err != nil {
		t.Fatalf("query archetype_calibration: %v", err)
	}
	if w <= 1.0 {
		t.Errorf("expected accuracy_weight > 1.0 after perfect rate, got %.4f", w)
	}
}

func TestRecordProposalAccuracy_ZeroRateDecreases(t *testing.T) {
	db := openCalibrationTestDB(t)
	c := NewCalibrator(db)

	const agentID = "agent-prop-2"

	// Seed a baseline weight above 1.0 with several samples so the EMA is stable
	db.Exec(`INSERT OR IGNORE INTO archetype_calibration (archetype_id, domain, accuracy_weight, sample_count) VALUES (?, 'market', 1.8, 10)`, agentID)

	var before float64
	db.QueryRow(`SELECT accuracy_weight FROM archetype_calibration WHERE archetype_id = ? AND domain = ?`, agentID, "market").Scan(&before)

	// Zero rate → deltaScore = -1.0 → weight should not increase
	err := c.RecordProposalAccuracy("market", []ProposalAccuracyRecord{
		{AgentID: agentID, Proposed: 5, Accepted: 0, Rate: 0.0},
	})
	if err != nil {
		t.Fatalf("RecordProposalAccuracy: %v", err)
	}

	var after float64
	if err := db.QueryRow(`SELECT accuracy_weight FROM archetype_calibration WHERE archetype_id = ? AND domain = ?`,
		agentID, "market").Scan(&after); err != nil {
		t.Fatalf("query archetype_calibration: %v", err)
	}
	if after > before {
		t.Errorf("expected weight to decrease (or stay same) after zero rate, before=%.4f after=%.4f", before, after)
	}
}

func TestRecordProposalAccuracy_EmptySliceNoError(t *testing.T) {
	db := openCalibrationTestDB(t)
	c := NewCalibrator(db)

	if err := c.RecordProposalAccuracy("market", nil); err != nil {
		t.Errorf("RecordProposalAccuracy(nil) returned error: %v", err)
	}
	if err := c.RecordProposalAccuracy("market", []ProposalAccuracyRecord{}); err != nil {
		t.Errorf("RecordProposalAccuracy(empty slice) returned error: %v", err)
	}
}

// ── GetFullCausalGraph ────────────────────────────────────────────────────────

func TestGetFullCausalGraph_EmptyDB(t *testing.T) {
	db := openCalibrationTestDBWithCausality(t)
	graph := NewCausalityGraph(db)

	data, err := graph.GetFullCausalGraph("")
	if err != nil {
		t.Fatalf("GetFullCausalGraph: %v", err)
	}
	if data == nil {
		t.Fatal("expected non-nil CausalGraphData, got nil")
	}
	if len(data.Nodes) != 0 {
		t.Errorf("expected 0 nodes on empty DB, got %d", len(data.Nodes))
	}
	if len(data.Edges) != 0 {
		t.Errorf("expected 0 edges on empty DB, got %d", len(data.Edges))
	}
}

func TestGetFullCausalGraph_WithData(t *testing.T) {
	db := openCalibrationTestDBWithCausality(t)
	graph := NewCausalityGraph(db)

	// Insert 2 nodes and 1 edge for company "test-co"
	const nodeA = "node-a-id"
	const nodeB = "node-b-id"
	db.Exec(`INSERT INTO causality_nodes (id, company_id, description, node_type) VALUES (?, 'test-co', 'decision A', 'decision')`, nodeA)
	db.Exec(`INSERT INTO causality_nodes (id, company_id, description, node_type) VALUES (?, 'test-co', 'outcome B', 'outcome')`, nodeB)
	db.Exec(`INSERT INTO causality_edges (from_node, to_node, strength, evidence, company_id) VALUES (?, ?, 0.7, 3, 'test-co')`, nodeA, nodeB)

	data, err := graph.GetFullCausalGraph("test-co")
	if err != nil {
		t.Fatalf("GetFullCausalGraph: %v", err)
	}
	if data == nil {
		t.Fatal("expected non-nil CausalGraphData, got nil")
	}
	if len(data.Nodes) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(data.Nodes))
	}
	if len(data.Edges) != 1 {
		t.Errorf("expected 1 edge, got %d", len(data.Edges))
	}
	if len(data.Edges) == 1 {
		edge := data.Edges[0]
		if edge.From != nodeA {
			t.Errorf("expected edge.From=%q, got %q", nodeA, edge.From)
		}
		if edge.To != nodeB {
			t.Errorf("expected edge.To=%q, got %q", nodeB, edge.To)
		}
	}
}
