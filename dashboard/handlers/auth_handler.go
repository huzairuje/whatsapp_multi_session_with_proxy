package handlers

import (
	"net/http"
	"strings"
	"time"

	"whatsapp_multi_session/commandhandler"
	"whatsapp_multi_session/dashboard/auth"
	"whatsapp_multi_session/dashboard/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	otpValidityDuration = 5 * time.Minute  // OTP is valid for 5 minutes
	tokenDuration       = 24 * time.Hour // JWT is valid for 24 hours
	otpLength           = 6                // Length of the OTP
)

// AuthHandler holds dependencies for authentication handlers
type AuthHandler struct {
	DB             *gorm.DB
	CmdHandler     *commandhandler.CommandHandler
	JwtSecretKey   string
	OtpSenderJID   string
	TokenBlacklist map[string]time.Time // For a simple blacklist, if needed for logout
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(db *gorm.DB, cmdHandler *commandhandler.CommandHandler, jwtSecretKey string, otpSenderJID string) *AuthHandler {
	return &AuthHandler{
		DB:             db,
		CmdHandler:     cmdHandler,
		JwtSecretKey:   jwtSecretKey,
		OtpSenderJID:   otpSenderJID,
		TokenBlacklist: make(map[string]time.Time), // Initialize blacklist if used
	}
}

// --- Request/Response Structs ---

type RegisterRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
}

type OTPRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
}

type LoginRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
	OTP         string `json:"otp" binding:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type AuthDetailResponse struct {
	ID          uint      `json:"id"`
	PhoneNumber string    `json:"phone_number"`
	CreatedAt   time.Time `json:"created_at"`
}

// --- Handler Implementations ---

// Register handles new admin user registration
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	// Basic phone number validation (can be more sophisticated)
	if strings.TrimSpace(req.PhoneNumber) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Phone number cannot be empty"})
		return
	}

	var existingUser models.AdminUser
	if err := h.DB.Where("phone_number = ?", req.PhoneNumber).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Admin user with this phone number already exists"})
		return
	} else if err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error: " + err.Error()})
		return
	}

	newUser := models.AdminUser{
		PhoneNumber: req.PhoneNumber,
		// HashedOTP and OTPGeneratedAt are initially null (zero value for pointers)
	}

	if err := h.DB.Create(&newUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create admin user: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Admin user registered successfully. Please request an OTP to login."})
}

// RequestOTP handles OTP generation and sending
func (h *AuthHandler) RequestOTP(c *gin.Context) {
	var req OTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	var user models.AdminUser
	if err := h.DB.Where("phone_number = ?", req.PhoneNumber).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Admin user not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error: " + err.Error()})
		}
		return
	}

	plainOTP, err := auth.GenerateOTP(otpLength)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate OTP: " + err.Error()})
		return
	}

	hashedOTP, err := auth.HashOTP(plainOTP)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash OTP: " + err.Error()})
		return
	}

	now := time.Now()
	user.HashedOTP = &hashedOTP
	user.OTPGeneratedAt = &now

	if err := h.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save OTP details: " + err.Error()})
		return
	}

	// Ensure CmdHandler and OtpSenderJID are configured
	if h.CmdHandler == nil || h.OtpSenderJID == "" {
		// Log this issue for the admin, but don't expose internal config details to the user.
		// For the user, the OTP sending will appear to have failed.
		// Consider if this should be a startup check for the handler.
		c.JSON(http.StatusInternalServerError, gin.H{"error": "OTP sending service not configured. OTP generated but not sent."})
		return
	}

	err = auth.SendOTPMessage(h.CmdHandler, h.OtpSenderJID, user.PhoneNumber, plainOTP)
	if err != nil {
		// Log the detailed error for debugging.
		// For the user, provide a generic message.
		// It's important not to reveal too much about why sending failed (e.g., invalid sender JID).
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send OTP. Please try again later."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "OTP sent successfully to " + user.PhoneNumber})
}

// Login handles admin user login with OTP
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	var user models.AdminUser
	if err := h.DB.Where("phone_number = ?", req.PhoneNumber).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Admin user not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error: " + err.Error()})
		}
		return
	}

	if user.HashedOTP == nil || user.OTPGeneratedAt == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "OTP not requested or already used. Please request a new OTP."})
		return
	}

	if time.Since(*user.OTPGeneratedAt) > otpValidityDuration {
		// Optionally clear the expired OTP
		user.HashedOTP = nil
		user.OTPGeneratedAt = nil
		h.DB.Save(&user) // Best effort save
		c.JSON(http.StatusUnauthorized, gin.H{"error": "OTP expired. Please request a new one."})
		return
	}

	err := auth.VerifyOTP(*user.HashedOTP, req.OTP)
	if err != nil {
		// bcrypt.ErrMismatchedHashAndPassword indicates wrong OTP
		if err == bcrypt.ErrMismatchedHashAndPassword {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid OTP provided."})
		} else {
			// Other errors during verification
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error verifying OTP: " + err.Error()})
		}
		return
	}

	// OTP is valid, generate JWT
	tokenString, err := auth.GenerateJWT(user.ID, user.PhoneNumber, h.JwtSecretKey, tokenDuration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token: " + err.Error()})
		return
	}

	// Clear OTP after successful login
	user.HashedOTP = nil
	user.OTPGeneratedAt = nil
	if err := h.DB.Save(&user).Error; err != nil {
		// Log this error, but the user has successfully logged in.
		// This is a cleanup step.
		// log.Printf("Warning: Failed to clear OTP for user %d: %v", user.ID, err)
	}

	c.JSON(http.StatusOK, LoginResponse{Token: tokenString})
}

// GetAuthDetail retrieves details of the authenticated admin user
func (h *AuthHandler) GetAuthDetail(c *gin.Context) {
	adminUserIDVal, exists := c.Get("admin_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Admin user ID not found in token claims (middleware error?)"})
		return
	}
	// adminUserID, ok := adminUserIDVal.(uint) // This would panic if adminUserIDVal is float64
    adminUserIDFloat, ok := adminUserIDVal.(float64) // JWT numbers are often float64
    if !ok {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid admin user ID type in token claims"})
        return
    }
    adminUserID := uint(adminUserIDFloat)


	phoneNumberVal, exists := c.Get("phone_number")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Phone number not found in token claims (middleware error?)"})
		return
	}
	phoneNumber, ok := phoneNumberVal.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid phone number type in token claims"})
		return
	}

	var user models.AdminUser
	// Fetch using ID and phone number for extra verification, though ID should be sufficient
	if err := h.DB.Where("id = ? AND phone_number = ?", adminUserID, phoneNumber).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Authenticated admin user not found in database"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error: " + err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, AuthDetailResponse{
		ID:          user.ID,
		PhoneNumber: user.PhoneNumber,
		CreatedAt:   user.CreatedAt,
	})
}

// Logout handles admin user logout (stateless for now)
func (h *AuthHandler) Logout(c *gin.Context) {
	// For stateless JWT, client is responsible for discarding the token.
	// If a blacklist is implemented, the token could be added here.
	// Example:
	// tokenString := c.GetHeader("Authorization") // Assuming Bearer token
	// if strings.HasPrefix(tokenString, "Bearer ") {
	// token := strings.TrimPrefix(tokenString, "Bearer ")
	// h.TokenBlacklist[token] = time.Now().Add(tokenDuration) // Store with its expiry
	// }
	// Cleanup expired tokens from blacklist periodically in a real scenario.

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully. Please discard your token."})
}
