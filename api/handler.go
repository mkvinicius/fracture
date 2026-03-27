package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/fracture/fracture/deepsearch"
	"github.com/fracture/fracture/db"
	"github.com/fracture/fracture/engine"
	"github.com/fracture/fracture/llm"
	"github.com/fracture/fracture/memory"
	"github.com/fracture/fracture/security"
	"github.com/fracture/fracture/telemetry"
	"github.com/go-chi/chi/v5"
)

// Handler holds all API dependencies.
type Handler struct {
	db          *db.DB
	signer      *security.Signer
	sanitizer   *security.Sanitizer
	auditLogger *security.AuditLogger
	tel         *telemetry.Client
	calibrator  *memory.Calibrator
	ragStore    *memory.RAGStore

	// simulation state: in-memory mirror of simulation_jobs table.
	// All mutations are written through to the DB for resilience across restarts.
	simMu   sync.RWMutex
	simJobs map[string]*simJob
}

type simJob struct {
	ID              string `json:"id"`
	Status          string `json:"status"` // queued | researching | running | done | error
	Question        string `json:"question"`
	Department      string `json:"department"`
	Rounds          int    `json:"rounds"`
	Mode            string `json:"mode,omitempty"` // standard | premium
	CreatedAt       int64  `json:"created_at"`
	DurationMs      int64  `json:"duration_ms,omitempty"`
	Error           string `json:"error,omitempty"`
	ResearchSources int    `json:"research_sources,omitempty"` // web sources found by DeepSearch
	ResearchTokens  int    `json:"research_tokens,omitempty"`  // tokens used by DeepSearch
	Company         string `json:"company,omitempty"`
	// Live progress fields — updated after each round, streamed via SSE
	CurrentRound    int     `json:"current_round,omitempty"`     // last completed round number
	CurrentTension  float64 `json:"current_tension,omitempty"`   // tension level after last round
	FractureCount   int     `json:"fracture_count,omitempty"`    // fracture points triggered so far
	LastAgentName   string  `json:"last_agent_name,omitempty"`   // name of last agent to act
	LastAgentAction string  `json:"last_agent_action,omitempty"` // truncated text of last action
	TotalTokens     int     `json:"total_tokens,omitempty"`      // cumulative tokens used
}

// NewHandler creates a new API Handler.
// It marks any interrupted jobs as failed (resilience across restarts) and
// re-hydrates the in-memory map from the DB so the UI sees correct state.
func NewHandler(
	database *db.DB,
	signer *security.Signer,
	sanitizer *security.Sanitizer,
	auditLogger *security.AuditLogger,
	tel *telemetry.Client,
) *Handler {
	h := &Handler{
		db:          database,
		signer:      signer,
		sanitizer:   sanitizer,
		auditLogger: auditLogger,
		tel:         tel,
		calibrator:  memory.NewCalibrator(database.DB),
		ragStore:    memory.NewRAGStore(database.DB),
		simJobs:     make(map[string]*simJob),
	}

	// Mark jobs that were in-flight when the process last stopped as failed.
	if n, err := database.MarkInterruptedJobsFailed(); err == nil && n > 0 {
		log.Printf("[FRACTURE] marked %d interrupted job(s) as failed (process restart)", n)
	}

	// Re-hydrate in-memory map from DB so the UI sees correct state immediately.
	if jobs, err := database.ListJobs(); err == nil {
		for _, j := range jobs {
			j := j // capture
			h.simJobs[j.ID] = &simJob{
				ID:              j.ID,
				Status:          j.Status,
				Question:        j.Question,
				Department:      j.Department,
				Rounds:          j.Rounds,
				CreatedAt:       j.CreatedAt,
				DurationMs:      j.DurationMs,
				Error:           j.Error,
				ResearchSources: j.ResearchSources,
				ResearchTokens:  j.ResearchTokens,
				Company:         j.Company,
				// Restore live progress from DB so SSE is accurate after restart
				CurrentRound:    j.CurrentRound,
				CurrentTension:  j.CurrentTension,
				FractureCount:   j.FractureCount,
				LastAgentName:   j.LastAgentName,
				LastAgentAction: j.LastAgentAction,
				TotalTokens:     j.TotalTokens,
			}
		}
		log.Printf("[FRACTURE] re-hydrated %d job(s) from DB", len(jobs))
	}

	return h
}

// persistJob writes the current in-memory job state to the DB.
// Must be called with simMu held (or after releasing it for the read).
func (h *Handler) persistJob(j *simJob) {
	_ = h.db.UpsertJob(&db.JobRow{
		ID:              j.ID,
		Status:          j.Status,
		Question:        j.Question,
		Department:      j.Department,
		Rounds:          j.Rounds,
		Mode:            j.Mode,
		Company:         j.Company,
		Error:           j.Error,
		ResearchSources: j.ResearchSources,
		ResearchTokens:  j.ResearchTokens,
		DurationMs:      j.DurationMs,
		CreatedAt:       j.CreatedAt,
		// Live progress fields — persisted so they survive a restart
		CurrentRound:    j.CurrentRound,
		CurrentTension:  j.CurrentTension,
		FractureCount:   j.FractureCount,
		LastAgentName:   j.LastAgentName,
		LastAgentAction: j.LastAgentAction,
		TotalTokens:     j.TotalTokens,
	})
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

	// Simulations — full implementation
	r.Post("/simulations", h.createSimulation)
	r.Get("/simulations", h.listSimulations)
	r.Get("/simulations/compare", h.compareSimulations) // must be before {id}
	r.Get("/simulations/{id}", h.getSimulation)
	r.Get("/simulations/{id}/stream", h.streamSimulation) // SSE
	r.Delete("/simulations/{id}", h.deleteSimulation)

	// Results, export & feedback
	r.Get("/simulations/{id}/results", h.getResults)
	r.Get("/simulations/{id}/report", h.getReport)
	r.Get("/simulations/{id}/export/markdown", h.exportMarkdown)
	r.Get("/simulations/{id}/export/json", h.exportJSON)
	r.Get("/simulations/{id}/events", h.getSimulationEvents)
	r.Post("/simulations/{id}/feedback", h.submitFeedback)

	// Quick pulse (fast tension check, no full simulation)
	r.Post("/pulse", h.quickPulse)

	// Templates
	r.Get("/templates", h.listTemplates)
	r.Get("/templates/{id}", h.getTemplate)

	// Archetypes — built-ins + custom from DB
	r.Get("/archetypes", h.listArchetypes)
	r.Post("/archetypes", h.createArchetype)
	r.Get("/archetypes/{id}", h.getArchetype)
	r.Put("/archetypes/{id}", h.updateArchetype)
	r.Delete("/archetypes/{id}", h.deleteArchetype)

	// Rules — built-ins merged with custom from DB
	r.Get("/rules", h.listRules)
	r.Get("/rules/domain/{domain}", h.listRulesByDomain)
	r.Post("/rules", h.createRule)
	r.Get("/rules/{id}", h.getCustomRule)
	r.Put("/rules/{id}", h.updateRule)
	r.Delete("/rules/{id}", h.deleteCustomRule)

	// Audit log
	r.Get("/audit", h.getAuditLog)

	// Telemetry opt-in/opt-out
	r.Get("/telemetry", h.getTelemetry)
	r.Post("/telemetry", h.setTelemetry)

	// Update check
	r.Get("/update-check", h.checkForUpdate)

	// Context extraction from URLs
	r.Post("/extract-context", h.extractContext)

	return r
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

// companyID retrieves the current company name from config (used as company identifier).
func (h *Handler) companyID() string {
	companyJSON, _ := h.db.GetConfig("company_json")
	if companyJSON == "" {
		return ""
	}
	var cp map[string]interface{}
	if json.Unmarshal([]byte(companyJSON), &cp) != nil {
		return ""
	}
	name, _ := cp["name"].(string)
	return name
}

// buildLLMRouter constructs an LLM router from API keys stored in the DB.
func (h *Handler) buildLLMRouter() (*llm.Router, error) {
	cfg := llm.DefaultRouterConfig()

	openaiKey, _ := h.db.GetConfig("openai_api_key")
	anthropicKey, _ := h.db.GetConfig("anthropic_api_key")
	googleKey, _ := h.db.GetConfig("google_api_key")
	ollamaEnabled, _ := h.db.GetConfig("ollama_enabled")

	hasAny := false

	if openaiKey != "" {
		cfg[llm.RoleConformist] = llm.ModelConfig{Provider: "openai", Model: "gpt-4o-mini", APIKey: openaiKey}
		cfg[llm.RoleDisruptor] = llm.ModelConfig{Provider: "openai", Model: "gpt-4o", APIKey: openaiKey}
		hasAny = true
	}
	if anthropicKey != "" {
		cfg[llm.RoleSynthesis] = llm.ModelConfig{Provider: "anthropic", Model: "claude-sonnet-4-20250514", APIKey: anthropicKey}
		cfg[llm.RoleSanitizer] = llm.ModelConfig{Provider: "anthropic", Model: "claude-haiku-4-5-20251001", APIKey: anthropicKey}
		if !hasAny {
			cfg[llm.RoleConformist] = llm.ModelConfig{Provider: "anthropic", Model: "claude-haiku-4-5-20251001", APIKey: anthropicKey}
			cfg[llm.RoleDisruptor] = llm.ModelConfig{Provider: "anthropic", Model: "claude-sonnet-4-20250514", APIKey: anthropicKey}
		}
		hasAny = true
	}
	if googleKey != "" {
		cfg[llm.RoleCoherence] = llm.ModelConfig{Provider: "google", Model: "gemini-1.5-flash", APIKey: googleKey}
		if !hasAny {
			cfg[llm.RoleConformist] = llm.ModelConfig{Provider: "google", Model: "gemini-1.5-flash", APIKey: googleKey}
			cfg[llm.RoleDisruptor] = llm.ModelConfig{Provider: "google", Model: "gemini-1.5-pro", APIKey: googleKey}
		}
		hasAny = true
	}
	if ollamaEnabled == "true" {
		ollamaModel, _ := h.db.GetConfig("ollama_model")
		if ollamaModel == "" {
			ollamaModel = "llama3"
		}
		cfg[llm.RoleConformist] = llm.ModelConfig{Provider: "ollama", Model: ollamaModel}
		cfg[llm.RoleDisruptor] = llm.ModelConfig{Provider: "ollama", Model: ollamaModel}
		cfg[llm.RoleSynthesis] = llm.ModelConfig{Provider: "ollama", Model: ollamaModel}
		cfg[llm.RoleCoherence] = llm.ModelConfig{Provider: "ollama", Model: ollamaModel}
		hasAny = true
	}

	if !hasAny {
		return nil, fmt.Errorf("no LLM keys configured")
	}

	// Fill any missing roles with the first available
	for _, role := range []llm.ModelRole{llm.RoleConformist, llm.RoleDisruptor, llm.RoleSynthesis, llm.RoleCoherence, llm.RoleSanitizer} {
		if _, ok := cfg[role]; !ok {
			// Use conformist as fallback
			if c, ok2 := cfg[llm.RoleConformist]; ok2 {
				cfg[role] = c
			}
		}
	}

	return llm.NewRouter(cfg), nil
}

// domainResultsToSignals converts a DomainResearchResult map to a []memory.DomainSignal
// slice, breaking the deepsearch→memory circular import by doing the conversion here.
func domainResultsToSignals(results map[engine.RuleDomain]*deepsearch.DomainResearchResult) []memory.DomainSignal {
	signals := make([]memory.DomainSignal, 0, len(results))
	for domain, res := range results {
		if res == nil {
			continue
		}
		signals = append(signals, memory.DomainSignal{
			Domain:    string(domain),
			Summary:   res.Summary,
			Signals:   res.KeySignals,
			Sentiment: res.SentimentScore,
		})
	}
	return signals
}

// loadFullReport fetches result_json for a simulation and unmarshals it as FullReport.
func (h *Handler) loadFullReport(id string) (*engine.FullReport, error) {
	sim, err := h.db.GetSimulation(id)
	if err != nil {
		return nil, fmt.Errorf("simulation not found")
	}
	var report engine.FullReport
	if err := json.Unmarshal([]byte(sim.ResultJSON), &report); err != nil {
		return nil, fmt.Errorf("failed to parse report")
	}
	if report.SimulationID == "" {
		return nil, fmt.Errorf("report not yet generated for this simulation")
	}
	return &report, nil
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
