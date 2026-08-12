package chat

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"real-time-forum/backend/notification"
)

var (
	ErrInvalidRecipient = errors.New("invalid message recipient")
	ErrInvalidContent   = errors.New("invalid message content")
)

type Service struct {
	db            *sql.DB
	repository    *Repository
	notifications *notification.Repository
}

func NewService(db *sql.DB, repository *Repository, notifications *notification.Repository) *Service {
	return &Service{db: db, repository: repository, notifications: notifications}
}

func (s *Service) SendMessage(senderID int64, receiver, content string) (Message, error) {
	sender, err := s.repository.UserByID(senderID)
	if err != nil {
		return Message{}, err
	}
	if receiver == "" || receiver == sender {
		return Message{}, ErrInvalidRecipient
	}
	content = strings.TrimSpace(content)
	if len(content) < 1 || len(content) > 5000 {
		return Message{}, ErrInvalidContent
	}
	receiverID, err := s.repository.UserIDByNickname(receiver)
	if err != nil {
		return Message{}, ErrInvalidRecipient
	}

	tx, err := s.db.Begin()
	if err != nil {
		return Message{}, err
	}
	message, err := s.repository.InsertMessage(tx, Message{
		SenderID:   senderID,
		ReceiverID: receiverID,
		From:       sender,
		To:         receiver,
		Content:    content,
		Timestamp:  time.Now().Format(time.RFC3339),
	})
	if err != nil {
		_ = tx.Rollback()
		return Message{}, err
	}
	if err := s.notifications.IncrementUnread(tx, receiverID, senderID); err != nil {
		_ = tx.Rollback()
		return Message{}, err
	}
	if err := tx.Commit(); err != nil {
		return Message{}, err
	}

	message.Content = content
	return message, nil
}
