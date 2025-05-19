package backend

import (
	"encoding/json"
	"fmt"
	"net/http"
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

	fmt.Println(user)

	nickname, hashedPassword, err := S.GetHashedPasswordFromDB(user.Identifier)
	if err != nil {
		//err
		return
	}

	if err := CheckPassword(hashedPassword, user.Password); err != nil {
		renderErrorPage(w, "Inccorect password", http.StatusInternalServerError)
		return
	}

	S.MakeTocken(w, nickname)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
