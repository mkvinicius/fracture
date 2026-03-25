package engine

import (
	"fmt"
	"strings"
)

// ReportToMarkdown converts a FullReport to a structured Markdown document.
func ReportToMarkdown(report *FullReport) string {
	var b strings.Builder

	// ── Header ───────────────────────────────────────────────────────────────
	fmt.Fprintf(&b, "# FRACTURE Report — %s\n\n", report.Question)
	fmt.Fprintf(&b, "> Generated: %s | Tokens: %d | Duration: %dms\n\n",
		report.Watermark.GeneratedAt, report.TotalTokens, report.DurationMs)

	// ── Probable Future ───────────────────────────────────────────────────────
	b.WriteString("## Probable Future\n\n")
	b.WriteString(report.ProbableFuture.Narrative)
	b.WriteString("\n\n")
	fmt.Fprintf(&b, "**Confidence:** %.0f%%\n\n", report.ProbableFuture.Confidence*100)

	if len(report.ProbableFuture.Timeline) > 0 {
		b.WriteString("### Timeline\n\n")
		b.WriteString("| Horizon | Description | Confidence |\n")
		b.WriteString("|---------|-------------|------------|\n")
		for _, t := range report.ProbableFuture.Timeline {
			fmt.Fprintf(&b, "| %s | %s | %.0f%% |\n",
				t.Horizon, mdCell(t.Description), t.Confidence*100)
		}
		b.WriteString("\n")
	}

	if len(report.ProbableFuture.KeyAssumptions) > 0 {
		b.WriteString("### Key Assumptions\n\n")
		for _, a := range report.ProbableFuture.KeyAssumptions {
			fmt.Fprintf(&b, "- %s\n", a)
		}
		b.WriteString("\n")
	}

	// ── Tension Map ───────────────────────────────────────────────────────────
	if len(report.TensionMap) > 0 {
		b.WriteString("## Tension Map\n\n")
		b.WriteString("| Domain | Rule | Tension | Status |\n")
		b.WriteString("|--------|------|---------|--------|\n")
		for _, t := range report.TensionMap {
			fmt.Fprintf(&b, "| %s | %s | %.0f%% | %s |\n",
				t.Domain, mdCell(t.Description), t.Tension*100, t.Color)
		}
		b.WriteString("\n")
	}

	// ── Rupture Scenarios ─────────────────────────────────────────────────────
	if len(report.RuptureScenarios) > 0 {
		b.WriteString("## Rupture Scenarios\n\n")
		for _, s := range report.RuptureScenarios {
			fmt.Fprintf(&b, "### %s\n\n", s.RuleDescription)
			fmt.Fprintf(&b, "**Probability:** %.0f%%\n\n", s.Probability*100)
			fmt.Fprintf(&b, "**Who breaks it:** %s\n\n", s.WhoBreaks)
			fmt.Fprintf(&b, "**How:** %s\n\n", s.HowItHappens)
			fmt.Fprintf(&b, "**Impact:** %s\n\n", s.ImpactOnCompany)
			if s.HowToBeFirst != "" {
				fmt.Fprintf(&b, "**How to be first:** %s\n\n", s.HowToBeFirst)
			}
		}
	}

	// ── Coalitions ────────────────────────────────────────────────────────────
	if len(report.Coalitions) > 0 {
		b.WriteString("## Coalitions\n\n")
		for _, c := range report.Coalitions {
			tag := ""
			if c.IsDisruptive {
				tag = " *(disruptive)*"
			}
			fmt.Fprintf(&b, "- **%s**%s — %s (strength: %.0f%%)\n",
				c.Name, tag, c.SharedGoal, c.Strength*100)
		}
		b.WriteString("\n")
	}

	// ── Rupture Timeline ──────────────────────────────────────────────────────
	if len(report.RuptureTimeline) > 0 {
		b.WriteString("## Rupture Timeline\n\n")
		b.WriteString("| Horizon | Event | Trigger | Probability |\n")
		b.WriteString("|---------|-------|---------|-------------|\n")
		for _, e := range report.RuptureTimeline {
			fmt.Fprintf(&b, "| %s | %s | %s | %.0f%% |\n",
				e.Horizon, mdCell(e.Description), mdCell(e.Trigger), e.Probability*100)
		}
		b.WriteString("\n")
	}

	// ── Action Playbook ───────────────────────────────────────────────────────
	if report.ActionPlaybook != nil {
		b.WriteString("## Action Playbook\n\n")
		mdPlaybook(&b, "90 Days", report.ActionPlaybook.Horizon90Days)
		mdPlaybook(&b, "1 Year", report.ActionPlaybook.Horizon1Year)
		mdPlaybook(&b, "3 Years", report.ActionPlaybook.Horizon3Years)
		mdPlaybook(&b, "Quick Wins", report.ActionPlaybook.QuickWins)
		mdPlaybook(&b, "Critical Risks", report.ActionPlaybook.CriticalRisks)
	}

	// ── Fracture Events ───────────────────────────────────────────────────────
	if len(report.FractureEvents) > 0 {
		b.WriteString("## Fracture Events\n\n")
		b.WriteString("| Round | Proposed By | Rule | Accepted | Confidence |\n")
		b.WriteString("|-------|-------------|------|----------|------------|\n")
		for _, fe := range report.FractureEvents {
			accepted := "No"
			if fe.Accepted {
				accepted = "Yes"
			}
			proposer := fe.ProposedBy
			if len(proposer) > 16 {
				proposer = proposer[:16]
			}
			fmt.Fprintf(&b, "| %d | %s | %s | %s | %.0f%% |\n",
				fe.Round, proposer, mdCell(fe.Proposal.NewDescription), accepted, fe.Confidence*100)
		}
		b.WriteString("\n")
	}

	// ── Ensemble ──────────────────────────────────────────────────────────────
	if report.EnsembleResult != nil {
		b.WriteString("## Ensemble\n\n")
		fmt.Fprintf(&b, "**Runs:** %d\n\n", report.EnsembleResult.RunCount)
		if len(report.EnsembleResult.Consensus) > 0 {
			b.WriteString("### Consensus\n\n")
			for _, fe := range report.EnsembleResult.Consensus {
				fmt.Fprintf(&b, "- Round %d: %s\n", fe.Round, mdCell(fe.Proposal.NewDescription))
			}
			b.WriteString("\n")
		}
		if len(report.EnsembleResult.WeakSignals) > 0 {
			b.WriteString("### Weak Signals\n\n")
			for _, fe := range report.EnsembleResult.WeakSignals {
				fmt.Fprintf(&b, "- Round %d: %s\n", fe.Round, mdCell(fe.Proposal.NewDescription))
			}
			b.WriteString("\n")
		}
		if len(report.EnsembleResult.Minority) > 0 {
			b.WriteString("### Minority Scenarios\n\n")
			for _, fe := range report.EnsembleResult.Minority {
				fmt.Fprintf(&b, "- Round %d: %s\n", fe.Round, mdCell(fe.Proposal.NewDescription))
			}
			b.WriteString("\n")
		}
	}

	// ── Watermark ─────────────────────────────────────────────────────────────
	fmt.Fprintf(&b, "---\n\n*%s*\n", report.Watermark.Notice)

	return b.String()
}

// mdCell escapes pipe characters and newlines for use in Markdown table cells.
func mdCell(s string) string {
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}

func mdPlaybook(b *strings.Builder, title string, items []string) {
	if len(items) == 0 {
		return
	}
	fmt.Fprintf(b, "### %s\n\n", title)
	for _, item := range items {
		fmt.Fprintf(b, "- %s\n", item)
	}
	b.WriteString("\n")
}
