package engine

import (
	"testing"
)

func makeReportForComparison(id, question string, scenarios []RuptureScenario, confidence float64, tensions []TensionEntry) *FullReport {
	return &FullReport{
		SimulationID:     id,
		Question:         question,
		ProbableFuture:   ProbableFuture{Confidence: confidence},
		RuptureScenarios: scenarios,
		TensionMap:       tensions,
	}
}

func TestCompareReportsCommonFractures(t *testing.T) {
	scenario := RuptureScenario{RuleID: "pricing", RuleDescription: "Premium pricing collapses"}
	r1 := makeReportForComparison("sim-1", "Q1", []RuptureScenario{
		scenario,
		{RuleID: "trust", RuleDescription: "Trust breaks down"},
	}, 0.7, nil)
	r2 := makeReportForComparison("sim-2", "Q1", []RuptureScenario{
		scenario,
		{RuleID: "regulation", RuleDescription: "Regulation enforced"},
	}, 0.65, nil)

	result := CompareReports([]*FullReport{r1, r2})

	if len(result.CommonFractures) != 1 {
		t.Fatalf("expected 1 common fracture, got %d: %v", len(result.CommonFractures), result.CommonFractures)
	}
	if result.CommonFractures[0] != "Premium pricing collapses" {
		t.Errorf("wrong common fracture: %q", result.CommonFractures[0])
	}
}

func TestCompareReportsDivergentFractures(t *testing.T) {
	r1 := makeReportForComparison("sim-1", "Q1", []RuptureScenario{
		{RuleID: "a", RuleDescription: "Unique to sim-1"},
	}, 0.7, nil)
	r2 := makeReportForComparison("sim-2", "Q1", []RuptureScenario{
		{RuleID: "b", RuleDescription: "Unique to sim-2"},
	}, 0.7, nil)

	result := CompareReports([]*FullReport{r1, r2})

	if len(result.CommonFractures) != 0 {
		t.Errorf("expected 0 common fractures, got %d", len(result.CommonFractures))
	}
	if len(result.DivergentFractures["sim-1"]) != 1 {
		t.Errorf("sim-1 should have 1 divergent fracture")
	}
	if len(result.DivergentFractures["sim-2"]) != 1 {
		t.Errorf("sim-2 should have 1 divergent fracture")
	}
}

func TestCompareReportsTensionDelta(t *testing.T) {
	tensions1 := []TensionEntry{{RuleID: "r1", Description: "Rule one", Domain: "market", Tension: 0.9, Color: "red"}}
	tensions2 := []TensionEntry{{RuleID: "r1", Description: "Rule one", Domain: "market", Tension: 0.3, Color: "green"}}

	r1 := makeReportForComparison("sim-1", "Q", nil, 0.7, tensions1)
	r2 := makeReportForComparison("sim-2", "Q", nil, 0.7, tensions2)

	result := CompareReports([]*FullReport{r1, r2})

	if len(result.TensionDelta) == 0 {
		t.Fatal("expected at least one TensionDelta")
	}
	delta := result.TensionDelta[0]
	if delta.Delta < 0.55 || delta.Delta > 0.65 {
		t.Errorf("expected delta ~0.6 for tensions [0.9, 0.3], got %.4f", delta.Delta)
	}
}

func TestCompareReportsConfidenceDelta(t *testing.T) {
	r1 := makeReportForComparison("sim-1", "Q", nil, 0.9, nil)
	r2 := makeReportForComparison("sim-2", "Q", nil, 0.5, nil)
	r3 := makeReportForComparison("sim-3", "Q", nil, 0.7, nil)

	result := CompareReports([]*FullReport{r1, r2, r3})

	if result.ConfidenceDelta < 0.39 || result.ConfidenceDelta > 0.41 {
		t.Errorf("expected ConfidenceDelta ~0.40, got %.4f", result.ConfidenceDelta)
	}
}

func TestCompareReportsSummaryContent(t *testing.T) {
	r1 := makeReportForComparison("sim-1", "Q", nil, 0.7, nil)
	r2 := makeReportForComparison("sim-2", "Q", nil, 0.5, nil)
	result := CompareReports([]*FullReport{r1, r2})

	if result.Summary == "" {
		t.Error("summary should not be empty")
	}
	if len(result.SimulationIDs) != 2 {
		t.Errorf("expected 2 simulation IDs, got %d", len(result.SimulationIDs))
	}
}

func TestCompareReportsEmptyInput(t *testing.T) {
	result := CompareReports(nil)
	if result == nil {
		t.Fatal("CompareReports(nil) should not return nil")
	}
	if len(result.SimulationIDs) != 0 {
		t.Errorf("empty input should yield empty SimulationIDs")
	}
}

func TestCompareReportsSingleReport(t *testing.T) {
	r := makeReportForComparison("sim-1", "Q", []RuptureScenario{
		{RuleID: "r1", RuleDescription: "Only one run"},
	}, 0.7, nil)
	result := CompareReports([]*FullReport{r})

	// With one report, ConfidenceDelta must be 0
	if result.ConfidenceDelta != 0 {
		t.Errorf("single report ConfidenceDelta should be 0, got %.4f", result.ConfidenceDelta)
	}
}
