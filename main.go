package main

import (
	"whatsapp_multi_session/app/db"
	"whatsapp_multi_session/app/handlers"
	"whatsapp_multi_session/app/sessions"
	"fmt" // Now needed for server startup message
	"log"
	"net/http"

	"github.com/gorilla/mux"
	_ "gorm.io/driver/sqlite" // Ensure driver is included for database connection
)

func main() {
	db.ConnectDatabase()
	err := db.DB.AutoMigrate(&db.User{}, &db.Device{}, &db.MessageCount{})
	if err != nil {
		log.Fatalf("Failed to auto-migrate database: %v", err)
	}
	log.Println("Database auto-migration completed.")

	sessions.InitSessionStore()
	handlers.InitAuthTemplates()
	handlers.InitDashboardTemplates()
	handlers.InitDeviceTemplates()
	handlers.InitMessageTemplates()
	handlers.InitSettingsTemplates() // Initialize settings templates

	r := mux.NewRouter()

	r.HandleFunc("/", handlers.HomeHandler).Methods(http.MethodGet)

	r.HandleFunc("/register", handlers.ShowRegistrationPage).Methods(http.MethodGet)
	r.HandleFunc("/register", handlers.HandleRegistration).Methods(http.MethodPost)
	r.HandleFunc("/login", handlers.ShowLoginPage).Methods(http.MethodGet)
	r.HandleFunc("/login", handlers.HandleLogin).Methods(http.MethodPost)
	r.HandleFunc("/logout", handlers.LogoutHandler).Methods(http.MethodGet)

	r.HandleFunc("/devices", handlers.ShowDevicesPage).Methods(http.MethodGet)
	r.HandleFunc("/messages", handlers.ShowMessagesPage).Methods(http.MethodGet)

	// Settings page
	r.HandleFunc("/settings", handlers.SettingsPageGET).Methods(http.MethodGet)
	r.HandleFunc("/settings/password", handlers.HandleChangePassword).Methods(http.MethodPost)

	staticDir := "./app/static/" // Relative to project root
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))

	fmt.Println("Server starting on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", r))
}
