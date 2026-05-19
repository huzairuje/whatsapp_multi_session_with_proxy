package handler

import (
	"net/http"
	"strconv"

	"whatsapp_multi_session/activity"

	"github.com/gin-gonic/gin"
)

func (h Handler) HandleLogActivity(c *gin.Context) {
	var req activity.LogActivityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	act, err := h.ActivityService.LogActivity(req.Type, req.Message, req.Sender, req.User, req.Details, req.Status, req.Error)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to log activity"})
		return
	}

	c.JSON(http.StatusOK, act)
}

func (h Handler) HandleGetRecentActivities(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 50
	}

	activities, err := h.ActivityService.GetRecentActivities(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get activities"})
		return
	}

	c.JSON(http.StatusOK, activities)
}

func (h Handler) HandleGetActivitiesBySender(c *gin.Context) {
	sender := c.Query("sender")
	if sender == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sender parameter required"})
		return
	}

	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 50
	}

	activities, err := h.ActivityService.GetActivitiesBySender(sender, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get activities"})
		return
	}

	c.JSON(http.StatusOK, activities)
}

func (h Handler) HandleGetActivitiesByType(c *gin.Context) {
	typeStr := c.Query("type")
	if typeStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "type parameter required"})
		return
	}

	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 50
	}

	activities, err := h.ActivityService.GetActivitiesByType(activity.ActivityType(typeStr), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get activities"})
		return
	}

	c.JSON(http.StatusOK, activities)
}

func (h Handler) HandleGetActivityStats(c *gin.Context) {
	stats, err := h.ActivityService.GetStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}
