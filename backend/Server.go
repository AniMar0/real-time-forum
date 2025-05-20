package backend

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/twinj/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Server struct {
	db  *sql.DB
	Mux *http.ServeMux
}

func (S *Server) Run() {
	S.Mux = http.NewServeMux()
	S.DataBase()
	S.initRoutes()

	fmt.Println("Server running on http://localhost:8080")
	err := http.ListenAndServe(":8080", S.Mux)
	if err != nil {
		log.Println("Server error:", err)
		return
	}
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
	_, err = S.db.Exec(query, user.Nickname, user.FirstName, user.LastName, user.Email, string(hashedPassword), user.Age, user.Gender)
	if err != nil {
		return error.Error(err)
	}
	return ""
}

func (s *Server) SessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err != nil {
			fmt.Println("kaka")
			http.Error(w, "Unauthorized - no session", http.StatusUnauthorized)
			return
		}

		var nickname string
		err = s.db.QueryRow(`
			SELECT nickname FROM sessions 
			WHERE session_id = ? AND expires_at > CURRENT_TIMESTAMP
		`, cookie.Value).Scan(&nickname)
		if err != nil {
			fmt.Println(err.Error())
			http.Error(w, "Unauthorized - invalid session", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "nickname", nickname)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (S *Server) MakeToken(Writer http.ResponseWriter, username string) {
	sessionID := uuid.NewV4().String()
	expirationTime := time.Now().Add(24 * time.Hour)

	_, err := S.db.Exec("INSERT INTO sessions (session_id, nickname, expires_at) VALUES (?, ?, ?)",
		sessionID, username, expirationTime)
	if err != nil {
		fmt.Println("'tocken err'")
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

func (S *Server) initRoutes() {
	S.Mux.Handle("/", http.FileServer(http.Dir("./static")))

	S.Mux.Handle("/createPost", S.SessionMiddleware(http.HandlerFunc(S.CreatePostHandler)))
	S.Mux.HandleFunc("/posts", S.GetPostsHandler)

	S.Mux.HandleFunc("/register", S.RegisterHandler)
	S.Mux.HandleFunc("/login", S.LoginHandler)

	S.Mux.HandleFunc("/logout", S.LogoutHandler)
}

func (S *Server) DataBase() {
	var err error
	S.db, err = sql.Open("sqlite3", "database/forum.db")
	if err != nil {
		log.Fatal(err)
	}
}

func (S *Server) Shutdown() {
	S.db.Close()
}
