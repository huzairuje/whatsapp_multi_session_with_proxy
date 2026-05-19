package warmup

import (
	"database/sql"
	"fmt"
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
		return nil, fmt.Errorf("failed to create warmup table: %w", err)
	}

	return s, nil
}

func (s *Service) createTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS warmup_configs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		sender_jid TEXT NOT NULL UNIQUE,
		enabled INTEGER NOT NULL DEFAULT 0,
		current_day INTEGER NOT NULL DEFAULT 1,
		start_date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		daily_limit INTEGER NOT NULL DEFAULT 5,
		increment_amount INTEGER NOT NULL DEFAULT 5,
		increment_days INTEGER NOT NULL DEFAULT 3,
		max_daily_limit INTEGER NOT NULL DEFAULT 1000,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`

	if s.isPostgres {
		query = `
		CREATE TABLE IF NOT EXISTS warmup_configs (
			id SERIAL PRIMARY KEY,
			sender_jid TEXT NOT NULL UNIQUE,
			enabled BOOLEAN NOT NULL DEFAULT FALSE,
			current_day INTEGER NOT NULL DEFAULT 1,
			start_date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			daily_limit INTEGER NOT NULL DEFAULT 5,
			increment_amount INTEGER NOT NULL DEFAULT 5,
			increment_days INTEGER NOT NULL DEFAULT 3,
			max_daily_limit INTEGER NOT NULL DEFAULT 1000,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`
	}

	_, err := s.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create warmup_configs table: %w", err)
	}

	log.Info("[WarmUp] warmup_configs table ready")
	return nil
}

func (s *Service) Create(req CreateWarmUpRequest) (*WarmUpConfig, error) {
	now := time.Now()
	
	var query string
	var id int64

	if s.isPostgres {
		query = `
		INSERT INTO warmup_configs (sender_jid, enabled, current_day, start_date, daily_limit, increment_amount, increment_days, max_daily_limit, created_at, updated_at)
		VALUES ($1, $2, 1, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id`
		
		err := s.db.QueryRow(query, req.SenderJID, req.Enabled, now, req.DailyLimit, req.IncrementAmount, req.IncrementDays, req.MaxDailyLimit, now, now).Scan(&id)
		if err != nil {
			return nil, fmt.Errorf("failed to create warmup config: %w", err)
		}
	} else {
		query = `
		INSERT INTO warmup_configs (sender_jid, enabled, current_day, start_date, daily_limit, increment_amount, increment_days, max_daily_limit, created_at, updated_at)
		VALUES (?, ?, 1, ?, ?, ?, ?, ?, ?, ?)`
		
		result, err := s.db.Exec(query, req.SenderJID, req.Enabled, now, req.DailyLimit, req.IncrementAmount, req.IncrementDays, req.MaxDailyLimit, now, now)
		if err != nil {
			return nil, fmt.Errorf("failed to create warmup config: %w", err)
		}
		
		id, err = result.LastInsertId()
		if err != nil {
			return nil, fmt.Errorf("failed to get last insert id: %w", err)
		}
	}

	return s.GetByID(int(id))
}

func (s *Service) GetByID(id int) (*WarmUpConfig, error) {
	var config WarmUpConfig
	
	query := `SELECT id, sender_jid, enabled, current_day, start_date, daily_limit, increment_amount, increment_days, max_daily_limit, created_at, updated_at
	          FROM warmup_configs WHERE id = ?`
	
	if s.isPostgres {
		query = `SELECT id, sender_jid, enabled, current_day, start_date, daily_limit, increment_amount, increment_days, max_daily_limit, created_at, updated_at
		         FROM warmup_configs WHERE id = $1`
	}

	err := s.db.QueryRow(query, id).Scan(
		&config.ID, &config.SenderJID, &config.Enabled, &config.CurrentDay, &config.StartDate,
		&config.DailyLimit, &config.IncrementAmount, &config.IncrementDays, &config.MaxDailyLimit,
		&config.CreatedAt, &config.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("warmup config not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get warmup config: %w", err)
	}

	return &config, nil
}

func (s *Service) GetBySenderJID(senderJID string) (*WarmUpConfig, error) {
	var config WarmUpConfig
	
	query := `SELECT id, sender_jid, enabled, current_day, start_date, daily_limit, increment_amount, increment_days, max_daily_limit, created_at, updated_at
	          FROM warmup_configs WHERE sender_jid = ?`
	
	if s.isPostgres {
		query = `SELECT id, sender_jid, enabled, current_day, start_date, daily_limit, increment_amount, increment_days, max_daily_limit, created_at, updated_at
		         FROM warmup_configs WHERE sender_jid = $1`
	}

	err := s.db.QueryRow(query, senderJID).Scan(
		&config.ID, &config.SenderJID, &config.Enabled, &config.CurrentDay, &config.StartDate,
		&config.DailyLimit, &config.IncrementAmount, &config.IncrementDays, &config.MaxDailyLimit,
		&config.CreatedAt, &config.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get warmup config: %w", err)
	}

	return &config, nil
}

func (s *Service) GetAll() ([]WarmUpConfig, error) {
	query := `SELECT id, sender_jid, enabled, current_day, start_date, daily_limit, increment_amount, increment_days, max_daily_limit, created_at, updated_at
	          FROM warmup_configs ORDER BY created_at DESC`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get warmup configs: %w", err)
	}
	defer rows.Close()

	var configs []WarmUpConfig
	for rows.Next() {
		var config WarmUpConfig
		err := rows.Scan(
			&config.ID, &config.SenderJID, &config.Enabled, &config.CurrentDay, &config.StartDate,
			&config.DailyLimit, &config.IncrementAmount, &config.IncrementDays, &config.MaxDailyLimit,
			&config.CreatedAt, &config.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan warmup config: %w", err)
		}
		configs = append(configs, config)
	}

	return configs, nil
}

func (s *Service) Update(senderJID string, req UpdateWarmUpRequest) (*WarmUpConfig, error) {
	config, err := s.GetBySenderJID(senderJID)
	if err != nil {
		return nil, err
	}
	if config == nil {
		return nil, fmt.Errorf("warmup config not found for sender %s", senderJID)
	}

	updates := []string{}
	args := []interface{}{}
	argIndex := 1

	if req.Enabled != nil {
		if s.isPostgres {
			updates = append(updates, fmt.Sprintf("enabled = $%d", argIndex))
		} else {
			updates = append(updates, "enabled = ?")
		}
		args = append(args, *req.Enabled)
		argIndex++
	}

	if req.DailyLimit != nil {
		if s.isPostgres {
			updates = append(updates, fmt.Sprintf("daily_limit = $%d", argIndex))
		} else {
			updates = append(updates, "daily_limit = ?")
		}
		args = append(args, *req.DailyLimit)
		argIndex++
	}

	if req.IncrementAmount != nil {
		if s.isPostgres {
			updates = append(updates, fmt.Sprintf("increment_amount = $%d", argIndex))
		} else {
			updates = append(updates, "increment_amount = ?")
		}
		args = append(args, *req.IncrementAmount)
		argIndex++
	}

	if req.IncrementDays != nil {
		if s.isPostgres {
			updates = append(updates, fmt.Sprintf("increment_days = $%d", argIndex))
		} else {
			updates = append(updates, "increment_days = ?")
		}
		args = append(args, *req.IncrementDays)
		argIndex++
	}

	if req.MaxDailyLimit != nil {
		if s.isPostgres {
			updates = append(updates, fmt.Sprintf("max_daily_limit = $%d", argIndex))
		} else {
			updates = append(updates, "max_daily_limit = ?")
		}
		args = append(args, *req.MaxDailyLimit)
		argIndex++
	}

	if len(updates) == 0 {
		return config, nil
	}

	now := time.Now()
	if s.isPostgres {
		updates = append(updates, fmt.Sprintf("updated_at = $%d", argIndex))
	} else {
		updates = append(updates, "updated_at = ?")
	}
	args = append(args, now)
	argIndex++

	args = append(args, senderJID)

	var query string
	if s.isPostgres {
		query = fmt.Sprintf("UPDATE warmup_configs SET %s WHERE sender_jid = $%d", 
			joinUpdates(updates), argIndex)
	} else {
		query = fmt.Sprintf("UPDATE warmup_configs SET %s WHERE sender_jid = ?", 
			joinUpdates(updates))
	}

	_, err = s.db.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update warmup config: %w", err)
	}

	return s.GetBySenderJID(senderJID)
}

func (s *Service) Delete(senderJID string) error {
	query := `DELETE FROM warmup_configs WHERE sender_jid = ?`
	
	if s.isPostgres {
		query = `DELETE FROM warmup_configs WHERE sender_jid = $1`
	}

	result, err := s.db.Exec(query, senderJID)
	if err != nil {
		return fmt.Errorf("failed to delete warmup config: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("warmup config not found for sender %s", senderJID)
	}

	return nil
}

func (s *Service) GetCurrentDailyLimit(senderJID string) (int, error) {
	config, err := s.GetBySenderJID(senderJID)
	if err != nil {
		return 0, err
	}
	if config == nil || !config.Enabled {
		return 0, nil
	}

	daysSinceStart := int(time.Since(config.StartDate).Hours() / 24)
	
	incrementsApplied := daysSinceStart / config.IncrementDays
	
	currentLimit := config.DailyLimit + (incrementsApplied * config.IncrementAmount)
	
	if currentLimit > config.MaxDailyLimit {
		currentLimit = config.MaxDailyLimit
	}

	if daysSinceStart != config.CurrentDay {
		s.updateCurrentDay(senderJID, daysSinceStart)
	}

	return currentLimit, nil
}

func (s *Service) updateCurrentDay(senderJID string, currentDay int) error {
	query := `UPDATE warmup_configs SET current_day = ?, updated_at = ? WHERE sender_jid = ?`
	
	if s.isPostgres {
		query = `UPDATE warmup_configs SET current_day = $1, updated_at = $2 WHERE sender_jid = $3`
	}

	_, err := s.db.Exec(query, currentDay, time.Now(), senderJID)
	return err
}

func joinUpdates(updates []string) string {
	result := ""
	for i, update := range updates {
		if i > 0 {
			result += ", "
		}
		result += update
	}
	return result
}
