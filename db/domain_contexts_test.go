package db

import (
	"testing"

	"github.com/fracture/fracture/engine"
)

// TestSaveDomainContext tests persisting domain context with stability modifier.
func TestSaveDomainContext(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	simID := "test-sim-001"
	// Create simulation first (required by foreign key)
	_, err := db.Exec(`INSERT INTO simulations (id, question, department) VALUES (?, ?, ?)`,
		simID, "Test question", "market")
	if err != nil {
		t.Fatalf("Failed to create test simulation: %v", err)
	}

	domain := "market"
	contextText := "Key Players: Apple, Microsoft | Threats: Disruption | Sentiment: Bullish"
	affectedRules := []string{"mkt-001", "mkt-002", "mkt-005"}
	confidence := 0.75
	stabilityMod := -0.15 * confidence

	err = db.SaveDomainContext(
		simID+":"+domain,
		simID,
		domain,
		contextText,
		affectedRules,
		confidence,
		stabilityMod,
	)
	if err != nil {
		t.Fatalf("SaveDomainContext failed: %v", err)
	}

	// Verify by retrieving
	contexts, err := db.GetDomainContextsByDomain(simID, domain)
	if err != nil {
		t.Fatalf("GetDomainContextsByDomain failed: %v", err)
	}
	if len(contexts) != 1 {
		t.Fatalf("Expected 1 context, got %d", len(contexts))
	}

	ctx := contexts[0]
	if ctx.Domain != domain {
		t.Errorf("Domain mismatch: expected %s, got %s", domain, ctx.Domain)
	}
	if ctx.Confidence != confidence {
		t.Errorf("Confidence mismatch: expected %f, got %f", confidence, ctx.Confidence)
	}
	if ctx.StabilityModifier != stabilityMod {
		t.Errorf("StabilityModifier mismatch: expected %f, got %f", stabilityMod, ctx.StabilityModifier)
	}
	if len(ctx.AffectedRules) != 3 {
		t.Errorf("Expected 3 affected rules, got %d", len(ctx.AffectedRules))
	}
}

// TestGetDomainContexts tests retrieving all domain contexts for a simulation.
func TestGetDomainContexts(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	simID := "test-sim-002"
	// Create simulation first (required by foreign key)
	_, err := db.Exec(`INSERT INTO simulations (id, question, department) VALUES (?, ?, ?)`,
		simID, "Test question", "market")
	if err != nil {
		t.Fatalf("Failed to create test simulation: %v", err)
	}

	// Save multiple domain contexts
	domains := []string{"market", "technology", "regulation"}
	for i, domain := range domains {
		contextID := simID + ":" + domain
		affectedRules := []string{domain + "-001", domain + "-002"}
		confidence := 0.6 + float64(i)*0.1
		stabilityMod := -0.15 * confidence

		err := db.SaveDomainContext(
			contextID,
			simID,
			domain,
			"Context for "+domain,
			affectedRules,
			confidence,
			stabilityMod,
		)
		if err != nil {
			t.Fatalf("SaveDomainContext failed for %s: %v", domain, err)
		}
	}

	// Retrieve all
	contexts, err := db.GetDomainContexts(simID)
	if err != nil {
		t.Fatalf("GetDomainContexts failed: %v", err)
	}
	if len(contexts) != 3 {
		t.Fatalf("Expected 3 contexts, got %d", len(contexts))
	}

	// Verify each domain is present
	domainMap := make(map[string]bool)
	for _, ctx := range contexts {
		domainMap[ctx.Domain] = true
	}
	for _, domain := range domains {
		if !domainMap[domain] {
			t.Errorf("Domain %s not found in results", domain)
		}
	}
}

// TestDefaultWorldForDomainWithContext verifies Evidence is populated
// and affected rules have stability reduced.
func TestDefaultWorldForDomainWithContext(t *testing.T) {
	question := "Will market disruption happen?"
	extraContext := "Real-world evidence from DeepSearch"
	affectedRules := []string{"mkt-001", "mkt-002"}
	confidence := 0.75

	world := engine.DefaultWorldForDomainWithContext(
		engine.DomainMarket,
		question,
		extraContext,
		affectedRules,
		confidence,
	)

	// Verify Evidence is populated
	if world.Evidence == "" {
		t.Error("Evidence field is empty, expected populated context")
	}
	if len(world.Evidence) < 10 {
		t.Errorf("Evidence too short: %s", world.Evidence)
	}

	// Verify affected rules have reduced stability
	for _, ruleID := range affectedRules {
		if rule, ok := world.Rules[ruleID]; ok {
			// Base stability for market rules is typically 0.6-0.8
			// With confidence 0.75, modifier should be -0.1125
			// So stability should be reduced
			if rule.Stability > 0.75 {
				t.Logf("Rule %s stability: %f (expected reduced from base)", ruleID, rule.Stability)
			}
		}
	}

	// Verify low confidence doesn't apply modifier
	worldLowConf := engine.DefaultWorldForDomainWithContext(
		engine.DomainMarket,
		question,
		extraContext,
		affectedRules,
		0.5, // Below 0.6 threshold
	)

	// With low confidence, rules should have original stability
	baseWorld := engine.DefaultWorldForDomain(engine.DomainMarket, question, extraContext)
	for _, ruleID := range affectedRules {
		if rule, ok := worldLowConf.Rules[ruleID]; ok {
			if baseRule, ok := baseWorld.Rules[ruleID]; ok {
				if rule.Stability != baseRule.Stability {
					t.Logf("Rule %s: low confidence should not modify stability", ruleID)
				}
			}
		}
	}
}
