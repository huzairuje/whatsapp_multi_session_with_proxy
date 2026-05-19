package contacts

import "time"

type Contact struct {
	ID            int       `json:"id" db:"id"`
	SenderJID     string    `json:"sender_jid" db:"sender_jid"`
	ContactJID    string    `json:"contact_jid" db:"contact_jid"`
	ContactName   string    `json:"contact_name" db:"contact_name"`
	PushName      string    `json:"push_name" db:"push_name"`
	BusinessName  string    `json:"business_name" db:"business_name"`
	FirstName     string    `json:"first_name" db:"first_name"`
	FullName      string    `json:"full_name" db:"full_name"`
	IsBlocked     bool      `json:"is_blocked" db:"is_blocked"`
	IsBusiness    bool      `json:"is_business" db:"is_business"`
	IsEnterprise  bool      `json:"is_enterprise" db:"is_enterprise"`
	LastSyncedAt  time.Time `json:"last_synced_at" db:"last_synced_at"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

type SyncContactsRequest struct {
	SenderJID string `json:"sender_jid" validate:"required"`
	Force     bool   `json:"force"`
}

type SyncContactsResponse struct {
	TotalContacts int       `json:"total_contacts"`
	NewContacts   int       `json:"new_contacts"`
	UpdatedContacts int     `json:"updated_contacts"`
	SyncedAt      time.Time `json:"synced_at"`
}

type DeleteContactRequest struct {
	SenderJID  string `json:"sender_jid" validate:"required"`
	ContactJID string `json:"contact_jid" validate:"required"`
}

type ContactFilter struct {
	SenderJID    string
	SearchQuery  string
	IsBlocked    *bool
	IsBusiness   *bool
	IsEnterprise *bool
	Limit        int
	Offset       int
}
