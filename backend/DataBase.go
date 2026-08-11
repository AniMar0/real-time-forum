package backend

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func MakeDataBase() {
	MakeDataBaseAt("database/forum.db")
}

func MakeDataBaseAt(databasePath string) {
	if err := os.MkdirAll(filepath.Dir(databasePath), os.ModePerm); err != nil {
		log.Fatalf("Failed to create database directory: %v", err)
	}

	db, err := sql.Open("sqlite", databasePath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		log.Fatalf("Failed to enable foreign keys: %v", err)
	}

	if err := runMigrations(db); err != nil {
		log.Fatalf("Failed to run database migrations: %v", err)
	}

	fmt.Println("Database and migrations applied successfully!")
}
