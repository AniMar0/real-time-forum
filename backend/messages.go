package backend

import "database/sql"

func saveMessages(db *sql.DB, msg Message) error {
	_, err := db.Exec(`
	INSER INTO messages (sender_id, reciever_id, content)
	VALUES (?, ?, ?)`,
	msg.SenderID, msg.ReceiverID, msg.Content)
	return err
}
