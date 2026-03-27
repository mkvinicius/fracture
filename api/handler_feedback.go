package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/fracture/fracture/memory"
	"github.com/go-chi/chi/v5"
)

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
