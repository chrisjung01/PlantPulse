# PlantPulse Backend

Go HTTP API server that receives sensor readings from ESP32 devices and exposes them to the SvelteKit frontend. Data is stored in an embedded SQLite database — no external dependencies required.

## Prerequisites

- Go 1.22 or later

## Running the server

```bash
cd backend
go run ./cmd/server
```

The server starts on port **8080**.

## Environment variables

| Variable         | Default                    | Description                                      |
|------------------|----------------------------|--------------------------------------------------|
| `DATABASE_PATH`  | `./plantpulse.db`          | Path to the SQLite database file                 |
| `ALLOWED_ORIGIN` | `http://localhost:5173`    | Origin allowed by the CORS middleware            |

Example with custom values:

```bash
DATABASE_PATH=/data/plant.db ALLOWED_ORIGIN=https://plantpulse.example.com go run ./cmd/server
```

## Endpoints

### GET /health

Returns 200 OK when the server is up.

```bash
curl http://localhost:8080/health
```

---

### POST /readings

Ingest a sensor reading. All fields are required.

```bash
curl -X POST http://localhost:8080/readings \
  -H "Content-Type: application/json" \
  -d '{
    "sensor_id": "esp32-greenhouse-01",
    "temperature": 22.5,
    "humidity": 58.3,
    "soil_moisture": 41.0,
    "light_lux": 3200
  }'
```

---

### GET /readings

Return the most recent readings, newest first.

```bash
# Default limit
curl "http://localhost:8080/readings"

# Custom limit
curl "http://localhost:8080/readings?limit=50"
```

## Running tests

```bash
cd backend
go test ./...
```

Expected output:

```
ok  	plantpulse/backend/internal/handler
```
