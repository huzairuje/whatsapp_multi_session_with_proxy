package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time" // For creating users with CreatedAt

	"whatsapp_multi_session/app/db"
	"whatsapp_multi_session/app/sessions"
	"golang.org/x/crypto/bcrypt" // For creating test users with hashed passwords
	// "log"
	// "os"
	// "github.com/gorilla/mux"
	// "gorm.io/driver/sqlite"
	// "gorm.io/gorm"
)

// testDB and testRouter are now global, defined in auth_handlers_test.go's TestMain

// clearTables_settings helper (uses global testDB)
func clearTables_settings() {
	testDB.Exec("DELETE FROM message_counts")
	testDB.Exec("DELETE FROM devices")
	testDB.Exec("DELETE FROM users")
}

// createAuthenticatedUserAndRequest_settings creates a request with a valid session for the given user ID.
// It also returns the created user for convenience. (uses global testDB)
func createAuthenticatedUserAndRequest_settings(t *testing.T, method, urlStr string, body strings.Reader) (*db.User, *http.Request) {
	t.Helper()
	// Create a dummy user
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("oldpassword"), bcrypt.DefaultCost)
	user := db.User{
		Username:     "settingsuser" + time.Now().Format("20060102150405.000000"), // Ensure unique username
		PasswordHash: string(hashedPassword),
		CreatedAt:    time.Now(),
	}
	if err := testDB.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create user for authenticated request: %v", err)
	}

	req, err := http.NewRequest(method, urlStr, &body)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder() // Used to save the session cookie
	session, _ := sessions.Store.Get(req, sessions.SessionKey)
	session.Values[sessions.UserIDKey] = user.ID
	err = sessions.Store.Save(req, rr, session)
	if err != nil {
		t.Fatalf("Failed to save session: %v", err)
	}

	// Apply the cookie from the recorder to the actual request
	for _, cookie := range rr.Result().Cookies() {
		req.AddCookie(cookie)
	}
	return &user, req
}

func TestSettingsPageGET_Authenticated(t *testing.T) {
	clearTables_settings()
	_, req := createAuthenticatedUserAndRequest_settings(t, http.MethodGet, "/settings", *strings.NewReader(""))

	rr := httptest.NewRecorder()
	testRouter.ServeHTTP(rr, req) // Use global testRouter

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	if !strings.Contains(rr.Body.String(), "Change Your Password") {
		t.Errorf("handler returned unexpected body, expected to contain 'Change Your Password'")
	}
}

func TestSettingsPageGET_Unauthenticated(t *testing.T) {
	clearTables_settings()
	req, _ := http.NewRequest(http.MethodGet, "/settings", nil)
	rr := httptest.NewRecorder()
	testRouter.ServeHTTP(rr, req) // Use global testRouter

	if status := rr.Code; status != http.StatusFound { // Expect redirect to /login
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

func TestHandleChangePassword_Success_NonHTMX(t *testing.T) {
	clearTables_settings()
	user, req := createAuthenticatedUserAndRequest_settings(t, http.MethodPost, "/settings/password",
		*strings.NewReader(url.Values{
			"current_password": {"oldpassword"},
			"new_password":     {"newpassword123"},
			"confirm_password": {"newpassword123"},
		}.Encode()),
	)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	testRouter.ServeHTTP(rr, req) // Use global testRouter

	if status := rr.Code; status != http.StatusOK { // Renders the page again with a message
		t.Errorf("handler returned wrong status code: got %v want %v. Body: %s", status, http.StatusOK, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "Password changed successfully.") {
		t.Errorf("expected success message not found in body")
	}

	// Verify password in DB
	var updatedUser db.User
	testDB.First(&updatedUser, user.ID) // Use global testDB
	err := bcrypt.CompareHashAndPassword([]byte(updatedUser.PasswordHash), []byte("newpassword123"))
	if err != nil {
		t.Errorf("new password was not correctly set in the database")
	}
}

func TestHandleChangePassword_Success_HTMX(t *testing.T) {
	clearTables_settings()
	user, req := createAuthenticatedUserAndRequest_settings(t, http.MethodPost, "/settings/password",
		*strings.NewReader(url.Values{
			"current_password": {"oldpassword"},
			"new_password":     {"newpassword123"},
			"confirm_password": {"newpassword123"},
		}.Encode()),
	)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("HX-Request", "true")

	rr := httptest.NewRecorder()
	testRouter.ServeHTTP(rr, req) // Use global testRouter

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v. Body: %s", status, http.StatusOK, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "Password changed successfully.") {
		t.Errorf("expected success message not found in partial response body: %s", rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "message-content") { // Check if it's the partial
		t.Errorf("expected partial response for HTMX request")
	}

	var updatedUser db.User
	testDB.First(&updatedUser, user.ID) // Use global testDB
	err := bcrypt.CompareHashAndPassword([]byte(updatedUser.PasswordHash), []byte("newpassword123"))
	if err != nil {
		t.Errorf("new password was not correctly set in the database (HTMX)")
	}
}
// Placeholder for the TestSettingsHandlersPlaceholder to avoid "imported and not used"
// This was TestDeviceHandlersPlaceholder in the previous step, needs to be unique or removed if other tests exist.
// The actual tests above make this redundant. I will remove it.
// func TestSettingsHandlersPlaceholder(t *testing.T) {}


func runChangePasswordSubTest(t *testing.T, isHtmx bool, formData url.Values, expectedStatus int, expectedMsg string, userPasswordBeforeChange string) {
	t.Helper()
	clearTables_settings() // Ensure clean state for each sub-test run

	// Create a user with a known current password
	hashedCurrentPassword, _ := bcrypt.GenerateFromPassword([]byte(userPasswordBeforeChange), bcrypt.DefaultCost)
	user := db.User{
		Username:     "testuser_cpsub_" + time.Now().Format("150405.000000"), // Unique username
		PasswordHash: string(hashedCurrentPassword),
		CreatedAt:    time.Now(),
	}
	if err := testDB.Create(&user).Error; err != nil { // Use global testDB
		t.Fatalf("SubTest: Failed to create user: %v", err)
	}

	// Create authenticated request
	req, err := http.NewRequest(http.MethodPost, "/settings/password", strings.NewReader(formData.Encode()))
	if err != nil {
		t.Fatalf("SubTest: Failed to create request: %v", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	if isHtmx {
		req.Header.Add("HX-Request", "true")
	}

	rr := httptest.NewRecorder() // Recorder for session saving
	session, _ := sessions.Store.Get(req, sessions.SessionKey)
	session.Values[sessions.UserIDKey] = user.ID
	if err := sessions.Store.Save(req, rr, session); err != nil {
		t.Fatalf("SubTest: Failed to save session: %v", err)
	}
	for _, cookie := range rr.Result().Cookies() { // Apply cookie to request
		req.AddCookie(cookie)
	}

	// Execute request
	rr = httptest.NewRecorder() // Fresh recorder for actual response
	testRouter.ServeHTTP(rr, req) // Use global testRouter

	// Assertions
	if status := rr.Code; status != expectedStatus {
		t.Errorf("SubTest (HTMX: %v): Handler returned wrong status code: got %v want %v. Body: %s", isHtmx, status, expectedStatus, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), expectedMsg) {
		t.Errorf("SubTest (HTMX: %v): Expected message '%s' not found in body: %s", isHtmx, expectedMsg, rr.Body.String())
	}
	if isHtmx && !strings.Contains(rr.Body.String(), "message-content") {
		t.Errorf("SubTest (HTMX: %v): Expected partial response for HTMX request", isHtmx)
	}
}

func TestHandleChangePassword_ValidationAndErrors(t *testing.T) {
	// This single test function will cover multiple validation/error scenarios using subtests.
	// The `userPasswordBeforeChange` is the actual password for user created in `runChangePasswordSubTest`
	const userPasswordBeforeChange = "currentGoodPassword"

	testCases := []struct {
		name           string
		isHtmx         bool
		formData       url.Values
		expectedStatus int    // For full page, or for HTMX if server sets specific error codes
		expectedMsg    string
	}{
		// Incorrect Current Password
		{
			"IncorrectCurrentPassword_NonHTMX", false,
			url.Values{"current_password": {"wrongOldPassword"}, "new_password": {"newPass123"}, "confirm_password": {"newPass123"}},
			http.StatusOK, // ShowSettingsPage renders 200 with error message
			"Current password is incorrect.",
		},
		{
			"IncorrectCurrentPassword_HTMX", true,
			url.Values{"current_password": {"wrongOldPassword"}, "new_password": {"newPass123"}, "confirm_password": {"newPass123"}},
			http.StatusOK, // Partial is rendered with 200
			"Current password is incorrect.",
		},
		// New Passwords Do Not Match
		{
			"NewPasswordsDoNotMatch_NonHTMX", false,
			url.Values{"current_password": {userPasswordBeforeChange}, "new_password": {"newPass123"}, "confirm_password": {"NOMATCHY"}},
			http.StatusOK,
			"New password and confirmation password do not match.",
		},
		{
			"NewPasswordsDoNotMatch_HTMX", true,
			url.Values{"current_password": {userPasswordBeforeChange}, "new_password": {"newPass123"}, "confirm_password": {"NOMATCHY"}},
			http.StatusOK,
			"New password and confirmation password do not match.",
		},
		// New Password Too Short
		{
			"NewPasswordTooShort_NonHTMX", false,
			url.Values{"current_password": {userPasswordBeforeChange}, "new_password": {"short"}, "confirm_password": {"short"}},
			http.StatusOK,
			"New password must be at least 6 characters long.",
		},
		{
			"NewPasswordTooShort_HTMX", true,
			url.Values{"current_password": {userPasswordBeforeChange}, "new_password": {"short"}, "confirm_password": {"short"}},
			http.StatusOK,
			"New password must be at least 6 characters long.",
		},
		// Empty fields (example: current password empty)
		{
			"EmptyCurrentPassword_NonHTMX", false,
			url.Values{"current_password": {""}, "new_password": {"newPass123"}, "confirm_password": {"newPass123"}},
			http.StatusOK,
			"All password fields are required.",
		},
		{
			"EmptyCurrentPassword_HTMX", true,
			url.Values{"current_password": {""}, "new_password": {"newPass123"}, "confirm_password": {"newPass123"}},
			http.StatusOK,
			"All password fields are required.",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			runChangePasswordSubTest(t, tc.isHtmx, tc.formData, tc.expectedStatus, tc.expectedMsg, userPasswordBeforeChange)
		})
	}
}


// (Removed the wrapper TestMain that called TestMain_Settings)
