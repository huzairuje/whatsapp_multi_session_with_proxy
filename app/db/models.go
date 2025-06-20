package db

import (
	"time"
)

type User struct {
	ID        uint           `gorm:"primaryKey"`
	Username  string         `gorm:"uniqueIndex;not null"`
	PasswordHash string      `gorm:"not null"`
	CreatedAt time.Time
	Devices   []Device       `gorm:"foreignKey:UserID"`
	MessageCounts MessageCount `gorm:"foreignKey:UserID"` // Assuming one-to-one or one-to-many where we only care about the latest/summary
}

type Device struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"not null"`
	DeviceName string   `gorm:"not null"`
	CreatedAt time.Time
}

type MessageCount struct {
	ID           uint      `gorm:"primaryKey"`
	UserID       uint      `gorm:"uniqueIndex;not null"` // UserID should be unique if it's a summary table for each user
	Count        int       `gorm:"default:0"`
	LastUpdatedAt time.Time
}
