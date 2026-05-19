package routers

import (
	"whatsapp_multi_session/auth"
	"whatsapp_multi_session/handler"

	"github.com/gin-gonic/gin"
)

type Router struct {
	Handler     handler.Handler
	AuthService *auth.Service
}

func NewRoutes(handler handler.Handler, authService *auth.Service) Router {
	return Router{
		Handler:     handler,
		AuthService: authService,
	}
}

func (r Router) V1(router *gin.Engine) *gin.Engine {
	// use middleware for recover
	router.Use(gin.Recovery())

	// Public routes (no auth required)
	router.POST("/login", r.Handler.HandleLogin)
	router.POST("/refresh-token", r.Handler.HandleRefreshToken)
	router.GET("/health-check", r.Handler.HandleHealthCheck)

	// Protected routes (auth required)
	protected := router.Group("/")
	protected.Use(auth.AuthMiddleware(r.AuthService))
	{
		protected.POST("/change-password", r.Handler.HandleChangePassword)
		protected.GET("/connect", r.Handler.HandleConnect)
		protected.POST("/connect-bulk", r.Handler.HandleConnectBulk)
		protected.GET("/disconnect", r.Handler.HandleDisconnect)
		protected.POST("/disconnect-bulk", r.Handler.HandleDisconnectBulk)
		protected.GET("/autologin", r.Handler.ServeAutoLogin)
		protected.GET("/auto-disconnect", r.Handler.ServeAutoDisconnect)
		protected.GET("/qr", r.Handler.HandleQR)
		protected.GET("/pair-code", r.Handler.HandlePairCode)
		protected.GET("/qr-json", r.Handler.HandleQRResponseJson)
		protected.POST("/presence", r.Handler.ServeSendPresence)
		protected.POST("/send", r.Handler.ServeSendText)
		protected.POST("/send-bulk", r.Handler.ServeSendTextBulk)
		protected.GET("/send-bulk/status", r.Handler.ServeBulkSendStatus)
		protected.GET("/status", r.Handler.ServeStatus)
		protected.POST("/check-user", r.Handler.ServeCheckUser)
		protected.POST("/check-user-single", r.Handler.ServeCheckUserSingle)
		protected.POST("/upload", r.Handler.NewUploadHandler)
		protected.POST("/upload-single", r.Handler.NewUploadSingleHandler)
		protected.GET("/devices", r.Handler.ServeAllDevices)
		protected.GET("/device-proxies", r.Handler.ServeDeviceProxies)
		protected.GET("/devices/:jid", r.Handler.ServeDetailDevices)
		protected.POST("/logout", r.Handler.Logout)
		protected.DELETE("/message", r.Handler.DeleteMessages)
		
		// Message tracking endpoints
		protected.GET("/messages", r.Handler.HandleGetMessages)
		protected.GET("/messages/stats", r.Handler.HandleGetMessageStats)
		protected.GET("/messages/stats/all", r.Handler.HandleGetAllMessageStats)
		protected.POST("/messages/status", r.Handler.HandleUpdateMessageStatus)
		
		// Activity logging endpoints
		protected.POST("/activities/log", r.Handler.HandleLogActivity)
		protected.GET("/activities", r.Handler.HandleGetRecentActivities)
		protected.GET("/activities/sender", r.Handler.HandleGetActivitiesBySender)
		protected.GET("/activities/type", r.Handler.HandleGetActivitiesByType)
		protected.GET("/activities/stats", r.Handler.HandleGetActivityStats)
		
		// Warm-up endpoints
		protected.POST("/warmup", r.Handler.HandleCreateWarmUp)
		protected.GET("/warmup", r.Handler.HandleGetWarmUp)
		protected.GET("/warmup/all", r.Handler.HandleGetAllWarmUp)
		protected.PUT("/warmup", r.Handler.HandleUpdateWarmUp)
		protected.DELETE("/warmup", r.Handler.HandleDeleteWarmUp)
		protected.GET("/warmup/status", r.Handler.HandleGetWarmUpStatus)
		
		// Template endpoints
		protected.POST("/templates", r.Handler.HandleCreateTemplate)
		protected.GET("/templates", r.Handler.HandleGetTemplate)
		protected.GET("/templates/all", r.Handler.HandleGetAllTemplates)
		protected.PUT("/templates", r.Handler.HandleUpdateTemplate)
		protected.DELETE("/templates", r.Handler.HandleDeleteTemplate)
		protected.POST("/templates/preview", r.Handler.HandlePreviewTemplate)
		
		// Scheduler endpoints
		protected.POST("/scheduler/jobs", r.Handler.HandleCreateScheduledJob)
		protected.GET("/scheduler/jobs", r.Handler.HandleGetScheduledJob)
		protected.GET("/scheduler/jobs/pending", r.Handler.HandleGetPendingJobs)
		
		// Contacts endpoints
		protected.GET("/contacts", r.Handler.HandleGetContacts)
		protected.GET("/contacts/search", r.Handler.HandleSearchContacts)
		protected.POST("/contacts/sync", r.Handler.HandleSyncContacts)
		protected.DELETE("/contacts", r.Handler.HandleDeleteContact)
	}

	return router
}
