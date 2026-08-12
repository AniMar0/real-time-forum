package backend

import (
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
	"real-time-forum/backend/chat"
	"real-time-forum/backend/notification"
)

func TestChatServicePersistsMessageAndUnreadNotification(t *testing.T) {
	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "chat.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE users (id INTEGER PRIMARY KEY, nickname TEXT UNIQUE);
		CREATE TABLE messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			sender_id INTEGER,
			receiver_id INTEGER,
			sender TEXT,
			receiver TEXT,
			content TEXT,
			timestamp DATETIME
		);
		CREATE TABLE notifications (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			receiver_id INTEGER,
			sender_id INTEGER,
			receiver_nickname TEXT,
			sender_nickname TEXT,
			unread_messages INTEGER
		);
		INSERT INTO users (id, nickname) VALUES (1, 'alice'), (2, 'bob');
	`)
	if err != nil {
		t.Fatal(err)
	}

	service := chat.NewService(
		db,
		chat.NewRepository(db),
		notification.NewRepository(db),
	)
	message, err := service.SendMessage(1, "bob", "<b>hello</b>")
	if err != nil {
		t.Fatalf("SendMessage failed: %v", err)
	}
	if message.ID == 0 || message.From != "alice" || message.To != "bob" || message.Content != "<b>hello</b>" {
		t.Fatalf("unexpected message: %#v", message)
	}

	var storedCount, unread int
	var storedContent string
	if err := db.QueryRow("SELECT COUNT(*) FROM messages WHERE sender_id = 1 AND receiver_id = 2").Scan(&storedCount); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow("SELECT unread_messages FROM notifications WHERE receiver_id = 2 AND sender_id = 1").Scan(&unread); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow("SELECT content FROM messages WHERE sender_id = 1 AND receiver_id = 2").Scan(&storedContent); err != nil {
		t.Fatal(err)
	}
	if storedCount != 1 || unread != 1 || storedContent != "<b>hello</b>" {
		t.Fatalf("got storedCount=%d unread=%d content=%q, want 1, 1, raw content", storedCount, unread, storedContent)
	}
}
