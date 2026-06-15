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

## API endpoints (Slice 1)

| Method | Path                      | Notes                        |
|--------|---------------------------|------------------------------|
| POST   | `/readings`               | JSON body, insert one row    |
| GET    | `/readings?limit=N`       | Return last N rows, desc ts  |

## Conventions

- Return `application/json`; errors as `{"error": "..."}`.
- No global state — pass dependencies explicitly.
- `go vet` and `go test ./...` must pass before committing.
