package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/fracture/fracture/archetypes"
	"github.com/fracture/fracture/contextextractor"
	"github.com/fracture/fracture/db"
	"github.com/fracture/fracture/deepsearch"
	"github.com/fracture/fracture/engine"
	"github.com/fracture/fracture/llm"
	"github.com/fracture/fracture/memory"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (h *Handler) createSimulation(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Question   string   `json:"question"`
		Department string   `json:"department"`
		Rounds     int      `json:"rounds"`
		Mode       string   `json:"mode"`
		Context    string   `json:"context"`
		URLs       []string `json:"urls"` // optional: company website + social media URLs
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
		job.Error = "no LLM keys configured — add at least one API key in Settings"
		h.persistJob(job)
		h.simMu.Unlock()
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

	cfg := engine.SimulationConfig{
		ID:         job.ID,
		Question:   job.Question,
		Department: job.Department,
		MaxRounds:  jobModeCfg.MaxRounds,
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
	report, reportErr := rg.GenerateReport(ctx, &result, job.Question)
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
		finalData = &result
		durationMs = result.DurationMs
		log.Printf("[FRACTURE] Saving raw result for sim %s (no report)", job.ID)
	}

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
	var result []simSummary
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
