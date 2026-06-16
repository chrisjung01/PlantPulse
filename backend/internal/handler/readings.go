package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"plantpulse/backend/internal/db"
)

type Readings struct {
	q *db.Queries
}

func NewReadings(q *db.Queries) *Readings {
	return &Readings{q: q}
}

type insertRequest struct {
	SensorID            string   `json:"sensor_id"`
	RecordedAt          int64    `json:"recorded_at"`
	TemperatureCelsius  *float64 `json:"temperature_celsius"`
	HumidityPercent     *float64 `json:"humidity_percent"`
	SoilMoisturePercent *float64 `json:"soil_moisture_percent"`
}

type readingResponse struct {
	ID                  int64    `json:"id"`
	SensorID            string   `json:"sensor_id"`
	RecordedAt          int64    `json:"recorded_at"`
	TemperatureCelsius  *float64 `json:"temperature_celsius"`
	HumidityPercent     *float64 `json:"humidity_percent"`
	SoilMoisturePercent *float64 `json:"soil_moisture_percent"`
}

func toResponse(r db.Reading) readingResponse {
	resp := readingResponse{
		ID:         r.ID,
		SensorID:   r.SensorID,
		RecordedAt: r.RecordedAt,
	}
	if r.TemperatureCelsius.Valid {
		resp.TemperatureCelsius = &r.TemperatureCelsius.Float64
	}
	if r.HumidityPercent.Valid {
		resp.HumidityPercent = &r.HumidityPercent.Float64
	}
	if r.SoilMoisturePercent.Valid {
		resp.SoilMoisturePercent = &r.SoilMoisturePercent.Float64
	}
	return resp
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func (h *Readings) Insert(w http.ResponseWriter, r *http.Request) {
	var req insertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if req.SensorID == "" {
		writeError(w, http.StatusBadRequest, "sensor_id is required")
		return
	}
	if req.RecordedAt == 0 {
		writeError(w, http.StatusBadRequest, "recorded_at is required")
		return
	}

	params := db.InsertReadingParams{
		SensorID:   req.SensorID,
		RecordedAt: req.RecordedAt,
	}
	if req.TemperatureCelsius != nil {
		params.TemperatureCelsius = sql.NullFloat64{Float64: *req.TemperatureCelsius, Valid: true}
	}
	if req.HumidityPercent != nil {
		params.HumidityPercent = sql.NullFloat64{Float64: *req.HumidityPercent, Valid: true}
	}
	if req.SoilMoisturePercent != nil {
		params.SoilMoisturePercent = sql.NullFloat64{Float64: *req.SoilMoisturePercent, Valid: true}
	}

	reading, err := h.q.InsertReading(context.Background(), params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}

	writeJSON(w, http.StatusCreated, toResponse(reading))
}

func (h *Readings) List(w http.ResponseWriter, r *http.Request) {
	limit := int64(20)
	if raw := r.URL.Query().Get("limit"); raw != "" {
		n, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || n < 1 || n > 500 {
			writeError(w, http.StatusBadRequest, "limit must be between 1 and 500")
			return
		}
		limit = n
	}

	rows, err := h.q.ListReadings(context.Background(), limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}

	resp := make([]readingResponse, 0, len(rows))
	for _, row := range rows {
		resp = append(resp, toResponse(row))
	}
	writeJSON(w, http.StatusOK, resp)
}
