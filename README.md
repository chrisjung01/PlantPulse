# PlantPulse

Self-hosted, vendor-independent plant monitoring platform. ESP32 sensors report environmental data (temperature, humidity, soil moisture) to a Go backend; a SvelteKit frontend visualises it.

> **Disclaimer:** This software is provided "as is", without warranty of any kind. Use at your own risk — see [LICENSE](LICENSE) for full terms.

## Features

- **No cloud dependency** — runs entirely on your own hardware
- **ESP32 ready** — sensors POST readings directly over HTTP
- **Simple data model** — flat `readings` table, no over-engineering
- **SvelteKit dashboard** — live view of your plant data

## Tech Stack

| Layer    | Choice                                      |
|----------|---------------------------------------------|
| Backend  | Go, `net/http`, `modernc.org/sqlite`, sqlc  |
| Frontend | SvelteKit + TypeScript                      |
| Database | SQLite (embedded, single file)              |
| Sensors  | ESP32 → HTTP POST `/readings`               |

## Getting Started

### Prerequisites

- Go 1.22+
- Node.js 18+

### Run locally

```bash
# Backend (port 8080)
cd backend && go run ./cmd/server/

# Frontend (port 5173)
cd frontend && npm install && npm run dev
```

### Send a test reading

```bash
curl -X POST http://localhost:8080/readings \
  -H "Content-Type: application/json" \
  -d '{
    "sensor_id": "esp32-balcony",
    "recorded_at": 1750000000,
    "temperature_celsius": 22.4,
    "humidity_percent": 61.0,
    "soil_moisture_percent": 38.5
  }'
```

See [`backend/README.md`](backend/README.md) for the full API reference.

## License

[MIT](LICENSE) — free to use, modify, and self-host.