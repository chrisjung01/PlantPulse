package db

import (
	_ "embed"
	"database/sql"
)

//go:embed schema.sql
var schema string

func ApplySchema(db *sql.DB) error {
	_, err := db.Exec(schema)
	return err
}
