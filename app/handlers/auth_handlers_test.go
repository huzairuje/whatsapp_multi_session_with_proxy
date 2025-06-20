// app/handlers/auth_handlers_test.go
package handlers

import (
	"whatsapp_multi_session/app/db"
	"whatsapp_multi_session/app/sessions"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
    "log"
    "time" // Added for time.Now()
    "golang.org/x/crypto/bcrypt" // Added for bcrypt

	"github.com/gorilla/mux"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var testDB *gorm.DB
var testRouter *mux.Router // Keep a router instance for tests

// TestMain sets up the test environment
func TestMain(m *testing.M) {
	// Setup: Initialize a temporary in-memory SQLite database for tests
	var err error
	// Using "file::memory:?cache=shared" allows multiple connections to the same in-memory DB
	// which can be useful, but ensure proper isolation if tests run in parallel.
	// For simple sequential tests, "file:test.db?mode=memory&cache=shared" or just in-memory without file.
	testDB, err = gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to in-memory database for testing: %v", err)
	}
	db.DB = testDB // Override the global DB variable in the db package

	// Run migrations
	// Make sure all necessary models are included
	err = db.DB.AutoMigrate(&db.User{}, &db.Device{}, &db.MessageCount{})
	if err != nil {
		log.Fatalf("Failed to auto-migrate test database: %v", err)
	}

	// Initialize sessions and templates
	// It's crucial that InitSessionStore uses fixed keys for testing or allows key injection.
    // The current InitSessionStore generates random keys if ENV is not set, which is bad for tests.
    // For now, we'll rely on the fallback fixed key if no ENV is set.
    // A better approach would be:
    // os.Setenv("SESSION_AUTHENTICATION_KEY", "test-auth-key-32-bytes-long--")
    // os.Setenv("SESSION_ENCRYPTION_KEY", "test-enc-key-32-bytes-long---")
    // For simplicity, the current InitSessionStore has a hardcoded fallback if rand.Read fails or key is empty.
    // Let's ensure it uses a predictable key for tests by setting ENV var.
    os.Setenv("SESSION_AUTHENTICATION_KEY", "testkey1234567890123456789012345678") // 32 bytes
	sessions.InitSessionStore()

    InitAuthTemplates()
    InitDashboardTemplates()
    InitDeviceTemplates()
    InitMessageTemplates()
    InitSettingsTemplates()

	// Setup a comprehensive router for all handlers in the package
	testRouter = mux.NewRouter()
	// Auth routes
	testRouter.HandleFunc("/register", ShowRegistrationPage).Methods(http.MethodGet)
	testRouter.HandleFunc("/register", HandleRegistration).Methods(http.MethodPost)
	testRouter.HandleFunc("/login", ShowLoginPage).Methods(http.MethodGet)
	testRouter.HandleFunc("/login", HandleLogin).Methods(http.MethodPost)
    testRouter.HandleFunc("/logout", LogoutHandler).Methods(http.MethodGet)
	// Home route
    testRouter.HandleFunc("/", HomeHandler).Methods(http.MethodGet)
	// Settings routes
	testRouter.HandleFunc("/settings", SettingsPageGET).Methods(http.MethodGet)
	testRouter.HandleFunc("/settings/password", HandleChangePassword).Methods(http.MethodPost)
	// Device route
	testRouter.HandleFunc("/devices", ShowDevicesPage).Methods(http.MethodGet)
	// Message route
	testRouter.HandleFunc("/messages", ShowMessagesPage).Methods(http.MethodGet)


	// Run tests
	exitCode := m.Run()

	// Teardown: (In-memory DB doesn't need explicit cleanup of files)
	// If using a file-based test DB, os.Remove("test_dashboard.db") here.
	sqlDB, err := testDB.DB()
    if err == nil {
        sqlDB.Close()
    }
	os.Exit(exitCode)
}

// Helper to clear tables before a test
func clearTables() {
    // Order matters due to foreign key constraints if they are enforced by SQLite build / PRAGMAs.
    // Typically, for tests, disabling FK constraints temporarily or deleting in reverse order of creation is safer.
    // GORM's Delete with an empty condition deletes all records.
    testDB.Exec("DELETE FROM message_counts") // No dependencies
    testDB.Exec("DELETE FROM devices")        // Depends on users
    testDB.Exec("DELETE FROM users")          // Base table
}


func TestShowRegistrationPage_Unauthenticated(t *testing.T) {
    clearTables() // Ensure clean state

	req, _ := http.NewRequest(http.MethodGet, "/register", nil)
	rr := httptest.NewRecorder()
	testRouter.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check if the response body contains some expected text from the registration form
	if !strings.Contains(rr.Body.String(), "Create Account") { // Title might have HTMX now
		t.Errorf("handler returned unexpected body: got %v, expected to contain 'Create Account'", rr.Body.String())
	}
}

func TestHandleRegistration_Success(t *testing.T) {
    clearTables()

	form := url.Values{}
	form.Add("username", "testuserreg") // Unique username for this test
	form.Add("password", "password123")

	req, _ := http.NewRequest(http.MethodPost, "/register", strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	testRouter.ServeHTTP(rr, req)

	// For non-HTMX, expect a 200 OK with success message on the page.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v. Body: %s", status, http.StatusOK, rr.Body.String())
	}

	// Check for success message (HTMX or full page)
	if !strings.Contains(rr.Body.String(), "Registration successful") {
		t.Errorf("handler returned unexpected body: got %v, expected 'Registration successful'", rr.Body.String())
	}

	// Verify user in database
	var user db.User
	if err := testDB.Where("username = ?", "testuserreg").First(&user).Error; err != nil {
		t.Errorf("user 'testuserreg' was not created in database: %v", err)
	}
	if user.Username != "testuserreg" {
		t.Errorf("created user has wrong username: got %s want testuserreg", user.Username)
	}

    // Verify message_count created
    var mc db.MessageCount
    if err := testDB.Where("user_id = ?", user.ID).First(&mc).Error; err != nil {
        t.Errorf("message_count was not created for user %d ('testuserreg'): %v", user.ID, err)
    }
    if mc.Count != 0 {
        t.Errorf("message_count for new user should be 0, got %d", mc.Count)
    }
}

func TestShowRegistrationPage_Authenticated(t *testing.T) {
    clearTables()
    // Create a dummy user
    dummyUser := db.User{Username: "authedtestuser", PasswordHash: "somehash"}
    result := testDB.Create(&dummyUser)
    if result.Error != nil {
        t.Fatalf("Failed to create dummy user for auth test: %v", result.Error)
    }

    req, _ := http.NewRequest(http.MethodGet, "/register", nil)
    rr := httptest.NewRecorder()

    // Create and save a session for the request
    // The session store uses the request to get/save cookies.
    // The recorder (rr) captures the Set-Cookie header from session.Save.
    session, err := sessions.Store.Get(req, sessions.SessionKey)
    if err != nil {
        t.Fatalf("Failed to get session: %v", err)
    }
    session.Values[sessions.UserIDKey] = dummyUser.ID
    err = sessions.Store.Save(req, rr, session) // Use sessions.Store.Save
    if err != nil {
        t.Fatalf("Failed to save session: %v", err)
    }

    // The httptest.Recorder (rr) now has the session cookie.
    // We need to copy this cookie to the new request being sent to the router.
    // This simulates a browser sending the cookie back.
    // This is a common point of confusion in Go HTTP testing.

    // Create a new request for the actual test, now with the cookie.
    authedReq, _ := http.NewRequest(http.MethodGet, "/register", nil)
    for _, cookie := range rr.Result().Cookies() { // rr.Result().Cookies() gets cookies set by session.Save
        authedReq.AddCookie(cookie)
    }

    // Serve the new request that has the session cookie.
    finalRR := httptest.NewRecorder() // Use a new recorder for the actual handler response
    testRouter.ServeHTTP(finalRR, authedReq)

    if status := finalRR.Code; status != http.StatusFound {
        t.Errorf("handler returned wrong status code: got %v want %v. Body: %s", status, http.StatusFound, finalRR.Body.String())
    }

    redirectURL, err := finalRR.Result().Location()
    if err != nil {
        t.Errorf("Failed to get redirect location: %v", err)
    }
    if redirectURL.Path != "/" {
        t.Errorf("Expected redirect to '/', got %s", redirectURL.Path)
    }
}

// TODO: Add more tests for other auth functions:
// - TestHandleRegistration_UsernameExists
// - TestHandleRegistration_ValidationErrors (short password, invalid username etc.)
// - TestShowLoginPage_Unauthenticated
// - TestShowLoginPage_Authenticated
// - TestHandleLogin_Success (check session, HX-Redirect for HTMX)
// - TestHandleLogin_UserNotFound
// - TestHandleLogin_IncorrectPassword
// - TestLogoutHandler

// TestHandleRegistration_UsernameExists tests registration with an already existing username.
func TestHandleRegistration_UsernameExists(t *testing.T) {
    clearTables()
    // 1. Create an initial user
    initialUser := db.User{
        Username:     "existinguser",
        PasswordHash: "somehash", // Hash doesn't matter for this test
        CreatedAt:    time.Now(),
    }
    if err := testDB.Create(&initialUser).Error; err != nil {
        t.Fatalf("Failed to create initial user: %v", err)
    }

    // 2. Attempt to register the same username again (non-HTMX)
    form := url.Values{}
    form.Add("username", "existinguser")
    form.Add("password", "newpassword123")

    req, _ := http.NewRequest(http.MethodPost, "/register", strings.NewReader(form.Encode()))
    req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

    rr := httptest.NewRecorder()
    testRouter.ServeHTTP(rr, req)

    if status := rr.Code; status != http.StatusConflict {
        t.Errorf("Non-HTMX: handler returned wrong status code: got %v want %v. Body: %s", status, http.StatusConflict, rr.Body.String())
    }
    if !strings.Contains(rr.Body.String(), "Username already taken") {
        t.Errorf("Non-HTMX: handler returned unexpected body: got %v, expected 'Username already taken'", rr.Body.String())
    }

    // 3. Attempt to register the same username again (HTMX)
    reqHtmx, _ := http.NewRequest(http.MethodPost, "/register", strings.NewReader(form.Encode()))
    reqHtmx.Header.Add("Content-Type", "application/x-www-form-urlencoded")
    reqHtmx.Header.Add("HX-Request", "true")

    rrHtmx := httptest.NewRecorder()
    testRouter.ServeHTTP(rrHtmx, reqHtmx)

    // HTMX requests might return 200 OK but with an error message in the partial,
    // or a specific error code like 409 or 422. The current HandleRegistration sets 409 via w.WriteHeader
    // before calling renderAuthResponse. renderAuthResponse then renders the partial.
    // So, the status code should ideally reflect the error.
    // However, the provided renderAuthResponse for HTMX path doesn't explicitly set status code based on data.Error.
    // Let's assume for now the handler sets it before calling renderAuthResponse, or that 200 with error message is acceptable.
    // The current implementation of HandleRegistration calls w.WriteHeader(http.StatusConflict) before renderAuthResponse.
    // So, for HTMX, it should also be StatusConflict if WriteHeader is called before template execution for partials.
    // Let's test for StatusConflict for HTMX as well, as the WriteHeader is called regardless of HTMX.
    if status := rrHtmx.Code; status != http.StatusConflict {
         t.Errorf("HTMX: handler returned wrong status code: got %v want %v. Body: %s", status, http.StatusConflict, rrHtmx.Body.String())
    }
    if !strings.Contains(rrHtmx.Body.String(), "Username already taken") {
        t.Errorf("HTMX: handler returned unexpected partial body: got %v, expected 'Username already taken'", rrHtmx.Body.String())
    }
}


func TestHandleRegistration_ValidationErrors(t *testing.T) {
    clearTables()
    tests := []struct {
        name            string
        username        string
        password        string
        expectedMsg     string
        expectedStatus  int // For non-HTMX
        isHtmx          bool
    }{
        {"EmptyUsername_NonHTMX", "", "password123", "Username and password are required", http.StatusBadRequest, false},
        {"EmptyUsername_HTMX", "", "password123", "Username and password are required", http.StatusBadRequest, true},
        {"EmptyPassword_NonHTMX", "user1", "", "Username and password are required", http.StatusBadRequest, false},
        {"EmptyPassword_HTMX", "user1", "", "Username and password are required", http.StatusBadRequest, true},
        {"ShortPassword_NonHTMX", "user2", "12345", "Password must be at least 6 characters long", http.StatusBadRequest, false},
        {"ShortPassword_HTMX", "user2", "12345", "Password must be at least 6 characters long", http.StatusBadRequest, true},
        {"InvalidUsernameChars_NonHTMX", "user!@#", "password123", "Username must be 3-20 characters long", http.StatusBadRequest, false},
        {"InvalidUsernameChars_HTMX", "user!@#", "password123", "Username must be 3-20 characters long", http.StatusBadRequest, true},
        {"UsernameTooShort_NonHTMX", "us", "password123", "Username must be 3-20 characters long", http.StatusBadRequest, false},
        {"UsernameTooShort_HTMX", "us", "password123", "Username must be 3-20 characters long", http.StatusBadRequest, true},
        {"UsernameTooLong_NonHTMX", "averylongusernameover20chars", "password123", "Username must be 3-20 characters long", http.StatusBadRequest, false},
        {"UsernameTooLong_HTMX", "averylongusernameover20chars", "password123", "Username must be 3-20 characters long", http.StatusBadRequest, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            form := url.Values{}
            form.Add("username", tt.username)
            form.Add("password", tt.password)

            req, _ := http.NewRequest(http.MethodPost, "/register", strings.NewReader(form.Encode()))
            req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
            if tt.isHtmx {
                req.Header.Add("HX-Request", "true")
            }

            rr := httptest.NewRecorder()
            testRouter.ServeHTTP(rr, req)

            // As with UsernameExists, HandleRegistration calls WriteHeader before renderAuthResponse.
            // So, the status code should be tt.expectedStatus for both HTMX and non-HTMX.
            if status := rr.Code; status != tt.expectedStatus {
                t.Errorf("%s: handler returned wrong status code: got %v want %v. Body: %s", tt.name, status, tt.expectedStatus, rr.Body.String())
            }
            if !strings.Contains(rr.Body.String(), tt.expectedMsg) {
                t.Errorf("%s: handler returned unexpected body: got %v, expected to contain '%s'", tt.name, rr.Body.String(), tt.expectedMsg)
            }
        })
    }
}

func TestShowLoginPage_Unauthenticated(t *testing.T) {
    clearTables()
    req, _ := http.NewRequest(http.MethodGet, "/login", nil)
    rr := httptest.NewRecorder()
    testRouter.ServeHTTP(rr, req)

    if status := rr.Code; status != http.StatusOK {
        t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
    }
    if !strings.Contains(rr.Body.String(), "Login to Dashboard") {
        t.Errorf("handler returned unexpected body, expected 'Login to Dashboard'")
    }
}

func TestShowLoginPage_Authenticated(t *testing.T) {
    clearTables()
    dummyUser := db.User{Username: "authedloginuser", PasswordHash: "hash"}
    testDB.Create(&dummyUser)

    req, _ := http.NewRequest(http.MethodGet, "/login", nil)
    rr := httptest.NewRecorder()

    session, _ := sessions.Store.Get(req, sessions.SessionKey)
    session.Values[sessions.UserIDKey] = dummyUser.ID
    sessions.Store.Save(req, rr, session)

    authedReq, _ := http.NewRequest(http.MethodGet, "/login", nil)
    for _, cookie := range rr.Result().Cookies() {
        authedReq.AddCookie(cookie)
    }

    finalRR := httptest.NewRecorder()
    testRouter.ServeHTTP(finalRR, authedReq)

    if status := finalRR.Code; status != http.StatusFound {
        t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusFound)
    }
    if loc, _ := finalRR.Result().Location(); loc.Path != "/" {
        t.Errorf("expected redirect to '/', got %s", loc.Path)
    }
}

func TestHandleLogin_Success_NonHTMX(t *testing.T) {
    clearTables()
    hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
    testDB.Create(&db.User{Username: "normallogin", PasswordHash: string(hashedPassword)})

    form := url.Values{}
    form.Add("username", "normallogin")
    form.Add("password", "password123")

    req, _ := http.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
    req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

    rr := httptest.NewRecorder()
    testRouter.ServeHTTP(rr, req)

    if status := rr.Code; status != http.StatusFound {
        t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusFound)
    }
    if loc, _ := rr.Result().Location(); loc.Path != "/" {
        t.Errorf("expected redirect to '/', got %s", loc.Path)
    }

    // Check session cookie
    var userID uint
    rawCookies := rr.Result().Cookies()
    parsedSession, _ := sessions.Store.Get(&http.Request{Header: http.Header{"Cookie": {rawCookies[0].String()}}}, sessions.SessionKey)
    if idVal, ok := parsedSession.Values[sessions.UserIDKey]; ok {
        userID = idVal.(uint)
    }
    if userID == 0 {
        t.Errorf("session was not created or userID not set correctly")
    }
}

func TestHandleLogin_Success_HTMX(t *testing.T) {
    clearTables()
    hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
    testDB.Create(&db.User{Username: "htmxlogin", PasswordHash: string(hashedPassword)})

    form := url.Values{}
    form.Add("username", "htmxlogin")
    form.Add("password", "password123")

    req, _ := http.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
    req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
    req.Header.Add("HX-Request", "true")

    rr := httptest.NewRecorder()
    testRouter.ServeHTTP(rr, req)

    if status := rr.Code; status != http.StatusOK { // HandleLogin sends HX-Redirect, which is a 200 OK with special header
        t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
    }
    if hxRedirect := rr.Header().Get("HX-Redirect"); hxRedirect != "/" {
        t.Errorf("expected HX-Redirect header to '/', got %s", hxRedirect)
    }
    // Session check similar to NonHTMX can be added here too
}

func TestHandleLogin_UserNotFound(t *testing.T) {
    clearTables()
    form := url.Values{}
    form.Add("username", "nonexistentuser")
    form.Add("password", "password123")

    // Non-HTMX
    reqNonHtmx, _ := http.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
    reqNonHtmx.Header.Add("Content-Type", "application/x-www-form-urlencoded")
    rrNonHtmx := httptest.NewRecorder()
    testRouter.ServeHTTP(rrNonHtmx, reqNonHtmx)
    if status := rrNonHtmx.Code; status != http.StatusUnauthorized {
        t.Errorf("Non-HTMX: Expected status %v, got %v", http.StatusUnauthorized, status)
    }
    if !strings.Contains(rrNonHtmx.Body.String(), "Invalid username or password") {
        t.Errorf("Non-HTMX: Expected error message not found in body")
    }

    // HTMX
    reqHtmx, _ := http.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
    reqHtmx.Header.Add("Content-Type", "application/x-www-form-urlencoded")
    reqHtmx.Header.Add("HX-Request", "true")
    rrHtmx := httptest.NewRecorder()
    testRouter.ServeHTTP(rrHtmx, reqHtmx)
    // HandleLogin calls WriteHeader(http.StatusUnauthorized) before renderAuthResponse
    if status := rrHtmx.Code; status != http.StatusUnauthorized {
         t.Errorf("HTMX: Expected status %v, got %v. Body: %s", http.StatusUnauthorized, status, rrHtmx.Body.String())
    }
    if !strings.Contains(rrHtmx.Body.String(), "Invalid username or password") {
        t.Errorf("HTMX: Expected error message not found in partial body")
    }
}

func TestHandleLogin_IncorrectPassword(t *testing.T) {
    clearTables()
    hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("realpassword"), bcrypt.DefaultCost)
    testDB.Create(&db.User{Username: "correctuser", PasswordHash: string(hashedPassword)})

    form := url.Values{}
    form.Add("username", "correctuser")
    form.Add("password", "wrongpassword")

    // Non-HTMX
    reqNonHtmx, _ := http.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
    reqNonHtmx.Header.Add("Content-Type", "application/x-www-form-urlencoded")
    rrNonHtmx := httptest.NewRecorder()
    testRouter.ServeHTTP(rrNonHtmx, reqNonHtmx)
    if status := rrNonHtmx.Code; status != http.StatusUnauthorized {
        t.Errorf("Non-HTMX: Expected status %v, got %v", http.StatusUnauthorized, status)
    }
    if !strings.Contains(rrNonHtmx.Body.String(), "Invalid username or password") {
        t.Errorf("Non-HTMX: Expected error message not found")
    }

    // HTMX
    reqHtmx, _ := http.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
    reqHtmx.Header.Add("Content-Type", "application/x-www-form-urlencoded")
    reqHtmx.Header.Add("HX-Request", "true")
    rrHtmx := httptest.NewRecorder()
    testRouter.ServeHTTP(rrHtmx, reqHtmx)
    if status := rrHtmx.Code; status != http.StatusUnauthorized {
        t.Errorf("HTMX: Expected status %v, got %v", http.StatusUnauthorized, status)
    }
    if !strings.Contains(rrHtmx.Body.String(), "Invalid username or password") {
        t.Errorf("HTMX: Expected error message not found in partial")
    }
}

func TestLogoutHandler(t *testing.T) {
    clearTables()
    dummyUser := db.User{Username: "logoutuser", PasswordHash: "hash"}
    testDB.Create(&dummyUser)

    // 1. Simulate login to get a session cookie
    loginForm := url.Values{}
    loginForm.Add("username", "logoutuser")
    loginForm.Add("password", "hash") // This test doesn't check password, just needs a valid user for session

    // Temporarily override password for login to succeed without bcrypt for this setup
    // This is a bit of a hack. Better would be to use bcrypt for the dummy user.
    // For now, let's assume login works and gives us a cookie.
    // A more robust way: create user with known bcrypt hash, or mock bcrypt.Compare.
    // Let's just create a session manually for logout test simplicity.

    reqLogout, _ := http.NewRequest(http.MethodGet, "/logout", nil)
    rrLogoutSetup := httptest.NewRecorder() // To capture cookie from manual session save

    session, _ := sessions.Store.Get(reqLogout, sessions.SessionKey)
    session.Values[sessions.UserIDKey] = dummyUser.ID
    sessions.Store.Save(reqLogout, rrLogoutSetup, session) // Save session to get cookie in rrLogoutSetup

    // Add cookie to actual logout request
    for _, cookie := range rrLogoutSetup.Result().Cookies() {
        reqLogout.AddCookie(cookie)
    }

    // 2. Perform logout
    rr := httptest.NewRecorder()
    testRouter.ServeHTTP(rr, reqLogout)

    if status := rr.Code; status != http.StatusFound {
        t.Errorf("Logout: Expected status %v, got %v", http.StatusFound, status)
    }
    if loc, _ := rr.Result().Location(); loc.Path != "/login" {
        t.Errorf("Logout: Expected redirect to /login, got %s", loc.Path)
    }

    // Verify cookie is cleared (MaxAge < 0 or Expires in the past)
    foundCookie := false
    for _, cookie := range rr.Result().Cookies() {
        if cookie.Name == sessions.SessionKey {
            foundCookie = true
            if cookie.MaxAge >= 0 {
                // Note: Some browsers/libraries might interpret MaxAge=0 as session cookie, not deletion.
                // gorilla/sessions sets MaxAge = -1 for deletion.
                t.Errorf("Logout: Expected session cookie to have MaxAge < 0, got %d", cookie.MaxAge)
            }
            break
        }
    }
    if !foundCookie {
        t.Errorf("Logout: Session cookie not found in response, expected it to be set for clearing.")
    }
}

// Ensure time is imported for TestHandleRegistration_UsernameExists
// Ensure bcrypt is imported for TestHandleLogin_Success_NonHTMX and TestHandleLogin_Success_HTMX
// (It is imported in handlers, but for test setup, sometimes direct use is needed)
// "time" is already imported in the original test file.
// "golang.org/x/crypto/bcrypt" is not directly used in test functions here but in handlers.
