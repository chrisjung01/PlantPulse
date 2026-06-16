CREATE TABLE IF NOT EXISTS readings (
    id                    INTEGER PRIMARY KEY AUTOINCREMENT,
    sensor_id             TEXT    NOT NULL,
    recorded_at           INTEGER NOT NULL,
    temperature_celsius   REAL,
    humidity_percent      REAL,
    soil_moisture_percent REAL
);
