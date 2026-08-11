package chat

import "database/sql"

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) UserIDByNickname(nickname string) (int64, error) {
	var userID int64
	err := r.db.QueryRow("SELECT id FROM users WHERE nickname = ?", nickname).Scan(&userID)
	return userID, err
}

func (r *Repository) UserByID(userID int64) (string, error) {
	var nickname string
	err := r.db.QueryRow("SELECT nickname FROM users WHERE id = ?", userID).Scan(&nickname)
	return nickname, err
}

func (r *Repository) InsertMessage(tx *sql.Tx, message Message) (Message, error) {
	result, err := tx.Exec(`
		INSERT INTO messages (sender_id, receiver_id, sender, receiver, content, timestamp)
		VALUES (?, ?, ?, ?, ?, ?)`,
		message.SenderID, message.ReceiverID, message.From, message.To, message.Content, message.Timestamp)
	if err != nil {
		return Message{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return Message{}, err
	}
	message.ID = int(id)
	return message, nil
}

func (r *Repository) ListHistory(from, to string, beforeID, offset int) ([]Message, error) {
	var rows *sql.Rows
	var err error
	if beforeID > 0 {
		rows, err = r.db.Query(`
			SELECT id, sender, receiver, content, timestamp
			FROM messages
			WHERE ((sender_id = (SELECT id FROM users WHERE nickname = ?) AND receiver_id = (SELECT id FROM users WHERE nickname = ?))
			   OR (sender_id = (SELECT id FROM users WHERE nickname = ?) AND receiver_id = (SELECT id FROM users WHERE nickname = ?)))
			  AND id < ?
			ORDER BY id DESC
			LIMIT 10`, from, to, to, from, beforeID)
	} else {
		rows, err = r.db.Query(`
			SELECT id, sender, receiver, content, timestamp
			FROM messages
			WHERE (sender_id = (SELECT id FROM users WHERE nickname = ?) AND receiver_id = (SELECT id FROM users WHERE nickname = ?))
			   OR (sender_id = (SELECT id FROM users WHERE nickname = ?) AND receiver_id = (SELECT id FROM users WHERE nickname = ?))
			ORDER BY id DESC
			LIMIT 10 OFFSET ?`, from, to, to, from, offset)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var message Message
		if err := rows.Scan(&message.ID, &message.From, &message.To, &message.Content, &message.Timestamp); err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return messages, nil
}
