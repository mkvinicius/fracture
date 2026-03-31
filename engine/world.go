package engine

import (
	"encoding/json"
	"math"
	"sync"
)

// RuleDomain classifies the nature of a world rule.
type RuleDomain string

const (
	DomainMarket      RuleDomain = "market"
	DomainTechnology  RuleDomain = "technology"
	DomainRegulation  RuleDomain = "regulation"
	DomainBehavior    RuleDomain = "behavior"
	DomainCulture     RuleDomain = "culture"
	DomainGeopolitics RuleDomain = "geopolitics"
	DomainFinance     RuleDomain = "finance"
)

// Rule represents a fundamental rule of the simulated world.
type Rule struct {
	ID          string     `json:"id"`
	Description string     `json:"description"`
	Domain      RuleDomain `json:"domain"`
	Stability   float64    `json:"stability"`   // 0.0 (fragile) to 1.0 (immutable)
	DependsOn   []string   `json:"depends_on"`  // IDs of rules this one depends on
	IsActive    bool       `json:"is_active"`
}

// RuleProposal is a Disruptor's proposal to mutate a world rule.
type RuleProposal struct {
	OriginalRuleID  string  `json:"original_rule_id"`
	NewDescription  string  `json:"new_description"`
	NewDomain       RuleDomain `json:"new_domain"`
	NewStability    float64 `json:"new_stability"`
	Rationale       string  `json:"rationale"`
	ProposedByAgent string  `json:"proposed_by_agent"`
}

// World holds the graph of rules and tracks accumulated tension.
type World struct {
	mu                  sync.RWMutex
	Rules               map[string]*Rule   `json:"rules"`
	TensionMap          map[string]float64 `json:"tension_map"`    // ruleID -> tension 0.0-1.0
	RoundHistory        []WorldSnapshot    `json:"round_history"`
	Evidence            string             `json:"evidence,omitempty"` // domain context from DeepSearch (read-only, not voted on)
	prevRoundInfluence  string             // top agent signals from the previous round, not serialised
}

// WorldSnapshot is an immutable snapshot of the world at a given round.
type WorldSnapshot struct {
	Round      int              `json:"round"`
	Rules      map[string]*Rule `json:"rules"`
	TensionMap map[string]float64 `json:"tension_map"`
}

// NewWorld creates a World from a slice of rules.
func NewWorld(rules []*Rule) *World {
	w := &World{
		Rules:      make(map[string]*Rule, len(rules)),
		TensionMap: make(map[string]float64, len(rules)),
	}
	for _, r := range rules {
		w.Rules[r.ID] = r
		w.TensionMap[r.ID] = 0.0
	}
	return w
}

// CalculateTension returns the average tension across all active rules.
// Tension rises when agents consistently push against a rule.
func (w *World) CalculateTension() float64 {
	w.mu.RLock()
	defer w.mu.RUnlock()

	if len(w.TensionMap) == 0 {
		return 0
	}
	var sum float64
	for _, t := range w.TensionMap {
		sum += t
	}
	return sum / float64(len(w.TensionMap))
}

// IncreaseTension adds pressure to a specific rule.
// Called when an agent's action implies dissatisfaction with that rule.
// If the target rule is listed in other rules' DependsOn fields, those dependent
// rules receive 30% of the delta as cascade pressure — destabilising a foundational
// rule shakes everything built on top of it.
func (w *World) IncreaseTension(ruleID string, delta float64) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if _, ok := w.TensionMap[ruleID]; ok {
		w.TensionMap[ruleID] = math.Min(1.0, w.TensionMap[ruleID]+delta)

		// Cascade: rules that depend on this one also feel pressure.
		const cascadeFactor = 0.30
		for id, rule := range w.Rules {
			if !rule.IsActive || id == ruleID {
				continue
			}
			for _, dep := range rule.DependsOn {
				if dep == ruleID {
					w.TensionMap[id] = math.Min(1.0, w.TensionMap[id]+delta*cascadeFactor)
					break
				}
			}
		}
	}
}

// SetPrevRoundInfluence stores a summary of the most influential agent signals
// from the last completed round. Agents read this via PrevRoundInfluence() to
// understand the social landscape before reacting.
func (w *World) SetPrevRoundInfluence(summary string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.prevRoundInfluence = summary
}

// PrevRoundInfluence returns the influence summary set by SetPrevRoundInfluence.
func (w *World) PrevRoundInfluence() string {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.prevRoundInfluence
}

// ApplyProposal mutates a rule after a successful FRACTURE POINT vote.
// The new rule starts with low stability (it's fragile — just emerged).
func (w *World) ApplyProposal(proposal RuleProposal) *Rule {
	w.mu.Lock()
	defer w.mu.Unlock()

	if original, ok := w.Rules[proposal.OriginalRuleID]; ok {
		// Keep original as inactive historical record
		original.IsActive = false

		// Create new rule with reduced stability (newly emerged rules are fragile)
		newRule := &Rule{
			ID:          proposal.OriginalRuleID + "_v2",
			Description: proposal.NewDescription,
			Domain:      proposal.NewDomain,
			Stability:   math.Max(0.1, proposal.NewStability*0.4), // starts fragile
			DependsOn:   original.DependsOn,
			IsActive:    true,
		}
		w.Rules[newRule.ID] = newRule
		w.TensionMap[newRule.ID] = 0.0
		// Reset tension on original rule
		w.TensionMap[proposal.OriginalRuleID] = 0.0
		return newRule
	}
	return nil
}

// Snapshot returns an immutable copy of the current world state.
func (w *World) Snapshot(round int) WorldSnapshot {
	w.mu.RLock()
	defer w.mu.RUnlock()

	rules := make(map[string]*Rule, len(w.Rules))
	for k, v := range w.Rules {
		cp := *v
		rules[k] = &cp
	}
	tension := make(map[string]float64, len(w.TensionMap))
	for k, v := range w.TensionMap {
		tension[k] = v
	}
	return WorldSnapshot{Round: round, Rules: rules, TensionMap: tension}
}

// ActiveRules returns only rules that are currently active.
func (w *World) ActiveRules() []*Rule {
	w.mu.RLock()
	defer w.mu.RUnlock()
	var active []*Rule
	for _, r := range w.Rules {
		if r.IsActive {
			active = append(active, r)
		}
	}
	return active
}

// ToJSON serialises the world for storage or prompt injection.
func (w *World) ToJSON() (string, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	b, err := json.Marshal(w)
	return string(b), err
}

// MostTenseRules returns the top N rules by tension level.
func (w *World) MostTenseRules(n int) []*Rule {
	w.mu.RLock()
	defer w.mu.RUnlock()

	type pair struct {
		rule    *Rule
		tension float64
	}
	var pairs []pair
	for id, tension := range w.TensionMap {
		if r, ok := w.Rules[id]; ok && r.IsActive {
			pairs = append(pairs, pair{r, tension})
		}
	}
	// Simple selection sort for small N
	for i := 0; i < len(pairs) && i < n; i++ {
		maxIdx := i
		for j := i + 1; j < len(pairs); j++ {
			if pairs[j].tension > pairs[maxIdx].tension {
				maxIdx = j
			}
		}
		pairs[i], pairs[maxIdx] = pairs[maxIdx], pairs[i]
	}
	result := make([]*Rule, 0, n)
	for i := 0; i < len(pairs) && i < n; i++ {
		result = append(result, pairs[i].rule)
	}
	return result
}
