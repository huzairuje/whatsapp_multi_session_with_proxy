package handler

import (
	"net/http"

	"whatsapp_multi_session/primitive"
	"whatsapp_multi_session/warmup"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow/types"
)

func (h Handler) HandleCreateWarmUp(c *gin.Context) {
	var req warmup.CreateWarmUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.CommandHandler.Validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config, err := h.CommandHandler.WarmUpService.Create(req)
	if err != nil {
		log.Errorf("[WarmUp] Failed to create warmup config: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create warmup config"})
		return
	}

	c.JSON(http.StatusCreated, config)
}

func (h Handler) HandleGetWarmUp(c *gin.Context) {
	senderString := c.Query("sender")
	if senderString == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": primitive.MessageSenderShouldBeFilled})
		return
	}

	senderJID := types.NewJID(senderString, types.DefaultUserServer)

	config, err := h.CommandHandler.WarmUpService.GetBySenderJID(senderJID.String())
	if err != nil {
		log.Errorf("[WarmUp] Failed to get warmup config: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get warmup config"})
		return
	}

	if config == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "warmup config not found"})
		return
	}

	c.JSON(http.StatusOK, config)
}

func (h Handler) HandleGetAllWarmUp(c *gin.Context) {
	configs, err := h.CommandHandler.WarmUpService.GetAll()
	if err != nil {
		log.Errorf("[WarmUp] Failed to get all warmup configs: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get warmup configs"})
		return
	}

	c.JSON(http.StatusOK, configs)
}

func (h Handler) HandleUpdateWarmUp(c *gin.Context) {
	senderString := c.Query("sender")
	if senderString == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": primitive.MessageSenderShouldBeFilled})
		return
	}

	senderJID := types.NewJID(senderString, types.DefaultUserServer)

	var req warmup.UpdateWarmUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config, err := h.CommandHandler.WarmUpService.Update(senderJID.String(), req)
	if err != nil {
		log.Errorf("[WarmUp] Failed to update warmup config: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update warmup config"})
		return
	}

	c.JSON(http.StatusOK, config)
}

func (h Handler) HandleDeleteWarmUp(c *gin.Context) {
	senderString := c.Query("sender")
	if senderString == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": primitive.MessageSenderShouldBeFilled})
		return
	}

	senderJID := types.NewJID(senderString, types.DefaultUserServer)

	err := h.CommandHandler.WarmUpService.Delete(senderJID.String())
	if err != nil {
		log.Errorf("[WarmUp] Failed to delete warmup config: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete warmup config"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "warmup config deleted successfully"})
}

func (h Handler) HandleGetWarmUpStatus(c *gin.Context) {
	senderString := c.Query("sender")
	if senderString == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": primitive.MessageSenderShouldBeFilled})
		return
	}

	senderJID := types.NewJID(senderString, types.DefaultUserServer)

	config, err := h.CommandHandler.WarmUpService.GetBySenderJID(senderJID.String())
	if err != nil {
		log.Errorf("[WarmUp] Failed to get warmup config: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get warmup config"})
		return
	}

	if config == nil {
		c.JSON(http.StatusOK, gin.H{
			"enabled":        false,
			"current_limit":  0,
			"message":        "warmup not configured",
		})
		return
	}

	currentLimit, err := h.CommandHandler.WarmUpService.GetCurrentDailyLimit(senderJID.String())
	if err != nil {
		log.Errorf("[WarmUp] Failed to get current daily limit: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get current daily limit"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"enabled":        config.Enabled,
		"current_day":    config.CurrentDay,
		"current_limit":  currentLimit,
		"max_limit":      config.MaxDailyLimit,
		"start_date":     config.StartDate,
		"config":         config,
	})
}
