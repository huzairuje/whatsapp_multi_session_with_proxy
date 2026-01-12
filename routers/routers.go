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

	// Dashboard routes
	// Redirect /dashboard to /dashboard/login
	router.GET("/dashboard", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/dashboard/login")
	})
	// Login page (GET and POST)
	router.GET("/dashboard/login", r.Handler.ServeLoginPage())
	router.POST("/dashboard/login", r.Handler.ServeLoginPage())

	// Authenticated dashboard group
	dashboardAuthGroup := router.Group("/dashboard").Use(handler.AuthMiddleware())
	{
		// Dashboard home page
		dashboardAuthGroup.GET("/home", r.Handler.ServeHomePage())
		// Dashboard content pages (for HTMX partials that require Go templating/logic)
		dashboardAuthGroup.GET("/page/:page", r.Handler.ServeDashboardContent())
		// API routes for dashboard data (HTMX targets)
		dashboardAuthGroup.GET("/api/sent-messages", r.Handler.GetSentMessages())
		dashboardAuthGroup.GET("/api/device-count", r.Handler.GetDeviceCount())
		dashboardAuthGroup.GET("/api/message-graphic", r.Handler.GetMessageCountGraphic())
		dashboardAuthGroup.GET("/api/active-senders", r.Handler.GetActiveSenders())
		dashboardAuthGroup.POST("/api/send-message", r.Handler.HandleDashboardSendMessage()) // New route for sending messages
	}

	// Serve static HTML files for dashboard partials like messages.html, devices.html etc.
	// These are requested by hx-get attributes in base.html's navigation.
	// These are kept outside the auth group as they are static assets.
	// The handlers that render base.html (which then loads these) ARE protected.
	// If direct access to these partials also needs protection, StaticFS would need a more complex setup.
	router.StaticFS("/dashboard", http.Dir("dashboard"))


	return router
}
