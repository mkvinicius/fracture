package engine

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/fracture/fracture/skills"
)

// CouncilResult is the output of a single council debate.
type CouncilResult struct {
	Domain   RuleDomain `json:"domain"`
	Evidence string     `json:"evidence"`
	Round    int        `json:"round"`
}

// Council orchestrates a single-domain expert debate using one LLM call (+ 1 retry).
type Council struct {
	domain RuleDomain
	llm    LLMCaller
}

// BuildCouncils creates one Council per RuleDomain.
func BuildCouncils(llm LLMCaller) []Council {
	domains := []RuleDomain{
		DomainMarket, DomainTechnology, DomainRegulation,
		DomainBehavior, DomainCulture, DomainGeopolitics, DomainFinance,
	}
	councils := make([]Council, 0, len(domains))
	for _, d := range domains {
		councils = append(councils, Council{domain: d, llm: llm})
	}
	return councils
}

// RunCouncilDebate calls the LLM once to produce domain evidence.
// On failure it retries once; if both attempts fail the error is returned.
func (c *Council) RunCouncilDebate(ctx context.Context, world *World, round int) (CouncilResult, error) {
	prompt := c.buildPrompt(world, round)

	systemPrompt := c.systemPrompt()
	if world.SkillID != "" {
		allNames := skills.AllMindNames(world.SkillID)
		relCtx := skills.FormatRelationsContext(world.SkillID, allNames)
		if relCtx != "" {
			systemPrompt += "\n\n" + relCtx
		}
	}

	raw, _, err := c.llm.Call(ctx, systemPrompt, prompt, 400)
	if err != nil {
		// Single retry
		raw, _, err = c.llm.Call(ctx, systemPrompt, prompt, 400)
		if err != nil {
			return CouncilResult{}, fmt.Errorf("council %s round %d: %w", c.domain, round, err)
		}
	}

	evidence := strings.TrimSpace(raw)
	return CouncilResult{
		Domain:   c.domain,
		Evidence: evidence,
		Round:    round,
	}, nil
}

func (c *Council) systemPrompt() string {
	return fmt.Sprintf(`You are a senior expert in the %s domain serving on a strategic council.
Your task: synthesize the current simulation state into a concise expert judgment (2-3 sentences).
Focus on the most critical tension or opportunity visible in the %s domain right now.
Be specific and actionable. Respond with plain text only — no JSON, no markdown.`, c.domain, c.domain)
}

func (c *Council) buildPrompt(world *World, round int) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Round %d — %s domain council.\n\n", round, c.domain))
	sb.WriteString("Active rules in this domain:\n")

	for _, r := range world.ActiveRules() {
		if r.Domain == c.domain {
			sb.WriteString(fmt.Sprintf("- %s (stability: %.2f)\n", r.Description, r.Stability))
		}
	}

	sb.WriteString("\nTop tensions:\n")
	count := 0
	for ruleID, t := range world.TensionMap {
		if count >= 3 {
			break
		}
		if r, ok := world.Rules[ruleID]; ok && r.Domain == c.domain {
			sb.WriteString(fmt.Sprintf("- %s: %.2f\n", r.Description, t))
			count++
		}
	}

	sb.WriteString("\nWhat is the most important signal the council must act on?")
	return sb.String()
}

// RunAllCouncils runs all councils concurrently (max 3 at a time) and merges evidence into the world.
func RunAllCouncils(ctx context.Context, councils []Council, world *World, round int) []CouncilResult {
	sem := make(chan struct{}, 3)
	var (
		mu      sync.Mutex
		wg      sync.WaitGroup
		results []CouncilResult
	)

	for i := range councils {
		wg.Add(1)
		go func(c *Council) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			res, err := c.RunCouncilDebate(ctx, world, round)
			if err != nil {
				return
			}
			mu.Lock()
			results = append(results, res)
			mu.Unlock()
		}(&councils[i])
	}
	wg.Wait()

	// Merge all council outputs into World.Evidence
	if len(results) > 0 {
		var ev strings.Builder
		ev.WriteString(fmt.Sprintf("[Council Round %d]\n", round))
		for _, r := range results {
			ev.WriteString(fmt.Sprintf("%s: %s\n", r.Domain, r.Evidence))
		}
		world.mu.Lock()
		world.Evidence = ev.String()
		world.mu.Unlock()
	}

	return results
}
