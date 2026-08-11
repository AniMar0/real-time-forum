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
		_, err = tx.Exec(`
			INSERT INTO notifications (unread_messages, receiver_id, sender_id)
			VALUES (?, ?, ?)`, count, receiverID, senderID)
		return err
	}

	_, err = tx.Exec(`
		UPDATE notifications
		SET unread_messages = ?
		WHERE receiver_id = ? AND sender_id = ?`, count, receiverID, senderID)
	return err
}

func (r *Repository) ListUnread(receiverID int64) (map[string]int, error) {
	rows, err := r.db.Query(`
		SELECT users.nickname, notifications.unread_messages
		FROM notifications
		JOIN users ON users.id = notifications.sender_id
		WHERE notifications.receiver_id = ? AND notifications.unread_messages > 0`, receiverID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	notifications := make(map[string]int)
	for rows.Next() {
		var sender string
		var unread int
		if err := rows.Scan(&sender, &unread); err != nil {
			return nil, err
		}
		notifications[sender] = unread
	}
	return notifications, rows.Err()
}

func (r *Repository) MarkRead(receiverID int64, senderNickname string) error {
	_, err := r.db.Exec(`
		UPDATE notifications
		SET unread_messages = 0
		WHERE receiver_id = ?
		  AND sender_id = (SELECT id FROM users WHERE nickname = ?)`, receiverID, senderNickname)
	return err
}
