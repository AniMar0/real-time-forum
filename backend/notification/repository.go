package notification

import "database/sql"

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) IncrementUnread(tx *sql.Tx, receiverID, senderID int64) error {
	var count int
	err := tx.QueryRow(`
		SELECT unread_messages
		FROM notifications
		WHERE receiver_id = ? AND sender_id = ?`, receiverID, senderID).Scan(&count)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	count++
	if err == sql.ErrNoRows {
		var receiverNickname, senderNickname string
		if err := tx.QueryRow("SELECT nickname FROM users WHERE id = ?", receiverID).Scan(&receiverNickname); err != nil {
			return err
		}
		if err := tx.QueryRow("SELECT nickname FROM users WHERE id = ?", senderID).Scan(&senderNickname); err != nil {
			return err
		}
		_, err = tx.Exec(`
			INSERT INTO notifications (
				unread_messages, receiver_id, sender_id,
				receiver_nickname, sender_nickname
			)
			VALUES (?, ?, ?, ?, ?)`,
			count, receiverID, senderID, receiverNickname, senderNickname)
		return err
	}

	_, err = tx.Exec(`
		UPDATE notifications
		SET unread_messages = ?
		WHERE receiver_id = ? AND sender_id = ?`, count, receiverID, senderID)
	return err
}
