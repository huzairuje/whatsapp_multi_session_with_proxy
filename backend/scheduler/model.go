package scheduler

import "time"

type ScheduledJob struct {
	ID              int       `json:"id" db:"id"`
	SenderJID       string    `json:"sender_jid" db:"sender_jid"`
	TemplateID      *int      `json:"template_id" db:"template_id"`
	Recipients      string    `json:"recipients" db:"recipients"`
	MessageVariants string    `json:"message_variants" db:"message_variants"`
	TotalMessages   int       `json:"total_messages" db:"total_messages"`
	SentMessages    int       `json:"sent_messages" db:"sent_messages"`
	Status          string    `json:"status" db:"status"`
	ScheduledFor    time.Time `json:"scheduled_for" db:"scheduled_for"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

type CreateScheduledJobRequest struct {
	SenderJID       string   `json:"sender_jid" validate:"required"`
	TemplateID      *int     `json:"template_id"`
	Recipients      []string `json:"recipients" validate:"required"`
	MessageVariants []string `json:"message_variants"`
	StartDate       time.Time `json:"start_date"`
}

type ScheduleConfig struct {
	AllowedHourStart int
	AllowedHourEnd   int
	Timezone         string
	DailyLimit       int
	MinDelayMs       int
	MaxDelayMs       int
}

type ScheduledMessage struct {
	Recipient string
	Message   string
	ScheduledTime time.Time
}
