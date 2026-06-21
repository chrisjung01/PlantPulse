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
