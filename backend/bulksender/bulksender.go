package bulksender

import (
	"context"
	"fmt"
	"math/rand/v2"
	"strings"
	"sync"
	"time"

	"whatsapp_multi_session/config"
	"whatsapp_multi_session/warmup"

	log "github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

// BulkResult holds the result of a single message send attempt.
type BulkResult struct {
	Recipient string `json:"recipient"`
	MessageID string `json:"message_id,omitempty"`
	Success   bool   `json:"success"`
	Error     string `json:"error,omitempty"`
}

// dailyCounter tracks per-sender daily message counts.
var (
	dailyCounterMu sync.Mutex
	dailyCounts    = make(map[string]*senderCount)
)

type senderCount struct {
	Count int
	Date  string // YYYY-MM-DD
}

// GetDailyCount returns the current daily send count for a sender.
func GetDailyCount(senderUser string) int {
	dailyCounterMu.Lock()
	defer dailyCounterMu.Unlock()

	today := time.Now().Format("2006-01-02")
	sc, exists := dailyCounts[senderUser]
	if !exists || sc.Date != today {
		return 0
	}
	return sc.Count
}

// IncrementDailyCount increments and returns the new count. Returns false if limit exceeded.
func IncrementDailyCount(senderUser string, limit int) (int, bool) {
	dailyCounterMu.Lock()
	defer dailyCounterMu.Unlock()

	today := time.Now().Format("2006-01-02")
	sc, exists := dailyCounts[senderUser]
	if !exists || sc.Date != today {
		dailyCounts[senderUser] = &senderCount{Count: 1, Date: today}
		return 1, true
	}

	if sc.Count >= limit {
		return sc.Count, false
	}

	sc.Count++
	return sc.Count, true
}

// ResetDailyCount resets the daily counter for a sender (useful for testing).
func ResetDailyCount(senderUser string) {
	dailyCounterMu.Lock()
	defer dailyCounterMu.Unlock()
	delete(dailyCounts, senderUser)
}

// getConfig returns the bulk send config with defaults applied.
func getConfig() config.BulkSendConfig {
	cfg := config.Conf.BulkSend

	if cfg.MinDelay <= 0 {
		cfg.MinDelay = 15000 // 15 seconds
	}
	if cfg.MaxDelay <= 0 {
		cfg.MaxDelay = 45000 // 45 seconds
	}
	if cfg.MaxDelay < cfg.MinDelay {
		cfg.MaxDelay = cfg.MinDelay + 5000
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 10
	}
	if cfg.BatchPauseMin <= 0 {
		cfg.BatchPauseMin = 300 // 5 minutes
	}
	if cfg.BatchPauseMax <= 0 {
		cfg.BatchPauseMax = 600 // 10 minutes
	}
	if cfg.BatchPauseMax < cfg.BatchPauseMin {
		cfg.BatchPauseMax = cfg.BatchPauseMin + 60
	}
	if cfg.DailyLimit <= 0 {
		cfg.DailyLimit = 50
	}
	if cfg.TypingDelayMin <= 0 {
		cfg.TypingDelayMin = 2000 // 2 seconds
	}
	if cfg.TypingDelayMax <= 0 {
		cfg.TypingDelayMax = 5000 // 5 seconds
	}
	if cfg.TypingDelayMax < cfg.TypingDelayMin {
		cfg.TypingDelayMax = cfg.TypingDelayMin + 1000
	}
	if cfg.AllowedHourStart <= 0 {
		cfg.AllowedHourStart = 8 // 8 AM
	}
	if cfg.AllowedHourEnd <= 0 {
		cfg.AllowedHourEnd = 22 // 10 PM
	}
	if cfg.Timezone == "" {
		cfg.Timezone = "Local"
	}
	if cfg.ErrorBackoffMinutes <= 0 {
		cfg.ErrorBackoffMinutes = 30
	}
	if cfg.ValidationCacheDuration <= 0 {
		cfg.ValidationCacheDuration = 24
	}
	if cfg.MaxErrorRate <= 0 {
		cfg.MaxErrorRate = 0.3 // 30%
	}

	return cfg
}

// ApplyTemplate replaces {{key}} placeholders in the message with values from the variables map.
// It also adds slight random variation to avoid identical messages.
func ApplyTemplate(message string, variables map[string]string, recipientPhone string) string {
	result := message

	// Replace user-defined variables
	for key, value := range variables {
		placeholder := fmt.Sprintf("{{%s}}", key)
		result = strings.ReplaceAll(result, placeholder, value)
	}

	// Replace built-in variables
	result = strings.ReplaceAll(result, "{{phone}}", recipientPhone)

	return result
}

// isWithinAllowedHours checks if the current time is within allowed sending hours.
func isWithinAllowedHours(cfg config.BulkSendConfig) (bool, time.Duration) {
	if !cfg.EnableTimeRestrictions {
		return true, 0
	}

	// Load timezone
	var loc *time.Location
	var err error
	if cfg.Timezone == "Local" || cfg.Timezone == "" {
		loc = time.Local
	} else {
		loc, err = time.LoadLocation(cfg.Timezone)
		if err != nil {
			log.Warnf("[TimeCheck] Invalid timezone %s, using Local: %v", cfg.Timezone, err)
			loc = time.Local
		}
	}

	now := time.Now().In(loc)
	currentHour := now.Hour()

	// Check if within allowed hours
	if currentHour >= cfg.AllowedHourStart && currentHour < cfg.AllowedHourEnd {
		return true, 0
	}

	// Calculate wait time until next allowed hour
	var nextAllowedTime time.Time
	if currentHour < cfg.AllowedHourStart {
		// Wait until start hour today
		nextAllowedTime = time.Date(now.Year(), now.Month(), now.Day(), cfg.AllowedHourStart, 0, 0, 0, loc)
	} else {
		// Wait until start hour tomorrow
		tomorrow := now.AddDate(0, 0, 1)
		nextAllowedTime = time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), cfg.AllowedHourStart, 0, 0, 0, loc)
	}

	waitDuration := nextAllowedTime.Sub(now)
	return false, waitDuration
}

// isRateLimitError checks if an error is a rate limit or spam-related error.
func isRateLimitError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "rate") ||
		strings.Contains(errStr, "too many") ||
		strings.Contains(errStr, "spam") ||
		strings.Contains(errStr, "limit") ||
		strings.Contains(errStr, "throttle")
}

// SendBulkSequential sends messages one by one with human-like delays, presence simulation,
// batch pauses, and daily limit enforcement. This is the anti-ban bulk sender.
func SendBulkSequential(
	ctx context.Context,
	client *whatsmeow.Client,
	sender types.JID,
	recipients []string,
	message string,
	variables map[string]string,
	parseJID func(string) (types.JID, bool),
	warmUpService *warmup.Service,
) []BulkResult {
	cfg := getConfig()
	rng := rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), uint64(time.Now().UnixNano()<<1)))

	dailyLimit := cfg.DailyLimit
	if warmUpService != nil {
		warmUpLimit, err := warmUpService.GetCurrentDailyLimit(sender.String())
		if err != nil {
			log.Warnf("[BulkSend] Failed to get warm-up limit, using config default: %v", err)
		} else if warmUpLimit > 0 {
			dailyLimit = warmUpLimit
			log.Infof("[BulkSend] Using warm-up daily limit: %d for sender %s", dailyLimit, sender.User)
		}
	}

	// Check time-of-day restrictions
	if allowed, waitDuration := isWithinAllowedHours(cfg); !allowed {
		log.Warnf("[BulkSend] Outside allowed hours. Next allowed time in %v", waitDuration)
		// Return all as failed with time restriction message
		results := make([]BulkResult, len(recipients))
		for i, recipient := range recipients {
			results[i] = BulkResult{
				Recipient: recipient,
				Success:   false,
				Error:     fmt.Sprintf("outside allowed hours, retry in %v", waitDuration),
			}
		}
		return results
	}

	results := make([]BulkResult, 0, len(recipients))
	batchCount := 0

	for i, recipientStr := range recipients {
		// Check context cancellation
		select {
		case <-ctx.Done():
			log.Warnf("[BulkSend] Context cancelled, stopping at recipient %d/%d", i+1, len(recipients))
			results = append(results, BulkResult{
				Recipient: recipientStr,
				Success:   false,
				Error:     "cancelled",
			})
			return results
		default:
		}

		// Check daily limit
		_, allowed := IncrementDailyCount(sender.User, dailyLimit)
		if !allowed {
			log.Warnf("[BulkSend] Daily limit (%d) reached for sender %s, stopping", dailyLimit, sender.User)
			// Mark remaining as skipped
			for j := i; j < len(recipients); j++ {
				results = append(results, BulkResult{
					Recipient: recipients[j],
					Success:   false,
					Error:     fmt.Sprintf("daily limit reached (%d messages)", dailyLimit),
				})
			}
			return results
		}

		// Parse recipient JID
		recipient, ok := parseJID(recipientStr)
		if !ok {
			results = append(results, BulkResult{
				Recipient: recipientStr,
				Success:   false,
				Error:     "invalid JID",
			})
			continue
		}

		// Simulate typing presence (composing)
		if cfg.EnablePresenceSimulation {
			err := client.SendChatPresence(context.Background(), recipient, types.ChatPresenceComposing, types.ChatPresenceMediaText)
			if err != nil {
				log.Warnf("[BulkSend] Failed to send composing presence to %s: %v", recipientStr, err)
			}

			// Wait for typing duration
			typingDelay := time.Duration(rng.IntN(cfg.TypingDelayMax-cfg.TypingDelayMin)+cfg.TypingDelayMin) * time.Millisecond
			time.Sleep(typingDelay)

			// Stop composing
			_ = client.SendChatPresence(context.Background(), recipient, types.ChatPresencePaused, types.ChatPresenceMediaText)
		}

		// Apply message template with per-recipient variables
		finalMessage := ApplyTemplate(message, variables, recipientStr)

		// Send the message
		msg := &waE2E.Message{
			Conversation: proto.String(finalMessage),
		}

		messageID := client.GenerateMessageID()
		resp, err := client.SendMessage(ctx, recipient, msg, whatsmeow.SendRequestExtra{ID: messageID})
		if err != nil {
			log.Errorf("[BulkSend] Error sending to %s: %v", recipientStr, err)

			// Check if this is a rate limit error
			if isRateLimitError(err) {
				backoffDuration := time.Duration(cfg.ErrorBackoffMinutes) * time.Minute
				log.Errorf("[BulkSend] RATE LIMIT DETECTED! Backing off for %v", backoffDuration)

				// Mark remaining recipients as failed
				for j := i; j < len(recipients); j++ {
					results = append(results, BulkResult{
						Recipient: recipients[j],
						Success:   false,
						Error:     fmt.Sprintf("rate limit detected, backing off for %v", backoffDuration),
					})
				}

				log.Warnf("[BulkSend] Stopped at %d/%d due to rate limit", i+1, len(recipients))
				return results
			}

			results = append(results, BulkResult{
				Recipient: recipientStr,
				Success:   false,
				Error:     err.Error(),
			})
		} else {
			// Mark as read
			_ = client.MarkRead(context.Background(), []types.MessageID{resp.ID}, time.Now(), recipient, sender)

			// Delete after send if configured
			if config.Conf.DeleteAfterSend.Enable {
				_, errRevoke := client.SendMessage(ctx, recipient, client.BuildRevoke(recipient, types.EmptyJID, messageID))
				if errRevoke != nil {
					log.Warnf("[BulkSend] Error revoking message to %s: %v", recipientStr, errRevoke)
				}
			}

			results = append(results, BulkResult{
				Recipient: recipientStr,
				MessageID: resp.ID,
				Success:   true,
			})
			log.Infof("[BulkSend] Sent to %s (msg %d/%d)", recipientStr, i+1, len(recipients))
		}

		batchCount++

		// Batch pause: after every N messages, take a longer break
		if batchCount >= cfg.BatchSize && i < len(recipients)-1 {
			batchPause := time.Duration(rng.IntN(cfg.BatchPauseMax-cfg.BatchPauseMin)+cfg.BatchPauseMin) * time.Second
			log.Infof("[BulkSend] Batch pause: sleeping %v after %d messages", batchPause, batchCount)
			time.Sleep(batchPause)
			batchCount = 0
		} else if i < len(recipients)-1 {
			// Random delay between messages
			delay := time.Duration(rng.IntN(cfg.MaxDelay-cfg.MinDelay)+cfg.MinDelay) * time.Millisecond
			log.Debugf("[BulkSend] Delay before next message: %v", delay)
			time.Sleep(delay)
		}
	}

	log.Infof("[BulkSend] Completed: %d/%d messages sent for sender %s", countSuccess(results), len(recipients), sender.User)
	return results
}

func countSuccess(results []BulkResult) int {
	count := 0
	for _, r := range results {
		if r.Success {
			count++
		}
	}
	return count
}
