# PlantPulse — Agent Guidelines

## Project overview

Self-hosted plant monitoring. ESP32 sensors → Go HTTP API → SQLite → SvelteKit frontend.
This is a learning project; favour clarity and simplicity over cleverness.

## Monorepo structure

```
backend/    Go API  (see backend/AGENTS.md)
frontend/   SvelteKit app (see frontend/AGENTS.md)
```

## General rules

- **No auth, no MQTT** in Slice 1 — HTTP POST only.
- Keep the database schema flat (`readings` table). Do not introduce a `Plant` or `Sensor` model yet.
- Standard library first. Add a dependency only when the stdlib genuinely can't do the job.
- No co-author lines in commits.

## Cross-cutting conventions

- Timestamps in UTC, stored as Unix seconds.
- `sensor_id` is a plain string (e.g. `"balcony-1"`); no UUID library needed yet.
- CORS: allow `http://localhost:5173` in the Go server during development.
