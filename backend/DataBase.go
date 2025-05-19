package backend

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func RunDataBas() {
	db, err := sql.Open("sqlite3", "database/forum.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = createTables(db)
	if err != nil {
		log.Fatal("Failed to create tables:", err)
	}

	fmt.Println("Database and tables created successfully!")
}

func createTables(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		nickname TEXT UNIQUE,
		first_name TEXT,
		last_name TEXT,
		email TEXT UNIQUE,
		password TEXT,
		age INTEGER,
		gender TEXT
	);

	CREATE TABLE IF NOT EXISTS posts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		title TEXT,
		content TEXT,
		category TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(user_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS comments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		post_id INTEGER,
		user_id INTEGER,
		content TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(post_id) REFERENCES posts(id),
		FOREIGN KEY(user_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		sender_id INTEGER,
		receiver_id INTEGER,
		content TEXT,
		sent_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(sender_id) REFERENCES users(id),
		FOREIGN KEY(receiver_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS sessions (
    session_id TEXT PRIMARY KEY,
    nickname TEXT,
    expires_at DATETIME,
    FOREIGN KEY(nickname) REFERENCES users(nickname)
	);
	`
	_, err := db.Exec(schema)
	return err
}
