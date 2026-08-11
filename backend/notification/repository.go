package notification

import "database/sql"

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) IncrementUnread(tx *sql.Tx, receiver, sender string) error {
	var count int
	err := tx.QueryRow(`
		SELECT unread_messages
		FROM notifications
		WHERE receiver_nickname = ? AND sender_nickname = ?`, receiver, sender).Scan(&count)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	count++
	if err == sql.ErrNoRows {
		_, err = tx.Exec(`
			INSERT INTO notifications (unread_messages, receiver_nickname, sender_nickname)
			VALUES (?, ?, ?)`, count, receiver, sender)
		return err
	}

	_, err = tx.Exec(`
		UPDATE notifications
		SET unread_messages = ?
		WHERE receiver_nickname = ? AND sender_nickname = ?`, count, receiver, sender)
	return err
}
