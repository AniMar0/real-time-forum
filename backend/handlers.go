package backend

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func (S *Server) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		renderErrorPage(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		renderErrorPage(w, "Bad Request", http.StatusBadRequest)
		return
	}

	err, found := S.UserFound(user)
	if err != nil {
		renderErrorPage(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if found {
		renderErrorPage(w, "Status Conflict", http.StatusConflict)
		return
	}

	Err := S.AddUser(user)
	if Err != "" {
		renderErrorPage(w, Err, http.StatusInternalServerError)
		return
	}
}

func (S *Server) LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		renderErrorPage(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var user LoginUser
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		renderErrorPage(w, "Bad Request", http.StatusBadRequest)
		return
	}

	//fmt.Println(user)

	nickname, hashedPassword, err := S.GetHashedPasswordFromDB(user.Identifier)
	if err != nil {
		fmt.Println("undif")
		return
	}

	if err := CheckPassword(hashedPassword, user.Password); err != nil {
		fmt.Println("err")
		renderErrorPage(w, "Inccorect password", http.StatusInternalServerError)
		return
	}

	S.MakeToken(w, nickname)
}

func (S *Server) CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
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

	var post struct {
		Title    string `json:"title"`
		Content  string `json:"content"`
		Category string `json:"category"`
	}
	err = json.NewDecoder(r.Body).Decode(&post)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	_, err = S.db.Exec(
		"INSERT INTO posts (user_id, title, content, category) VALUES ((SELECT id FROM users WHERE nickname = ?), ?, ?, ?)",
		nickname, post.Title, post.Content, post.Category,
	)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (S *Server) GetPostsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
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

	type Post struct {
		ID        int    `json:"id"`
		Title     string `json:"title"`
		Content   string `json:"content"`
		Category  string `json:"category"`
		CreatedAt string `json:"created_at"`
		Author    string `json:"author"`
	}

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
	cookie, err := r.Cookie("session_token")
	if err != nil {
		http.Error(w, "No session", http.StatusBadRequest)
		return
	}

	_, err = S.db.Exec("DELETE FROM sessions WHERE session_id = ?", cookie.Value)
	if err != nil {
		http.Error(w, "Error deleting session", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   "",
		Expires: time.Unix(0, 0),
		Path:    "/",
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
