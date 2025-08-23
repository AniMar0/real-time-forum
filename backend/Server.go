package backend

import (
	"context"
	"database/sql"
	"fmt"
	"html"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/twinj/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Server struct {
	db       *sql.DB
	Mux      *http.ServeMux
	clients  map[string][]*Client // Changed: map username to slice of clients
	upgrader websocket.Upgrader
	sync.RWMutex
}

func (S *Server) Run(port string) {
	S.Mux = http.NewServeMux()

	var err error
	S.db, err = sql.Open("sqlite3", "database/forum.db")
	if err != nil {
		log.Fatal(err)
	}
	defer S.db.Close()

	S.initRoutes()

	S.clients = make(map[string][]*Client) // Updated initialization

	fmt.Println("Server running on http://localhost:" + port)
	err = http.ListenAndServe(":"+port, S.Mux)
	if err != nil {
		log.Println("Server error:", err)
		return
	}
}

func (S *Server) initRoutes() {
	home := http.FileServer(http.Dir("./static"))
	S.Mux.Handle("/", checkHome(home))
	S.Mux.HandleFunc("/logged", S.LoggedHandler)

	S.Mux.HandleFunc("/notification", S.Notification)

	S.Mux.Handle("/createPost", S.SessionMiddleware(http.HandlerFunc(S.CreatePostHandler)))
	S.Mux.HandleFunc("/posts", S.GetPostsHandler)

	S.Mux.Handle("/createComment", S.SessionMiddleware(http.HandlerFunc(S.CreateCommentHandler)))
	S.Mux.HandleFunc("/comments", S.GetCommentsHandler)

	S.Mux.HandleFunc("/register", S.RegisterHandler)
	S.Mux.HandleFunc("/login", S.LoginHandler)

	S.Mux.HandleFunc("/ws", S.HandleWebSocket)
	S.Mux.HandleFunc("/messages", S.GetMessagesHandler)

	S.Mux.HandleFunc("/logout", S.LogoutHandler)
}

func (S *Server) UserFound(user User) (error, bool) {
	var exists int
	err := S.db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ? OR nickname = ?", user.Email, user.Nickname).Scan(&exists)
	if err != nil {
		return err, false
	}
	if exists > 0 {
		return nil, true
	}
	return nil, false
}

func (S *Server) AddUser(user User) string {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return "hash Password Error"
	}
	query := `INSERT INTO users (nickname, first_name, last_name, email, password, age, gender)
		VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err = S.db.Exec(query, html.EscapeString(user.Nickname), html.EscapeString(user.FirstName), html.EscapeString(user.LastName), html.EscapeString(user.Email), string(hashedPassword), user.Age, user.Gender)
	if err != nil {
		return error.Error(err)
	}
	return ""
}

func (S *Server) SessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, _, err := S.CheckSession(r)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "username", username)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (S *Server) CheckSession(r *http.Request) (string, string, error) {

	cookie, err := r.Cookie("session_token")
	if err != nil {
		return "", "", fmt.Errorf("no session cookie")
	}
	sessionID := cookie.Value
	var username string
	err = S.db.QueryRow(`
        SELECT nickname FROM sessions 
        WHERE session_id = ? AND expires_at > CURRENT_TIMESTAMP
    `, sessionID).Scan(&username)

	if err != nil {
		return "", "", fmt.Errorf("invalid or expired session")
	}

	return username, sessionID, nil
}

func (S *Server) MakeToken(Writer http.ResponseWriter, username string) {
	sessionID := uuid.NewV4().String()
	expirationTime := time.Now().Add(24 * time.Hour)

	_, err := S.db.Exec("INSERT INTO sessions (session_id, nickname, expires_at) VALUES (?, ?, ?)",
		sessionID, username, expirationTime)
	if err != nil {
		http.Error(Writer, "Error creating session", http.StatusInternalServerError)
		return
	}

	http.SetCookie(Writer, &http.Cookie{
		Name:     "session_token",
		Value:    sessionID,
		Expires:  expirationTime,
		HttpOnly: true,
	})
}

func (S *Server) GetHashedPasswordFromDB(identifier string) (string, string, error) {
	var hashedPassword, nickname string

	err := S.db.QueryRow(`
		SELECT password, nickname FROM users 
		WHERE nickname = ? OR email = ?
	`, identifier, identifier).Scan(&nickname, &hashedPassword)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", "", fmt.Errorf("this user does not exist")
		}
		return "", "", err
	}
	return hashedPassword, nickname, nil
}

// Modified HandleWebSocket function - broadcasts status changes when user connects
func (S *Server) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	username, session_id, err := S.CheckSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := S.upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("WebSocket Upgrade Error:", err)
		return
	}

	client := &Client{
		ID:        uuid.NewV4().String(),
		Conn:      conn,
		Username:  username,
		SessionID: session_id,
	}

	S.RLock()
	if S.clients[username] == nil {
		S.clients[username] = []*Client{}
	}
	S.clients[username] = append(S.clients[username], client)
	S.RUnlock()

	fmt.Println(username, "connected to WebSocket")

	S.broadcastUserStatusChange()

	go S.receiveMessages(client)
}

// Modified receiveMessages function
func (s *Server) receiveMessages(client *Client) {
	defer s.removeClient(client)

	for {
		var msg Message
		err := client.Conn.ReadJSON(&msg)
		if err != nil {
			fmt.Println("WebSocket Read Error:", err)
			break
		}

		if strings.TrimSpace(msg.Content) == "" {
			fmt.Println("you cant't send an empty message")
			continue
		}

		msg.From = client.Username
		msg.Timestamp = time.Now().Format(time.RFC3339)

		_, err = s.db.Exec(`
			INSERT INTO messages (sender, receiver, content, timestamp)
			VALUES (?, ?, ?, ?)`,
			msg.From, msg.To, html.EscapeString(msg.Content), msg.Timestamp)
		if err != nil {
			fmt.Println("DB Insert Error:", err)
			continue
		}

		s.RLock()
		// Send to all sessions of the recipient
		if recipientSessions, ok := s.clients[msg.To]; ok {
			for _, recipient := range recipientSessions {
				s.broadcastUserStatusChange()
				err := recipient.Conn.WriteJSON(msg)
				if err != nil {
					fmt.Println("Send Error to recipient:", err)
				}
			}
		}
		s.RUnlock()

		s.RLock()
		// Send to all other sessions of the sender (excluding current session)
		if senderSessions, ok := s.clients[msg.From]; ok {
			for _, senderClient := range senderSessions {
				s.broadcastUserStatusChange()
				if senderClient.ID != client.ID { // Don't send back to the same session
					err := senderClient.Conn.WriteJSON(msg)
					if err != nil {
						fmt.Println("Send Error to sender session:", err)
					}
				}
			}
		}
		s.RUnlock()
	}
}

// Modified broadcastUserList function
func (S *Server) broadcastUserList(currentUser string) {
	query := `
	WITH 
	cte_latest_interaction AS (
	    SELECT
	        CASE 
	            WHEN sender = ? THEN receiver 
	            ELSE sender 
	        END AS user_nickname,
	        MAX(timestamp) AS lastInteraction,
	        content
	    FROM messages
	    WHERE sender = ? OR receiver = ?
	    GROUP BY user_nickname
	),
	cte_ordered_users AS (
	    SELECT 
	        i.content, 
	        COALESCE(i.lastInteraction, 0) AS lastInteraction,
	        u.id, 
	        u.nickname
	    FROM users u 
	    LEFT JOIN cte_latest_interaction i 
	        ON i.user_nickname = u.nickname
	    WHERE u.nickname != ?
	),
	cte_notifications AS (
	    SELECT 
	        sender AS sender_nickname,
	        COUNT(*) AS notifications 
	    FROM messages 
	    WHERE 1=0
	      AND receiver = ?
	    GROUP BY sender
	)
	SELECT 
	    u.id, 
	    u.nickname, 
	    COALESCE(u.content, ""), 
	    u.lastInteraction, 
	    COALESCE(n.notifications, 0)
	FROM cte_ordered_users u
	LEFT JOIN cte_notifications n ON u.nickname = n.sender_nickname
	ORDER BY u.lastInteraction DESC, u.nickname;
	`

	rows, err := S.db.Query(query,
		currentUser,
		currentUser,
		currentUser,
		currentUser,
		currentUser,
	)
	if err != nil {
		return
	}
	defer rows.Close()

	var results []UserConversation
	for rows.Next() {
		var uc UserConversation
		if err := rows.Scan(&uc.ID, &uc.Nickname, &uc.LastMessage, &uc.LastInteraction, &uc.Notifications); err != nil {
			continue
		}
		results = append(results, uc)
	}

	if err := rows.Err(); err != nil {
		return
	}

	var usernames []string
	for _, r := range results {
		usernames = append(usernames, r.Nickname)
	}
	var Users []UsersListe
	for _, user := range usernames {
		S.RLock()
		if S.clients[user] != nil {
			Users = append(Users, UsersListe{Nickname: user, Status: "online"})
		} else {
			Users = append(Users, UsersListe{Nickname: user, Status: "offline"})
		}
		S.RUnlock()
	}

	// Send to all client sessions
	S.RLock()
	for _, client := range S.clients[currentUser] {
		client.Conn.WriteJSON(map[string]interface{}{
			"type":  "user_list",
			"users": Users,
		})
	}
	S.RUnlock()
}

func (s *Server) removeClient(client *Client) {
	s.Lock()
	defer s.Unlock()

	clients, ok := s.clients[client.Username]
	if !ok {
		return
	}

	for i, c := range clients {
		if c.ID == client.ID {
			s.clients[client.Username] = append(clients[:i], clients[i+1:]...)
			break
		}
	}

	if len(s.clients[client.Username]) == 0 {
		delete(s.clients, client.Username)
	}

	fmt.Println(client.Username, "disconnected")

	go func() {
		time.Sleep(100 * time.Millisecond)
		s.broadcastUserStatusChange()
	}()
}

func (S *Server) broadcastUserStatusChange() {
	S.RLock()
	defer S.RUnlock()

	for username := range S.clients {
		S.broadcastUserList(username)
	}
}
