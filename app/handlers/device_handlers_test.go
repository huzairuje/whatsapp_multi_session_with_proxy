package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time" // For CreatedAt values

	"whatsapp_multi_session/app/db"
	"whatsapp_multi_session/app/sessions"
	// "log" // No longer directly used in tests
	// "os" // No longer directly used in tests
	// "github.com/gorilla/mux" // No longer directly used as testRouter is global
	// "gorm.io/driver/sqlite" // No longer directly used as testDB is global
	// "gorm.io/gorm" // No longer directly used as testDB is global
)

// testDB and testRouter are global, defined in auth_handlers_test.go's TestMain

func clearTables_device() {
	testDB.Exec("DELETE FROM message_counts") // Use global testDB
	testDB.Exec("DELETE FROM devices")
	testDB.Exec("DELETE FROM users")
}

// (User is created within tests for device page, this helper just sets up session)
func createAuthenticatedUserAndRequest_device(t *testing.T, user db.User, method, urlStr string) *http.Request {
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

func TestShowDevicesPage_Authenticated_WithDevices(t *testing.T) {
	clearTables_device()
	user := db.User{Username: "deviceUser1", PasswordHash: "hash", CreatedAt: time.Now()}
	testDB.Create(&user) // Use global testDB

	device1Time := time.Now().Add(-time.Hour)
	device2Time := time.Now().Add(-2 * time.Hour) // Ensure different times for potential order checks

	testDB.Create(&db.Device{UserID: user.ID, DeviceName: "Phone XL", CreatedAt: device1Time}) // Use global testDB
	testDB.Create(&db.Device{UserID: user.ID, DeviceName: "Laptop Pro", CreatedAt: device2Time}) // Use global testDB

	req := createAuthenticatedUserAndRequest_device(t, user, http.MethodGet, "/devices")
	rr := httptest.NewRecorder()
	testRouter.ServeHTTP(rr, req) // Use global testRouter

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v. Body: %s", status, http.StatusOK, rr.Body.String())
	}
	body := rr.Body.String()
	if !strings.Contains(body, "My Devices") {
		t.Errorf("expected 'My Devices' title not found")
	}
	if !strings.Contains(body, "Phone XL") {
		t.Errorf("expected device 'Phone XL' not found")
	}
	if !strings.Contains(body, "Laptop Pro") {
		t.Errorf("expected device 'Laptop Pro' not found")
	}
	// Check for formatted dates (using the formatDate func in template)
	// Example: device1Time.Format("Jan 02, 2006 03:04 PM")
	if !strings.Contains(body, device1Time.Format("Jan 02, 2006 03:04 PM")) {
		t.Errorf("expected formatted date for Phone XL not found: %s", device1Time.Format("Jan 02, 2006 03:04 PM"))
	}
}

func TestShowDevicesPage_Authenticated_NoDevices(t *testing.T) {
	clearTables_device()
	user := db.User{Username: "deviceUser2", PasswordHash: "hash", CreatedAt: time.Now()}
	testDB.Create(&user) // Use global testDB

	req := createAuthenticatedUserAndRequest_device(t, user, http.MethodGet, "/devices")
	rr := httptest.NewRecorder()
	testRouter.ServeHTTP(rr, req) // Use global testRouter

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	if !strings.Contains(rr.Body.String(), "You have not added any devices yet.") {
		t.Errorf("expected 'no devices' message not found")
	}
}

func TestShowDevicesPage_Unauthenticated(t *testing.T) {
	clearTables_device()
	req, _ := http.NewRequest(http.MethodGet, "/devices", nil)
	rr := httptest.NewRecorder()
	testRouter.ServeHTTP(rr, req) // Use global testRouter

	if status := rr.Code; status != http.StatusFound {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusFound)
	}
	redirectURL, _ := rr.Result().Location()
	if redirectURL.Path != "/login" {
		t.Errorf("expected redirect to /login, got %s", redirectURL.Path)
	}
}

// Removed the wrapper TestMain
