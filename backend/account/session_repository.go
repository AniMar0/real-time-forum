package account

import (
	"database/sql"
	"fmt"
	"time"
)

type SessionRepository struct {
	db *sql.DB
}

func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Create(sessionID, nickname string, expiresAt time.Time) error {
	var userID int64
	if err := r.db.QueryRow("SELECT id FROM users WHERE nickname = ?", nickname).Scan(&userID); err != nil {
		return err
	}
	_, err := r.db.Exec(`
		INSERT INTO sessions (session_id, user_id, nickname, expires_at)
		VALUES (?, ?, ?, ?)`, sessionID, userID, nickname, expiresAt)
	return err
}

func (r *SessionRepository) FindValid(sessionID string) (Identity, error) {
	var identity Identity
	err := r.db.QueryRow(`
		SELECT s.user_id, u.nickname
		FROM sessions s
		JOIN users u ON u.id = s.user_id
		WHERE s.session_id = ? AND s.expires_at > CURRENT_TIMESTAMP`, sessionID).Scan(&identity.UserID, &identity.Nickname)
	if err != nil {
		return Identity{}, fmt.Errorf("find valid session: %w", err)
	}
	identity.SessionID = sessionID
	return identity, nil
}

func (r *SessionRepository) Delete(sessionID string) error {
	_, err := r.db.Exec("DELETE FROM sessions WHERE session_id = ?", sessionID)
	return err
}
