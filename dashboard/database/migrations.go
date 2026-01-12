package database

import (
	"database/sql"
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"whatsapp_multi_session/dashboard/models" // Assuming 'whatsapp_multi_session' is the module name
)

// RunDashboardMigrations performs GORM auto-migration for the dashboard related tables.
// It accepts an existing database connection and the dialect ("postgres" or "sqlite").
func RunDashboardMigrations(db *sql.DB, dialect string) error {
	var gormDB *gorm.DB
	var err error

	log.Printf("Initializing GORM for dialect: %s", dialect)

	if dialect == "postgres" {
		gormDB, err = gorm.Open(postgres.New(postgres.Config{
			Conn: db,
		}), &gorm.Config{})
	} else if dialect == "sqlite" {
		// Using sqlite.Dialector{Conn: db} for an existing *sql.DB connection
		gormDB, err = gorm.Open(sqlite.Dialector{Conn: db}, &gorm.Config{})
	} else {
		errMsg := fmt.Sprintf("unsupported dialect for GORM migration: %s", dialect)
		log.Println(errMsg)
		return fmt.Errorf(errMsg)
	}

	if err != nil {
		log.Printf("Failed to initialize GORM: %v", err)
		return fmt.Errorf("failed to initialize GORM with dialect %s: %w", dialect, err)
	}

	log.Println("Running GORM AutoMigrate for dashboard models...")
	err = gormDB.AutoMigrate(
		&models.AdminUser{},
		&models.ManagedDevice{},
		&models.SentMessage{},
	)
	if err != nil {
		log.Printf("Failed to AutoMigrate dashboard models: %v", err)
		return fmt.Errorf("failed to AutoMigrate dashboard models: %w", err)
	}

	log.Println("Dashboard migrations completed successfully.")
	return nil
}
