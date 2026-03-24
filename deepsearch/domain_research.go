package deepsearch

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/fracture/fracture/engine"
)

const maxDomainConcurrency = 3

// DomainResearchResult holds the structured research output for a single domain.
// AffectedRules lists the rule IDs (e.g. "mkt-003") that the research findings
// indicate are under pressure or disruption. Confidence is a 0.0–1.0 score
// derived from source count and research depth.
type DomainResearchResult struct {
	Domain           engine.RuleDomain `json:"domain"`
	AffectedRules    []string          `json:"affected_rules"`
	Confidence       float64           `json:"confidence"`
	Summary          string            `json:"summary"`
	KeySignals       []string          `json:"key_signals"`
	Threats          []string          `json:"threats"`
	Opportunities    []string          `json:"opportunities"`
	SynthesizedContext string          `json:"synthesized_context"`
	CachedAt         time.Time         `json:"cached_at"`
}

// DomainResearcher enriches engine domains with real-world context via the
// DeepSearch Agent. It enforces a concurrency limit of maxDomainConcurrency
// concurrent in-flight research calls and caches results per domain in SQLite
// so that repeated calls for the same domain are served instantly.
type DomainResearcher struct {
	agent *Agent
	db    *sql.DB
	sem   chan struct{} // capacity = maxDomainConcurrency
	once  sync.Once    // ensures the cache table is created exactly once
}

// NewDomainResearcher returns a DomainResearcher backed by the given Agent
// and SQLite database handle.
func NewDomainResearcher(agent *Agent, sqlDB *sql.DB) *DomainResearcher {
	return &DomainResearcher{
		agent: agent,
		db:    sqlDB,
		sem:   make(chan struct{}, maxDomainConcurrency),
	}
}

// ensureTable creates the cache table on first use (idempotent).
func (dr *DomainResearcher) ensureTable() error {
	var execErr error
	dr.once.Do(func() {
		_, execErr = dr.db.Exec(`
			CREATE TABLE IF NOT EXISTS domain_research_cache (
				domain      TEXT    NOT NULL PRIMARY KEY,
				result_json TEXT    NOT NULL,
				cached_at   INTEGER NOT NULL
			)
		`)
	})
	return execErr
}

// ResearchDomain returns enriched research for a single domain.
// A cached entry is returned without touching the semaphore. On a cache miss
// the call blocks until a semaphore slot is available (up to maxDomainConcurrency
// concurrent callers), performs live research, then writes the result to cache.
func (dr *DomainResearcher) ResearchDomain(ctx context.Context, domain engine.RuleDomain, question, company, sector string) (*DomainResearchResult, error) {
	if err := dr.ensureTable(); err != nil {
		return nil, fmt.Errorf("ensure cache table: %w", err)
	}

	// Fast path: cache hit (no semaphore needed — read-only)
	if cached, err := dr.loadCache(string(domain)); err == nil && cached != nil {
		log.Printf("[DomainResearcher] cache hit for domain %q", domain)
		return cached, nil
	}

	// Acquire semaphore slot before any LLM/search work.
	select {
	case dr.sem <- struct{}{}:
		defer func() { <-dr.sem }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// Double-check cache after acquiring the slot — another goroutine may have
	// populated it while we were waiting.
	if cached, err := dr.loadCache(string(domain)); err == nil && cached != nil {
		return cached, nil
	}

	report, err := dr.agent.Research(ctx, question, company, sector)
	if err != nil {
		return nil, fmt.Errorf("research domain %q: %w", domain, err)
	}

	result := buildDomainResult(domain, report)

	if saveErr := dr.saveCache(string(domain), result); saveErr != nil {
		log.Printf("[DomainResearcher] cache write error for domain %q: %v", domain, saveErr)
	}

	return result, nil
}

// ResearchDomains runs ResearchDomain for every domain concurrently, honouring
// the maxDomainConcurrency semaphore, and returns a map of domain → result.
// The first non-nil error encountered is returned; successful results are still
// included in the map.
func (dr *DomainResearcher) ResearchDomains(
	ctx context.Context,
	domains []engine.RuleDomain,
	question, company, sector string,
) (map[engine.RuleDomain]*DomainResearchResult, error) {
	type item struct {
		domain engine.RuleDomain
		result *DomainResearchResult
		err    error
	}

	ch := make(chan item, len(domains))

	var wg sync.WaitGroup
	for _, d := range domains {
		wg.Add(1)
		go func(dom engine.RuleDomain) {
			defer wg.Done()
			res, err := dr.ResearchDomain(ctx, dom, question, company, sector)
			ch <- item{dom, res, err}
		}(d)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	results := make(map[engine.RuleDomain]*DomainResearchResult, len(domains))
	var firstErr error
	for it := range ch {
		if it.err != nil && firstErr == nil {
			firstErr = it.err
		}
		if it.result != nil {
			results[it.domain] = it.result
		}
	}

	return results, firstErr
}

// ─── cache helpers ─────────────────────────────────────────────────────────────

func (dr *DomainResearcher) loadCache(domain string) (*DomainResearchResult, error) {
	var rawJSON string
	err := dr.db.QueryRow(
		`SELECT result_json FROM domain_research_cache WHERE domain = ?`, domain,
	).Scan(&rawJSON)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var r DomainResearchResult
	if err := json.Unmarshal([]byte(rawJSON), &r); err != nil {
		return nil, err
	}
	return &r, nil
}

func (dr *DomainResearcher) saveCache(domain string, r *DomainResearchResult) error {
	b, err := json.Marshal(r)
	if err != nil {
		return err
	}
	_, err = dr.db.Exec(`
		INSERT INTO domain_research_cache (domain, result_json, cached_at)
		VALUES (?, ?, unixepoch())
		ON CONFLICT(domain) DO UPDATE SET
			result_json = excluded.result_json,
			cached_at   = excluded.cached_at
	`, domain, string(b))
	return err
}

// ─── result builder ────────────────────────────────────────────────────────────

// buildDomainResult maps a ContextReport onto a DomainResearchResult.
func buildDomainResult(domain engine.RuleDomain, report *ContextReport) *DomainResearchResult {
	r := &DomainResearchResult{
		Domain:        domain,
		AffectedRules: extractAffectedRules(domain, report),
		Confidence:    computeConfidence(report),
		Summary:       report.Summary,
		KeySignals:    report.RecentTrends,
		Threats:       report.Threats,
		Opportunities: report.Opportunities,
		CachedAt:      time.Now().UTC(),
	}
	r.SynthesizedContext = synthesizeDomainContext(domain, r)
	return r
}

// synthesizeDomainContext builds a structured markdown context string from a
// DomainResearchResult. The output is stored in SynthesizedContext and injected
// as World.Evidence so agents can reference it without it acting as a Rule.
func synthesizeDomainContext(domain engine.RuleDomain, r *DomainResearchResult) string {
	var b strings.Builder
	fmt.Fprintf(&b, "## Domain: %s\n", domain)

	b.WriteString("### Key Signals:\n")
	for _, s := range r.KeySignals {
		fmt.Fprintf(&b, "- %s\n", s)
	}
	if len(r.KeySignals) == 0 {
		b.WriteString("- (no signals)\n")
	}

	b.WriteString("### Threats:\n")
	for _, t := range r.Threats {
		fmt.Fprintf(&b, "- %s\n", t)
	}
	if len(r.Threats) == 0 {
		b.WriteString("- (none identified)\n")
	}

	b.WriteString("### Opportunities:\n")
	for _, o := range r.Opportunities {
		fmt.Fprintf(&b, "- %s\n", o)
	}
	if len(r.Opportunities) == 0 {
		b.WriteString("- (none identified)\n")
	}

	return b.String()
}

// buildDomainQueries returns 2 domain-specialised search queries for the given
// domain, company, and sector. These can be used to direct research agents
// toward domain-specific signals rather than generic context.
func buildDomainQueries(domain engine.RuleDomain, _, company, sector string) []string {
	switch domain {
	case engine.DomainMarket:
		return []string{"competitive landscape " + sector, "pricing pressure " + company}
	case engine.DomainTechnology:
		return []string{"technology disruption " + sector, "AI adoption " + company}
	case engine.DomainRegulation:
		return []string{"regulatory changes " + sector, "compliance " + company}
	case engine.DomainBehavior:
		return []string{"workforce trends " + sector, "talent market " + company}
	case engine.DomainCulture:
		return []string{"consumer sentiment " + sector, "brand trust " + company}
	case engine.DomainGeopolitics:
		return []string{"geopolitical risk " + sector, "supply chain " + company}
	case engine.DomainFinance:
		return []string{"funding trends " + sector, "capital allocation " + company}
	default:
		return []string{sector + " market overview", company + " competitive position"}
	}
}

// ResearchAllDomains is a convenience wrapper that researches all 7 engine
// domains concurrently, honouring the maxDomainConcurrency semaphore.
func (dr *DomainResearcher) ResearchAllDomains(
	ctx context.Context,
	question, company, sector string,
) (map[engine.RuleDomain]*DomainResearchResult, error) {
	return dr.ResearchDomains(ctx, []engine.RuleDomain{
		engine.DomainMarket,
		engine.DomainTechnology,
		engine.DomainRegulation,
		engine.DomainBehavior,
		engine.DomainCulture,
		engine.DomainGeopolitics,
		engine.DomainFinance,
	}, question, company, sector)
}

// extractAffectedRules infers which canonical rule IDs for the given domain are
// most affected by the research findings. It scans threats, opportunities, and
// recent trends for explicit rule-ID mentions (e.g. "mkt-003"). When no explicit
// IDs are found it falls back to the first N default rule IDs for the domain.
func extractAffectedRules(domain engine.RuleDomain, report *ContextReport) []string {
	prefix := domainPrefix(domain)
	if prefix == "" {
		return nil
	}

	signals := append(report.Threats, report.Opportunities...)
	signals = append(signals, report.RecentTrends...)

	seen := make(map[string]bool)
	var rules []string

	for _, sig := range signals {
		for i := 1; i <= 12; i++ {
			id := fmt.Sprintf("%s-%03d", prefix, i)
			if strings.Contains(sig, id) && !seen[id] {
				seen[id] = true
				rules = append(rules, id)
			}
		}
	}

	// Fallback: return the default rule IDs for the domain.
	if len(rules) == 0 {
		rules = defaultRuleIDs(prefix)
	}

	return rules
}

func defaultRuleIDs(prefix string) []string {
	ids := make([]string, 0, 8)
	for i := 1; i <= 8; i++ {
		ids = append(ids, fmt.Sprintf("%s-%03d", prefix, i))
	}
	return ids
}

func domainPrefix(d engine.RuleDomain) string {
	switch d {
	case engine.DomainMarket:
		return "mkt"
	case engine.DomainTechnology:
		return "tech"
	case engine.DomainRegulation:
		return "reg"
	case engine.DomainBehavior:
		return "beh"
	case engine.DomainCulture:
		return "cul"
	case engine.DomainGeopolitics:
		return "geo"
	case engine.DomainFinance:
		return "fin"
	default:
		return ""
	}
}

// computeConfidence derives a 0.0–1.0 confidence score from research quality.
// Rounds contribute up to 0.5 and source diversity contributes up to 0.5.
func computeConfidence(report *ContextReport) float64 {
	rounds := float64(len(report.Rounds))
	sources := float64(len(report.Sources))

	roundScore := clamp(rounds/6.0, 0, 0.5)
	sourceScore := clamp(sources/20.0, 0, 0.5)

	return roundScore + sourceScore
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
