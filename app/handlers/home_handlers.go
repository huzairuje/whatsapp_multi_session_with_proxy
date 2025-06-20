package handlers

import (
	"whatsapp_multi_session/app/db"
	"whatsapp_multi_session/app/sessions"
	"html/template"
	"log"
	"net/http"

	"gorm.io/gorm" // Added for gorm.ErrRecordNotFound
)

var dashboardTmpl *template.Template

// InitDashboardTemplates pre-parses the dashboard templates.
func InitDashboardTemplates() {
	// Assuming templates are in 'app/templates/' relative to run directory
	// For ExecuteTemplate to find "layout.html" and then the content block from "index.html",
	// ParseFiles should include all files that define templates you want to be part of the *same* Template set.
	// The first file name becomes the name of the template if not otherwise specified by {{define "name"}}
	// It's often better to parse them into a map or use template.Must for clarity if templates are named.
	// For this setup, layout.html will be the entry point.
    // Adjusted paths for tests run from app/handlers/ directory.
	tpl, err := template.ParseFiles("../templates/layout.html", "../templates/index.html")
	if err != nil {
		log.Fatalf("Error parsing dashboard templates: %v", err)
	}
	dashboardTmpl = tpl
}

// DashboardData holds the data to be passed to the dashboard template.
type DashboardData struct {
	LoggedIn     bool
	UserID       uint
	Username     string
	DeviceCount  int64
	MessageCount int64 // Changed from int to int64 to match GORM count
	PageTitle    string // For setting title in layout
}

// HomeHandler serves the main dashboard page.
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	session, err := sessions.Store.Get(r, sessions.SessionKey)
	if err != nil {
		log.Printf("Error getting session for HomeHandler: %v. Redirecting to login.", err)
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	userIDVal, ok := session.Values[sessions.UserIDKey]
	if !ok {
		log.Println("UserID not found in session. Redirecting to login.")
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	userID, ok := userIDVal.(uint)
	if !ok || userID == 0 { // Ensure userID is of type uint and not zero
		log.Printf("UserID in session is invalid (not uint or zero): %v. Redirecting to login.", userIDVal)
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	var user db.User
	if err := db.DB.First(&user, userID).Error; err != nil {
		log.Printf("Error fetching user %d from DB: %v. Redirecting to login.", userID, err)
		http.Redirect(w, r, "/login", http.StatusFound) // Could be an error page too
		return
	}

	var deviceCount int64
	db.DB.Model(&db.Device{}).Where("user_id = ?", userID).Count(&deviceCount)

	var messageSummary db.MessageCount
	// Fetch the specific message count record for the user.
	if err := db.DB.Where("user_id = ?", userID).First(&messageSummary).Error; err != nil {
		// If no record found, it means count is 0. Log error for other cases.
		if err != gorm.ErrRecordNotFound {
			log.Printf("Error fetching message count for user %d: %v. Count will be displayed as 0.", userID, err)
		}
		// messageSummary.Count will be 0 by default if not found and struct is initialized
	}

	data := DashboardData{
		LoggedIn:     true,
		UserID:       userID,
		Username:     user.Username,
		DeviceCount:  deviceCount,
		MessageCount: int64(messageSummary.Count), // Ensure it's int64
		PageTitle:    "Main Dashboard",
	}

	if dashboardTmpl == nil {
		log.Println("Dashboard template (dashboardTmpl) is nil. Cannot render.")
		http.Error(w, "Dashboard template not initialized", http.StatusInternalServerError)
		return
	}
	// When using ExecuteTemplate with a specific template name from ParseFiles (like "layout.html"),
	// it executes that template. If "layout.html" then uses {{template "content" .}}
	// and "index.html" defines "content", it should work.
	err = dashboardTmpl.ExecuteTemplate(w, "layout.html", data)
	if err != nil {
		log.Printf("Error executing dashboard template: %v", err)
		http.Error(w, "Failed to render dashboard: "+err.Error(), http.StatusInternalServerError)
	}
}
