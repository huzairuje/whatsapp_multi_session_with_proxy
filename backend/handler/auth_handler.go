package handler

import (
	"fmt"
	"net/http"

	"whatsapp_multi_session/activity"
	"whatsapp_multi_session/auth"

	"github.com/gin-gonic/gin"
)

func (h Handler) HandleLogin(c *gin.Context) {
	var req auth.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	resp, err := h.AuthService.Login(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	h.ActivityService.LogActivity(activity.TypeUserLogin, fmt.Sprintf("User %s logged in", req.Username), "", req.Username, "", "success", "")

	c.JSON(http.StatusOK, resp)
}

func (h Handler) HandleRefreshToken(c *gin.Context) {
	var req auth.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	resp, err := h.AuthService.RefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}

	h.ActivityService.LogActivity(activity.TypeUserLogin, fmt.Sprintf("User %s refreshed token", resp.Username), "", resp.Username, "", "success", "")

	c.JSON(http.StatusOK, resp)
}

func (h Handler) HandleChangePassword(c *gin.Context) {
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req auth.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if err := h.AuthService.ChangePassword(username.(string), req.OldPassword, req.NewPassword); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "password changed successfully"})
}
