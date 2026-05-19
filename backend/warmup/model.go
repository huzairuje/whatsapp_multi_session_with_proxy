package warmup

import "time"

type WarmUpConfig struct {
	ID              int       `json:"id" db:"id"`
	SenderJID       string    `json:"sender_jid" db:"sender_jid"`
	Enabled         bool      `json:"enabled" db:"enabled"`
	CurrentDay      int       `json:"current_day" db:"current_day"`
	StartDate       time.Time `json:"start_date" db:"start_date"`
	DailyLimit      int       `json:"daily_limit" db:"daily_limit"`
	IncrementAmount int       `json:"increment_amount" db:"increment_amount"`
	IncrementDays   int       `json:"increment_days" db:"increment_days"`
	MaxDailyLimit   int       `json:"max_daily_limit" db:"max_daily_limit"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

type CreateWarmUpRequest struct {
	SenderJID       string `json:"sender_jid" validate:"required"`
	Enabled         bool   `json:"enabled"`
	DailyLimit      int    `json:"daily_limit" validate:"required,min=1"`
	IncrementAmount int    `json:"increment_amount" validate:"required,min=1"`
	IncrementDays   int    `json:"increment_days" validate:"required,min=1"`
	MaxDailyLimit   int    `json:"max_daily_limit" validate:"required,min=1"`
}

type UpdateWarmUpRequest struct {
	Enabled         *bool `json:"enabled,omitempty"`
	DailyLimit      *int  `json:"daily_limit,omitempty"`
	IncrementAmount *int  `json:"increment_amount,omitempty"`
	IncrementDays   *int  `json:"increment_days,omitempty"`
	MaxDailyLimit   *int  `json:"max_daily_limit,omitempty"`
}
