package main

import (
	"testing"
	"time"
)

func TestLoadConfig_Defaults(t *testing.T) {
	t.Setenv("DATABASE_PATH", "")
	t.Setenv("ALLOWED_ORIGIN", "")
	t.Setenv("ADDR", "")
	t.Setenv("SHUTDOWN_TIMEOUT", "")

	cfg := loadConfig()

	if cfg.DatabasePath != "./plantpulse.db" {
		t.Errorf("DatabasePath: got %q, want %q", cfg.DatabasePath, "./plantpulse.db")
	}
	if cfg.AllowedOrigin != "http://localhost:5173" {
		t.Errorf("AllowedOrigin: got %q, want %q", cfg.AllowedOrigin, "http://localhost:5173")
	}
	if cfg.Addr != ":8080" {
		t.Errorf("Addr: got %q, want %q", cfg.Addr, ":8080")
	}
	if cfg.ShutdownTimeout != 10*time.Second {
		t.Errorf("ShutdownTimeout: got %v, want 10s", cfg.ShutdownTimeout)
	}
}

func TestLoadConfig_EnvOverrides(t *testing.T) {
	t.Setenv("DATABASE_PATH", "/data/plant.db")
	t.Setenv("ALLOWED_ORIGIN", "https://example.com")
	t.Setenv("ADDR", ":9090")
	t.Setenv("SHUTDOWN_TIMEOUT", "30s")

	cfg := loadConfig()

	if cfg.DatabasePath != "/data/plant.db" {
		t.Errorf("DatabasePath: got %q, want %q", cfg.DatabasePath, "/data/plant.db")
	}
	if cfg.AllowedOrigin != "https://example.com" {
		t.Errorf("AllowedOrigin: got %q, want %q", cfg.AllowedOrigin, "https://example.com")
	}
	if cfg.Addr != ":9090" {
		t.Errorf("Addr: got %q, want %q", cfg.Addr, ":9090")
	}
	if cfg.ShutdownTimeout != 30*time.Second {
		t.Errorf("ShutdownTimeout: got %v, want 30s", cfg.ShutdownTimeout)
	}
}

func TestLoadConfig_InvalidTimeoutFallsBackToDefault(t *testing.T) {
	t.Setenv("SHUTDOWN_TIMEOUT", "not-a-duration")

	cfg := loadConfig()

	if cfg.ShutdownTimeout != 10*time.Second {
		t.Errorf("ShutdownTimeout: got %v, want 10s (default)", cfg.ShutdownTimeout)
	}
}
