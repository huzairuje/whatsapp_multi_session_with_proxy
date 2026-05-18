package commandhandler

import (
	"sync"
	"time"

	"whatsapp_multi_session/config"

	log "github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow"
)

// SessionHealth tracks health metrics for a WhatsApp session.
type SessionHealth struct {
	mu                sync.RWMutex
	lastConnectTime   time.Time
	totalSends        int
	failedSends       int
	lastErrorTime     time.Time
	consecutiveErrors int
	isHealthy         bool
}

var (
	sessionHealthMap = make(map[string]*SessionHealth)
	healthMapMu      sync.RWMutex
)

// GetSessionHealth returns the health tracker for a user, creating if needed.
func GetSessionHealth(userID string) *SessionHealth {
	healthMapMu.RLock()
	health, exists := sessionHealthMap[userID]
	healthMapMu.RUnlock()

	if exists {
		return health
	}

	// Create new health tracker
	healthMapMu.Lock()
	defer healthMapMu.Unlock()

	// Double-check after acquiring write lock
	if health, exists := sessionHealthMap[userID]; exists {
		return health
	}

	health = &SessionHealth{
		lastConnectTime: time.Now(),
		isHealthy:       true,
	}
	sessionHealthMap[userID] = health
	return health
}

// RecordConnect records a successful connection.
func (sh *SessionHealth) RecordConnect() {
	sh.mu.Lock()
	defer sh.mu.Unlock()

	sh.lastConnectTime = time.Now()
	sh.consecutiveErrors = 0
	sh.isHealthy = true
}

// RecordSendSuccess records a successful message send.
func (sh *SessionHealth) RecordSendSuccess() {
	sh.mu.Lock()
	defer sh.mu.Unlock()

	sh.totalSends++
	sh.consecutiveErrors = 0
}

// RecordSendFailure records a failed message send.
func (sh *SessionHealth) RecordSendFailure() {
	sh.mu.Lock()
	defer sh.mu.Unlock()

	sh.totalSends++
	sh.failedSends++
	sh.lastErrorTime = time.Now()
	sh.consecutiveErrors++

	// Mark unhealthy if too many consecutive errors
	if sh.consecutiveErrors >= 5 {
		sh.isHealthy = false
		log.Warnf("[Health] Session marked unhealthy after %d consecutive errors", sh.consecutiveErrors)
	}
}

// GetErrorRate returns the current error rate (0.0 to 1.0).
func (sh *SessionHealth) GetErrorRate() float64 {
	sh.mu.RLock()
	defer sh.mu.RUnlock()

	if sh.totalSends == 0 {
		return 0.0
	}

	return float64(sh.failedSends) / float64(sh.totalSends)
}

// GetConsecutiveErrors returns the number of consecutive errors.
func (sh *SessionHealth) GetConsecutiveErrors() int {
	sh.mu.RLock()
	defer sh.mu.RUnlock()
	return sh.consecutiveErrors
}

// IsHealthy returns whether the session is considered healthy.
func (sh *SessionHealth) IsHealthy() bool {
	sh.mu.RLock()
	defer sh.mu.RUnlock()
	return sh.isHealthy
}

// ResetHealth resets the health status (useful after recovery).
func (sh *SessionHealth) ResetHealth() {
	sh.mu.Lock()
	defer sh.mu.Unlock()

	sh.consecutiveErrors = 0
	sh.isHealthy = true
	log.Infof("[Health] Session health reset")
}

// CheckSessionHealth performs a comprehensive health check on a WhatsApp client.
func CheckSessionHealth(client *whatsmeow.Client, userID string) (bool, string) {
	if client == nil {
		return false, "client is nil"
	}

	// Check if connected
	if !client.IsConnected() {
		return false, "client not connected"
	}

	// Check if logged in
	if !client.IsLoggedIn() {
		return false, "client not logged in"
	}

	// Get health tracker
	health := GetSessionHealth(userID)

	// Check error rate
	cfg := config.Conf.BulkSend
	maxErrorRate := cfg.MaxErrorRate
	if maxErrorRate <= 0 {
		maxErrorRate = 0.3 // Default 30%
	}

	errorRate := health.GetErrorRate()
	if errorRate > maxErrorRate {
		return false, "error rate too high"
	}

	// Check consecutive errors
	if health.GetConsecutiveErrors() >= 5 {
		return false, "too many consecutive errors"
	}

	// Check if recently reconnected (suspicious)
	timeSinceConnect := time.Since(health.lastConnectTime)
	if timeSinceConnect < 5*time.Minute {
		log.Warnf("[Health] Session recently reconnected (%v ago), proceeding with caution", timeSinceConnect)
	}

	// Check health flag
	if !health.IsHealthy() {
		return false, "session marked unhealthy"
	}

	return true, "healthy"
}

// GetHealthStats returns health statistics for a user.
func GetHealthStats(userID string) map[string]interface{} {
	health := GetSessionHealth(userID)
	health.mu.RLock()
	defer health.mu.RUnlock()

	return map[string]interface{}{
		"total_sends":        health.totalSends,
		"failed_sends":       health.failedSends,
		"error_rate":         health.GetErrorRate(),
		"consecutive_errors": health.consecutiveErrors,
		"is_healthy":         health.isHealthy,
		"last_connect_time":  health.lastConnectTime,
		"last_error_time":    health.lastErrorTime,
	}
}

// CleanupHealthTracking removes health tracking for a user (call on logout/disconnect).
func CleanupHealthTracking(userID string) {
	healthMapMu.Lock()
	defer healthMapMu.Unlock()

	delete(sessionHealthMap, userID)
	log.Infof("[Health] Cleaned up health tracking for user %s", userID)
}
