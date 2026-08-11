package backend

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

func runMigrations(db *sql.DB) error {
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`); err != nil {
		return err
	}

	entries, err := fs.ReadDir(migrationFiles, "migrations")
	if err != nil {
		return err
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		var applied int
		if err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", entry.Name()).Scan(&applied); err != nil {
			return err
		}
		if applied > 0 {
			continue
		}

		contents, err := migrationFiles.ReadFile(path.Join("migrations", entry.Name()))
		if err != nil {
			return err
		}
		tx, err := db.Begin()
		if err != nil {
			return err
		}
		if _, err := tx.Exec(string(contents)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("migration %s: %w", entry.Name(), err)
		}
		if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES (?)", entry.Name()); err != nil {
			_ = tx.Rollback()
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
	}
	return validateIdentityBackfill(db)
}

func validateIdentityBackfill(db *sql.DB) error {
	checks := []struct {
		name  string
		query string
	}{
		{
			name:  "messages sender and receiver IDs",
			query: "SELECT COUNT(*) FROM messages WHERE sender_id IS NULL OR receiver_id IS NULL",
		},
		{
			name:  "session user IDs",
			query: "SELECT COUNT(*) FROM sessions WHERE user_id IS NULL",
		},
		{
			name:  "notification sender and receiver IDs",
			query: "SELECT COUNT(*) FROM notifications WHERE sender_id IS NULL OR receiver_id IS NULL",
		},
	}

	for _, check := range checks {
		var invalid int
		if err := db.QueryRow(check.query).Scan(&invalid); err != nil {
			return fmt.Errorf("validate %s: %w", check.name, err)
		}
		if invalid > 0 {
			return fmt.Errorf("validate %s: %d invalid rows", check.name, invalid)
		}
	}
	return nil
}
