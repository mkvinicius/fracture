package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/fracture/fracture/db"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// RunScheduledSim fires a scheduled simulation as if triggered by a user.
// Called by the main.go scheduler goroutine.
func (h *Handler) RunScheduledSim(s db.ScheduledSim) {
	job := &simJob{
		ID:         uuid.New().String(),
		Status:     "queued",
		Question:   s.Question,
		Department: s.Department,
		Rounds:     s.Rounds,
		Mode:       "standard",
		CreatedAt:  time.Now().Unix(),
		injectCh:   make(chan string, 8),
	}
	h.simMu.Lock()
	h.simJobs[job.ID] = job
	h.persistJob(job)
	h.simMu.Unlock()
	go h.runWithDeepSearch(job, s.Context)
}

// ─── Share Link ───────────────────────────────────────────────────────────────

func (h *Handler) createShareLink(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	// verify simulation exists
	if _, err := h.db.GetSimulation(id); err != nil {
		writeError(w, http.StatusNotFound, "simulation not found")
		return
	}
	token, err := h.db.GenerateShareToken(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate share token")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}

func (h *Handler) getSharedReport(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	row, err := h.db.GetSimulationByShareToken(token)
	if err != nil {
		writeError(w, http.StatusNotFound, "shared report not found")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(row.ResultJSON))
}

// ─── God View ─────────────────────────────────────────────────────────────────

func (h *Handler) injectEvent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var body struct {
		Event string `json:"event"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Event == "" {
		writeError(w, http.StatusBadRequest, "event text is required")
		return
	}
	cleanEvent, err := h.sanitizer.Sanitize(r.Context(), body.Event)
	if err != nil || cleanEvent == "" {
		writeError(w, http.StatusBadRequest, "event contains invalid content")
		return
	}

	h.simMu.RLock()
	job, ok := h.simJobs[id]
	h.simMu.RUnlock()

	if !ok || job.Status != "running" {
		writeError(w, http.StatusConflict, "simulation is not currently running")
		return
	}
	if job.injectCh == nil {
		writeError(w, http.StatusConflict, "simulation does not support event injection")
		return
	}
	select {
	case job.injectCh <- cleanEvent:
		_ = h.auditLogger.Log("godview.inject", id, map[string]string{"event": cleanEvent})
		writeJSON(w, http.StatusOK, map[string]string{"status": "queued"})
	default:
		writeError(w, http.StatusTooManyRequests, "injection queue full — try again in a moment")
	}
}

// ─── Prediction Outcomes (Accuracy Tracking) ──────────────────────────────────

func (h *Handler) getOutcomes(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	outcomes, err := h.db.GetPredictionOutcomes(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load outcomes")
		return
	}
	if outcomes == nil {
		outcomes = []db.PredictionOutcome{}
	}
	writeJSON(w, http.StatusOK, outcomes)
}

func (h *Handler) saveOutcome(w http.ResponseWriter, r *http.Request) {
	simID := chi.URLParam(r, "id")
	var body struct {
		RuleID             string `json:"rule_id"`
		Prediction         string `json:"prediction"`
		Outcome            string `json:"outcome"` // pending | confirmed | refuted | partial
		Notes              string `json:"notes"`
		FractureEventRound int    `json:"fracture_event_round"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if body.Outcome == "" {
		body.Outcome = "pending"
	}
	p := db.PredictionOutcome{
		ID:                 uuid.New().String(),
		SimulationID:       simID,
		FractureEventRound: body.FractureEventRound,
		RuleID:             body.RuleID,
		Prediction:         body.Prediction,
		Outcome:            body.Outcome,
		Notes:              body.Notes,
	}
	if err := h.db.SavePredictionOutcome(p); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save outcome")
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (h *Handler) getAccuracy(w http.ResponseWriter, r *http.Request) {
	stats, err := h.db.GetAccuracyStats()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load accuracy stats")
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

// ─── Scheduled Simulations ────────────────────────────────────────────────────

func (h *Handler) listSchedules(w http.ResponseWriter, r *http.Request) {
	schedules, err := h.db.ListScheduledSims()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load schedules")
		return
	}
	if schedules == nil {
		schedules = []db.ScheduledSim{}
	}
	writeJSON(w, http.StatusOK, schedules)
}

func (h *Handler) createSchedule(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Question   string `json:"question"`
		Department string `json:"department"`
		Rounds     int    `json:"rounds"`
		Context    string `json:"context"`
		IntervalH  int    `json:"interval_h"` // hours: 24=daily, 168=weekly, 720=monthly
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
	if body.Rounds <= 0 {
		body.Rounds = 20
	}
	if body.IntervalH <= 0 {
		body.IntervalH = 168 // weekly by default
	}
	s := db.ScheduledSim{
		ID:         uuid.New().String(),
		Question:   body.Question,
		Department: body.Department,
		Rounds:     body.Rounds,
		Context:    body.Context,
		IntervalH:  body.IntervalH,
		Enabled:    true,
		NextRunAt:  time.Now().Unix(), // run immediately on first tick
	}
	if err := h.db.CreateScheduledSim(s); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create schedule")
		return
	}
	_ = h.auditLogger.Log("schedule.created", s.ID, map[string]interface{}{
		"question": s.Question, "interval_h": s.IntervalH,
	})
	writeJSON(w, http.StatusCreated, s)
}

func (h *Handler) deleteSchedule(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.db.DeleteScheduledSim(id); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete schedule")
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *Handler) toggleSchedule(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var body struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if err := h.db.UpdateScheduledSim(id, body.Enabled); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update schedule")
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// ─── API Keys ─────────────────────────────────────────────────────────────────

func (h *Handler) listAPIKeys(w http.ResponseWriter, r *http.Request) {
	keys, err := h.db.ListAPIKeys()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load API keys")
		return
	}
	if keys == nil {
		keys = []db.APIKey{}
	}
	writeJSON(w, http.StatusOK, keys)
}

func (h *Handler) createAPIKey(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name      string `json:"name"`
		SimsLimit int    `json:"sims_limit"` // 0 = unlimited
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if body.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	key, err := h.db.CreateAPIKey(uuid.New().String(), body.Name, body.SimsLimit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create API key")
		return
	}
	_ = h.auditLogger.Log("apikey.created", key.ID, map[string]string{"name": key.Name})
	writeJSON(w, http.StatusCreated, key)
}

func (h *Handler) deleteAPIKey(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.db.DeleteAPIKey(id); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete API key")
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
