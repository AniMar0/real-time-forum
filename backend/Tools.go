package backend

import (
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

type Error struct {
	Err       string
	ErrNumber string
}

func renderErrorPage(w http.ResponseWriter, r *http.Request, errMsg string, errCode int) {
	http.ServeFile(w, r, "./static/index.html")
}

func CheckPassword(hashedPassword, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err
}

func checkHome(next http.Handler) http.Handler {

	// Issue #5: Update to include register.js instead of regester.js
	Paths := []string{"/app.js", "/chat.js", "/comments.js", "/error.js", "/login.js", "/logout.js", "/posts.js", "/register.js", "/style.css", "/"}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, p := range Paths {
			if r.URL.Path == p {
				next.ServeHTTP(w, r)
				return
			}
		}

		if r.URL.Path != "/" || r.Method != http.MethodGet {
			renderErrorPage(w, r, "Not Found", http.StatusNotFound)
			return
		}
	})
}
