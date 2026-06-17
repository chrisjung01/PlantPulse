package handler_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"plantpulse/backend/internal/db"
	"plantpulse/backend/internal/handler"

	_ "modernc.org/sqlite"
)

func setupDB(t *testing.T) *db.Queries {
	t.Helper()
	sqlDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { sqlDB.Close() })
	if err := db.ApplySchema(sqlDB); err != nil {
		t.Fatal(err)
	}
	return db.New(sqlDB)
}

func TestInsertReading(t *testing.T) {
	queries := setupDB(t)
	h := handler.NewReadings(queries)

	body := `{
		"sensor_id": "esp32-test",
		"recorded_at": 1750000000,
		"temperature_celsius": 22.4,
		"humidity_percent": 61.0,
		"soil_moisture_percent": 38.5
	}`
	req := httptest.NewRequest(http.MethodPost, "/readings", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Insert(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp["sensor_id"] != "esp32-test" {
		t.Errorf("unexpected sensor_id: %v", resp["sensor_id"])
	}
	if resp["id"] == nil {
		t.Error("expected id in response")
	}
}

func TestListReadings(t *testing.T) {
	queries := setupDB(t)
	h := handler.NewReadings(queries)

	insertBody := `{"sensor_id":"esp32-test","recorded_at":1750000000,"temperature_celsius":22.4}`
	insertReq := httptest.NewRequest(http.MethodPost, "/readings", bytes.NewBufferString(insertBody))
	insertReq.Header.Set("Content-Type", "application/json")
	insertW := httptest.NewRecorder()
	h.Insert(insertW, insertReq)
	if insertW.Code != http.StatusCreated {
		t.Fatalf("insert failed: %d %s", insertW.Code, insertW.Body.String())
	}

	listReq := httptest.NewRequest(http.MethodGet, "/readings?limit=10", nil)
	listW := httptest.NewRecorder()
	h.List(listW, listReq)

	if listW.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", listW.Code, listW.Body.String())
	}
	var readings []map[string]any
	if err := json.NewDecoder(listW.Body).Decode(&readings); err != nil {
		t.Fatal(err)
	}
	if len(readings) != 1 {
		t.Errorf("expected 1 reading, got %d", len(readings))
	}
}

func TestHealth(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	handler.Health(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	body := w.Body.String()
	if body != `{"status":"ok"}`+"\n" {
		t.Errorf("unexpected body: %s", body)
	}
}

func TestInsertReading_MissingSensorID(t *testing.T) {
	queries := setupDB(t)
	h := handler.NewReadings(queries)

	body := `{"sensor_id":"","recorded_at":1750000000}`
	req := httptest.NewRequest(http.MethodPost, "/readings", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Insert(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestInsertReading_MissingRecordedAt(t *testing.T) {
	queries := setupDB(t)
	h := handler.NewReadings(queries)

	body := `{"sensor_id":"esp32-test","recorded_at":0}`
	req := httptest.NewRequest(http.MethodPost, "/readings", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Insert(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestInsertReading_InvalidJSON(t *testing.T) {
	queries := setupDB(t)
	h := handler.NewReadings(queries)

	req := httptest.NewRequest(http.MethodPost, "/readings", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Insert(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestInsertReading_NullableFields(t *testing.T) {
	queries := setupDB(t)
	h := handler.NewReadings(queries)

	body := `{"sensor_id":"esp32-test","recorded_at":1750000000}`
	req := httptest.NewRequest(http.MethodPost, "/readings", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Insert(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if _, ok := resp["temperature_celsius"]; !ok {
		t.Error("expected temperature_celsius key in response")
	}
	if resp["temperature_celsius"] != nil {
		t.Errorf("expected temperature_celsius to be null, got %v", resp["temperature_celsius"])
	}
	if _, ok := resp["humidity_percent"]; !ok {
		t.Error("expected humidity_percent key in response")
	}
	if resp["humidity_percent"] != nil {
		t.Errorf("expected humidity_percent to be null, got %v", resp["humidity_percent"])
	}
	if _, ok := resp["soil_moisture_percent"]; !ok {
		t.Error("expected soil_moisture_percent key in response")
	}
	if resp["soil_moisture_percent"] != nil {
		t.Errorf("expected soil_moisture_percent to be null, got %v", resp["soil_moisture_percent"])
	}
}

func TestListReadings_LimitZero(t *testing.T) {
	queries := setupDB(t)
	h := handler.NewReadings(queries)

	req := httptest.NewRequest(http.MethodGet, "/readings?limit=0", nil)
	w := httptest.NewRecorder()

	h.List(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestListReadings_LimitTooHigh(t *testing.T) {
	queries := setupDB(t)
	h := handler.NewReadings(queries)

	req := httptest.NewRequest(http.MethodGet, "/readings?limit=501", nil)
	w := httptest.NewRecorder()

	h.List(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestListReadings_LimitInvalid(t *testing.T) {
	queries := setupDB(t)
	h := handler.NewReadings(queries)

	req := httptest.NewRequest(http.MethodGet, "/readings?limit=abc", nil)
	w := httptest.NewRecorder()

	h.List(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestListReadings_Empty(t *testing.T) {
	queries := setupDB(t)
	h := handler.NewReadings(queries)

	req := httptest.NewRequest(http.MethodGet, "/readings?limit=10", nil)
	w := httptest.NewRecorder()

	h.List(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var readings []map[string]any
	if err := json.NewDecoder(w.Body).Decode(&readings); err != nil {
		t.Fatal(err)
	}
	if readings == nil {
		t.Error("expected [] not null")
	}
	if len(readings) != 0 {
		t.Errorf("expected 0 readings, got %d", len(readings))
	}
}

func insertTestReading(t *testing.T, queries *db.Queries, sensorID string, recordedAt int64) {
	t.Helper()
	_, err := queries.InsertReading(context.Background(), db.InsertReadingParams{
		SensorID:            sensorID,
		RecordedAt:          recordedAt,
		TemperatureCelsius:  sql.NullFloat64{Float64: 22.0, Valid: true},
		HumidityPercent:     sql.NullFloat64{Float64: 60.0, Valid: true},
		SoilMoisturePercent: sql.NullFloat64{Float64: 40.0, Valid: true},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestSensorsList(t *testing.T) {
	queries := setupDB(t)
	insertTestReading(t, queries, "esp32-A", 1750000000)

	h := handler.NewSensors(queries)
	req := httptest.NewRequest(http.MethodGet, "/sensors", nil)
	w := httptest.NewRecorder()
	h.List(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var ids []string
	if err := json.NewDecoder(w.Body).Decode(&ids); err != nil {
		t.Fatal(err)
	}
	if len(ids) != 1 || ids[0] != "esp32-A" {
		t.Errorf("unexpected ids: %v", ids)
	}
}

func TestSensorsList_Empty(t *testing.T) {
	queries := setupDB(t)
	h := handler.NewSensors(queries)
	req := httptest.NewRequest(http.MethodGet, "/sensors", nil)
	w := httptest.NewRecorder()
	h.List(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var ids []string
	if err := json.NewDecoder(w.Body).Decode(&ids); err != nil {
		t.Fatal(err)
	}
	if ids == nil {
		t.Error("expected [] not null")
	}
	if len(ids) != 0 {
		t.Errorf("expected 0 ids, got %d", len(ids))
	}
}
