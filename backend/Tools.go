package backend

import (
	"database/sql"
	"fmt"
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
	Paths := []string{"/app.js", "/chat.js", "/comments.js", "/dom.js", "/error.js", "/login.js", "/logout.js", "/posts.js", "/register.js", "/style.css", "/toast.js", "/"}
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
func (S *Server) addNewNotification(resiverID, senderID string) {
	qery := `UPDATE notifications SET unread_messages = ? WHERE receiver_nickname = ? And sender_nickname = ?`

	oldCount, err := S.getNotificationCount(resiverID, senderID)
	if err != nil {
		if err == sql.ErrNoRows {
			qery = `INSERT INTO notifications (unread_messages, receiver_nickname, sender_nickname) VALUES (?, ?, ?)`
		} else {
			fmt.Println("Error retrieving old notification count:", err)
			return
		}
	}
	newCount := oldCount + 1
	_, err = S.db.Exec(qery, newCount, resiverID, senderID)
	if err != nil {
		fmt.Println("Error updating notification count:", err)
		return
	}
}

func (S *Server) getNotificationCount(resiverID, senderID string) (int, error) {
	var count int
	qery := `SELECT unread_messages FROM notifications WHERE receiver_nickname = ? And sender_nickname = ?`
	err := S.db.QueryRow(qery, resiverID, senderID).Scan(&count)
	if err != nil {
		fmt.Println("Error getting notification count:", err)
		return 0, err
	}
	return count, nil
}
