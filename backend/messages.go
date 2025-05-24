package backend

import "database/sql"

func saveMessages(db *sql.DB, msg Message) error {
	_, err := db.Exec(`
	INSER INTO messages (sender_id, receiver_id, content)
	VALUES (?, ?, ?)`,
		msg.SenderID, msg.ReceiverID, msg.Content)
	return err
}

func loadMessages(db *sql.DB, u1, u2 int) ([]Message, error) {
	rows, err := db.Query(`
		SELECT id, sender_id, receiver_id, content, sent_at
		FROM messages
		WHERE (sender_idd = ? AND receiver_id = ?) OR
				(sender_idd = ? AND receiver_id = ?)
		ORDER BY sent_at ASC`, u1, u2, u2, u1)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var messages []Message
	for rows.Next() {
		var msg Message
		if err := rows.Scan(&msg.ID, &msg.SenderID, &msg.ReceiverID, &msg.Content, &msg.Timestamp); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}
	return messages, nil
}
