package skills

import (
	"testing"
)

func TestSkillRegistry(t *testing.T) {
	expectedIDs := []string{
		"healthcare", "fintech", "retail", "legal", "education",
		"agro", "construction", "logistics", "saas", "energy",
		"manufacturing", "media", "tourism",
	}
	for _, id := range expectedIDs {
		sk, ok := Registry[id]
		if !ok {
			t.Errorf("skill %q not found in Registry", id)
			continue
		}
		if sk.ID != id {
			t.Errorf("skill %q has wrong ID: got %q", id, sk.ID)
		}
		if sk.Name == "" {
			t.Errorf("skill %q has empty Name", id)
		}
		if len(sk.Rules) == 0 {
			t.Errorf("skill %q has no Rules", id)
		}
		if len(sk.Agents) == 0 {
			t.Errorf("skill %q has no Agents", id)
		}
	}
}

func TestSkillDetect(t *testing.T) {
	tests := []struct {
		question   string
		department string
		wantSkill  string
	}{
		{"How will ANVISA regulation affect our hospital expansion?", "Strategy", "healthcare"},
		{"If PIX kills our revenue, what's the future of fintech?", "Finance", "fintech"},
		{"How would a marketplace like Mercado Livre disrupt retail?", "Sales", "retail"},
		{"What happens if OAB bans AI in legal practice?", "Operations", "legal"},
		{"How will EAD affect university credentialism?", "Product", "education"},
		{"What if soja exports to China are blocked by EU deforestation rules?", "Strategy", "agro"},
		{"How would rising SELIC rates affect our incorporadora strategy?", "Finance", "construction"},
		{"If ANTT removes the tabelamento de fretes, what happens to logistics?", "Operations", "logistics"},
		{"How would TOTVS entering our ERP niche affect our SaaS growth?", "Product", "saas"},
		{"How will distributed solar generation affect our energia distribuída margins?", "Strategy", "energy"},
		{"What if Chinese imports eliminate our indústria competitive advantage?", "Strategy", "manufacturing"},
		{"How would TikTok banning creators affect our mídia strategy?", "Marketing", "media"},
		{"If Airbnb captures 30% of hotel bookings, what happens in turismo?", "Sales", "tourism"},
		{"Generic question about market disruption", "Marketing", ""},
	}

	for _, tt := range tests {
		sk := Detect(tt.question, tt.department)
		if tt.wantSkill == "" {
			if sk != nil {
				t.Errorf("Detect(%q, %q): expected nil, got %q", tt.question, tt.department, sk.ID)
			}
		} else {
			if sk == nil {
				t.Errorf("Detect(%q, %q): expected %q, got nil", tt.question, tt.department, tt.wantSkill)
			} else if sk.ID != tt.wantSkill {
				t.Errorf("Detect(%q, %q): expected %q, got %q", tt.question, tt.department, tt.wantSkill, sk.ID)
			}
		}
	}
}

func TestHealthcareRules(t *testing.T) {
	sk := Registry["healthcare"]
	if sk == nil {
		t.Fatal("healthcare skill not registered")
	}
	if len(sk.Rules) < 12 {
		t.Errorf("expected at least 12 rules, got %d", len(sk.Rules))
	}
	for _, r := range sk.Rules {
		if r.ID == "" {
			t.Error("rule has empty ID")
		}
		if r.Description == "" {
			t.Errorf("rule %q has empty Description", r.ID)
		}
		if r.Stability < 0 || r.Stability > 1 {
			t.Errorf("rule %q stability out of range: %.2f", r.ID, r.Stability)
		}
	}
}

func TestFintechRules(t *testing.T) {
	sk := Registry["fintech"]
	if sk == nil {
		t.Fatal("fintech skill not registered")
	}
	if len(sk.Rules) < 12 {
		t.Errorf("expected at least 12 rules, got %d", len(sk.Rules))
	}
	for _, r := range sk.Rules {
		if r.ID == "" {
			t.Error("rule has empty ID")
		}
		if r.Stability < 0 || r.Stability > 1 {
			t.Errorf("rule %q stability out of range: %.2f", r.ID, r.Stability)
		}
	}
}

func TestRetailRules(t *testing.T) {
	sk := Registry["retail"]
	if sk == nil {
		t.Fatal("retail skill not registered")
	}
	if len(sk.Rules) < 12 {
		t.Errorf("expected at least 12 rules, got %d", len(sk.Rules))
	}
}

func TestLegalAndEducationRules(t *testing.T) {
	for _, id := range []string{"legal", "education"} {
		sk := Registry[id]
		if sk == nil {
			t.Fatalf("%s skill not registered", id)
		}
		if len(sk.Rules) < 10 {
			t.Errorf("%s: expected at least 10 rules, got %d", id, len(sk.Rules))
		}
	}
}

func TestNewSkillsRules(t *testing.T) {
	twelveRuleSkills := []string{"agro", "construction", "logistics", "saas", "energy", "manufacturing", "media"}
	for _, id := range twelveRuleSkills {
		sk := Registry[id]
		if sk == nil {
			t.Fatalf("skill %q not registered", id)
		}
		if len(sk.Rules) < 12 {
			t.Errorf("skill %q: expected at least 12 rules, got %d", id, len(sk.Rules))
		}
		for _, r := range sk.Rules {
			if r.ID == "" {
				t.Errorf("skill %q has rule with empty ID", id)
			}
			if r.Stability < 0 || r.Stability > 1 {
				t.Errorf("skill %q rule %q stability out of range: %.2f", id, r.ID, r.Stability)
			}
		}
	}

	// Tourism has 10 rules
	sk := Registry["tourism"]
	if sk == nil {
		t.Fatal("tourism skill not registered")
	}
	if len(sk.Rules) < 10 {
		t.Errorf("tourism: expected at least 10 rules, got %d", len(sk.Rules))
	}
}

func TestSkillAgents(t *testing.T) {
	for id, sk := range Registry {
		if id != sk.ID {
			// Skip industry slug aliases
			continue
		}
		for _, a := range sk.Agents {
			if a.Name == "" {
				t.Errorf("skill %q has agent with empty Name", id)
			}
			if a.Power <= 0 || a.Power > 1 {
				t.Errorf("skill %q agent %q has invalid Power: %.2f", id, a.Name, a.Power)
			}
		}
	}
}

func TestTotalSkills(t *testing.T) {
	uniqueIDs := map[string]bool{}
	for _, sk := range Registry {
		uniqueIDs[sk.ID] = true
	}
	if len(uniqueIDs) < 13 {
		t.Errorf("expected at least 13 unique skills, got %d", len(uniqueIDs))
	}
}

func TestSkillContextAndQueries(t *testing.T) {
	for id, sk := range Registry {
		if id != sk.ID {
			continue
		}
		if sk.Context == "" {
			t.Errorf("skill %q has empty Context", id)
		}
		if len(sk.Queries) == 0 {
			t.Errorf("skill %q has no Queries", id)
		}
	}
}
