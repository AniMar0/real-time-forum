package backend

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

type Server struct {
	db *sql.DB
}

func (S *Server) Run() {
	var err error
	S.db, err = sql.Open("sqlite3", "database/forum.db")
	if err != nil {
		log.Fatal(err)
	}
	defer S.db.Close()

	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/register", S.RegisterHandler)
	http.HandleFunc("/login", S.LoginHandler)
	
	fmt.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
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
