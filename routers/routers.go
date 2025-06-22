// routers/routers.go
package routers

import (
	"whatsapp_multi_session/handler" // Existing WA handler
	dashboard_handlers "whatsapp_multi_session/app/handlers" // Our new dashboard handlers

	"github.com/gin-gonic/gin"
)

type Router struct {
	Handler handler.Handler // Existing WA handler
    // No need to add dashboard handlers to the struct if they are directly called in route setup
}

func NewRoutes(handler handler.Handler) Router {
	return Router{
		Handler: handler,
	}
}

// V1 sets up routes for the WhatsApp application
func (r Router) V1(router *gin.Engine) *gin.Engine {
	router.Use(gin.Recovery()) // Middleware for WA routes

	// Define WA routers
	router.GET("/connect", r.Handler.HandleConnect)
	// ... other existing WA routes from the original file...
    // Assuming original file had these based on common patterns:
    router.POST("/send-message", r.Handler.HandleSendMessage)
    router.POST("/send-image", r.Handler.HandleSendImage)
    router.POST("/send-video", r.Handler.HandleSendVideo)
    router.POST("/send-audio", r.Handler.HandleSendAudio)
    router.POST("/send-document", r.Handler.HandleSendDocument)
    router.GET("/logout", r.Handler.HandleLogout) // Note: WA /logout vs Dashboard /logout
    router.GET("/reconnect", r.Handler.HandleReconnect)
    router.GET("/get-group", r.Handler.HandleGetGroup)
    router.GET("/get-all-groups", r.Handler.HandleGetAllGroups)
    router.GET("/get-contact", r.Handler.HandleGetContact)
    router.GET("/get-all-contacts", r.Handler.HandleGetAllContacts)
    router.GET("/get-profile-pic", r.Handler.HandleGetProfilePic)
    router.GET("/get-qr-code", r.Handler.HandleGetQrCode)
    router.GET("/get-session-info", r.Handler.HandleGetSessionInfo)
    router.GET("/check-connection", r.Handler.HandleCheckConnection)
    router.GET("/list-sessions", r.Handler.HandleListSessions)
    router.POST("/save-session", r.Handler.HandleSaveSession)
    router.DELETE("/delete-session", r.Handler.HandleDeleteSession)
	router.GET("/health-check", r.Handler.HandleHealthCheck)

	return router
}

// AddDashboardRoutes sets up routes for the new dashboard application features
func AddDashboardRoutes(engine *gin.Engine) {
    // Dashboard routes
    // Grouping under /dashboard might be good for namespacing:
    // dashGroup := engine.Group("/dashboard")
    // For now, keep paths as they were for easier transition, but be mindful of conflicts.

    // Auth Routes
    // If there's a name conflict with WA's /logout, one must change.
    // For now, assuming they can coexist if methods differ or if WA's /logout is POST-only, for example.
    // If both are GET /logout, the last one defined for the same method will take precedence or behavior is undefined.
    // It's better to prefix dashboard routes. Example: /dash/register
    // However, sticking to the plan for this step.
    engine.GET("/register", dashboard_handlers.ShowRegistrationPageGin)
    engine.POST("/register", dashboard_handlers.HandleRegistrationGin)
    engine.GET("/login", dashboard_handlers.ShowLoginPageGin)
    engine.POST("/login", dashboard_handlers.HandleLoginGin)
    engine.GET("/logout", dashboard_handlers.LogoutHandlerGin) // This is a GET /logout for dashboard

    // Dashboard Main Page
    // Potential conflict with WA app if it also serves "/" and is registered on the same engine.
    // If WA app's V1 routes are added to 'engine' first, and then dashboard routes,
    // the order of engine.GET("/") definitions might matter or lead to unexpected behavior.
    // Best practice: use distinct paths or a group for the dashboard.
    engine.GET("/", dashboard_handlers.HomeHandlerGin)

    // Other Dashboard Pages
    engine.GET("/devices", dashboard_handlers.ShowDevicesPageGin)
    engine.GET("/messages", dashboard_handlers.ShowMessagesPageGin)
    engine.GET("/settings", dashboard_handlers.SettingsPageGETGin)
    engine.POST("/settings/password", dashboard_handlers.HandleChangePasswordGin)
}
