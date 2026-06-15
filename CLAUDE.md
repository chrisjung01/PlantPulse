# PlantPulse

Self-hosted, vendor-independent plant monitoring platform. ESP32 sensors report environmental data to a Go backend; a SvelteKit frontend visualises it.

## Repo layout

```
backend/    Go API server
frontend/   SvelteKit app
```

## Tech stack

| Layer    | Choice                                      |
|----------|---------------------------------------------|
| Backend  | Go, `net/http`, `modernc.org/sqlite`, sqlc  |
| Frontend | SvelteKit + TypeScript                      |
| Database | SQLite (embedded, single file)              |
| Sensors  | ESP32 → HTTP POST `/readings`               |

## Running locally

```bash
# backend  (port 8080)
cd backend && go run ./cmd/server

# frontend (port 5173)
cd frontend && npm run dev
```

CORS is enabled on the Go server for local development.

## Commit style

Conventional commits (`feat:`, `fix:`, `chore:`, …). No co-author lines.

## What's intentionally out of scope right now

- Authentication
- MQTT
- Plant / sensor-type domain model (flat `readings` table only)
- ML-based recommendations
