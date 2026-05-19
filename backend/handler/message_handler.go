package handler

import (
	"net/http"
	"strconv"

	"whatsapp_multi_session/message"

	"github.com/gin-gonic/gin"
)

func (h Handler) HandleRecordMessage(c *gin.Context) {
	var req message.CreateMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	msg, err := h.MessageService.RecordMessage(req.Sender, req.Recipient, req.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to record message"})
		return
	}

	c.JSON(http.StatusOK, msg)
}

func (h Handler) HandleUpdateMessageStatus(c *gin.Context) {
	var req message.UpdateMessageStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	err := h.MessageService.UpdateMessageStatus(req.MessageID, req.Status, req.Error)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update message status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "status updated"})
}

func (h Handler) HandleGetMessageStats(c *gin.Context) {
	sender := c.Query("sender")
	if sender == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sender parameter required"})
		return
	}

	stats, err := h.MessageService.GetStatsBySender(sender)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func (h Handler) HandleGetAllMessageStats(c *gin.Context) {
	stats, err := h.MessageService.GetAllStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func (h Handler) HandleGetMessages(c *gin.Context) {
	sender := c.Query("sender")
	if sender == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sender parameter required"})
		return
	}

	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 50
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0
	}

	messages, err := h.MessageService.GetMessagesBySender(sender, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get messages"})
		return
	}

	c.JSON(http.StatusOK, messages)
}
