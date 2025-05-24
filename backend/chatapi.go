package backend

import (
	"database/sql"
	"encoding/json"
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
	if err != nil {
		http.Error(w, "Failed loading messages", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}