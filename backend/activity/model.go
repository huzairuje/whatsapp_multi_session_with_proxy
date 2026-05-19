package activity

import (
	"time"
)

type ActivityType string

const (
	TypeSessionConnect    ActivityType = "session_connect"
	TypeSessionDisconnect ActivityType = "session_disconnect"
	TypeSessionLogout     ActivityType = "session_logout"
	TypeMessageSent       ActivityType = "message_sent"
	TypeMessageFailed     ActivityType = "message_failed"
	TypeBulkSendStart     ActivityType = "bulk_send_start"
	TypeBulkSendComplete  ActivityType = "bulk_send_complete"
	TypeBulkSendError     ActivityType = "bulk_send_error"
	TypeRateLimit         ActivityType = "rate_limit"
	TypeAutoLogin         ActivityType = "auto_login"
	TypeHealthCheck       ActivityType = "health_check"
	TypeHealthCheckFailed ActivityType = "health_check_failed"
	TypeQRGenerated       ActivityType = "qr_generated"
	TypeUserLogin         ActivityType = "user_login"
	TypeUserLogout        ActivityType = "user_logout"
)

type Activity struct {
	ID        int64        `json:"id"`
	Type      ActivityType `json:"type"`
	Sender    string       `json:"sender,omitempty"`
	User      string       `json:"user,omitempty"`
	Message   string       `json:"message"`
	Details   string       `json:"details,omitempty"`
	Status    string       `json:"status,omitempty"`
	ErrorMsg  string       `json:"error,omitempty"`
	CreatedAt time.Time    `json:"created_at"`
}

type ActivityStats struct {
	TotalActivities   int64                    `json:"total_activities"`
	ActivitiesByType  map[string]int64         `json:"activities_by_type"`
	RecentActivities  []*Activity              `json:"recent_activities"`
	SessionsConnected int64                    `json:"sessions_connected"`
	MessagesSent      int64                    `json:"messages_sent"`
	MessagesFailed    int64                    `json:"messages_failed"`
	RateLimitEvents   int64                    `json:"rate_limit_events"`
}

type LogActivityRequest struct {
	Type    ActivityType `json:"type" binding:"required"`
	Sender  string       `json:"sender"`
	User    string       `json:"user"`
	Message string       `json:"message" binding:"required"`
	Details string       `json:"details"`
	Status  string       `json:"status"`
	Error   string       `json:"error"`
}
