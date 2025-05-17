package handler

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"

	"whatsapp_multi_session/commandhandler"
	"whatsapp_multi_session/config"
	"whatsapp_multi_session/primitive"
	"whatsapp_multi_session/utils"
	"whatsapp_multi_session/validator"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow/types"
)

type Handler struct {
	CommandHandler commandhandler.CommandHandler
}

func NewHandler(commandhandler commandhandler.CommandHandler) Handler {
	return Handler{
		CommandHandler: commandhandler,
	}
}

// ServeSendPresence handles sending text messages
func (h Handler) ServeSendPresence(c *gin.Context) {
	// Get query parameters
	senderString := c.Query("sender")

	if senderString == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "sender seharusnya diisi dengan nomor yang valid"})
		return
	}
	senderJidTypes := types.NewJID(senderString, types.DefaultUserServer)

	clientSpecificUser, ok := commandhandler.LoadClientConcurrent(senderJidTypes.User)
	if !ok || clientSpecificUser == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "gagal kirim"})
		return
	}

	if config.Conf.Proxy.Enable {
		proxy := h.CommandHandler.EnabledProxy(senderJidTypes.User)
		clientSpecificUser.SetProxy(http.ProxyURL(proxy))
	}

	if !clientSpecificUser.IsLoggedIn() {
		c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"message": "gagal kirim, tolong hit endpoint untuk melakukan qrcode"})
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
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "sender seharusnya diisi dengan nomor yang valid"})
		return
	}
	senderJidTypes := types.NewJID(senderString, types.DefaultUserServer)

	clientSpecificUser, ok := commandhandler.LoadClientConcurrent(senderJidTypes.User)
	if !ok || clientSpecificUser == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "gagal kirim"})
		return
	}

	if config.Conf.Proxy.Enable {
		proxy := h.CommandHandler.EnabledProxy(senderJidTypes.User)
		clientSpecificUser.SetProxy(http.ProxyURL(proxy))
	}

	if !clientSpecificUser.IsLoggedIn() {
		c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"message": "gagal kirim, tolong hit endpoint untuk melakukan qrcode"})
		return
	}

	var requestBody primitive.SendTextSingleRequest
	if err := c.BindJSON(&requestBody); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Error decoding JSON"})
		return
	}

	errValidateStruct := validator.ValidateStructResponseSliceString(requestBody)
	if errValidateStruct != nil {
		log.Errorf("validator.ValidateStructResponseSliceString got err : %v", errValidateStruct)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": errValidateStruct})
		return
	}

	msgID, err := h.CommandHandler.HandleSendTextMessage(senderJidTypes, requestBody.Message, requestBody.Recipient)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

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
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "sender seharusnya diisi dengan nomor yang valid"})
		return
	}
	senderJidTypes := types.NewJID(senderString, types.DefaultUserServer)

	clientSpecificUser, ok := commandhandler.LoadClientConcurrent(senderJidTypes.User)
	if !ok || clientSpecificUser == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "gagal kirim"})
		return
	}

	if config.Conf.Proxy.Enable {
		proxy := h.CommandHandler.EnabledProxy(senderJidTypes.User)
		clientSpecificUser.SetProxy(http.ProxyURL(proxy))
	}

	if !clientSpecificUser.IsLoggedIn() {
		c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"message": "gagal kirim, tolong hit endpoint untuk melakukan qrcode"})
		return
	}

	var requestBody primitive.SendTextBulkRequest
	if err := c.BindJSON(&requestBody); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "error decoding JSON"})
		return
	}

	errValidateStruct := validator.ValidateStructResponseSliceString(requestBody)
	if errValidateStruct != nil {
		log.Errorf("validator.ValidateStructResponseSliceString got err : %v", errValidateStruct)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": errValidateStruct})
		return
	}

	h.CommandHandler.HandleSendTextMessageBulk(senderJidTypes, requestBody.Message, requestBody.Recipients)

	c.JSON(http.StatusOK, gin.H{"message": "success"})
	return

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
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "sender should be filled"})
		return
	}
	senderJidTypes := types.NewJID(senderString, types.DefaultUserServer)

	clientSpecificUser, ok := commandhandler.LoadClientConcurrent(senderJidTypes.User)
	if !ok || clientSpecificUser == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "gagal kirim"})
		return
	}

	if config.Conf.Proxy.Enable {
		proxy := h.CommandHandler.EnabledProxy(senderJidTypes.User)
		clientSpecificUser.SetProxy(http.ProxyURL(proxy))
	}

	if !clientSpecificUser.IsLoggedIn() {
		c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"message": "gagal kirim, tolong hit endpoint untuk melakukan qrcode"})
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
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "sender should be filled"})
		return
	}
	senderJidTypes := types.NewJID(senderString, types.DefaultUserServer)

	clientSpecificUser, ok := commandhandler.LoadClientConcurrent(senderJidTypes.User)
	if !ok || clientSpecificUser == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "gagal kirim"})
		return
	}

	if config.Conf.Proxy.Enable {
		proxy := h.CommandHandler.EnabledProxy(senderJidTypes.User)
		clientSpecificUser.SetProxy(http.ProxyURL(proxy))
	}

	if !clientSpecificUser.IsLoggedIn() {
		c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"message": "gagal kirim, tolong hit endpoint untuk melakukan qrcode kembali"})
		return
	}

	var request primitive.CheckUserBulkRequest
	if err := c.BindJSON(&request); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Error decoding JSON"})
		return
	}

	errValidateStruct := validator.ValidateStructResponseSliceString(request)
	if errValidateStruct != nil {
		log.Errorf("validator.ValidateStructResponseSliceString got err : %v", errValidateStruct)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": errValidateStruct})
		return
	}

	if len(request.Recipients) > 0 {
		for _, val := range request.Recipients {
			isRecipientValid := utils.ValidatePhoneNumber(val)
			if !isRecipientValid {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": `your number recipient is not contains "+"`})
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

	c.JSON(http.StatusOK, gin.H{"message": "success trigger auto login"})
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
	c.JSON(http.StatusOK, gin.H{"message": "success trigger auto disconnect"})
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
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "invalid jid request"})
		return
	}

	response := h.CommandHandler.HandleGetSingleDevices(c.Request.Context(), jid)
	if response.User == "" {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "your request jid is not found"})
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
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "sender should be filled"})
		return
	}
	senderJidTypes := types.NewJID(senderString, types.DefaultUserServer)

	clientSpecificUser, ok := commandhandler.LoadClientConcurrent(senderJidTypes.User)
	if !ok || clientSpecificUser == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "gagal kirim"})
		return
	}

	if config.Conf.Proxy.Enable {
		proxy := h.CommandHandler.EnabledProxy(senderJidTypes.User)
		clientSpecificUser.SetProxy(http.ProxyURL(proxy))
	}

	if !clientSpecificUser.IsLoggedIn() {
		c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"message": "gagal kirim, tolong hit endpoint untuk melakukan qrcode kembali"})
		return
	}

	var request primitive.CheckUserSingleRequest
	if err := c.BindJSON(&request); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Error decoding JSON"})
		return
	}

	errValidateStruct := validator.ValidateStructResponseSliceString(request)
	if errValidateStruct != nil {
		log.Errorf("validator.ValidateStructResponseSliceString got err : %v", errValidateStruct)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": errValidateStruct})
		return
	}

	isRecipientValid := utils.ValidatePhoneNumber(request.Recipient)
	if !isRecipientValid {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": `your number recipient is not contains "+"`})
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
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "sender should be filled"})
		return
	}
	senderJidTypes := types.NewJID(senderString, types.DefaultUserServer)

	clientSpecificUser, ok := commandhandler.LoadClientConcurrent(senderJidTypes.User)
	if !ok || clientSpecificUser == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "gagal kirim"})
		return
	}

	if config.Conf.Proxy.Enable {
		proxy := h.CommandHandler.EnabledProxy(senderJidTypes.User)
		clientSpecificUser.SetProxy(http.ProxyURL(proxy))
	}

	if !clientSpecificUser.IsLoggedIn() {
		c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"message": "gagal kirim, tolong hit endpoint untuk melakukan qrcode kembali"})
		return
	}

	err := c.Request.ParseMultipartForm(10 << 20)
	if err != nil {
		log.Errorf("Failed to parse multipart form : %v", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Ada galat pada saat parse multipart form atau request tidak valid"})
		return
	}

	// Get the files
	files, ok := c.Request.MultipartForm.File["file"]
	if !ok || len(files) == 0 {
		log.Errorf("No files found in the request")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "No files found in the request"})
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
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Failed to read file data"})
			return
		}

		// Read the file data
		data, err := io.ReadAll(file)
		if err != nil {
			log.Errorf("Failed to read file data : %v", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Failed to read file data"})
			return
		}

		// Close the file explicitly after reading the data
		if err := file.Close(); err != nil {
			log.Errorf("Error closing file: %v", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Failed to close file"})
			return
		}

		sliceJID, err := utils.ValidateStringArrayAsStringArray(recipientJIDs)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Ada galat pada parameter recipients atau parameter recipients salah atau tidak ditemukan"})
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
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Ada galat pada saat upload file"})
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
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "sender seharusnya diisi dengan nomor yang valid"})
		return
	}
	senderJidTypes := types.NewJID(senderString, types.DefaultUserServer)

	clientSpecificUser, ok := commandhandler.LoadClientConcurrent(senderJidTypes.User)
	if !ok || clientSpecificUser == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "gagal kirim"})
		return
	}

	if config.Conf.Proxy.Enable {
		proxy := h.CommandHandler.EnabledProxy(senderJidTypes.User)
		clientSpecificUser.SetProxy(http.ProxyURL(proxy))
	}

	if !clientSpecificUser.IsLoggedIn() {
		c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"message": "gagal kirim, tolong hit endpoint untuk melakukan qrcode"})
		return
	}

	var request primitive.SendSingleMediaRequest
	if err := c.BindJSON(&request); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Error decoding JSON"})
		return
	}

	errValidateStruct := validator.ValidateStructResponseSliceString(request)
	if errValidateStruct != nil {
		log.Errorf("validator.ValidateStructResponseSliceString got err : %v", errValidateStruct)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": errValidateStruct})
		return
	}

	// Logic for handling get file from link
	fileURL := request.File
	resp, err := http.Get(fileURL)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Failed to download file"})
		return
	}

	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Failed to download file, invalid URL"})
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
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Failed to read file data"})
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
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Failed send data"})
			return
		}
		response := uploadResp[0]
		c.JSON(http.StatusOK, response)
		return
	} else if utils.IsVideo(mimeType) {
		uploadResp, err = h.CommandHandler.HandleSendVideo(senderJidTypes, sliceJID, data, request.Caption)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Failed send data"})
			return
		}
		response := uploadResp[0]
		c.JSON(http.StatusOK, response)
		return
	} else if utils.IsAudio(mimeType) {
		uploadResp, err = h.CommandHandler.HandleSendAudio(senderJidTypes, sliceJID, data)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Failed send data"})
			return
		}
		response := uploadResp[0]
		c.JSON(http.StatusOK, response)
		return
	} else {
		uploadResp, err = h.CommandHandler.HandleSendDocument(senderJidTypes, sliceJID, fileName, data, request.Caption)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Failed send data"})
			return
		}
		response := uploadResp[0]
		c.JSON(http.StatusOK, response)
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
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "sender should be filled"})
		return
	}
	senderJidTypes := types.NewJID(senderString, types.DefaultUserServer)

	if len(devices) > 0 {
		// Get specific QR code
		base64qrcode, err := h.CommandHandler.HandleGetSpecificQR(c.Request.Context(), senderJidTypes)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}
		if base64qrcode == "" {
			err = errors.New("you are already login")
			c.AbortWithStatusJSON(http.StatusOK, gin.H{"message": err.Error()})
			return
		}

		c.Data(http.StatusOK, "image/png", []byte(base64qrcode))
		return

	} else {
		// Get specific QR code
		base64qrcode, err := h.CommandHandler.HandleGetSingleQR(c.Request.Context(), senderJidTypes)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}
		if base64qrcode == "" {
			err = errors.New("you are already login")
			c.AbortWithStatusJSON(http.StatusOK, gin.H{"message": err.Error()})
			return
		}
		c.Data(http.StatusOK, "image/png", []byte(base64qrcode))
		return
	}
}

func (h Handler) HandleConnect(c *gin.Context) {
	// Get query parameters
	senderString := c.Query("sender")
	if senderString == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "sender should be filled"})
		return
	}
	senderJidTypes := types.NewJID(senderString, types.DefaultUserServer)

	// get pair code
	err := h.CommandHandler.HandleConnectSingleDevice(c.Request.Context(), senderJidTypes)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "success reconnect!"})
	return

}

func (h Handler) HandleConnectBulk(c *gin.Context) {
	var request primitive.ConnectBulkDeviceRequest
	if err := c.BindJSON(&request); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Error decoding JSON"})
		return
	}

	errValidateStruct := validator.ValidateStructResponseSliceString(request)
	if errValidateStruct != nil {
		log.Errorf("validator.ValidateStructResponseSliceString got err : %v", errValidateStruct)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": errValidateStruct})
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
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("connected devices : %v", senderJidTypesBulk)})
	return
}

func (h Handler) HandleHealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "server is alive and ok!"})
	return
}

func (h Handler) HandleDisconnect(c *gin.Context) {
	// Get query parameters
	senderString := c.Query("sender")
	if senderString == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "sender should be filled"})
		return
	}
	senderJidTypes := types.NewJID(senderString, types.DefaultUserServer)

	// get pair code
	err := h.CommandHandler.HandleDisconnectSingleDevice(c.Request.Context(), senderJidTypes)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "success disconnect!"})
	return

}

func (h Handler) HandleDisconnectBulk(c *gin.Context) {
	var request primitive.DisconnectBulkDeviceRequest
	if err := c.BindJSON(&request); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Error decoding JSON"})
		return
	}

	errValidateStruct := validator.ValidateStructResponseSliceString(request)
	if errValidateStruct != nil {
		log.Errorf("validator.ValidateStructResponseSliceString got err : %v", errValidateStruct)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": errValidateStruct})
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
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("disconnected devices : %v", senderJidTypesBulk)})
	return
}

func (h Handler) HandlePairCode(c *gin.Context) {
	// Get query parameters
	senderString := c.Query("sender")
	if senderString == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "sender should be filled"})
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
		err = errors.New("you are already login")
		c.JSON(http.StatusOK, gin.H{"message": err.Error()})
		return
	}

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
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "sender should be filled"})
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
			err = errors.New("you are already login")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": err.Error(), "data": ""})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "success login", "data": code})
		return

	} else {
		// Get specific QR code
		code, err := h.CommandHandler.HandleGetSingleQRResponseCode(c.Request.Context(), senderJidTypes)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}
		if code == "" {
			err = errors.New("you are already login")
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
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "sender should be filled"})
		return
	}
	senderJidTypes := types.NewJID(senderString, types.DefaultUserServer)

	// logout device by sender
	err := h.CommandHandler.HandleLogoutSingleDevice(c.Request.Context(), senderJidTypes)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

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
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "sender should be filled"})
		return
	}
	senderJidTypes := types.NewJID(senderString, types.DefaultUserServer)

	clientSpecificUser, ok := commandhandler.LoadClientConcurrent(senderJidTypes.User)
	if !ok || clientSpecificUser == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "gagal kirim"})
		return
	}

	if config.Conf.Proxy.Enable {
		proxy := h.CommandHandler.EnabledProxy(senderJidTypes.User)
		clientSpecificUser.SetProxy(http.ProxyURL(proxy))
	}

	if !clientSpecificUser.IsLoggedIn() {
		c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"message": "gagal kirim, tolong hit endpoint untuk melakukan qrcode"})
		return
	}

	var request primitive.DeleteMessagesRequest
	if err := c.BindJSON(&request); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Error decoding JSON"})
		return
	}

	errValidateStruct := validator.ValidateStructResponseSliceString(request)
	if errValidateStruct != nil {
		log.Errorf("validator.ValidateStructResponseSliceString got err : %v", errValidateStruct)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": errValidateStruct})
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
