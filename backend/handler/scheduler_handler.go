package handler

import (
	"net/http"
	"strconv"

	"whatsapp_multi_session/scheduler"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func (h Handler) HandleCreateScheduledJob(c *gin.Context) {
	var req scheduler.CreateScheduledJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.CommandHandler.Validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config := scheduler.ScheduleConfig{
		AllowedHourStart: 8,
		AllowedHourEnd:   22,
		Timezone:         "Local",
		DailyLimit:       50,
		MinDelayMs:       15000,
		MaxDelayMs:       45000,
	}

	job, err := h.CommandHandler.SchedulerService.ScheduleBulkSend(req, config)
	if err != nil {
		log.Errorf("[Scheduler] Failed to create scheduled job: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, job)
}

func (h Handler) HandleGetScheduledJob(c *gin.Context) {
	idStr := c.Query("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id parameter required"})
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	job, err := h.CommandHandler.SchedulerService.GetByID(id)
	if err != nil {
		log.Errorf("[Scheduler] Failed to get scheduled job: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, job)
}

func (h Handler) HandleGetPendingJobs(c *gin.Context) {
	jobs, err := h.CommandHandler.SchedulerService.GetPendingJobs()
	if err != nil {
		log.Errorf("[Scheduler] Failed to get pending jobs: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, jobs)
}
