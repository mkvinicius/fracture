package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// ReportGenerator produces the three FRACTURE result types from a completed simulation.
type ReportGenerator struct {
	llm LLMCaller // synthesis model (Claude Sonnet recommended)
}

// NewReportGenerator creates a ReportGenerator with the given synthesis LLM.
func NewReportGenerator(llm LLMCaller) *ReportGenerator {
	return &ReportGenerator{llm: llm}
}

// GenerateReport produces all three result types from a SimulationResult.
func (rg *ReportGenerator) GenerateReport(ctx context.Context, result *SimulationResult, question string) (*FullReport, error) {
	// Prepare context summary for the LLM
	summary := rg.buildSummary(result)

	probableFuture, err := rg.generateProbableFuture(ctx, question, summary)
	if err != nil {
		return nil, fmt.Errorf("probable future: %w", err)
	}

	tensionMap := rg.buildTensionMap(result)

	ruptureScenarios, err := rg.generateRuptureScenarios(ctx, question, summary, tensionMap)
	if err != nil {
		return nil, fmt.Errorf("rupture scenarios: %w", err)
	}

	return &FullReport{
		SimulationID:     result.SimulationID,
		Question:         question,
		ProbableFuture:   probableFuture,
		TensionMap:       tensionMap,
		RuptureScenarios: ruptureScenarios,
		FractureEvents:   result.FractureEvents,
		TotalTokens:      result.TotalTokens,
		DurationMs:       result.DurationMs,
	}, nil
}

// FullReport is the complete output of a FRACTURE simulation.
type FullReport struct {
	SimulationID     string              `json:"simulation_id"`
	Question         string              `json:"question"`
	ProbableFuture   ProbableFuture      `json:"probable_future"`
	TensionMap       []TensionEntry      `json:"tension_map"`
	RuptureScenarios []RuptureScenario   `json:"rupture_scenarios"`
	FractureEvents   []FractureEvent     `json:"fracture_events"`
	TotalTokens      int                 `json:"total_tokens"`
	DurationMs       int64               `json:"duration_ms"`
}

// ProbableFuture describes the most likely evolution if no rules are broken.
type ProbableFuture struct {
	Narrative    string           `json:"narrative"`
	Timeline     []TimelineEntry  `json:"timeline"`
	Confidence   float64          `json:"confidence"`
	KeyAssumptions []string       `json:"key_assumptions"`
}

// TimelineEntry is a projected event at a specific time horizon.
type TimelineEntry struct {
	Horizon     string  `json:"horizon"`    // "6 months", "12 months", "24 months"
	Description string  `json:"description"`
	Confidence  float64 `json:"confidence"`
}

// TensionEntry represents a rule and its current tension level.
type TensionEntry struct {
	RuleID      string  `json:"rule_id"`
	Description string  `json:"description"`
	Domain      string  `json:"domain"`
	Tension     float64 `json:"tension"`
	Color       string  `json:"color"` // "green" | "yellow" | "orange" | "red"
}

// buildSummary creates a concise text summary of the simulation for the synthesis LLM.
func (rg *ReportGenerator) buildSummary(result *SimulationResult) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Simulation ran %d rounds with %d total tokens.\n\n", len(result.Rounds), result.TotalTokens))

	if len(result.FractureEvents) > 0 {
		sb.WriteString(fmt.Sprintf("%d FRACTURE POINT(s) were triggered:\n", len(result.FractureEvents)))
		for _, fe := range result.FractureEvents {
			accepted := "REJECTED"
			if fe.Accepted {
				accepted = "ACCEPTED"
			}
			sb.WriteString(fmt.Sprintf("- Round %d: %s proposed changing rule '%s' → %s\n",
				fe.Round, fe.ProposedBy, fe.Proposal.OriginalRuleID, accepted))
			if fe.Accepted {
				sb.WriteString(fmt.Sprintf("  New rule: %s\n", fe.Proposal.NewDescription))
			}
		}
		sb.WriteString("\n")
	} else {
		sb.WriteString("No FRACTURE POINTs were triggered — the world remained stable.\n\n")
	}

	// Summarize last 5 rounds of notable actions
	sb.WriteString("Key agent observations from final rounds:\n")
	start := len(result.Rounds) - 5
	if start < 0 {
		start = 0
	}
	for _, rr := range result.Rounds[start:] {
		for _, action := range rr.Actions {
			if len(action.Text) > 20 {
				sb.WriteString(fmt.Sprintf("- [%s] %s\n", action.AgentID[:8], action.Text[:min(150, len(action.Text))]))
			}
		}
	}

	return sb.String()
}

// buildTensionMap converts the world's tension data into sorted TensionEntries.
func (rg *ReportGenerator) buildTensionMap(result *SimulationResult) []TensionEntry {
	var entries []TensionEntry
	for ruleID, tension := range result.TensionMap {
		rule, ok := result.FinalWorld.Rules[ruleID]
		if !ok || !rule.IsActive {
			continue
		}
		color := "green"
		switch {
		case tension >= 0.7:
			color = "red"
		case tension >= 0.5:
			color = "orange"
		case tension >= 0.3:
			color = "yellow"
		}
		entries = append(entries, TensionEntry{
			RuleID:      ruleID,
			Description: rule.Description,
			Domain:      string(rule.Domain),
			Tension:     tension,
			Color:       color,
		})
	}
	// Sort by tension descending
	for i := 0; i < len(entries); i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[j].Tension > entries[i].Tension {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}
	return entries
}

// extractJSON removes markdown code fences from LLM responses.
func extractJSON(raw string) string {
	// Remove ```json ... ``` or ``` ... ``` blocks
	raw = strings.TrimSpace(raw)
	if strings.HasPrefix(raw, "```") {
		// Find the first newline after the opening fence
		newline := strings.Index(raw, "\n")
		if newline != -1 {
			raw = raw[newline+1:]
		}
		// Remove closing fence
		if idx := strings.LastIndex(raw, "```"); idx != -1 {
			raw = raw[:idx]
		}
		raw = strings.TrimSpace(raw)
	}
	return raw
}

// generateProbableFuture calls the synthesis LLM to produce the narrative.
func (rg *ReportGenerator) generateProbableFuture(ctx context.Context, question, summary string) (ProbableFuture, error) {
	systemPrompt := `You are a strategic analyst synthesizing a market simulation.
Your task: produce a structured "Probable Future" report based on simulation results.
Be specific, concrete, and grounded in the simulation data.
Respond in JSON format only.`

	userPrompt := fmt.Sprintf(`Question asked: %s

Simulation summary:
%s

Produce a Probable Future report in this exact JSON format:
{
  "narrative": "2-3 paragraph narrative of the most likely future if no rules change",
  "timeline": [
    {"horizon": "6 months", "description": "...", "confidence": 0.8},
    {"horizon": "12 months", "description": "...", "confidence": 0.7},
    {"horizon": "24 months", "description": "...", "confidence": 0.6}
  ],
  "confidence": 0.75,
  "key_assumptions": ["assumption 1", "assumption 2", "assumption 3"]
}`, question, summary)

	raw, _, err := rg.llm.Call(ctx, systemPrompt, userPrompt, 800)
	if err != nil {
		return ProbableFuture{}, err
	}

	clean := extractJSON(raw)
	var result ProbableFuture
	if err := json.Unmarshal([]byte(clean), &result); err != nil {
		result.Narrative = raw
		result.Confidence = 0.5
	}
	return result, nil
}

// generateRuptureScenarios calls the synthesis LLM to produce the top 3 rupture scenarios.
func (rg *ReportGenerator) generateRuptureScenarios(ctx context.Context, question, summary string, tensionMap []TensionEntry) ([]RuptureScenario, error) {
	// Build tension context
	var tensionLines []string
	for i, t := range tensionMap {
		if i >= 5 {
			break
		}
		tensionLines = append(tensionLines, fmt.Sprintf("- %s (tension: %.2f, domain: %s)", t.Description, t.Tension, t.Domain))
	}

	systemPrompt := `You are a strategic disruption analyst.
Your task: identify the 3 most likely rupture scenarios — moments where fundamental market rules could be rewritten.
For each scenario, explain who breaks the rule, how it happens, the impact, and crucially: how the company could be FIRST to break it themselves.
Respond in JSON format only.`

	userPrompt := fmt.Sprintf(`Question asked: %s

Simulation summary:
%s

Most tense rules (highest disruption potential):
%s

Produce exactly 3 rupture scenarios in this JSON format:
[
  {
    "rule_id": "...",
    "rule_description": "The current rule that could be broken",
    "probability": 0.65,
    "who_breaks": "Which archetype/player is most likely to break this rule",
    "how_it_happens": "Specific mechanism — what triggers the change and how it unfolds",
    "impact_on_company": "Direct impact on the company asking the question",
    "how_to_be_first": "Concrete steps the company could take to break this rule themselves before a competitor does"
  }
]`, question, summary, strings.Join(tensionLines, "\n"))

	raw, _, err := rg.llm.Call(ctx, systemPrompt, userPrompt, 1200)
	if err != nil {
		return nil, err
	}

	clean := extractJSON(raw)
	var scenarios []RuptureScenario
	if err := json.Unmarshal([]byte(clean), &scenarios); err != nil {
		return []RuptureScenario{{
			RuleDescription: "Analysis unavailable",
			HowItHappens:    raw,
		}}, nil
	}
	return scenarios, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
