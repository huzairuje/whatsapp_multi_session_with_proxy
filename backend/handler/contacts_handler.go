package handler

import (
	"net/http"
	"strconv"

	"whatsapp_multi_session/commandhandler"
	"whatsapp_multi_session/contacts"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow/types"
)

func (h Handler) HandleGetContacts(c *gin.Context) {
	senderString := c.Query("sender")
	if senderString == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sender parameter required"})
		return
	}

	limitStr := c.DefaultQuery("limit", "100")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	if limit > 1000 {
		limit = 1000
	}

	contactList, err := h.CommandHandler.ContactsService.GetContactsBySender(senderString, limit, offset)
	if err != nil {
		log.Errorf("[Contacts] Failed to get contacts: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get contacts"})
		return
	}

	count, err := h.CommandHandler.ContactsService.GetContactCount(senderString)
	if err != nil {
		log.Errorf("[Contacts] Failed to get contact count: %v", err)
		count = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"contacts": contactList,
		"total":    count,
		"limit":    limit,
		"offset":   offset,
	})
}

func (h Handler) HandleSearchContacts(c *gin.Context) {
	senderString := c.Query("sender")
	if senderString == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sender parameter required"})
		return
	}

	searchQuery := c.Query("q")
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	filter := contacts.ContactFilter{
		SenderJID:   senderString,
		SearchQuery: searchQuery,
		Limit:       limit,
		Offset:      offset,
	}

	contactList, err := h.CommandHandler.ContactsService.SearchContacts(filter)
	if err != nil {
		log.Errorf("[Contacts] Failed to search contacts: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to search contacts"})
		return
	}

	c.JSON(http.StatusOK, contactList)
}

func (h Handler) HandleDeleteContact(c *gin.Context) {
	var req contacts.DeleteContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.CommandHandler.Validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.CommandHandler.ContactsService.DeleteContact(req.SenderJID, req.ContactJID)
	if err != nil {
		log.Errorf("[Contacts] Failed to delete contact: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete contact"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "contact deleted successfully"})
}

func (h Handler) HandleSyncContacts(c *gin.Context) {
	var req contacts.SyncContactsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.CommandHandler.Validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	senderJID, err := types.ParseJID(req.SenderJID)
	if err != nil {
		log.Errorf("[Contacts] Failed to parse JID %s: %v", req.SenderJID, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sender JID"})
		return
	}

	log.Infof("[Contacts] Looking for client with key: %s (from JID: %s)", senderJID.User, req.SenderJID)
	
	client, ok := commandhandler.LoadClientConcurrent(senderJID.User)
	if !ok || client == nil {
		log.Errorf("[Contacts] Client not found for key: %s", senderJID.User)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "client not found"})
		return
	}

	if !client.IsLoggedIn() {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "client not logged in"})
		return
	}

	go func() {
		h.CommandHandler.SyncContactsFromWhatsApp(senderJID, req.Force)
	}()

	c.JSON(http.StatusOK, gin.H{
		"message": "contact sync started in background",
		"sender":  req.SenderJID,
	})
}
