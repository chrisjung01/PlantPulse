package db_test

import (
	"database/sql"
	"testing"

	"plantpulse/backend/internal/db"

	_ "modernc.org/sqlite"
)

func TestApplySchema_CreatesIndex(t *testing.T) {
	sqlDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()

	if err := db.ApplySchema(sqlDB); err != nil {
		t.Fatalf("ApplySchema failed: %v", err)
	}

	var count int
	err = sqlDB.QueryRow(
		`SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name='idx_readings_sensor_time'`,
	).Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatal("expected index idx_readings_sensor_time to exist, but it does not")
	}
}
