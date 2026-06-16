package db

import (
	_ "embed"
	"database/sql"
)

// schema embeds the SQL file at compile time so the binary is self-contained —
// no need to ship schema.sql alongside the executable.
//
//go:embed schema.sql
var schema string

// ApplySchema runs the embedded schema against the database. Safe to call on
// every startup because all statements use IF NOT EXISTS.
func ApplySchema(db *sql.DB) error {
	_, err := db.Exec(schema)
	return err
}
