package template

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

type Service struct {
	db         *sql.DB
	isPostgres bool
}

func NewService(db *sql.DB, isPostgres bool) (*Service, error) {
	s := &Service{
		db:         db,
		isPostgres: isPostgres,
	}

	if err := s.createTable(); err != nil {
		return nil, fmt.Errorf("failed to create template table: %w", err)
	}

	return s, nil
}

func (s *Service) createTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS message_templates (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		description TEXT,
		content TEXT NOT NULL,
		variables TEXT,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`

	if s.isPostgres {
		query = `
		CREATE TABLE IF NOT EXISTS message_templates (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			description TEXT,
			content TEXT NOT NULL,
			variables TEXT,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`
	}

	_, err := s.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create message_templates table: %w", err)
	}

	log.Info("[Template] message_templates table ready")
	return nil
}

func (s *Service) Create(req CreateTemplateRequest) (*MessageTemplate, error) {
	now := time.Now()
	
	variables := extractVariables(req.Content)
	variablesJSON, err := json.Marshal(variables)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal variables: %w", err)
	}

	var query string
	var id int64

	if s.isPostgres {
		query = `
		INSERT INTO message_templates (name, description, content, variables, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`
		
		err := s.db.QueryRow(query, req.Name, req.Description, req.Content, string(variablesJSON), now, now).Scan(&id)
		if err != nil {
			return nil, fmt.Errorf("failed to create template: %w", err)
		}
	} else {
		query = `
		INSERT INTO message_templates (name, description, content, variables, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)`
		
		result, err := s.db.Exec(query, req.Name, req.Description, req.Content, string(variablesJSON), now, now)
		if err != nil {
			return nil, fmt.Errorf("failed to create template: %w", err)
		}
		
		id, err = result.LastInsertId()
		if err != nil {
			return nil, fmt.Errorf("failed to get last insert id: %w", err)
		}
	}

	return s.GetByID(int(id))
}

func (s *Service) GetByID(id int) (*MessageTemplate, error) {
	var template MessageTemplate
	var variablesJSON string
	
	query := `SELECT id, name, description, content, variables, created_at, updated_at
	          FROM message_templates WHERE id = ?`
	
	if s.isPostgres {
		query = `SELECT id, name, description, content, variables, created_at, updated_at
		         FROM message_templates WHERE id = $1`
	}

	err := s.db.QueryRow(query, id).Scan(
		&template.ID, &template.Name, &template.Description, &template.Content,
		&variablesJSON, &template.CreatedAt, &template.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("template not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	if variablesJSON != "" {
		if err := json.Unmarshal([]byte(variablesJSON), &template.Variables); err != nil {
			log.Warnf("[Template] Failed to unmarshal variables: %v", err)
			template.Variables = []string{}
		}
	}

	return &template, nil
}

func (s *Service) GetByName(name string) (*MessageTemplate, error) {
	var template MessageTemplate
	var variablesJSON string
	
	query := `SELECT id, name, description, content, variables, created_at, updated_at
	          FROM message_templates WHERE name = ?`
	
	if s.isPostgres {
		query = `SELECT id, name, description, content, variables, created_at, updated_at
		         FROM message_templates WHERE name = $1`
	}

	err := s.db.QueryRow(query, name).Scan(
		&template.ID, &template.Name, &template.Description, &template.Content,
		&variablesJSON, &template.CreatedAt, &template.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	if variablesJSON != "" {
		if err := json.Unmarshal([]byte(variablesJSON), &template.Variables); err != nil {
			log.Warnf("[Template] Failed to unmarshal variables: %v", err)
			template.Variables = []string{}
		}
	}

	return &template, nil
}

func (s *Service) GetAll() ([]MessageTemplate, error) {
	query := `SELECT id, name, description, content, variables, created_at, updated_at
	          FROM message_templates ORDER BY created_at DESC`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get templates: %w", err)
	}
	defer rows.Close()

	var templates []MessageTemplate
	for rows.Next() {
		var template MessageTemplate
		var variablesJSON string
		
		err := rows.Scan(
			&template.ID, &template.Name, &template.Description, &template.Content,
			&variablesJSON, &template.CreatedAt, &template.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan template: %w", err)
		}

		if variablesJSON != "" {
			if err := json.Unmarshal([]byte(variablesJSON), &template.Variables); err != nil {
				log.Warnf("[Template] Failed to unmarshal variables: %v", err)
				template.Variables = []string{}
			}
		}

		templates = append(templates, template)
	}

	return templates, nil
}

func (s *Service) Update(id int, req UpdateTemplateRequest) (*MessageTemplate, error) {
	template, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}
	if template == nil {
		return nil, fmt.Errorf("template not found")
	}

	updates := []string{}
	args := []interface{}{}
	argIndex := 1

	if req.Name != nil {
		if s.isPostgres {
			updates = append(updates, fmt.Sprintf("name = $%d", argIndex))
		} else {
			updates = append(updates, "name = ?")
		}
		args = append(args, *req.Name)
		argIndex++
	}

	if req.Description != nil {
		if s.isPostgres {
			updates = append(updates, fmt.Sprintf("description = $%d", argIndex))
		} else {
			updates = append(updates, "description = ?")
		}
		args = append(args, *req.Description)
		argIndex++
	}

	if req.Content != nil {
		if s.isPostgres {
			updates = append(updates, fmt.Sprintf("content = $%d", argIndex))
		} else {
			updates = append(updates, "content = ?")
		}
		args = append(args, *req.Content)
		argIndex++

		variables := extractVariables(*req.Content)
		variablesJSON, err := json.Marshal(variables)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal variables: %w", err)
		}

		if s.isPostgres {
			updates = append(updates, fmt.Sprintf("variables = $%d", argIndex))
		} else {
			updates = append(updates, "variables = ?")
		}
		args = append(args, string(variablesJSON))
		argIndex++
	}

	if len(updates) == 0 {
		return template, nil
	}

	now := time.Now()
	if s.isPostgres {
		updates = append(updates, fmt.Sprintf("updated_at = $%d", argIndex))
	} else {
		updates = append(updates, "updated_at = ?")
	}
	args = append(args, now)
	argIndex++

	args = append(args, id)

	var query string
	if s.isPostgres {
		query = fmt.Sprintf("UPDATE message_templates SET %s WHERE id = $%d", 
			strings.Join(updates, ", "), argIndex)
	} else {
		query = fmt.Sprintf("UPDATE message_templates SET %s WHERE id = ?", 
			strings.Join(updates, ", "))
	}

	_, err = s.db.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update template: %w", err)
	}

	return s.GetByID(id)
}

func (s *Service) Delete(id int) error {
	query := `DELETE FROM message_templates WHERE id = ?`
	
	if s.isPostgres {
		query = `DELETE FROM message_templates WHERE id = $1`
	}

	result, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete template: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("template not found")
	}

	return nil
}

func (s *Service) ApplyTemplate(templateID int, recipientData RecipientData) (string, error) {
	template, err := s.GetByID(templateID)
	if err != nil {
		return "", err
	}

	message := template.Content

	for key, value := range recipientData.Variables {
		placeholder := fmt.Sprintf("{{%s}}", key)
		message = strings.ReplaceAll(message, placeholder, value)
	}

	message = strings.ReplaceAll(message, "{{phone}}", recipientData.Phone)

	return message, nil
}

func (s *Service) PreviewTemplate(templateID int, recipients []RecipientData) ([]TemplatePreview, error) {
	previews := make([]TemplatePreview, 0, len(recipients))

	for _, recipient := range recipients {
		message, err := s.ApplyTemplate(templateID, recipient)
		if err != nil {
			return nil, err
		}

		previews = append(previews, TemplatePreview{
			Phone:   recipient.Phone,
			Message: message,
		})
	}

	return previews, nil
}

func extractVariables(content string) []string {
	re := regexp.MustCompile(`\{\{([^}]+)\}\}`)
	matches := re.FindAllStringSubmatch(content, -1)

	variableMap := make(map[string]bool)
	for _, match := range matches {
		if len(match) > 1 {
			variableMap[match[1]] = true
		}
	}

	variables := make([]string, 0, len(variableMap))
	for variable := range variableMap {
		variables = append(variables, variable)
	}

	return variables
}
