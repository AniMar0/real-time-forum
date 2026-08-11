package chat

import "database/sql"

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) UserExists(nickname string) (bool, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM users WHERE nickname = ?", nickname).Scan(&count)
	return count > 0, err
}

func (r *Repository) InsertMessage(tx *sql.Tx, message Message) (Message, error) {
	result, err := tx.Exec(`
		INSERT INTO messages (sender, receiver, content, timestamp)
		VALUES (?, ?, ?, ?)`,
		message.From, message.To, message.Content, message.Timestamp)
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
			WHERE ((sender = ? AND receiver = ?) OR (sender = ? AND receiver = ?))
			  AND id < ?
			ORDER BY id DESC
			LIMIT 10`, from, to, to, from, beforeID)
	} else {
		rows, err = r.db.Query(`
			SELECT id, sender, receiver, content, timestamp
			FROM messages
			WHERE (sender = ? AND receiver = ?) OR (sender = ? AND receiver = ?)
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
