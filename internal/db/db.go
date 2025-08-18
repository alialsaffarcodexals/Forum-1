package db

import (
	"database/sql"
	"embed"
	"io/fs"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var schemaFS embed.FS

func Open(path string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	if err := migrate(db); err != nil {
		return nil, err
	}
	return db, nil
}

func migrate(db *sql.DB) error {
	sqlBytes, err := fs.ReadFile(schemaFS, "schema.sql")
	if err != nil {
		return err
	}
	if _, err := db.Exec(string(sqlBytes)); err != nil {
		return err
	}
	return nil
}
