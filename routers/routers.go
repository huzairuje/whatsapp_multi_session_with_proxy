package routers

import (
	"whatsapp_multi_session/handler"

	"github.com/gin-gonic/gin"
)

type Router struct {
	Handler handler.Handler
}

func NewRoutes(handler handler.Handler) Router {
	return Router{
		Handler: handler,
	}
}

func (r Router) V1(router *gin.Engine) *gin.Engine {
	// use middleware for recover
	router.Use(gin.Recovery())

	// Define routers
	router.GET("/connect", r.Handler.HandleConnect)
	router.POST("/connect-bulk", r.Handler.HandleConnectBulk)
	router.GET("/disconnect", r.Handler.HandleDisconnect)
	router.POST("/disconnect-bulk", r.Handler.HandleDisconnectBulk)
	router.GET("/autologin", r.Handler.ServeAutoLogin)
	router.GET("/auto-disconnect", r.Handler.ServeAutoDisconnect)
	router.GET("/qr", r.Handler.HandleQR)
	router.GET("/pair-code", r.Handler.HandlePairCode)
	router.GET("/qr-json", r.Handler.HandleQRResponseJson)
	router.POST("/presence", r.Handler.ServeSendPresence)
	router.POST("/send", r.Handler.ServeSendText)
	router.POST("/send-bulk", r.Handler.ServeSendTextBulk)
	router.GET("/status", r.Handler.ServeStatus)
	router.POST("/check-user", r.Handler.ServeCheckUser)
	router.POST("/check-user-single", r.Handler.ServeCheckUserSingle)
	router.POST("/upload", r.Handler.NewUploadHandler)
	router.POST("/upload-single", r.Handler.NewUploadSingleHandler)
	router.GET("/devices", r.Handler.ServeAllDevices)
	router.GET("/device-proxies", r.Handler.ServeDeviceProxies)
	router.GET("/devices/:jid", r.Handler.ServeDetailDevices)
	router.POST("/logout", r.Handler.Logout)
	router.DELETE("/message", r.Handler.DeleteMessages)
	router.GET("/health-check", r.Handler.HandleHealthCheck)

	return router
}
