package api

import (
	"encoding/json"
	"net/http"

	"github.com/fracture/fracture/db"
	"github.com/fracture/fracture/security"
	"github.com/go-chi/chi/v5"
)

// Handler holds all API dependencies.
type Handler struct {
	db          *db.DB
	signer      *security.Signer
	sanitizer   *security.Sanitizer
	auditLogger *security.AuditLogger
}

// NewHandler creates a new API Handler.
func NewHandler(
	database *db.DB,
	signer *security.Signer,
	sanitizer *security.Sanitizer,
	auditLogger *security.AuditLogger,
) *Handler {
	return &Handler{
		db:          database,
		signer:      signer,
		sanitizer:   sanitizer,
		auditLogger: auditLogger,
	}
}

// Routes returns the chi router with all API routes mounted.
func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()

	// Health check
	r.Get("/health", h.health)

	// Config / onboarding
	r.Get("/config", h.getConfig)
	r.Post("/config", h.setConfig)
	r.Get("/onboarding/status", h.onboardingStatus)
	r.Post("/onboarding/complete", h.completeOnboarding)

	// Company profile
	r.Get("/company", h.getCompany)
	r.Post("/company", h.upsertCompany)

	// LLM key validation
	r.Post("/keys/validate", h.validateKey)

	// Simulations
	r.Post("/simulations", h.createSimulation)
	r.Get("/simulations", h.listSimulations)
	r.Get("/simulations/{id}", h.getSimulation)
	r.Get("/simulations/{id}/stream", h.streamSimulation) // SSE
	r.Delete("/simulations/{id}", h.deleteSimulation)

	// Results
	r.Get("/simulations/{id}/results", h.getResults)

	// Feedback
	r.Post("/simulations/{id}/feedback", h.submitFeedback)

	// Templates
	r.Get("/templates", h.listTemplates)
	r.Get("/templates/{id}", h.getTemplate)

	// Archetypes
	r.Get("/archetypes", h.listArchetypes)
	r.Post("/archetypes", h.createArchetype)
	r.Put("/archetypes/{id}", h.updateArchetype)

	// Rules
	r.Get("/rules", h.listRules)
	r.Post("/rules", h.createRule)
	r.Put("/rules/{id}", h.updateRule)

	// Audit log
	r.Get("/audit", h.getAuditLog)

	return r
}

// ─── Health ──────────────────────────────────────────────────────────────────

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "version": "1.0.0"})
}

// ─── Config ──────────────────────────────────────────────────────────────────

func (h *Handler) getConfig(w http.ResponseWriter, r *http.Request) {
	keys := []string{
		"openai_key_set", "anthropic_key_set", "google_key_set", "ollama_enabled",
		"default_model_conformist", "default_model_disruptor",
		"default_model_synthesis", "default_rounds",
	}
	result := make(map[string]string)
	for _, k := range keys {
		val, _ := h.db.GetConfig(k)
		result[k] = val
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) setConfig(w http.ResponseWriter, r *http.Request) {
	var body map[string]string
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	for k, v := range body {
		if err := h.db.SetConfig(k, v); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to save config")
			return
		}
	}
	_ = h.auditLogger.Log("config.updated", "system", body)
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// ─── Onboarding ──────────────────────────────────────────────────────────────

func (h *Handler) onboardingStatus(w http.ResponseWriter, r *http.Request) {
	done, _ := h.db.IsOnboarded()
	writeJSON(w, http.StatusOK, map[string]bool{"complete": done})
}

func (h *Handler) completeOnboarding(w http.ResponseWriter, r *http.Request) {
	if err := h.db.SetConfig("onboarding_complete", "true"); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save onboarding status")
		return
	}
	_ = h.auditLogger.Log("onboarding.completed", "system", nil)
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// ─── Company ─────────────────────────────────────────────────────────────────

func (h *Handler) getCompany(w http.ResponseWriter, r *http.Request) {
	val, err := h.db.GetConfig("company_json")
	if err != nil || val == "" {
		writeJSON(w, http.StatusOK, nil)
		return
	}
	var company map[string]interface{}
	json.Unmarshal([]byte(val), &company)
	writeJSON(w, http.StatusOK, company)
}

func (h *Handler) upsertCompany(w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	b, _ := json.Marshal(body)
	if err := h.db.SetConfig("company_json", string(b)); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save company")
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// ─── Key Validation ──────────────────────────────────────────────────────────

func (h *Handler) validateKey(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Provider string `json:"provider"`
		Key      string `json:"key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	// Sanitize key before storing
	cleanKey, err := h.sanitizer.Sanitize(r.Context(), body.Key)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid key format")
		return
	}

	// Store key (encrypted in production; plaintext for MVP)
	configKey := body.Provider + "_api_key"
	if err := h.db.SetConfig(configKey, cleanKey); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save key")
		return
	}
	if err := h.db.SetConfig(body.Provider+"_key_set", "true"); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update key status")
		return
	}

	_ = h.auditLogger.Log("key.configured", body.Provider, map[string]string{"provider": body.Provider})
	writeJSON(w, http.StatusOK, map[string]bool{"valid": true})
}

// ─── Simulations (stubs — full implementation in simulation.go) ──────────────

func (h *Handler) createSimulation(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusAccepted, map[string]string{"status": "queued"})
}

func (h *Handler) listSimulations(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, []interface{}{})
}

func (h *Handler) getSimulation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	writeJSON(w, http.StatusOK, map[string]string{"id": id})
}

func (h *Handler) streamSimulation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) deleteSimulation(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *Handler) getResults(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{})
}

func (h *Handler) submitFeedback(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *Handler) listTemplates(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, []interface{}{})
}

func (h *Handler) getTemplate(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{})
}

func (h *Handler) listArchetypes(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, []interface{}{})
}

func (h *Handler) createArchetype(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusCreated, map[string]bool{"ok": true})
}

func (h *Handler) updateArchetype(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *Handler) listRules(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, []interface{}{})
}

func (h *Handler) createRule(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusCreated, map[string]bool{"ok": true})
}

func (h *Handler) updateRule(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *Handler) getAuditLog(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, []interface{}{})
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
