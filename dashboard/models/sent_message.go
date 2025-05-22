package models

import (
	"time"
)

// SentMessage represents a message sent via a managed device
type SentMessage struct {
	ID                   uint      `gorm:"primaryKey;autoIncrement"`
	AdminUserID          uint      `gorm:"index"` // Foreign Key to AdminUser
	AdminUser            AdminUser `gorm:"foreignKey:AdminUserID"` // GORM relation
	DeviceJID            string    `gorm:"index"`                  // Sender device JID
	RecipientNumber      string    `gorm:"index"`
	MessageContent       string    `gorm:"type:text"`
	Status               string    // e.g., "sent", "failed", "delivered"
	SentAt               time.Time
	MessageIDFromWhatsApp *string   `gorm:"index"` // Nullable, ID of the message from WhatsApp network
	CreatedAt            time.Time
	UpdatedAt            time.Time
}
