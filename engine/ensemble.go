package engine

import (
	"context"
	"fmt"
)

// EnsembleConfig controls how many independent simulation runs are merged.
type EnsembleConfig struct {
	Runs int // number of independent runs to merge (Standard=1, Premium=2)
}

// RunResult captures the key outputs of a single simulation run for ensemble merging.
type RunResult struct {
	FractureEvents []FractureEvent
	FinalWorld     WorldSnapshot
	TensionMap     map[string]float64
	TotalTokens    int
}

// EnsembleResult classifies outcomes across multiple runs into three tiers.
type EnsembleResult struct {
	// Consensus: events that appeared in ≥60% of runs
	Consensus []FractureEvent `json:"consensus"`
	// WeakSignals: events that appeared in >1 run but <60%
	WeakSignals []FractureEvent `json:"weak_signals"`
	// Minority: events that appeared in exactly 1 run
	Minority []FractureEvent `json:"minority"`
	// MergedTensionMap: average tension across all runs
	MergedTensionMap map[string]float64 `json:"merged_tension_map"`
	// RunCount is how many independent runs were merged
	RunCount int `json:"run_count"`
}

// RunEnsemble executes cfg.Runs independent simulations and merges their results.
// The runFn callback produces one RunResult per invocation; it is called sequentially.
// If runs==1 the single result is classified trivially (all events are Minority).
func RunEnsemble(ctx context.Context, cfg EnsembleConfig, runFn func(ctx context.Context, runIdx int) (*RunResult, error)) (*EnsembleResult, error) {
	if cfg.Runs <= 0 {
		cfg.Runs = 1
	}

	runs := make([]*RunResult, 0, cfg.Runs)
	for i := 0; i < cfg.Runs; i++ {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("ensemble cancelled after %d/%d runs: %w", i, cfg.Runs, ctx.Err())
		default:
		}
		res, err := runFn(ctx, i)
		if err != nil {
			return nil, fmt.Errorf("ensemble run %d: %w", i, err)
		}
		runs = append(runs, res)
	}

	return mergeEnsemble(runs), nil
}

// mergeEnsemble classifies fracture events and averages tension maps.
func mergeEnsemble(runs []*RunResult) *EnsembleResult {
	n := len(runs)
	if n == 0 {
		return &EnsembleResult{RunCount: 0}
	}

	// Count how many runs each unique fracture event appeared in.
	// Key: ProposedRuleID + ProposedBy (coarse identity).
	type eventKey struct {
		ruleID     string
		proposedBy string
	}
	counts := make(map[eventKey]int)
	exemplar := make(map[eventKey]FractureEvent)

	for _, r := range runs {
		seen := make(map[eventKey]bool)
		for _, fe := range r.FractureEvents {
			k := eventKey{ruleID: fe.Proposal.OriginalRuleID, proposedBy: fe.ProposedBy}
			if !seen[k] {
				counts[k]++
				exemplar[k] = fe
				seen[k] = true
			}
		}
	}

	threshold60 := int(float64(n)*0.6 + 0.5) // ceil(60% of runs)
	if threshold60 < 1 {
		threshold60 = 1
	}

	result := &EnsembleResult{
		MergedTensionMap: make(map[string]float64),
		RunCount:         n,
	}

	for k, count := range counts {
		fe := exemplar[k]
		switch {
		case count >= threshold60:
			result.Consensus = append(result.Consensus, fe)
		case count > 1:
			result.WeakSignals = append(result.WeakSignals, fe)
		default:
			result.Minority = append(result.Minority, fe)
		}
	}

	// Average tension maps across all runs
	tensionSums := make(map[string]float64)
	tensionCounts := make(map[string]int)
	for _, r := range runs {
		for ruleID, t := range r.TensionMap {
			tensionSums[ruleID] += t
			tensionCounts[ruleID]++
		}
	}
	for ruleID, sum := range tensionSums {
		result.MergedTensionMap[ruleID] = sum / float64(tensionCounts[ruleID])
	}

	return result
}
