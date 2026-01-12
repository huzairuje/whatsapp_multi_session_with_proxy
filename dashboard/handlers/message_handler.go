package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"whatsapp_multi_session/commandhandler"
	"whatsapp_multi_session/dashboard/models"
	"whatsapp_multi_session/utils" // For utils.ParseJID

	"github.com/gin-gonic/gin"
	"go.mau.fi/whatsmeow/types"
	"gorm.io/gorm"
)

// MessageHandler holds dependencies for message sending and reporting handlers
type MessageHandler struct {
	DB         *gorm.DB
	CmdHandler *commandhandler.CommandHandler
}

// NewMessageHandler creates a new MessageHandler
func NewMessageHandler(db *gorm.DB, cmdHandler *commandhandler.CommandHandler) *MessageHandler {
	return &MessageHandler{
		DB:         db,
		CmdHandler: cmdHandler,
	}
}

// --- Request/Response Structs ---

type SendMessageRequest struct {
	DeviceJID       string `json:"device_jid" binding:"required"`
	RecipientNumber string `json:"recipient_number" binding:"required"`
	MessageContent  string `json:"message_content" binding:"required"`
}

type SendBulkMessageRequest struct {
	DeviceJID        string   `json:"device_jid" binding:"required"`
	RecipientNumbers []string `json:"recipient_numbers" binding:"required"`
	MessageContent   string   `json:"message_content" binding:"required"`
}

// BulkMessageReportItem represents the status of sending a message to a single recipient in a bulk operation.
type BulkMessageReportItem struct {
	RecipientNumber string `json:"recipient_number"`
	Status          string `json:"status"` // e.g., "sent", "failed"
	Error           string `json:"error,omitempty"`
	MessageID       string `json:"message_id,omitempty"` // WhatsApp Message ID
}

type SendBulkMessageResponse struct {
	OverallStatus string                  `json:"overall_status"` // e.g., "completed_with_errors", "completed_successfully"
	Results         []BulkMessageReportItem `json:"results"`
}

// SentMessageResponse is used for the message report
type SentMessageResponse struct {
	ID                   uint      `json:"id"`
	AdminUserID          uint      `json:"admin_user_id"`
	DeviceJID            string    `json:"device_jid"`
	RecipientNumber      string    `json:"recipient_number"`
	MessageContent       string    `json:"message_content"`
	Status               string    `json:"status"`
	SentAt               time.Time `json:"sent_at"`
	MessageIDFromWhatsApp *string   `json:"message_id_from_whatsapp"`
	CreatedAt            time.Time `json:"created_at"`
}

type MessageReportResponse struct {
	Messages    []SentMessageResponse `json:"messages"`
	Page        int                   `json:"page"`
	PageSize    int                   `json:"page_size"`
	TotalCount  int64                 `json:"total_count"`
	TotalPages  int                   `json:"total_pages"`
}

// --- Helper Function for common device validation ---
func (h *MessageHandler) validateDevice(c *gin.Context, adminUserID uint, deviceJIDStr string) (types.JID, bool) {
	// Check if device_jid is managed by adminUserID
	var managedDevice models.ManagedDevice
	err := h.DB.Where("admin_user_id = ? AND device_jid = ?", adminUserID, deviceJIDStr).First(&managedDevice).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusForbidden, gin.H{"error": "Device JID " + deviceJIDStr + " is not managed by this admin user or does not exist."})
			return types.JID{}, false
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error validating device: " + err.Error()})
		return types.JID{}, false
	}

	// Parse DeviceJID
	parsedDeviceJID, ok := utils.ParseJID(deviceJIDStr)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Device JID format for " + deviceJIDStr})
		return types.JID{}, false
	}

	// Check if the selected device is logged in
	if h.CmdHandler == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Messaging service not available."})
		return types.JID{}, false
	}
	client, clientOk := commandhandler.LoadClientConcurrent(parsedDeviceJID.User)
	if !clientOk || client == nil || !client.IsLoggedIn() {
		c.JSON(http.StatusPreconditionFailed, gin.H{"error": "Selected device " + deviceJIDStr + " is not currently logged in."})
		return types.JID{}, false
	}
	return parsedDeviceJID, true
}

// --- Handler Implementations ---

// SendMessage handles sending a single message
func (h *MessageHandler) SendMessage(c *gin.Context) {
	adminUserIDVal, exists := c.Get("admin_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Admin user ID not found in token claims"})
		return
	}
	adminUserIDFloat, ok := adminUserIDVal.(float64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid admin user ID type in token claims"})
		return
	}
	adminUserID := uint(adminUserIDFloat)

	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	if strings.TrimSpace(req.RecipientNumber) == "" || strings.TrimSpace(req.MessageContent) == "" || strings.TrimSpace(req.DeviceJID) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Device JID, recipient number, and message content cannot be empty"})
		return
	}

	parsedDeviceJID, isValidDevice := h.validateDevice(c, adminUserID, req.DeviceJID)
	if !isValidDevice {
		return // Error response already sent by validateDevice
	}

	// Send the message
	sentAt := time.Now()
	// The first return is a timestamp string from WhatsApp, often referred to as MessageID
	waMsgIDStr, err := h.CmdHandler.HandleSendTextMessage(parsedDeviceJID, req.MessageContent, req.RecipientNumber)

	status := "sent"
	if err != nil {
		status = "failed"
	}

	// Create SentMessage record
	sentMessage := models.SentMessage{
		AdminUserID:     adminUserID,
		DeviceJID:       req.DeviceJID,
		RecipientNumber: req.RecipientNumber,
		MessageContent:  req.MessageContent,
		Status:          status,
		SentAt:          sentAt,
		// MessageIDFromWhatsApp: &waMsgIDStr, // Store pointer to string
	}
    if waMsgIDStr != "" && err == nil { // Only store if successfully sent and ID is not empty
        sentMessage.MessageIDFromWhatsApp = &waMsgIDStr
    }


	if dbErr := h.DB.Create(&sentMessage).Error; dbErr != nil {
		// Log this, but the message might have been sent.
		// Consider how to handle this case: maybe a retry mechanism for DB save?
		// For now, if sending failed, the client knows. If sending succeeded but DB failed,
		// the client might get a success but the record isn't there.
		// Or, return an error if DB save fails.
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record message sending status: " + dbErr.Error()})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to send message: " + err.Error(),
			"message_id": sentMessage.ID, // ID from our database
			"status": status,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Message sent successfully",
		"message_id": sentMessage.ID, // ID from our database
		"whatsapp_message_id": waMsgIDStr,
		"status": status,
	})
}

// SendBulkMessage handles sending messages to multiple recipients
func (h *MessageHandler) SendBulkMessage(c *gin.Context) {
	adminUserIDVal, exists := c.Get("admin_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Admin user ID not found in token claims"})
		return
	}
	adminUserIDFloat, ok := adminUserIDVal.(float64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid admin user ID type in token claims"})
		return
	}
	adminUserID := uint(adminUserIDFloat)

	var req SendBulkMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	if strings.TrimSpace(req.DeviceJID) == "" || strings.TrimSpace(req.MessageContent) == "" || len(req.RecipientNumbers) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Device JID, message content, and at least one recipient number are required"})
		return
	}
	for _, num := range req.RecipientNumbers {
		if strings.TrimSpace(num) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Recipient numbers cannot be empty"})
			return
		}
	}


	parsedDeviceJID, isValidDevice := h.validateDevice(c, adminUserID, req.DeviceJID)
	if !isValidDevice {
		return // Error response already sent by validateDevice
	}

	var results []BulkMessageReportItem
	errorsOccurred := false

	for _, recipient := range req.RecipientNumbers {
		sentAt := time.Now()
		waMsgIDStr, err := h.CmdHandler.HandleSendTextMessage(parsedDeviceJID, req.MessageContent, recipient)
		
		item := BulkMessageReportItem{
			RecipientNumber: recipient,
		}

		status := "sent"
		if err != nil {
			status = "failed"
			item.Error = err.Error()
			errorsOccurred = true
		} else {
		    item.MessageID = waMsgIDStr
        }
        item.Status = status
        

		// Create SentMessage record regardless of send status for now, to log the attempt
		sentMessage := models.SentMessage{
			AdminUserID:     adminUserID,
			DeviceJID:       req.DeviceJID,
			RecipientNumber: recipient,
			MessageContent:  req.MessageContent,
			Status:          status,
			SentAt:          sentAt,
		}
        if waMsgIDStr != "" && err == nil {
             sentMessage.MessageIDFromWhatsApp = &waMsgIDStr
        }

		if dbErr := h.DB.Create(&sentMessage).Error; dbErr != nil {
			// Log this critical error: failed to save a message sending attempt
			// log.Printf("CRITICAL: Failed to save SentMessage record for %s to %s: %v", req.DeviceJID, recipient, dbErr)
			// Add to results that DB save failed for this recipient
			item.Status = "failed_to_record"
			item.Error = "Database error: " + dbErr.Error()
			errorsOccurred = true // Mark that an error occurred
		}
		results = append(results, item)
	}

	overallStatus := "completed_successfully"
	if errorsOccurred {
		overallStatus = "completed_with_errors"
	}

	c.JSON(http.StatusOK, SendBulkMessageResponse{
		OverallStatus: overallStatus,
		Results:       results,
	})
}

// GetMessageReport retrieves a paginated report of sent messages
func (h *MessageHandler) GetMessageReport(c *gin.Context) {
	adminUserIDVal, exists := c.Get("admin_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Admin user ID not found in token claims"})
		return
	}
	adminUserIDFloat, ok := adminUserIDVal.(float64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid admin user ID type in token claims"})
		return
	}
	adminUserID := uint(adminUserIDFloat)

	pageQuery := c.DefaultQuery("page", "1")
	pageSizeQuery := c.DefaultQuery("page_size", "10")

	page, err := strconv.Atoi(pageQuery)
	if err != nil || page < 1 {
		page = 1
	}
	pageSize, err := strconv.Atoi(pageSizeQuery)
	if err != nil || pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 { // Max page size
		pageSize = 100
	}

	offset := (page - 1) * pageSize

	var sentMessages []models.SentMessage
	var totalCount int64

	// Count total records for this admin
	if err := h.DB.Model(&models.SentMessage{}).Where("admin_user_id = ?", adminUserID).Count(&totalCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count messages: " + err.Error()})
		return
	}

	// Fetch paginated records
	err = h.DB.Where("admin_user_id = ?", adminUserID).
		Order("sent_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&sentMessages).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve messages: " + err.Error()})
		return
	}

	responseMessages := make([]SentMessageResponse, len(sentMessages))
	for i, msg := range sentMessages {
		responseMessages[i] = SentMessageResponse{
			ID:                   msg.ID,
			AdminUserID:          msg.AdminUserID,
			DeviceJID:            msg.DeviceJID,
			RecipientNumber:      msg.RecipientNumber,
			MessageContent:       msg.MessageContent,
			Status:               msg.Status,
			SentAt:               msg.SentAt,
			MessageIDFromWhatsApp: msg.MessageIDFromWhatsApp,
			CreatedAt:            msg.CreatedAt,
		}
	}

	totalPages := 0
	if totalCount > 0 {
		totalPages = (int(totalCount) + pageSize - 1) / pageSize
	}


	c.JSON(http.StatusOK, MessageReportResponse{
		Messages:    responseMessages,
		Page:        page,
		PageSize:    pageSize,
		TotalCount:  totalCount,
		TotalPages:  totalPages,
	})
}
