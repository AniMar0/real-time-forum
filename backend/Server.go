package backend

import (
	"context"
	"database/sql"
	"fmt"
	"html"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/twinj/uuid"
	"golang.org/x/crypto/bcrypt"
	"real-time-forum/backend/account"
	"real-time-forum/backend/chat"
	"real-time-forum/backend/forum"
	"real-time-forum/backend/notification"
)

type Server struct {
	db            *sql.DB
	Mux           *http.ServeMux
	hub           *Hub
	config        Config
	httpServer    *http.Server
	sessions      *account.SessionRepository
	forum         *forum.Repository
	chat          *chat.Repository
	chatService   *chat.Service
	notifications *notification.Repository
	upgrader      websocket.Upgrader
}

const (
	webSocketWriteWait = 10 * time.Second
	webSocketPongWait  = 60 * time.Second
	webSocketPingEvery = (webSocketPongWait * 9) / 10
	webSocketMaxSize   = 8 * 1024
)

func (S *Server) initUpgrader() {
	S.upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			// Allow localhost and production domains
			origin := r.Header.Get("Origin")
			if origin == "" {
				return true // Allow connections without Origin header (like from same origin)
			}
			// Add your production domain here if needed
			for _, allowed := range S.config.AllowedWSOrigins {
				if origin == allowed {
					return true
				}
			}
			return false
		},
	}
}

func (S *Server) Run(port string) {
	config := LoadConfig()
	config.HTTPAddress = ":" + port
	S.RunWithConfig(config)
}

func (S *Server) RunWithConfig(config Config) {
	S.Mux = http.NewServeMux()
	S.config = config

	var err error
	S.db, err = sql.Open("sqlite", config.DatabasePath)
	if err != nil {
		log.Fatal(err)
	}
	defer S.db.Close()
	S.sessions = account.NewSessionRepository(S.db)
	S.forum = forum.NewRepository(S.db)
	S.chat = chat.NewRepository(S.db)
	S.notifications = notification.NewRepository(S.db)
	S.chatService = chat.NewService(S.db, S.chat, S.notifications)

	// Initialize WebSocket upgrader with CORS protection
	S.initUpgrader()

	S.initRoutes()

	S.hub = NewHub()

	S.httpServer = &http.Server{
		Addr:              config.HTTPAddress,
		Handler:           S.Mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	fmt.Println("Server running on " + config.HTTPAddress)
	err = S.httpServer.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Println("Server error:", err)
		return
	}
}

func (S *Server) Shutdown(ctx context.Context) error {
	if S.httpServer == nil {
		return nil
	}
	return S.httpServer.Shutdown(ctx)
}

func (S *Server) initRoutes() {
	home := http.FileServer(http.Dir(S.config.StaticPath))
	S.Mux.Handle("/", checkHome(home))
	S.Mux.HandleFunc("/logged", S.LoggedHandler)

	S.Mux.HandleFunc("/notifications", S.GetNotifications)
	S.Mux.HandleFunc("/notifications/mark-read", S.MarkNotificationsRead)

	S.Mux.Handle("/createPost", S.SessionMiddleware(http.HandlerFunc(S.CreatePostHandler)))
	S.Mux.Handle("/posts", S.SessionMiddleware(http.HandlerFunc(S.GetPostsHandler)))

	S.Mux.Handle("/createComment", S.SessionMiddleware(http.HandlerFunc(S.CreateCommentHandler)))
	S.Mux.Handle("/comments", S.SessionMiddleware(http.HandlerFunc(S.GetCommentsHandler)))

	S.Mux.HandleFunc("/register", S.RegisterHandler)
	S.Mux.HandleFunc("/login", S.LoginHandler)

	S.Mux.Handle("/ws", S.SessionMiddleware(http.HandlerFunc(S.HandleWebSocket)))
	S.Mux.Handle("/messages", S.SessionMiddleware(http.HandlerFunc(S.GetMessagesHandler)))

	S.Mux.Handle("/sendMessage", S.SessionMiddleware(http.HandlerFunc(S.SendMessageHandler)))
	S.Mux.Handle("/logout", S.SessionMiddleware(http.HandlerFunc(S.LogoutHandler)))
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
		identity, err := S.CheckSessionIdentity(r)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := account.WithIdentity(r.Context(), identity)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (S *Server) CheckSession(r *http.Request) (string, string, error) {
	identity, err := S.CheckSessionIdentity(r)
	if err != nil {
		return "", "", err
	}
	return identity.Nickname, identity.SessionID, nil
}

func (S *Server) CheckSessionIdentity(r *http.Request) (account.Identity, error) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return account.Identity{}, fmt.Errorf("no session cookie")
	}
	sessionID := cookie.Value
	if S.sessions == nil {
		return account.Identity{}, fmt.Errorf("session repository is not initialized")
	}
	identity, err := S.sessions.FindValid(sessionID)
	if err != nil {
		return account.Identity{}, fmt.Errorf("invalid or expired session")
	}
	return identity, nil
}

func (S *Server) MakeToken(Writer http.ResponseWriter, username string) {
	sessionID := uuid.NewV4().String()
	expirationTime := time.Now().Add(24 * time.Hour)

	if S.sessions == nil {
		http.Error(Writer, "Session repository is not initialized", http.StatusInternalServerError)
		return
	}
	err := S.sessions.Create(sessionID, username, expirationTime)
	if err != nil {
		http.Error(Writer, "Error creating session", http.StatusInternalServerError)
		return
	}

	http.SetCookie(Writer, &http.Cookie{
		Name:     "session_token",
		Value:    sessionID,
		Expires:  expirationTime,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   S.config.SecureCookies(),
	})
}

func (S *Server) GetHashedPasswordFromDB(identifier string) (string, string, error) {
	var hashedPassword, nickname string

	// Issue #3: Fix SQL Scan order - must match SELECT order
	err := S.db.QueryRow(`
		SELECT password, nickname FROM users 
		WHERE nickname = ? OR email = ?
	`, identifier, identifier).Scan(&hashedPassword, &nickname)
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
		Send:      make(chan interface{}, 10),
	}

	S.hub.Register(client)

	fmt.Println(username, "connected to WebSocket")

	S.broadcastUserStatusChange()

	go StartWriter(client)
	go S.receiveMessages(client)
}

// Modified receiveMessages function
func (s *Server) receiveMessages(client *Client) {
	defer s.removeClient(client)

	client.Conn.SetReadLimit(webSocketMaxSize)
	_ = client.Conn.SetReadDeadline(time.Now().Add(webSocketPongWait))
	client.Conn.SetPongHandler(func(string) error {
		return client.Conn.SetReadDeadline(time.Now().Add(webSocketPongWait))
	})

	for {
		var msg Message
		err := client.Conn.ReadJSON(&msg)
		if err != nil {
			fmt.Println("WebSocket Read Error:", err)
			break
		}

		if msg.Type == "typing_indicator" {
			msg.From = client.Username
			s.sendTypingIndicator(msg)

		} else if msg.Type == "chat_message" {
			if strings.TrimSpace(msg.Content) == "" {
				fmt.Println("you cant't send an empty message")
				continue
			}

			msg.From = client.Username
			if s.chatService == nil {
				continue
			}
			storedMessage, err := s.chatService.SendMessage(msg.From, msg.To, msg.Content)
			if err != nil {
				fmt.Println("Failed to persist WebSocket message:", err)
				continue
			}
			s.broadcastUserStatusChange()
			s.sendMessageToRecipient(Message{
				ID:        storedMessage.ID,
				From:      storedMessage.From,
				To:        storedMessage.To,
				Content:   storedMessage.Content,
				Timestamp: storedMessage.Timestamp,
				Type:      "chat_message",
			})
		}

	}
}

func (s *Server) sendMessageToRecipient(msg Message) {
	for _, recipient := range s.hub.ClientsForUser(msg.To) {
		recipient.Enqueue(msg)
	}
	for _, senderClient := range s.hub.ClientsForUser(msg.From) {
		senderClient.Enqueue(msg)
	}
}

func (s *Server) sendTypingIndicator(msg Message) {
	for _, recipient := range s.hub.ClientsForUser(msg.To) {
		recipient.Enqueue(msg)
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
	)
	SELECT 
	    u.id, 
	    u.nickname, 
	    COALESCE(u.content, ""), 
	    u.lastInteraction
	FROM cte_ordered_users u
	ORDER BY u.lastInteraction DESC, u.nickname;
	`

	rows, err := S.db.Query(query,
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
		if err := rows.Scan(&uc.ID, &uc.Nickname, &uc.LastMessage, &uc.LastInteraction); err != nil {
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
		if S.hub.IsOnline(user) {
			Users = append(Users, UsersListe{Nickname: user, Status: "online"})
		} else {
			Users = append(Users, UsersListe{Nickname: user, Status: "offline"})
		}
	}

	// Send to all client sessions
	for _, client := range S.hub.ClientsForUser(currentUser) {
		client.Enqueue(map[string]interface{}{
			"type":  "user_list",
			"users": Users,
		})
	}
}

func (s *Server) removeClient(client *Client) {
	s.hub.Unregister(client)

	fmt.Println(client.Username, "disconnected")

	go func() {
		time.Sleep(100 * time.Millisecond)
		s.broadcastUserStatusChange()
	}()
}

func (S *Server) broadcastUserStatusChange() {
	for _, username := range S.hub.Usernames() {
		S.broadcastUserList(username)
	}
}

func StartWriter(c *Client) {
	// Issue #8: Add panic recovery and proper error handling
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Writer panic recovered:", r)
		}
	}()

	ticker := time.NewTicker(webSocketPingEvery)
	defer ticker.Stop()

	for {
		select {
		case msg, ok := <-c.Send:
			if !ok {
				return
			}
			if err := c.Conn.SetWriteDeadline(time.Now().Add(webSocketWriteWait)); err != nil {
				return
			}
			if err := c.Conn.WriteJSON(msg); err != nil {
				c.Close()
				return
			}
		case <-ticker.C:
			if err := c.Conn.SetWriteDeadline(time.Now().Add(webSocketWriteWait)); err != nil {
				c.Close()
				return
			}
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.Close()
				return
			}
		}
	}
}
