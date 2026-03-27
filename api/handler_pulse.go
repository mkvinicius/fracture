package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/fracture/fracture/llm"
)

// ─── Quick Pulse ─────────────────────────────────────────────────────────────

func (h *Handler) quickPulse(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Situation string `json:"situation"`
		Domain    string `json:"domain"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if body.Situation == "" {
		writeError(w, http.StatusBadRequest, "situation is required")
		return
	}

	cleanSit, err := h.sanitizer.Sanitize(r.Context(), body.Situation)
	if err != nil {
		writeError(w, http.StatusBadRequest, "situation contains invalid content")
		return
	}

	router, err := h.buildLLMRouter()
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "no LLM keys configured")
		return
	}

	// Use conformist role for pulse — it falls back to the available provider
	caller := router.ForRole(llm.RoleConformist)
	systemPrompt := `You are a rapid market tension analyst. Given a business situation, output a JSON object with:
- score: integer 0-100 (0=no tension, 100=maximum disruption risk)
- level: "low" | "medium" | "high" | "critical"
- summary: one sentence explaining the tension
- top_risks: array of 3 strings, each a specific risk
Respond with JSON only.`

	userPrompt := fmt.Sprintf("Domain: %s\nSituation: %s", body.Domain, cleanSit)

	raw, _, err := caller.Call(r.Context(), systemPrompt, userPrompt, 400)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "LLM call failed: "+err.Error())
		return
	}

	var pulse map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &pulse); err != nil {
		pulse = map[string]interface{}{
			"score":     50,
			"level":     "medium",
			"summary":   raw,
			"top_risks": []string{},
		}
	}
	writeJSON(w, http.StatusOK, pulse)
}
