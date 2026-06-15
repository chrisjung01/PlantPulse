# Backend — Agent Guidelines

## Language & dependencies

- Go only. No framework — `net/http` is enough for Slice 1.
- `modernc.org/sqlite` for the database driver (pure Go, no CGO required).
- sqlc for query generation. Add new queries to `schema.sql` / `.sql` query files, then run `sqlc generate`.

## What to implement (Slice 1)

1. `POST /readings` — decode JSON, validate required fields, insert into `readings`.
2. `GET /readings?limit=N` — query last N rows ordered by `ts DESC`, return JSON array.
3. CORS middleware: `Access-Control-Allow-Origin: http://localhost:5173`.

## Code rules

- No global variables. Wire dependencies in `main.go` and pass them down.
- Errors bubble up; handlers write a JSON error response and an appropriate HTTP status.
- No logging library — `log/slog` (stdlib) if logging is needed.
- Test files live alongside the code they test (`_test.go`).

## Out of scope right now

- Authentication / API keys
- MQTT or any message broker
- Plant, Sensor, or GrowCycle models
- Migrations tooling (plain `schema.sql` applied at startup is fine)
