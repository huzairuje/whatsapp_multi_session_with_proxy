package utils

import (
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// ValidationResult stores the result of a recipient validation check.
type ValidationResult struct {
	IsValid   bool
	JID       string
	CheckedAt time.Time
	ExpiresAt time.Time
}

// ValidationCache caches recipient validation results to avoid repeated checks.
type ValidationCache struct {
	mu    sync.RWMutex
	cache map[string]*ValidationResult
}

var (
	globalValidationCache *ValidationCache
	cacheOnce             sync.Once
)

// GetValidationCache returns the global validation cache singleton.
func GetValidationCache() *ValidationCache {
	cacheOnce.Do(func() {
		globalValidationCache = &ValidationCache{
			cache: make(map[string]*ValidationResult),
		}
		// Start cleanup goroutine
		go globalValidationCache.cleanupExpired()
	})
	return globalValidationCache
}

// Get retrieves a cached validation result if it exists and hasn't expired.
func (vc *ValidationCache) Get(recipient string) (*ValidationResult, bool) {
	vc.mu.RLock()
	defer vc.mu.RUnlock()

	result, exists := vc.cache[recipient]
	if !exists {
		return nil, false
	}

	// Check if expired
	if time.Now().After(result.ExpiresAt) {
		return nil, false
	}

	return result, true
}

// Set stores a validation result in the cache with the specified TTL.
func (vc *ValidationCache) Set(recipient string, isValid bool, jid string, ttlHours int) {
	vc.mu.Lock()
	defer vc.mu.Unlock()

	if ttlHours <= 0 {
		ttlHours = 24 // Default 24 hours
	}

	result := &ValidationResult{
		IsValid:   isValid,
		JID:       jid,
		CheckedAt: time.Now(),
		ExpiresAt: time.Now().Add(time.Duration(ttlHours) * time.Hour),
	}

	vc.cache[recipient] = result
	log.Debugf("[ValidationCache] Cached result for %s: valid=%v, expires in %dh", recipient, isValid, ttlHours)
}

// Delete removes a recipient from the cache.
func (vc *ValidationCache) Delete(recipient string) {
	vc.mu.Lock()
	defer vc.mu.Unlock()

	delete(vc.cache, recipient)
}

// Clear removes all entries from the cache.
func (vc *ValidationCache) Clear() {
	vc.mu.Lock()
	defer vc.mu.Unlock()

	vc.cache = make(map[string]*ValidationResult)
	log.Info("[ValidationCache] Cache cleared")
}

// Size returns the number of entries in the cache.
func (vc *ValidationCache) Size() int {
	vc.mu.RLock()
	defer vc.mu.RUnlock()

	return len(vc.cache)
}

// GetStats returns cache statistics.
func (vc *ValidationCache) GetStats() map[string]interface{} {
	vc.mu.RLock()
	defer vc.mu.RUnlock()

	validCount := 0
	invalidCount := 0
	expiredCount := 0
	now := time.Now()

	for _, result := range vc.cache {
		if now.After(result.ExpiresAt) {
			expiredCount++
		} else if result.IsValid {
			validCount++
		} else {
			invalidCount++
		}
	}

	return map[string]interface{}{
		"total_entries":  len(vc.cache),
		"valid_entries":  validCount,
		"invalid_entries": invalidCount,
		"expired_entries": expiredCount,
	}
}

// cleanupExpired periodically removes expired entries from the cache.
func (vc *ValidationCache) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		vc.mu.Lock()
		now := time.Now()
		removed := 0

		for recipient, result := range vc.cache {
			if now.After(result.ExpiresAt) {
				delete(vc.cache, recipient)
				removed++
			}
		}

		vc.mu.Unlock()

		if removed > 0 {
			log.Infof("[ValidationCache] Cleaned up %d expired entries", removed)
		}
	}
}
