package db

import (
	"log"
	"os"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {
	var err error
	// Ensure the directory for the database file exists
	err = os.MkdirAll("./database_files", 0755) // Storing db file in a sub-directory
	if err != nil {
		log.Fatalf("Failed to create database directory: %v", err)
	}

	DB, err = gorm.Open(sqlite.Open("./database_files/dashboard.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Database connection successfully established.")
}
