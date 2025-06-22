package handlers

import (
	"whatsapp_multi_session/app/db"
	"whatsapp_multi_session/app/sessions"
	"html/template"
	"log"
	"net/http"
	"time" // For date formatting in template

	"github.com/gin-gonic/gin" // Added Gin import
	"gorm.io/gorm" // For gorm.ErrRecordNotFound
)

var messagesTmpl *template.Template

// --- Gin Handlers (Placeholders) ---
func ShowMessagesPageGin(c *gin.Context) {
	log.Println("GIN HANDLER: ShowMessagesPageGin called")
	c.String(http.StatusOK, "Placeholder for Gin ShowMessagesPage")
}

// MessagePageData holds data for the messages page
type MessagePageData struct {
	LoggedIn     bool
	UserID       uint
	Username     string
	PageTitle    string
	MessageCount *db.MessageCount // Pointer to handle case where it might not exist
}

// InitMessageTemplates pre-parses the message-related templates.
func InitMessageTemplates() {
	tpl, err := template.New("messages.html").Funcs(template.FuncMap{
		"formatDate": func(t time.Time) string {
			// Check if time is zero, return empty string or placeholder
			if t.IsZero() {
				return "N/A"
			}
			return t.Format("Jan 02, 2006 03:04 PM")
		},
	}).ParseFiles("app/templates/layout.html", "app/templates/messages.html") // Reverted paths
	if err != nil {
		log.Fatalf("Error parsing message templates: %v", err)
	}
	messagesTmpl = tpl
}

// ShowMessagesPage displays message summary for the logged-in user.
func ShowMessagesPage(w http.ResponseWriter, r *http.Request) {
	session, err := sessions.Store.Get(r, sessions.SessionKey)
	if err != nil {
		log.Printf("Error getting session for ShowMessagesPage: %v. Redirecting to login.", err)
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	userIDVal, ok := session.Values[sessions.UserIDKey]
	if !ok {
		log.Println("UserID not found in session for ShowMessagesPage. Redirecting to login.")
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
		log.Printf("Error fetching user %d for messages page: %v. Redirecting to login.", userID, err)
		http.Redirect(w, r, "/login", http.StatusFound) // Or an error page
		return
	}

	var msgCount db.MessageCount
	dbErr := db.DB.Where("user_id = ?", userID).First(&msgCount).Error

	pageData := MessagePageData{
		LoggedIn:  true,
		UserID:    userID,
		Username:  user.Username,
		PageTitle: "My Messages",
	}

	if dbErr != nil {
		if dbErr == gorm.ErrRecordNotFound {
			// No message count record found, this is not an error for display purposes.
			// pageData.MessageCount will remain nil, template handles this.
			log.Printf("No message count record found for user %d. Message count will be displayed as unavailable or 0.", userID)
		} else {
			// Actual database error
			log.Printf("Error fetching message count for user %d: %v. Message count will be displayed as unavailable.", userID, dbErr)
			// Optionally, set an error message in pageData to display to the user on the page.
		}
	} else {
		pageData.MessageCount = &msgCount
	}

	if messagesTmpl == nil {
		log.Println("Messages template (messagesTmpl) is nil. Cannot render.")
		http.Error(w, "Messages template not initialized", http.StatusInternalServerError)
		return
	}

	templateErr := messagesTmpl.ExecuteTemplate(w, "layout.html", pageData)
	if templateErr != nil {
		log.Printf("Error executing messages template: %v", templateErr)
		http.Error(w, "Failed to render messages page: "+templateErr.Error(), http.StatusInternalServerError)
	}
}
