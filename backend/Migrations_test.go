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
	if count != 4 {
		t.Fatalf("got %d applied migrations, want 4", count)
	}
}

func TestValidateIdentityBackfillPassesCompleteSchema(t *testing.T) {
	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "forum.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err := runMigrations(db); err != nil {
		t.Fatal(err)
	}
	if err := validateIdentityBackfill(db); err != nil {
		t.Fatalf("expected complete migrated schema to validate: %v", err)
	}
}
