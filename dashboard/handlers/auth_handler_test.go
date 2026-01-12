package handlers

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"whatsapp_multi_session/dashboard/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Helper function to setup GORM with sqlmock
func setupMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp)) // Use regexp matching
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
		// PreferSimpleProtocol: true, // Recommended for pgx, might affect query format with sqlmock
	}), &gorm.Config{})

	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a gorm database connection: %v", err, err)
	}
	return gormDB, mock
}

// AnyTime argument for sqlmock
type AnyTime struct{}

// Match satisfies sqlmock.Argument interface
func (a AnyTime) Match(v driver.Value) bool {
	_, ok := v.(time.Time)
	return ok
}

func TestAuthHandler_Register(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// --- Test Case 1: Successful Registration ---
	t.Run("Successful Registration", func(t *testing.T) {
		gormDB, mock := setupMockDB(t)
		authHandler := NewAuthHandler(gormDB, nil, "test-secret", "test-sender-jid")

		rr := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rr)

		requestBody := RegisterRequest{PhoneNumber: "1234567890"}
		jsonBody, _ := json.Marshal(requestBody)
		req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req

		// 1. Expect query for existing user (return no rows -> gorm.ErrRecordNotFound)
		// GORM's First method constructs a query like: SELECT * FROM "admin_users" WHERE phone_number = $1 ORDER BY "admin_users"."id" LIMIT 1
		expectedSQLSelect := `SELECT \* FROM "admin_users" WHERE phone_number = \$1 ORDER BY "admin_users"\."id" LIMIT 1`
		mock.ExpectQuery(expectedSQLSelect).
			WithArgs(requestBody.PhoneNumber).
			WillReturnError(gorm.ErrRecordNotFound) // Simulate user not found

		// 2. Expect insert for new user
		// GORM's Create method for Postgres usually uses RETURNING "id" or similar
		// The exact fields depend on your model and GORM's behavior.
		// Using AnyArg for ID, CreatedAt, UpdatedAt, HashedOTP, OTPGeneratedAt
		expectedSQLInsert := `INSERT INTO "admin_users" \("phone_number","hashed_otp","otp_generated_at","created_at","updated_at","id"\) VALUES \(\$1,\$2,\$3,\$4,\$5,\$6\) RETURNING "id"`
		mock.ExpectBegin()
		mock.ExpectQuery(expectedSQLInsert). // Use ExpectQuery for RETURNING
									WithArgs(requestBody.PhoneNumber, nil, nil, AnyTime{}, AnyTime{}, sqlmock.AnyArg()).
									WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1)) // Return a new ID
		mock.ExpectCommit()

		authHandler.Register(c)

		assert.Equal(t, http.StatusCreated, rr.Code)
		expectedResponse := gin.H{"message": "Admin user registered successfully. Please request an OTP to login."}
		var actualResponse gin.H
		err := json.Unmarshal(rr.Body.Bytes(), &actualResponse)
		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, actualResponse)

		assert.NoError(t, mock.ExpectationsWereMet(), "DB expectations not met")
	})

	// --- Test Case 2: User Already Exists ---
	t.Run("User Already Exists", func(t *testing.T) {
		gormDB, mock := setupMockDB(t)
		authHandler := NewAuthHandler(gormDB, nil, "test-secret", "test-sender-jid")

		rr := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rr)

		requestBody := RegisterRequest{PhoneNumber: "0987654321"}
		jsonBody, _ := json.Marshal(requestBody)
		req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req

		// Expect query for existing user (return a row)
		rows := sqlmock.NewRows([]string{"id", "phone_number", "created_at", "updated_at"}).
			AddRow(1, requestBody.PhoneNumber, time.Now(), time.Now()) // Mock an existing user
		expectedSQLSelect := `SELECT \* FROM "admin_users" WHERE phone_number = \$1 ORDER BY "admin_users"\."id" LIMIT 1`
		mock.ExpectQuery(expectedSQLSelect).
			WithArgs(requestBody.PhoneNumber).
			WillReturnRows(rows)

		authHandler.Register(c)

		assert.Equal(t, http.StatusConflict, rr.Code)
		expectedResponse := gin.H{"error": "Admin user with this phone number already exists"}
		var actualResponse gin.H
		err := json.Unmarshal(rr.Body.Bytes(), &actualResponse)
		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, actualResponse)

		assert.NoError(t, mock.ExpectationsWereMet(), "DB expectations not met")
	})

	// --- Test Case 3: Invalid Input (Missing phone_number) ---
	t.Run("Invalid Input - Missing PhoneNumber", func(t *testing.T) {
		gormDB, _ := setupMockDB(t) // Mock is not strictly needed here but setup provides DB
		authHandler := NewAuthHandler(gormDB, nil, "test-secret", "test-sender-jid")

		rr := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rr)

		// Invalid body (empty JSON)
		req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer([]byte("{}")))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req

		authHandler.Register(c)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		// The error message comes from Gin's binding:"required"
		// "Key: 'RegisterRequest.PhoneNumber' Error:Field validation for 'PhoneNumber' failed on the 'required' tag"
		// We can check for a substring.
		var actualResponse gin.H
		err := json.Unmarshal(rr.Body.Bytes(), &actualResponse)
		assert.NoError(t, err)
		assert.Contains(t, actualResponse["error"], "Invalid request body")
		assert.Contains(t, actualResponse["error"], "PhoneNumber")
	})

	// --- Test Case 4: Invalid Input (Empty phone_number string) ---
	t.Run("Invalid Input - Empty PhoneNumber String", func(t *testing.T) {
		gormDB, _ := setupMockDB(t)
		authHandler := NewAuthHandler(gormDB, nil, "test-secret", "test-sender-jid")

		rr := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rr)

		requestBody := RegisterRequest{PhoneNumber: "  "} // Empty string / whitespace
		jsonBody, _ := json.Marshal(requestBody)
		req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req

		authHandler.Register(c)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		expectedResponse := gin.H{"error": "Phone number cannot be empty"}
		var actualResponse gin.H
		err := json.Unmarshal(rr.Body.Bytes(), &actualResponse)
		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, actualResponse)
		// No DB interaction expected
	})

	// --- Test Case 5: Database error on checking existence ---
	t.Run("Database Error on User Existence Check", func(t *testing.T) {
		gormDB, mock := setupMockDB(t)
		authHandler := NewAuthHandler(gormDB, nil, "test-secret", "test-sender-jid")

		rr := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rr)

		requestBody := RegisterRequest{PhoneNumber: "111222333"}
		jsonBody, _ := json.Marshal(requestBody)
		req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req

		expectedSQLSelect := `SELECT \* FROM "admin_users" WHERE phone_number = \$1 ORDER BY "admin_users"\."id" LIMIT 1`
		mock.ExpectQuery(expectedSQLSelect).
			WithArgs(requestBody.PhoneNumber).
			WillReturnError(gorm.ErrUnsupportedDriver) // Simulate a generic DB error

		authHandler.Register(c)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		var actualResponse gin.H
		err := json.Unmarshal(rr.Body.Bytes(), &actualResponse)
		assert.NoError(t, err)
		assert.Contains(t, actualResponse["error"], "Database error")
		assert.NoError(t, mock.ExpectationsWereMet(), "DB expectations not met")
	})

	// --- Test Case 6: Database error on user creation ---
	t.Run("Database Error on User Creation", func(t *testing.T) {
		gormDB, mock := setupMockDB(t)
		authHandler := NewAuthHandler(gormDB, nil, "test-secret", "test-sender-jid")

		rr := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rr)

		requestBody := RegisterRequest{PhoneNumber: "444555666"}
		jsonBody, _ := json.Marshal(requestBody)
		req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req

		// 1. Expect query for existing user (return no rows)
		expectedSQLSelect := `SELECT \* FROM "admin_users" WHERE phone_number = \$1 ORDER BY "admin_users"\."id" LIMIT 1`
		mock.ExpectQuery(expectedSQLSelect).
			WithArgs(requestBody.PhoneNumber).
			WillReturnError(gorm.ErrRecordNotFound)

		// 2. Expect insert for new user, but it fails
		expectedSQLInsert := `INSERT INTO "admin_users" \("phone_number","hashed_otp","otp_generated_at","created_at","updated_at","id"\) VALUES \(\$1,\$2,\$3,\$4,\$5,\$6\) RETURNING "id"`
		mock.ExpectBegin()
		mock.ExpectQuery(expectedSQLInsert).
			WithArgs(requestBody.PhoneNumber, nil, nil, AnyTime{}, AnyTime{}, sqlmock.AnyArg()).
			WillReturnError(gorm.ErrInvalidDB) // Simulate DB error on insert
		mock.ExpectRollback() // GORM should rollback on error

		authHandler.Register(c)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		var actualResponse gin.H
		err := json.Unmarshal(rr.Body.Bytes(), &actualResponse)
		assert.NoError(t, err)
		assert.Contains(t, actualResponse["error"], "Failed to create admin user")
		assert.NoError(t, mock.ExpectationsWereMet(), "DB expectations not met")
	})
}
