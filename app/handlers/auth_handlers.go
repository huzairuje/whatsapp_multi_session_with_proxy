package handlers

import (
	"whatsapp_multi_session/app/db"
	"whatsapp_multi_session/app/sessions"
	"html/template"
	"log"
	"net/http"
	"regexp"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	registerTmpl           *template.Template
	loginTmpl              *template.Template
	authMessagePartialTmpl *template.Template
)

// AuthMessageData is used for partial responses and full-page data
type AuthMessageData struct {
	Error    bool
	Message  string
	Username string // To preserve username input on error for full page reload
	// Add any other fields needed by login.html or register.html for initial display
}

func InitAuthTemplates() {
	// Parse main templates (register.html, login.html)
	// These are standalone pages and don't use layout.html
	// They have their own internal message display logic for non-HTMX GET requests.
	// Adjusted paths for tests run from app/handlers/ directory.
	var err error
	registerTmpl, err = template.ParseFiles("../templates/register.html")
	if err != nil {
		log.Fatalf("Error parsing register.html: %v", err)
	}

	loginTmpl, err = template.ParseFiles("../templates/login.html")
	if err != nil {
		log.Fatalf("Error parsing login.html: %v", err)
	}

	// Parse the message partial separately for HTMX responses
	// This partial is defined in app/templates/_messages.html
	authMessagePartialTmpl, err = template.ParseFiles("../templates/_messages.html")
	if err != nil {
		log.Fatalf("Error parsing auth message partial template (_messages.html): %v", err)
	}
}

// renderAuthResponse handles rendering for both HTMX and regular auth requests
func renderAuthResponse(w http.ResponseWriter, r *http.Request, pageTmpl *template.Template, data AuthMessageData) {
	isHtmx := r.Header.Get("HX-Request") == "true"

	if isHtmx {
		if authMessagePartialTmpl == nil {
			log.Println("ERROR: authMessagePartialTmpl is nil")
			http.Error(w, "Auth Message partial template not initialized", http.StatusInternalServerError)
			return
		}
		// Set appropriate status code for HTMX error responses if desired, e.g. based on data.Error
		// For example: if data.Error { w.WriteHeader(http.StatusUnprocessableEntity) }
		err := authMessagePartialTmpl.ExecuteTemplate(w, "_messages.html", data)
		if err != nil {
			log.Printf("ERROR rendering auth message partial: %v", err)
			http.Error(w, "Failed to render auth message partial: "+err.Error(), http.StatusInternalServerError)
		}
	} else {
		// For regular requests, render the full page template
		if pageTmpl == nil {
			log.Println("ERROR: Full page template (e.g., registerTmpl) is nil")
			http.Error(w, "Page template not initialized", http.StatusInternalServerError)
			return
		}
		// Status code for full page (e.g. http.StatusBadRequest) should be set by caller before this function
		err := pageTmpl.Execute(w, data) // Execute the main page template (e.g., register.html)
		if err != nil {
			log.Printf("ERROR rendering full auth page: %v", err)
			http.Error(w, "Failed to render page: "+err.Error(), http.StatusInternalServerError)
		}
	}
}

// ShowRegistrationPage serves the registration form
func ShowRegistrationPage(w http.ResponseWriter, r *http.Request) {
	session, _ := sessions.Store.Get(r, sessions.SessionKey)
	if _, ok := session.Values[sessions.UserIDKey]; ok { // User is already logged in
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	if registerTmpl == nil {
		log.Println("ERROR: registerTmpl is nil in ShowRegistrationPage")
		http.Error(w, "Registration template not initialized", http.StatusInternalServerError)
		return
	}
	// For a GET request, there's no message or error yet.
	// Pass empty AuthMessageData or nil. If AuthMessageData is used, ensure fields are appropriately zero/empty.
	err := registerTmpl.Execute(w, AuthMessageData{})
	if err != nil {
		log.Printf("Error rendering registration page: %v", err)
		http.Error(w, "Failed to render registration page: "+err.Error(), http.StatusInternalServerError)
	}
}

// HandleRegistration processes the registration form submission
func HandleRegistration(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	// Preserve username for full page reload on error, and for message partial if needed
	responseData := AuthMessageData{Username: username}

	if username == "" || password == "" {
		responseData.Error = true
		responseData.Message = "Username and password are required."
		w.WriteHeader(http.StatusBadRequest) // Set status for non-HTMX full page reload
		renderAuthResponse(w, r, registerTmpl, responseData)
		return
	}

	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_]{3,20}$`)
	if !usernameRegex.MatchString(username) {
		responseData.Error = true
		responseData.Message = "Username must be 3-20 characters long and contain only alphanumeric characters and underscores."
		w.WriteHeader(http.StatusBadRequest)
		renderAuthResponse(w, r, registerTmpl, responseData)
		return
	}

	if len(password) < 6 {
		responseData.Error = true
		responseData.Message = "Password must be at least 6 characters long."
		w.WriteHeader(http.StatusBadRequest)
		renderAuthResponse(w, r, registerTmpl, responseData)
		return
	}

	var existingUser db.User
	if err := db.DB.Where("username = ?", username).First(&existingUser).Error; err == nil {
		responseData.Error = true
		responseData.Message = "Username already taken. Please choose another."
		w.WriteHeader(http.StatusConflict) // Conflict for existing username
		renderAuthResponse(w, r, registerTmpl, responseData)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		responseData.Error = true
		responseData.Message = "An internal error occurred while securing your password. Please try again."
		w.WriteHeader(http.StatusInternalServerError)
		renderAuthResponse(w, r, registerTmpl, responseData)
		return
	}

	newUser := db.User{Username: username, PasswordHash: string(hashedPassword), CreatedAt: time.Now()}
	result := db.DB.Create(&newUser)
	if result.Error != nil {
		log.Printf("Error creating user: %v", result.Error)
		responseData.Error = true
		responseData.Message = "Failed to create account due to a database error. Please try again."
		w.WriteHeader(http.StatusInternalServerError)
		renderAuthResponse(w, r, registerTmpl, responseData)
		return
	}

	// Create initial message count for the user
	newMessageCount := db.MessageCount{UserID: newUser.ID, Count: 0, LastUpdatedAt: time.Now()}
	if err := db.DB.Create(&newMessageCount).Error; err != nil {
		log.Printf("Error creating initial message count for user %d: %v. This is non-critical for registration.", newUser.ID, err)
		// Do not send this error to user as registration itself was successful.
	}

	responseData.Error = false
	responseData.Message = "Registration successful! You can now login."
	// For HTMX, the form remains, and message updates. User then clicks login link.
	// For non-HTMX, original code showed message on register page. This is fine.
	// If successful and non-HTMX, a redirect to login might be better:
	// if r.Header.Get("HX-Request") != "true" { http.Redirect(w, r, "/login?registration_success=true", http.StatusSeeOther); return }
	renderAuthResponse(w, r, registerTmpl, responseData)
}

// ShowLoginPage serves the login form
func ShowLoginPage(w http.ResponseWriter, r *http.Request) {
    session, _ := sessions.Store.Get(r, sessions.SessionKey)
	if _, ok := session.Values[sessions.UserIDKey]; ok { // User is already logged in
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	if loginTmpl == nil {
        log.Println("ERROR: loginTmpl is nil in ShowLoginPage")
		http.Error(w, "Login template not initialized", http.StatusInternalServerError)
		return
	}
	// Check for registration_success message from redirect
	// This is for non-HTMX flow. HTMX shows message on same page.
	// var message string
	// if r.URL.Query().Get("registration_success") == "true" {
	// 	message = "Registration successful! You can now login."
	// }
	// err := loginTmpl.Execute(w, AuthMessageData{Message: message})

    // For now, keeping it simple, no message on initial GET for login page.
	err := loginTmpl.Execute(w, AuthMessageData{})
	if err != nil {
        log.Printf("Error rendering login page: %v", err)
		http.Error(w, "Failed to render login page: "+err.Error(), http.StatusInternalServerError)
	}
}

// HandleLogin processes the login form submission
func HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	responseData := AuthMessageData{Username: username}

	if username == "" || password == "" {
		responseData.Error = true
		responseData.Message = "Username and password are required."
		w.WriteHeader(http.StatusBadRequest)
		renderAuthResponse(w, r, loginTmpl, responseData) // Will use loginTmpl for full page
		return
	}

	var user db.User
	if err := db.DB.Where("username = ?", username).First(&user).Error; err != nil {
		responseData.Error = true
		responseData.Message = "Invalid username or password."
		w.WriteHeader(http.StatusUnauthorized)
		renderAuthResponse(w, r, loginTmpl, responseData)
		return
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		responseData.Error = true
		responseData.Message = "Invalid username or password."
		w.WriteHeader(http.StatusUnauthorized)
		renderAuthResponse(w, r, loginTmpl, responseData)
		return
	}

	session, _ := sessions.Store.Get(r, sessions.SessionKey)
	session.Values[sessions.UserIDKey] = user.ID
	err = session.Save(r, w)
	if err != nil {
		log.Printf("Error saving session: %v", err)
		responseData.Error = true
		responseData.Message = "Login failed due to a server error. Please try again."
		w.WriteHeader(http.StatusInternalServerError)
		renderAuthResponse(w, r, loginTmpl, responseData)
		return
	}

    // For HTMX, if login is successful, the partial response will be empty or a success message.
    // The page won't redirect automatically. Client-side JS (or hx-redirect) would be needed.
    // For non-HTMX, redirect is standard.
    if r.Header.Get("HX-Request") == "true" {
        // Option 1: Send a special header to tell HTMX to redirect
        w.Header().Set("HX-Redirect", "/")
        // Option 2: Send a script to redirect, or just a success message
        // For now, sending HX-Redirect is cleanest. If not supported/desired, a message is fine.
        // responseData.Message = "Login successful! Redirecting..."
        // renderAuthResponse(w, r, loginTmpl, responseData)
    } else {
	    http.Redirect(w, r, "/", http.StatusFound)
    }
}

// LogoutHandler clears the session
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := sessions.Store.Get(r, sessions.SessionKey)
	delete(session.Values, sessions.UserIDKey)
	session.Options.MaxAge = -1
	err := session.Save(r, w)
	if err != nil {
		log.Printf("Error saving session during logout: %v", err)
		http.Error(w, "Failed to logout. Please try again.", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/login", http.StatusFound)
}
