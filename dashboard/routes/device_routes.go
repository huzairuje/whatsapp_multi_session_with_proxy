package routes

import (
	"whatsapp_multi_session/dashboard/handlers" // Adjust if your handler package alias is different

	"github.com/gin-gonic/gin"
)

// SetupDeviceRoutes registers the device management routes for the dashboard.
// It expects the routerGroup to already be protected by JWT middleware.
func SetupDeviceRoutes(routerGroup *gin.RouterGroup, deviceHandler *handlers.DeviceHandler) {
	routerGroup.POST("/devices", deviceHandler.AddDevice)
	routerGroup.GET("/devices", deviceHandler.ListDevices)
}
