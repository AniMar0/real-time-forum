package account

import (
	"database/sql"
	"errors"
	"fmt"
	"html"

	"golang.org/x/crypto/bcrypt"
)

var ErrUserNotFound = errors.New("user not found")

type UserRecord struct {
	Nickname  string
	FirstName string
	LastName  string
	Email     string
	Password  string
	Age       int
	Gender    string
}

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Exists(email, nickname string) (bool, error) {
	var count int
	err := r.db.QueryRow(
		"SELECT COUNT(*) FROM users WHERE email = ? OR nickname = ?",
		email,
		nickname,
	).Scan(&count)
	return count > 0, err
}

func (r *UserRepository) Create(user UserRecord) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	_, err = r.db.Exec(`
		INSERT INTO users (nickname, first_name, last_name, email, password, age, gender)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		html.EscapeString(user.Nickname),
		html.EscapeString(user.FirstName),
		html.EscapeString(user.LastName),
		html.EscapeString(user.Email),
		string(hashedPassword),
		user.Age,
		user.Gender,
	)
	return err
}

func (r *UserRepository) PasswordByIdentifier(identifier string) (string, string, error) {
	var hashedPassword, nickname string
	err := r.db.QueryRow(`
		SELECT password, nickname FROM users
		WHERE nickname = ? OR email = ?`, identifier, identifier).
		Scan(&hashedPassword, &nickname)
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", ErrUserNotFound
	}
	if err != nil {
		return "", "", fmt.Errorf("find user credentials: %w", err)
	}
	return hashedPassword, nickname, nil
}
