package routes

import (
	"whatsapp_multi_session/dashboard/handlers" // Path to the AuthHandler

	"github.com/gin-gonic/gin"
)

// SetupAuthRoutes registers the authentication routes for the dashboard.
// It takes a Gin router group (e.g., /dashboard/auth) and the authentication handler.
func SetupAuthRoutes(routerGroup *gin.RouterGroup, authHandler *handlers.AuthHandler) {
	// Public routes
	routerGroup.POST("/register", authHandler.Register)
	routerGroup.POST("/request-otp", authHandler.RequestOTP)
	routerGroup.POST("/login", authHandler.Login)

	// Routes that will later be protected by JWT middleware.
	// For now, they are registered directly. The middleware will be added
	// to this group or these specific routes when the main dashboard router is set up.
	routerGroup.GET("/detail", authHandler.GetAuthDetail)
	routerGroup.POST("/logout", authHandler.Logout)
}
