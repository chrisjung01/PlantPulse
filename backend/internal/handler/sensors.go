package handler

import (
	"context"
	"database/sql"
	"net/http"
	"strconv"

	"plantpulse/backend/internal/db"
)

type Sensors struct {
	q *db.Queries
}

func NewSensors(q *db.Queries) *Sensors {
	return &Sensors{q: q}
}

type aggregateResponse struct {
	Bucket              string   `json:"bucket"`
	TemperatureCelsius  *float64 `json:"temperature_celsius"`
	HumidityPercent     *float64 `json:"humidity_percent"`
	SoilMoisturePercent *float64 `json:"soil_moisture_percent"`
	SampleCount         int64    `json:"sample_count"`
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

func (h *Sensors) ListReadings(w http.ResponseWriter, r *http.Request) {
	sensorID := r.PathValue("id")

	fromRaw := r.URL.Query().Get("from")
	toRaw := r.URL.Query().Get("to")
	if fromRaw == "" || toRaw == "" {
		respondError(w, http.StatusBadRequest, "from and to are required")
		return
	}
	from, err := strconv.ParseInt(fromRaw, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "from must be a unix timestamp")
		return
	}
	to, err := strconv.ParseInt(toRaw, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "to must be a unix timestamp")
		return
	}
	if from >= to {
		respondError(w, http.StatusBadRequest, "from must be before to")
		return
	}

	limit := int64(100)
	if raw := r.URL.Query().Get("limit"); raw != "" {
		n, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || n < 1 || n > 1000 {
			respondError(w, http.StatusBadRequest, "limit must be between 1 and 1000")
			return
		}
		limit = n
	}

	rows, err := h.q.ListReadingsBySensor(context.Background(), db.ListReadingsBySensorParams{
		SensorID: sensorID,
		From:     from,
		To:       to,
		Limit:    limit,
	})
	if err != nil {
		respondError(w, http.StatusInternalServerError, "database error")
		return
	}

	resp := make([]readingResponse, 0, len(rows))
	for _, row := range rows {
		resp = append(resp, toResponse(row))
	}
	respondJSON(w, http.StatusOK, resp)
}

func (h *Sensors) AggregateReadings(w http.ResponseWriter, r *http.Request) {
	sensorID := r.PathValue("id")

	fromRaw := r.URL.Query().Get("from")
	toRaw := r.URL.Query().Get("to")
	granularity := r.URL.Query().Get("granularity")

	if fromRaw == "" || toRaw == "" {
		respondError(w, http.StatusBadRequest, "from and to are required")
		return
	}
	if granularity != "hour" && granularity != "day" {
		respondError(w, http.StatusBadRequest, "granularity must be hour or day")
		return
	}

	from, err := strconv.ParseInt(fromRaw, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "from must be a unix timestamp")
		return
	}
	to, err := strconv.ParseInt(toRaw, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "to must be a unix timestamp")
		return
	}
	if from >= to {
		respondError(w, http.StatusBadRequest, "from must be before to")
		return
	}

	resp := make([]aggregateResponse, 0)

	if granularity == "hour" {
		rows, err := h.q.AggregateReadingsBySensorByHour(context.Background(), db.AggregateReadingsBySensorByHourParams{
			SensorID: sensorID,
			From:     from,
			To:       to,
		})
		if err != nil {
			respondError(w, http.StatusInternalServerError, "database error")
			return
		}
		for _, row := range rows {
			resp = append(resp, toAggregateResponse(row.Bucket, row.TemperatureCelsius, row.HumidityPercent, row.SoilMoisturePercent, row.SampleCount))
		}
	} else {
		rows, err := h.q.AggregateReadingsBySensorByDay(context.Background(), db.AggregateReadingsBySensorByDayParams{
			SensorID: sensorID,
			From:     from,
			To:       to,
		})
		if err != nil {
			respondError(w, http.StatusInternalServerError, "database error")
			return
		}
		for _, row := range rows {
			resp = append(resp, toAggregateResponse(row.Bucket, row.TemperatureCelsius, row.HumidityPercent, row.SoilMoisturePercent, row.SampleCount))
		}
	}

	respondJSON(w, http.StatusOK, resp)
}

func toAggregateResponse(bucket string, temp, hum, soil sql.NullFloat64, count int64) aggregateResponse {
	a := aggregateResponse{Bucket: bucket, SampleCount: count}
	if temp.Valid {
		a.TemperatureCelsius = &temp.Float64
	}
	if hum.Valid {
		a.HumidityPercent = &hum.Float64
	}
	if soil.Valid {
		a.SoilMoisturePercent = &soil.Float64
	}
	return a
}
