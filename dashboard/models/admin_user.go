package models

import (
	"time"
)

// AdminUser represents an administrator user in the dashboard
type AdminUser struct {
	ID             uint       `gorm:"primaryKey;autoIncrement"`
	PhoneNumber    string     `gorm:"unique;index"`
	HashedOTP      *string    // Pointer to string to allow null values
	OTPGeneratedAt *time.Time // Pointer to time.Time to allow null values
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
