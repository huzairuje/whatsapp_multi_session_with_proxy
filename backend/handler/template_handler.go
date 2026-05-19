package handler

import (
	"net/http"
	"strconv"

	"whatsapp_multi_session/template"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func (h Handler) HandleCreateTemplate(c *gin.Context) {
	var req template.CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.CommandHandler.Validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tmpl, err := h.CommandHandler.TemplateService.Create(req)
	if err != nil {
		log.Errorf("[Template] Failed to create template: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create template"})
		return
	}

	c.JSON(http.StatusCreated, tmpl)
}

func (h Handler) HandleGetTemplate(c *gin.Context) {
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

	tmpl, err := h.CommandHandler.TemplateService.GetByID(id)
	if err != nil {
		log.Errorf("[Template] Failed to get template: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get template"})
		return
	}

	if tmpl == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "template not found"})
		return
	}

	c.JSON(http.StatusOK, tmpl)
}

func (h Handler) HandleGetAllTemplates(c *gin.Context) {
	templates, err := h.CommandHandler.TemplateService.GetAll()
	if err != nil {
		log.Errorf("[Template] Failed to get all templates: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get templates"})
		return
	}

	c.JSON(http.StatusOK, templates)
}

func (h Handler) HandleUpdateTemplate(c *gin.Context) {
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

	var req template.UpdateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tmpl, err := h.CommandHandler.TemplateService.Update(id, req)
	if err != nil {
		log.Errorf("[Template] Failed to update template: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update template"})
		return
	}

	c.JSON(http.StatusOK, tmpl)
}

func (h Handler) HandleDeleteTemplate(c *gin.Context) {
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

	err = h.CommandHandler.TemplateService.Delete(id)
	if err != nil {
		log.Errorf("[Template] Failed to delete template: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete template"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "template deleted successfully"})
}

func (h Handler) HandlePreviewTemplate(c *gin.Context) {
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

	var req struct {
		Recipients []template.RecipientData `json:"recipients" validate:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	previews, err := h.CommandHandler.TemplateService.PreviewTemplate(id, req.Recipients)
	if err != nil {
		log.Errorf("[Template] Failed to preview template: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to preview template"})
		return
	}

	c.JSON(http.StatusOK, previews)
}
