package message

import (
	"database/sql"
	"time"
)

type Repository struct {
	db         *sql.DB
	isPostgres bool
}

func NewRepository(db *sql.DB, isPostgres bool) *Repository {
	return &Repository{
		db:         db,
		isPostgres: isPostgres,
	}
}

func (r *Repository) CreateMessagesTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			sender TEXT NOT NULL,
			recipient TEXT NOT NULL,
			content TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			message_id TEXT,
			error TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_messages_sender ON messages(sender);
		CREATE INDEX IF NOT EXISTS idx_messages_status ON messages(status);
		CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at);
		CREATE INDEX IF NOT EXISTS idx_messages_message_id ON messages(message_id);
	`
	_, err := r.db.Exec(query)
	return err
}

func (r *Repository) CreateMessagesTablePostgres() error {
	query := `
		CREATE TABLE IF NOT EXISTS messages (
			id SERIAL PRIMARY KEY,
			sender VARCHAR(255) NOT NULL,
			recipient VARCHAR(255) NOT NULL,
			content TEXT NOT NULL,
			status VARCHAR(50) NOT NULL DEFAULT 'pending',
			message_id VARCHAR(255),
			error TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_messages_sender ON messages(sender);
		CREATE INDEX IF NOT EXISTS idx_messages_status ON messages(status);
		CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at);
		CREATE INDEX IF NOT EXISTS idx_messages_message_id ON messages(message_id);
	`
	_, err := r.db.Exec(query)
	return err
}

func (r *Repository) InsertMessage(msg *Message) error {
	if r.isPostgres {
		query := `
			INSERT INTO messages (sender, recipient, content, status, message_id, error, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			RETURNING id
		`
		return r.db.QueryRow(query, msg.Sender, msg.Recipient, msg.Content, msg.Status, msg.MessageID, msg.Error, msg.CreatedAt, msg.UpdatedAt).Scan(&msg.ID)
	}

	query := `
		INSERT INTO messages (sender, recipient, content, status, message_id, error, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.Exec(query, msg.Sender, msg.Recipient, msg.Content, msg.Status, msg.MessageID, msg.Error, msg.CreatedAt, msg.UpdatedAt)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	msg.ID = id
	return nil
}

func (r *Repository) UpdateMessageStatus(messageID string, status MessageStatus, errorMsg string) error {
	if r.isPostgres {
		query := `UPDATE messages SET status = $1, error = $2, updated_at = $3 WHERE message_id = $4`
		_, err := r.db.Exec(query, status, errorMsg, time.Now(), messageID)
		return err
	}

	query := `UPDATE messages SET status = ?, error = ?, updated_at = ? WHERE message_id = ?`
	_, err := r.db.Exec(query, status, errorMsg, time.Now(), messageID)
	return err
}

func (r *Repository) GetMessageByID(id int64) (*Message, error) {
	msg := &Message{}
	var query string
	if r.isPostgres {
		query = `SELECT id, sender, recipient, content, status, message_id, error, created_at, updated_at FROM messages WHERE id = $1`
	} else {
		query = `SELECT id, sender, recipient, content, status, message_id, error, created_at, updated_at FROM messages WHERE id = ?`
	}

	err := r.db.QueryRow(query, id).Scan(&msg.ID, &msg.Sender, &msg.Recipient, &msg.Content, &msg.Status, &msg.MessageID, &msg.Error, &msg.CreatedAt, &msg.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return msg, nil
}

func (r *Repository) GetMessagesBySender(sender string, limit, offset int) ([]*Message, error) {
	var query string
	if r.isPostgres {
		query = `SELECT id, sender, recipient, content, status, message_id, error, created_at, updated_at 
				 FROM messages WHERE sender = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	} else {
		query = `SELECT id, sender, recipient, content, status, message_id, error, created_at, updated_at 
				 FROM messages WHERE sender = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`
	}

	rows, err := r.db.Query(query, sender, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := []*Message{}
	for rows.Next() {
		msg := &Message{}
		err := rows.Scan(&msg.ID, &msg.Sender, &msg.Recipient, &msg.Content, &msg.Status, &msg.MessageID, &msg.Error, &msg.CreatedAt, &msg.UpdatedAt)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}
	return messages, nil
}

func (r *Repository) GetStatsBySender(sender string) (*MessageStats, error) {
	stats := &MessageStats{}

	var totalQuery, failedQuery, deliveredQuery, dailyQuery string
	if r.isPostgres {
		totalQuery = `SELECT COUNT(*) FROM messages WHERE sender = $1 AND status IN ('sent', 'delivered', 'read')`
		failedQuery = `SELECT COUNT(*) FROM messages WHERE sender = $1 AND status = 'failed'`
		deliveredQuery = `SELECT COUNT(*) FROM messages WHERE sender = $1 AND status IN ('delivered', 'read')`
		dailyQuery = `SELECT COUNT(*) FROM messages WHERE sender = $1 AND DATE(created_at) = CURRENT_DATE`
	} else {
		totalQuery = `SELECT COUNT(*) FROM messages WHERE sender = ? AND status IN ('sent', 'delivered', 'read')`
		failedQuery = `SELECT COUNT(*) FROM messages WHERE sender = ? AND status = 'failed'`
		deliveredQuery = `SELECT COUNT(*) FROM messages WHERE sender = ? AND status IN ('delivered', 'read')`
		dailyQuery = `SELECT COUNT(*) FROM messages WHERE sender = ? AND DATE(created_at) = DATE('now')`
	}

	err := r.db.QueryRow(totalQuery, sender).Scan(&stats.TotalSent)
	if err != nil {
		return nil, err
	}

	err = r.db.QueryRow(failedQuery, sender).Scan(&stats.TotalFailed)
	if err != nil {
		return nil, err
	}

	err = r.db.QueryRow(deliveredQuery, sender).Scan(&stats.TotalDelivered)
	if err != nil {
		return nil, err
	}

	err = r.db.QueryRow(dailyQuery, sender).Scan(&stats.DailyCount)
	if err != nil {
		return nil, err
	}

	total := stats.TotalSent + stats.TotalFailed
	if total > 0 {
		stats.SuccessRate = float64(stats.TotalSent) / float64(total) * 100
	}

	stats.DailyLimit = 50

	return stats, nil
}

func (r *Repository) GetAllStats() (*MessageStats, error) {
	stats := &MessageStats{}

	var totalQuery, failedQuery, deliveredQuery, dailyQuery string
	if r.isPostgres {
		totalQuery = `SELECT COUNT(*) FROM messages WHERE status IN ('sent', 'delivered', 'read')`
		failedQuery = `SELECT COUNT(*) FROM messages WHERE status = 'failed'`
		deliveredQuery = `SELECT COUNT(*) FROM messages WHERE status IN ('delivered', 'read')`
		dailyQuery = `SELECT COUNT(*) FROM messages WHERE DATE(created_at) = CURRENT_DATE`
	} else {
		totalQuery = `SELECT COUNT(*) FROM messages WHERE status IN ('sent', 'delivered', 'read')`
		failedQuery = `SELECT COUNT(*) FROM messages WHERE status = 'failed'`
		deliveredQuery = `SELECT COUNT(*) FROM messages WHERE status IN ('delivered', 'read')`
		dailyQuery = `SELECT COUNT(*) FROM messages WHERE DATE(created_at) = DATE('now')`
	}

	err := r.db.QueryRow(totalQuery).Scan(&stats.TotalSent)
	if err != nil {
		return nil, err
	}

	err = r.db.QueryRow(failedQuery).Scan(&stats.TotalFailed)
	if err != nil {
		return nil, err
	}

	err = r.db.QueryRow(deliveredQuery).Scan(&stats.TotalDelivered)
	if err != nil {
		return nil, err
	}

	err = r.db.QueryRow(dailyQuery).Scan(&stats.DailyCount)
	if err != nil {
		return nil, err
	}

	total := stats.TotalSent + stats.TotalFailed
	if total > 0 {
		stats.SuccessRate = float64(stats.TotalSent) / float64(total) * 100
	}

	return stats, nil
}
