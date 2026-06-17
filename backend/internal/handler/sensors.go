package handler

import (
	"context"
	"net/http"

	"plantpulse/backend/internal/db"
)

type Sensors struct {
	q *db.Queries
}

func NewSensors(q *db.Queries) *Sensors {
	return &Sensors{q: q}
}

func (h *Sensors) List(w http.ResponseWriter, r *http.Request) {
	ids, err := h.q.ListSensorIDs(context.Background())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "database error")
		return
	}
	if ids == nil {
		ids = []string{}
	}
	respondJSON(w, http.StatusOK, ids)
}
