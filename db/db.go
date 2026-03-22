package db

import (
	"database/sql"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var schemaFS embed.FS

// DB wraps *sql.DB with FRACTURE-specific helpers.
type DB struct {
	*sql.DB
}

// DataDir returns the platform-appropriate data directory for FRACTURE.
func DataDir() (string, error) {
	var base string
	switch runtime.GOOS {
	case "windows":
		base = os.Getenv("APPDATA")
		if base == "" {
			return "", fmt.Errorf("APPDATA not set")
		}
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, "Library", "Application Support")
	default: // linux and others
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, ".local", "share")
	}
	dir := filepath.Join(base, "FRACTURE")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	return dir, nil
}

// Open opens (or creates) the FRACTURE SQLite database and applies the schema.
func Open() (*DB, error) {
	dir, err := DataDir()
	if err != nil {
		return nil, fmt.Errorf("data dir: %w", err)
	}

	dbPath := filepath.Join(dir, "data.db")
	sqlDB, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	// Apply schema (idempotent — uses CREATE TABLE IF NOT EXISTS)
	schema, err := schemaFS.ReadFile("schema.sql")
	if err != nil {
		return nil, fmt.Errorf("read schema: %w", err)
	}
	if _, err := sqlDB.Exec(string(schema)); err != nil {
		return nil, fmt.Errorf("apply schema: %w", err)
	}

	return &DB{sqlDB}, nil
}

// GetConfig retrieves a config value by key.
func (d *DB) GetConfig(key string) (string, error) {
	var value string
	err := d.QueryRow(`SELECT value FROM config WHERE key = ?`, key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}

// SetConfig upserts a config key-value pair.
func (d *DB) SetConfig(key, value string) error {
	_, err := d.Exec(`
		INSERT INTO config (key, value, updated_at)
		VALUES (?, ?, unixepoch())
		ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = unixepoch()
	`, key, value)
	return err
}

// IsOnboarded returns true if the user has completed onboarding.
func (d *DB) IsOnboarded() (bool, error) {
	val, err := d.GetConfig("onboarding_complete")
	if err != nil {
		return false, err
	}
	return val == "true", nil
}
