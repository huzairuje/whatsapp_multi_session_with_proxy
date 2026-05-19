package template

import "time"

type MessageTemplate struct {
	ID          int       `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Content     string    `json:"content" db:"content"`
	Variables   []string  `json:"variables" db:"variables"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type CreateTemplateRequest struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
	Content     string `json:"content" validate:"required"`
}

type UpdateTemplateRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Content     *string `json:"content,omitempty"`
}

type ApplyTemplateRequest struct {
	TemplateID int               `json:"template_id" validate:"required"`
	Recipients []RecipientData   `json:"recipients" validate:"required"`
}

type RecipientData struct {
	Phone     string            `json:"phone" validate:"required"`
	Variables map[string]string `json:"variables"`
}

type TemplatePreview struct {
	Phone   string `json:"phone"`
	Message string `json:"message"`
}
