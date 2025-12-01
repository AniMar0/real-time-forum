package backend

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func (S *Server) SendMessageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/404", http.StatusSeeOther)
		return
	}

	_, _, err := S.CheckSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var message Message
	err = json.NewDecoder(r.Body).Decode(&message)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Validate message content
	if !isValidTextLength(message.Content, 1, 5000) {
		http.Error(w, "Message must be 1-5000 characters", http.StatusBadRequest)
		return
	}

	message.Timestamp = time.Now().Format(time.RFC3339)

	fmt.Println((message.Content))
	_, err = S.db.Exec(`
		INSERT INTO messages (sender, receiver, content, timestamp)
		VALUES (?, ?, ?, ?)`,
		message.From, message.To, html.EscapeString(message.Content), message.Timestamp)
	if err != nil {
		http.Error(w, "Failed to insert message", http.StatusInternalServerError)
		return
	}

	S.broadcastUserStatusChange()
	S.addNewNotification(message.To, message.From)

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(message)
}

func (S *Server) GetNotifications(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Redirect(w, r, "/404", http.StatusSeeOther)
		return
	}

	cookie, err := r.Cookie("session_token")
	if err != nil {
		http.Error(w, "Unauthorized - No session", http.StatusUnauthorized)
		return
	}

	sessionID := cookie.Value
	var nickname string
	err = S.db.QueryRow(`
		SELECT nickname FROM sessions 
		WHERE session_id = ? AND expires_at > datetime('now')`, sessionID).Scan(&nickname)
	if err != nil {
		http.Error(w, "Unauthorized - Invalid session", http.StatusUnauthorized)
		return
	}

	rows, err := S.db.Query(`
		SELECT sender_nickname, unread_messages 
		FROM notifications 
		WHERE receiver_nickname = ? AND unread_messages > 0`, nickname)
	if err != nil {
		http.Error(w, "Failed to fetch notifications", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	notifications := make(map[string]int)
	for rows.Next() {
		var sender string
		var unread int
		if err := rows.Scan(&sender, &unread); err != nil {
			http.Error(w, "Failed to scan notifications", http.StatusInternalServerError)
			return
		}
		notifications[sender] = unread
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notifications)
}

func (S *Server) MarkNotificationsRead(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/404", http.StatusSeeOther)
		return
	}

	cookie, err := r.Cookie("session_token")
	if err != nil {
		http.Error(w, "Unauthorized - No session", http.StatusUnauthorized)
		return
	}

	sessionID := cookie.Value
	var nickname string
	err = S.db.QueryRow(`
		SELECT nickname FROM sessions 
		WHERE session_id = ? AND expires_at > datetime('now')`, sessionID).Scan(&nickname)
	if err != nil {
		http.Error(w, "Unauthorized - Invalid session", http.StatusUnauthorized)
		return
	}

	var request struct {
		Sender string `json:"sender"`
	}
	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	_, err = S.db.Exec(`
		UPDATE notifications 
		SET unread_messages = 0 
		WHERE receiver_nickname = ? AND sender_nickname = ?`, nickname, request.Sender)
	if err != nil {
		http.Error(w, "Failed to mark notifications as read", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (S *Server) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/404", http.StatusSeeOther)
		return
	}

	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		renderErrorPage(w, r, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Issue #4: Add backend validation
	if !isValidNickname(user.Nickname) {
		http.Error(w, "Invalid nickname: must be 3-20 characters, alphanumeric and underscore only", http.StatusBadRequest)
		return
	}

	if !isValidEmail(user.Email) {
		http.Error(w, "Invalid email format", http.StatusBadRequest)
		return
	}

	if !isValidPassword(user.Password) {
		http.Error(w, "Password must be at least 8 characters with uppercase, lowercase, and number", http.StatusBadRequest)
		return
	}

	if !isValidAge(user.Age) {
		http.Error(w, "Invalid age: must be between 13 and 120", http.StatusBadRequest)
		return
	}

	if !isValidTextLength(user.FirstName, 1, 50) {
		http.Error(w, "First name must be 1-50 characters", http.StatusBadRequest)
		return
	}

	if !isValidTextLength(user.LastName, 1, 50) {
		http.Error(w, "Last name must be 1-50 characters", http.StatusBadRequest)
		return
	}

	if user.Gender != "male" && user.Gender != "female" {
		http.Error(w, "Invalid gender", http.StatusBadRequest)
		return
	}

	err, found := S.UserFound(user)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if found {
		http.Error(w, "Email or nickname already exists", http.StatusConflict)
		return
	}

	Err := S.AddUser(user)
	if Err != "" {
		renderErrorPage(w, r, Err, http.StatusInternalServerError)
		return
	}
}

// Modified LoginHandler - broadcast status change after successful login
func (S *Server) LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/404", http.StatusSeeOther)
		return
	}

	var user LoginUser
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		renderErrorPage(w, r, "Bad Request", http.StatusBadRequest)
		return
	}
	if user.Identifier == "" || user.Password == "" {
		renderErrorPage(w, r, "Bad Request", http.StatusBadRequest)
		return
	}
	hashedPassword, nickname, err := S.GetHashedPasswordFromDB(user.Identifier)
	if err != nil {
		renderErrorPage(w, r, "User not found", http.StatusNotFound)
		return
	}

	if err := CheckPassword(hashedPassword, user.Password); err != nil {
		renderErrorPage(w, r, "Incorrect password", http.StatusUnauthorized)
		return
	}

	S.MakeToken(w, nickname)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"username": nickname,
	})

	S.broadcastUserStatusChange()
}

func (S *Server) CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/404", http.StatusSeeOther)
		return
	}

	cookie, err := r.Cookie("session_token")
	if err != nil {
		http.Error(w, "Unauthorized - No session", http.StatusUnauthorized)
		return
	}

	sessionID := cookie.Value
	var nickname string
	err = S.db.QueryRow("SELECT nickname FROM sessions WHERE session_id = ? AND expires_at > datetime('now')", sessionID).Scan(&nickname)
	if err != nil {
		http.Error(w, "Unauthorized - Invalid session", http.StatusUnauthorized)
		return
	}

	var post Post

	err = json.NewDecoder(r.Body).Decode(&post)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(post.Content) == "" || strings.TrimSpace(post.Title) == "" || strings.TrimSpace(post.Category) == "" {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Validate text lengths
	if !isValidTextLength(post.Title, 3, 200) {
		http.Error(w, "Title must be 3-200 characters", http.StatusBadRequest)
		return
	}

	if !isValidTextLength(post.Content, 10, 10000) {
		http.Error(w, "Content must be 10-10000 characters", http.StatusBadRequest)
		return
	}

	if !isValidTextLength(post.Category, 3, 50) {
		http.Error(w, "Category must be 3-50 characters", http.StatusBadRequest)
		return
	}

	_, err = S.db.Exec(
		"INSERT INTO posts (user_id, title, content, category) VALUES ((SELECT id FROM users WHERE nickname = ?), ?, ?, ?)",
		html.EscapeString(nickname), html.EscapeString(post.Title), html.EscapeString(post.Content), html.EscapeString(post.Category),
	)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (S *Server) GetPostsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Redirect(w, r, "/404", http.StatusSeeOther)
		return
	}

	rows, err := S.db.Query(`
        SELECT posts.id, posts.title, posts.content, posts.category, posts.created_at, users.nickname
        FROM posts
        JOIN users ON posts.user_id = users.id
        ORDER BY posts.created_at DESC
    `)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var p Post
		err := rows.Scan(&p.ID, &p.Title, &p.Content, &p.Category, &p.CreatedAt, &p.Author)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		posts = append(posts, p)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

func (S *Server) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/404", http.StatusSeeOther)
		return
	}
	cookie, err := r.Cookie("session_token")
	if err != nil {
		http.Error(w, "No session", http.StatusBadRequest)
		return
	}

	var username string
	S.db.QueryRow("SELECT nickname FROM sessions WHERE session_id = ?", cookie.Value).Scan(&username)

	_, err = S.db.Exec("DELETE FROM sessions WHERE session_id = ?", cookie.Value)
	if err != nil {
		http.Error(w, "Error deleting session", http.StatusInternalServerError)
		return
	}

	S.RLock()
	if clients, exists := S.clients[username]; exists {
		for _, client := range clients {
			if client.SessionID == cookie.Value {
				client.Send <- map[string]string{
					"event":   "logout",
					"message": "Session terminated",
				}
				client.Conn.Close()
				break
			}
		}
	}
	S.RUnlock()

	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   "",
		Expires: time.Unix(0, 0),
		Path:    "/",
	})

	// Broadcast user status change to remaining connected clients
	go func() {
		time.Sleep(100 * time.Millisecond)
		S.broadcastUserStatusChange()
	}()

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (S *Server) LoggedHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/404", http.StatusSeeOther)
		return
	}
	username, _, err := S.CheckSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"username": username,
	})
}

func (S *Server) CreateCommentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/404", http.StatusSeeOther)
		return
	}
	cookie, err := r.Cookie("session_token")
	if err != nil {
		http.Error(w, "Unauthorized - No session", http.StatusUnauthorized)
		return
	}
	sessionID := cookie.Value
	var nickname string
	err = S.db.QueryRow("SELECT nickname FROM sessions WHERE session_id = ? AND expires_at > datetime('now')", sessionID).Scan(&nickname)
	if err != nil {
		http.Error(w, "Unauthorized - Invalid session", http.StatusUnauthorized)
		return
	}
	var comment Comment
	err = json.NewDecoder(r.Body).Decode(&comment)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(comment.Content) == "" {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Validate comment length
	if !isValidTextLength(comment.Content, 1, 1000) {
		http.Error(w, "Comment must be 1-1000 characters", http.StatusBadRequest)
		return
	}

	_, err = S.db.Exec(
		"INSERT INTO comments (post_id, user_id, content) VALUES (?, (SELECT id FROM users WHERE nickname = ?), ?)",
		(comment.PostID), html.EscapeString(nickname), html.EscapeString(comment.Content),
	)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (S *Server) GetCommentsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Redirect(w, r, "/404", http.StatusSeeOther)
		return
	}
	postID := r.URL.Query().Get("post_id")
	if postID == "" {
		http.Error(w, "Missing post_id parameter", http.StatusBadRequest)
		return
	}
	rows, err := S.db.Query(`
        SELECT comments.id, comments.content, comments.created_at, users.nickname
        FROM comments
        JOIN users ON comments.user_id = users.id
        WHERE comments.post_id = ?
        ORDER BY comments.created_at ASC
    `, postID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var c Comment
		err := rows.Scan(&c.ID, &c.Content, &c.CreatedAt, &c.Author)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		comments = append(comments, c)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comments)
}

func (s *Server) GetMessagesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/404", http.StatusSeeOther)
		return
	}
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	username, _, err := s.CheckSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if from == "" || to == "" || from != username {
		http.Error(w, "Missing parameters", http.StatusBadRequest)
		return
	}

	offsetStr := r.URL.Query().Get("offset")
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0
	}

	rows, err := s.db.Query(`
	SELECT sender, receiver, content, timestamp
	FROM messages
	WHERE (sender = ? AND receiver = ?) OR (sender = ? AND receiver = ?)
	ORDER BY id DESC
	LIMIT 10 OFFSET ?
`, from, to, to, from, offset)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "No messages found", http.StatusNotFound)
			return
		}
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		err := rows.Scan(&msg.From, &msg.To, &msg.Content, &msg.Timestamp)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "No messages found", http.StatusNotFound)
				return
			}
			http.Error(w, "Scan error", http.StatusInternalServerError)
			return
		}
		msg.Content = html.UnescapeString(msg.Content)
		messages = append([]Message{msg}, messages...)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}
