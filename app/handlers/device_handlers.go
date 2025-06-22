package handlers

import (
	"whatsapp_multi_session/app/db"
	"whatsapp_multi_session/app/sessions" // Assuming sessions package is correctly set up
	"html/template"
	"log"
	"net/http"
	"time" // For CreatedAt formatting

	"github.com/gin-gonic/gin" // Added Gin import
)

var devicesTmpl *template.Template

// --- Gin Handlers (Placeholders) ---
func ShowDevicesPageGin(c *gin.Context) {
	log.Println("GIN HANDLER: ShowDevicesPageGin called")
	c.String(http.StatusOK, "Placeholder for Gin ShowDevicesPage")
}

// DevicePageData holds data for the devices page
type DevicePageData struct {
	LoggedIn  bool
	UserID    uint
	Username  string
	PageTitle string
	Devices   []db.Device
}

// InitDeviceTemplates pre-parses the device-related templates.
func InitDeviceTemplates() {
	// It needs to parse layout.html as well to define content within it.
	// The template name is derived from the base name of the first file in ParseFiles.
	// To execute "layout.html" as the entry point, ParseFiles should have "layout.html" first,
	// or the template name should be explicitly used.
	// Here, we name the template "devices.html" explicitly by New().
	// Then ParseFiles adds definitions from layout.html and devices.html into this named template.
	tpl, err := template.New("devices.html").Funcs(template.FuncMap{
		"formatDate": func(t time.Time) string {
			if t.IsZero() {
				return "N/A" // Handle zero time gracefully
			}
			return t.Format("Jan 02, 2006 03:04 PM")
		},
	}).ParseFiles("app/templates/layout.html", "app/templates/devices.html") // Reverted paths
	if err != nil {
		log.Fatalf("Error parsing device templates: %v", err)
	}
	devicesTmpl = tpl
}

// ShowDevicesPage displays the list of devices for the logged-in user.
func ShowDevicesPage(w http.ResponseWriter, r *http.Request) {
	session, err := sessions.Store.Get(r, sessions.SessionKey)
	if err != nil {
		log.Printf("Error getting session for ShowDevicesPage: %v. Redirecting to login.", err)
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	userIDVal, ok := session.Values[sessions.UserIDKey]
	if !ok {
		log.Println("UserID not found in session for ShowDevicesPage. Redirecting to login.")
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
		log.Printf("Error fetching user %d for devices page: %v. Redirecting to login.", userID, err)
		http.Redirect(w, r, "/login", http.StatusFound) // Or an error page
		return
	}

	var devices []db.Device
	if err := db.DB.Where("user_id = ?", userID).Order("created_at desc").Find(&devices).Error; err != nil {
		log.Printf("Error fetching devices for user %d: %v. Page will be rendered with an empty list or error message.", userID, err)
		// Depending on requirements, you might want to show an error on the page
		// For now, it will just show an empty list if there's a DB error.
	}

	data := DevicePageData{
		LoggedIn:  true,
		UserID:    userID,
		Username:  user.Username,
		PageTitle: "My Devices",
		Devices:   devices,
	}

	if devicesTmpl == nil {
		log.Println("Devices template (devicesTmpl) is nil. Cannot render.")
		http.Error(w, "Devices template not initialized", http.StatusInternalServerError)
		return
	}

	// Execute the layout, which will in turn render the "devices.html" content block
	err = devicesTmpl.ExecuteTemplate(w, "layout.html", data)
	if err != nil {
		log.Printf("Error executing devices template: %v", err)
		http.Error(w, "Failed to render devices page: "+err.Error(), http.StatusInternalServerError)
	}
}
