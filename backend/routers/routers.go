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
	}

	return router
}
