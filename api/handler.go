package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/fracture/fracture/archetypes"
	"github.com/fracture/fracture/contextextractor"
	"github.com/fracture/fracture/deepsearch"
	"github.com/fracture/fracture/db"
	"github.com/fracture/fracture/engine"
	"github.com/fracture/fracture/llm"
	"github.com/fracture/fracture/memory"
	"github.com/fracture/fracture/security"
	"github.com/fracture/fracture/telemetry"
	"github.com/fracture/fracture/updater"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
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
	CompanySize     string `json:"company_size,omitempty"`     // pme | media | enterprise
	CompanySector   string `json:"company_sector,omitempty"`
	CompanyLocation string `json:"company_location,omitempty"`
	// Live progress fields — updated after each round, streamed via SSE
	CurrentRound    int     `json:"current_round,omitempty"`     // last completed round number
	CurrentTension  float64 `json:"current_tension,omitempty"`   // tension level after last round
	FractureCount   int     `json:"fracture_count,omitempty"`    // fracture points triggered so far
	LastAgentName   string  `json:"last_agent_name,omitempty"`   // name of last agent to act
	LastAgentAction string  `json:"last_agent_action,omitempty"` // truncated text of last action
	TotalTokens     int     `json:"total_tokens,omitempty"`      // cumulative tokens used
	// God View: channel for injecting live events into a running simulation
	injectCh        chan string `json:"-"`
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
	r.Get("/keys", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, []struct{}{})
	})

	// Notifications (stub)
	r.Get("/notifications", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, []struct{}{})
	})

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

	// Share link (public, no auth)
	r.Post("/simulations/{id}/share", h.createShareLink)
	r.Get("/share/{token}", h.getSharedReport)

	// God View — inject event into running simulation
	r.Post("/simulations/{id}/inject", h.injectEvent)

	// Prediction outcome tracking (accuracy)
	r.Get("/simulations/{id}/outcomes", h.getOutcomes)
	r.Post("/simulations/{id}/outcomes", h.saveOutcome)
	r.Get("/accuracy", h.getAccuracy)

	// Scheduled simulations
	r.Get("/schedules", h.listSchedules)
	r.Post("/schedules", h.createSchedule)
	r.Delete("/schedules/{id}", h.deleteSchedule)
	r.Patch("/schedules/{id}", h.toggleSchedule)

	// Public API keys management
	r.Get("/api-keys", h.listAPIKeys)
	r.Post("/api-keys", h.createAPIKey)
	r.Delete("/api-keys/{id}", h.deleteAPIKey)

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

// ─── Health ──────────────────────────────────────────────────────────────────

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "version": updater.CurrentVersion})
}

// ─── Update Check ─────────────────────────────────────────────────────────────

func (h *Handler) checkForUpdate(w http.ResponseWriter, r *http.Request) {
	result, err := updater.CheckForUpdate()
	if err != nil {
		// Don't fail — just return current version with no update
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"has_update":      false,
			"current_version": updater.CurrentVersion,
			"error":           err.Error(),
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"has_update":      result.HasUpdate,
		"current_version": result.CurrentVersion,
		"latest_version":  result.LatestVersion,
		"release_url":     result.ReleaseURL,
		"release_name":    result.ReleaseName,
		"release_notes":   result.ReleaseNotes,
	})
}

// ─── Context Extraction ───────────────────────────────────────────────────────

func (h *Handler) extractContext(w http.ResponseWriter, r *http.Request) {
	var body struct {
		URLs []string `json:"urls"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if len(body.URLs) == 0 {
		writeError(w, http.StatusBadRequest, "urls array is required")
		return
	}
	if len(body.URLs) > 10 {
		body.URLs = body.URLs[:10] // max 10 URLs
	}

	ctx := contextextractor.ExtractFromURLs(body.URLs)

	type sourceResult struct {
		URL        string `json:"url"`
		SourceType string `json:"source_type"`
		Title      string `json:"title"`
		Description string `json:"description"`
		Content    string `json:"content"`
		Error      string `json:"error,omitempty"`
	}

	var sources []sourceResult
	for _, s := range ctx.Sources {
		sources = append(sources, sourceResult{
			URL:         s.URL,
			SourceType:  string(s.SourceType),
			Title:       s.Title,
			Description: s.Description,
			Content:     s.Content,
			Error:       s.Error,
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"sources": sources,
		"summary": ctx.Summary,
		"has_errors": ctx.HasErrors,
	})
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

	cleanKey, err := h.sanitizer.Sanitize(r.Context(), body.Key)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid key format")
		return
	}

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

// ─── Simulations ─────────────────────────────────────────────────────────────

func (h *Handler) createSimulation(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Question        string   `json:"question"`
		Department      string   `json:"department"`
		Rounds          int      `json:"rounds"`
		Mode            string   `json:"mode"`
		Context         string   `json:"context"`
		URLs            []string `json:"urls"`             // optional: company website + social media URLs
		CompanySize     string   `json:"company_size"`     // pme | media | enterprise
		CompanySector   string   `json:"company_sector"`
		CompanyLocation string   `json:"company_location"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if body.Question == "" {
		writeError(w, http.StatusBadRequest, "question is required")
		return
	}
	if body.Department == "" {
		body.Department = "market"
	}
	// Determine mode and derive MaxRounds from it; caller-supplied rounds are ignored
	// when a mode is provided — the mode is the source of truth.
	simMode := engine.SimulationMode(body.Mode)
	if simMode != engine.ModeStandard && simMode != engine.ModePremium {
		simMode = engine.ModeStandard
	}
	modeCfg := engine.DefaultConfigForMode(simMode)
	if body.Rounds <= 0 {
		body.Rounds = modeCfg.MaxRounds
	}

	// Sanitize inputs
	cleanQ, err := h.sanitizer.Sanitize(r.Context(), body.Question)
	if err != nil {
		writeError(w, http.StatusBadRequest, "question contains invalid content")
		return
	}
	cleanCtx, _ := h.sanitizer.Sanitize(r.Context(), body.Context)

	// If URLs provided, extract context automatically and prepend to context
	if len(body.URLs) > 0 && len(body.URLs) <= 10 {
		extracted := contextextractor.ExtractFromURLs(body.URLs)
		if extracted.Summary != "" {
			if cleanCtx != "" {
				cleanCtx = extracted.Summary + "\n\n" + cleanCtx
			} else {
				cleanCtx = extracted.Summary
			}
		}
	}

	job := &simJob{
		ID:         uuid.New().String(),
		Status:     "queued",
		Question:   cleanQ,
		Department: body.Department,
		Rounds:     body.Rounds,
		Mode:       string(simMode),
		CreatedAt:  time.Now().Unix(),
	}

	// Extract company name from saved profile (best-effort)
	companyName := ""
	if companyJSON, _ := h.db.GetConfig("company_json"); companyJSON != "" {
		var cp map[string]interface{}
		if json.Unmarshal([]byte(companyJSON), &cp) == nil {
			if name, ok := cp["name"].(string); ok {
				companyName = name
			}
		}
	}
	job.Company = companyName
	job.CompanySize = body.CompanySize
	job.CompanySector = body.CompanySector
	job.CompanyLocation = body.CompanyLocation
	job.injectCh = make(chan string, 8)

	h.simMu.Lock()
	h.simJobs[job.ID] = job
	h.persistJob(job) // persist initial state immediately
	h.simMu.Unlock()

	// Run DeepSearch + simulation asynchronously
	go h.runWithDeepSearch(job, cleanCtx)

	_ = h.auditLogger.Log("simulation.created", job.ID, map[string]interface{}{
		"question": cleanQ, "department": body.Department, "rounds": body.Rounds,
	})

	writeJSON(w, http.StatusAccepted, map[string]string{"id": job.ID, "status": "queued"})
}

// runWithDeepSearch runs the DeepSearch agent first, then launches the simulation.
// This enriches the simulation with real-world context before the 32 agents start.
func (h *Handler) runWithDeepSearch(job *simJob, manualContext string) {
	// Step 1: Mark as researching
	h.simMu.Lock()
	job.Status = "researching"
	h.persistJob(job)
	h.simMu.Unlock()

	// Build LLM router first (needed for DeepSearch too)
	router, err := h.buildLLMRouter()
	if err != nil {
		h.simMu.Lock()
		job.Status = "error"
		job.Error = "no LLM keys configured — add at least one API key in Settings"
		h.persistJob(job)
		h.simMu.Unlock()
		return
	}

	// Step 2: Run DeepSearch to gather real-world context
	researchCtx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	dsAgent := deepsearch.New(
		router.ForRole(llm.RoleSynthesis), // use synthesis model for research
		deepsearch.DefaultConfig(),
	)

	contextReport, dsErr := dsAgent.Research(researchCtx, job.Question, job.Company, job.Department)

	// Build enriched context: DeepSearch findings + manual context
	enrichedContext := manualContext
	if dsErr == nil && contextReport != nil {
		researchContext := contextReport.ToSimulationContext()
		if enrichedContext != "" {
			enrichedContext = researchContext + "\n\n" + enrichedContext
		} else {
			enrichedContext = researchContext
		}
		h.simMu.Lock()
		job.ResearchSources = len(contextReport.Sources)
		job.ResearchTokens = contextReport.TokensUsed
		h.persistJob(job)
		h.simMu.Unlock()
		log.Printf("[FRACTURE] DeepSearch completed for sim %s: %d sources, %d tokens",
			job.ID, len(contextReport.Sources), contextReport.TokensUsed)

		// persiste causalidades no grafo — fire and forget, não bloqueia
		if h.calibrator != nil {
			synthLLM := router.ForRole(llm.RoleSynthesis)
			cr := contextReport
			sector := job.Department
			go func() {
				triples := deepsearch.ExtractCausalities(context.Background(), synthLLM, cr, sector)
				deepsearch.PersistFromDeepSearch(h.calibrator.Graph, sector, triples)
			}()
		}

		// Synthesize domain contexts and persist them
		domainContexts := dsAgent.SynthesizeDomainContext(contextReport)
		for domain, ctx := range domainContexts {
			// Calculate stability modifier based on confidence
			stabilityMod := -0.15 * ctx.Confidence
			if stabilityMod < -0.95 {
				stabilityMod = -0.95
			}
			if err := h.db.SaveDomainContext(job.ID, domain, db.DomainContextRow{
				SimulationID:      job.ID,
				Domain:            domain,
				Context:           ctx.ContextText,
				AffectedRules:     fmt.Sprintf("%v", ctx.AffectedRules),
				Signals:           "[]",
				StabilityModifier: stabilityMod,
				Confidence:        ctx.Confidence,
			}); err != nil {
				log.Printf("[FRACTURE] Failed to save domain context for %s:%s: %v", job.ID, domain, err)
			}
		}
	} else if dsErr != nil {
		log.Printf("[FRACTURE] DeepSearch failed for sim %s: %v — continuing without research context", job.ID, dsErr)
	}

	// Step 3: Run domain-specific research for all 7 domains concurrently
	dr := deepsearch.NewDomainResearcher(dsAgent, h.db.DB)
	domainResults, drErr := dr.ResearchAllDomains(researchCtx, job.Question, job.Company, job.Department)
	if drErr != nil {
		log.Printf("[FRACTURE] DomainResearcher partial error for sim %s: %v", job.ID, drErr)
	}
	if domainResults == nil {
		domainResults = make(map[engine.RuleDomain]*deepsearch.DomainResearchResult)
	}

	// Persist each domain context to the DB
	for domain, res := range domainResults {
		if res == nil {
			continue
		}
		afJSON, _ := json.Marshal(res.AffectedRules)
		sigJSON, _ := json.Marshal(res.KeySignals)
		_ = h.db.SaveDomainContext(job.ID, string(domain), db.DomainContextRow{
			SimulationID:      job.ID,
			Domain:            string(domain),
			Context:           res.SynthesizedContext,
			Signals:           string(sigJSON),
			StabilityModifier: res.Confidence,
			Confidence:        res.Confidence,
			AffectedRules:     string(afJSON),
			SentimentScore:    res.SentimentScore,
		})
	}

	// Step 4: Enrich with past simulation history (if company known)
	if job.Company != "" {
		historyCtx := dsAgent.EnrichWithHistory(researchCtx, h.ragStore, job.Company, job.Question)
		if historyCtx != "" {
			if enrichedContext != "" {
				enrichedContext = enrichedContext + "\n\n" + historyCtx
			} else {
				enrichedContext = historyCtx
			}
			log.Printf("[FRACTURE] History context injected for sim %s (company: %s)", job.ID, job.Company)
		}
	}

	// Step 5: Run the full FRACTURE simulation with enriched context
	h.runSimulation(job, enrichedContext, domainResults)
}

func (h *Handler) runSimulation(job *simJob, extraContext string, domainResults map[engine.RuleDomain]*deepsearch.DomainResearchResult) {
	h.simMu.Lock()
	job.Status = "running"
	h.persistJob(job)
	h.simMu.Unlock()

	// Build LLM router from stored keys
	router, err := h.buildLLMRouter()
	if err != nil {
		h.simMu.Lock()
		job.Status = "error"
		job.Error = fmt.Sprintf("LLM router error: %v — add at least one API key in Settings", err)
		h.persistJob(job)
		h.simMu.Unlock()
		log.Printf("[FRACTURE] buildLLMRouter failed for sim %s: %v", job.ID, err)
		return
	}

	// Build world from department domain, enriched with DeepSearch domain context
	domain := engine.RuleDomain(job.Department)
	simCtx := context.Background()
	var domainContext string
	var affectedRules []string
	var confidence float64
	if res, ok := domainResults[domain]; ok && res != nil {
		domainContext = res.SynthesizedContext
		affectedRules = res.AffectedRules
		confidence = res.Confidence
	}
	if extraContext != "" {
		if domainContext != "" {
			domainContext = extraContext + "\n\n" + domainContext
		} else {
			domainContext = extraContext
		}
	}
	world, _ := engine.DefaultWorldForDomainWithContext(simCtx, domain, job.Question, domainContext, affectedRules, confidence)

	// Build agents
	conformistLLM := router.ForRole(llm.RoleConformist)
	disruptorLLM := router.ForRole(llm.RoleDisruptor)
	agents := append(
		archetypes.BuiltinConformists(conformistLLM),
		archetypes.BuiltinDisruptors(disruptorLLM)...,
	)

	// Apply archetype calibration: agents with higher accuracy_weight get more influence
	if job.Company != "" {
		if cals, err := h.calibrator.GetCalibrationReport(job.Company); err == nil && len(cals) > 0 {
			engineCals := make([]engine.AgentCalibration, len(cals))
			for i, c := range cals {
				engineCals[i] = engine.AgentCalibration{
					AgentID:        c.ArchetypeID,
					AccuracyWeight: c.AccuracyWeight,
				}
			}
			agents = engine.ApplyCalibration(agents, engineCals)
			log.Printf("[FRACTURE] Applied calibration for %d archetypes (sim %s)", len(cals), job.ID)
		}
	}

	// Build memory store
	memStore := memory.NewStore(h.db.DB)

	jobMode := engine.SimulationMode(job.Mode)
	if jobMode != engine.ModeStandard && jobMode != engine.ModePremium {
		jobMode = engine.ModeStandard
	}
	jobModeCfg := engine.DefaultConfigForMode(jobMode)
	// Use user-specified rounds if provided and within range, otherwise fall back to mode default
	maxRounds := job.Rounds
	if maxRounds <= 0 || maxRounds > 200 {
		maxRounds = jobModeCfg.MaxRounds
	}

	cfg := engine.SimulationConfig{
		ID:         job.ID,
		Question:   job.Question,
		Department: job.Department,
		MaxRounds:  maxRounds,
		Agents:     agents,
		World:      world,
		Memory:     memStore,
		CouncilLLM: router.ForRole(llm.RoleSynthesis),
		Mode:       jobModeCfg,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()

	// Run ensemble: for Premium (EnsembleRuns=2) we run twice independently;
	// for Standard (EnsembleRuns=1) this is a single run with no overhead.
	var primaryResult engine.SimulationResult
	ensembleCfg := engine.EnsembleConfig{Runs: jobModeCfg.EnsembleRuns}

	ensembleResult, ensErr := engine.RunEnsemble(ctx, ensembleCfg, func(ctx context.Context, runIdx int) (*engine.RunResult, error) {
		// Each ensemble run needs a fresh world (independent runs).
		freshWorld, _ := engine.DefaultWorldForDomainWithContext(simCtx, domain, job.Question, domainContext, affectedRules, confidence)
		runCfg := cfg
		runCfg.World = freshWorld
		runCfg.ID = job.ID // keep same ID for round persistence on run 0 only

		sim := engine.NewSimulation(runCfg)
		for rr := range sim.Run(ctx) {
			if runIdx == 0 {
				// Only persist rounds for the primary run
				h.persistRound(job.ID, rr)
				// God View: drain injected events into world evidence
				if job.injectCh != nil {
					for {
						select {
						case ev := <-job.injectCh:
							runCfg.World.InjectEvent(ev)
							log.Printf("[FRACTURE] God View event injected into sim %s: %s", job.ID, ev[:min(80, len(ev))])
						default:
							goto drained
						}
					}
				drained:
				}
			}
		}
		res := sim.Finalize()
		if runIdx == 0 {
			primaryResult = res
		}
		return &engine.RunResult{
			FractureEvents: res.FractureEvents,
			FinalWorld:     res.FinalWorld,
			TensionMap:     res.TensionMap,
			TotalTokens:    res.TotalTokens,
		}, nil
	})
	if ensErr != nil {
		log.Printf("[FRACTURE] Ensemble error for sim %s: %v — using primary result only", job.ID, ensErr)
	}

	result := primaryResult

	// Generate final report — tracked in report_generations table
	synthesisLLM := router.ForRole(llm.RoleSynthesis)
	rg := engine.NewReportGenerator(synthesisLLM)
	reportGenID := uuid.New().String()
	reportStart := time.Now()
	_ = h.db.StartReportGen(reportGenID, job.ID, "full")
	report, reportErr := rg.GenerateReport(ctx, &result, job.Question, engine.CompanyContext{
		Name:     job.Company,
		Size:     job.CompanySize,
		Sector:   job.CompanySector,
		Location: job.CompanyLocation,
	})
	reportDurationMs := time.Since(reportStart).Milliseconds()
	if reportErr != nil {
		log.Printf("[FRACTURE] ReportGenerator error for sim %s: %v", job.ID, reportErr)
		_ = h.db.CompleteReportGen(reportGenID, "error", reportErr.Error(), 0, reportDurationMs)
	} else if report != nil {
		// Attach ensemble results to the report (Premium only — Standard has RunCount=1)
		if ensErr == nil && ensembleResult != nil && ensembleResult.RunCount > 1 {
			report.EnsembleResult = ensembleResult
		}
		_ = h.db.CompleteReportGen(reportGenID, "done", "", report.TotalTokens, reportDurationMs)
	}

	var finalData interface{}
	var durationMs int64
	if reportErr == nil && report != nil {
		finalData = report
		durationMs = report.DurationMs
		log.Printf("[FRACTURE] Report generated for sim %s: %d tokens, %d rupture scenarios", job.ID, report.TotalTokens, len(report.RuptureScenarios))
	} else {
		// No LLM available — build a structured heuristic report from raw simulation data
		// so the UI always receives a valid FullReport (no zeros, no 404)
		finalData = buildHeuristicReport(job.ID, job.Question, &result)
		durationMs = result.DurationMs
		log.Printf("[FRACTURE] Saving heuristic report for sim %s (no LLM)", job.ID)
	}

	// --- JUDGE + EWC fire-and-forget ---
	go func() {
		jCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// round real do primeiro fracture point (0 se não houve)
		actualFractureRound := 0
		if len(primaryResult.FractureEvents) > 0 {
			actualFractureRound = primaryResult.FractureEvents[0].Round
		}

		// sem fonte de round previsto no relatório — usa o real para creditar
		// participantes quando um fracture ocorreu (roundAccuracy = 1.0)
		predictedFractureRound := actualFractureRound

		// monta participação dos agentes
		participation := make(map[string]bool)
		for _, rr := range primaryResult.Rounds {
			for _, action := range rr.Actions {
				participation[action.AgentID] = true
			}
		}

		// constrói e executa julgamento
		judgement := memory.BuildJudgement(
			primaryResult.SimulationID,
			job.Company,
			predictedFractureRound,
			actualFractureRound,
			participation,
		)
		if err := h.calibrator.Judge(jCtx, judgement); err != nil {
			log.Printf("[FRACTURE] judge failed sim=%s: %v", primaryResult.SimulationID, err)
		}

		// consolida pesos EWC para o setor atual
		if h.calibrator.EWC != nil {
			fishers, err := h.calibrator.EWC.ComputeFisherWeights(jCtx, job.Department)
			if err == nil {
				_ = h.calibrator.EWC.ConsolidateWeights(jCtx, job.Department, fishers)
			}
		}
	}()
	// --- fim Judge + EWC ---

	h.simMu.Lock()
	job.Status = "done"
	job.DurationMs = durationMs
	h.persistJob(job)
	h.simMu.Unlock()

	// Persist full result to simulations table
	_ = h.db.SaveSimulation(job.ID, job.Question, job.Department, job.Rounds, finalData)
	_ = h.auditLogger.Log("simulation.completed", job.ID, map[string]interface{}{
		"duration_ms": durationMs,
		"tokens":      result.TotalTokens,
		"fractures":   len(result.FractureEvents),
	})

	// Index simulation artifacts in the RAG store for future simulations
	if job.Company != "" && report != nil && reportErr == nil {
		signals := domainResultsToSignals(domainResults)
		if err := h.ragStore.IndexSimulation(job.Company, *report, signals); err != nil {
			log.Printf("[FRACTURE] RAG index error for sim %s: %v", job.ID, err)
		} else {
			log.Printf("[FRACTURE] RAG indexed sim %s for company %q", job.ID, job.Company)
		}
	}
}

// persistRound saves each agent action and fracture votes from a RoundResult to the DB.
// It also updates the live progress fields on the in-memory simJob for SSE streaming.
// Errors are logged but never fatal — the simulation continues regardless.
func (h *Handler) persistRound(simID string, rr engine.RoundResult) {
	// Update live progress on the in-memory job
	h.simMu.Lock()
	if job, ok := h.simJobs[simID]; ok {
		job.CurrentRound = rr.Round
		job.CurrentTension = rr.Tension
		job.FractureCount += len(rr.FractureEvents)
		for _, action := range rr.Actions {
			job.TotalTokens += action.TokensUsed
			if action.Text != "" {
				job.LastAgentName = action.AgentID
				if len(action.Text) > 120 {
					job.LastAgentAction = action.Text[:120] + "…"
				} else {
					job.LastAgentAction = action.Text
				}
			}
		}
		h.persistJob(job)
	}
	h.simMu.Unlock()
	for _, action := range rr.Actions {
		var newRuleJSON string
		if action.Proposal != nil {
			b, _ := json.Marshal(action.Proposal)
			newRuleJSON = string(b)
		}
		row := &db.RoundRow{
			ID:               uuid.New().String(),
			SimulationID:     simID,
			RoundNumber:      rr.Round,
			AgentID:          action.AgentID,
			AgentType:        string(action.AgentType),
			ActionText:       action.Text,
			TensionLevel:     rr.Tension,
			FractureProposed: action.IsFractureProposal,
			NewRuleJSON:      newRuleJSON,
			TokensUsed:       action.TokensUsed,
			CreatedAt:        time.Now().Unix(),
		}
		if err := h.db.SaveRound(row); err != nil {
			log.Printf("[FRACTURE] SaveRound error sim=%s round=%d agent=%s: %v", simID, rr.Round, action.AgentID, err)
		}
	}
	// Persist fracture votes if any.
	// ProposalID: a stable composite key (simID + round + proposer) that uniquely
	// identifies this proposal — ProposedBy alone is not unique across rounds.
	// VoterType: the agent's archetype type (conformist|disruptor), not its name.
	for _, fe := range rr.FractureEvents {
		// Build a stable proposal ID from simulation + round + proposing agent.
		proposalID := fmt.Sprintf("%s:round%d:%s", simID, fe.Round, fe.ProposedBy)
		for _, vr := range fe.VoteBreakdown {
			// Derive voter_type from agent ID prefix (conformist- / disruptor-).
			voterType := "conformist"
			if len(vr.AgentID) >= 9 && vr.AgentID[:9] == "disruptor" {
				voterType = "disruptor"
			}
			voteRow := &db.VoteRow{
				ID:           uuid.New().String(),
				SimulationID: simID,
				RoundNumber:  fe.Round,
				ProposalID:   proposalID,
				VoterID:      vr.AgentID,
				VoterType:    voterType,
				Vote:         vr.Vote,
				Weight:       vr.Weight,
				Reasoning:    vr.Rationale,
				CreatedAt:    time.Now().Unix(),
			}
			if err := h.db.SaveVote(voteRow); err != nil {
				log.Printf("[FRACTURE] SaveVote error sim=%s round=%d voter=%s: %v", simID, fe.Round, vr.AgentID, err)
			}
		}
	}
}

func (h *Handler) listSimulations(w http.ResponseWriter, r *http.Request) {
	// Return in-memory jobs + DB history
	h.simMu.RLock()
	jobs := make([]*simJob, 0, len(h.simJobs))
	for _, j := range h.simJobs {
		jobs = append(jobs, j)
	}
	h.simMu.RUnlock()

	// Also load from DB
	dbSims, _ := h.db.ListSimulations()
	type simSummary struct {
		ID         string `json:"id"`
		Status     string `json:"status"`
		Question   string `json:"question"`
		Department string `json:"department"`
		Rounds     int    `json:"rounds"`
		CreatedAt  int64  `json:"created_at"`
		DurationMs int64  `json:"duration_ms,omitempty"`
	}

	seen := map[string]bool{}
	result := []simSummary{}
	for _, j := range jobs {
		seen[j.ID] = true
		result = append(result, simSummary{
			ID: j.ID, Status: j.Status, Question: j.Question,
			Department: j.Department, Rounds: j.Rounds,
			CreatedAt: j.CreatedAt, DurationMs: j.DurationMs,
		})
	}
	for _, s := range dbSims {
		if !seen[s.ID] {
			result = append(result, simSummary{
				ID: s.ID, Status: "done", Question: s.Question,
				Department: s.Department, Rounds: s.Rounds,
				CreatedAt: s.CreatedAt, DurationMs: s.DurationMs,
			})
		}
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) getSimulation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	h.simMu.RLock()
	job, ok := h.simJobs[id]
	h.simMu.RUnlock()
	if ok {
		writeJSON(w, http.StatusOK, job)
		return
	}
	// Try DB
	sim, err := h.db.GetSimulation(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "simulation not found")
		return
	}
	writeJSON(w, http.StatusOK, sim)
}

func (h *Handler) streamSimulation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)

	flusher, ok := w.(http.Flusher)
	if !ok {
		return
	}

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	timeout := time.After(12 * time.Minute)

	for {
		select {
		case <-r.Context().Done():
			return
		case <-timeout:
			fmt.Fprintf(w, "event: timeout\ndata: {}\n\n")
			flusher.Flush()
			return
		case <-ticker.C:
			h.simMu.RLock()
			job, ok := h.simJobs[id]
			h.simMu.RUnlock()
			if !ok {
				fmt.Fprintf(w, "event: error\ndata: {\"error\":\"not found\"}\n\n")
				flusher.Flush()
				return
			}
			b, _ := json.Marshal(job)
			fmt.Fprintf(w, "event: update\ndata: %s\n\n", b)
			flusher.Flush()
			if job.Status == "done" || job.Status == "error" {
				return
			}
		}
	}
}

func (h *Handler) deleteSimulation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	h.simMu.Lock()
	delete(h.simJobs, id)
	h.simMu.Unlock()
	_ = h.db.DeleteSimulation(id)
	_ = h.db.DeleteJob(id) // also remove from jobs table
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *Handler) getResults(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	// Results are always fetched from DB (persisted after simulation completes)
	sim, err := h.db.GetSimulation(id)
	if err != nil {
		// Check if still running in memory
		h.simMu.RLock()
		job, ok := h.simJobs[id]
		h.simMu.RUnlock()
		if ok {
			writeJSON(w, http.StatusOK, map[string]string{"status": job.Status, "id": job.ID})
			return
		}
		writeError(w, http.StatusNotFound, "results not found")
		return
	}
	writeJSON(w, http.StatusOK, sim)
}

func (h *Handler) submitFeedback(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var body struct {
		Outcome           string  `json:"outcome"`            // accurate | inaccurate | partial
		PredictedFracture string  `json:"predicted_fracture"` // what the simulation predicted
		ActualOutcome     string  `json:"actual_outcome"`     // what actually happened
		DeltaScore        float64 `json:"delta_score"`        // -1.0 to 1.0
		Notes             string  `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	// a) Persist feedback to DB
	if err := h.db.SaveFeedback(id, body.Outcome, body.Notes); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save feedback")
		return
	}

	// b) Recalibrate archetypes that participated in this simulation
	if body.DeltaScore != 0 {
		companyID := h.companyID()
		feedback := memory.FeedbackRecord{
			SimulationID:      id,
			PredictedFracture: body.PredictedFracture,
			ActualOutcome:     body.ActualOutcome,
			DeltaScore:        body.DeltaScore,
			Notes:             body.Notes,
		}
		if err := h.calibrator.RecordFeedback(companyID, feedback); err != nil {
			log.Printf("[FRACTURE] calibration error for sim %s: %v", id, err)
		}
	}

	// c) Re-index simulation in RAG with feedback metadata (best-effort)
	if body.PredictedFracture != "" || body.ActualOutcome != "" {
		companyID := h.companyID()
		if companyID != "" {
			meta, _ := json.Marshal(map[string]interface{}{
				"simulation_id":      id,
				"feedback_outcome":   body.Outcome,
				"predicted_fracture": body.PredictedFracture,
				"actual_outcome":     body.ActualOutcome,
				"delta_score":        body.DeltaScore,
			})
			content := fmt.Sprintf(
				"Feedback para simulação %s: previsto=%q, real=%q, delta=%.2f",
				id, body.PredictedFracture, body.ActualOutcome, body.DeltaScore,
			)
			_ = h.ragStore.Index(companyID, memory.RAGDocument{
				ID:       "feedback-" + id,
				Type:     memory.DocCompanyContext,
				Content:  content,
				Metadata: string(meta),
			})
		}
	}

	_ = h.auditLogger.Log("feedback.submitted", id, map[string]interface{}{
		"outcome":     body.Outcome,
		"delta_score": body.DeltaScore,
	})
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

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
			"score":    50,
			"level":    "medium",
			"summary":  raw,
			"top_risks": []string{},
		}
	}
	writeJSON(w, http.StatusOK, pulse)
}

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
// builtinArchetypes builds the display list from the real 56 simulation agents
// (37 conformists + 19 disruptors) so the UI always reflects what the engine uses.
func builtinArchetypes() []db.ArchetypeRow {
	agents := append(
		archetypes.BuiltinConformists(nil),
		archetypes.BuiltinDisruptors(nil)...,
	)
	out := make([]db.ArchetypeRow, 0, len(agents))
	for _, a := range agents {
		p := a.Personality()
		agentType := "conformist"
		if a.Type() == engine.AgentDisruptor {
			agentType = "disruptor"
		}
		desc := p.Role
		if len(p.Traits) > 0 {
			desc = p.Role + ": " + strings.Join(p.Traits[:min(3, len(p.Traits))], ", ")
		}
		out = append(out, db.ArchetypeRow{
			ID:           a.ID(),
			Name:         p.Name,
			AgentType:    agentType,
			Description:  desc,
			MemoryWeight: p.PowerWeight,
			IsActive:     true,
		})
	}
	return out
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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

// ─── Audit Log ───────────────────────────────────────────────────────────────

func (h *Handler) getAuditLog(w http.ResponseWriter, r *http.Request) {
	logs, err := h.db.GetAuditLog(50)
	if err != nil {
		writeJSON(w, http.StatusOK, []interface{}{})
		return
	}
	writeJSON(w, http.StatusOK, logs)
}

// ─── LLM Router Builder ──────────────────────────────────────────────────────

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

// ─── Helpers ─────────────────────────────────────────────────────────────────

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

// ─── Report / Export / Compare / Events ──────────────────────────────────────

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

func (h *Handler) getReport(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	report, err := h.loadFullReport(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, report)
}

func (h *Handler) exportMarkdown(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	report, err := h.loadFullReport(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	md := engine.ReportToMarkdown(report)
	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="fracture-%s.md"`, id))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(md))
}

func (h *Handler) exportJSON(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	report, err := h.loadFullReport(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	b, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to encode report")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="fracture-%s.json"`, id))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(b)
}

func (h *Handler) compareSimulations(w http.ResponseWriter, r *http.Request) {
	idsParam := strings.TrimSpace(r.URL.Query().Get("ids"))
	if idsParam == "" {
		writeError(w, http.StatusBadRequest, "ids parameter required (comma-separated, 2–5 IDs)")
		return
	}
	parts := strings.Split(idsParam, ",")
	if len(parts) < 2 || len(parts) > 5 {
		writeError(w, http.StatusBadRequest, "provide between 2 and 5 simulation IDs")
		return
	}
	reports := make([]*engine.FullReport, 0, len(parts))
	for _, rawID := range parts {
		id := strings.TrimSpace(rawID)
		if id == "" {
			continue
		}
		report, err := h.loadFullReport(id)
		if err != nil {
			writeError(w, http.StatusNotFound, fmt.Sprintf("simulation %q: %s", id, err.Error()))
			return
		}
		reports = append(reports, report)
	}
	if len(reports) < 2 {
		writeError(w, http.StatusBadRequest, "at least 2 valid simulation IDs are required")
		return
	}
	writeJSON(w, http.StatusOK, engine.CompareReports(reports))
}

func (h *Handler) getSimulationEvents(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	tensions, err := h.db.GetRoundTensions(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load events: "+err.Error())
		return
	}
	if tensions == nil {
		tensions = []db.TensionPoint{}
	}
	writeJSON(w, http.StatusOK, tensions)
}

// buildHeuristicReport converts a raw SimulationResult into a valid FullReport
// for when no LLM is available (heuristic mode / no API key configured).
func buildHeuristicReport(simID, question string, result *engine.SimulationResult) *engine.FullReport {
	// Build tension entries from the raw map + world rules
	var tensionEntries []engine.TensionEntry
	for ruleID, tension := range result.TensionMap {
		rule, ok := result.FinalWorld.Rules[ruleID]
		if !ok {
			continue
		}
		color := "green"
		switch {
		case tension >= 0.7:
			color = "red"
		case tension >= 0.5:
			color = "orange"
		case tension >= 0.3:
			color = "yellow"
		}
		tensionEntries = append(tensionEntries, engine.TensionEntry{
			RuleID:      ruleID,
			Description: rule.Description,
			Domain:      string(rule.Domain),
			Tension:     tension,
			Color:       color,
		})
	}
	// Sort by tension descending
	for i := 0; i < len(tensionEntries); i++ {
		for j := i + 1; j < len(tensionEntries); j++ {
			if tensionEntries[j].Tension > tensionEntries[i].Tension {
				tensionEntries[i], tensionEntries[j] = tensionEntries[j], tensionEntries[i]
			}
		}
	}

	// Build narrative from simulation data
	fractureCount := len(result.FractureEvents)
	narrative := fmt.Sprintf(
		"Simulação executada em modo heurístico: %d rodadas completadas, %d ponto(s) de fratura detectado(s). "+
			"Para análise narrativa completa com IA, configure a variável de ambiente OPENAI_API_KEY.",
		len(result.Rounds), fractureCount,
	)
	if result.ProbableFuture != "" {
		narrative = result.ProbableFuture
	}

	return &engine.FullReport{
		SimulationID: simID,
		Question:     question,
		ProbableFuture: engine.ProbableFuture{
			Narrative:  narrative,
			Confidence: 0.5,
			KeyAssumptions: []string{
				"Modo heurístico — sem síntese LLM",
				fmt.Sprintf("%d rodadas simuladas", len(result.Rounds)),
				fmt.Sprintf("%d agentes participantes", 56),
			},
		},
		TensionMap:       tensionEntries,
		RuptureScenarios: result.RuptureScenarios,
		FractureEvents:   result.FractureEvents,
		TotalTokens:      result.TotalTokens,
		DurationMs:       result.DurationMs,
		Watermark: engine.Watermark{
			Tool:        "FRACTURE",
			Version:     updater.CurrentVersion,
			URL:         "https://github.com/mkvinicius/fracture",
			License:     "AGPL-3.0",
			GeneratedAt: time.Now().UTC().Format(time.RFC3339),
			Notice:      "Generated by FRACTURE v" + updater.CurrentVersion + " (heuristic mode — configure OPENAI_API_KEY for full AI narrative)",
		},
	}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// ─── Telemetry ───────────────────────────────────────────────────────────────

// getTelemetry returns the current telemetry opt-in status.
func (h *Handler) getTelemetry(w http.ResponseWriter, r *http.Request) {
	enabled := false
	if h.tel != nil {
		enabled = h.tel.IsEnabled()
	}
	writeJSON(w, http.StatusOK, map[string]bool{"enabled": enabled})
}

// setTelemetry updates the telemetry opt-in preference.
func (h *Handler) setTelemetry(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if h.tel != nil {
		if body.Enabled {
			_ = h.tel.Enable()
		} else {
			_ = h.tel.Disable()
		}
	}
	_ = h.auditLogger.Log("telemetry.updated", "system", map[string]bool{"enabled": body.Enabled})
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
