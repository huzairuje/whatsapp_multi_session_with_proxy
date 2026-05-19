package message

import (
	"database/sql"
	"time"
)

type Service struct {
	repo *Repository
}

func NewService(db *sql.DB, isPostgres bool) *Service {
	return &Service{
		repo: NewRepository(db, isPostgres),
	}
}

func (s *Service) InitializeDatabase() error {
	if s.repo.isPostgres {
		return s.repo.CreateMessagesTablePostgres()
	}
	return s.repo.CreateMessagesTable()
}

func (s *Service) RecordMessage(sender, recipient, content string) (*Message, error) {
	msg := &Message{
		Sender:    sender,
		Recipient: recipient,
		Content:   content,
		Status:    StatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := s.repo.InsertMessage(msg)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

func (s *Service) RecordMessageWithID(sender, recipient, content, messageID string) (*Message, error) {
	msg := &Message{
		Sender:    sender,
		Recipient: recipient,
		Content:   content,
		MessageID: messageID,
		Status:    StatusSent,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := s.repo.InsertMessage(msg)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

func (s *Service) UpdateMessageStatus(messageID string, status MessageStatus, errorMsg string) error {
	return s.repo.UpdateMessageStatus(messageID, status, errorMsg)
}

func (s *Service) GetMessageByID(id int64) (*Message, error) {
	return s.repo.GetMessageByID(id)
}

func (s *Service) GetMessagesBySender(sender string, limit, offset int) ([]*Message, error) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}
	return s.repo.GetMessagesBySender(sender, limit, offset)
}

func (s *Service) GetStatsBySender(sender string) (*MessageStats, error) {
	return s.repo.GetStatsBySender(sender)
}

func (s *Service) GetAllStats() (*MessageStats, error) {
	return s.repo.GetAllStats()
}
