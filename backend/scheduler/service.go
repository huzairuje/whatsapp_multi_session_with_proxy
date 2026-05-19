package scheduler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand/v2"
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
		return nil, fmt.Errorf("failed to create scheduler table: %w", err)
	}

	return s, nil
}

func (s *Service) createTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS scheduled_jobs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		sender_jid TEXT NOT NULL,
		template_id INTEGER,
		recipients TEXT NOT NULL,
		message_variants TEXT,
		total_messages INTEGER NOT NULL,
		sent_messages INTEGER NOT NULL DEFAULT 0,
		status TEXT NOT NULL DEFAULT 'pending',
		scheduled_for TIMESTAMP NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`

	if s.isPostgres {
		query = `
		CREATE TABLE IF NOT EXISTS scheduled_jobs (
			id SERIAL PRIMARY KEY,
			sender_jid TEXT NOT NULL,
			template_id INTEGER,
			recipients TEXT NOT NULL,
			message_variants TEXT,
			total_messages INTEGER NOT NULL,
			sent_messages INTEGER NOT NULL DEFAULT 0,
			status TEXT NOT NULL DEFAULT 'pending',
			scheduled_for TIMESTAMP NOT NULL,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		)`
	}

	_, err := s.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create scheduled_jobs table: %w", err)
	}

	log.Info("[Scheduler] scheduled_jobs table ready")
	return nil
}

func (s *Service) ScheduleBulkSend(req CreateScheduledJobRequest, config ScheduleConfig) (*ScheduledJob, error) {
	if len(req.Recipients) <= 10 {
		return nil, fmt.Errorf("bulk send must have more than 10 recipients to use scheduler")
	}

	recipientsJSON, err := json.Marshal(req.Recipients)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal recipients: %w", err)
	}

	variantsJSON, err := json.Marshal(req.MessageVariants)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message variants: %w", err)
	}

	now := time.Now()
	scheduledFor := req.StartDate
	if scheduledFor.IsZero() {
		scheduledFor = s.getNextAllowedTime(now, config)
	}

	var query string
	var id int64

	if s.isPostgres {
		query = `
		INSERT INTO scheduled_jobs (sender_jid, template_id, recipients, message_variants, total_messages, status, scheduled_for, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, 'pending', $6, $7, $8)
		RETURNING id`

		err := s.db.QueryRow(query, req.SenderJID, req.TemplateID, string(recipientsJSON), string(variantsJSON), len(req.Recipients), scheduledFor, now, now).Scan(&id)
		if err != nil {
			return nil, fmt.Errorf("failed to create scheduled job: %w", err)
		}
	} else {
		query = `
		INSERT INTO scheduled_jobs (sender_jid, template_id, recipients, message_variants, total_messages, status, scheduled_for, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, 'pending', ?, ?, ?)`

		result, err := s.db.Exec(query, req.SenderJID, req.TemplateID, string(recipientsJSON), string(variantsJSON), len(req.Recipients), scheduledFor, now, now)
		if err != nil {
			return nil, fmt.Errorf("failed to create scheduled job: %w", err)
		}

		id, err = result.LastInsertId()
		if err != nil {
			return nil, fmt.Errorf("failed to get last insert id: %w", err)
		}
	}

	return s.GetByID(int(id))
}

func (s *Service) GetByID(id int) (*ScheduledJob, error) {
	var job ScheduledJob

	query := `SELECT id, sender_jid, template_id, recipients, message_variants, total_messages, sent_messages, status, scheduled_for, created_at, updated_at
	          FROM scheduled_jobs WHERE id = ?`

	if s.isPostgres {
		query = `SELECT id, sender_jid, template_id, recipients, message_variants, total_messages, sent_messages, status, scheduled_for, created_at, updated_at
		         FROM scheduled_jobs WHERE id = $1`
	}

	err := s.db.QueryRow(query, id).Scan(
		&job.ID, &job.SenderJID, &job.TemplateID, &job.Recipients, &job.MessageVariants,
		&job.TotalMessages, &job.SentMessages, &job.Status, &job.ScheduledFor, &job.CreatedAt, &job.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("scheduled job not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get scheduled job: %w", err)
	}

	return &job, nil
}

func (s *Service) GetPendingJobs() ([]ScheduledJob, error) {
	query := `SELECT id, sender_jid, template_id, recipients, message_variants, total_messages, sent_messages, status, scheduled_for, created_at, updated_at
	          FROM scheduled_jobs WHERE status = 'pending' AND scheduled_for <= ? ORDER BY scheduled_for ASC`

	if s.isPostgres {
		query = `SELECT id, sender_jid, template_id, recipients, message_variants, total_messages, sent_messages, status, scheduled_for, created_at, updated_at
		         FROM scheduled_jobs WHERE status = 'pending' AND scheduled_for <= $1 ORDER BY scheduled_for ASC`
	}

	rows, err := s.db.Query(query, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to get pending jobs: %w", err)
	}
	defer rows.Close()

	var jobs []ScheduledJob
	for rows.Next() {
		var job ScheduledJob
		err := rows.Scan(
			&job.ID, &job.SenderJID, &job.TemplateID, &job.Recipients, &job.MessageVariants,
			&job.TotalMessages, &job.SentMessages, &job.Status, &job.ScheduledFor, &job.CreatedAt, &job.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan scheduled job: %w", err)
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}

func (s *Service) UpdateJobStatus(id int, status string, sentMessages int) error {
	query := `UPDATE scheduled_jobs SET status = ?, sent_messages = ?, updated_at = ? WHERE id = ?`

	if s.isPostgres {
		query = `UPDATE scheduled_jobs SET status = $1, sent_messages = $2, updated_at = $3 WHERE id = $4`
	}

	_, err := s.db.Exec(query, status, sentMessages, time.Now(), id)
	return err
}

func (s *Service) getNextAllowedTime(from time.Time, config ScheduleConfig) time.Time {
	loc := time.Local
	if config.Timezone != "" && config.Timezone != "Local" {
		var err error
		loc, err = time.LoadLocation(config.Timezone)
		if err != nil {
			log.Warnf("[Scheduler] Invalid timezone %s, using Local", config.Timezone)
			loc = time.Local
		}
	}

	current := from.In(loc)
	hour := current.Hour()

	if hour >= config.AllowedHourStart && hour < config.AllowedHourEnd {
		return current
	}

	if hour < config.AllowedHourStart {
		return time.Date(current.Year(), current.Month(), current.Day(), config.AllowedHourStart, 0, 0, 0, loc)
	}

	tomorrow := current.AddDate(0, 0, 1)
	return time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), config.AllowedHourStart, 0, 0, 0, loc)
}

func (s *Service) CalculateSchedule(totalRecipients int, dailyLimit int, config ScheduleConfig) []time.Time {
	schedule := []time.Time{}
	rng := rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), uint64(time.Now().UnixNano()<<1)))

	now := time.Now()
	currentTime := s.getNextAllowedTime(now, config)
	sentToday := 0

	for len(schedule) < totalRecipients {
		if sentToday >= dailyLimit {
			currentTime = s.getNextAllowedTime(currentTime.AddDate(0, 0, 1), config)
			sentToday = 0
		}

		randomMinutes := rng.IntN(60)
		randomSeconds := rng.IntN(60)
		currentTime = currentTime.Add(time.Duration(randomMinutes)*time.Minute + time.Duration(randomSeconds)*time.Second)

		schedule = append(schedule, currentTime)
		sentToday++
	}

	return schedule
}

func (s *Service) GetMessageVariant(variants []string, index int) string {
	if len(variants) == 0 {
		return ""
	}
	return variants[index%len(variants)]
}
