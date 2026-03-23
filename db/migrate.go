package db

import (
	"database/sql"
	"embed"
	"fmt"
	"sort"
	"strings"
)

//go:embed migrations
var migrationsFS embed.FS

// Migrate applies all pending SQL migrations in order.
// It creates a schema_migrations table to track which migrations have been applied.
// Migrations are applied inside individual transactions so a failure rolls back only
// the failing migration, leaving previously applied ones intact.
func Migrate(sqlDB *sql.DB) error {
	if err := ensureMigrationsTable(sqlDB); err != nil {
		return fmt.Errorf("migrations table: %w", err)
	}

	files, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	// Sort by filename (001_init.sql, 002_jobs.sql, …)
	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})

	for _, f := range files {
		if f.IsDir() || !strings.HasSuffix(f.Name(), ".sql") {
			continue
		}

		name := f.Name()
		applied, err := isMigrationApplied(sqlDB, name)
		if err != nil {
			return fmt.Errorf("check migration %s: %w", name, err)
		}
		if applied {
			continue
		}

		content, err := migrationsFS.ReadFile("migrations/" + name)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}

		if err := applyMigration(sqlDB, name, string(content)); err != nil {
			return fmt.Errorf("apply migration %s: %w", name, err)
		}
	}
	return nil
}

// ensureMigrationsTable creates the schema_migrations tracking table if needed.
func ensureMigrationsTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			name       TEXT PRIMARY KEY,
			applied_at INTEGER NOT NULL DEFAULT (unixepoch())
		)
	`)
	return err
}

// isMigrationApplied returns true if the migration has already been applied.
func isMigrationApplied(db *sql.DB, name string) (bool, error) {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM schema_migrations WHERE name = ?`, name).Scan(&count)
	return count > 0, err
}

// applyMigration executes the SQL content inside a transaction and records it.
func applyMigration(db *sql.DB, name, content string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// Execute each statement separately (SQLite driver does not support multi-statement exec)
	statements := splitStatements(content)
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, execErr := tx.Exec(stmt); execErr != nil {
			// ALTER TABLE ADD COLUMN is not idempotent in SQLite.
			// If the column already exists (schema.sql was updated), ignore the error.
			if strings.Contains(strings.ToUpper(stmt), "ALTER TABLE") &&
				strings.Contains(strings.ToUpper(stmt), "ADD COLUMN") &&
				strings.Contains(execErr.Error(), "duplicate column") {
				continue
			}
			err = execErr
			return fmt.Errorf("statement failed: %w\nSQL: %s", err, stmt[:min(200, len(stmt))])
		}
	}

	if _, err = tx.Exec(`INSERT INTO schema_migrations (name) VALUES (?)`, name); err != nil {
		return fmt.Errorf("record migration: %w", err)
	}

	return tx.Commit()
}

// splitStatements splits a SQL file into individual statements by semicolon.
// It ignores semicolons inside single-line comments.
func splitStatements(sql string) []string {
	var stmts []string
	var current strings.Builder

	lines := strings.Split(sql, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Skip pure comment lines
		if strings.HasPrefix(trimmed, "--") {
			continue
		}
		current.WriteString(line)
		current.WriteByte('\n')
		if strings.Contains(trimmed, ";") {
			stmt := strings.TrimSpace(current.String())
			if stmt != "" {
				stmts = append(stmts, stmt)
			}
			current.Reset()
		}
	}
	// Flush any remaining content
	if remaining := strings.TrimSpace(current.String()); remaining != "" {
		stmts = append(stmts, remaining)
	}
	return stmts
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
