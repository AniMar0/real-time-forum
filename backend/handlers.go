package backend

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
)

func (S *Server) HomeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		renderErrorPage(w, "Template error", http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
}

func (S *Server) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		renderErrorPage(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		fmt.Println("err")
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

func (S Server) StaticFileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet || r.URL.Path == "/static/" {
		renderErrorPage(w, "Bad Request", http.StatusBadRequest)
		return
	}

	fileCssServe := http.FileServer(http.Dir("static"))
	http.StripPrefix("/static/", fileCssServe).ServeHTTP(w, r)
}
