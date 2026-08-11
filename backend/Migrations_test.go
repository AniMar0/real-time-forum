package backend

import (
	"database/sql"
	"path/filepath"
	"testing"
)

func TestRunMigrationsIsIdempotent(t *testing.T) {
	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "forum.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err := runMigrations(db); err != nil {
		t.Fatalf("first migration run failed: %v", err)
	}
	if err := runMigrations(db); err != nil {
		t.Fatalf("second migration run failed: %v", err)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 3 {
		t.Fatalf("got %d applied migrations, want 3", count)
	}
}

func TestValidateIdentityBackfillRejectsIncompleteRows(t *testing.T) {
	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "forum.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err := runMigrations(db); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`
		INSERT INTO users (id, nickname) VALUES (1, 'alice');
		INSERT INTO sessions (session_id, nickname, user_id, expires_at)
		VALUES ('invalid-session', 'alice', NULL, CURRENT_TIMESTAMP);`); err != nil {
		t.Fatal(err)
	}
	if err := validateIdentityBackfill(db); err == nil {
		t.Fatal("expected incomplete identity rows to be rejected")
	}
}
