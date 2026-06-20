package main

import (
	"database/sql"
	"log/slog"
	"net/http"
	"os"

	"plantpulse/backend/internal/db"
	"plantpulse/backend/internal/handler"

	_ "modernc.org/sqlite"
)

func main() {
	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "./plantpulse.db"
	}

	allowedOrigin := os.Getenv("ALLOWED_ORIGIN")
	if allowedOrigin == "" {
		allowedOrigin = "http://localhost:5173"
	}

	sqlDB, err := sql.Open("sqlite", dbPath)
	if err != nil {
		slog.Error("open db", "err", err)
		os.Exit(1)
	}
	defer sqlDB.Close()

	if err := db.ApplySchema(sqlDB); err != nil {
		slog.Error("apply schema", "err", err)
		os.Exit(1)
	}

	queries := db.New(sqlDB)
	readings := handler.NewReadings(queries)
	sensors := handler.NewSensors(queries)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handler.Health)
	mux.HandleFunc("POST /readings", readings.Insert)
	mux.HandleFunc("GET /readings", readings.List)
	mux.HandleFunc("GET /sensors", sensors.List)
	mux.HandleFunc("GET /sensors/{id}/readings", sensors.ListReadings)
	mux.HandleFunc("GET /sensors/{id}/readings/aggregated", sensors.AggregateReadings)

	slog.Info("server starting", "addr", ":8080")
	if err := http.ListenAndServe(":8080", handler.LogRequests(cors(allowedOrigin, mux))); err != nil {
		slog.Error("server stopped", "err", err)
		os.Exit(1)
	}
}

func cors(origin string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
