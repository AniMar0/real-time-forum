package backend

import (
	"encoding/json"
	"html"
	"net/http"
	"strconv"
	"strings"
	"time"

	"real-time-forum/backend/account"
	"real-time-forum/backend/forum"
)

func (S *Server) GetNotifications(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Redirect(w, r, "/404", http.StatusSeeOther)
		return
	}

	identity, err := S.CheckSessionIdentity(r)
	if err != nil {
		http.Error(w, "Unauthorized - Invalid session", http.StatusUnauthorized)
		return
	}

	if S.notifications == nil {
		http.Error(w, "Notification repository is not initialized", http.StatusInternalServerError)
		return
	}
	notifications, err := S.notifications.ListUnread(identity.UserID)
	if err != nil {
		http.Error(w, "Failed to fetch notifications", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notifications)
}

func (S *Server) MarkNotificationsRead(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/404", http.StatusSeeOther)
		return
	}

	identity, err := S.CheckSessionIdentity(r)
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

	if S.notifications == nil {
		http.Error(w, "Notification repository is not initialized", http.StatusInternalServerError)
		return
	}
	err = S.notifications.MarkRead(identity.UserID, request.Sender)
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

	if S.users == nil {
		http.Error(w, "Account repository is not initialized", http.StatusInternalServerError)
		return
	}
	found, err := S.users.Exists(user.Email, user.Nickname)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if found {
		http.Error(w, "Email or nickname already exists", http.StatusConflict)
		return
	}

	err = S.users.Create(account.UserRecord{
		Nickname: user.Nickname, FirstName: user.FirstName, LastName: user.LastName,
		Email: user.Email, Password: user.Password, Age: user.Age, Gender: user.Gender,
	})
	if err != nil {
		renderErrorPage(w, r, "Unable to create account", http.StatusInternalServerError)
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
	if S.users == nil {
		http.Error(w, "Account repository is not initialized", http.StatusInternalServerError)
		return
	}
	hashedPassword, nickname, err := S.users.PasswordByIdentifier(user.Identifier)
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

	nickname, _, err := S.CheckSession(r)
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

	if S.forum == nil {
		http.Error(w, "Forum repository is not initialized", http.StatusInternalServerError)
		return
	}
	err = S.forum.CreatePost(
		html.EscapeString(nickname),
		html.EscapeString(post.Title),
		html.EscapeString(post.Content),
		html.EscapeString(post.Category),
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

	if S.forum == nil {
		http.Error(w, "Forum repository is not initialized", http.StatusInternalServerError)
		return
	}
	posts, err := S.forum.ListPosts()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

func (S *Server) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/404", http.StatusSeeOther)
		return
	}
	identity, err := S.CheckSessionIdentity(r)
	if err != nil {
		http.Error(w, "No session", http.StatusBadRequest)
		return
	}
	if S.sessions == nil {
		http.Error(w, "Session repository is not initialized", http.StatusInternalServerError)
		return
	}
	err = S.sessions.Delete(identity.SessionID)
	if err != nil {
		http.Error(w, "Error deleting session", http.StatusInternalServerError)
		return
	}

	for _, session := range S.hub.ClientsForUser(identity.UserID) {
		if session.SessionID == identity.SessionID {
			session.Enqueue(map[string]string{
				"event":   "logout",
				"message": "Session terminated",
			})
			S.hub.Unregister(session)
			break
		}
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Expires:  time.Unix(0, 0),
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   S.config.SecureCookies(),
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
	nickname, _, err := S.CheckSession(r)
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

	if S.forum == nil {
		http.Error(w, "Forum repository is not initialized", http.StatusInternalServerError)
		return
	}
	err = S.forum.CreateComment(comment.PostID, html.EscapeString(nickname), html.EscapeString(comment.Content))
	if err != nil {
		if err == forum.ErrPostNotFound {
			http.Error(w, "Post not found", http.StatusBadRequest)
			return
		}
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
	if S.forum == nil {
		http.Error(w, "Forum repository is not initialized", http.StatusInternalServerError)
		return
	}
	comments, err := S.forum.ListComments(postID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
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
	beforeID := 0
	if beforeIDValue := r.URL.Query().Get("before_id"); beforeIDValue != "" {
		beforeID, err = strconv.Atoi(beforeIDValue)
		if err != nil || beforeID < 1 {
			http.Error(w, "Invalid before_id", http.StatusBadRequest)
			return
		}
	}

	if s.chat == nil {
		http.Error(w, "Chat repository is not initialized", http.StatusInternalServerError)
		return
	}
	messagesFromRepository, err := s.chat.ListHistory(from, to, beforeID, offset)
	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	var messages []Message
	for _, storedMessage := range messagesFromRepository {
		messages = append([]Message{{
			ID:        storedMessage.ID,
			From:      storedMessage.From,
			To:        storedMessage.To,
			Content:   storedMessage.Content,
			Timestamp: storedMessage.Timestamp,
		}}, messages...)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}
