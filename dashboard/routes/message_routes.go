package routes

import (
	"whatsapp_multi_session/dashboard/handlers" // Adjust if your handler package alias is different

	"github.com/gin-gonic/gin"
)

// SetupMessageRoutes registers the message sending and reporting routes for the dashboard.
// It expects the routerGroup to already be protected by JWT middleware.
// The base path for this routerGroup would typically be something like "/dashboard".
func SetupMessageRoutes(routerGroup *gin.RouterGroup, messageHandler *handlers.MessageHandler) {
	// Group for message sending operations, e.g., /dashboard/messages
	messageSendingRoutes := routerGroup.Group("/messages")
	{
		// Maps to POST /dashboard/messages
		messageSendingRoutes.POST("", messageHandler.SendMessage)
		// Maps to POST /dashboard/messages/bulk
		messageSendingRoutes.POST("/bulk", messageHandler.SendBulkMessage)
	}

	// Group for reporting operations, e.g., /dashboard/report
	reportRoutes := routerGroup.Group("/report")
	{
		// Maps to GET /dashboard/report/messages
		reportRoutes.GET("/messages", messageHandler.GetMessageReport)
	}
}
