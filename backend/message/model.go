package message

import (
	"time"
)

type MessageStatus string

const (
	StatusPending   MessageStatus = "pending"
	StatusSent      MessageStatus = "sent"
	StatusFailed    MessageStatus = "failed"
	StatusDelivered MessageStatus = "delivered"
	StatusRead      MessageStatus = "read"
)

type Message struct {
	ID        int64         `json:"id"`
	Sender    string        `json:"sender"`
	Recipient string        `json:"recipient"`
	Content   string        `json:"content"`
	Status    MessageStatus `json:"status"`
	MessageID string        `json:"message_id"`
	Error     string        `json:"error,omitempty"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

type MessageStats struct {
	TotalSent     int64   `json:"total_sent"`
	TotalFailed   int64   `json:"total_failed"`
	TotalDelivered int64  `json:"total_delivered"`
	SuccessRate   float64 `json:"success_rate"`
	DailyCount    int64   `json:"daily_count"`
	DailyLimit    int64   `json:"daily_limit"`
}

type CreateMessageRequest struct {
	Sender    string `json:"sender" binding:"required"`
	Recipient string `json:"recipient" binding:"required"`
	Content   string `json:"content" binding:"required"`
}

type UpdateMessageStatusRequest struct {
	MessageID string        `json:"message_id" binding:"required"`
	Status    MessageStatus `json:"status" binding:"required"`
	Error     string        `json:"error"`
}
