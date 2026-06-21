package main

import (
	"log/slog"
	"os"
	"time"
)

type config struct {
	DatabasePath    string
	AllowedOrigin   string
	Addr            string
	ShutdownTimeout time.Duration
}

func loadConfig() config {
	return config{
		DatabasePath:    envOr("DATABASE_PATH", "./plantpulse.db"),
		AllowedOrigin:   envOr("ALLOWED_ORIGIN", "http://localhost:5173"),
		Addr:            envOr("ADDR", ":8080"),
		ShutdownTimeout: parseDuration("SHUTDOWN_TIMEOUT", 10*time.Second),
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func parseDuration(key string, fallback time.Duration) time.Duration {
	if raw := os.Getenv(key); raw != "" {
		d, err := time.ParseDuration(raw)
		if err != nil {
			slog.Warn("invalid duration env var, using default",
				"key", key, "value", raw, "default", fallback)
			return fallback
		}
		return d
	}
	return fallback
}
