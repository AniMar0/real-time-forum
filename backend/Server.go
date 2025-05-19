package backend

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
)

type Server struct {
	db *sql.DB
}

func (S *Server) Run() {
	var err error
	S.db, err = sql.Open("sqlite3", "./forum.db")
	if err != nil {
		log.Fatal(err)
	}
	defer S.db.Close()

	http.HandleFunc("/", HomeHandler)

	fmt.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
