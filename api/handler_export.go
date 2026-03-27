package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/fracture/fracture/engine"
	"github.com/go-chi/chi/v5"
)

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
