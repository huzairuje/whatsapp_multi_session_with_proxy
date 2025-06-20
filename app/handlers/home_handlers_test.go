package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time" // For user CreatedAt

	"whatsapp_multi_session/app/db"
	"whatsapp_multi_session/app/sessions"
	// "log"
	// "os"
	// "github.com/gorilla/mux"
	// "gorm.io/driver/sqlite"
	// "gorm.io/gorm"
)

// testDB and testRouter are global, defined in auth_handlers_test.go's TestMain

func clearTables_home() {
	testDB.Exec("DELETE FROM message_counts") // Use global testDB
	testDB.Exec("DELETE FROM devices")
	testDB.Exec("DELETE FROM users")
}

// createAuthenticatedUserAndRequest_home creates a request with a valid session for the given user.
// (Uses global testDB if user creation was part of it, but here user is passed in)
func createAuthenticatedUserAndRequest_home(t *testing.T, user db.User, method, urlStr string) *http.Request {
	t.Helper()

	req, err := http.NewRequest(method, urlStr, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	session, _ := sessions.Store.Get(req, sessions.SessionKey)
	session.Values[sessions.UserIDKey] = user.ID
	err = sessions.Store.Save(req, rr, session)
	if err != nil {
		t.Fatalf("Failed to save session: %v", err)
	}
	for _, cookie := range rr.Result().Cookies() {
		req.AddCookie(cookie)
	}
	return req
}

func TestHomeHandler_Authenticated(t *testing.T) {
	clearTables_home()

	// 1. Setup: Create a user
	user := db.User{Username: "homeUser", PasswordHash: "hash", CreatedAt: time.Now()}
	if err := testDB.Create(&user).Error; err != nil { // Use global testDB
		t.Fatalf("Failed to create user: %v", err)
	}

	// 2. Setup: Add devices for the user
	testDB.Create(&db.Device{UserID: user.ID, DeviceName: "Device1", CreatedAt: time.Now()}) // Use global testDB
	testDB.Create(&db.Device{UserID: user.ID, DeviceName: "Device2", CreatedAt: time.Now()}) // Use global testDB

	// 3. Setup: Add message count for the user
	testDB.Create(&db.MessageCount{UserID: user.ID, Count: 123, LastUpdatedAt: time.Now()}) // Use global testDB

	// 4. Action: Access / with the user's session
	req := createAuthenticatedUserAndRequest_home(t, user, http.MethodGet, "/")
	rr := httptest.NewRecorder()
	testRouter.ServeHTTP(rr, req) // Use global testRouter

	// 5. Assertions
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v. Body: %s", status, http.StatusOK, rr.Body.String())
	}
	body := rr.Body.String()
	if !strings.Contains(body, "Main Dashboard") {
		t.Errorf("expected 'Main Dashboard' title not found")
	}
	if !strings.Contains(body, "Welcome, homeUser") {
		t.Errorf("expected welcome message for 'homeUser' not found")
	}
	// Check for device count (stringified "2")
	if !strings.Contains(body, "You currently have <strong class=\"text-blue-600\">2</strong> registered device(s).") {
		t.Errorf("expected device count of 2 not found or not matching format. Body: %s", body)
	}
	// Check for message count (stringified "123")
	if !strings.Contains(body, "Total messages recorded: <strong class=\"text-green-600\">123</strong>.") {
		t.Errorf("expected message count of 123 not found or not matching format. Body: %s", body)
	}
}

func TestHomeHandler_Unauthenticated(t *testing.T) {
	clearTables_home()
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	testRouter.ServeHTTP(rr, req) // Use global testRouter

	if status := rr.Code; status != http.StatusFound {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusFound)
	}
	redirectURL, err := rr.Result().Location()
	if err != nil {
		t.Fatalf("Could not get redirect location: %v", err)
	}
	if redirectURL.Path != "/login" {
		t.Errorf("expected redirect to /login, got %s", redirectURL.Path)
	}
}

// Removed the wrapper TestMain
