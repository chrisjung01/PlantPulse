# Backend Slice 1 — Foundation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Harden the PlantPulse Go backend with a DB index for query performance, a structured config struct, and graceful HTTP server shutdown on SIGTERM/SIGINT.

**Architecture:** Three independent changes: `schema.sql` gains a composite index applied at startup via the existing `ApplySchema` path; a new `config.go` in `cmd/server/` collects all env-var reads into a typed struct; `main.go` switches from `http.ListenAndServe` to `http.Server` with signal-based graceful shutdown using `cfg.ShutdownTimeout`.

**Tech Stack:** Go stdlib (`net/http`, `os/signal`, `syscall`, `context`), `modernc.org/sqlite`, sqlc — no new dependencies.

## Global Constraints

- `go vet ./...` and `go test ./...` must pass after every commit
- No new external dependencies beyond Go stdlib
- Conventional commits (`feat:`, `fix:`, `chore:`), no co-author lines
- All work is inside `backend/`

---

### Task 1: DB index on (sensor_id, recorded_at)

**Files:**
- Modify: `backend/internal/db/schema.sql`
- Create: `backend/internal/db/migrations_test.go`

**Interfaces:**
- Produces: index `idx_readings_sensor_time` applied by `db.ApplySchema(sqlDB)` on every startup

- [ ] **Step 1: Write the failing test**

Create `backend/internal/db/migrations_test.go`:

```go
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
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd backend && go test ./internal/db/... -run TestApplySchema_CreatesIndex -v
```

Expected: `FAIL` — the index does not exist yet.

- [ ] **Step 3: Add the index to schema.sql**

Replace the content of `backend/internal/db/schema.sql` with:

```sql
CREATE TABLE IF NOT EXISTS readings (
    id                    INTEGER PRIMARY KEY AUTOINCREMENT,
    sensor_id             TEXT    NOT NULL,
    recorded_at           INTEGER NOT NULL,
    temperature_celsius   REAL,
    humidity_percent      REAL,
    soil_moisture_percent REAL
);

CREATE INDEX IF NOT EXISTS idx_readings_sensor_time
    ON readings (sensor_id, recorded_at);
```

- [ ] **Step 4: Run test to verify it passes**

```bash
cd backend && go test ./internal/db/... -run TestApplySchema_CreatesIndex -v
```

Expected: `PASS`

- [ ] **Step 5: Run full test suite**

```bash
cd backend && go vet ./... && go test ./...
```

Expected: all green.

- [ ] **Step 6: Commit**

```bash
git add backend/internal/db/schema.sql backend/internal/db/migrations_test.go
git commit -m "feat(backend): add composite index on (sensor_id, recorded_at)"
```

---

### Task 2: Structured configuration

**Files:**
- Create: `backend/cmd/server/config.go`
- Create: `backend/cmd/server/config_test.go`
- Modify: `backend/cmd/server/main.go`

**Interfaces:**
- Produces: `type config struct` with fields `DatabasePath string`, `AllowedOrigin string`, `Addr string`, `ShutdownTimeout time.Duration`; function `loadConfig() config`; these are consumed by Task 3

- [ ] **Step 1: Write the failing tests**

Create `backend/cmd/server/config_test.go`:

```go
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
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd backend && go test ./cmd/server/... -run TestLoadConfig -v
```

Expected: compile error — `loadConfig` undefined.

- [ ] **Step 3: Create config.go**

Create `backend/cmd/server/config.go`:

```go
package main

import (
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
		if d, err := time.ParseDuration(raw); err == nil {
			return d
		}
	}
	return fallback
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd backend && go test ./cmd/server/... -run TestLoadConfig -v
```

Expected: `PASS` for all three tests.

- [ ] **Step 5: Update main.go to use loadConfig()**

Replace `backend/cmd/server/main.go` with:

```go
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

	slog.Info("server starting", "addr", cfg.Addr)
	if err := http.ListenAndServe(cfg.Addr, handler.LogRequests(cors(cfg.AllowedOrigin, mux))); err != nil {
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
```

- [ ] **Step 6: Run full test suite**

```bash
cd backend && go vet ./... && go test ./...
```

Expected: all green.

- [ ] **Step 7: Commit**

```bash
git add backend/cmd/server/config.go backend/cmd/server/config_test.go backend/cmd/server/main.go
git commit -m "feat(backend): structured config via loadConfig()"
```

---

### Task 3: Graceful shutdown

**Files:**
- Modify: `backend/cmd/server/main.go`
- Create: `backend/cmd/server/shutdown_test.go`

**Interfaces:**
- Consumes: `config.Addr` and `config.ShutdownTimeout` from Task 2 (`loadConfig() config`)
- Produces: server that drains in-flight requests within `ShutdownTimeout` on SIGTERM or SIGINT, then exits 0

- [ ] **Step 1: Write the failing test**

Create `backend/cmd/server/shutdown_test.go`:

```go
package main

import (
	"context"
	"net"
	"net/http"
	"testing"
	"time"
)

// TestGracefulShutdown_CompletesInflightRequests verifies that http.Server.Shutdown
// waits for an in-flight request to finish before returning.
func TestGracefulShutdown_CompletesInflightRequests(t *testing.T) {
	slow := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})

	srv := &http.Server{Handler: slow}

	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}

	go srv.Serve(ln) //nolint:errcheck

	reqDone := make(chan struct{})
	go func() {
		resp, err := http.Get("http://" + ln.Addr().String() + "/")
		if err == nil {
			resp.Body.Close()
		}
		close(reqDone)
	}()

	// Let the request reach the handler, then initiate shutdown.
	time.Sleep(10 * time.Millisecond)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown returned error: %v", err)
	}

	select {
	case <-reqDone:
		// correct: in-flight request completed before shutdown returned
	case <-time.After(200 * time.Millisecond):
		t.Fatal("in-flight request did not complete within 200ms of Shutdown")
	}
}
```

- [ ] **Step 2: Run test to verify it passes**

```bash
cd backend && go test ./cmd/server/... -run TestGracefulShutdown -v
```

Expected: `PASS` — this establishes a regression guard for the stdlib shutdown behaviour we are about to rely on.

- [ ] **Step 3: Update main.go with signal-based graceful shutdown**

Replace `backend/cmd/server/main.go` with the final version:

```go
package main

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

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
		Addr:    cfg.Addr,
		Handler: handler.LogRequests(cors(cfg.AllowedOrigin, mux)),
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
```

- [ ] **Step 4: Run full test suite**

```bash
cd backend && go vet ./... && go test -race ./...
```

Expected: all green (race detector enabled to match CI).

- [ ] **Step 5: Commit**

```bash
git add backend/cmd/server/main.go backend/cmd/server/shutdown_test.go
git commit -m "feat(backend): graceful shutdown on SIGTERM/SIGINT"
```
