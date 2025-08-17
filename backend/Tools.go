package backend

import (
	"encoding/json"
	"net/http"
	"strconv"

	"golang.org/x/crypto/bcrypt"
)

type Error struct {
	Err       string
	ErrNumber string
}

func renderErrorPage(w http.ResponseWriter, errMsg string, errCode int) {
	w.WriteHeader(errCode)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Error{Err: errMsg, ErrNumber: strconv.Itoa(errCode)})
}

func CheckPassword(hashedPassword, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err
}
