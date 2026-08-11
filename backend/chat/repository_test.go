package chat

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func TestListHistoryWithJoinedUsers(t *testing.T) {
	db, err := sql.Open("sqlite", "file:history-test?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE users (id INTEGER PRIMARY KEY, nickname TEXT UNIQUE);
		CREATE TABLE messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			sender_id INTEGER NOT NULL,
			receiver_id INTEGER NOT NULL,
			content TEXT NOT NULL,
			timestamp TEXT NOT NULL
		);
		INSERT INTO users (id, nickname) VALUES (1, 'User1'), (2, 'User2');
		INSERT INTO messages (sender_id, receiver_id, content, timestamp)
		VALUES (1, 2, 'first', '2026-08-11T10:00:00Z'),
		       (2, 1, 'second', '2026-08-11T10:01:00Z');
	`)
	if err != nil {
		t.Fatal(err)
	}

	messages, err := NewRepository(db).ListHistory("User1", "User2", 0, 0)
	if err != nil {
		t.Fatalf("ListHistory failed: %v", err)
	}
	if len(messages) != 2 {
		t.Fatalf("got %d messages, want 2", len(messages))
	}
	if messages[0].From != "User2" || messages[1].From != "User1" {
		t.Fatalf("unexpected message order: %#v", messages)
	}
}
