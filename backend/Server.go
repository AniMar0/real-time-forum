package backend

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/twinj/uuid"
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
	users         *account.UserRepository
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
	S.users = account.NewUserRepository(S.db)
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

	log.Println("Server running on " + config.HTTPAddress)
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

	S.Mux.Handle("/logout", S.SessionMiddleware(http.HandlerFunc(S.LogoutHandler)))
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

// Modified HandleWebSocket function - broadcasts status changes when user connects
func (S *Server) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	identity, err := S.CheckSessionIdentity(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := S.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}

	client := &Client{
		ID:        uuid.NewV4().String(),
		Conn:      conn,
		Username:  identity.Nickname,
		UserID:    identity.UserID,
		SessionID: identity.SessionID,
		Send:      make(chan interface{}, 10),
	}

	S.hub.Register(client)

	log.Printf("user %s connected to WebSocket", identity.Nickname)

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
			log.Printf("WebSocket read error: %v", err)
			break
		}

		s.handleWebSocketMessage(client, msg)
	}
}

func (s *Server) handleWebSocketMessage(client *Client, msg Message) {
	switch msg.Type {
	case "typing_indicator":
		msg.From = client.Username
		s.sendTypingIndicator(msg)
	case "chat_message":
		s.handleChatMessage(client, msg)
	}
}

func (s *Server) handleChatMessage(client *Client, msg Message) {
	if strings.TrimSpace(msg.Content) == "" {
		log.Printf("ignored empty chat message from user %d", client.UserID)
		return
	}
	if s.chatService == nil {
		log.Printf("chat service is not initialized")
		return
	}

	storedMessage, err := s.chatService.SendMessage(client.UserID, msg.To, msg.Content)
	if err != nil {
		log.Printf("failed to persist WebSocket message: %v", err)
		return
	}
	s.broadcastUserStatusChange()
	s.sendMessageToRecipient(Message{
		ID:        storedMessage.ID,
		From:      storedMessage.From,
		To:        storedMessage.To,
		Content:   storedMessage.Content,
		Timestamp: storedMessage.Timestamp,
		Type:      "chat_message",
	}, storedMessage.ReceiverID, storedMessage.SenderID)
}

func (s *Server) sendMessageToRecipient(msg Message, recipientID, senderID int64) {
	for _, recipient := range s.hub.ClientsForUser(recipientID) {
		recipient.Enqueue(msg)
	}
	for _, senderClient := range s.hub.ClientsForUser(senderID) {
		senderClient.Enqueue(msg)
	}
}

func (s *Server) sendTypingIndicator(msg Message) {
	recipientID, err := s.chat.UserIDByNickname(msg.To)
	if err != nil {
		return
	}
	for _, recipient := range s.hub.ClientsForUser(recipientID) {
		recipient.Enqueue(msg)
	}
}

func (S *Server) broadcastUserList(currentUserID int64) {
	conversations, err := S.chat.ListConversations(currentUserID)
	if err != nil {
		return
	}

	users := make([]UsersListe, 0, len(conversations))
	for _, conversation := range conversations {
		status := "offline"
		if S.hub.IsOnline(conversation.UserID) {
			status = "online"
		}
		users = append(users, UsersListe{Nickname: conversation.Nickname, Status: status})
	}

	for _, client := range S.hub.ClientsForUser(currentUserID) {
		client.Enqueue(map[string]interface{}{
			"type":  "user_list",
			"users": users,
		})
	}
}

func (s *Server) removeClient(client *Client) {
	s.hub.Unregister(client)

	log.Printf("user %s disconnected", client.Username)

	go func() {
		time.Sleep(100 * time.Millisecond)
		s.broadcastUserStatusChange()
	}()
}

func (S *Server) broadcastUserStatusChange() {
	for _, userID := range S.hub.UserIDs() {
		S.broadcastUserList(userID)
	}
}

func StartWriter(c *Client) {
	// Issue #8: Add panic recovery and proper error handling
	defer func() {
		if r := recover(); r != nil {
			log.Printf("writer panic recovered: %v", r)
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
