package activity

import (
	"database/sql"
	"time"
)

type Service struct {
	repo *Repository
}

func NewService(db *sql.DB, isPostgres bool) *Service {
	return &Service{
		repo: NewRepository(db, isPostgres),
	}
}

func (s *Service) InitializeDatabase() error {
	if s.repo.isPostgres {
		return s.repo.CreateActivitiesTablePostgres()
	}
	return s.repo.CreateActivitiesTable()
}

func (s *Service) LogActivity(actType ActivityType, message string, sender, user, details, status, errorMsg string) (*Activity, error) {
	activity := &Activity{
		Type:      actType,
		Sender:    sender,
		User:      user,
		Message:   message,
		Details:   details,
		Status:    status,
		ErrorMsg:  errorMsg,
		CreatedAt: time.Now(),
	}

	err := s.repo.InsertActivity(activity)
	if err != nil {
		return nil, err
	}

	return activity, nil
}

func (s *Service) GetRecentActivities(limit int) ([]*Activity, error) {
	return s.repo.GetRecentActivities(limit)
}

func (s *Service) GetActivitiesBySender(sender string, limit int) ([]*Activity, error) {
	return s.repo.GetActivitiesBySender(sender, limit)
}

func (s *Service) GetActivitiesByType(actType ActivityType, limit int) ([]*Activity, error) {
	return s.repo.GetActivitiesByType(actType, limit)
}

func (s *Service) GetStats() (*ActivityStats, error) {
	return s.repo.GetStats()
}
