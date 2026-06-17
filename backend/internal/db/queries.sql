-- name: InsertReading :one
INSERT INTO readings (sensor_id, recorded_at, temperature_celsius, humidity_percent, soil_moisture_percent)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: ListReadings :many
SELECT * FROM readings ORDER BY recorded_at DESC LIMIT ?;

-- name: ListSensorIDs :many
SELECT DISTINCT sensor_id FROM readings ORDER BY sensor_id;

-- name: ListReadingsBySensor :many
SELECT id, sensor_id, recorded_at, temperature_celsius, humidity_percent, soil_moisture_percent
FROM readings
WHERE sensor_id = @sensor_id
  AND recorded_at >= @from
  AND recorded_at <= @to
ORDER BY recorded_at ASC
LIMIT @limit;

-- name: AggregateReadingsBySensorByHour :many
SELECT
  strftime('%Y-%m-%dT%H:00:00Z', datetime(recorded_at, 'unixepoch')) AS bucket,
  AVG(temperature_celsius)   AS temperature_celsius,
  AVG(humidity_percent)      AS humidity_percent,
  AVG(soil_moisture_percent) AS soil_moisture_percent,
  COUNT(*)                   AS sample_count
FROM readings
WHERE sensor_id = @sensor_id
  AND recorded_at >= @from
  AND recorded_at <= @to
GROUP BY bucket
ORDER BY bucket ASC;

-- name: AggregateReadingsBySensorByDay :many
SELECT
  strftime('%Y-%m-%dT00:00:00Z', datetime(recorded_at, 'unixepoch')) AS bucket,
  AVG(temperature_celsius)   AS temperature_celsius,
  AVG(humidity_percent)      AS humidity_percent,
  AVG(soil_moisture_percent) AS soil_moisture_percent,
  COUNT(*)                   AS sample_count
FROM readings
WHERE sensor_id = @sensor_id
  AND recorded_at >= @from
  AND recorded_at <= @to
GROUP BY bucket
ORDER BY bucket ASC;
