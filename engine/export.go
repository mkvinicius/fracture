package engine

import (
	"fmt"
	"html"
	"strings"
	"time"
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

// ── HTML Export ───────────────────────────────────────────────────────────────

// ReportToHTML generates a self-contained, print-optimised HTML document from a
// FullReport. No external assets are required. The caller supplies the optional
// vertical-skill ID (e.g. "healthcare") for the header badge.
func ReportToHTML(report *FullReport, skillID string) string {
	var b strings.Builder

	// ── title (truncate long questions) ──────────────────────────────────────
	titleQ := report.Question
	if len(titleQ) > 60 {
		titleQ = titleQ[:60] + "…"
	}

	// ── generated-at formatted for humans ────────────────────────────────────
	genAt := report.Watermark.GeneratedAt
	if t, err := time.Parse(time.RFC3339, genAt); err == nil {
		genAt = t.Format("02 Jan 2006, 15:04")
	}

	b.WriteString(`<!DOCTYPE html>
<html lang="pt-BR">
<head>
<meta charset="UTF-8">
<title>FRACTURE Report — ` + html.EscapeString(titleQ) + `</title>
<style>
*,*::before,*::after{box-sizing:border-box;margin:0;padding:0}
body{font-family:'Georgia','Times New Roman',serif;font-size:11pt;line-height:1.6;color:#1a1a1a;background:#fff;max-width:210mm;margin:0 auto;padding:15mm 20mm}
h1{font-size:20pt;font-weight:700;color:#0f172a;margin-bottom:8pt}
h2{font-size:14pt;font-weight:700;color:#1e293b;margin-top:24pt;margin-bottom:8pt;padding-bottom:4pt;border-bottom:2px solid #e2e8f0}
h3{font-size:11pt;font-weight:700;color:#334155;margin-top:12pt;margin-bottom:4pt}
p{margin-bottom:8pt}
.report-header{border-bottom:3px solid #0f172a;padding-bottom:12pt;margin-bottom:20pt}
.report-meta{font-size:9pt;color:#64748b;margin-top:6pt}
.report-meta span{margin-right:16pt}
.skill-badge{display:inline-block;background:#0f172a;color:#fff;font-size:8pt;font-weight:700;padding:3pt 8pt;border-radius:4pt;margin-bottom:8pt;letter-spacing:.5pt}
.confidence{display:inline-block;font-size:9pt;font-weight:700;padding:2pt 6pt;border-radius:3pt}
.conf-high{background:#dcfce7;color:#166534}
.conf-med{background:#fef9c3;color:#854d0e}
.conf-low{background:#fee2e2;color:#991b1b}
table{width:100%;border-collapse:collapse;margin:8pt 0 16pt;font-size:10pt}
th{background:#f1f5f9;font-weight:700;text-align:left;padding:6pt 8pt;border-bottom:2px solid #cbd5e1}
td{padding:5pt 8pt;border-bottom:1px solid #e2e8f0;vertical-align:top}
tr:last-child td{border-bottom:none}
.tension-bar-container{display:inline-block;width:80pt;height:8pt;background:#e2e8f0;border-radius:4pt;vertical-align:middle;margin-right:4pt}
.tension-bar{height:100%;border-radius:4pt}
.tension-green{background:#22c55e}
.tension-yellow{background:#eab308}
.tension-orange{background:#f97316}
.tension-red{background:#ef4444}
.scenario-card{border:1px solid #e2e8f0;border-radius:6pt;padding:10pt 12pt;margin-bottom:12pt;break-inside:avoid}
.scenario-card h3{margin-top:0}
.scenario-field{margin-top:6pt;font-size:10pt}
.scenario-field strong{color:#475569;font-size:9pt;text-transform:uppercase;letter-spacing:.3pt}
.how-to-be-first{background:#f0fdf4;border-left:3px solid #22c55e;padding:6pt 10pt;margin-top:8pt;font-size:10pt}
.how-to-be-first strong{color:#166534}
.coalition-card{border:1px solid #e2e8f0;border-radius:6pt;padding:8pt 12pt;margin-bottom:8pt;break-inside:avoid}
.coalition-disruptive{border-color:#f97316}
.disruptive-badge{display:inline-block;background:#fff7ed;color:#c2410c;font-size:8pt;font-weight:700;padding:2pt 6pt;border-radius:3pt;margin-left:6pt}
.agent-pills{margin-top:4pt}
.agent-pill{display:inline-block;background:#f1f5f9;color:#475569;font-size:8pt;padding:2pt 6pt;border-radius:3pt;margin:2pt 2pt 0 0}
.playbook-grid{display:grid;grid-template-columns:1fr 1fr 1fr;gap:8pt;margin-bottom:12pt}
.playbook-col{border:1px solid #e2e8f0;border-radius:6pt;padding:8pt}
.playbook-col h4{font-size:9pt;font-weight:700;color:#475569;text-transform:uppercase;letter-spacing:.3pt;margin-bottom:6pt;padding-bottom:4pt;border-bottom:1px solid #e2e8f0}
.playbook-col ul{padding-left:12pt}
.playbook-col li{font-size:9pt;margin-bottom:3pt}
.playbook-bottom{display:grid;grid-template-columns:1fr 1fr;gap:8pt}
.quick-wins{border-color:#22c55e}
.critical-risks{border-color:#ef4444}
.quick-wins h4{color:#166534}
.critical-risks h4{color:#991b1b}
.fracture-event{padding:5pt 0;border-bottom:1px solid #f1f5f9;font-size:10pt}
.event-accepted{color:#166534;font-weight:700;font-size:8pt}
.event-rejected{color:#991b1b;font-size:8pt}
.ensemble-section{background:#0f172a;color:#fff;border-radius:8pt;padding:12pt 16pt;margin-bottom:16pt}
.ensemble-section h2{color:#fff;border-bottom-color:#334155}
.ensemble-section h3{color:#94a3b8}
.ensemble-item{background:#1e293b;padding:6pt 10pt;border-radius:4pt;margin-bottom:6pt;font-size:10pt}
.watermark{margin-top:32pt;padding-top:12pt;border-top:1px solid #e2e8f0;text-align:center;font-size:8pt;color:#94a3b8}
@media print{
  body{padding:0;max-width:100%}
  h2{page-break-after:avoid}
  .scenario-card,.coalition-card,.playbook-grid,.ensemble-section{page-break-inside:avoid}
  @page{size:A4;margin:15mm 20mm}
  @page :first{margin-top:20mm}
}
</style>
</head>
<body>
`)

	// ── Header ───────────────────────────────────────────────────────────────
	b.WriteString(`<div class="report-header">`)
	if emoji, name := htmlSkillBadge(skillID); name != "" {
		fmt.Fprintf(&b, `<div class="skill-badge">%s %s SKILL</div>`+"\n", emoji, html.EscapeString(strings.ToUpper(name)))
	}
	fmt.Fprintf(&b, "<h1>%s</h1>\n", html.EscapeString(report.Question))
	fmt.Fprintf(&b, `<div class="report-meta"><span>&#128336; %s</span><span>&#9889; %s tokens</span><span>&#9201; %.1fs</span><span>&#128202; FRACTURE %s</span></div>`+"\n",
		html.EscapeString(genAt),
		formatInt(report.TotalTokens),
		float64(report.DurationMs)/1000,
		html.EscapeString(report.Watermark.Version),
	)
	b.WriteString("</div>\n")

	// ── Probable Future ───────────────────────────────────────────────────────
	b.WriteString("<h2>Prov&#225;vel Futuro</h2>\n")
	fmt.Fprintf(&b, "<p>%s</p>\n", html.EscapeString(report.ProbableFuture.Narrative))
	fmt.Fprintf(&b, `<p>Confian&#231;a geral: <span class="confidence conf-%s">%s</span></p>`+"\n",
		htmlConfLevel(report.ProbableFuture.Confidence),
		htmlPct(report.ProbableFuture.Confidence),
	)

	if len(report.ProbableFuture.Timeline) > 0 {
		b.WriteString("<h3>Linha do Tempo</h3>\n")
		b.WriteString("<table><tr><th>Horizonte</th><th>Descri&#231;&#227;o</th><th>Confian&#231;a</th></tr>\n")
		for _, t := range report.ProbableFuture.Timeline {
			fmt.Fprintf(&b, "<tr><td><strong>%s</strong></td><td>%s</td><td><span class=\"confidence conf-%s\">%s</span></td></tr>\n",
				html.EscapeString(t.Horizon), html.EscapeString(t.Description),
				htmlConfLevel(t.Confidence), htmlPct(t.Confidence))
		}
		b.WriteString("</table>\n")
	}

	if len(report.ProbableFuture.KeyAssumptions) > 0 {
		b.WriteString("<h3>Premissas Fundamentais</h3>\n<ul>\n")
		for _, a := range report.ProbableFuture.KeyAssumptions {
			fmt.Fprintf(&b, "<li>%s</li>\n", html.EscapeString(a))
		}
		b.WriteString("</ul>\n")
	}

	// ── Tension Map ───────────────────────────────────────────────────────────
	if len(report.TensionMap) > 0 {
		b.WriteString("<h2>Mapa de Tens&#227;o</h2>\n")
		b.WriteString("<table><tr><th>Dom&#237;nio</th><th>Regra</th><th>Tens&#227;o</th><th>Status</th></tr>\n")
		for _, t := range report.TensionMap {
			barW := fmt.Sprintf("%.0f%%", t.Tension*100)
			fmt.Fprintf(&b,
				`<tr><td>%s</td><td>%s</td><td><div class="tension-bar-container"><div class="tension-bar tension-%s" style="width:%s"></div></div> %s</td><td>%s</td></tr>`+"\n",
				html.EscapeString(t.Domain), html.EscapeString(t.Description),
				html.EscapeString(t.Color), barW, htmlPct(t.Tension),
				html.EscapeString(htmlTensionLabel(t.Color)),
			)
		}
		b.WriteString("</table>\n")
	}

	// ── Rupture Scenarios ─────────────────────────────────────────────────────
	if len(report.RuptureScenarios) > 0 {
		b.WriteString("<h2>Cen&#225;rios de Ruptura</h2>\n")
		for _, s := range report.RuptureScenarios {
			b.WriteString(`<div class="scenario-card">`)
			fmt.Fprintf(&b, "<h3>%s</h3>\n", html.EscapeString(s.RuleDescription))
			fmt.Fprintf(&b, `<p>Probabilidade: <span class="confidence conf-%s">%s</span></p>`+"\n",
				htmlConfLevel(s.Probability), htmlPct(s.Probability))
			if s.WhoBreaks != "" {
				fmt.Fprintf(&b, `<div class="scenario-field"><strong>Quem rompe</strong><p>%s</p></div>`+"\n", html.EscapeString(s.WhoBreaks))
			}
			if s.HowItHappens != "" {
				fmt.Fprintf(&b, `<div class="scenario-field"><strong>Como acontece</strong><p>%s</p></div>`+"\n", html.EscapeString(s.HowItHappens))
			}
			if s.ImpactOnCompany != "" {
				fmt.Fprintf(&b, `<div class="scenario-field"><strong>Impacto</strong><p>%s</p></div>`+"\n", html.EscapeString(s.ImpactOnCompany))
			}
			if s.HowToBeFirst != "" {
				fmt.Fprintf(&b, `<div class="how-to-be-first"><strong>Como ser o primeiro a romper</strong><p>%s</p></div>`+"\n", html.EscapeString(s.HowToBeFirst))
			}
			b.WriteString("</div>\n")
		}
	}

	// ── Coalitions ────────────────────────────────────────────────────────────
	if len(report.Coalitions) > 0 {
		b.WriteString("<h2>Coaliz&#245;es</h2>\n")
		for _, c := range report.Coalitions {
			cls := "coalition-card"
			if c.IsDisruptive {
				cls += " coalition-disruptive"
			}
			fmt.Fprintf(&b, `<div class="%s">`, cls)
			if c.IsDisruptive {
				fmt.Fprintf(&b, "<h3>%s<span class=\"disruptive-badge\">DISRUPTIVA</span></h3>\n", html.EscapeString(c.Name))
			} else {
				fmt.Fprintf(&b, "<h3>%s</h3>\n", html.EscapeString(c.Name))
			}
			fmt.Fprintf(&b, `<p style="font-size:10pt;color:#475569;">%s</p>`+"\n", html.EscapeString(c.SharedGoal))
			if len(c.AgentNames) > 0 {
				b.WriteString(`<div class="agent-pills">`)
				for _, n := range c.AgentNames {
					fmt.Fprintf(&b, `<span class="agent-pill">%s</span>`, html.EscapeString(n))
				}
				b.WriteString("</div>\n")
			}
			b.WriteString("</div>\n")
		}
	}

	// ── Rupture Timeline ──────────────────────────────────────────────────────
	if len(report.RuptureTimeline) > 0 {
		b.WriteString("<h2>Timeline de Rupturas</h2>\n")
		b.WriteString("<table><tr><th>Horizonte</th><th>Evento</th><th>Gatilho</th><th>Prob.</th></tr>\n")
		for _, e := range report.RuptureTimeline {
			fmt.Fprintf(&b, "<tr><td><strong>%s</strong></td><td>%s</td><td>%s</td><td><span class=\"confidence conf-%s\">%s</span></td></tr>\n",
				html.EscapeString(e.Horizon), html.EscapeString(e.Description),
				html.EscapeString(e.Trigger),
				htmlConfLevel(e.Probability), htmlPct(e.Probability))
		}
		b.WriteString("</table>\n")
	}

	// ── Action Playbook ───────────────────────────────────────────────────────
	if report.ActionPlaybook != nil {
		b.WriteString("<h2>Playbook de A&#231;&#227;o</h2>\n")
		b.WriteString(`<div class="playbook-grid">`)
		htmlPlaybookCol(&b, "90 Dias", "", report.ActionPlaybook.Horizon90Days)
		htmlPlaybookCol(&b, "1 Ano", "", report.ActionPlaybook.Horizon1Year)
		htmlPlaybookCol(&b, "3 Anos", "", report.ActionPlaybook.Horizon3Years)
		b.WriteString("</div>\n")
		b.WriteString(`<div class="playbook-bottom">`)
		htmlPlaybookCol(&b, "Quick Wins", "quick-wins", report.ActionPlaybook.QuickWins)
		htmlPlaybookCol(&b, "Riscos Cr&#237;ticos", "critical-risks", report.ActionPlaybook.CriticalRisks)
		b.WriteString("</div>\n")
	}

	// ── Fracture Events ───────────────────────────────────────────────────────
	if len(report.FractureEvents) > 0 {
		b.WriteString("<h2>Eventos de Fratura</h2>\n")
		for _, fe := range report.FractureEvents {
			proposer := fe.ProposedBy
			if len(proposer) > 16 {
				proposer = proposer[:16]
			}
			b.WriteString(`<div class="fracture-event">`)
			fmt.Fprintf(&b, "<strong>Round %d</strong> &#183; %s &#8594; %s ",
				fe.Round, html.EscapeString(proposer), html.EscapeString(fe.Proposal.NewDescription))
			if fe.Accepted {
				conf := ""
				if fe.Confidence > 0 {
					conf = fmt.Sprintf(" &#183; Confian&#231;a: %s", htmlPct(fe.Confidence))
				}
				fmt.Fprintf(&b, `<span class="event-accepted">&#10003; ACEITO%s</span>`, conf)
			} else {
				b.WriteString(`<span class="event-rejected">&#10007; REJEITADO</span>`)
			}
			b.WriteString("</div>\n")
		}
	}

	// ── Ensemble (Premium) ────────────────────────────────────────────────────
	if report.EnsembleResult != nil {
		b.WriteString(`<div class="ensemble-section">`)
		b.WriteString("<h2>Resultado Ensemble &#8212; Premium</h2>\n")
		if len(report.EnsembleResult.Consensus) > 0 {
			b.WriteString("<h3>Consenso entre simula&#231;&#245;es</h3>\n")
			for _, fe := range report.EnsembleResult.Consensus {
				fmt.Fprintf(&b, `<div class="ensemble-item">Round %d: %s</div>`+"\n",
					fe.Round, html.EscapeString(fe.Proposal.NewDescription))
			}
		}
		if len(report.EnsembleResult.WeakSignals) > 0 {
			b.WriteString("<h3>Sinais Fracos</h3>\n")
			for _, fe := range report.EnsembleResult.WeakSignals {
				fmt.Fprintf(&b, `<div class="ensemble-item">Round %d: %s</div>`+"\n",
					fe.Round, html.EscapeString(fe.Proposal.NewDescription))
			}
		}
		if len(report.EnsembleResult.Minority) > 0 {
			b.WriteString("<h3>Cen&#225;rios Majorit&#225;rios</h3>\n")
			for _, fe := range report.EnsembleResult.Minority {
				fmt.Fprintf(&b, `<div class="ensemble-item">Round %d: %s</div>`+"\n",
					fe.Round, html.EscapeString(fe.Proposal.NewDescription))
			}
		}
		b.WriteString("</div>\n")
	}

	// ── Watermark ─────────────────────────────────────────────────────────────
	b.WriteString(`<div class="watermark">`)
	fmt.Fprintf(&b, "<p><strong>%s</strong> %s</p>\n",
		html.EscapeString(report.Watermark.Tool), html.EscapeString(report.Watermark.Version))
	fmt.Fprintf(&b, "<p>%s</p>\n", html.EscapeString(report.Watermark.Notice))
	fmt.Fprintf(&b, `<p style="margin-top:4pt;">%s &#183; %s</p>`+"\n",
		html.EscapeString(report.Watermark.URL), html.EscapeString(report.Watermark.License))
	b.WriteString("</div>\n")

	b.WriteString("</body>\n</html>")
	return b.String()
}

// htmlConfLevel returns CSS class suffix based on confidence value.
func htmlConfLevel(f float64) string {
	if f > 0.65 {
		return "high"
	}
	if f > 0.35 {
		return "med"
	}
	return "low"
}

// htmlPct formats a 0-1 float as "73%".
func htmlPct(f float64) string {
	return fmt.Sprintf("%.0f%%", f*100)
}

// htmlTensionLabel maps color codes to Portuguese labels.
func htmlTensionLabel(color string) string {
	switch color {
	case "green":
		return "Est&#225;vel"
	case "yellow":
		return "Aten&#231;&#227;o"
	case "orange":
		return "Press&#227;o Alta"
	case "red":
		return "Cr&#237;tico"
	}
	return color
}

// htmlSkillBadge returns (emoji, displayName) for a skill ID.
func htmlSkillBadge(skillID string) (string, string) {
	badges := map[string][2]string{
		"healthcare":    {"🏥", "Healthcare"},
		"fintech":       {"💳", "Fintech"},
		"retail":        {"🛒", "Retail"},
		"legal":         {"⚖️", "Legal"},
		"education":     {"🎓", "Education"},
		"agro":          {"🌱", "Agro"},
		"construction":  {"🏗️", "Construction"},
		"logistics":     {"🚚", "Logistics"},
		"saas":          {"💻", "SaaS"},
		"energy":        {"⚡", "Energy"},
		"manufacturing": {"🏭", "Manufacturing"},
		"media":         {"📺", "Media"},
		"tourism":       {"✈️", "Tourism"},
	}
	if v, ok := badges[skillID]; ok {
		return v[0], v[1]
	}
	return "", ""
}

// htmlPlaybookCol writes a single playbook column div.
func htmlPlaybookCol(b *strings.Builder, title, extraClass string, items []string) {
	if len(items) == 0 {
		return
	}
	cls := "playbook-col"
	if extraClass != "" {
		cls += " " + extraClass
	}
	fmt.Fprintf(b, `<div class="%s"><h4>%s</h4><ul>`, cls, title)
	for _, item := range items {
		fmt.Fprintf(b, "<li>%s</li>", html.EscapeString(item))
	}
	b.WriteString("</ul></div>")
}

// formatInt formats an integer with thousands separators.
func formatInt(n int) string {
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}
	var result []byte
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, byte(c))
	}
	return string(result)
}
