package handler

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"

	"whatsapp_multi_session/activity"
	"whatsapp_multi_session/auth"
	"whatsapp_multi_session/bulksender"
	"whatsapp_multi_session/commandhandler"
	"whatsapp_multi_session/config"
	"whatsapp_multi_session/message"
	"whatsapp_multi_session/primitive"
	"whatsapp_multi_session/scheduler"
	"whatsapp_multi_session/utils"
	"whatsapp_multi_session/validator"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow/types"
)

type Handler struct {
	CommandHandler  commandhandler.CommandHandler
	AuthService     *auth.Service
	MessageService  *message.Service
	ActivityService *activity.Service
}

func NewHandler(commandhandler commandhandler.CommandHandler, authService *auth.Service, messageService *message.Service, activityService *activity.Service) Handler {
	return Handler{
		CommandHandler:  commandhandler,
		AuthService:     authService,
		MessageService:  messageService,
		ActivityService: activityService,
	}
}

// ServeSendPresence handles sending text messages
func (h Handler) ServeSendPresence(c *gin.Context) {
	// Get query parameters
	senderString := c.Query("sender")

	if senderString == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": primitive.MessageSenderShouldBeFilled})
		return
	}
	senderJidTypes := types.NewJID(senderString, types.DefaultUserServer)

	clientSpecificUser, ok := commandhandler.LoadClientConcurrent(senderJidTypes.User)
	if !ok || clientSpecificUser == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": primitive.MessageFailedToSend})
		return
	}

	if config.Conf.Proxy.Enable {
		proxy := h.CommandHandler.EnabledProxy(senderJidTypes.User)
		clientSpecificUser.SetProxy(http.ProxyURL(proxy))
	}

	if !clientSpecificUser.IsLoggedIn() {
		c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"message": primitive.MessageFailedToSend})
		return
	}

	err := h.CommandHandler.HandleSendPresence(senderJidTypes)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "success", "send_presence": http.StatusText(http.StatusOK)})
	return
}

// ServeSendText handles sending text messages
func (h Handler) ServeSendText(c *gin.Context) {
	// Get query parameters
	senderString := c.Query("sender")

	if senderString == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": primitive.MessageSenderShouldBeFilled})
		return
	}
	senderJidTypes := types.NewJID(senderString, types.DefaultUserServer)

	clientSpecificUser, ok := commandhandler.LoadClientConcurrent(senderJidTypes.User)
	if !ok || clientSpecificUser == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": primitive.MessageFailedToSend})
		return
	}

	if config.Conf.Proxy.Enable {
		proxy := h.CommandHandler.EnabledProxy(senderJidTypes.User)
		clientSpecificUser.SetProxy(http.ProxyURL(proxy))
	}

	if !clientSpecificUser.IsLoggedIn() {
		c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"message": primitive.MessageFailedToSend})
		return
	}

	var requestBody primitive.SendTextSingleRequest
	if err := c.BindJSON(&requestBody); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": primitive.MessageInvalidJson})
		return
	}

	errValidateStruct := validator.ValidateStructResponseSliceString(requestBody)
	if errValidateStruct != nil {
		log.Errorf("validator.ValidateStructResponseSliceString got err : %v", errValidateStruct)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("%v", errValidateStruct)})
		return
	}

	msgID, err := h.CommandHandler.HandleSendTextMessage(senderJidTypes, requestBody.Message, requestBody.Recipient)
	if err != nil {
		h.ActivityService.LogActivity(activity.TypeMessageFailed, fmt.Sprintf("Failed to send message to %s", requestBody.Recipient), senderString, "", fmt.Sprintf("Recipient: %s", requestBody.Recipient), "failed", err.Error())
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	// Record message to database
	_, recordErr := h.MessageService.RecordMessageWithID(senderString, requestBody.Recipient, requestBody.Message, msgID)
	if recordErr != nil {
		log.Errorf("failed to record message: %v", recordErr)
	}

	h.ActivityService.LogActivity(activity.TypeMessageSent, fmt.Sprintf("Message sent to %s", requestBody.Recipient), senderString, "", fmt.Sprintf("Message ID: %s, Recipient: %s", msgID, requestBody.Recipient), "success", "")

	c.JSON(http.StatusOK, gin.H{"message": "success", "id_pesan": msgID})
	return

}

// ServeSendTextBulk handles sending bulk text messages
func (h Handler) ServeSendTextBulk(c *gin.Context) {
	if c.Request.Method == "OPTIONS" {
		c.Status(http.StatusOK)
		return
	}

	// Get query parameters
	senderString := c.Query("sender")
	if senderString == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": primitive.MessageSenderShouldBeFilled})
		return
	}
	senderJidTypes := types.NewJID(senderString, types.DefaultUserServer)

	clientSpecificUser, ok := commandhandler.LoadClientConcurrent(senderJidTypes.User)
	if !ok || clientSpecificUser == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": primitive.MessageFailedToSend})
		return
	}

	if config.Conf.Proxy.Enable {
		proxy := h.CommandHandler.EnabledProxy(senderJidTypes.User)
		clientSpecificUser.SetProxy(http.ProxyURL(proxy))
	}

	if !clientSpecificUser.IsLoggedIn() {
		c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"message": primitive.MessageFailedToSend})
		return
	}

	var requestBody primitive.SendTextBulkRequest
	if err := c.BindJSON(&requestBody); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": primitive.MessageInvalidJson})
		return
	}

	errValidateStruct := validator.ValidateStructResponseSliceString(requestBody)
	if errValidateStruct != nil {
		log.Errorf("validator.ValidateStructResponseSliceString got err : %v", errValidateStruct)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("%v", errValidateStruct)})
		return
	}

	// Check if bulk send should be scheduled (> 10 recipients)
	if len(requestBody.Recipients) > 10 {
		// Use scheduler for large bulk sends
		schedReq := scheduler.CreateScheduledJobRequest{
			SenderJID:  senderString,
			Recipients: requestBody.Recipients,
		}

		if requestBody.TemplateID != nil {
			schedReq.TemplateID = requestBody.TemplateID
		}

		schedConfig := scheduler.ScheduleConfig{
			AllowedHourStart: config.Conf.BulkSend.AllowedHourStart,
			AllowedHourEnd:   config.Conf.BulkSend.AllowedHourEnd,
			Timezone:         config.Conf.BulkSend.Timezone,
			DailyLimit:       config.Conf.BulkSend.DailyLimit,
			MinDelayMs:       config.Conf.BulkSend.MinDelay,
			MaxDelayMs:       config.Conf.BulkSend.MaxDelay,
		}

		job, err := h.CommandHandler.SchedulerService.ScheduleBulkSend(schedReq, schedConfig)
		if err != nil {
			log.Errorf("[BulkSend] Failed to schedule bulk send: %v", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "failed to schedule bulk send"})
			return
		}

		h.ActivityService.LogActivity(activity.TypeBulkSendStart, fmt.Sprintf("Bulk send scheduled for %d recipients", len(requestBody.Recipients)), senderString, "", fmt.Sprintf("Job ID: %d, Recipients: %d", job.ID, len(requestBody.Recipients)), "scheduled", "")

		c.JSON(http.StatusOK, gin.H{
			"message":      "bulk send scheduled",
			"job_id":       job.ID,
			"recipients":   len(requestBody.Recipients),
			"scheduled_for": job.ScheduledFor,
			"note":         "messages will be sent gradually across multiple days to avoid account flagging",
		})
		return
	}

	// Run bulk send in background (it's sequential and slow by design)
	go func() {
		h.ActivityService.LogActivity(activity.TypeBulkSendStart, fmt.Sprintf("Bulk send started for %d recipients", len(requestBody.Recipients)), senderString, "", fmt.Sprintf("Recipients: %d", len(requestBody.Recipients)), "started", "")
		h.CommandHandler.HandleSendTextMessageBulk(senderJidTypes, requestBody.Message, requestBody.Recipients, requestBody.Variables)
	}()

	c.JSON(http.StatusOK, gin.H{
		"message":    "bulk send started",
		"recipients": len(requestBody.Recipients),
		"note":       "messages are being sent sequentially with anti-ban delays. Check logs for progress.",
	})
	return

}

// ServeBulkSendStatus returns the daily send count and limit for a sender.
func (h Handler) ServeBulkSendStatus(c *gin.Context) {
	senderString := c.Query("sender")
	if senderString == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": primitive.MessageSenderShouldBeFilled})
		return
	}

	dailyCount := bulksender.GetDailyCount(senderString)
	dailyLimit := config.Conf.BulkSend.DailyLimit
	if dailyLimit <= 0 {
		dailyLimit = 50
	}

	c.JSON(http.StatusOK, gin.H{
		"sender":      senderString,
		"daily_count": dailyCount,
		"daily_limit": dailyLimit,
		"remaining":   dailyLimit - dailyCount,
	})
}

// ServeStatus returns the current status of the client
func (h Handler) ServeStatus(c *gin.Context) {
	if c.Request.Method == "OPTIONS" {
		c.Status(http.StatusOK)
		return
	}

	// Get query parameters
	senderString := c.Query("sender")
	if senderString == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": primitive.MessageSenderShouldBeFilled})
		return
	}
	senderJidTypes := types.NewJID(senderString, types.DefaultUserServer)

	clientSpecificUser, ok := commandhandler.LoadClientConcurrent(senderJidTypes.User)
	if !ok || clientSpecificUser == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": primitive.MessageFailedToSend})
		return
	}

	if config.Conf.Proxy.Enable {
		proxy := h.CommandHandler.EnabledProxy(senderJidTypes.User)
		clientSpecificUser.SetProxy(http.ProxyURL(proxy))
	}

	if !clientSpecificUser.IsLoggedIn() {
		c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"message": primitive.MessageFailedToSend})
		return
	}

	response := primitive.StatusResponse{
		ID:       clientSpecificUser.Store.ID.String(),
		PushName: clientSpecificUser.Store.PushName,
		IsLogin:  clientSpecificUser.IsLoggedIn(),
	}

	c.JSON(http.StatusOK, response)
	return

}

// ServeCheckUser checks user status
func (h Handler) ServeCheckUser(c *gin.Context) {
	if c.Request.Method == "OPTIONS" {
		c.Status(http.StatusOK)
		return
	}

	// Get query parameters
	senderString := c.Query("sender")
	if senderString == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": primitive.MessageSenderShouldBeFilled})
		return
	}
	senderJidTypes := types.NewJID(senderString, types.DefaultUserServer)

	clientSpecificUser, ok := commandhandler.LoadClientConcurrent(senderJidTypes.User)
	if !ok || clientSpecificUser == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": primitive.MessageFailedToSend})
		return
	}

	if config.Conf.Proxy.Enable {
		proxy := h.CommandHandler.EnabledProxy(senderJidTypes.User)
		clientSpecificUser.SetProxy(http.ProxyURL(proxy))
	}

	if !clientSpecificUser.IsLoggedIn() {
		c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"message": primitive.MessageFailedToSend})
		return
	}

	var request primitive.CheckUserBulkRequest
	if err := c.BindJSON(&request); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": primitive.MessageInvalidJson})
		return
	}

	errValidateStruct := h.CommandHandler.Validator.Struct(request)
	if errValidateStruct != nil {
		log.Errorf("validator.Struct got err : %v", errValidateStruct)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("%v", errValidateStruct)})
		return
	}

	if len(request.Recipients) > 0 {
		for _, val := range request.Recipients {
			isRecipientValid := utils.ValidatePhoneNumber(val)
			if !isRecipientValid {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": primitive.MessageInvalidPhoneNumber})
				return
			}
		}
	}

	response := h.CommandHandler.HandleCheckUser(senderJidTypes, request.Recipients)
	c.JSON(http.StatusOK, response)
	return

}

func (h Handler) ServeAutoLogin(c *gin.Context) {
	if c.Request.Method == "OPTIONS" {
		c.Status(http.StatusOK)
		return
	}

	go func() {
		defer func() {
			// Recover from panic if any
			if r := recover(); r != nil {
				// Log the panic or handle it as needed
				log.Printf("Recovered from panic: %v", r)
			}
			// Clean up resources here if needed, e.g., closing connections, files, etc.
		}()

		// Execute the HandleLoginAllDevices function
		h.CommandHandler.HandleLoginAllDevices()
	}()

	h.ActivityService.LogActivity(activity.TypeAutoLogin, "Auto-login triggered for all devices", "", "", "", "started", "")
	c.JSON(http.StatusOK, gin.H{"message": primitive.MessageTriggeredAutoLogin})
	return
}

// ServeAutoDisconnect checks user status
func (h Handler) ServeAutoDisconnect(c *gin.Context) {
	if c.Request.Method == "OPTIONS" {
		c.Status(http.StatusOK)
		return
	}

	go func() {
		defer func() {
			// Recover from panic if any
			if r := recover(); r != nil {
				// Log the panic or handle it as needed
				log.Printf("Recovered from panic: %v", r)
			}
			// Clean up resources here if needed, e.g., closing connections, files, etc.
		}()

		// Execute the HandleDisconnectAllDevices function
		h.CommandHandler.HandleDisconnectAllDevices()
	}()
	h.ActivityService.LogActivity(activity.TypeAutoDisconnect, "Auto-disconnect triggered for all devices", "", "", "", "started", "")
	c.JSON(http.StatusOK, gin.H{"message": primitive.MessageTriggeredAutoDisconnect})
	return
}

// ServeAllDevices checks user status
func (h Handler) ServeAllDevices(c *gin.Context) {
	if c.Request.Method == "OPTIONS" {
		c.Status(http.StatusOK)
		return
	}

	response := h.CommandHandler.HandleGetAllDevices(c.Request.Context())
	c.JSON(http.StatusOK, response)
	return
}

func (h Handler) ServeDeviceProxies(c *gin.Context) {
	if c.Request.Method == "OPTIONS" {
		c.Status(http.StatusOK)
		return
	}

	list, err := h.CommandHandler.HandleDeviceProxies(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

// ServeDetailDevices checks user status
func (h Handler) ServeDetailDevices(c *gin.Context) {
	if c.Request.Method == "OPTIONS" {
		c.Status(http.StatusOK)
		return
	}

	jidStringReq := c.Param("jid")
	jid, ok := utils.ParseJID(jidStringReq)
	if !ok {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": primitive.MessageInvalidJidRequest})
		return
	}

	response := h.CommandHandler.HandleGetSingleDevices(c.Request.Context(), jid)
	if response.User == "" {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": primitive.MessageJidNotFound})
		return
	}

	c.JSON(http.StatusOK, response)
	return
}

// ServeCheckUserSingle checks user status
func (h Handler) ServeCheckUserSingle(c *gin.Context) {
	if c.Request.Method == "OPTIONS" {
		c.Status(http.StatusOK)
		return
	}

	// Get query parameters
	senderString := c.Query("sender")
	if senderString == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": primitive.MessageSenderShouldBeFilled})
		return
	}
	senderJidTypes := types.NewJID(senderString, types.DefaultUserServer)

	clientSpecificUser, ok := commandhandler.LoadClientConcurrent(senderJidTypes.User)
	if !ok || clientSpecificUser == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": primitive.MessageFailedToSend})
		return
	}

	if config.Conf.Proxy.Enable {
		proxy := h.CommandHandler.EnabledProxy(senderJidTypes.User)
		clientSpecificUser.SetProxy(http.ProxyURL(proxy))
	}

	if !clientSpecificUser.IsLoggedIn() {
		c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"message": primitive.MessageFailedToSend})
		return
	}

	var request primitive.CheckUserSingleRequest
	if err := c.BindJSON(&request); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": primitive.MessageInvalidJson})
		return
	}

	errValidateStruct := h.CommandHandler.Validator.Struct(request)
	if errValidateStruct != nil {
		log.Errorf("validator.Struct got err : %v", errValidateStruct)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("%v", errValidateStruct)})
		return
	}

	isRecipientValid := utils.ValidatePhoneNumber(request.Recipient)
	if !isRecipientValid {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": primitive.MessageInvalidPhoneNumber})
		return
	}

	response, err := h.CommandHandler.HandleCheckUserSingle(senderJidTypes, request.Recipient)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
	return

}

func (h Handler) NewUploadHandler(c *gin.Context) {
	// Get query parameters
	senderString := c.Query("sender")
	if senderString == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": primitive.MessageSenderShouldBeFilled})
		return
	}
	senderJidTypes := types.NewJID(senderString, types.DefaultUserServer)

	clientSpecificUser, ok := commandhandler.LoadClientConcurrent(senderJidTypes.User)
	if !ok || clientSpecificUser == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": primitive.MessageFailedToSend})
		return
	}

	if config.Conf.Proxy.Enable {
		proxy := h.CommandHandler.EnabledProxy(senderJidTypes.User)
		clientSpecificUser.SetProxy(http.ProxyURL(proxy))
	}

	if !clientSpecificUser.IsLoggedIn() {
		c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"message": primitive.MessageFailedToSend})
		return
	}

	err := c.Request.ParseMultipartForm(10 << 20)
	if err != nil {
		log.Errorf("Failed to parse multipart form : %v", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": primitive.MessageErrorRequestMultiPart})
		return
	}

	// Get the files
	files, ok := c.Request.MultipartForm.File["file"]
	if !ok || len(files) == 0 {
		log.Errorf("No files found in the request")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": primitive.MessageFileNotFound})
		return
	}

	recipientJIDs := c.Request.FormValue("recipients")
	captionMsg := c.Request.FormValue("caption")

	var resp []primitive.Message

	for _, handler := range files {
		// Open the file
		file, err := handler.Open()
		if err != nil {
			log.Errorf("Failed to open file : %v", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": primitive.MessageFailedToReadFileData})
			return
		}

		// Read the file data
		data, err := io.ReadAll(file)
		if err != nil {
			log.Errorf("Failed to read file data : %v", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": primitive.MessageFailedToReadFileData})
			return
		}

		// Close the file explicitly after reading the data
		if err := file.Close(); err != nil {
			log.Errorf("Error closing file: %v", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": primitive.MessageFailedToCloseFile})
			return
		}

		sliceJID, err := utils.ValidateStringArrayAsStringArray(recipientJIDs)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": primitive.MessageInvalidRecipient})
			return
		}

		var uploadResp []primitive.Message
		mimeType := http.DetectContentType(data)
		if utils.IsImage(mimeType) {
			uploadResp, err = h.CommandHandler.HandleSendImage(senderJidTypes, sliceJID, data, captionMsg)
		} else if utils.IsVideo(mimeType) {
			uploadResp, err = h.CommandHandler.HandleSendVideo(senderJidTypes, sliceJID, data, captionMsg)
		} else if utils.IsAudio(mimeType) {
			uploadResp, err = h.CommandHandler.HandleSendAudio(senderJidTypes, sliceJID, data)
		} else {
			uploadResp, err = h.CommandHandler.HandleSendDocument(senderJidTypes, sliceJID, handler.Filename, data, captionMsg)
		}
		if err != nil {
			log.Errorf("Failed to handle file upload : %v", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": primitive.MessageInvalidUploadFile})
			return
		}

		resp = append(resp, uploadResp...)
	}

	c.JSON(http.StatusOK, resp)
	return

}

func (h Handler) NewUploadSingleHandler(c *gin.Context) {
	// Get query parameters
	senderString := c.Query("sender")

	if senderString == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": primitive.MessageSenderShouldBeFilled})
		return
	}
	senderJidTypes := types.NewJID(senderString, types.DefaultUserServer)

	clientSpecificUser, ok := commandhandler.LoadClientConcurrent(senderJidTypes.User)
	if !ok || clientSpecificUser == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": primitive.MessageFailedToSend})
		return
	}

	if config.Conf.Proxy.Enable {
		proxy := h.CommandHandler.EnabledProxy(senderJidTypes.User)
		clientSpecificUser.SetProxy(http.ProxyURL(proxy))
	}

	if !clientSpecificUser.IsLoggedIn() {
		c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"message": primitive.MessageFailedToSend})
		return
	}

	var request primitive.SendSingleMediaRequest
	if err := c.BindJSON(&request); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": primitive.MessageInvalidJson})
		return
	}

	errValidateStruct := validator.ValidateStructResponseSliceString(request)
	if errValidateStruct != nil {
		log.Errorf("validator.ValidateStructResponseSliceString got err : %v", errValidateStruct)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("%v", errValidateStruct)})
		return
	}

	// Logic for handling get file from link
	fileURL := request.File
	resp, err := http.Get(fileURL)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": primitive.MessageFailedToDownloadFile})
		return
	}

	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": primitive.MessageFailedToDownloadFile})
		return
	}

	// Get the file name from the Content-Disposition header
	var fileName string
	contentDisposition := resp.Header.Get("Content-Disposition")
	if contentDisposition != "" {
		_, params, err := mime.ParseMediaType(contentDisposition)
		if err == nil {
			fileName = params["filename"]
		}
	}

	// If filename is not present in headers, you might want to parse it from the URL
	if fileName == "" {
		urlParts := strings.Split(fileURL, "/")
		fileName = urlParts[len(urlParts)-1]
	}

	// Read the file content
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": primitive.MessageFailedToReadFileData})
		return
	}

	// Close the response body immediately after reading
	if err = resp.Body.Close(); err != nil {
		log.Errorf("Error closing response body: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Failed to close response body"})
		return
	}

	var uploadResp []primitive.Message
	mimeType := http.DetectContentType(data)
	sliceJID := []string{request.Recipient}
	if utils.IsImage(mimeType) {
		uploadResp, err = h.CommandHandler.HandleSendImage(senderJidTypes, sliceJID, data, request.Caption)
		if err != nil {
			log.Errorf("Failed to handle file upload : %v", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": primitive.MessageFailedToSendData})
			return
		}
		response := uploadResp[0]
		c.JSON(http.StatusOK, response)
		return
	} else if utils.IsVideo(mimeType) {
		uploadResp, err = h.CommandHandler.HandleSendVideo(senderJidTypes, sliceJID, data, request.Caption)
		if err != nil {
			log.Errorf("Failed to handle file upload : %v", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": primitive.MessageFailedToSendData})
			return
		}
		response := uploadResp[0]
		c.JSON(http.StatusOK, response)
		return
	} else if utils.IsAudio(mimeType) {
		uploadResp, err = h.CommandHandler.HandleSendAudio(senderJidTypes, sliceJID, data)
		if err != nil {
			log.Errorf("Failed to handle file upload : %v", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": primitive.MessageFailedToSend})
			return
		}
		response := uploadResp[0]
		c.JSON(http.StatusOK, response)
		return
	} else {
		uploadResp, err = h.CommandHandler.HandleSendDocument(senderJidTypes, sliceJID, fileName, data, request.Caption)
		if err != nil {
			log.Errorf("Failed to handle file upload : %v", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": primitive.MessageFailedToSendData})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": primitive.MessageSuccessSent})
		return
	}
}

func (h Handler) HandleQR(c *gin.Context) {
	container := h.CommandHandler.Container
	if container == nil {
		err := errors.New("container is nil")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	devices, err := container.GetAllDevices(c.Request.Context())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	// Get query parameters
	senderString := c.Query("sender")
	if senderString == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": primitive.MessageSenderShouldBeFilled})
		return
	}
	senderJidTypes := types.NewJID(senderString, types.DefaultUserServer)

	if len(devices) > 0 {
		// Get specific QR code
		base64qrcode, err := h.CommandHandler.HandleGetSpecificQR(senderJidTypes)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}
		if base64qrcode == "" {
			err = errors.New(primitive.MessageAlreadyLoggedIn)
			c.AbortWithStatusJSON(http.StatusOK, gin.H{"message": err.Error()})
			return
		}

		h.ActivityService.LogActivity(activity.TypeQRGenerated, fmt.Sprintf("QR code generated for %s", senderString), senderString, "", "", "success", "")
		c.Data(http.StatusOK, "image/png", []byte(base64qrcode))
		return

	} else {
		// Get specific QR code
		base64qrcode, err := h.CommandHandler.HandleGetSingleQR(senderJidTypes)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}
		if base64qrcode == "" {
			err = errors.New(primitive.MessageAlreadyLoggedIn)
			c.AbortWithStatusJSON(http.StatusOK, gin.H{"message": err.Error()})
			return
		}
		h.ActivityService.LogActivity(activity.TypeQRGenerated, fmt.Sprintf("QR code generated for %s", senderString), senderString, "", "", "success", "")
		c.Data(http.StatusOK, "image/png", []byte(base64qrcode))
		return
	}
}

func (h Handler) HandleConnect(c *gin.Context) {
	// Get query parameters
	senderString := c.Query("sender")
	if senderString == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": primitive.MessageSenderShouldBeFilled})
		return
	}
	senderJidTypes := types.NewJID(senderString, types.DefaultUserServer)

	// get pair code
	err := h.CommandHandler.HandleConnectSingleDevice(c.Request.Context(), senderJidTypes)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	h.ActivityService.LogActivity(activity.TypeSessionConnect, fmt.Sprintf("Session %s connected", senderString), senderString, "", "", "success", "")

	c.JSON(http.StatusOK, gin.H{"message": "success reconnect!"})
	return

}

func (h Handler) HandleConnectBulk(c *gin.Context) {
	var request primitive.ConnectBulkDeviceRequest
	if err := c.BindJSON(&request); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": primitive.MessageInvalidJson})
		return
	}

	errValidateStruct := validator.ValidateStructResponseSliceString(request)
	if errValidateStruct != nil {
		log.Errorf("validator.ValidateStructResponseSliceString got err : %v", errValidateStruct)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("%v", errValidateStruct)})
		return
	}

	var senderJidTypesBulk []types.JID
	for _, sender := range request.Senders {
		senderJidTypesBulk = append(senderJidTypesBulk, types.NewJID(sender, types.DefaultUserServer))
	}

	err := h.CommandHandler.HandleConnectBulkDevices(c.Request.Context(), senderJidTypesBulk)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	h.ActivityService.LogActivity(activity.TypeSessionConnect, fmt.Sprintf("Bulk connect: %d devices connected", len(senderJidTypesBulk)), "", "", fmt.Sprintf("Devices: %v", senderJidTypesBulk), "success", "")
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("connected devices : %v", senderJidTypesBulk)})
	return
}

func (h Handler) HandleHealthCheck(c *gin.Context) {
	h.ActivityService.LogActivity(activity.TypeHealthCheck, "Health check performed", "", "", "", "success", "")
	c.JSON(http.StatusOK, gin.H{"message": "server is alive and ok!"})
	return
}

func (h Handler) HandleDisconnect(c *gin.Context) {
	// Get query parameters
	senderString := c.Query("sender")
	if senderString == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": primitive.MessageSenderShouldBeFilled})
		return
	}
	senderJidTypes := types.NewJID(senderString, types.DefaultUserServer)

	// get pair code
	err := h.CommandHandler.HandleDisconnectSingleDevice(c.Request.Context(), senderJidTypes)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	h.ActivityService.LogActivity(activity.TypeSessionDisconnect, fmt.Sprintf("Session %s disconnected", senderString), senderString, "", "", "success", "")

	c.JSON(http.StatusOK, gin.H{"message": "success disconnect!"})
	return

}

func (h Handler) HandleDisconnectBulk(c *gin.Context) {
	var request primitive.DisconnectBulkDeviceRequest
	if err := c.BindJSON(&request); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": primitive.MessageInvalidJson})
		return
	}

	errValidateStruct := validator.ValidateStructResponseSliceString(request)
	if errValidateStruct != nil {
		log.Errorf("validator.ValidateStructResponseSliceString got err : %v", errValidateStruct)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("%v", errValidateStruct)})
		return
	}

	var senderJidTypesBulk []types.JID
	for _, sender := range request.Senders {
		senderJidTypesBulk = append(senderJidTypesBulk, types.NewJID(sender, types.DefaultUserServer))
	}

	err := h.CommandHandler.HandleDisconnectBulkDevices(c.Request.Context(), senderJidTypesBulk)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	h.ActivityService.LogActivity(activity.TypeSessionDisconnect, fmt.Sprintf("Bulk disconnect: %d devices disconnected", len(senderJidTypesBulk)), "", "", fmt.Sprintf("Devices: %v", senderJidTypesBulk), "success", "")
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("disconnected devices : %v", senderJidTypesBulk)})
	return
}

func (h Handler) HandlePairCode(c *gin.Context) {
	// Get query parameters
	senderString := c.Query("sender")
	if senderString == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": primitive.MessageSenderShouldBeFilled})
		return
	}
	senderJidTypes := types.NewJID(senderString, types.DefaultUserServer)

	// get pair code
	pairCode, err := h.CommandHandler.HandleGetPairCode(c.Request.Context(), senderJidTypes)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	if pairCode == "" {
		err = errors.New(primitive.MessageAlreadyLoggedIn)
		c.JSON(http.StatusOK, gin.H{"message": err.Error()})
		return
	}

	h.ActivityService.LogActivity(activity.TypePairCode, fmt.Sprintf("Pair code generated for %s", senderString), senderString, "", fmt.Sprintf("Code: %s", pairCode), "success", "")
	c.JSON(http.StatusOK, gin.H{"pair_code": pairCode})
	return

}

func (h Handler) HandleQRResponseJson(c *gin.Context) {
	container := h.CommandHandler.Container
	if container == nil {
		err := errors.New("container is nil")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	devices, err := container.GetAllDevices(c.Request.Context())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	// Get query parameters
	senderString := c.Query("sender")
	if senderString == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": primitive.MessageSenderShouldBeFilled})
		return
	}
	senderJidTypes := types.NewJID(senderString, types.DefaultUserServer)

	if len(devices) > 0 {
		// Get specific QR code
		code, err := h.CommandHandler.HandleSpecificQRResponseCode(c.Request.Context(), senderJidTypes)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}
		if code == "" {
			err = errors.New(primitive.MessageAlreadyLoggedIn)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": err.Error(), "data": ""})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "success login", "data": code})
		return

	} else {
		// Get specific QR code
		code, err := h.CommandHandler.HandleGetSingleQRResponseCode(senderJidTypes)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}
		if code == "" {
			err = errors.New(primitive.MessageAlreadyLoggedIn)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": err.Error(), "data": ""})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "success login", "data": code})
		return
	}
}

// Logout checks user status
func (h Handler) Logout(c *gin.Context) {
	if c.Request.Method == "OPTIONS" {
		c.Status(http.StatusOK)
		return
	}

	// Get query parameters
	senderString := c.Query("sender")
	if senderString == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": primitive.MessageSenderShouldBeFilled})
		return
	}
	senderJidTypes := types.NewJID(senderString, types.DefaultUserServer)

	// logout device by sender
	err := h.CommandHandler.HandleLogoutSingleDevice(c.Request.Context(), senderJidTypes)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	h.ActivityService.LogActivity(activity.TypeLogout, fmt.Sprintf("Session %s logged out", senderString), senderString, "", "", "success", "")
	c.JSON(http.StatusOK, gin.H{"message": "success logout"})
	return
}

// DeleteMessages delete message checks user status
func (h Handler) DeleteMessages(c *gin.Context) {
	if c.Request.Method == "OPTIONS" {
		c.Status(http.StatusOK)
		return
	}

	// Get query parameters
	senderString := c.Query("sender")
	if senderString == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": primitive.MessageSenderShouldBeFilled})
		return
	}
	senderJidTypes := types.NewJID(senderString, types.DefaultUserServer)

	clientSpecificUser, ok := commandhandler.LoadClientConcurrent(senderJidTypes.User)
	if !ok || clientSpecificUser == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": primitive.MessageFailedToSend})
		return
	}

	if config.Conf.Proxy.Enable {
		proxy := h.CommandHandler.EnabledProxy(senderJidTypes.User)
		clientSpecificUser.SetProxy(http.ProxyURL(proxy))
	}

	if !clientSpecificUser.IsLoggedIn() {
		c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"message": primitive.MessageFailedToSend})
		return
	}

	var request primitive.DeleteMessagesRequest
	if err := c.BindJSON(&request); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": primitive.MessageInvalidJson})
		return
	}

	errValidateStruct := validator.ValidateStructResponseSliceString(request)
	if errValidateStruct != nil {
		log.Errorf("validator.ValidateStructResponseSliceString got err : %v", errValidateStruct)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("%v", errValidateStruct)})
		return
	}

	for _, msgID := range request.MessageIDs {
		recipientJidTypes, ok := utils.ParseJID(request.Recipient)
		if !ok {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Error parse recipient"})
			return
		}
		// delete message by message id
		_, errRevoke := clientSpecificUser.SendMessage(c.Request.Context(), recipientJidTypes, clientSpecificUser.BuildRevoke(recipientJidTypes, types.EmptyJID, msgID))
		if errRevoke != nil {
			log.Errorf("Error sendingrevoke message: %v", errRevoke)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Error sendingrevoke message"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "success delete message"})
	return

}
