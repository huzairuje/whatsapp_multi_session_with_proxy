package activity

import (
	"database/sql"
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

func (r *Repository) CreateActivitiesTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS activities (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			type TEXT NOT NULL,
			sender TEXT,
			user TEXT,
			message TEXT NOT NULL,
			details TEXT,
			status TEXT,
			error TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_activities_type ON activities(type);
		CREATE INDEX IF NOT EXISTS idx_activities_sender ON activities(sender);
		CREATE INDEX IF NOT EXISTS idx_activities_created_at ON activities(created_at);
	`
	_, err := r.db.Exec(query)
	return err
}

func (r *Repository) CreateActivitiesTablePostgres() error {
	query := `
		CREATE TABLE IF NOT EXISTS activities (
			id SERIAL PRIMARY KEY,
			type VARCHAR(100) NOT NULL,
			sender VARCHAR(255),
			user VARCHAR(255),
			message TEXT NOT NULL,
			details TEXT,
			status VARCHAR(50),
			error TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_activities_type ON activities(type);
		CREATE INDEX IF NOT EXISTS idx_activities_sender ON activities(sender);
		CREATE INDEX IF NOT EXISTS idx_activities_created_at ON activities(created_at);
	`
	_, err := r.db.Exec(query)
	return err
}

func (r *Repository) InsertActivity(activity *Activity) error {
	if r.isPostgres {
		query := `
			INSERT INTO activities (type, sender, user, message, details, status, error, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			RETURNING id
		`
		return r.db.QueryRow(query, activity.Type, activity.Sender, activity.User, activity.Message, 
			activity.Details, activity.Status, activity.ErrorMsg, activity.CreatedAt).Scan(&activity.ID)
	}

	query := `
		INSERT INTO activities (type, sender, user, message, details, status, error, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.Exec(query, activity.Type, activity.Sender, activity.User, activity.Message,
		activity.Details, activity.Status, activity.ErrorMsg, activity.CreatedAt)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	activity.ID = id
	return nil
}

func (r *Repository) GetRecentActivities(limit int) ([]*Activity, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 500 {
		limit = 500
	}

	var query string
	if r.isPostgres {
		query = `SELECT id, type, sender, user, message, details, status, error, created_at 
				 FROM activities ORDER BY created_at DESC LIMIT $1`
	} else {
		query = `SELECT id, type, sender, user, message, details, status, error, created_at 
				 FROM activities ORDER BY created_at DESC LIMIT ?`
	}

	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	activities := []*Activity{}
	for rows.Next() {
		a := &Activity{}
		err := rows.Scan(&a.ID, &a.Type, &a.Sender, &a.User, &a.Message, &a.Details, &a.Status, &a.ErrorMsg, &a.CreatedAt)
		if err != nil {
			return nil, err
		}
		activities = append(activities, a)
	}
	return activities, nil
}

func (r *Repository) GetActivitiesBySender(sender string, limit int) ([]*Activity, error) {
	if limit <= 0 {
		limit = 50
	}

	var query string
	if r.isPostgres {
		query = `SELECT id, type, sender, user, message, details, status, error, created_at 
				 FROM activities WHERE sender = $1 ORDER BY created_at DESC LIMIT $2`
	} else {
		query = `SELECT id, type, sender, user, message, details, status, error, created_at 
				 FROM activities WHERE sender = ? ORDER BY created_at DESC LIMIT ?`
	}

	rows, err := r.db.Query(query, sender, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	activities := []*Activity{}
	for rows.Next() {
		a := &Activity{}
		err := rows.Scan(&a.ID, &a.Type, &a.Sender, &a.User, &a.Message, &a.Details, &a.Status, &a.ErrorMsg, &a.CreatedAt)
		if err != nil {
			return nil, err
		}
		activities = append(activities, a)
	}
	return activities, nil
}

func (r *Repository) GetActivitiesByType(activityType ActivityType, limit int) ([]*Activity, error) {
	if limit <= 0 {
		limit = 50
	}

	var query string
	if r.isPostgres {
		query = `SELECT id, type, sender, user, message, details, status, error, created_at 
				 FROM activities WHERE type = $1 ORDER BY created_at DESC LIMIT $2`
	} else {
		query = `SELECT id, type, sender, user, message, details, status, error, created_at 
				 FROM activities WHERE type = ? ORDER BY created_at DESC LIMIT ?`
	}

	rows, err := r.db.Query(query, activityType, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	activities := []*Activity{}
	for rows.Next() {
		a := &Activity{}
		err := rows.Scan(&a.ID, &a.Type, &a.Sender, &a.User, &a.Message, &a.Details, &a.Status, &a.ErrorMsg, &a.CreatedAt)
		if err != nil {
			return nil, err
		}
		activities = append(activities, a)
	}
	return activities, nil
}

func (r *Repository) GetStats() (*ActivityStats, error) {
	stats := &ActivityStats{
		ActivitiesByType: make(map[string]int64),
	}

	// Total activities
	var totalQuery string
	if r.isPostgres {
		totalQuery = `SELECT COUNT(*) FROM activities`
	} else {
		totalQuery = `SELECT COUNT(*) FROM activities`
	}
	err := r.db.QueryRow(totalQuery).Scan(&stats.TotalActivities)
	if err != nil {
		return nil, err
	}

	// Activities by type
	var typeQuery string
	if r.isPostgres {
		typeQuery = `SELECT type, COUNT(*) FROM activities GROUP BY type`
	} else {
		typeQuery = `SELECT type, COUNT(*) FROM activities GROUP BY type`
	}
	rows, err := r.db.Query(typeQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var actType string
		var count int64
		if err := rows.Scan(&actType, &count); err != nil {
			return nil, err
		}
		stats.ActivitiesByType[actType] = count
	}

	// Recent activities
	stats.RecentActivities, _ = r.GetRecentActivities(10)

	// Specific counts
	stats.SessionsConnected = stats.ActivitiesByType[string(TypeSessionConnect)]
	stats.MessagesSent = stats.ActivitiesByType[string(TypeMessageSent)]
	stats.MessagesFailed = stats.ActivitiesByType[string(TypeMessageFailed)]
	stats.RateLimitEvents = stats.ActivitiesByType[string(TypeRateLimit)]

	return stats, nil
}
