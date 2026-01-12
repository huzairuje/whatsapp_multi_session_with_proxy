package models

import (
	"time"
)

// ManagedDevice represents a WhatsApp device managed by an admin user
type ManagedDevice struct {
	ID          uint   `gorm:"primaryKey;autoIncrement"`
	AdminUserID uint   `gorm:"index"` // Foreign Key to AdminUser
	AdminUser   AdminUser `gorm:"foreignKey:AdminUserID"` // GORM relation
	DeviceJID   string `gorm:"index"` // WhatsApp JID of the managed device
	DeviceName  string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
