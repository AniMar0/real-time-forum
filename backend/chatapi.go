package backend

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
)

var MessageDB *sql.DB

func getMessageHistor(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(int)
	otherIDstr := r.URL.Query().Get("with")
	otherID, err := strconv.Atoi(otherIDstr)
	if err != nil {
		http.Error(w, "Ivalid user ID", http.StatusBadRequest)
		return 
	}
	messages, err := loadMessages(MessageDB, userID, otherID)
}