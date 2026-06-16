-- name: InsertReading :one
INSERT INTO readings (sensor_id, recorded_at, temperature_celsius, humidity_percent, soil_moisture_percent)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: ListReadings :many
SELECT * FROM readings ORDER BY recorded_at DESC LIMIT ?;
