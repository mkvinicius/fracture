// Package deepsearch implements a multi-round deep research agent inspired by DeepSearchAgent.
// It runs BEFORE the FRACTURE simulation to automatically gather real-world context
// (news, competitors, trends, sentiment) and inject it into the simulation world.
//
// Architecture:
//   1. Decompose the user's question into 3-5 research sub-queries
//   2. For each sub-query: search the web, extract key findings
//   3. Reflect: identify gaps, generate follow-up queries
//   4. Repeat up to MaxRounds (default 3) until context is rich enough
//   5. Synthesize all findings into a structured ContextReport
//
// The ContextReport is then injected as extraContext into the FRACTURE simulation,
// giving the 32 agents real-world grounding instead of generic assumptions.
package deepsearch

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// LLMCaller is the interface used to call language models.
type LLMCaller interface {
	Call(ctx context.Context, systemPrompt, userPrompt string, maxTokens int) (string, int, error)
}

// SearchResult holds a single web search result.
type SearchResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
}

// ResearchRound captures what was searched and found in one round.
type ResearchRound struct {
	Round    int            `json:"round"`
	Queries  []string       `json:"queries"`
	Results  []SearchResult `json:"results"`
	Findings string         `json:"findings"` // LLM synthesis of this round
	Gaps     []string       `json:"gaps"`     // identified knowledge gaps
}

// ContextReport is the final output of the DeepSearch agent.
// It is injected into the FRACTURE simulation as extraContext.
type ContextReport struct {
	Question        string          `json:"question"`
	Company         string          `json:"company,omitempty"`
	Sector          string          `json:"sector"`
	KeyPlayers      []string        `json:"key_players"`
	RecentTrends    []string        `json:"recent_trends"`
	Threats         []string        `json:"threats"`
	Opportunities   []string        `json:"opportunities"`
	MarketSentiment string          `json:"market_sentiment"`
	Rounds          []ResearchRound `json:"rounds"`
	Summary         string          `json:"summary"`
	Sources         []string        `json:"sources"`
	GeneratedAt     time.Time       `json:"generated_at"`
	TokensUsed      int             `json:"tokens_used"`
}

// Config controls the deep search agent behaviour.
type Config struct {
	MaxRounds               int           // default 3
	QueriesPerRound         int           // default 4
	Timeout                 time.Duration // default 90s
	SearchAPIKey            string        // optional: Tavily or SerpAPI key
	SearchProvider          string        // "tavily" | "serpapi" | "duckduckgo" (default, free)
	MaxReflectionsPerDomain int           // default 0 (use per-domain defaults); if > 0, use same for all domains
}

// DefaultConfig returns sensible defaults that work without any API key.
func DefaultConfig() Config {
	return Config{
		MaxRounds:               3,
		QueriesPerRound:         4,
		Timeout:                 90 * time.Second,
		SearchProvider:          "duckduckgo",
		MaxReflectionsPerDomain: 0, // use per-domain defaults
	}
}

// domainReflectionDepth defines how many reflection cycles each domain gets by default.
// Regulation and Geopolitics are complex; Behavior and Culture are simpler.
var domainReflectionDepth = map[string]int{
	"regulation":  3,
	"geopolitics": 3,
	"technology":  2,
	"finance":     2,
	"market":      2,
	"behavior":    1,
	"culture":     1,
}

// Agent is the deep search research agent.
type Agent struct {
	llm    LLMCaller
	cfg    Config
	client *http.Client
}

// New creates a new DeepSearch agent.
func New(llm LLMCaller, cfg Config) *Agent {
	if cfg.MaxRounds == 0 {
		cfg.MaxRounds = 3
	}
	if cfg.QueriesPerRound == 0 {
		cfg.QueriesPerRound = 4
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 90 * time.Second
	}
	if cfg.SearchProvider == "" {
		cfg.SearchProvider = "duckduckgo"
	}
	if cfg.MaxReflectionsPerDomain < 0 {
		cfg.MaxReflectionsPerDomain = 0
	}
	return &Agent{
		llm: llm,
		cfg: cfg,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Research runs the full multi-round deep search for a given question.
// It returns a ContextReport ready to be injected into the FRACTURE simulation.
func (a *Agent) Research(ctx context.Context, question, company, sector string) (*ContextReport, error) {
	ctx, cancel := context.WithTimeout(ctx, a.cfg.Timeout)
	defer cancel()

	report := &ContextReport{
		Question:    question,
		Company:     company,
		Sector:      sector,
		GeneratedAt: time.Now().UTC(),
	}

	// Step 1: Decompose the question into initial research queries
	initialQueries, tokens, err := a.decomposeQuestion(ctx, question, company, sector)
	report.TokensUsed += tokens
	if err != nil {
		log.Printf("[DeepSearch] decompose error: %v — using fallback queries", err)
		initialQueries = a.fallbackQueries(question, company, sector)
	}

	allFindings := []string{}
	allSources := []string{}
	currentQueries := initialQueries

	// Step 2: Multi-round research loop
	for round := 1; round <= a.cfg.MaxRounds; round++ {
		select {
		case <-ctx.Done():
			break
		default:
		}

		log.Printf("[DeepSearch] Round %d/%d — %d queries", round, a.cfg.MaxRounds, len(currentQueries))

		rr := ResearchRound{Round: round, Queries: currentQueries}

		// Search for each query
		for _, q := range currentQueries {
			results, err := a.search(ctx, q)
			if err != nil {
				log.Printf("[DeepSearch] search error for %q: %v", q, err)
				continue
			}
			rr.Results = append(rr.Results, results...)
			for _, r := range results {
				allSources = append(allSources, r.URL)
			}
		}

		// Synthesize findings for this round
		findings, gaps, t, err := a.synthesizeRound(ctx, question, rr.Results, allFindings)
		report.TokensUsed += t
		if err != nil {
			log.Printf("[DeepSearch] synthesize error round %d: %v", round, err)
			findings = a.extractSnippets(rr.Results)
		}

		rr.Findings = findings
		rr.Gaps = gaps
		allFindings = append(allFindings, findings)
		report.Rounds = append(report.Rounds, rr)

		// If no gaps identified or last round, stop
		if len(gaps) == 0 || round == a.cfg.MaxRounds {
			break
		}

		// Generate follow-up queries from gaps
		followUp, t2, err := a.generateFollowUpQueries(ctx, gaps, company, sector)
		report.TokensUsed += t2
		if err != nil || len(followUp) == 0 {
			break
		}
		currentQueries = followUp
	}

	// Step 3: Final synthesis into structured ContextReport
	err = a.finalSynthesis(ctx, report, allFindings)
	if err != nil {
		log.Printf("[DeepSearch] final synthesis error: %v — using raw findings", err)
		report.Summary = strings.Join(allFindings, "\n\n")
	}

	// Deduplicate sources
	report.Sources = deduplicate(allSources)

	return report, nil
}

// ToSimulationContext converts the ContextReport into a string
// suitable for injection as extraContext in the FRACTURE simulation.
func (r *ContextReport) ToSimulationContext() string {
	var sb strings.Builder

	sb.WriteString("=== REAL-WORLD CONTEXT (gathered by DeepSearch Agent) ===\n\n")

	if r.Company != "" {
		sb.WriteString(fmt.Sprintf("Company under analysis: %s\n", r.Company))
	}
	if r.Sector != "" {
		sb.WriteString(fmt.Sprintf("Sector: %s\n\n", r.Sector))
	}

	if len(r.KeyPlayers) > 0 {
		sb.WriteString("Key players in this market:\n")
		for _, p := range r.KeyPlayers {
			sb.WriteString(fmt.Sprintf("  - %s\n", p))
		}
		sb.WriteString("\n")
	}

	if len(r.RecentTrends) > 0 {
		sb.WriteString("Recent market trends:\n")
		for _, t := range r.RecentTrends {
			sb.WriteString(fmt.Sprintf("  - %s\n", t))
		}
		sb.WriteString("\n")
	}

	if len(r.Threats) > 0 {
		sb.WriteString("Identified threats:\n")
		for _, t := range r.Threats {
			sb.WriteString(fmt.Sprintf("  - %s\n", t))
		}
		sb.WriteString("\n")
	}

	if len(r.Opportunities) > 0 {
		sb.WriteString("Identified opportunities:\n")
		for _, o := range r.Opportunities {
			sb.WriteString(fmt.Sprintf("  - %s\n", o))
		}
		sb.WriteString("\n")
	}

	if r.MarketSentiment != "" {
		sb.WriteString(fmt.Sprintf("Market sentiment: %s\n\n", r.MarketSentiment))
	}

	if r.Summary != "" {
		sb.WriteString("Research summary:\n")
		sb.WriteString(r.Summary)
		sb.WriteString("\n\n")
	}

	if len(r.Sources) > 0 && len(r.Sources) <= 10 {
		sb.WriteString(fmt.Sprintf("Sources consulted: %d web sources\n", len(r.Sources)))
	}

	sb.WriteString("=== END OF REAL-WORLD CONTEXT ===\n")

	return sb.String()
}

// ─── LLM helpers ─────────────────────────────────────────────────────────────

func (a *Agent) decomposeQuestion(ctx context.Context, question, company, sector string) ([]string, int, error) {
	system := `You are a strategic research analyst. Your task is to decompose a business question into specific web search queries that will gather real-world market intelligence.

Return a JSON array of 4-5 search queries. Each query should target different aspects:
1. Recent news about the company or sector
2. Key competitors and their moves
3. Market trends and disruptions
4. Regulatory or technological changes
5. Customer sentiment or industry analyst views

Return ONLY a JSON array of strings, no explanation.`

	user := fmt.Sprintf(`Question: %s
Company: %s
Sector: %s

Generate 4-5 targeted search queries to research this question.`, question, company, sector)

	raw, tokens, err := a.llm.Call(ctx, system, user, 300)
	if err != nil {
		return nil, tokens, err
	}

	// Extract JSON array from response
	queries, err := extractStringArray(raw)
	if err != nil {
		return nil, tokens, fmt.Errorf("parse queries: %w", err)
	}

	return queries, tokens, nil
}

func (a *Agent) synthesizeRound(ctx context.Context, question string, results []SearchResult, previousFindings []string) (findings string, gaps []string, tokens int, err error) {
	if len(results) == 0 {
		return "No results found for this round.", nil, 0, nil
	}

	// Build snippets text
	var snippets strings.Builder
	for i, r := range results {
		if i >= 12 {
			break
		}
		snippets.WriteString(fmt.Sprintf("[%d] %s\n%s\n\n", i+1, r.Title, r.Snippet))
	}

	prevContext := ""
	if len(previousFindings) > 0 {
		prevContext = "Previous findings:\n" + strings.Join(previousFindings, "\n") + "\n\n"
	}

	system := `You are a strategic market intelligence analyst. Synthesize the search results into key findings relevant to the research question. Be concise and factual.

Return a JSON object with:
{
  "findings": "2-3 paragraph synthesis of key insights",
  "gaps": ["list of important questions still unanswered", "..."]
}

Return ONLY valid JSON.`

	user := fmt.Sprintf(`Research question: %s

%sSearch results:
%s

Synthesize the findings and identify knowledge gaps.`, question, prevContext, snippets.String())

	raw, t, err := a.llm.Call(ctx, system, user, 600)
	tokens = t
	if err != nil {
		return "", nil, tokens, err
	}

	// Parse response
	var resp struct {
		Findings string   `json:"findings"`
		Gaps     []string `json:"gaps"`
	}
	if err := json.Unmarshal([]byte(extractJSON(raw)), &resp); err != nil {
		// Return error so the caller can fall back to extractSnippets — do NOT
		// return raw LLM output as findings; it would contaminate agent prompts.
		return "", nil, tokens, fmt.Errorf("parse synthesis response: %w", err)
	}

	return resp.Findings, resp.Gaps, tokens, nil
}

func (a *Agent) generateFollowUpQueries(ctx context.Context, gaps []string, company, sector string) ([]string, int, error) {
	system := `You are a research analyst. Convert knowledge gaps into specific web search queries.
Return ONLY a JSON array of 3-4 search query strings.`

	user := fmt.Sprintf(`Company: %s | Sector: %s
Knowledge gaps to fill:
%s

Generate targeted search queries to fill these gaps.`, company, sector, strings.Join(gaps, "\n"))

	raw, tokens, err := a.llm.Call(ctx, system, user, 200)
	if err != nil {
		return nil, tokens, err
	}

	queries, err := extractStringArray(raw)
	return queries, tokens, err
}

func (a *Agent) finalSynthesis(ctx context.Context, report *ContextReport, allFindings []string) error {
	if len(allFindings) == 0 {
		return fmt.Errorf("no findings to synthesize")
	}

	system := `You are a senior strategic analyst. Based on research findings, produce a structured market intelligence report.

Return a JSON object with these exact fields:
{
  "key_players": ["list of 3-6 key market players mentioned"],
  "recent_trends": ["list of 3-5 recent market trends"],
  "threats": ["list of 2-4 key threats"],
  "opportunities": ["list of 2-4 key opportunities"],
  "market_sentiment": "one sentence describing overall market sentiment",
  "summary": "3-4 paragraph executive summary of all findings"
}

Return ONLY valid JSON.`

	user := fmt.Sprintf(`Question: %s
Company: %s
Sector: %s

Research findings from %d rounds:
%s

Produce the structured market intelligence report.`,
		report.Question, report.Company, report.Sector,
		len(allFindings), strings.Join(allFindings, "\n\n---\n\n"))

	raw, tokens, err := a.llm.Call(ctx, system, user, 800)
	report.TokensUsed += tokens
	if err != nil {
		return err
	}

	var resp struct {
		KeyPlayers      []string `json:"key_players"`
		RecentTrends    []string `json:"recent_trends"`
		Threats         []string `json:"threats"`
		Opportunities   []string `json:"opportunities"`
		MarketSentiment string   `json:"market_sentiment"`
		Summary         string   `json:"summary"`
	}
	if err := json.Unmarshal([]byte(extractJSON(raw)), &resp); err != nil {
		report.Summary = strings.Join(allFindings, "\n\n")
		return nil
	}

	report.KeyPlayers = resp.KeyPlayers
	report.RecentTrends = resp.RecentTrends
	report.Threats = resp.Threats
	report.Opportunities = resp.Opportunities
	report.MarketSentiment = resp.MarketSentiment
	report.Summary = resp.Summary

	return nil
}

// ─── Search providers ─────────────────────────────────────────────────────────

func (a *Agent) search(ctx context.Context, query string) ([]SearchResult, error) {
	switch a.cfg.SearchProvider {
	case "tavily":
		return a.searchTavily(ctx, query)
	case "serpapi":
		return a.searchSerpAPI(ctx, query)
	default:
		return a.searchDuckDuckGo(ctx, query)
	}
}

// searchDuckDuckGo uses DuckDuckGo Instant Answer API (free, no key required).
func (a *Agent) searchDuckDuckGo(ctx context.Context, query string) ([]SearchResult, error) {
	encoded := url.QueryEscape(query)
	apiURL := fmt.Sprintf("https://api.duckduckgo.com/?q=%s&format=json&no_html=1&skip_disambig=1", encoded)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "FRACTURE-DeepSearch/1.4.0")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if err != nil {
		return nil, err
	}

	var ddg struct {
		AbstractText   string `json:"AbstractText"`
		AbstractSource string `json:"AbstractSource"`
		AbstractURL    string `json:"AbstractURL"`
		RelatedTopics  []struct {
			Text     string `json:"Text"`
			FirstURL string `json:"FirstURL"`
		} `json:"RelatedTopics"`
	}

	if err := json.Unmarshal(body, &ddg); err != nil {
		return nil, err
	}

	var results []SearchResult

	if ddg.AbstractText != "" {
		results = append(results, SearchResult{
			Title:   ddg.AbstractSource,
			URL:     ddg.AbstractURL,
			Snippet: ddg.AbstractText,
		})
	}

	for i, t := range ddg.RelatedTopics {
		if i >= 6 {
			break
		}
		if t.Text != "" {
			results = append(results, SearchResult{
				Title:   extractTitle(t.Text),
				URL:     t.FirstURL,
				Snippet: t.Text,
			})
		}
	}

	return results, nil
}

// searchTavily uses Tavily Search API (requires API key, higher quality).
func (a *Agent) searchTavily(ctx context.Context, query string) ([]SearchResult, error) {
	if a.cfg.SearchAPIKey == "" {
		return a.searchDuckDuckGo(ctx, query)
	}

	payload := fmt.Sprintf(`{"api_key":"%s","query":%q,"search_depth":"basic","max_results":5}`,
		a.cfg.SearchAPIKey, query)

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.tavily.com/search",
		strings.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if err != nil {
		return nil, err
	}

	var tv struct {
		Results []struct {
			Title   string `json:"title"`
			URL     string `json:"url"`
			Content string `json:"content"`
		} `json:"results"`
	}

	if err := json.Unmarshal(body, &tv); err != nil {
		return nil, err
	}

	var results []SearchResult
	for _, r := range tv.Results {
		results = append(results, SearchResult{
			Title:   r.Title,
			URL:     r.URL,
			Snippet: r.Content,
		})
	}

	return results, nil
}

// searchSerpAPI uses SerpAPI (requires API key).
func (a *Agent) searchSerpAPI(ctx context.Context, query string) ([]SearchResult, error) {
	if a.cfg.SearchAPIKey == "" {
		return a.searchDuckDuckGo(ctx, query)
	}

	encoded := url.QueryEscape(query)
	apiURL := fmt.Sprintf("https://serpapi.com/search.json?q=%s&api_key=%s&num=5",
		encoded, a.cfg.SearchAPIKey)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if err != nil {
		return nil, err
	}

	var serp struct {
		OrganicResults []struct {
			Title   string `json:"title"`
			Link    string `json:"link"`
			Snippet string `json:"snippet"`
		} `json:"organic_results"`
	}

	if err := json.Unmarshal(body, &serp); err != nil {
		return nil, err
	}

	var results []SearchResult
	for _, r := range serp.OrganicResults {
		results = append(results, SearchResult{
			Title:   r.Title,
			URL:     r.Link,
			Snippet: r.Snippet,
		})
	}

	return results, nil
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func (a *Agent) fallbackQueries(question, company, sector string) []string {
	queries := []string{
		fmt.Sprintf("%s market trends 2024 2025", sector),
		fmt.Sprintf("%s competitors analysis", company),
		fmt.Sprintf("%s industry disruption news", sector),
		fmt.Sprintf("%s strategic challenges opportunities", company),
	}
	return queries
}

func (a *Agent) extractSnippets(results []SearchResult) string {
	var sb strings.Builder
	for i, r := range results {
		if i >= 5 {
			break
		}
		sb.WriteString(r.Snippet)
		sb.WriteString(" ")
	}
	return sb.String()
}

func extractStringArray(raw string) ([]string, error) {
	// Find JSON array in response
	start := strings.Index(raw, "[")
	end := strings.LastIndex(raw, "]")
	if start == -1 || end == -1 || end <= start {
		return nil, fmt.Errorf("no JSON array found in: %s", raw)
	}

	var arr []string
	if err := json.Unmarshal([]byte(raw[start:end+1]), &arr); err != nil {
		return nil, err
	}
	return arr, nil
}

func extractJSON(raw string) string {
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start == -1 || end == -1 || end <= start {
		return raw
	}
	return raw[start : end+1]
}

func extractTitle(text string) string {
	if len(text) > 60 {
		return text[:60] + "..."
	}
	return text
}

func deduplicate(items []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, item := range items {
		if !seen[item] && item != "" {
			seen[item] = true
			result = append(result, item)
		}
	}
	return result
}


// SynthesizeDomainContext extracts domain-specific context from a ContextReport
// and returns a map of domain -> (context_text, affected_rules, confidence).
// This is used to inject real-world evidence into the FRACTURE simulation.
func (a *Agent) SynthesizeDomainContext(report *ContextReport) map[string]struct {
	ContextText   string
	AffectedRules []string
	Confidence    float64
} {
	result := make(map[string]struct {
		ContextText   string
		AffectedRules []string
		Confidence    float64
	})

	// Market domain: key players, threats, opportunities, sentiment
	if len(report.KeyPlayers) > 0 || len(report.Threats) > 0 {
		ctx := struct {
			ContextText   string
			AffectedRules []string
			Confidence    float64
		}{
			ContextText: fmt.Sprintf(
				"Key Players: %s | Threats: %s | Opportunities: %s | Sentiment: %s",
				strings.Join(report.KeyPlayers, ", "),
				strings.Join(report.Threats, ", "),
				strings.Join(report.Opportunities, ", "),
				report.MarketSentiment,
			),
			AffectedRules: []string{"mkt-001", "mkt-002", "mkt-004", "mkt-005", "mkt-007"},
			Confidence:    0.75,
		}
		result["market"] = ctx
	}

	// Technology domain: trends
	if len(report.RecentTrends) > 0 {
		ctx := struct {
			ContextText   string
			AffectedRules []string
			Confidence    float64
		}{
			ContextText: fmt.Sprintf("Technology Trends: %s", strings.Join(report.RecentTrends, ", ")),
			AffectedRules: []string{"tech-001", "tech-003", "tech-006", "tech-007"},
			Confidence:    0.70,
		}
		result["technology"] = ctx
	}

	// Regulation domain: threats often include regulatory changes
	if len(report.Threats) > 0 {
		threatText := strings.Join(report.Threats, ", ")
		if strings.Contains(strings.ToLower(threatText), "regulat") || strings.Contains(strings.ToLower(threatText), "compliance") {
			ctx := struct {
				ContextText   string
				AffectedRules []string
				Confidence    float64
			}{
				ContextText:   fmt.Sprintf("Regulatory Threats: %s", threatText),
				AffectedRules: []string{"reg-001", "reg-003", "reg-005", "reg-006"},
				Confidence:    0.65,
			}
			result["regulation"] = ctx
		}
	}

	// Behavior domain: market sentiment and opportunities
	if report.MarketSentiment != "" {
		ctx := struct {
			ContextText   string
			AffectedRules []string
			Confidence    float64
		}{
			ContextText:   fmt.Sprintf("Market Sentiment: %s", report.MarketSentiment),
			AffectedRules: []string{"beh-003", "beh-005", "beh-007"},
			Confidence:    0.60,
		}
		result["behavior"] = ctx
	}

	// Culture domain: recent trends and sentiment
	if len(report.RecentTrends) > 0 {
		ctx := struct {
			ContextText   string
			AffectedRules []string
			Confidence    float64
		}{
			ContextText:   fmt.Sprintf("Cultural Trends: %s | Sentiment: %s", strings.Join(report.RecentTrends, ", "), report.MarketSentiment),
			AffectedRules: []string{"cul-002", "cul-004", "cul-005", "cul-006"},
			Confidence:    0.65,
		}
		result["culture"] = ctx
	}

	// Finance domain: threats and opportunities
	if len(report.Opportunities) > 0 || len(report.Threats) > 0 {
		ctx := struct {
			ContextText   string
			AffectedRules []string
			Confidence    float64
		}{
			ContextText:   fmt.Sprintf("Financial Opportunities: %s | Threats: %s", strings.Join(report.Opportunities, ", "), strings.Join(report.Threats, ", ")),
			AffectedRules: []string{"fin-001", "fin-003", "fin-005", "fin-006"},
			Confidence:    0.70,
		}
		result["finance"] = ctx
	}

	return result
}


// buildDomainQueries generates initial search queries for a specific domain.
func (a *Agent) buildDomainQueries(domain, question, company, sector string) []string {
	domainKeywords := map[string][]string{
		"market":      {"market trends", "competitive landscape", "market disruption", "market share"},
		"technology":  {"technology trends", "innovation", "digital transformation", "tech adoption"},
		"regulation":  {"regulatory changes", "compliance", "government policy", "legal framework"},
		"behavior":    {"consumer behavior", "user adoption", "market adoption", "behavioral trends"},
		"culture":     {"cultural trends", "social trends", "cultural shift", "societal changes"},
		"geopolitics": {"geopolitical risks", "trade policy", "international relations", "political risks"},
		"finance":     {"financial markets", "funding trends", "investment", "financial outlook"},
	}

	keywords, ok := domainKeywords[domain]
	if !ok {
		keywords = []string{"market trends", "industry analysis"}
	}

	var queries []string
	for _, kw := range keywords {
		queries = append(queries, fmt.Sprintf("%s %s %s %d", company, kw, sector, time.Now().Year()))
	}
	return queries
}

// buildComplementaryQueries generates follow-up queries based on identified gaps.
func (a *Agent) buildComplementaryQueries(domain string, gaps []string) []string {
	var queries []string
	for _, gap := range gaps {
		queries = append(queries, fmt.Sprintf("%s %s latest news", gap, domain))
	}
	return queries
}

// synthesizeDomainFindings uses the LLM to synthesize findings for a domain.
func (a *Agent) synthesizeDomainFindings(ctx context.Context, domain, question string, queries []string) (string, int, error) {
	prompt := fmt.Sprintf(
		"Synthesize the key findings for the %s domain in response to: %s\nSearch queries used: %s\nProvide a concise summary of the most important insights.",
		domain, question, strings.Join(queries, "; "),
	)

	findings, tokens, err := a.llm.Call(ctx, "You are a domain expert synthesizing research findings.", prompt, 500)
	return findings, tokens, err
}

// identifyGaps uses the LLM to identify knowledge gaps in the current findings.
func (a *Agent) identifyGaps(ctx context.Context, domain, question string, findings []string) ([]string, int, error) {
	prompt := fmt.Sprintf(
		"Given the question '%s' and these findings about the %s domain:\n%s\n\nIdentify 2-3 critical knowledge gaps that should be researched further. Return as a JSON array of strings.",
		question, domain, strings.Join(findings, "\n"),
	)

	response, tokens, err := a.llm.Call(ctx, "You are a research analyst identifying gaps.", prompt, 300)
	if err != nil {
		return nil, tokens, err
	}

	gaps, err := extractStringArray(response)
	if err != nil {
		log.Printf("[DeepSearch] failed to parse gaps: %v", err)
		return nil, tokens, nil
	}

	return gaps, tokens, nil
}


// hashQuestion generates a stable hash for a research question.
// Used as key for resumable state in the database.
func hashQuestion(question, company, sector string) string {
	h := sha256.Sum256([]byte(question + "|" + company + "|" + sector))
	return hex.EncodeToString(h[:8]) // First 16 hex chars (8 bytes)
}
