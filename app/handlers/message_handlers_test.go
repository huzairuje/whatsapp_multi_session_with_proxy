package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"whatsapp_multi_session/app/db"
	"whatsapp_multi_session/app/sessions"
	// "log"
	// "os"
	// "github.com/gorilla/mux"
	// "gorm.io/driver/sqlite"
	// "gorm.io/gorm"
)

// testDB and testRouter are global, defined in auth_handlers_test.go's TestMain

func clearTables_message() {
	testDB.Exec("DELETE FROM message_counts") // Use global testDB
	testDB.Exec("DELETE FROM devices")
	testDB.Exec("DELETE FROM users")
}

// (User is created within tests for message page, this helper just sets up session)
func createAuthenticatedUserAndRequest_message(t *testing.T, user db.User, method, urlStr string) *http.Request {
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

func TestShowMessagesPage_Authenticated_WithMessages(t *testing.T) {
	clearTables_message()
	user := db.User{Username: "messageUser1", PasswordHash: "hash", CreatedAt: time.Now()}
	testDB.Create(&user) // Use global testDB

	msgTime := time.Now().Add(-5 * time.Minute)
	testDB.Create(&db.MessageCount{UserID: user.ID, Count: 42, LastUpdatedAt: msgTime}) // Use global testDB

	req := createAuthenticatedUserAndRequest_message(t, user, http.MethodGet, "/messages")
	rr := httptest.NewRecorder()
	testRouter.ServeHTTP(rr, req) // Use global testRouter

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v. Body: %s", status, http.StatusOK, rr.Body.String())
	}
	body := rr.Body.String()
	if !strings.Contains(body, "Message Summary") {
		t.Errorf("expected 'Message Summary' title not found")
	}
	expectedCountStr := "<strong class=\"text-green-600\">42</strong>"
	if !strings.Contains(body, expectedCountStr) {
		t.Errorf("expected message count '42' not found or not matching format. Body: %s", body)
	}
	formattedDate := msgTime.Format("Jan 02, 2006 03:04 PM")
	if !strings.Contains(body, formattedDate) {
		t.Errorf("expected formatted LastUpdatedAt '%s' not found. Body: %s", formattedDate, body)
	}
}

func TestShowMessagesPage_Authenticated_NoMessagesRecord(t *testing.T) {
	clearTables_message()
	user := db.User{Username: "messageUser2", PasswordHash: "hash", CreatedAt: time.Now()}
	testDB.Create(&user) // Use global testDB
	// No MessageCount record is created for this user

	req := createAuthenticatedUserAndRequest_message(t, user, http.MethodGet, "/messages")
	rr := httptest.NewRecorder()
	testRouter.ServeHTTP(rr, req) // Use global testRouter

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v. Body: %s", status, http.StatusOK, rr.Body.String())
	}
	// The handler sets MessageCount to nil if record not found.
	// The template `messages.html` has: {{if .MessageCount}} ... {{else}} <p>No message activity...</p> {{end}}
	if !strings.Contains(rr.Body.String(), "No message activity recorded yet") {
		t.Errorf("expected 'No message activity' message not found. Body: %s", rr.Body.String())
	}
}

func TestShowMessagesPage_Authenticated_ZeroCountMessages(t *testing.T) {
	clearTables_message()
	user := db.User{Username: "messageUser3", PasswordHash: "hash", CreatedAt: time.Now()}
	testDB.Create(&user) // Use global testDB

	msgTime := time.Now().Add(-10 * time.Minute)
	testDB.Create(&db.MessageCount{UserID: user.ID, Count: 0, LastUpdatedAt: msgTime}) // Use global testDB


	req := createAuthenticatedUserAndRequest_message(t, user, http.MethodGet, "/messages")
	rr := httptest.NewRecorder()
	testRouter.ServeHTTP(rr, req) // Use global testRouter

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v. Body: %s", status, http.StatusOK, rr.Body.String())
	}
	body := rr.Body.String()
	expectedCountStr := "<strong class=\"text-green-600\">0</strong>"
	if !strings.Contains(body, expectedCountStr) {
		t.Errorf("expected message count '0' not found or not matching format. Body: %s", body)
	}
	formattedDate := msgTime.Format("Jan 02, 2006 03:04 PM")
	if !strings.Contains(body, formattedDate) {
		t.Errorf("expected formatted LastUpdatedAt '%s' for zero count not found. Body: %s", formattedDate, body)
	}
}


func TestShowMessagesPage_Unauthenticated(t *testing.T) {
	clearTables_message()
	req, _ := http.NewRequest(http.MethodGet, "/messages", nil)
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
