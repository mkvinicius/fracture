package deepsearch

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// CausalTriple representa uma relação causal extraída do DeepSearch.
// Exemplo: "mudança regulatória ANS" → "impacto em precificação de planos" (0.75)
type CausalTriple struct {
	Cause      string
	Effect     string
	Confidence float64
	Domain     string
}

// ExtractCausalities usa LLM para extrair triplas causais de um ContextReport.
// Retorna lista vazia sem erro se a extração falhar — nunca bloqueia o fluxo principal.
func ExtractCausalities(ctx context.Context, llm LLMCaller, report *ContextReport, sector string) []CausalTriple {
	if report == nil {
		return nil
	}

	prompt := fmt.Sprintf(`Você é um analista estratégico. Analise o contexto de mercado abaixo e extraia relações causais concretas.

SETOR: %s

CONTEXTO:
- Tendências: %s
- Ameaças: %s
- Oportunidades: %s
- Players principais: %s

Extraia até 10 relações causais no formato JSON:
[
  {
    "cause": "evento ou condição que causa",
    "effect": "impacto ou consequência",
    "confidence": 0.0-1.0,
    "domain": "market|technology|regulation|behavior|culture|finance"
  }
]

Retorne APENAS o JSON, sem texto adicional.`,
		sector,
		strings.Join(report.RecentTrends, "; "),
		strings.Join(report.Threats, "; "),
		strings.Join(report.Opportunities, "; "),
		strings.Join(report.KeyPlayers, "; "),
	)

	resp, _, err := llm.Call(ctx, prompt, "", 800)
	if err != nil {
		return nil
	}

	resp = extractJSONArray(resp)
	if resp == "" {
		return nil
	}

	var raw []struct {
		Cause      string  `json:"cause"`
		Effect     string  `json:"effect"`
		Confidence float64 `json:"confidence"`
		Domain     string  `json:"domain"`
	}

	if err := json.Unmarshal([]byte(resp), &raw); err != nil {
		return nil
	}

	triples := make([]CausalTriple, 0, len(raw))
	for _, r := range raw {
		if r.Cause == "" || r.Effect == "" {
			continue
		}
		if r.Confidence < 0.3 {
			continue // descarta baixa confiança
		}
		triples = append(triples, CausalTriple{
			Cause:      r.Cause,
			Effect:     r.Effect,
			Confidence: r.Confidence,
			Domain:     r.Domain,
		})
	}
	return triples
}

// extractJSONArray extrai o primeiro array JSON de uma string.
func extractJSONArray(s string) string {
	start := strings.Index(s, "[")
	end := strings.LastIndex(s, "]")
	if start == -1 || end == -1 || end <= start {
		return ""
	}
	return s[start : end+1]
}
