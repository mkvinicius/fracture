package api

import (
	"encoding/json"
	"net/http"

	"github.com/fracture/fracture/contextextractor"
	"github.com/fracture/fracture/updater"
)

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
		URL         string `json:"url"`
		SourceType  string `json:"source_type"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Content     string `json:"content"`
		Error       string `json:"error,omitempty"`
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
		"sources":    sources,
		"summary":    ctx.Summary,
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

// ─── Audit Log ───────────────────────────────────────────────────────────────

func (h *Handler) getAuditLog(w http.ResponseWriter, r *http.Request) {
	logs, err := h.db.GetAuditLog(50)
	if err != nil {
		writeJSON(w, http.StatusOK, []interface{}{})
		return
	}
	writeJSON(w, http.StatusOK, logs)
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
