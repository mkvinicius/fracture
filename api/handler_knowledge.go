package api

import (
	"encoding/json"
	"net/http"

	"github.com/fracture/fracture/db"
	"github.com/fracture/fracture/engine"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// ─── Templates ───────────────────────────────────────────────────────────────

func (h *Handler) listTemplates(w http.ResponseWriter, r *http.Request) {
	templates := []map[string]interface{}{
		{"id": "competitor-free-tier", "name": "Competitor launches free tier", "domain": "market", "rounds": 20,
			"question": "What happens if a major competitor launches a free tier targeting our core customers?"},
		{"id": "ai-disruption", "name": "AI disrupts our core product", "domain": "technology", "rounds": 20,
			"question": "How would AI automation affect our product category in the next 18 months?"},
		{"id": "regulation-change", "name": "New regulation in our sector", "domain": "regulation", "rounds": 15,
			"question": "What if new data privacy regulation requires us to change our business model?"},
		{"id": "price-increase", "name": "We raise prices by 30%", "domain": "market", "rounds": 10,
			"question": "What is the market reaction if we raise prices by 30% next quarter?"},
		{"id": "talent-war", "name": "Talent war in our sector", "domain": "behavior", "rounds": 15,
			"question": "How will the talent shortage in our sector evolve and what rules will change?"},
		{"id": "new-entrant", "name": "Well-funded new entrant", "domain": "market", "rounds": 20,
			"question": "A well-funded startup enters our market with a radically different approach. What happens?"},
	}
	writeJSON(w, http.StatusOK, templates)
}

func (h *Handler) getTemplate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	templates := map[string]map[string]interface{}{
		"competitor-free-tier": {"id": "competitor-free-tier", "name": "Competitor launches free tier", "domain": "market", "rounds": 20,
			"question": "What happens if a major competitor launches a free tier targeting our core customers?"},
		"ai-disruption": {"id": "ai-disruption", "name": "AI disrupts our core product", "domain": "technology", "rounds": 20,
			"question": "How would AI automation affect our product category in the next 18 months?"},
	}
	t, ok := templates[id]
	if !ok {
		writeError(w, http.StatusNotFound, "template not found")
		return
	}
	writeJSON(w, http.StatusOK, t)
}

// ─── Archetypes ──────────────────────────────────────────────────────────────

// builtinArchetypes returns the hardcoded built-in archetype list as ArchetypeRow slices.
// These are never stored in the DB; they are merged with custom archetypes at query time.
func builtinArchetypes() []db.ArchetypeRow {
	type meta struct {
		id, name, agentType, description string
		weight                           float64
	}
	list := []meta{
		{"pragmatist", "The Pragmatist", "conformist", "Mid-level manager: data-driven, risk-averse, process-oriented", 0.7},
		{"loyalist", "The Loyalist", "conformist", "Long-term customer: brand-loyal, resistant to change, word-of-mouth", 0.6},
		{"analyst", "The Analyst", "conformist", "Industry analyst: evidence-based, conservative, benchmark-focused", 0.8},
		{"opportunist", "The Opportunist", "conformist", "Competitor executive: market-watching, fast-follower, profit-driven", 0.75},
		{"traditionalist", "The Traditionalist", "conformist", "Regulator / policy maker: rule-enforcing, slow-moving, stability-focused", 0.65},
		{"regulator", "The Regulator", "conformist", "Compliance officer: risk-averse, rule-based, conservative", 0.7},
		{"consumer", "The Consumer", "conformist", "End user / customer: value-seeking, convenience-driven, price-sensitive", 0.55},
		{"investor", "The Investor", "conformist", "Institutional investor: ROI-focused, long-term, risk-calibrated", 0.85},
		{"visionary", "The Visionary", "disruptor", "Startup founder: contrarian, first-principles, high-risk tolerance", 0.9},
		{"rebel", "The Rebel", "disruptor", "Activist / whistleblower: anti-establishment, viral, unpredictable", 0.7},
		{"tech-accelerator", "The Tech Accelerator", "disruptor", "AI/tech researcher: exponential thinking, automation-first, impatient", 0.85},
		{"arbitrageur", "The Arbitrageur", "disruptor", "Financial disruptor: gap-finder, speed-focused, asymmetric bets", 0.8},
	}
	out := make([]db.ArchetypeRow, 0, len(list))
	for _, m := range list {
		out = append(out, db.ArchetypeRow{
			ID: m.id, Name: m.name, AgentType: m.agentType,
			Description: m.description, MemoryWeight: m.weight, IsActive: true,
		})
	}
	return out
}

func (h *Handler) listArchetypes(w http.ResponseWriter, r *http.Request) {
	companyID := r.URL.Query().Get("company_id")

	// Start with built-ins
	result := builtinArchetypes()

	// Merge custom archetypes from DB (company-specific or all)
	custom, err := h.db.ListArchetypes(companyID)
	if err == nil {
		// Index built-ins by ID so company overrides replace them
		index := make(map[string]int, len(result))
		for i, a := range result {
			index[a.ID] = i
		}
		for _, a := range custom {
			if i, ok := index[a.ID]; ok {
				// Override built-in with company-calibrated version
				result[i] = a
			} else {
				// Append new custom archetype
				result = append(result, a)
			}
		}
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) getArchetype(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	// Check DB first (custom archetypes)
	a, err := h.db.GetArchetype(id)
	if err == nil {
		writeJSON(w, http.StatusOK, a)
		return
	}
	// Fall back to built-in
	for _, b := range builtinArchetypes() {
		if b.ID == id {
			writeJSON(w, http.StatusOK, b)
			return
		}
	}
	writeError(w, http.StatusNotFound, "archetype not found")
}

func (h *Handler) deleteArchetype(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	// Built-in archetypes cannot be deleted
	for _, b := range builtinArchetypes() {
		if b.ID == id {
			writeError(w, http.StatusForbidden, "built-in archetypes cannot be deleted")
			return
		}
	}
	if err := h.db.DeleteArchetype(id); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete archetype: "+err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) createArchetype(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name         string  `json:"name"`
		AgentType    string  `json:"agent_type"` // conformist | disruptor
		Description  string  `json:"description"`
		MemoryWeight float64 `json:"memory_weight"`
		CompanyID    string  `json:"company_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" || req.AgentType == "" {
		writeError(w, http.StatusBadRequest, "name and agent_type are required")
		return
	}
	if req.AgentType != "conformist" && req.AgentType != "disruptor" {
		writeError(w, http.StatusBadRequest, "agent_type must be conformist or disruptor")
		return
	}
	if req.MemoryWeight == 0 {
		req.MemoryWeight = 1.0
	}
	row := &db.ArchetypeRow{
		ID:           uuid.New().String(),
		CompanyID:    req.CompanyID,
		Name:         req.Name,
		AgentType:    req.AgentType,
		Description:  req.Description,
		MemoryWeight: req.MemoryWeight,
		IsActive:     true,
	}
	if err := h.db.CreateArchetype(row); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create archetype: "+err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, row)
}

func (h *Handler) updateArchetype(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req struct {
		Name         string  `json:"name"`
		Description  string  `json:"description"`
		MemoryWeight float64 `json:"memory_weight"`
		IsActive     bool    `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	if req.MemoryWeight == 0 {
		req.MemoryWeight = 1.0
	}
	if err := h.db.UpdateArchetype(id, req.Name, req.Description, req.MemoryWeight, req.IsActive); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update archetype: "+err.Error())
		return
	}
	updated, err := h.db.GetArchetype(id)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

// ─── Rules ───────────────────────────────────────────────────────────────────

// mergeRulesWithCustom takes a slice of built-in engine.Rule pointers and appends
// active custom rules from the DB for the given companyID, converting them to the
// same engine.Rule shape so callers receive a unified list.
func (h *Handler) mergeRulesWithCustom(builtins []*engine.Rule, domain, companyID string) []interface{} {
	type ruleView struct {
		ID          string  `json:"id"`
		Description string  `json:"description"`
		Domain      string  `json:"domain"`
		Stability   float64 `json:"stability"`
		IsCustom    bool    `json:"is_custom"`
		CompanyID   string  `json:"company_id,omitempty"`
	}
	result := make([]interface{}, 0, len(builtins)+8)
	for _, r := range builtins {
		result = append(result, ruleView{
			ID: r.ID, Description: r.Description,
			Domain: string(r.Domain), Stability: r.Stability,
			IsCustom: false,
		})
	}
	if companyID != "" {
		custom, err := h.db.ListCustomRules(companyID)
		if err == nil {
			for _, cr := range custom {
				if !cr.IsActive {
					continue
				}
				if domain != "" && cr.Domain != domain {
					continue
				}
				result = append(result, ruleView{
					ID: cr.ID, Description: cr.Description,
					Domain: cr.Domain, Stability: cr.Stability,
					IsCustom: true, CompanyID: cr.CompanyID,
				})
			}
		}
	}
	return result
}

func (h *Handler) listRules(w http.ResponseWriter, r *http.Request) {
	companyID := r.URL.Query().Get("company_id")
	world := engine.DefaultWorldForDomain("market", "", "")
	builtins := make([]*engine.Rule, 0, len(world.Rules))
	for _, rule := range world.Rules {
		builtins = append(builtins, rule)
	}
	writeJSON(w, http.StatusOK, h.mergeRulesWithCustom(builtins, "", companyID))
}

func (h *Handler) listRulesByDomain(w http.ResponseWriter, r *http.Request) {
	domain := chi.URLParam(r, "domain")
	companyID := r.URL.Query().Get("company_id")
	world := engine.DefaultWorldForDomain(engine.RuleDomain(domain), "", "")
	builtins := make([]*engine.Rule, 0, len(world.Rules))
	for _, rule := range world.Rules {
		builtins = append(builtins, rule)
	}
	writeJSON(w, http.StatusOK, h.mergeRulesWithCustom(builtins, domain, companyID))
}

func (h *Handler) getCustomRule(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	rule, err := h.db.GetCustomRule(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "rule not found")
		return
	}
	writeJSON(w, http.StatusOK, rule)
}

func (h *Handler) deleteCustomRule(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.db.DeleteCustomRule(id); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete rule: "+err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) createRule(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Description string  `json:"description"`
		Domain      string  `json:"domain"`
		Stability   float64 `json:"stability"`
		CompanyID   string  `json:"company_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Description == "" {
		writeError(w, http.StatusBadRequest, "description is required")
		return
	}
	if req.Domain == "" {
		req.Domain = "market"
	}
	if req.Stability == 0 {
		req.Stability = 0.5
	}
	row := &db.CustomRuleRow{
		ID:          uuid.New().String(),
		CompanyID:   req.CompanyID,
		Description: req.Description,
		Domain:      req.Domain,
		Stability:   req.Stability,
		IsActive:    true,
	}
	if err := h.db.CreateCustomRule(row); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create rule: "+err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, row)
}

func (h *Handler) updateRule(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req struct {
		Description string  `json:"description"`
		Domain      string  `json:"domain"`
		Stability   float64 `json:"stability"`
		IsActive    bool    `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Description == "" {
		writeError(w, http.StatusBadRequest, "description is required")
		return
	}
	if req.Domain == "" {
		req.Domain = "market"
	}
	if req.Stability == 0 {
		req.Stability = 0.5
	}
	if err := h.db.UpdateCustomRule(id, req.Description, req.Domain, req.Stability, req.IsActive); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update rule: "+err.Error())
		return
	}
	updated, err := h.db.GetCustomRule(id)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
		return
	}
	writeJSON(w, http.StatusOK, updated)
}
