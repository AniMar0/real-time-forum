package chat

import (
	"database/sql"
	"errors"
	"html"
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

func (s *Service) SendMessage(sender, receiver, content string) (Message, error) {
	if receiver == "" || receiver == sender {
		return Message{}, ErrInvalidRecipient
	}
	content = strings.TrimSpace(content)
	if len(content) < 1 || len(content) > 5000 {
		return Message{}, ErrInvalidContent
	}
	receiverExists, err := s.repository.UserExists(receiver)
	if err != nil {
		return Message{}, err
	}
	if !receiverExists {
		return Message{}, ErrInvalidRecipient
	}

	tx, err := s.db.Begin()
	if err != nil {
		return Message{}, err
	}
	message, err := s.repository.InsertMessage(tx, Message{
		From:      sender,
		To:        receiver,
		Content:   html.EscapeString(content),
		Timestamp: time.Now().Format(time.RFC3339),
	})
	if err != nil {
		_ = tx.Rollback()
		return Message{}, err
	}
	if err := s.notifications.IncrementUnread(tx, receiver, sender); err != nil {
		_ = tx.Rollback()
		return Message{}, err
	}
	if err := tx.Commit(); err != nil {
		return Message{}, err
	}

	message.Content = content
	return message, nil
}
