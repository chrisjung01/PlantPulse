package handler_test

import (
	"bytes"
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
