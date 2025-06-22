package db

import (
	"log"
	"os"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB // Exported for access from other packages

// ConnectDashboardDB initializes the database for the dashboard features.
func ConnectDashboardDB() {
	var err error
	// Ensure path is relative to the executable's working directory at runtime.
	// For tests, this path might need adjustment or an alternative setup.
	err = os.MkdirAll("./database_files", 0755)
	if err != nil {
		log.Fatalf("Failed to create dashboard database directory: %v", err)
	}

	DB, err = gorm.Open(sqlite.Open("./database_files/dashboard.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to dashboard database: %v", err)
	}
	log.Println("Dashboard database connection successfully established.")
}

// MigrateDashboardSchema runs auto-migration for dashboard specific tables.
func MigrateDashboardSchema() {
	if DB == nil {
		log.Fatalf("Dashboard database not initialized before migration.")
		return
	}
	// Assumes User, Device, MessageCount structs are defined in models.go in this 'db' package.
	err := DB.AutoMigrate(&User{}, &Device{}, &MessageCount{})
	if err != nil {
		log.Fatalf("Failed to auto-migrate dashboard database schema: %v", err)
	}
	log.Println("Dashboard database auto-migration completed.")
}
