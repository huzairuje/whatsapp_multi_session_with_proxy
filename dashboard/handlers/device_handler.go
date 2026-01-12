package handlers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"whatsapp_multi_session/commandhandler"
	"whatsapp_multi_session/dashboard/models"
	"whatsapp_multi_session/utils" // For utils.ParseJID

	"github.com/gin-gonic/gin"
	"go.mau.fi/whatsmeow/types"
	"gorm.io/gorm"
)

// DeviceHandler holds dependencies for device management handlers
type DeviceHandler struct {
	DB         *gorm.DB
	CmdHandler *commandhandler.CommandHandler
}

// NewDeviceHandler creates a new DeviceHandler
func NewDeviceHandler(db *gorm.DB, cmdHandler *commandhandler.CommandHandler) *DeviceHandler {
	if db == nil {
		// Or panic, depending on desired behavior for uninitialized dependencies
		// log.Fatal("DeviceHandler: gorm.DB is nil")
	}
	if cmdHandler == nil {
		// Or panic
		// log.Fatal("DeviceHandler: commandhandler.CommandHandler is nil")
	}
	return &DeviceHandler{
		DB:         db,
		CmdHandler: cmdHandler,
	}
}

// --- Request/Response Structs ---

type AddDeviceRequest struct {
	DeviceJID  string `json:"device_jid" binding:"required"`
	DeviceName string `json:"device_name" binding:"required"`
}

// ManagedDeviceResponse is used for listing devices, including their live status
type ManagedDeviceResponse struct {
	ID              uint      `json:"id"`
	AdminUserID     uint      `json:"admin_user_id"`
	DeviceJID       string    `json:"device_jid"`
	DeviceName      string    `json:"device_name"`
	IsLoggedInLive  bool      `json:"is_logged_in_live"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// --- Handler Implementations ---

// AddDevice handles adding a new device to be managed by an admin user
func (h *DeviceHandler) AddDevice(c *gin.Context) {
	adminUserIDVal, exists := c.Get("admin_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Admin user ID not found in token claims"})
		return
	}
	// adminUserID, ok := adminUserIDVal.(uint) // Previous assumption
	adminUserIDFloat, ok := adminUserIDVal.(float64) // JWT numbers are often float64
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid admin user ID type in token claims"})
		return
	}
	adminUserID := uint(adminUserIDFloat)

	var req AddDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	// Validate input
	if strings.TrimSpace(req.DeviceJID) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Device JID cannot be empty"})
		return
	}
	if strings.TrimSpace(req.DeviceName) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Device name cannot be empty"})
		return
	}

	// Parse DeviceJID
	parsedDeviceJID, ok := utils.ParseJID(req.DeviceJID)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Device JID format. Ensure it's a valid WhatsApp JID (e.g., 1234567890@s.whatsapp.net or just 1234567890)."})
		return
	}

	// Validation with main app: Check if the device is known to whatsmeow
	// Note: CmdHandler.Container might be nil if not initialized properly in the main app
	// or if the CommandHandler struct changes. Assuming Container is the sqlstore.Container.
	if h.CmdHandler == nil || h.CmdHandler.Container == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Device validation service not available."})
		return
	}
	_, err := h.CmdHandler.Container.GetDevice(context.Background(), parsedDeviceJID) // Use c.Request.Context() if preferred
	if err != nil {
		// This error usually means "sql: no rows in result set" if not found
		// We treat any error here as the device not being registered/known by the system.
		c.JSON(http.StatusNotFound, gin.H{"error": "Device JID " + parsedDeviceJID.String() + " not registered or recognized by the underlying WhatsApp system. Please ensure this device has connected to the service before."})
		return
	}

	// Check for duplicates for this admin and device JID
	var existingManagedDevice models.ManagedDevice
	if err := h.DB.Where("admin_user_id = ? AND device_jid = ?", adminUserID, parsedDeviceJID.String()).First(&existingManagedDevice).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "This device JID is already managed by you."})
		return
	} else if err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error checking for existing device: " + err.Error()})
		return
	}

	// Create new ManagedDevice record
	newManagedDevice := models.ManagedDevice{
		AdminUserID: adminUserID,
		DeviceJID:   parsedDeviceJID.String(), // Store the standardized JID string
		DeviceName:  req.DeviceName,
	}

	if err := h.DB.Create(&newManagedDevice).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add managed device: " + err.Error()})
		return
	}

	// Prepare response, similar to ManagedDeviceResponse but without live status for this specific call
	response := gin.H{
		"id":            newManagedDevice.ID,
		"admin_user_id": newManagedDevice.AdminUserID,
		"device_jid":    newManagedDevice.DeviceJID,
		"device_name":   newManagedDevice.DeviceName,
		"created_at":    newManagedDevice.CreatedAt,
		"updated_at":    newManagedDevice.UpdatedAt,
	}
	c.JSON(http.StatusCreated, response)
}

// ListDevices lists all devices managed by the authenticated admin user
func (h *DeviceHandler) ListDevices(c *gin.Context) {
	adminUserIDVal, exists := c.Get("admin_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Admin user ID not found in token claims"})
		return
	}
	// adminUserID, ok := adminUserIDVal.(uint)
	adminUserIDFloat, ok := adminUserIDVal.(float64) // JWT numbers are often float64
    if !ok {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid admin user ID type in token claims"})
        return
    }
    adminUserID := uint(adminUserIDFloat)

	var managedDevices []models.ManagedDevice
	if err := h.DB.Where("admin_user_id = ?", adminUserID).Find(&managedDevices).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve managed devices: " + err.Error()})
		return
	}

	var responseDevices []ManagedDeviceResponse
	for _, device := range managedDevices {
		parsedJID, pOk := utils.ParseJID(device.DeviceJID)
		isLoggedIn := false
		if pOk && h.CmdHandler != nil {
			// commandhandler.LoadClientConcurrent expects JID.User part
			client, clientOk := commandhandler.LoadClientConcurrent(parsedJID.User)
			if clientOk && client != nil && client.IsLoggedIn() {
				isLoggedIn = true
			}
		}

		responseDevices = append(responseDevices, ManagedDeviceResponse{
			ID:              device.ID,
			AdminUserID:     device.AdminUserID,
			DeviceJID:       device.DeviceJID,
			DeviceName:      device.DeviceName,
			IsLoggedInLive:  isLoggedIn,
			CreatedAt:       device.CreatedAt,
			UpdatedAt:       device.UpdatedAt,
		})
	}

	if responseDevices == nil {
		responseDevices = []ManagedDeviceResponse{} // Return empty list instead of null
	}
	c.JSON(http.StatusOK, responseDevices)
}
