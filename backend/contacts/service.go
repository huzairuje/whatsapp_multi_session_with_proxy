package contacts

import (
	"database/sql"
	"fmt"
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
		return nil, fmt.Errorf("failed to create contacts table: %w", err)
	}

	return s, nil
}

func (s *Service) createTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS contacts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		sender_jid TEXT NOT NULL,
		contact_jid TEXT NOT NULL,
		contact_name TEXT,
		push_name TEXT,
		business_name TEXT,
		first_name TEXT,
		full_name TEXT,
		is_blocked INTEGER NOT NULL DEFAULT 0,
		is_business INTEGER NOT NULL DEFAULT 0,
		is_enterprise INTEGER NOT NULL DEFAULT 0,
		last_synced_at TIMESTAMP,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(sender_jid, contact_jid)
	)`

	if s.isPostgres {
		query = `
		CREATE TABLE IF NOT EXISTS contacts (
			id SERIAL PRIMARY KEY,
			sender_jid TEXT NOT NULL,
			contact_jid TEXT NOT NULL,
			contact_name TEXT,
			push_name TEXT,
			business_name TEXT,
			first_name TEXT,
			full_name TEXT,
			is_blocked BOOLEAN NOT NULL DEFAULT FALSE,
			is_business BOOLEAN NOT NULL DEFAULT FALSE,
			is_enterprise BOOLEAN NOT NULL DEFAULT FALSE,
			last_synced_at TIMESTAMP,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(sender_jid, contact_jid)
		)`
	}

	_, err := s.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create contacts table: %w", err)
	}

	log.Info("[Contacts] contacts table ready")
	return nil
}

func (s *Service) UpsertContact(contact Contact) error {
	now := time.Now()

	query := `
	INSERT INTO contacts (sender_jid, contact_jid, contact_name, push_name, business_name, first_name, full_name, is_blocked, is_business, is_enterprise, last_synced_at, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(sender_jid, contact_jid) DO UPDATE SET
		contact_name = excluded.contact_name,
		push_name = excluded.push_name,
		business_name = excluded.business_name,
		first_name = excluded.first_name,
		full_name = excluded.full_name,
		is_blocked = excluded.is_blocked,
		is_business = excluded.is_business,
		is_enterprise = excluded.is_enterprise,
		last_synced_at = excluded.last_synced_at,
		updated_at = excluded.updated_at`

	if s.isPostgres {
		query = `
		INSERT INTO contacts (sender_jid, contact_jid, contact_name, push_name, business_name, first_name, full_name, is_blocked, is_business, is_enterprise, last_synced_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT(sender_jid, contact_jid) DO UPDATE SET
			contact_name = excluded.contact_name,
			push_name = excluded.push_name,
			business_name = excluded.business_name,
			first_name = excluded.first_name,
			full_name = excluded.full_name,
			is_blocked = excluded.is_blocked,
			is_business = excluded.is_business,
			is_enterprise = excluded.is_enterprise,
			last_synced_at = excluded.last_synced_at,
			updated_at = excluded.updated_at`
	}

	_, err := s.db.Exec(query, contact.SenderJID, contact.ContactJID, contact.ContactName, contact.PushName, contact.BusinessName, contact.FirstName, contact.FullName, contact.IsBlocked, contact.IsBusiness, contact.IsEnterprise, now, now, now)
	return err
}

func (s *Service) GetContactsBySender(senderJID string, limit int, offset int) ([]Contact, error) {
	query := `SELECT id, sender_jid, contact_jid, contact_name, push_name, business_name, first_name, full_name, is_blocked, is_business, is_enterprise, last_synced_at, created_at, updated_at
	          FROM contacts WHERE sender_jid = ? ORDER BY contact_name ASC LIMIT ? OFFSET ?`

	if s.isPostgres {
		query = `SELECT id, sender_jid, contact_jid, contact_name, push_name, business_name, first_name, full_name, is_blocked, is_business, is_enterprise, last_synced_at, created_at, updated_at
		         FROM contacts WHERE sender_jid = $1 ORDER BY contact_name ASC LIMIT $2 OFFSET $3`
	}

	rows, err := s.db.Query(query, senderJID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get contacts: %w", err)
	}
	defer rows.Close()

	var contacts []Contact
	for rows.Next() {
		var contact Contact
		err := rows.Scan(&contact.ID, &contact.SenderJID, &contact.ContactJID, &contact.ContactName, &contact.PushName, &contact.BusinessName, &contact.FirstName, &contact.FullName, &contact.IsBlocked, &contact.IsBusiness, &contact.IsEnterprise, &contact.LastSyncedAt, &contact.CreatedAt, &contact.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan contact: %w", err)
		}
		contacts = append(contacts, contact)
	}

	return contacts, nil
}

func (s *Service) SearchContacts(filter ContactFilter) ([]Contact, error) {
	query := `SELECT id, sender_jid, contact_jid, contact_name, push_name, business_name, first_name, full_name, is_blocked, is_business, is_enterprise, last_synced_at, created_at, updated_at FROM contacts WHERE sender_jid = ?`
	args := []interface{}{filter.SenderJID}

	if filter.SearchQuery != "" {
		query += ` AND (contact_name LIKE ? OR push_name LIKE ? OR contact_jid LIKE ?)`
		searchTerm := "%" + filter.SearchQuery + "%"
		args = append(args, searchTerm, searchTerm, searchTerm)
	}

	if filter.IsBlocked != nil {
		query += ` AND is_blocked = ?`
		args = append(args, *filter.IsBlocked)
	}

	if filter.IsBusiness != nil {
		query += ` AND is_business = ?`
		args = append(args, *filter.IsBusiness)
	}

	query += ` ORDER BY contact_name ASC LIMIT ? OFFSET ?`
	args = append(args, filter.Limit, filter.Offset)

	if s.isPostgres {
		query = strings.ReplaceAll(query, "?", "$")
		for i := 1; i <= len(args); i++ {
			query = strings.Replace(query, "$", fmt.Sprintf("$%d", i), 1)
		}
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search contacts: %w", err)
	}
	defer rows.Close()

	var contacts []Contact
	for rows.Next() {
		var contact Contact
		err := rows.Scan(&contact.ID, &contact.SenderJID, &contact.ContactJID, &contact.ContactName, &contact.PushName, &contact.BusinessName, &contact.FirstName, &contact.FullName, &contact.IsBlocked, &contact.IsBusiness, &contact.IsEnterprise, &contact.LastSyncedAt, &contact.CreatedAt, &contact.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan contact: %w", err)
		}
		contacts = append(contacts, contact)
	}

	return contacts, nil
}

func (s *Service) DeleteContact(senderJID, contactJID string) error {
	query := `DELETE FROM contacts WHERE sender_jid = ? AND contact_jid = ?`

	if s.isPostgres {
		query = `DELETE FROM contacts WHERE sender_jid = $1 AND contact_jid = $2`
	}

	_, err := s.db.Exec(query, senderJID, contactJID)
	return err
}

func (s *Service) GetContactCount(senderJID string) (int, error) {
	query := `SELECT COUNT(*) FROM contacts WHERE sender_jid = ?`

	if s.isPostgres {
		query = `SELECT COUNT(*) FROM contacts WHERE sender_jid = $1`
	}

	var count int
	err := s.db.QueryRow(query, senderJID).Scan(&count)
	return count, err
}
