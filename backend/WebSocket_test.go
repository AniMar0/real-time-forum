package backend

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	_ "modernc.org/sqlite"
	"real-time-forum/backend/account"
	"real-time-forum/backend/chat"
	"real-time-forum/backend/notification"
)

func TestWebSocketMessageIsPersistedAndDeliveredToBothSessions(t *testing.T) {
	db, err := sql.Open("sqlite", "file:websocket-test?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := runMigrations(db); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`
		INSERT INTO users (nickname, first_name, last_name, email, password, age, gender)
		VALUES ('alice', 'Alice', 'Test', 'alice@example.com', 'hash', 30, 'female'),
		       ('bob', 'Bob', 'Test', 'bob@example.com', 'hash', 30, 'male')`); err != nil {
		t.Fatal(err)
	}

	sessions := account.NewSessionRepository(db)
	if err := sessions.Create("alice-session", "alice", time.Now().Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := sessions.Create("bob-session", "bob", time.Now().Add(time.Hour)); err != nil {
		t.Fatal(err)
	}

	server := &Server{
		db:            db,
		sessions:      sessions,
		hub:           NewHub(),
		chat:          chat.NewRepository(db),
		notifications: notification.NewRepository(db),
		config:        Config{AllowedWSOrigins: []string{"http://example.test"}},
	}
	server.chatService = chat.NewService(db, server.chat, server.notifications)
	server.initUpgrader()

	mux := http.NewServeMux()
	mux.Handle("/ws", server.SessionMiddleware(http.HandlerFunc(server.HandleWebSocket)))
	httpServer := httptest.NewServer(mux)
	defer httpServer.Close()

	aliceConn := dialWebSocketTestClient(t, httpServer.URL, "alice-session")
	defer aliceConn.Close()
	bobConn := dialWebSocketTestClient(t, httpServer.URL, "bob-session")
	defer bobConn.Close()
	drainWebSocketEvents(t, aliceConn, 2)
	drainWebSocketEvents(t, bobConn, 1)

	if err := aliceConn.WriteJSON(Message{To: "bob", Content: "hello", Type: "chat_message"}); err != nil {
		t.Fatal(err)
	}

	aliceMessage := readChatMessage(t, aliceConn)
	bobMessage := readChatMessage(t, bobConn)
	if aliceMessage.ID == 0 || aliceMessage.ID != bobMessage.ID {
		t.Fatalf("got mismatched message IDs: alice=%d bob=%d", aliceMessage.ID, bobMessage.ID)
	}
	if aliceMessage.From != "alice" || bobMessage.To != "bob" || aliceMessage.Content != "hello" {
		t.Fatalf("unexpected delivered messages: alice=%#v bob=%#v", aliceMessage, bobMessage)
	}

	var messageCount, unread int
	if err := db.QueryRow("SELECT COUNT(*) FROM messages WHERE sender_id = 1 AND receiver_id = 2").Scan(&messageCount); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow("SELECT unread_messages FROM notifications WHERE receiver_id = 2 AND sender_id = 1").Scan(&unread); err != nil {
		t.Fatal(err)
	}
	if messageCount != 1 || unread != 1 {
		t.Fatalf("got messageCount=%d unread=%d, want 1 and 1", messageCount, unread)
	}
}

func dialWebSocketTestClient(t *testing.T, serverURL, sessionID string) *websocket.Conn {
	t.Helper()
	wsURL := "ws" + strings.TrimPrefix(serverURL, "http") + "/ws"
	connection, _, err := websocket.DefaultDialer.Dial(wsURL, http.Header{
		"Cookie": []string{"session_token=" + sessionID},
	})
	if err != nil {
		t.Fatalf("WebSocket dial failed: %v", err)
	}
	return connection
}

func drainWebSocketEvents(t *testing.T, connection *websocket.Conn, count int) {
	t.Helper()
	for i := 0; i < count; i++ {
		var event map[string]interface{}
		if err := connection.ReadJSON(&event); err != nil {
			t.Fatalf("failed to drain WebSocket event: %v", err)
		}
	}
	_ = connection.SetReadDeadline(time.Now().Add(3 * time.Second))
}

func readChatMessage(t *testing.T, connection *websocket.Conn) Message {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	_ = connection.SetReadDeadline(deadline)
	for {
		var event Message
		if err := connection.ReadJSON(&event); err != nil {
			t.Fatalf("failed to read WebSocket message: %v", err)
		}
		if event.Type == "chat_message" {
			return event
		}
	}
}
