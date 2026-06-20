# Backend — Go API

## Stack

- **Language**: Go (standard library `net/http`, no framework)
- **Database**: SQLite via `modernc.org/sqlite` (pure Go, no CGO)
- **Queries**: sqlc — write SQL, get typed Go functions

## Project layout (target)

```
backend/
  cmd/server/     main.go entry point
  internal/
    db/           sqlc-generated code + migrations
    handler/      HTTP handlers
  schema.sql      canonical schema
```

## Database schema (Slice 1)

```sql
CREATE TABLE readings (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    sensor_id     TEXT    NOT NULL,
    ts            INTEGER NOT NULL,  -- Unix seconds, UTC
    temperature   REAL,
    humidity      REAL,
    soil_moisture REAL
);
```

## API endpoints

| Method | Path                                         | Notes                                                              |
|--------|----------------------------------------------|--------------------------------------------------------------------|
| POST   | `/readings`                                  | JSON body, insert one row                                          |
| GET    | `/readings?limit=N`                          | Return last N rows, desc ts                                        |
| GET    | `/sensors`                                   | Distinct sensor IDs as JSON array                                  |
| GET    | `/sensors/{id}/readings`                     | `from` (unix, req), `to` (unix, req), `limit` (1–1000, def 100)   |
| GET    | `/sensors/{id}/readings/aggregated`          | `from`, `to`, `granularity` (`hour`\|`day`) — all required        |

## Conventions

- Return `application/json`; errors as `{"error": "..."}`.
- No global state — pass dependencies explicitly.
- `go vet` and `go test ./...` must pass before committing.
