package handlers

import (
	"whatsapp_multi_session/app/db"
	"whatsapp_multi_session/app/sessions"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin" // Added Gin import
	"golang.org/x/crypto/bcrypt"
)

var settingsTmpl *template.Template
var messagePartialTmpl *template.Template // For HTMX responses

// --- Gin Handlers (Placeholders) ---
func SettingsPageGETGin(c *gin.Context) {
	log.Println("GIN HANDLER: SettingsPageGETGin called")
	c.String(http.StatusOK, "Placeholder for Gin SettingsPageGET")
}

func HandleChangePasswordGin(c *gin.Context) {
	log.Println("GIN HANDLER: HandleChangePasswordGin called")
	c.String(http.StatusOK, "Placeholder for Gin HandleChangePassword")
}

// SettingsPageData remains the same
type SettingsPageData struct {
	LoggedIn  bool
	UserID    uint
	Username  string
	PageTitle string
	Error     bool
	Message   string
}

// InitSettingsTemplates pre-parses the settings-related templates.
func InitSettingsTemplates() {
	// Main template for full page loads
	// It now also knows about _messages.html if we want to define it as a sub-template within settings.html,
	// but for standalone partial rendering, messagePartialTmpl is better.
	mainTpl, err := template.New("settings.html").Funcs(template.FuncMap{
		"formatDate": func(t time.Time) string {
			if t.IsZero() {
				return "N/A"
			}
			return t.Format("Jan 02, 2006") // Or any other consistent format
		},
	}).ParseFiles("app/templates/layout.html", "app/templates/settings.html", "app/templates/_messages.html") // Reverted paths
	if err != nil {
		log.Fatalf("Error parsing main settings templates (including _messages.html): %v", err)
	}
	settingsTmpl = mainTpl

	// Standalone partial template for HTMX message updates
	// The name of the template defined by ParseFiles is the base name of the file.
	partialTpl, err := template.ParseFiles("app/templates/_messages.html") // Reverted path
	if err != nil {
		log.Fatalf("Error parsing message partial template: %v", err)
	}
	messagePartialTmpl = partialTpl
}

// renderSettingsResponse handles rendering for both HTMX and regular requests
func renderSettingsResponse(w http.ResponseWriter, r *http.Request, data SettingsPageData) {
	isHtmx := r.Header.Get("HX-Request") == "true"

	if isHtmx {
		// For HTMX, just render the message partial
		if messagePartialTmpl == nil {
			log.Println("ERROR: messagePartialTmpl is nil")
			http.Error(w, "Message template not initialized", http.StatusInternalServerError)
			return
		}
		// If there's an error, HTMX can use this status code.
		// However, often it's simpler to return 200 and let HTMX swap the content,
		// which includes the error styling.
		// if data.Error {
		// w.WriteHeader(http.StatusUnprocessableEntity) // Example for validation errors
		// }
		err := messagePartialTmpl.ExecuteTemplate(w, "_messages.html", data)
		if err != nil {
			log.Printf("ERROR rendering message partial: %v", err)
			http.Error(w, "Failed to render message partial: "+err.Error(), http.StatusInternalServerError)
		}
	} else {
		// For regular requests, render the full page
		if settingsTmpl == nil {
			log.Println("ERROR: settingsTmpl is nil")
			http.Error(w, "Settings template not initialized", http.StatusInternalServerError)
			return
		}
		// For non-HTMX, status codes for errors are typically handled by redirects or specific error pages.
		// Here, we're re-rendering the form with a message, so 200 is often fine,
		// but specific error codes like 400/422 could be used if the client were set up to handle them.
		// Example: if data.Error { w.WriteHeader(http.StatusBadRequest) }
		err := settingsTmpl.ExecuteTemplate(w, "layout.html", data)
		if err != nil {
			log.Printf("ERROR rendering full settings page: %v", err)
			http.Error(w, "Failed to render settings page: "+err.Error(), http.StatusInternalServerError)
		}
	}
}

// ShowSettingsPage displays the change password form.
func ShowSettingsPage(w http.ResponseWriter, r *http.Request, message string, isError bool) {
	session, err := sessions.Store.Get(r, sessions.SessionKey)
	if err != nil {
		log.Printf("Error getting session for ShowSettingsPage: %v. Redirecting to login.", err)
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	userIDVal, ok := session.Values[sessions.UserIDKey]
	if !ok {
		log.Println("UserID not found in session for ShowSettingsPage. Redirecting to login.")
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	userID, ok := userIDVal.(uint)
	if !ok || userID == 0 {
		log.Printf("UserID in session is invalid (not uint or zero): %v. Redirecting to login.", userIDVal)
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	var user db.User
	if err := db.DB.First(&user, userID).Error; err != nil {
		log.Printf("Error fetching user %d for settings page: %v", userID, err)
		http.Error(w, "Could not retrieve user data.", http.StatusInternalServerError)
		return
	}

	data := SettingsPageData{
		LoggedIn:  true, UserID: userID, Username: user.Username, // Added UserID and Username
		PageTitle: "Settings - Change Password", Message: message, Error: isError,
	}
	renderSettingsResponse(w, r, data) // Use the unified render function
}

func SettingsPageGET(w http.ResponseWriter, r *http.Request) {
	ShowSettingsPage(w, r, "", false) // No message on initial GET
}

// HandleChangePassword processes the password change request.
func HandleChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	session, errSess := sessions.Store.Get(r, sessions.SessionKey)
	if errSess != nil {
		log.Printf("Error getting session for HandleChangePassword: %v. Redirecting to login.", errSess)
		http.Redirect(w, r, "/login", http.StatusFound) // Should not happen if page access is controlled
		return
	}
	userIDVal, okSess := session.Values[sessions.UserIDKey]
	if !okSess {
		log.Println("UserID not found in session for HandleChangePassword. Redirecting to login.")
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	userID, okCast := userIDVal.(uint)
	if !okCast || userID == 0 {
		log.Printf("UserID in session is invalid for HandleChangePassword: %v. Redirecting to login.", userIDVal)
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	currentPassword := r.FormValue("current_password")
	newPassword := r.FormValue("new_password")
	confirmPassword := r.FormValue("confirm_password")

	// Initialize data for response. User details are needed for full page re-render.
	var user db.User
	var username string
	if errDb := db.DB.First(&user, userID).Error; errDb != nil {
		log.Printf("Error fetching user %d for password change response: %v", userID, errDb)
		// This is tricky. If we can't get user, full page re-render might be an issue.
		// For HTMX, username is not critical. For full page, it is.
		// For now, proceed, but this could lead to partial data on full page error.
		username = "User" // Fallback username
	} else {
		username = user.Username
	}

	data := SettingsPageData{
		LoggedIn:  true, UserID: userID, Username: username,
		PageTitle: "Settings - Change Password", // Needed for full page reload
	}


	if currentPassword == "" || newPassword == "" || confirmPassword == "" {
		data.Message = "All password fields are required."
		data.Error = true
		renderSettingsResponse(w, r, data)
		return
	}
	if newPassword != confirmPassword {
		data.Message = "New password and confirmation password do not match."
		data.Error = true
		renderSettingsResponse(w, r, data)
		return
	}
	if len(newPassword) < 6 {
		data.Message = "New password must be at least 6 characters long."
		data.Error = true
		renderSettingsResponse(w, r, data)
		return
	}

	// Re-fetch user if not already fetched, or if the earlier fetch was just for username
	// Here, 'user' variable from username fetch above is already populated.

	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(currentPassword))
	if err != nil {
		data.Message = "Current password is incorrect."
		data.Error = true
		renderSettingsResponse(w, r, data)
		return
	}

	hashedNewPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing new password for user %d: %v", userID, err)
		data.Message = "Error processing new password. Please try again."
		data.Error = true
		renderSettingsResponse(w, r, data)
		return
	}

	user.PasswordHash = string(hashedNewPassword)
	if err := db.DB.Save(&user).Error; err != nil {
		log.Printf("Error saving new password for user %d: %v", userID, err)
		data.Message = "Failed to update password. Please try again."
		data.Error = true
		renderSettingsResponse(w, r, data)
		return
	}

	data.Message = "Password changed successfully."
	data.Error = false
	renderSettingsResponse(w, r, data)
}
