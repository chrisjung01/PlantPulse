package main

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"plantpulse/backend/internal/db"
	"plantpulse/backend/internal/handler"

	_ "modernc.org/sqlite"
)

func main() {
	cfg := loadConfig()

	sqlDB, err := sql.Open("sqlite", cfg.DatabasePath)
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

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           handler.LogRequests(cors(cfg.AllowedOrigin, mux)),
		ReadHeaderTimeout: 5 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		slog.Info("server starting", "addr", cfg.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	<-quit
	signal.Stop(quit)
	slog.Info("shutting down", "timeout", cfg.ShutdownTimeout)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("shutdown error", "err", err)
		os.Exit(1)
	}
	slog.Info("server stopped")
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
