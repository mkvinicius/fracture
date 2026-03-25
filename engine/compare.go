package engine

import (
	"fmt"
	"math"
	"strings"
)

// ComparisonReport summarises the similarities and differences between 2–5 simulations.
type ComparisonReport struct {
	SimulationIDs      []string            `json:"simulation_ids"`
	Questions          []string            `json:"questions"`
	CommonFractures    []string            `json:"common_fractures"`
	DivergentFractures map[string][]string `json:"divergent_fractures"`
	TensionDelta       []TensionDelta      `json:"tension_delta"`
	ConfidenceDelta    float64             `json:"confidence_delta"`
	Summary            string              `json:"summary"`
}

// TensionDelta captures tension variation for a single rule across simulations.
type TensionDelta struct {
	RuleID      string    `json:"rule_id"`
	Description string    `json:"description"`
	Tensions    []float64 `json:"tensions"` // one entry per simulation
	Delta       float64   `json:"delta"`    // max − min
}

// CompareReports analyses commonalities and divergences across 2–5 FullReports.
func CompareReports(reports []*FullReport) *ComparisonReport {
	if len(reports) == 0 {
		return &ComparisonReport{}
	}

	ids := make([]string, len(reports))
	questions := make([]string, len(reports))
	for i, r := range reports {
		ids[i] = r.SimulationID
		questions[i] = r.Question
	}

	commonFractures := findCommonFractures(reports)

	commonSet := make(map[string]bool, len(commonFractures))
	for _, f := range commonFractures {
		commonSet[strings.ToLower(f)] = true
	}

	divergent := make(map[string][]string)
	for i, r := range reports {
		key := r.SimulationID
		if key == "" {
			key = fmt.Sprintf("sim_%d", i)
		}
		for _, s := range r.RuptureScenarios {
			if !commonSet[strings.ToLower(s.RuleDescription)] {
				divergent[key] = append(divergent[key], s.RuleDescription)
			}
		}
	}

	tensionDeltas := computeTensionDeltas(reports)
	confidenceDelta := computeConfidenceDelta(reports)
	summary := buildComparisonSummary(reports, commonFractures, tensionDeltas, confidenceDelta)

	return &ComparisonReport{
		SimulationIDs:      ids,
		Questions:          questions,
		CommonFractures:    commonFractures,
		DivergentFractures: divergent,
		TensionDelta:       tensionDeltas,
		ConfidenceDelta:    confidenceDelta,
		Summary:            summary,
	}
}

// findCommonFractures returns rupture scenario descriptions appearing in ALL reports.
func findCommonFractures(reports []*FullReport) []string {
	if len(reports) == 0 {
		return nil
	}
	counts := make(map[string]int)
	canonical := make(map[string]string) // normalised → original description
	for _, r := range reports {
		seen := make(map[string]bool)
		for _, s := range r.RuptureScenarios {
			norm := strings.ToLower(strings.TrimSpace(s.RuleDescription))
			if !seen[norm] {
				counts[norm]++
				canonical[norm] = s.RuleDescription
				seen[norm] = true
			}
		}
	}
	var common []string
	for norm, count := range counts {
		if count == len(reports) {
			common = append(common, canonical[norm])
		}
	}
	return common
}

// computeTensionDeltas finds the top-10 rules by tension variance across runs.
func computeTensionDeltas(reports []*FullReport) []TensionDelta {
	type entry struct {
		desc     string
		tensions []float64
	}
	byRule := make(map[string]*entry)

	for i, r := range reports {
		for _, t := range r.TensionMap {
			norm := strings.ToLower(t.RuleID)
			if _, ok := byRule[norm]; !ok {
				byRule[norm] = &entry{
					desc:     t.Description,
					tensions: make([]float64, len(reports)),
				}
			}
			if i < len(byRule[norm].tensions) {
				byRule[norm].tensions[i] = t.Tension
			}
		}
	}

	deltas := make([]TensionDelta, 0, len(byRule))
	for norm, e := range byRule {
		maxT := -math.MaxFloat64
		minT := math.MaxFloat64
		for _, v := range e.tensions {
			if v > maxT {
				maxT = v
			}
			if v < minT {
				minT = v
			}
		}
		delta := 0.0
		if maxT != -math.MaxFloat64 {
			delta = maxT - minT
		}
		deltas = append(deltas, TensionDelta{
			RuleID:      norm,
			Description: e.desc,
			Tensions:    e.tensions,
			Delta:       delta,
		})
	}

	// Insertion sort descending by delta
	for i := 1; i < len(deltas); i++ {
		for j := i; j > 0 && deltas[j].Delta > deltas[j-1].Delta; j-- {
			deltas[j], deltas[j-1] = deltas[j-1], deltas[j]
		}
	}
	if len(deltas) > 10 {
		deltas = deltas[:10]
	}
	return deltas
}

// computeConfidenceDelta returns the spread in ProbableFuture.Confidence across runs.
func computeConfidenceDelta(reports []*FullReport) float64 {
	if len(reports) < 2 {
		return 0
	}
	maxC := -math.MaxFloat64
	minC := math.MaxFloat64
	for _, r := range reports {
		c := r.ProbableFuture.Confidence
		if c > maxC {
			maxC = c
		}
		if c < minC {
			minC = c
		}
	}
	if maxC == -math.MaxFloat64 {
		return 0
	}
	return maxC - minC
}

func buildComparisonSummary(reports []*FullReport, common []string, deltas []TensionDelta, confDelta float64) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "%d simulations compared. ", len(reports))
	if len(common) == 0 {
		sb.WriteString("No common rupture patterns — high divergence between runs. ")
	} else {
		fmt.Fprintf(&sb, "%d rupture pattern(s) appear in all runs (robust signal). ", len(common))
	}
	if len(deltas) > 0 && deltas[0].Delta > 0.1 {
		fmt.Fprintf(&sb, "Highest tension variance on %q (Δ=%.2f). ", deltas[0].Description, deltas[0].Delta)
	}
	if confDelta > 0.15 {
		fmt.Fprintf(&sb, "Confidence spread of %.0f%% suggests model sensitivity to initial conditions.", confDelta*100)
	} else {
		sb.WriteString("Confidence levels are stable across runs.")
	}
	return sb.String()
}
